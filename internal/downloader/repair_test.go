package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyExistingFile_And_Finalize(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "hf_repair_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	data := []byte("testing chunk repair and verification payload")
	hasher := sha256.New()
	hasher.Write(data)
	correctHash := hex.EncodeToString(hasher.Sum(nil))

	partPath := filepath.Join(tempDir, "model.safetensors.part")
	finalPath := filepath.Join(tempDir, "model.safetensors")

	// 1. Test non-existent file
	res, err := VerifyExistingFile(finalPath, correctHash, int64(len(data)))
	if err != nil {
		t.Fatalf("VerifyExistingFile failed: %v", err)
	}
	if res.Exists {
		t.Errorf("Expected file to not exist")
	}

	// 2. Write part file
	if err := os.WriteFile(partPath, data, 0644); err != nil {
		t.Fatalf("Failed to write part file: %v", err)
	}

	// 3. Test Finalize with mismatched hash
	err = FinalizeDownload(partPath, finalPath, "wronghash123")
	if err == nil {
		t.Errorf("Expected finalize to fail on hash mismatch")
	}

	// 4. Test Finalize with correct hash
	err = FinalizeDownload(partPath, finalPath, correctHash)
	if err != nil {
		t.Fatalf("FinalizeDownload failed: %v", err)
	}

	// 5. Test VerifyExistingFile on promoted final file
	finalRes, err := VerifyExistingFile(finalPath, correctHash, int64(len(data)))
	if err != nil {
		t.Fatalf("VerifyExistingFile on final failed: %v", err)
	}
	if !finalRes.Exists || !finalRes.Valid {
		t.Errorf("Expected final file to exist and be valid: %+v", finalRes)
	}
	if finalRes.ActualSHA256 != correctHash {
		t.Errorf("Expected ActualSHA256 %s, got %s", correctHash, finalRes.ActualSHA256)
	}
	if finalRes.ExpectedSHA256 != correctHash {
		t.Errorf("Expected ExpectedSHA256 %s, got %s", correctHash, finalRes.ExpectedSHA256)
	}

	// 6. Test VerifyExistingFile without expected hash (should still calculate ActualSHA256)
	noHashRes, err := VerifyExistingFile(finalPath, "", int64(len(data)))
	if err != nil {
		t.Fatalf("VerifyExistingFile without hash failed: %v", err)
	}
	if !noHashRes.Exists || !noHashRes.Valid {
		t.Errorf("Expected file without expected hash to be valid: %+v", noHashRes)
	}
	if noHashRes.ActualSHA256 != correctHash {
		t.Errorf("Expected ActualSHA256 to be computed as %s even without expected hash, got %s", correctHash, noHashRes.ActualSHA256)
	}
	if noHashRes.ExpectedSHA256 != "" {
		t.Errorf("Expected empty ExpectedSHA256, got %s", noHashRes.ExpectedSHA256)
	}
}
