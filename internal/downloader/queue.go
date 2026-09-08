package downloader

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"hf-downloader/internal/config"
	"hf-downloader/internal/hfapi"
	"hf-downloader/internal/platform"
)

// ItemStatus defines the lifecycle status of a download task.
type ItemStatus string

const (
	StatusStaged      ItemStatus = "staged"
	StatusQueued      ItemStatus = "queued"
	StatusVerifying   ItemStatus = "verifying"
	StatusDownloading ItemStatus = "downloading"
	StatusPaused      ItemStatus = "paused"
	StatusCompleted   ItemStatus = "completed"
	StatusFailed      ItemStatus = "failed"
)

// DownloadItem represents a staged or active download task.
type DownloadItem struct {
	ID              string     `json:"id"`
	RepoID          string     `json:"repoId"`
	Revision        string     `json:"revision"`
	RemotePath      string     `json:"remotePath"`
	DestinationDir  string     `json:"destinationDir"`
	FinalFilename   string     `json:"finalFilename"`
	Size            int64      `json:"size"`
	ExpectedSHA256  string     `json:"expectedSha256"`
	Status          ItemStatus `json:"status"`
	AutoStart       bool       `json:"autoStart"`
	DownloadedBytes int64      `json:"downloadedBytes"`
	Progress        float64    `json:"progress"`
	SpeedBPS        float64    `json:"speedBps"`
	SpeedFormatted  string     `json:"speedFormatted"`
	ETASeconds      int64      `json:"etaSeconds"`
	ErrorMessage    string     `json:"errorMessage,omitempty"`
	CreatedAt       int64      `json:"createdAt"`
}

// QueueManager manages staged, running, and completed download items.
type QueueManager struct {
	mu           sync.RWMutex
	items        []DownloadItem
	cancels      map[string]context.CancelFunc
	engine       *TransferEngine
	resolver     *hfapi.Resolver
	configMgr    *config.ConfigManager
	progressSubs []func(item DownloadItem)
	sem          chan struct{}
	autoDownload bool
	stopChan     chan struct{}
}

// NewQueueManager initializes the QueueManager.
func NewQueueManager(cm *config.ConfigManager) *QueueManager {
	if cm == nil {
		cm = config.GetManager()
	}
	settings := cm.GetSettings()
	maxConcurrent := settings.MaxConcurrentFiles
	if maxConcurrent <= 0 {
		maxConcurrent = 2
	}

	httpClient := &http.Client{Timeout: 60 * time.Second}
	qm := &QueueManager{
		items:        make([]DownloadItem, 0),
		cancels:      make(map[string]context.CancelFunc),
		engine:       NewTransferEngine(httpClient),
		resolver:     hfapi.NewResolver(httpClient),
		configMgr:    cm,
		progressSubs: make([]func(item DownloadItem), 0),
		sem:          make(chan struct{}, maxConcurrent),
		autoDownload: settings.AutoDownload,
		stopChan:     make(chan struct{}),
	}

	go qm.processLoop()
	return qm
}

// OnProgress registers a callback for item updates.
func (qm *QueueManager) OnProgress(fn func(item DownloadItem)) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.progressSubs = append(qm.progressSubs, fn)
}

func (qm *QueueManager) broadcast(item DownloadItem) {
	qm.mu.RLock()
	subs := append([]func(item DownloadItem){}, qm.progressSubs...)
	qm.mu.RUnlock()

	for _, sub := range subs {
		sub(item)
	}
}

// SetAutoDownload updates the auto-download toggle.
func (qm *QueueManager) SetAutoDownload(enabled bool) {
	qm.mu.Lock()
	qm.autoDownload = enabled
	s := qm.configMgr.GetSettings()
	s.AutoDownload = enabled
	_ = qm.configMgr.UpdateSettings(s)
	qm.mu.Unlock()
}

// IsAutoDownload returns current auto-download setting.
func (qm *QueueManager) IsAutoDownload() bool {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	return qm.autoDownload
}

// QueueItems adds one or more items to the staging queue.
func (qm *QueueManager) QueueItems(newItems []DownloadItem, autoStart bool) []DownloadItem {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	var added []DownloadItem
	for _, item := range newItems {
		if item.ID == "" {
			item.ID = "task-" + uuid.New().String()[:8]
		}
		if item.Revision == "" {
			item.Revision = "main"
		}
		if item.FinalFilename == "" {
			item.FinalFilename = filepath.Base(item.RemotePath)
		}
		if item.DestinationDir == "" {
			item.DestinationDir = qm.configMgr.ResolveDestination(item.RemotePath)
		}
		if item.CreatedAt == 0 {
			item.CreatedAt = time.Now().Unix()
		}

		if autoStart || qm.autoDownload {
			item.Status = StatusQueued
			item.AutoStart = true
		} else {
			item.Status = StatusStaged
			item.AutoStart = false
		}

		qm.items = append(qm.items, item)
		added = append(added, item)
		go qm.broadcast(item)
	}

	return added
}

// UpdateItemDestination modifies the destination directory for a staged or paused item.
func (qm *QueueManager) UpdateItemDestination(id, newDestDir string) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	for i := range qm.items {
		if qm.items[i].ID == id {
			if qm.items[i].Status == StatusDownloading {
				return fmt.Errorf("cannot change destination directory while item is downloading")
			}
			qm.items[i].DestinationDir = newDestDir
			go qm.broadcast(qm.items[i])
			return nil
		}
	}
	return fmt.Errorf("item %s not found", id)
}

// StartItem transitions a staged or paused item to queued status.
func (qm *QueueManager) StartItem(id string) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	for i := range qm.items {
		if qm.items[i].ID == id {
			if qm.items[i].Status == StatusDownloading {
				return nil
			}
			qm.items[i].Status = StatusQueued
			qm.items[i].ErrorMessage = ""
			go qm.broadcast(qm.items[i])
			return nil
		}
	}
	return fmt.Errorf("item %s not found", id)
}

// UpdateItemExpectedSHA256 sets or backfills the expected SHA-256 hash for a specific item.
func (qm *QueueManager) UpdateItemExpectedSHA256(id, hash string) {
	if hash == "" {
		return
	}
	qm.updateItemStatus(id, func(it *DownloadItem) {
		it.ExpectedSHA256 = hash
	})
}

// PauseItem pauses an in-progress or queued download.
func (qm *QueueManager) PauseItem(id string) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if cancel, exists := qm.cancels[id]; exists {
		cancel()
		delete(qm.cancels, id)
	}

	for i := range qm.items {
		if qm.items[i].ID == id {
			if qm.items[i].Status == StatusDownloading || qm.items[i].Status == StatusQueued {
				qm.items[i].Status = StatusPaused
				qm.items[i].SpeedBPS = 0
				qm.items[i].SpeedFormatted = ""
				go qm.broadcast(qm.items[i])
			}
			return nil
		}
	}
	return fmt.Errorf("item %s not found", id)
}

// ResumeItem resumes a paused item.
func (qm *QueueManager) ResumeItem(id string) error {
	return qm.StartItem(id)
}

// CancelItem cancels an active item and resets it to staged.
func (qm *QueueManager) CancelItem(id string) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if cancel, exists := qm.cancels[id]; exists {
		cancel()
		delete(qm.cancels, id)
	}

	for i := range qm.items {
		if qm.items[i].ID == id {
			qm.items[i].Status = StatusStaged
			qm.items[i].SpeedBPS = 0
			qm.items[i].SpeedFormatted = ""
			go qm.broadcast(qm.items[i])
			return nil
		}
	}
	return fmt.Errorf("item %s not found", id)
}

// RemoveItem removes an item completely from the queue.
func (qm *QueueManager) RemoveItem(id string) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if cancel, exists := qm.cancels[id]; exists {
		cancel()
		delete(qm.cancels, id)
	}

	updated := make([]DownloadItem, 0, len(qm.items))
	for _, it := range qm.items {
		if it.ID != id {
			updated = append(updated, it)
		}
	}
	qm.items = updated
	return nil
}

// GetItems returns a snapshot of all queue items.
func (qm *QueueManager) GetItems() []DownloadItem {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	copied := make([]DownloadItem, len(qm.items))
	copy(copied, qm.items)
	return copied
}

func (qm *QueueManager) findNextQueued() *DownloadItem {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	for i := range qm.items {
		if qm.items[i].Status == StatusQueued {
			qm.items[i].Status = StatusDownloading
			itemCopy := qm.items[i]
			return &itemCopy
		}
	}
	return nil
}

func (qm *QueueManager) updateItemStatus(id string, updateFn func(*DownloadItem)) {
	qm.mu.Lock()
	var updatedItem DownloadItem
	found := false
	for i := range qm.items {
		if qm.items[i].ID == id {
			updateFn(&qm.items[i])
			updatedItem = qm.items[i]
			found = true
			break
		}
	}
	qm.mu.Unlock()

	if found {
		qm.broadcast(updatedItem)
	}
}

func (qm *QueueManager) processLoop() {
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-qm.stopChan:
			return
		case <-ticker.C:
			// Check if we can start a job
			select {
			case qm.sem <- struct{}{}:
				item := qm.findNextQueued()
				if item == nil {
					<-qm.sem
					continue
				}

				go func(it DownloadItem) {
					defer func() { <-qm.sem }()
					qm.executeDownload(it)
				}(*item)
			default:
				// Semaphore is full
			}
		}
	}
}

func (qm *QueueManager) executeDownload(item DownloadItem) {
	ctx, cancel := context.WithCancel(context.Background())
	qm.mu.Lock()
	qm.cancels[item.ID] = cancel
	qm.mu.Unlock()

	defer func() {
		qm.mu.Lock()
		delete(qm.cancels, item.ID)
		qm.mu.Unlock()
	}()

	finalPath := filepath.Join(item.DestinationDir, item.FinalFilename)

	// Step 1: Pre-download hash verification of existing completed file
	qm.updateItemStatus(item.ID, func(it *DownloadItem) {
		it.Status = StatusVerifying
	})

	verifyRes, err := VerifyExistingFile(finalPath, item.ExpectedSHA256, item.Size)
	if err == nil && verifyRes.Exists && verifyRes.Valid {
		// File already exists and matches hash perfectly!
		qm.updateItemStatus(item.ID, func(it *DownloadItem) {
			it.Status = StatusCompleted
			it.Progress = 100.0
			it.DownloadedBytes = item.Size
		})
		return
	}

	// Step 2: Resolve URL & capabilities
	downloadURL := hfapi.BuildDownloadURL(item.RepoID, item.Revision, item.RemotePath)
	token := qm.configMgr.GetSettings().HFToken

	insp, inspErr := qm.resolver.InspectFile(downloadURL, token)
	totalSize := item.Size
	acceptRanges := true
	if inspErr == nil && insp != nil {
		if insp.ContentLength > 0 {
			totalSize = insp.ContentLength
		}
		acceptRanges = insp.AcceptRanges
		if item.ExpectedSHA256 == "" && insp.SHA256 != "" {
			item.ExpectedSHA256 = insp.SHA256
			qm.updateItemStatus(item.ID, func(it *DownloadItem) {
				it.ExpectedSHA256 = insp.SHA256
			})
		}
	}

	if item.ExpectedSHA256 == "" && item.RepoID != "" && item.RemotePath != "" {
		hfClient := hfapi.NewClient(hfapi.WithToken(token))
		if nodes, err := hfClient.FetchPathsInfo(item.RepoID, item.Revision, []string{item.RemotePath}); err == nil && len(nodes) > 0 && nodes[0].SHA256 != "" {
			item.ExpectedSHA256 = nodes[0].SHA256
			qm.updateItemStatus(item.ID, func(it *DownloadItem) {
				it.ExpectedSHA256 = nodes[0].SHA256
			})
		}
	}

	// Step 2.5: Safety check storage drive space to ensure sufficient room and prevent stalls
	if totalSize > 0 {
		if spaceInfo, spaceErr := platform.CheckDiskSpace(item.DestinationDir); spaceErr == nil {
			neededBytes := totalSize
			partPath := filepath.Join(item.DestinationDir, item.FinalFilename+".part")
			if fi, statErr := os.Stat(partPath); statErr == nil && fi.Size() > 0 {
				if fi.Size() < totalSize {
					neededBytes = totalSize - fi.Size()
				} else {
					neededBytes = 0
				}
			}

			// Require additional 100 MB buffer to prevent drive exhaustion and OS stalls
			const safetyBuffer = 100 * 1024 * 1024
			if neededBytes > 0 && spaceInfo.AvailableBytes < uint64(neededBytes+safetyBuffer) {
				qm.updateItemStatus(item.ID, func(it *DownloadItem) {
					it.Status = StatusFailed
					it.ErrorMessage = fmt.Sprintf("Insufficient disk space on %s: needs %s (+100MB buffer), only %s available",
						spaceInfo.Path, FormatBytes(neededBytes), FormatBytes(int64(spaceInfo.AvailableBytes)))
					it.SpeedBPS = 0
					it.SpeedFormatted = ""
				})
				return
			}
		}
	}

	qm.updateItemStatus(item.ID, func(it *DownloadItem) {
		it.Status = StatusDownloading
		it.Size = totalSize
	})

	// Step 3: Run multi-part chunk transfer engine
	job := &DownloadJob{
		TaskID:         item.ID,
		URL:            downloadURL,
		DestinationDir: item.DestinationDir,
		FinalFilename:  item.FinalFilename,
		TotalSize:      totalSize,
		ExpectedSHA256: item.ExpectedSHA256,
		Token:          token,
		MaxConnections: qm.configMgr.GetSettings().MaxConnectionsPerFile,
		AcceptRanges:   acceptRanges,
	}

	dlErr := qm.engine.Download(ctx, job, func(p DownloadProgress) {
		qm.updateItemStatus(item.ID, func(it *DownloadItem) {
			it.Progress = p.Percentage
			it.DownloadedBytes = p.DownloadedBytes
			it.SpeedBPS = p.SpeedBPS
			it.SpeedFormatted = p.SpeedFormatted
			it.ETASeconds = p.ETASeconds
		})
	})

	if dlErr != nil {
		if ctx.Err() != nil {
			// Context cancelled / paused
			return
		}
		qm.updateItemStatus(item.ID, func(it *DownloadItem) {
			it.Status = StatusFailed
			it.ErrorMessage = dlErr.Error()
			it.SpeedBPS = 0
			it.SpeedFormatted = ""
		})
		return
	}

	// Step 4: Finalize download (post-download SHA-256 verification and atomic promotion)
	qm.updateItemStatus(item.ID, func(it *DownloadItem) {
		it.Status = StatusVerifying
	})

	partPath := filepath.Join(item.DestinationDir, item.FinalFilename+".part")
	finErr := FinalizeDownload(partPath, finalPath, item.ExpectedSHA256)
	if finErr != nil {
		qm.updateItemStatus(item.ID, func(it *DownloadItem) {
			it.Status = StatusFailed
			it.ErrorMessage = finErr.Error()
			it.SpeedBPS = 0
			it.SpeedFormatted = ""
		})
		return
	}

	// Step 5: Mark completed!
	qm.updateItemStatus(item.ID, func(it *DownloadItem) {
		it.Status = StatusCompleted
		it.Progress = 100.0
		it.DownloadedBytes = totalSize
		it.SpeedBPS = 0
		it.SpeedFormatted = ""
		it.ETASeconds = 0
		if it.ExpectedSHA256 == "" && item.ExpectedSHA256 != "" {
			it.ExpectedSHA256 = item.ExpectedSHA256
		}
	})
}

// Close stops the background worker loop and cancels in-flight jobs.
func (qm *QueueManager) Close() {
	close(qm.stopChan)
	qm.mu.Lock()
	for _, cancel := range qm.cancels {
		cancel()
	}
	qm.cancels = make(map[string]context.CancelFunc)
	qm.mu.Unlock()
}
