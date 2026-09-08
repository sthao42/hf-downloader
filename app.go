package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"hf-downloader/internal/config"
	"hf-downloader/internal/downloader"
	"hf-downloader/internal/hfapi"
	"hf-downloader/internal/platform"
)

// InspectResponse contains the parsed target details and discovered files.
type InspectResponse struct {
	Target *hfapi.ParsedTarget `json:"target"`
	Files  []hfapi.FileNode    `json:"files"`
}

// App coordinates the desktop frontend with backend services.
type App struct {
	ctx       context.Context
	configMgr *config.ConfigManager
	queueMgr  *downloader.QueueManager
	hfClient  *hfapi.Client
	resolver  *hfapi.Resolver
}

// NewApp initializes backend services.
func NewApp() *App {
	cm := config.GetManager()
	qm := downloader.NewQueueManager(cm)
	httpClient := &http.Client{Timeout: 30 * time.Second}

	return &App{
		configMgr: cm,
		queueMgr:  qm,
		hfClient:  hfapi.NewClient(hfapi.WithToken(cm.GetSettings().HFToken)),
		resolver:  hfapi.NewResolver(httpClient),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Subscribe to queue updates and emit to Wails frontend
	a.queueMgr.OnProgress(func(item downloader.DownloadItem) {
		wruntime.EventsEmit(a.ctx, "download:progress", item)
	})
}

func (a *App) shutdown(ctx context.Context) {
	if a.queueMgr != nil {
		a.queueMgr.Close()
	}
}

// ParseAndInspect sniffs a pasted URL and discovers matching repository files.
func (a *App) ParseAndInspect(inputURL string) (*InspectResponse, error) {
	target, err := hfapi.ParseHFURL(inputURL)
	if err != nil {
		return nil, err
	}

	token := a.configMgr.GetSettings().HFToken
	a.hfClient.SetToken(token)

	res := &InspectResponse{
		Target: target,
		Files:  make([]hfapi.FileNode, 0),
	}

	switch target.Type {
	case hfapi.TargetFile:
		// Inspect single file directly
		downloadURL := hfapi.BuildDownloadURL(target.RepoID, target.Revision, target.Subpath)
		insp, inspErr := a.resolver.InspectFile(downloadURL, token)

		node := hfapi.FileNode{
			Type:        "file",
			Path:        target.Subpath,
			DownloadURL: downloadURL,
		}

		if inspErr == nil && insp != nil {
			node.Size = insp.ContentLength
		}

		res.Files = append(res.Files, node)
		return res, nil

	case hfapi.TargetFolder, hfapi.TargetRepo:
		// Fetch repository tree from API
		files, treeErr := a.hfClient.FetchTree(target, true)
		if treeErr != nil {
			return nil, treeErr
		}

		// Filter for only files (skip directories in selection list)
		var onlyFiles []hfapi.FileNode
		for _, f := range files {
			if f.Type == "file" {
				onlyFiles = append(onlyFiles, f)
			}
		}
		res.Files = onlyFiles
		return res, nil

	default:
		return nil, fmt.Errorf("unsupported target type: %s", target.Type)
	}
}

// GetSettings returns current application settings.
func (a *App) GetSettings() config.Settings {
	return a.configMgr.GetSettings()
}

// SaveSettings persists updated application settings and refreshes clients.
func (a *App) SaveSettings(settings config.Settings) error {
	err := a.configMgr.UpdateSettings(settings)
	if err == nil {
		a.hfClient.SetToken(settings.HFToken)
		a.queueMgr.SetAutoDownload(settings.AutoDownload)
	}
	return err
}

// GetBookmarks returns the list of bookmarked destination folders.
func (a *App) GetBookmarks() []config.FolderBookmark {
	return a.configMgr.GetSettings().Bookmarks
}

// AddBookmark creates a new bookmark.
func (a *App) AddBookmark(label, path string) (config.FolderBookmark, error) {
	return a.configMgr.AddBookmark(label, path)
}

// DeleteBookmark removes a bookmark by ID.
func (a *App) DeleteBookmark(id string) error {
	return a.configMgr.DeleteBookmark(id)
}

// SelectDirectoryDialog opens the OS native directory picker.
func (a *App) SelectDirectoryDialog(defaultPath string) (string, error) {
	if defaultPath == "" {
		defaultPath = a.configMgr.GetSettings().DefaultDownloadDir
	}

	chosen, err := wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{
		DefaultDirectory: defaultPath,
		Title:            "Select Destination Directory",
	})
	if err != nil {
		return "", err
	}

	if chosen != "" {
		a.configMgr.AddRecentPath(chosen)
	}
	return chosen, nil
}

// QueueItems stages or enqueues items for download.
func (a *App) QueueItems(items []downloader.DownloadItem, autoStart bool) []downloader.DownloadItem {
	return a.queueMgr.QueueItems(items, autoStart)
}

// GetQueueItems returns all active, queued, and completed items.
func (a *App) GetQueueItems() []downloader.DownloadItem {
	return a.queueMgr.GetItems()
}

// StartItem manually starts a staged or paused download.
func (a *App) StartItem(id string) error {
	return a.queueMgr.StartItem(id)
}

// PauseItem pauses a download in progress.
func (a *App) PauseItem(id string) error {
	return a.queueMgr.PauseItem(id)
}

// ResumeItem resumes a paused download.
func (a *App) ResumeItem(id string) error {
	return a.queueMgr.ResumeItem(id)
}

// CancelItem cancels an active item.
func (a *App) CancelItem(id string) error {
	return a.queueMgr.CancelItem(id)
}

// RemoveItem completely deletes an item from the queue list.
func (a *App) RemoveItem(id string) error {
	return a.queueMgr.RemoveItem(id)
}

// UpdateItemDestination changes the target directory of an item.
func (a *App) UpdateItemDestination(id, newDestDir string) error {
	return a.queueMgr.UpdateItemDestination(id, newDestDir)
}

// VerifyLocalFile checks if a file already exists locally and matches SHA-256.
func (a *App) VerifyLocalFile(item downloader.DownloadItem) (*downloader.VerificationResult, error) {
	filePath := filepath.Join(item.DestinationDir, item.FinalFilename)
	return downloader.VerifyExistingFile(filePath, item.ExpectedSHA256, item.Size)
}

// OpenHFTokenPage opens the Hugging Face settings page in the default web browser.
func (a *App) OpenHFTokenPage() {
	wruntime.BrowserOpenURL(a.ctx, "https://huggingface.co/settings/tokens")
}

// OpenFolder opens the destination directory in the OS file explorer.
func (a *App) OpenFolder(folderPath string) error {
	if folderPath == "" {
		return fmt.Errorf("folder path cannot be empty")
	}

	_ = os.MkdirAll(folderPath, 0755)

	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", folderPath).Start()
	case "darwin":
		return exec.Command("open", folderPath).Start()
	default:
		return exec.Command("xdg-open", folderPath).Start()
	}
}

// CheckDiskSpace returns drive free and total space for the destination folder path.
func (a *App) CheckDiskSpace(targetPath string) (platform.DiskSpaceInfo, error) {
	return platform.CheckDiskSpace(targetPath)
}

// UpdateBookmarks saves the reordered or modified list of bookmarks.
func (a *App) UpdateBookmarks(bookmarks []config.FolderBookmark) error {
	return a.configMgr.UpdateBookmarks(bookmarks)
}

// UpdateRecentPaths saves the reordered recent paths list.
func (a *App) UpdateRecentPaths(paths []string) error {
	return a.configMgr.UpdateRecentPaths(paths)
}
