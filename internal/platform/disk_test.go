package platform

import (
	"os"
	"testing"
)

func TestCheckDiskSpace(t *testing.T) {
	// Test current directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	info, err := CheckDiskSpace(cwd)
	if err != nil {
		t.Fatalf("CheckDiskSpace failed for cwd (%s): %v", cwd, err)
	}

	if info.TotalBytes == 0 {
		t.Errorf("expected TotalBytes > 0, got 0")
	}

	if info.AvailableBytes == 0 {
		t.Errorf("expected AvailableBytes > 0, got 0")
	}

	t.Logf("CWD: %s | Total: %d GB | Available: %d GB",
		info.Path,
		info.TotalBytes/(1024*1024*1024),
		info.AvailableBytes/(1024*1024*1024),
	)

	// Test non-existent subfolder
	subPath := cwd + string(os.PathSeparator) + "non_existent_folder_xyz_123"
	subInfo, err := CheckDiskSpace(subPath)
	if err != nil {
		t.Fatalf("CheckDiskSpace failed for non-existent path: %v", err)
	}

	if subInfo.AvailableBytes == 0 {
		t.Errorf("expected AvailableBytes > 0 for non-existent subpath")
	}
}
