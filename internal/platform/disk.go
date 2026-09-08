package platform

// DiskSpaceInfo contains volume space statistics in bytes.
type DiskSpaceInfo struct {
	Path           string `json:"path"`
	FreeBytes      uint64 `json:"freeBytes"`
	TotalBytes     uint64 `json:"totalBytes"`
	AvailableBytes uint64 `json:"availableBytes"`
}
