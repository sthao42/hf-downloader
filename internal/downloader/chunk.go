package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ChunkWorker downloads a specific byte range of a file.
type ChunkWorker struct {
	client    *http.Client
	token     string
	userAgent string
}

// NewChunkWorker constructs a new range worker.
func NewChunkWorker(client *http.Client, token string) *ChunkWorker {
	if client == nil {
		client = &http.Client{
			Timeout: 60 * time.Second,
		}
	}
	return &ChunkWorker{
		client:    client,
		token:     token,
		userAgent: "hf-downloader/1.0",
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

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	buf := make([]byte, 128*1024) // 128KB buffer
	currOffset := start

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, rErr := bodyReader.Read(buf)
		if n > 0 {
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
			return fmt.Errorf("error streaming range data: %w", rErr)
		}
	}

	return nil
}
