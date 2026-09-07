package downloader

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestTransferEngine_Download(t *testing.T) {
	// Generate 10MB deterministic test payload
	const payloadSize = 10 * 1024 * 1024
	payload := make([]byte, payloadSize)
	for i := range payload {
		payload[i] = byte(i % 251)
	}

	hasher := sha256.New()
	hasher.Write(payload)
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	// Local HTTP test server supporting Range requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept-Ranges", "bytes")

		rangeHeader := r.Header.Get("Range")
		if rangeHeader == "" {
			w.Header().Set("Content-Length", strconv.Itoa(payloadSize))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(payload)
			return
		}

		// Parse Range: bytes=start-end
		if !strings.HasPrefix(rangeHeader, "bytes=") {
			http.Error(w, "invalid range", http.StatusBadRequest)
			return
		}

		parts := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
		start, _ := strconv.ParseInt(parts[0], 10, 64)
		end := int64(payloadSize - 1)
		if len(parts) > 1 && parts[1] != "" {
			end, _ = strconv.ParseInt(parts[1], 10, 64)
		}
		if end >= int64(payloadSize) {
			end = int64(payloadSize - 1)
		}

		contentLength := end - start + 1
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, payloadSize))
		w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(payload[start : end+1])
	}))
	defer server.Close()

	// Temporary output dir
	tempDir, err := os.MkdirTemp("", "hf_down_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	engine := NewTransferEngine(server.Client())
	job := &DownloadJob{
		TaskID:         "test-job-1",
		URL:            server.URL,
		DestinationDir: tempDir,
		FinalFilename:  "test_model.safetensors",
		TotalSize:      int64(payloadSize),
		ExpectedSHA256: expectedHash,
		MaxConnections: 4,
		AcceptRanges:   true,
	}

	var lastReportedBytes int64
	err = engine.Download(context.Background(), job, func(p DownloadProgress) {
		lastReportedBytes = p.DownloadedBytes
	})
	if err != nil {
		t.Fatalf("Engine download failed: %v", err)
	}
	if lastReportedBytes <= 0 {
		t.Logf("Warning: progress callback last reported bytes: %d", lastReportedBytes)
	}

	// Verify downloaded .part file
	partPath := filepath.Join(tempDir, "test_model.safetensors.part")
	downloadedData, err := os.ReadFile(partPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded part file: %v", err)
	}

	if len(downloadedData) != payloadSize {
		t.Fatalf("Downloaded data size mismatch: got %d, want %d", len(downloadedData), payloadSize)
	}

	if !bytes.Equal(downloadedData, payload) {
		t.Fatalf("Downloaded data content does not match original payload bit-for-bit")
	}

	downHasher := sha256.New()
	downHasher.Write(downloadedData)
	downloadedHash := hex.EncodeToString(downHasher.Sum(nil))

	if downloadedHash != expectedHash {
		t.Fatalf("Hash mismatch: got %s, want %s", downloadedHash, expectedHash)
	}
}

func TestStatePersistence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "hf_state_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	statePath := filepath.Join(tempDir, "model.part.json")
	state := &DownloadState{
		TaskID:         "task-123",
		URL:            "https://example.com/test",
		DestinationDir: tempDir,
		FinalFilename:  "model.bin",
		TotalSize:      1000,
		Chunks: []ChunkState{
			{Index: 0, Start: 0, End: 499, CurrentOffset: 500, Downloaded: 500, Done: true},
			{Index: 1, Start: 500, End: 999, CurrentOffset: 700, Downloaded: 200, Done: false},
		},
	}

	if err := SaveState(statePath, state); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	loaded, err := LoadState(statePath)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if loaded.TaskID != "task-123" || len(loaded.Chunks) != 2 || loaded.Chunks[0].Done != true {
		t.Fatalf("Loaded state does not match saved state: %+v", loaded)
	}

	if err := RemoveState(statePath); err != nil {
		t.Fatalf("RemoveState failed: %v", err)
	}

	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Fatalf("Expected state file to be removed")
	}
}
