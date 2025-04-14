package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/babelcloud/gbox/packages/api-server/config"
)

const (
	// Default reclaim interval
	defaultFileReclaimInterval = 14 * 24 * time.Hour // 14 days
)

// FileService handles file operations for the share directory
type FileService struct {
	shareDir string
}

// New creates a new Service
func New() (*FileService, error) {
	cfg := config.GetInstance()
	shareDir := cfg.File.Share

	// Create share directory if it doesn't exist
	if err := os.MkdirAll(shareDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create share directory: %v", err)
	}

	log.Info("File service initialized with share directory: %s", shareDir)

	return &FileService{
		shareDir: shareDir,
	}, nil
}

// validateAndCleanPath validates and cleans a file path
func (s *FileService) validateAndCleanPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	// Clean and validate the path
	cleanPath := filepath.Clean(path)
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}

	return cleanPath, nil
}

// getFullPath gets the full path in the share directory
func (s *FileService) getFullPath(cleanPath string) string {
	return filepath.Join(s.shareDir, cleanPath)
}
