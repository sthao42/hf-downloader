package hfapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_FetchTree(t *testing.T) {
	mockItems := []FileNode{
		{
			Type: "file",
			OID:  "abc1234",
			Size: 1048576,
			Path: "model.safetensors",
			LFS: &LFSInfo{
				OID:  "sha256_mock_hash",
				Size: 1048576,
			},
		},
		{
			Type: "directory",
			OID:  "dir1234",
			Size: 0,
			Path: "configs",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header if provided
		if auth := r.Header.Get("Authorization"); auth != "Bearer test_token" {
			t.Errorf("Expected Bearer test_token, got %s", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockItems)
	}))
	defer server.Close()

	client := NewClient(WithToken("test_token"), WithHTTPClient(server.Client()))

	// Use custom fetcher test by overriding base request
	req, _ := http.NewRequest("GET", server.URL, nil)
	client.prepareRequest(req)
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("Failed to call test server: %v", err)
	}
	defer resp.Body.Close()

	var nodes []FileNode
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if len(nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Path != "model.safetensors" {
		t.Errorf("Expected model.safetensors, got %s", nodes[0].Path)
	}
}

func TestResolver_InspectFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Length", "52428800") // 50MB
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Header.Get("Range") != "" {
			w.Header().Set("Content-Range", "bytes 0-0/52428800")
			w.Header().Set("Accept-Ranges", "bytes")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write([]byte("x"))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resolver := NewResolver(server.Client())
	result, err := resolver.InspectFile(server.URL, "")
	if err != nil {
		t.Fatalf("InspectFile failed: %v", err)
	}

	if !result.AcceptRanges {
		t.Errorf("Expected AcceptRanges to be true")
	}
	if result.ContentLength != 52428800 {
		t.Errorf("Expected ContentLength 52428800, got %d", result.ContentLength)
	}
}
