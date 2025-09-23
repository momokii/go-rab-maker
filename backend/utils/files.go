package utils

import (
	"os"
	"path/filepath"
)

const (
	APP_DATA_DIR_NAME = "RABMaker"
)

// getBaseDir returns the appropriate base directory for storing application data
// In development, it uses a local directory; in production, it uses a user-specific directory
func GetBaseDir() string {
	// Try to use a directory in user's home for production
	userConfigDir, err := os.UserConfigDir()

	if err == nil {
		// Use a directory in the user's config folder (e.g., AppData on Windows)
		appDataDir := filepath.Join(userConfigDir, APP_DATA_DIR_NAME)

		// Check/create the directory
		if err := os.MkdirAll(appDataDir, 0755); err == nil {
			return appDataDir
		}
	}

	// Fallback: try to use system temp directory
	tempBaseDir := filepath.Join(os.TempDir(), APP_DATA_DIR_NAME)
	if err := os.MkdirAll(tempBaseDir, 0755); err == nil {
		return tempBaseDir
	}

	// Last resort: use a local directory (for development)
	return filepath.Join(".", "backend", "core")
}
