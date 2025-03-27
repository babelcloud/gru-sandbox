package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
)

// FileConfig handles file service configuration
type FileConfig struct {
	ShareDir string
}

// NewFileConfig creates a new FileConfig
func NewFileConfig() Config {
	return &FileConfig{}
}

// Initialize initializes the file configuration
func (c *FileConfig) Initialize(logger *log.Logger) error {
	// First try to get explicit share directory from config
	if shareDir := v.GetString("gbox.share"); shareDir != "" {
		c.ShareDir = shareDir
		logger.Info("Using configured share directory: \"%s\"", c.ShareDir)
	} else {
		// If share dir not configured, try to use gbox.home
		if homeDir := v.GetString("gbox.home"); homeDir != "" {
			c.ShareDir = filepath.Join(homeDir, "share")
			logger.Info("Using share directory from gbox.home: \"%s\"", c.ShareDir)
		} else {
			// If gbox.home not configured, use default path
			userHomeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %v", err)
			}
			c.ShareDir = filepath.Join(userHomeDir, ".gbox", "share")
			logger.Info("Using default share directory: \"%s\"", c.ShareDir)
		}
	}

	// Create share directory if it doesn't exist
	if err := os.MkdirAll(c.ShareDir, 0755); err != nil {
		return fmt.Errorf("failed to create share directory: %v", err)
	}

	return nil
}

// GetFileShareDir returns the share directory path
func (c *FileConfig) GetFileShareDir() string {
	return c.ShareDir
}
