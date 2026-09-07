package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"hf-downloader/internal/platform"
)

// FolderBookmark stores a favorite destination folder.
type FolderBookmark struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Path      string `json:"path"`
	IsDefault bool   `json:"isDefault"`
	LastUsed  int64  `json:"lastUsed"`
}

// RoutingRule maps a filename glob/substring pattern to a target folder.
type RoutingRule struct {
	Pattern   string `json:"pattern"`
	TargetDir string `json:"targetDir"`
	Label     string `json:"label"`
}

// Settings stores all application configuration and persistent state.
type Settings struct {
	HFToken               string           `json:"hfToken"`
	DefaultDownloadDir    string           `json:"defaultDownloadDir"`
	AutoDownload          bool             `json:"autoDownload"`
	MaxConcurrentFiles    int              `json:"maxConcurrentFiles"`
	MaxConnectionsPerFile int              `json:"maxConnectionsPerFile"`
	Bookmarks             []FolderBookmark `json:"bookmarks"`
	RoutingRules          []RoutingRule    `json:"routingRules"`
	RecentPaths           []string         `json:"recentPaths"`
}

// ConfigManager handles thread-safe loading and saving of Settings.
type ConfigManager struct {
	mu       sync.RWMutex
	settings Settings
	filePath string
}

var (
	defaultConfigManager *ConfigManager
	onceConfig           sync.Once
)

// DefaultSettings returns sensible defaults.
func DefaultSettings() Settings {
	home, _ := os.UserHomeDir()
	defaultDownloads := filepath.Join(home, "Downloads")

	return Settings{
		HFToken:               "",
		DefaultDownloadDir:    defaultDownloads,
		AutoDownload:          false,
		MaxConcurrentFiles:    2,
		MaxConnectionsPerFile: 8,
		Bookmarks: []FolderBookmark{
			{
				ID:        "bm-checkpoints",
				Label:     "Checkpoints",
				Path:      filepath.Join(defaultDownloads, "models", "checkpoints"),
				IsDefault: true,
				LastUsed:  time.Now().Unix(),
			},
			{
				ID:        "bm-lora",
				Label:     "LoRAs",
				Path:      filepath.Join(defaultDownloads, "models", "loras"),
				IsDefault: false,
				LastUsed:  time.Now().Unix(),
			},
			{
				ID:        "bm-vae",
				Label:     "VAE",
				Path:      filepath.Join(defaultDownloads, "models", "vae"),
				IsDefault: false,
				LastUsed:  time.Now().Unix(),
			},
			{
				ID:        "bm-text-encoders",
				Label:     "Text Encoders",
				Path:      filepath.Join(defaultDownloads, "models", "text_encoders"),
				IsDefault: false,
				LastUsed:  time.Now().Unix(),
			},
		},
		RoutingRules: []RoutingRule{
			{
				Pattern:   "*vae*.safetensors",
				TargetDir: filepath.Join(defaultDownloads, "models", "vae"),
				Label:     "VAE Rule",
			},
			{
				Pattern:   "*lora*.safetensors",
				TargetDir: filepath.Join(defaultDownloads, "models", "loras"),
				Label:     "LoRA Rule",
			},
			{
				Pattern:   "*clip*.safetensors",
				TargetDir: filepath.Join(defaultDownloads, "models", "text_encoders"),
				Label:     "CLIP Text Encoder Rule",
			},
			{
				Pattern:   "*t5*.safetensors",
				TargetDir: filepath.Join(defaultDownloads, "models", "text_encoders"),
				Label:     "T5 Text Encoder Rule",
			},
		},
		RecentPaths: []string{defaultDownloads},
	}
}

// GetManager returns the singleton ConfigManager.
func GetManager() *ConfigManager {
	onceConfig.Do(func() {
		cfgPath, err := platform.GetConfigFilePath()
		if err != nil {
			cfgPath = "config.json"
		}
		cm := &ConfigManager{
			filePath: cfgPath,
			settings: DefaultSettings(),
		}
		_ = cm.Load()
		defaultConfigManager = cm
	})
	return defaultConfigManager
}

// Load loads settings from the config file.
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.filePath)
	if err != nil {
		// If file does not exist, save default settings
		if os.IsNotExist(err) {
			_ = cm.saveLocked()
		}
		return err
	}

	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	cm.settings = s
	return nil
}

// Save saves settings to the config file.
func (cm *ConfigManager) Save() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.saveLocked()
}

func (cm *ConfigManager) saveLocked() error {
	data, err := json.MarshalIndent(cm.settings, "", "  ")
	if err != nil {
		return err
	}

	temp := fmt.Sprintf("%s.tmp.%d", cm.filePath, os.Getpid())
	if err := os.WriteFile(temp, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(temp, cm.filePath); err != nil {
		_ = os.Remove(cm.filePath)
		if retry := os.Rename(temp, cm.filePath); retry != nil {
			_ = os.Remove(temp)
			return retry
		}
	}
	return nil
}

// GetSettings returns a copy of the current settings.
func (cm *ConfigManager) GetSettings() Settings {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.settings
}

// UpdateSettings updates the settings and persists to disk.
func (cm *ConfigManager) UpdateSettings(s Settings) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.settings = s
	return cm.saveLocked()
}

// ResolveDestination applies routing rules to determine target directory for a file.
func (cm *ConfigManager) ResolveDestination(remotePath string) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	fileName := strings.ToLower(filepath.Base(remotePath))

	for _, rule := range cm.settings.RoutingRules {
		pattern := strings.ToLower(rule.Pattern)
		// Check glob match
		if matched, _ := filepath.Match(pattern, fileName); matched {
			if rule.TargetDir != "" {
				return rule.TargetDir
			}
		}
		// Also check substring if pattern doesn't contain glob stars
		if !strings.Contains(pattern, "*") && strings.Contains(fileName, pattern) {
			if rule.TargetDir != "" {
				return rule.TargetDir
			}
		}
	}

	if cm.settings.DefaultDownloadDir != "" {
		return cm.settings.DefaultDownloadDir
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Downloads")
}
