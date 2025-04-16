package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	model "github.com/babelcloud/gbox/packages/api-server/pkg/file"
)

// ReclaimFiles removes files that haven't been accessed for more than 14 days
func (s *FileService) ReclaimFiles(ctx context.Context) (*model.FileShareResult, error) {
	cutoffTime := time.Now().Add(-defaultFileReclaimInterval)
	var reclaimedFiles []string
	var fileStats []model.FileStat
	var errors []string
	emptyDirs := make(map[string]bool) // Track empty directories

	// Walk through the share directory
	err := filepath.Walk(s.shareDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errors = append(errors, fmt.Sprintf("Error accessing path %s: %v", path, err))
			return nil
		}

		// Skip the share directory itself
		if path == s.shareDir {
			return nil
		}

		// Check if file is older than cutoff time
		if info.ModTime().Before(cutoffTime) {
			// Collect file stats before deletion
			fileStats = append(fileStats, model.FileStat{
				Name:    info.Name(),
				Path:    path,
				Size:    info.Size(),
				Mode:    info.Mode().String(),
				ModTime: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
				Type:    getFileType(info),
				Mime:    getMimeType(path, info),
			})

			// For symbolic links, only remove the link itself
			if info.Mode()&os.ModeSymlink != 0 {
				if err := os.Remove(path); err != nil {
					errors = append(errors, fmt.Sprintf("Error removing symlink %s: %v", path, err))
				} else {
					reclaimedFiles = append(reclaimedFiles, path)
					// Mark parent directory as potentially empty
					parentDir := filepath.Dir(path)
					emptyDirs[parentDir] = true
				}
				return nil
			}

			// For regular files and directories, use RemoveAll
			if err := os.RemoveAll(path); err != nil {
				errors = append(errors, fmt.Sprintf("Error removing %s: %v", path, err))
			} else {
				reclaimedFiles = append(reclaimedFiles, path)
				// Mark parent directory as potentially empty
				parentDir := filepath.Dir(path)
				emptyDirs[parentDir] = true
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	// Check and remove empty directories
	for dir := range emptyDirs {
		// Skip the share directory itself
		if dir == s.shareDir {
			continue
		}

		// Check if directory is empty
		entries, err := os.ReadDir(dir)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Error checking directory %s: %v", dir, err))
			continue
		}

		// If directory is empty, remove it
		if len(entries) == 0 {
			if err := os.Remove(dir); err != nil {
				errors = append(errors, fmt.Sprintf("Error removing empty directory %s: %v", dir, err))
			} else {
				reclaimedFiles = append(reclaimedFiles, dir)
				// Mark parent directory as potentially empty
				parentDir := filepath.Dir(dir)
				emptyDirs[parentDir] = true
			}
		}
	}

	// Prepare response
	response := &model.FileShareResult{
		Success:  true,
		Message:  "Files reclaimed successfully",
		FileList: fileStats,
	}

	return response, nil
}
