//go:build !windows

package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"golang.org/x/sys/unix"
)

// CheckDiskSpace returns disk space information for the filesystem containing targetPath.
func CheckDiskSpace(targetPath string) (DiskSpaceInfo, error) {
	if targetPath == "" {
		targetPath = "."
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		absPath = targetPath
	}

	curr := absPath
	for {
		fi, statErr := os.Stat(curr)
		if statErr == nil {
			if !fi.IsDir() {
				curr = filepath.Dir(curr)
			}
			break
		}

		parent := filepath.Dir(curr)
		if parent == curr || parent == "" || parent == "." || parent == "/" {
			curr = "/"
			break
		}
		curr = parent
	}

	var stat unix.Statfs_t
	err = unix.Statfs(curr, &stat)
	if err != nil {
		return DiskSpaceInfo{Path: targetPath}, fmt.Errorf("failed to query disk space for %s: %w", curr, err)
	}

	availableBytes := stat.Bavail * uint64(stat.Bsize)
	freeBytes := stat.Bfree * uint64(stat.Bsize)
	totalBytes := stat.Blocks * uint64(stat.Bsize)

	return DiskSpaceInfo{
		Path:           curr,
		FreeBytes:      freeBytes,
		TotalBytes:     totalBytes,
		AvailableBytes: availableBytes,
	}, nil
}
