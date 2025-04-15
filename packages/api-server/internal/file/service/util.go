package service

import (
	"os"
	"path/filepath"

	"github.com/babelcloud/gbox/packages/api-server/pkg/file"
	"github.com/babelcloud/gbox/packages/api-server/pkg/logger"
	"github.com/gabriel-vasile/mimetype"
)

var log = logger.New()

// getMimeType determines the MIME type of a file
func getMimeType(path string, info os.FileInfo) string {
	if info.IsDir() {
		return "application/x-directory"
	}

	// Use mimetype library to detect MIME type
	mime, err := mimetype.DetectFile(path)
	if err != nil {
		log.Error("Error detecting MIME type: %v", err)
		return "application/octet-stream"
	}

	return mime.String()
}

// getFileList returns a list of files in the given path
func getFileList(path string) ([]model.FileStat, error) {
	// First check if path is a file
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// If it's a file, return its info directly
	if !info.IsDir() {
		return []model.FileStat{
			{
				Name:    info.Name(),
				Path:    path,
				Size:    info.Size(),
				Mode:    info.Mode().String(),
				ModTime: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
				Type:    getFileType(info),
				Mime:    getMimeType(path, info),
			},
		}, nil
	}

	// If it's a directory, read its contents
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []model.FileStat
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, model.FileStat{
			Name:    info.Name(),
			Path:    filepath.Join(path, info.Name()),
			Size:    info.Size(),
			Mode:    info.Mode().String(),
			ModTime: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
			Type:    getFileType(info),
			Mime:    getMimeType(filepath.Join(path, info.Name()), info),
		})
	}

	return files, nil
}

// getFileType determines the type of a file
func getFileType(info os.FileInfo) model.FileType {
	if info.IsDir() {
		return model.FileTypeDirectory
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return model.FileTypeSymlink
	}
	if info.Mode()&os.ModeSocket != 0 {
		return model.FileTypeSocket
	}
	if info.Mode()&os.ModeNamedPipe != 0 {
		return model.FileTypePipe
	}
	if info.Mode()&os.ModeDevice != 0 {
		return model.FileTypeDevice
	}
	return model.FileTypeFile
}
