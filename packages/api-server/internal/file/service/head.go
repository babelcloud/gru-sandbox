package service

import (
	"context"
	"fmt"
	"os"

	"github.com/babelcloud/gbox/packages/api-server/pkg/file"
)

// HeadFile gets metadata about a file
func (s *FileService) HeadFile(ctx context.Context, path string) (*model.FileStat, error) {
	cleanPath, err := s.validateAndCleanPath(path)
	if err != nil {
		return nil, err
	}

	// Construct full path
	fullPath := s.getFullPath(cleanPath)

	// Get file info
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("error getting file info: %v", err)
	}

	// Create file stat response
	stat := &model.FileStat{
		Name:    info.Name(),
		Path:    cleanPath,
		Size:    info.Size(),
		Mode:    info.Mode().String(),
		ModTime: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
		Type:    getFileType(info),
		Mime:    getMimeType(fullPath, info),
	}

	return stat, nil
}