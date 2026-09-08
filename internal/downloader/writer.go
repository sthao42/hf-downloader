package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"hf-downloader/internal/platform"
)

// PartFileWriter provides safe concurrent WriteAt operations on a partially downloaded file.
type PartFileWriter struct {
	mu       sync.Mutex
	file     *os.File
	path     string
	filesize int64
}

// NewPartFileWriter opens or creates the target .part file and pre-allocates disk space if size > 0.
func NewPartFileWriter(partPath string, totalSize int64) (*PartFileWriter, error) {
	dir := filepath.Dir(partPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination folder %s: %w", dir, err)
	}

	// Safety check: verify storage drive has enough space to prevent mid-download stalls or system errors
	if totalSize > 0 {
		if spaceInfo, spaceErr := platform.CheckDiskSpace(dir); spaceErr == nil {
			var currentSize int64
			if fi, statErr := os.Stat(partPath); statErr == nil {
				currentSize = fi.Size()
			}
			needed := totalSize - currentSize
			const safetyBuffer = 50 * 1024 * 1024 // 50MB reserve buffer
			if needed > 0 && spaceInfo.AvailableBytes < uint64(needed+safetyBuffer) {
				return nil, fmt.Errorf("insufficient disk space on %s: need %s, only %s available",
					spaceInfo.Path, FormatBytes(needed), FormatBytes(int64(spaceInfo.AvailableBytes)))
			}
		}
	}

	f, err := os.OpenFile(partPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open/create part file %s: %w", partPath, err)
	}

	if totalSize > 0 {
		stat, statErr := f.Stat()
		if statErr == nil && stat.Size() < totalSize {
			if truncErr := f.Truncate(totalSize); truncErr != nil {
				f.Close()
				return nil, fmt.Errorf("failed to pre-allocate disk space (%d bytes): %w", totalSize, truncErr)
			}
		}
	}

	return &PartFileWriter{
		file:     f,
		path:     partPath,
		filesize: totalSize,
	}, nil
}

// WriteAt writes p to the underlying file at offset off in a thread-safe manner.
func (w *PartFileWriter) WriteAt(p []byte, off int64) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return 0, fmt.Errorf("part file is closed")
	}
	n, err := w.file.WriteAt(p, off)
	if err != nil {
		return n, fmt.Errorf("write error (possible disk full or I/O failure): %w", err)
	}
	return n, nil
}

// Sync commits the current contents of the file to stable storage.
func (w *PartFileWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}

// Close flushes and closes the underlying file.
func (w *PartFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		_ = w.file.Sync()
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}

// Path returns the path of the part file.
func (w *PartFileWriter) Path() string {
	return w.path
}
