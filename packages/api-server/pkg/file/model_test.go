package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// FileType represents the type of a file
type FileType string

const (
	FileTypeDirectory FileType = "directory"
	FileTypeFile      FileType = "file"
	FileTypeSymlink   FileType = "symlink"
	FileTypeSocket    FileType = "socket"
	FileTypePipe      FileType = "pipe"
	FileTypeDevice    FileType = "device"
)

// FileStat represents file metadata
type FileStat struct {
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Size    int64    `json:"size"`
	Mode    string   `json:"mode"`
	ModTime string   `json:"modTime"`
	Type    FileType `json:"type"`
	Mime    string   `json:"mime"`
}

// FileError represents a file operation error response
type FileError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// FileOperationParams represents a request to share a file from a box
type FileOperationParams struct {
	BoxID     string `json:"boxId"`   // ID of the box to share from
	Path      string `json:"path"`    // Path to the file in the box
	Content   string `json:"content"` // Content to write to the file
	Operation string `json:"operation"` // Operation to perform (share, write, reclaim)
}

// FileShareResult represents the response for file sharing operations
type FileShareResult struct {
	Success  bool       `json:"success"`
	Message  string     `json:"message"`
	FileList []FileStat `json:"fileList"`
}

func TestFileType(t *testing.T) {
	tests := []struct {
		name     string
		fileType FileType
		expected string
	}{
		{
			name:     "Directory type",
			fileType: FileTypeDirectory,
			expected: "directory",
		},
		{
			name:     "File type",
			fileType: FileTypeFile,
			expected: "file",
		},
		{
			name:     "Symlink type",
			fileType: FileTypeSymlink,
			expected: "symlink",
		},
		{
			name:     "Socket type",
			fileType: FileTypeSocket,
			expected: "socket",
		},
		{
			name:     "Pipe type",
			fileType: FileTypePipe,
			expected: "pipe",
		},
		{
			name:     "Device type",
			fileType: FileTypeDevice,
			expected: "device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.fileType))
		})
	}
}

func TestFileStat(t *testing.T) {
	fileStat := FileStat{
		Name:    "test.txt",
		Path:    "/path/to/test.txt",
		Size:    1024,
		Mode:    "0644",
		ModTime: "2025-05-01T10:00:00Z",
		Type:    FileTypeFile,
		Mime:    "text/plain",
	}

	assert.Equal(t, "test.txt", fileStat.Name)
	assert.Equal(t, "/path/to/test.txt", fileStat.Path)
	assert.Equal(t, int64(1024), fileStat.Size)
	assert.Equal(t, "0644", fileStat.Mode)
	assert.Equal(t, "2025-05-01T10:00:00Z", fileStat.ModTime)
	assert.Equal(t, FileTypeFile, fileStat.Type)
	assert.Equal(t, "text/plain", fileStat.Mime)
}

func TestFileError(t *testing.T) {
	fileError := FileError{
		Code:    "NOT_FOUND",
		Message: "File not found",
	}

	assert.Equal(t, "NOT_FOUND", fileError.Code)
	assert.Equal(t, "File not found", fileError.Message)
}

func TestFileOperationParams(t *testing.T) {
	params := FileOperationParams{
		BoxID:     "box123",
		Path:      "/test/file.txt",
		Content:   "test content",
		Operation: "share",
	}

	assert.Equal(t, "box123", params.BoxID)
	assert.Equal(t, "/test/file.txt", params.Path)
	assert.Equal(t, "test content", params.Content)
	assert.Equal(t, "share", params.Operation)
}

func TestFileShareResult(t *testing.T) {
	fileList := []FileStat{
		{
			Name: "file1.txt",
			Path: "/path/to/file1.txt",
			Type: FileTypeFile,
		},
		{
			Name: "dir1",
			Path: "/path/to/dir1",
			Type: FileTypeDirectory,
		},
	}

	result := FileShareResult{
		Success:  true,
		Message:  "Files shared successfully",
		FileList: fileList,
	}

	assert.True(t, result.Success)
	assert.Equal(t, "Files shared successfully", result.Message)
	assert.Len(t, result.FileList, 2)
	assert.Equal(t, "file1.txt", result.FileList[0].Name)
	assert.Equal(t, "dir1", result.FileList[1].Name)
}
