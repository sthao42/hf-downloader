package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// VerificationResult provides details on existing local file inspection.
type VerificationResult struct {
	Exists      bool   `json:"exists"`
	Valid       bool   `json:"valid"`
	ActualSize  int64  `json:"actualSize"`
	ActualSHA256 string `json:"actualSha256,omitempty"`
	Message     string `json:"message"`
}

// VerifyExistingFile inspects an existing file on disk against expected size and SHA-256.
func VerifyExistingFile(filePath, expectedHash string, expectedSize int64) (*VerificationResult, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &VerificationResult{
				Exists:  false,
				Valid:   false,
				Message: "file does not exist",
			}, nil
		}
		return nil, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	actualSize := info.Size()
	res := &VerificationResult{
		Exists:     true,
		ActualSize: actualSize,
	}

	if expectedSize > 0 && actualSize != expectedSize {
		res.Valid = false
		res.Message = fmt.Sprintf("size mismatch: found %d bytes, expected %d bytes", actualSize, expectedSize)
		return res, nil
	}

	if expectedHash == "" {
		// No hash provided to verify against, but size matches
		res.Valid = true
		res.Message = "file exists and size matches"
		return res, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for hashing: %w", err)
	}
	defer f.Close()

	hasher := sha256.New()
	buf := make([]byte, 1024*1024) // 1MB streaming buffer
	if _, err := io.CopyBuffer(hasher, f, buf); err != nil {
		return nil, fmt.Errorf("error hashing file %s: %w", filePath, err)
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	res.ActualSHA256 = actualHash

	if strings.EqualFold(actualHash, expectedHash) {
		res.Valid = true
		res.Message = "file verified: SHA-256 digest matched"
	} else {
		res.Valid = false
		res.Message = fmt.Sprintf("hash mismatch: expected %s, calculated %s", expectedHash, actualHash)
	}

	return res, nil
}

// FinalizeDownload verifies the completed .part file against expected SHA-256 and atomically promotes it to finalPath.
func FinalizeDownload(partPath, finalPath, expectedHash string) error {
	// If expectedHash is present, verify before promoting
	if expectedHash != "" {
		f, err := os.Open(partPath)
		if err != nil {
			return fmt.Errorf("failed to open part file for final verification: %w", err)
		}
		hasher := sha256.New()
		buf := make([]byte, 1024*1024)
		_, cErr := io.CopyBuffer(hasher, f, buf)
		f.Close()
		if cErr != nil {
			return fmt.Errorf("error computing hash for %s: %w", partPath, cErr)
		}

		actualHash := hex.EncodeToString(hasher.Sum(nil))
		if !strings.EqualFold(actualHash, expectedHash) {
			return fmt.Errorf("hash verification failed: expected %s, calculated %s", expectedHash, actualHash)
		}
	}

	// Remove target if exists (required on Windows before Rename)
	_ = os.Remove(finalPath)

	if err := os.Rename(partPath, finalPath); err != nil {
		return fmt.Errorf("failed to rename part file to final file: %w", err)
	}

	// Remove associated .part.json
	_ = os.Remove(partPath + ".json")
	return nil
}
