package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/babelcloud/gbox/packages/api-server/internal/common"
	"github.com/babelcloud/gbox/packages/api-server/pkg/file"
)

// ShareFile shares a file from a box to the share directory
func (s *FileService) ShareFile(ctx context.Context, boxID, path string) (*model.FileShareResult, error) {
	if boxID == "" {
		return nil, fmt.Errorf("box ID is required")
	}

	cleanPath, err := s.validateAndCleanPath(path)
	if err != nil {
		return nil, err
	}

	// Check if path starts with default share directory path
	if strings.HasPrefix(cleanPath, "/"+boxID+common.DefaultShareDirPath) {
		// Remove the prefix to get the relative path
		relativePath := strings.TrimPrefix(cleanPath, "/"+boxID+common.DefaultShareDirPath)
		relativePath = filepath.Clean(relativePath)
		if !strings.HasPrefix(relativePath, "/") {
			relativePath = "/" + relativePath
		}

		// Check if file already exists in share directory
		sharePath := filepath.Join(s.shareDir, boxID, relativePath)
		if _, err := os.Stat(sharePath); err == nil {
			// File exists, return success with file list
			fileList, err := getFileList(sharePath)
			if err != nil {
				log.Error("Error getting file list: %v", err)
			}
			return &model.FileShareResult{
				Success:  true,
				Message:  "File already exists",
				FileList: fileList,
			}, nil
		}
	}

	// Create share directory for the box if it doesn't exist
	boxShareDir := filepath.Join(s.shareDir, boxID)
	if err := os.MkdirAll(boxShareDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create share directory: %v", err)
	}

	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(filepath.Join(boxShareDir, cleanPath))
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %v", err)
	}

	// TODO: Implement file sharing from box to share directory
	// This will require integration with the box service to get the file content

	// Get the list of shared files
	fileList, err := getFileList(filepath.Join(boxShareDir, cleanPath))
	if err != nil {
		log.Error("Error getting file list: %v", err)
	}

	return &model.FileShareResult{
		Success:  true,
		Message:  "File shared successfully",
		FileList: fileList,
	}, nil
}
