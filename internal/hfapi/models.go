package hfapi

// URLType indicates whether a target points to a full repository, subfolder, or file.
type URLType string

const (
	TargetRepo   URLType = "repo"
	TargetFolder URLType = "folder"
	TargetFile   URLType = "file"
)

// ParsedTarget holds the normalized decomposition of a Hugging Face target.
type ParsedTarget struct {
	RawInput string  `json:"rawInput"`
	RepoID   string  `json:"repoId"`
	Revision string  `json:"revision"` // defaults to "main"
	Type     URLType `json:"type"`
	Subpath  string  `json:"subpath"` // Empty if full repo; path to folder/file if specific
}

// LFSInfo holds Git-LFS metadata from the Hugging Face API.
type LFSInfo struct {
	OID         string `json:"oid"`
	Size        int64  `json:"size"`
	PointerSize int64  `json:"pointerSize,omitempty"`
	Sha256      string `json:"sha256,omitempty"`
}

// FileNode represents an item in the Hugging Face file tree.
type FileNode struct {
	Type        string   `json:"type"` // "file" or "directory"
	OID         string   `json:"oid"`
	Size        int64    `json:"size"`
	Path        string   `json:"path"`
	LFS         *LFSInfo `json:"lfs,omitempty"`
	SHA256      string   `json:"sha256,omitempty"`
	DownloadURL string   `json:"downloadUrl,omitempty"`
}

// InspectionResult holds HTTP headers and range support discovered for a file.
type InspectionResult struct {
	URL           string `json:"url"`
	ContentLength int64  `json:"contentLength"`
	AcceptRanges  bool   `json:"acceptRanges"`
	StatusCode    int    `json:"statusCode"`
	ETag          string `json:"etag,omitempty"`
	SHA256        string `json:"sha256,omitempty"`
	IsGated       bool   `json:"isGated"`
	FinalURL      string `json:"finalUrl,omitempty"`
}
