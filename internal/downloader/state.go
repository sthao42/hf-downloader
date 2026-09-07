package downloader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ChunkState represents the download status of an individual byte range.
type ChunkState struct {
	Index         int   `json:"index"`
	Start         int64 `json:"start"`
	End           int64 `json:"end"`
	CurrentOffset int64 `json:"currentOffset"`
	Downloaded    int64 `json:"downloaded"`
	Done          bool  `json:"done"`
}

// DownloadState holds full recovery state for an in-flight file download.
type DownloadState struct {
	TaskID         string       `json:"taskId"`
	URL            string       `json:"url"`
	DestinationDir string       `json:"destinationDir"`
	FinalFilename  string       `json:"finalFilename"`
	PartPath       string       `json:"partPath"`
	TotalSize      int64        `json:"totalSize"`
	ExpectedSHA256 string       `json:"expectedSha256"`
	Chunks         []ChunkState `json:"chunks"`
	UpdatedEpoch   int64        `json:"updatedEpoch"`
}

// StateManager provides synchronized atomic persistence for .part.json files.
type StateManager struct {
	mu sync.Mutex
}

var globalStateManager = &StateManager{}

// SaveState atomically serializes the state to stateFilePath (.part.json).
func SaveState(stateFilePath string, state *DownloadState) error {
	globalStateManager.mu.Lock()
	defer globalStateManager.mu.Unlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	dir := filepath.Dir(stateFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	tempFile := fmt.Sprintf("%s.tmp.%d", stateFilePath, os.Getpid())
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp state file: %w", err)
	}

	if err := os.Rename(tempFile, stateFilePath); err != nil {
		// On Windows, if destination exists, rename might fail if not removed first
		_ = os.Remove(stateFilePath)
		if retryErr := os.Rename(tempFile, stateFilePath); retryErr != nil {
			_ = os.Remove(tempFile)
			return fmt.Errorf("failed to atomically rename state file: %w", retryErr)
		}
	}

	return nil
}

// LoadState reads and parses the .part.json file if it exists.
func LoadState(stateFilePath string) (*DownloadState, error) {
	globalStateManager.mu.Lock()
	defer globalStateManager.mu.Unlock()

	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		return nil, err
	}

	var state DownloadState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

// RemoveState removes the .part.json state file upon completion.
func RemoveState(stateFilePath string) error {
	globalStateManager.mu.Lock()
	defer globalStateManager.mu.Unlock()

	if err := os.Remove(stateFilePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
