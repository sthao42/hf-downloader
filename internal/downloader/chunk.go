package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// ChunkWorker downloads a specific byte range of a file.
type ChunkWorker struct {
	client    *http.Client
	token     string
	userAgent string
	limiter   *SpeedLimiter
}

// NewChunkWorker constructs a new range worker with optional speed limiter.
func NewChunkWorker(client *http.Client, token string, limiter ...*SpeedLimiter) *ChunkWorker {
	if client == nil {
		client = NewDownloadHTTPClient()
	}
	var lim *SpeedLimiter
	if len(limiter) > 0 {
		lim = limiter[0]
	}
	return &ChunkWorker{
		client:    client,
		token:     token,
		userAgent: "hf-downloader/1.0",
		limiter:   lim,
	}
}

// DownloadRange streams the specified range [start, end] into writer at appropriate offsets.
// onBytesRead callback is invoked with number of bytes read in each buffer cycle.
func (cw *ChunkWorker) DownloadRange(
	ctx context.Context,
	url string,
	start int64,
	end int64,
	writer *PartFileWriter,
	onBytesRead func(n int),
) error {
	if end >= 0 && start > end {
		return nil // Already complete
	}

	// Inactivity watchdog: abort if no data arrives for 30s so the engine can retry seamlessly
	watchdogCtx, cancelWatchdog := context.WithCancel(ctx)
	defer cancelWatchdog()

	activityChan := make(chan struct{}, 1)
	stalled := false
	var stallMu sync.Mutex

	go func() {
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()

		for {
			select {
			case <-watchdogCtx.Done():
				return
			case <-activityChan:
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(30 * time.Second)
			case <-timer.C:
				stallMu.Lock()
				stalled = true
				stallMu.Unlock()
				cancelWatchdog()
				return
			}
		}
	}()

	req, err := http.NewRequestWithContext(watchdogCtx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create range request: %w", err)
	}

	var rangeHeader string
	if end >= 0 {
		rangeHeader = fmt.Sprintf("bytes=%d-%d", start, end)
		req.Header.Set("Range", rangeHeader)
	} else if start > 0 {
		rangeHeader = fmt.Sprintf("bytes=%d-", start)
		req.Header.Set("Range", rangeHeader)
	}

	req.Header.Set("User-Agent", cw.userAgent)
	if cw.token != "" {
		req.Header.Set("Authorization", "Bearer "+cw.token)
	}

	resp, err := cw.client.Do(req)
	if err != nil {
		stallMu.Lock()
		wasStalled := stalled
		stallMu.Unlock()
		if wasStalled {
			return fmt.Errorf("connection stalled while connecting: no response for 30s")
		}
		return fmt.Errorf("range request failed (%s): %w", rangeHeader, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		if resp.StatusCode == http.StatusOK && start == 0 {
			// Single chunk from beginning or server ignored Range header
		} else {
			return fmt.Errorf("unexpected status code %d for %s (server may not support byte ranges)", resp.StatusCode, rangeHeader)
		}
	}

	var bodyReader io.Reader = resp.Body
	if resp.StatusCode == http.StatusOK && end >= start && end >= 0 {
		bodyReader = io.LimitReader(resp.Body, end-start+1)
	}

	if cw.limiter != nil {
		bodyReader = cw.limiter.NewReader(watchdogCtx, bodyReader)
	}

	buf := make([]byte, 128*1024) // 128KB buffer
	currOffset := start

	for {
		select {
		case <-watchdogCtx.Done():
			stallMu.Lock()
			wasStalled := stalled
			stallMu.Unlock()
			if wasStalled {
				return fmt.Errorf("connection stalled: no data received for 30s")
			}
			return ctx.Err()
		default:
		}

		n, rErr := bodyReader.Read(buf)
		if n > 0 {
			// Signal activity to watchdog
			select {
			case activityChan <- struct{}{}:
			default:
			}

			if _, wErr := writer.WriteAt(buf[:n], currOffset); wErr != nil {
				return fmt.Errorf("failed to write at offset %d: %w", currOffset, wErr)
			}
			currOffset += int64(n)
			if onBytesRead != nil {
				onBytesRead(n)
			}
		}

		if rErr != nil {
			if rErr == io.EOF {
				break
			}
			stallMu.Lock()
			wasStalled := stalled
			stallMu.Unlock()
			if wasStalled {
				return fmt.Errorf("connection stalled: no data received for 30s")
			}
			return fmt.Errorf("error streaming range data: %w", rErr)
		}
	}

	return nil
}
