//go:build windows

package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// CheckDiskSpace returns disk space information for the drive containing targetPath.
func CheckDiskSpace(targetPath string) (DiskSpaceInfo, error) {
	if targetPath == "" {
		targetPath = "."
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		absPath = targetPath
	}

	// Find the closest existing ancestor directory
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
		if parent == curr || parent == "" || parent == "." {
			vol := filepath.VolumeName(absPath)
			if vol != "" {
				curr = vol + `\`
			}
			break
		}
		curr = parent
	}

	// Ensure drive root like "C:" has trailing slash "C:\" for Windows API
	if len(curr) == 2 && curr[1] == ':' {
		curr += `\`
	} else if !strings.HasSuffix(curr, `\`) && !strings.HasSuffix(curr, `/`) {
		// It's safe to keep path as is or ensure clean directory
		curr = filepath.Clean(curr)
	}

	utf16Ptr, err := windows.UTF16PtrFromString(curr)
	if err != nil {
		return DiskSpaceInfo{Path: targetPath}, fmt.Errorf("invalid path format for %s: %w", curr, err)
	}

	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64
	err = windows.GetDiskFreeSpaceEx(utf16Ptr, &freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes)
	if err != nil {
		// Try fallback to drive root if path wasn't root
		vol := filepath.VolumeName(curr)
		if vol != "" {
			volRoot := vol + `\`
			if volPtr, volErr := windows.UTF16PtrFromString(volRoot); volErr == nil {
				if err2 := windows.GetDiskFreeSpaceEx(volPtr, &freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes); err2 == nil {
					return DiskSpaceInfo{
						Path:           volRoot,
						FreeBytes:      totalNumberOfFreeBytes,
						TotalBytes:     totalNumberOfBytes,
						AvailableBytes: freeBytesAvailable,
					}, nil
				}
			}
		}
		return DiskSpaceInfo{Path: targetPath}, fmt.Errorf("failed to query disk space for %s: %w", curr, err)
	}

	return DiskSpaceInfo{
		Path:           curr,
		FreeBytes:      totalNumberOfFreeBytes,
		TotalBytes:     totalNumberOfBytes,
		AvailableBytes: freeBytesAvailable,
	}, nil
}
