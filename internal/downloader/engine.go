package downloader

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// DownloadProgress captures live metrics of an in-progress transfer.
type DownloadProgress struct {
	TaskID          string  `json:"taskId"`
	Percentage      float64 `json:"percentage"`
	DownloadedBytes int64   `json:"downloadedBytes"`
	TotalBytes      int64   `json:"totalBytes"`
	SpeedBPS        float64 `json:"speedBps"`
	SpeedFormatted  string  `json:"speedFormatted"`
	ETASeconds      int64   `json:"etaSeconds"`
	Status          string  `json:"status"`
	ErrorMessage    string  `json:"errorMessage,omitempty"`
}

// DownloadJob defines all parameters necessary to download a target file.
type DownloadJob struct {
	TaskID         string `json:"taskId"`
	URL            string `json:"url"`
	DestinationDir string `json:"destinationDir"`
	FinalFilename  string `json:"finalFilename"`
	TotalSize      int64  `json:"totalSize"`
	ExpectedSHA256 string `json:"expectedSha256"`
	Token          string `json:"token"`
	MaxConnections int    `json:"maxConnections"`
	AcceptRanges   bool   `json:"acceptRanges"`
}

// FormatSpeed formats bytes per second into human readable units.
func FormatSpeed(bytesPerSec float64) string {
	if bytesPerSec >= 1024*1024*1024 {
		return fmt.Sprintf("%.2f GB/s", bytesPerSec/(1024*1024*1024))
	} else if bytesPerSec >= 1024*1024 {
		return fmt.Sprintf("%.2f MB/s", bytesPerSec/(1024*1024))
	} else if bytesPerSec >= 1024 {
		return fmt.Sprintf("%.2f KB/s", bytesPerSec/1024)
	}
	return fmt.Sprintf("%.0f B/s", bytesPerSec)
}

// FormatBytes formats byte counts into human readable strings.
func FormatBytes(bytes int64) string {
	b := float64(bytes)
	if b >= 1024*1024*1024 {
		return fmt.Sprintf("%.2f GB", b/(1024*1024*1024))
	} else if b >= 1024*1024 {
		return fmt.Sprintf("%.2f MB", b/(1024*1024))
	} else if b >= 1024 {
		return fmt.Sprintf("%.2f KB", b/1024)
	}
	return fmt.Sprintf("%d B", bytes)
}

// TransferEngine coordinates concurrent multi-socket chunk downloads.
type TransferEngine struct {
	httpClient *http.Client
}

// NewTransferEngine initializes a TransferEngine.
func NewTransferEngine(client *http.Client) *TransferEngine {
	if client == nil {
		client = &http.Client{
			Timeout: 60 * time.Second,
		}
	}
	return &TransferEngine{httpClient: client}
}

// Download executes a multi-connection chunked download or resumed transfer.
func (te *TransferEngine) Download(
	ctx context.Context,
	job *DownloadJob,
	progressCb func(progress DownloadProgress),
) error {
	if job == nil {
		return fmt.Errorf("download job is nil")
	}

	partPath := filepath.Join(job.DestinationDir, job.FinalFilename+".part")
	statePath := filepath.Join(job.DestinationDir, job.FinalFilename+".part.json")

	maxConn := job.MaxConnections
	if maxConn <= 0 {
		maxConn = 8
	}

	// 1. Check for existing state (.part.json) to resume
	var state *DownloadState
	if loaded, err := LoadState(statePath); err == nil && loaded.TotalSize == job.TotalSize && loaded.TotalSize > 0 {
		state = loaded
	} else {
		// Initialize chunks
		state = te.initChunks(job, maxConn, partPath)
	}

	// Calculate already downloaded bytes
	var downloadedBytes int64
	for _, ch := range state.Chunks {
		downloadedBytes += ch.Downloaded
	}

	// 2. Open part file writer (pre-allocating disk space)
	writer, err := NewPartFileWriter(partPath, job.TotalSize)
	if err != nil {
		return fmt.Errorf("failed to open part file writer: %w", err)
	}
	defer writer.Close()

	// 3. Setup speed sampler and progress ticker
	var liveDownloaded int64 = downloadedBytes
	var speedBPS float64 = 0
	var lastSampleTime = time.Now()
	var lastSampleBytes = downloadedBytes

	var stateMu sync.Mutex
	var tickerWg sync.WaitGroup
	stopTicker := make(chan struct{})
	var stopOnce sync.Once
	stopTickerFunc := func() {
		stopOnce.Do(func() {
			close(stopTicker)
			tickerWg.Wait()
		})
	}
	defer stopTickerFunc()

	tickerWg.Add(1)
	go func() {
		defer tickerWg.Done()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-stopTicker:
				return
			case <-ticker.C:
				now := time.Now()
				elapsed := now.Sub(lastSampleTime).Seconds()
				currentTotal := atomic.LoadInt64(&liveDownloaded)

				if elapsed >= 0.5 {
					delta := float64(currentTotal - lastSampleBytes)
					instSpeed := delta / elapsed
					if speedBPS == 0 {
						speedBPS = instSpeed
					} else {
						// 2-second moving average smoothing
						speedBPS = (speedBPS * 0.7) + (instSpeed * 0.3)
					}
					lastSampleTime = now
					lastSampleBytes = currentTotal
				}

				var percentage float64 = 0
				var eta int64 = 0
				if job.TotalSize > 0 {
					percentage = (float64(currentTotal) / float64(job.TotalSize)) * 100
					if speedBPS > 0 {
						remaining := job.TotalSize - currentTotal
						if remaining > 0 {
							eta = int64(float64(remaining) / speedBPS)
						}
					}
				}

				if progressCb != nil {
					progressCb(DownloadProgress{
						TaskID:          job.TaskID,
						Percentage:      percentage,
						DownloadedBytes: currentTotal,
						TotalBytes:      job.TotalSize,
						SpeedBPS:        speedBPS,
						SpeedFormatted:  FormatSpeed(speedBPS),
						ETASeconds:      eta,
						Status:          "downloading",
					})
				}

				// Periodic state flush
				stateMu.Lock()
				state.UpdatedEpoch = time.Now().Unix()
				_ = SaveState(statePath, state)
				stateMu.Unlock()
			}
		}
	}()

	// 4. Concurrently download chunks
	chunkCtx, chunkCancel := context.WithCancel(ctx)
	defer chunkCancel()

	worker := NewChunkWorker(te.httpClient, job.Token)
	errChan := make(chan error, len(state.Chunks))
	var wg sync.WaitGroup

	// Concurrency limiter channel
	sem := make(chan struct{}, maxConn)

	for i := range state.Chunks {
		chunk := &state.Chunks[i]
		if chunk.Done {
			continue
		}

		wg.Add(1)
		go func(c *ChunkState) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-chunkCtx.Done():
				errChan <- chunkCtx.Err()
				return
			}

			// Stream from current offset
			cErr := worker.DownloadRange(
				chunkCtx,
				job.URL,
				c.CurrentOffset,
				c.End,
				writer,
				func(n int) {
					atomic.AddInt64(&liveDownloaded, int64(n))
					stateMu.Lock()
					c.CurrentOffset += int64(n)
					c.Downloaded += int64(n)
					stateMu.Unlock()
				},
			)

			if cErr != nil {
				chunkCancel() // Fast cancel sibling chunks
				errChan <- cErr
				return
			}

			stateMu.Lock()
			c.Done = true
			stateMu.Unlock()
		}(chunk)
	}

	wg.Wait()
	close(errChan)

	// Check if any worker reported an error
	for err := range errChan {
		if err != nil {
			// Save current state before exiting
			stateMu.Lock()
			_ = SaveState(statePath, state)
			stateMu.Unlock()
			return err
		}
	}

	// Flush and commit writer
	if err := writer.Sync(); err != nil {
		return fmt.Errorf("failed to sync part file: %w", err)
	}

	// Stop speed sampling ticker before sending final completed status
	stopTickerFunc()

	// Final progress update
	if progressCb != nil {
		progressCb(DownloadProgress{
			TaskID:          job.TaskID,
			Percentage:      100.0,
			DownloadedBytes: job.TotalSize,
			TotalBytes:      job.TotalSize,
			SpeedBPS:        0,
			SpeedFormatted:  "",
			ETASeconds:      0,
			Status:          "completed",
		})
	}

	return nil
}

func (te *TransferEngine) initChunks(job *DownloadJob, maxConn int, partPath string) *DownloadState {
	state := &DownloadState{
		TaskID:         job.TaskID,
		URL:            job.URL,
		DestinationDir: job.DestinationDir,
		FinalFilename:  job.FinalFilename,
		PartPath:       partPath,
		TotalSize:      job.TotalSize,
		ExpectedSHA256: job.ExpectedSHA256,
		UpdatedEpoch:   time.Now().Unix(),
	}

	if job.TotalSize <= 0 || !job.AcceptRanges {
		// Single chunk if range not supported or size unknown
		state.Chunks = []ChunkState{
			{
				Index:         0,
				Start:         0,
				End:           job.TotalSize - 1,
				CurrentOffset: 0,
				Downloaded:    0,
				Done:          false,
			},
		}
		return state
	}

	// Determine number of chunks (minimum 4MB per chunk)
	const minChunkSize = 4 * 1024 * 1024
	numChunks := int(job.TotalSize / minChunkSize)
	if numChunks < 1 {
		numChunks = 1
	}
	if numChunks > maxConn {
		numChunks = maxConn
	}

	chunkSize := job.TotalSize / int64(numChunks)
	chunks := make([]ChunkState, numChunks)

	for i := 0; i < numChunks; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == numChunks-1 {
			end = job.TotalSize - 1
		}
		chunks[i] = ChunkState{
			Index:         i,
			Start:         start,
			End:           end,
			CurrentOffset: start,
			Downloaded:    0,
			Done:          false,
		}
	}

	state.Chunks = chunks
	return state
}
