package hfapi

import (
	"testing"
)

func TestParseHFURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		wantRepo    string
		wantRev     string
		wantType    URLType
		wantSubpath string
	}{
		{
			name:        "Full repo URL with https",
			input:       "https://huggingface.co/black-forest-labs/FLUX.1-dev",
			wantErr:     false,
			wantRepo:    "black-forest-labs/FLUX.1-dev",
			wantRev:     "main",
			wantType:    TargetRepo,
			wantSubpath: "",
		},
		{
			name:        "Full repo URL without protocol",
			input:       "huggingface.co/black-forest-labs/FLUX.1-dev/",
			wantErr:     false,
			wantRepo:    "black-forest-labs/FLUX.1-dev",
			wantRev:     "main",
			wantType:    TargetRepo,
			wantSubpath: "",
		},
		{
			name:        "Short repo identifier",
			input:       "meta-llama/Llama-3.1-8B-Instruct",
			wantErr:     false,
			wantRepo:    "meta-llama/Llama-3.1-8B-Instruct",
			wantRev:     "main",
			wantType:    TargetRepo,
			wantSubpath: "",
		},
		{
			name:        "Single file blob URL",
			input:       "https://huggingface.co/black-forest-labs/FLUX.1-dev/blob/main/vae/diffusion_pytorch_model.safetensors",
			wantErr:     false,
			wantRepo:    "black-forest-labs/FLUX.1-dev",
			wantRev:     "main",
			wantType:    TargetFile,
			wantSubpath: "vae/diffusion_pytorch_model.safetensors",
		},
		{
			name:        "Single file resolve URL with query params",
			input:       "https://huggingface.co/black-forest-labs/FLUX.1-dev/resolve/main/text_encoders/t5xxl_fp16.safetensors?download=true",
			wantErr:     false,
			wantRepo:    "black-forest-labs/FLUX.1-dev",
			wantRev:     "main",
			wantType:    TargetFile,
			wantSubpath: "text_encoders/t5xxl_fp16.safetensors",
		},
		{
			name:        "Folder tree URL",
			input:       "https://huggingface.co/Comfy-Org/flux1-dev/tree/main/split_files/text_encoders",
			wantErr:     false,
			wantRepo:    "Comfy-Org/flux1-dev",
			wantRev:     "main",
			wantType:    TargetFolder,
			wantSubpath: "split_files/text_encoders",
		},
		{
			name:        "Repo tree root URL",
			input:       "https://huggingface.co/Comfy-Org/flux1-dev/tree/main",
			wantErr:     false,
			wantRepo:    "Comfy-Org/flux1-dev",
			wantRev:     "main",
			wantType:    TargetRepo,
			wantSubpath: "",
		},
		{
			name:        "Custom branch revision in blob URL",
			input:       "https://huggingface.co/user/model/blob/v1.2/weights.bin",
			wantErr:     false,
			wantRepo:    "user/model",
			wantRev:     "v1.2",
			wantType:    TargetFile,
			wantSubpath: "weights.bin",
		},
		{
			name:     "Empty input",
			input:    "   ",
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseHFURL(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseHFURL(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got.RepoID != tc.wantRepo {
				t.Errorf("RepoID = %v, want %v", got.RepoID, tc.wantRepo)
			}
			if got.Revision != tc.wantRev {
				t.Errorf("Revision = %v, want %v", got.Revision, tc.wantRev)
			}
			if got.Type != tc.wantType {
				t.Errorf("Type = %v, want %v", got.Type, tc.wantType)
			}
			if got.Subpath != tc.wantSubpath {
				t.Errorf("Subpath = %v, want %v", got.Subpath, tc.wantSubpath)
			}
		})
	}
}

func TestBuildDownloadURL(t *testing.T) {
	url := BuildDownloadURL("owner/repo", "main", "vae/model.safetensors")
	expected := "https://huggingface.co/owner/repo/resolve/main/vae/model.safetensors"
	if url != expected {
		t.Errorf("BuildDownloadURL() = %v, want %v", url, expected)
	}
}
