package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
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
	return w.file.WriteAt(p, off)
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
