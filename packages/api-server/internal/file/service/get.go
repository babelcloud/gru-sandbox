package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
)

// FileContent represents the content of a file
type FileContent struct {
	Reader   io.ReadCloser
	MimeType string
	Size     int64
}

// GetFile gets the content of a file
func (s *FileService) GetFile(ctx context.Context, path string) (*FileContent, error) {
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

	// Handle directories
	if info.IsDir() {
		// For directories, return a reader with directory listing
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return nil, fmt.Errorf("error reading directory: %v", err)
		}

		// Create a buffer to hold the directory listing
		var buf bytes.Buffer
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			fmt.Fprintf(&buf, "%s\t%d\t%s\t%s\n",
				info.Mode().String(),
				info.Size(),
				info.ModTime().Format("2006-01-02 15:04:05"),
				info.Name())
		}

		return &FileContent{
			Reader:   io.NopCloser(&buf),
			MimeType: "application/x-directory",
			Size:     int64(buf.Len()),
		}, nil
	}

	// Open file
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}

	return &FileContent{
		Reader:   file,
		MimeType: getMimeType(fullPath, info),
		Size:     info.Size(),
	}, nil
}
