package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "hf_cfg_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfgPath := filepath.Join(tempDir, "config.json")
	cm := &ConfigManager{
		filePath: cfgPath,
		settings: DefaultSettings(),
	}

	// 1. Test Save and Load
	if err := cm.Save(); err != nil {
		t.Fatalf("cm.Save() failed: %v", err)
	}

	cmLoaded := &ConfigManager{
		filePath: cfgPath,
	}
	if err := cmLoaded.Load(); err != nil {
		t.Fatalf("cmLoaded.Load() failed: %v", err)
	}

	if cmLoaded.settings.DefaultDownloadDir != cm.settings.DefaultDownloadDir {
		t.Errorf("DefaultDownloadDir mismatch: got %s, want %s", cmLoaded.settings.DefaultDownloadDir, cm.settings.DefaultDownloadDir)
	}

	// 2. Test Add and Delete Bookmark
	bm, err := cm.AddBookmark("Test Checkpoints", "D:\\test\\checkpoints")
	if err != nil {
		t.Fatalf("AddBookmark failed: %v", err)
	}
	if bm.ID == "" || bm.Label != "Test Checkpoints" {
		t.Errorf("Unexpected bookmark: %+v", bm)
	}

	if err := cm.DeleteBookmark(bm.ID); err != nil {
		t.Fatalf("DeleteBookmark failed: %v", err)
	}

	// 3. Test Pattern Resolution
	vaeDest := cm.ResolveDestination("flux1-dev-vae.safetensors")
	if !filepath.IsAbs(vaeDest) || filepath.Base(vaeDest) != "vae" {
		t.Errorf("Expected vae destination, got: %s", vaeDest)
	}

	clipDest := cm.ResolveDestination("text_encoders/t5xxl_fp16.safetensors")
	if !filepath.IsAbs(clipDest) || filepath.Base(clipDest) != "text_encoders" {
		t.Errorf("Expected text_encoders destination, got: %s", clipDest)
	}

	// 4. Test UpdateBookmarks reordering
	b1, _ := cm.AddBookmark("Folder A", "D:\\test\\a")
	b2, _ := cm.AddBookmark("Folder B", "D:\\test\\b")
	reordered := []FolderBookmark{b2, b1}
	if err := cm.UpdateBookmarks(reordered); err != nil {
		t.Fatalf("UpdateBookmarks failed: %v", err)
	}
	currentBMs := cm.GetSettings().Bookmarks
	if len(currentBMs) < 2 || currentBMs[0].ID != b2.ID {
		t.Errorf("Expected reordered bookmarks to have %s first, got %v", b2.ID, currentBMs)
	}

	// 5. Test UpdateRecentPaths reordering
	recentList := []string{"D:\\path1", "D:\\path2"}
	if err := cm.UpdateRecentPaths(recentList); err != nil {
		t.Fatalf("UpdateRecentPaths failed: %v", err)
	}
	if len(cm.GetSettings().RecentPaths) != 2 || cm.GetSettings().RecentPaths[0] != "D:\\path1" {
		t.Errorf("Unexpected RecentPaths: %v", cm.GetSettings().RecentPaths)
	}
}
