package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetConfigDir returns the persistent configuration directory for hf-downloader.
func GetConfigDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData != "" {
		// Windows
		dir := filepath.Join(appData, "hf-downloader")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
		return dir, nil
	}

	userConfig, err := os.UserConfigDir()
	if err == nil {
		dir := filepath.Join(userConfig, "hf-downloader")
		if mkErr := os.MkdirAll(dir, 0755); mkErr == nil {
			return dir, nil
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user configuration directory: %w", err)
	}

	dir := filepath.Join(home, ".config", "hf-downloader")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// GetConfigFilePath returns the absolute path to config.json.
func GetConfigFilePath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}
