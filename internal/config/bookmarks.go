package config

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AddBookmark registers a new bookmark.
func (cm *ConfigManager) AddBookmark(label, path string) (FolderBookmark, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	bm := FolderBookmark{
		ID:        "bm-" + uuid.New().String()[:8],
		Label:     label,
		Path:      path,
		IsDefault: false,
		LastUsed:  time.Now().Unix(),
	}

	cm.settings.Bookmarks = append(cm.settings.Bookmarks, bm)
	return bm, cm.saveLocked()
}

// DeleteBookmark removes a bookmark by ID.
func (cm *ConfigManager) DeleteBookmark(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	updated := make([]FolderBookmark, 0, len(cm.settings.Bookmarks))
	found := false
	for _, b := range cm.settings.Bookmarks {
		if b.ID == id {
			found = true
			continue
		}
		updated = append(updated, b)
	}

	if !found {
		return fmt.Errorf("bookmark with ID %s not found", id)
	}

	cm.settings.Bookmarks = updated
	return cm.saveLocked()
}

// AddRecentPath pushes a directory path to the recent list (keeping last 10).
func (cm *ConfigManager) AddRecentPath(path string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var newRecents []string
	newRecents = append(newRecents, path)
	for _, p := range cm.settings.RecentPaths {
		if p != path {
			newRecents = append(newRecents, p)
		}
		if len(newRecents) >= 10 {
			break
		}
	}
	cm.settings.RecentPaths = newRecents
	_ = cm.saveLocked()
}
