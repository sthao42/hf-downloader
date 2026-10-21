package downloader

import (
	"os"
	"path/filepath"
	"testing"
	"hf-downloader/internal/config"
)

func TestQueue_PathTraversalSanitization(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "queue_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cm := config.GetManager()
	qm := NewQueueManager(cm)
	defer qm.Close()

	items := []DownloadItem{
		{
			RemotePath:     "weights/model.safetensors",
			DestinationDir: tempDir,
			FinalFilename:  "../../evil.exe",
		},
		{
			RemotePath:     "sub/deep/lora.safetensors",
			DestinationDir: tempDir,
			FinalFilename:  `..\..\windows\system32\calc.exe`,
		},
		{
			RemotePath:     "config.json",
			DestinationDir: filepath.Join(tempDir, "subdir/../"),
			FinalFilename:  "",
		},
	}

	added := qm.QueueItems(items, false)
	if len(added) != 3 {
		t.Fatalf("expected 3 items, got %d", len(added))
	}

	// First item: ../../evil.exe must be sanitized to evil.exe
	if added[0].FinalFilename != "evil.exe" {
		t.Errorf("expected 'evil.exe', got %q", added[0].FinalFilename)
	}

	// Second item: ..\..\windows\system32\calc.exe must be sanitized to calc.exe
	if added[1].FinalFilename != "calc.exe" {
		t.Errorf("expected 'calc.exe', got %q", added[1].FinalFilename)
	}

	// Third item: FinalFilename was empty, should fallback to filepath.Base(RemotePath) -> config.json
	if added[2].FinalFilename != "config.json" {
		t.Errorf("expected 'config.json', got %q", added[2].FinalFilename)
	}

	// Third item: DestinationDir must be cleaned
	if added[2].DestinationDir != filepath.Clean(tempDir) {
		t.Errorf("expected cleaned %q, got %q", filepath.Clean(tempDir), added[2].DestinationDir)
	}
}

func TestQueue_CloseIdempotency(t *testing.T) {
	cm := config.GetManager()
	qm := NewQueueManager(cm)

	// Calling Close multiple times should be safe and not panic
	qm.Close()
	qm.Close()
	qm.Close()
}

func TestQueue_SpeedLimit(t *testing.T) {
	cm := config.GetManager()
	qm := NewQueueManager(cm)
	defer qm.Close()

	initialLimit := qm.GetSpeedLimit()
	defer qm.SetSpeedLimit(initialLimit)

	qm.SetSpeedLimit(25.0)
	if qm.GetSpeedLimit() != 25.0 {
		t.Errorf("expected speed limit 25.0, got %f", qm.GetSpeedLimit())
	}
	if cm.GetSettings().SpeedLimitMBps != 25.0 {
		t.Errorf("expected config speed limit 25.0, got %f", cm.GetSettings().SpeedLimitMBps)
	}

	qm.SetSpeedLimit(0)
	if qm.GetSpeedLimit() != 0 {
		t.Errorf("expected speed limit 0, got %f", qm.GetSpeedLimit())
	}
	if cm.GetSettings().SpeedLimitMBps != 0 {
		t.Errorf("expected config speed limit 0, got %f", cm.GetSettings().SpeedLimitMBps)
	}
}

