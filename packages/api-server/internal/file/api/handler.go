package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/babelcloud/gbox/packages/api-server/pkg/file"
	"github.com/babelcloud/gbox/packages/api-server/internal/file/service"
	"github.com/emicklei/go-restful/v3"
)

// FileHandler handles file operations for the share directory
type FileHandler struct {
	service service.FileService
}

// NewFileHandler creates a new FileHandler
func NewFileHandler(service service.FileService) *FileHandler {
	return &FileHandler{
		service: service,
	}
}

// HeadFile handles HEAD requests to get file metadata
func (h *FileHandler) HeadFile(req *restful.Request, resp *restful.Response) {
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

	// Get file metadata
	stat, err := h.service.HeadFile(req.Request.Context(), cleanPath)
	if err != nil {
		writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error getting file metadata: %v", err))
		return
	}

	// Convert stat to JSON string
	statJSON, err := json.Marshal(stat)
	if err != nil {
		writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error marshaling stat: %v", err))
		return
	}

	// Set response headers
	resp.Header().Set("Content-Type", stat.Mime)
	resp.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))
	resp.Header().Set("X-Gbox-File-Stat", string(statJSON))

	resp.WriteHeader(http.StatusOK)
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

	// Get file content
	content, err := h.service.GetFile(req.Request.Context(), cleanPath)
	if err != nil {
		writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error getting file content: %v", err))
		return
	}
	defer content.Reader.Close()

	// Set response headers
	resp.Header().Set("Content-Type", content.MimeType)
	resp.Header().Set("Content-Length", fmt.Sprintf("%d", content.Size))

	// Copy file content to response
	if _, err := io.Copy(resp, content.Reader); err != nil {
		writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error copying file content: %v", err))
		return
	}
}

// HandleFileOperation handles file operations like reclaim and share
func (h *FileHandler) HandleFileOperation(req *restful.Request, resp *restful.Response) {
	operation := req.QueryParameter("operation")

	switch operation {
	case "reclaim":
		h.ReclaimFiles(req, resp)
	case "share":
		h.ShareFile(req, resp)
	default:
		writeFileError(resp, http.StatusBadRequest, "INVALID_REQUEST", "Invalid operation")
	}
}

// ReclaimFiles handles reclaiming files that haven't been accessed for more than 14 days
func (h *FileHandler) ReclaimFiles(req *restful.Request, resp *restful.Response) {
	response, err := h.service.ReclaimFiles(req.Request.Context())
	if err != nil {
		writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error reclaiming files: %v", err))
		return
	}

	resp.WriteAsJson(response)
}

// ShareFile handles sharing a file from a box to the share directory
func (h *FileHandler) ShareFile(req *restful.Request, resp *restful.Response) {
	var shareReq model.FileShareParams
	if err := req.ReadEntity(&shareReq); err != nil {
		writeFileError(resp, http.StatusBadRequest, "INVALID_REQUEST", fmt.Sprintf("Error reading request body: %v", err))
		return
	}

	if shareReq.BoxID == "" || shareReq.Path == "" {
		writeFileError(resp, http.StatusBadRequest, "INVALID_REQUEST", "Box ID and path are required")
		return
	}

	response, err := h.service.ShareFile(req.Request.Context(), shareReq.BoxID, shareReq.Path)
	if err != nil {
		writeFileError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error sharing file: %v", err))
		return
	}

	resp.WriteAsJson(response)
}

// writeFileError writes a structured error response
func writeFileError(resp *restful.Response, statusCode int, code, message string) {
	resp.WriteHeader(statusCode)
	resp.WriteAsJson(model.FileError{
		Code:    code,
		Message: message,
	})
}
