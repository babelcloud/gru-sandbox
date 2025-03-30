package handlers

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/babelcloud/gru-sandbox/packages/api-server/config"
	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/common"
	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
	"github.com/babelcloud/gru-sandbox/packages/api-server/types"
	"github.com/emicklei/go-restful/v3"
	"github.com/gabriel-vasile/mimetype"
)

var logger = log.New()

const (
	// Default reclaim interval
	defaultFileReclaimInterval = 14 * 24 * time.Hour // 14 days
)

// FileHandler handles file operations for the share directory
type FileHandler struct {
	shareDir string
}

// NewFileHandler creates a new FileHandler
func NewFileHandler(cfg *config.FileConfig) (*FileHandler, error) {
	shareDir := cfg.GetFileShareDir()

	// Create share directory if it doesn't exist
	if err := os.MkdirAll(shareDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create share directory: %v", err)
	}

	return &FileHandler{
		shareDir: shareDir,
	}, nil
}

// writeFileError writes a structured error response
func writeFileError(resp *restful.Response, statusCode int, code, message string) {
	resp.WriteHeader(statusCode)
	resp.WriteAsJson(models.FileError{
		Code:    code,
		Message: message,
	})
}

// getFileType determines the type of a file
func getFileType(info os.FileInfo) models.FileType {
	if info.IsDir() {
		return models.FileTypeDirectory
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return models.FileTypeSymlink
	}
	if info.Mode()&os.ModeSocket != 0 {
		return models.FileTypeSocket
	}
	if info.Mode()&os.ModeNamedPipe != 0 {
		return models.FileTypePipe
	}
	if info.Mode()&os.ModeDevice != 0 {
		return models.FileTypeDevice
	}
	return models.FileTypeFile
}

// HeadFile handles HEAD requests to get file metadata
func (h *FileHandler) HeadFile(req *restful.Request, resp *restful.Response) {
	path := req.PathParameter("path")
	logger.Info("Received HEAD request for file: %s", path)

	if path == "" {
		logger.Error("Invalid request: path is empty")
		writeFileError(resp, http.StatusBadRequest, "INVALID_REQUEST", "Path is required")
		return
	}

	// Clean and validate the path
	cleanPath := filepath.Clean(path)
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}
	logger.Debug("Cleaned path: %s", cleanPath)

	// Construct full path
	fullPath := filepath.Join(h.shareDir, cleanPath)
	logger.Debug("Full file path: %s", fullPath)

	// Get file info
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Error("File not found: %s", path)
			writeFileError(resp, http.StatusNotFound, "FILE_NOT_FOUND", fmt.Sprintf("File not found: %s", path))
		} else {
			logger.Error("Error getting file info: %v", err)
			writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error getting file info: %v", err))
		}
		return
	}

	// Create file stat response
	stat := models.FileStat{
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode().String(),
		ModTime: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
		Type:    getFileType(info),
		Mime:    getMimeType(fullPath, info),
	}
	logger.Debug("File stat: %+v", stat)

	// Convert stat to JSON string
	statJSON, err := json.Marshal(stat)
	if err != nil {
		logger.Error("Error marshaling stat: %v", err)
		writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error marshaling stat: %v", err))
		return
	}

	// Set response headers
	mimeType := getMimeType(fullPath, info)
	resp.Header().Set("Content-Type", mimeType)
	resp.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	resp.Header().Set("X-Gbox-File-Stat", string(statJSON))
	logger.Debug("Response headers: Content-Type=%s, Content-Length=%d", mimeType, info.Size())

	resp.WriteHeader(http.StatusOK)
	logger.Info("Successfully processed HEAD request for file: %s", path)
}

// GetFile handles GET requests to retrieve file content
func (h *FileHandler) GetFile(req *restful.Request, resp *restful.Response) {
	path := req.PathParameter("path")
	if path == "" {
		writeFileError(resp, http.StatusBadRequest, "INVALID_REQUEST", "Path is required")
		return
	}

	// Clean and validate the path
	cleanPath := filepath.Clean(path)
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}

	// Construct full path
	fullPath := filepath.Join(h.shareDir, cleanPath)

	// Get file info
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeFileError(resp, http.StatusNotFound, "FILE_NOT_FOUND", fmt.Sprintf("File not found: %s", path))
		} else {
			logger.Printf("Error getting file info: %v", err)
			writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error getting file info: %v", err))
		}
		return
	}

	// Handle directories
	if info.IsDir() {
		writeFileError(resp, http.StatusBadRequest, "INVALID_REQUEST", "Cannot get content of a directory")
		return
	}

	// Open file
	file, err := os.Open(fullPath)
	if err != nil {
		logger.Printf("Error opening file: %v", err)
		writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error opening file: %v", err))
		return
	}
	defer file.Close()

	// Set response headers
	mimeType := getMimeType(fullPath, info)
	resp.Header().Set("Content-Type", mimeType)
	resp.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))

	logger.Debug("Response headers: Content-Type=%s, Content-Length=%d", mimeType, info.Size())

	// Copy file content to response
	if _, err := io.Copy(resp, file); err != nil {
		logger.Printf("Error copying file content: %v", err)
		writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error copying file content: %v", err))
		return
	}

	logger.Info("Successfully processed HEAD request for file: %s", path)
}

// getMimeType determines the MIME type of a file using mimetype library
func getMimeType(path string, info os.FileInfo) string {
	if info.IsDir() {
		return "application/x-directory"
	}

	// Use mimetype library to detect MIME type
	mime, err := mimetype.DetectFile(path)
	if err != nil {
		logger.Printf("Error detecting MIME type: %v", err)
		return "application/octet-stream"
	}

	return mime.String()
}

// ReclaimFiles removes files that haven't been accessed for more than 14 days
func (h *FileHandler) ReclaimFiles(req *restful.Request, resp *restful.Response) {
	cutoffTime := time.Now().Add(-defaultFileReclaimInterval)
	var reclaimedFiles []string
	var errors []string
	emptyDirs := make(map[string]bool) // Track empty directories

	// Walk through the share directory
	err := filepath.Walk(h.shareDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errors = append(errors, fmt.Sprintf("Error accessing path %s: %v", path, err))
			return nil
		}

		// Skip the share directory itself
		if path == h.shareDir {
			return nil
		}

		// Check if file is older than cutoff time
		if info.ModTime().Before(cutoffTime) {
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
		writeFileError(resp, http.StatusInternalServerError, "RECLAIM_ERROR", fmt.Sprintf("Error walking directory: %v", err))
		return
	}

	// Check and remove empty directories
	for dir := range emptyDirs {
		// Skip the share directory itself
		if dir == h.shareDir {
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
	response := struct {
		ReclaimedFiles []string `json:"reclaimed_files"`
		Errors         []string `json:"errors,omitempty"`
	}{
		ReclaimedFiles: reclaimedFiles,
		Errors:         errors,
	}

	resp.WriteAsJson(response)
}

// ShareFile handles sharing a file from a box to the share directory
func (h *FileHandler) ShareFile(req *restful.Request, resp *restful.Response) {
	var shareReq models.FileShareRequest
	if err := req.ReadEntity(&shareReq); err != nil {
		writeFileError(resp, http.StatusBadRequest, "INVALID_REQUEST", fmt.Sprintf("Error reading request body: %v", err))
		return
	}

	if shareReq.BoxID == "" || shareReq.Path == "" {
		writeFileError(resp, http.StatusBadRequest, "INVALID_REQUEST", "Box ID and path are required")
		return
	}

	// Clean and validate the path
	cleanPath := filepath.Clean(shareReq.Path)
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}

	// Check if path starts with default share directory path
	if strings.HasPrefix(cleanPath, "/" + shareReq.BoxID + common.DefaultShareDirPath) {
		// Remove the prefix to get the relative path
		relativePath := strings.TrimPrefix(cleanPath, "/" + shareReq.BoxID + common.DefaultShareDirPath)
		relativePath = filepath.Clean(relativePath)
		if !strings.HasPrefix(relativePath, "/") {
			relativePath = "/" + relativePath
		}

		// Check if file already exists in share directory
		sharePath := filepath.Join(h.shareDir, shareReq.BoxID, relativePath)
		if _, err := os.Stat(sharePath); err == nil {
			// File exists, return success with file list
			fileList, err := getFileList(sharePath)
			if err != nil {
				logger.Printf("Error getting file list: %v", err)
			}
			resp.WriteAsJson(models.FileShareResponse{
				Success:  true,
				Message:  "File already exists",
				FileList: fileList,
			})
			return
		}
	}

	// Get box handler from request context
	boxHandler, ok := req.Attribute("boxHandler").(types.BoxHandler)
	if !ok {
		writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", "Box handler not found in request context")
		return
	}

	// Create share directory for the box if it doesn't exist
	boxShareDir := filepath.Join(h.shareDir, shareReq.BoxID)
	if err := os.MkdirAll(boxShareDir, 0755); err != nil {
		writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Failed to create share directory: %v", err))
		return
	}

	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(filepath.Join(boxShareDir, cleanPath))
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Failed to create target directory: %v", err))
		return
	}

	// Create a custom response writer that collects the archive data
	archiveData := &bytes.Buffer{}
	archiveResp := &restful.Response{
		ResponseWriter: &bufferResponseWriter{buffer: archiveData},
	}

	// Get file from container using GetArchive
	archiveReq := restful.NewRequest(req.Request.Clone(req.Request.Context()))
	archiveReq.Request.URL.Path = fmt.Sprintf("/api/v1/boxes/%s/archive", shareReq.BoxID)
	archiveReq.PathParameters()["id"] = shareReq.BoxID // Add boxId as path parameter
	q := archiveReq.Request.URL.Query()
	q.Set("path", cleanPath)
	archiveReq.Request.URL.RawQuery = q.Encode()

	boxHandler.GetArchive(archiveReq, archiveResp)

	// Extract the archive to the target directory
	if err := extractArchive(archiveData, targetDir); err != nil {
		writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Failed to extract archive: %v", err))
		return
	}

	// Get the list of shared files
	fileList, err := getFileList(filepath.Join(boxShareDir, cleanPath))
	if err != nil {
		logger.Printf("Error getting file list: %v", err)
	}

	resp.WriteAsJson(models.FileShareResponse{
		Success:  true,
		Message:  "File shared successfully",
		FileList: fileList,
	})
}

// bufferResponseWriter implements http.ResponseWriter to write to a buffer
type bufferResponseWriter struct {
	buffer *bytes.Buffer
}

func (w *bufferResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *bufferResponseWriter) Write(data []byte) (int, error) {
	return w.buffer.Write(data)
}

func (w *bufferResponseWriter) WriteHeader(statusCode int) {
	// Ignore status code as we're writing to a buffer
}

// extractArchive extracts a tar archive to the target directory
func extractArchive(archive io.Reader, targetDir string) error {
	// Create a new tar reader
	tr := tar.NewReader(archive)

	// Iterate through the archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("error reading tar header: %v", err)
		}

		// Construct the target path
		targetPath := filepath.Join(targetDir, header.Name)

		// Handle different types of files
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("error creating directory %s: %v", targetPath, err)
			}
		case tar.TypeReg, tar.TypeRegA:
			// Create parent directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("error creating parent directory for %s: %v", targetPath, err)
			}

			// Create the file
			f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("error creating file %s: %v", targetPath, err)
			}

			// Copy the file content
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("error copying file content for %s: %v", targetPath, err)
			}

			// Close the file
			if err := f.Close(); err != nil {
				return fmt.Errorf("error closing file %s: %v", targetPath, err)
			}
		case tar.TypeSymlink:
			// Create parent directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("error creating parent directory for symlink %s: %v", targetPath, err)
			}

			// Create the symlink
			if err := os.Symlink(header.Linkname, targetPath); err != nil {
				return fmt.Errorf("error creating symlink %s: %v", targetPath, err)
			}
		}
	}

	return nil
}

// getFileList returns a list of files in the given path
func getFileList(path string) ([]models.FileStat, error) {
	// First check if path is a file
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// If it's a file, return its info directly
	if !info.IsDir() {
		return []models.FileStat{
			{
				Name:    info.Name(),
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

	var files []models.FileStat
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, models.FileStat{
			Name:    info.Name(),
			Size:    info.Size(),
			Mode:    info.Mode().String(),
			ModTime: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
			Type:    getFileType(info),
			Mime:    getMimeType(filepath.Join(path, info.Name()), info),
		})
	}

	return files, nil
}
