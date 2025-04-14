package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/babelcloud/gbox/packages/api-server/internal/box/service"
	"github.com/babelcloud/gbox/packages/api-server/pkg/logger"

	"github.com/emicklei/go-restful/v3"
)

var log = logger.New()

// Local constants replacing models.MediaType*
const (
	mediaTypeRawStream         = "application/vnd.gbox.raw-stream"
	mediaTypeMultiplexedStream = "application/vnd.gbox.multiplexed-stream"
)

// Local error struct replacing models.BoxError
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// BoxHandler handles HTTP requests for box operations
type BoxHandler struct {
	service service.BoxService
}

// NewBoxHandler creates a new BoxHandler
func NewBoxHandler(service service.BoxService) *BoxHandler {
	return &BoxHandler{
		service: service,
	}
}

// ListBoxes returns all boxes
func (h *BoxHandler) ListBoxes(req *restful.Request, resp *restful.Response) {
	// Parse query parameters into BoxListParams
	params := &model.BoxListParams{}

	// Get filters from query parameters
	queryFilters := req.QueryParameters("filter")
	for _, filter := range queryFilters {
		// Parse filter format: field=value
		parts := strings.SplitN(filter, "=", 2)
		if len(parts) != 2 {
			continue
		}
		params.Filters = append(params.Filters, model.Filter{
			Field:    parts[0],
			Operator: model.FilterOperatorEquals,
			Value:    parts[1],
		})
	}

	result, err := h.service.List(req.Request.Context(), params)
	if err != nil {
		writeError(resp, http.StatusInternalServerError, "ListBoxesError", err.Error())
		return
	}

	resp.WriteEntity(result)
}

// GetBox returns a box by ID
func (h *BoxHandler) GetBox(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	box, err := h.service.Get(req.Request.Context(), boxID)
	if err != nil {
		if err == service.ErrBoxNotFound {
			writeError(resp, http.StatusNotFound, "BoxNotFound", err.Error())
			return
		}
		writeError(resp, http.StatusInternalServerError, "GetBoxError", err.Error())
		return
	}

	resp.WriteEntity(box)
}

// CreateBox creates a new box
func (h *BoxHandler) CreateBox(req *restful.Request, resp *restful.Response) {
	// Read request body directly into the internal model type
	var createParams model.BoxCreateParams
	if err := req.ReadEntity(&createParams); err != nil {
		writeError(resp, http.StatusBadRequest, "InvalidRequest", err.Error())
		return
	}

	// Call the service method with the populated struct
	box, err := h.service.Create(req.Request.Context(), &createParams)
	if err != nil {
		writeError(resp, http.StatusInternalServerError, "CreateBoxError", err.Error())
		return
	}

	resp.WriteHeaderAndEntity(http.StatusCreated, box)
}

// DeleteBox deletes a box by ID
func (h *BoxHandler) DeleteBox(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	if boxID == "" {
		writeError(resp, http.StatusBadRequest, "InvalidRequest", "Box ID is required")
		return
	}

	// Parse request body directly into model.BoxDeleteParams
	var deleteParams model.BoxDeleteParams
	if err := req.ReadEntity(&deleteParams); err != nil {
		// If no request body or error parsing, continue with default force=false
		// (Assuming ReadEntity handles empty body gracefully or returns specific error)
		// If ReadEntity fails on empty body, add check: if err != nil && err != io.EOF { ... }
		deleteParams.Force = false // Default
	}

	result, err := h.service.Delete(req.Request.Context(), boxID, &deleteParams)
	if err != nil {
		if err == service.ErrBoxNotFound {
			writeError(resp, http.StatusNotFound, "BoxNotFound", err.Error())
			return
		}
		writeError(resp, http.StatusInternalServerError, "DeleteBoxError", err.Error())
		return
	}
	// Write the internal model.BoxDeleteResult directly
	resp.WriteHeaderAndEntity(http.StatusOK, result)
}

// DeleteBoxes deletes all boxes
func (h *BoxHandler) DeleteBoxes(req *restful.Request, resp *restful.Response) {
	// Parse request body directly into model.BoxesDeleteParams
	var deleteParams model.BoxesDeleteParams
	if err := req.ReadEntity(&deleteParams); err != nil {
		// If no request body or error parsing, continue with default force=false
		deleteParams.Force = false // Default
	}

	result, err := h.service.DeleteAll(req.Request.Context(), &deleteParams)
	if err != nil {
		writeError(resp, http.StatusInternalServerError, "DeleteBoxesError", err.Error())
		return
	}
	// Write the internal model.BoxesDeleteResult directly
	resp.WriteHeaderAndEntity(http.StatusOK, result)
}

// ReclaimBoxes reclaims inactive boxes
func (h *BoxHandler) ReclaimBoxes(req *restful.Request, resp *restful.Response) {
	result, err := h.service.Reclaim(req.Request.Context())
	if err != nil {
		writeError(resp, http.StatusInternalServerError, "ReclaimBoxesError", err.Error())
		return
	}
	// Write the internal model.BoxReclaimResult directly
	resp.WriteHeaderAndEntity(http.StatusOK, result)
}

// ExecBox executes a command in a box
func (h *BoxHandler) ExecBox(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	// Read request body directly into model.BoxExecParams
	var execReq model.BoxExecParams
	if err := req.ReadEntity(&execReq); err != nil {
		writeError(resp, http.StatusBadRequest, "InvalidRequest", err.Error())
		return
	}

	// Check if we need to hijack the connection
	upgrade := req.HeaderParameter("Upgrade")
	connection := req.HeaderParameter("Connection")
	accept := req.HeaderParameter("Accept")
	if accept == "" {
		accept = mediaTypeMultiplexedStream // Use local constant
	}

	// Validate Accept header
	if accept != mediaTypeRawStream && accept != mediaTypeMultiplexedStream { // Use local constants
		writeError(resp, http.StatusNotAcceptable, "UnsupportedMediaType",
			fmt.Sprintf("Unsupported Accept header: %s", accept))
		return
	}

	// Hijack the connection if needed
	httpResp := resp.ResponseWriter
	hijacker, ok := httpResp.(http.Hijacker)
	if !ok {
		writeError(resp, http.StatusInternalServerError, "HijackError",
			"response does not support hijacking")
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		writeError(resp, http.StatusInternalServerError, "HijackError",
			fmt.Sprintf("failed to hijack connection: %v", err))
		return
	}
	defer clientConn.Close()

	// Set the connection in the execReq (already read into this struct)
	execReq.Conn = clientConn

	// Write HTTP response headers directly to the hijacked connection
	writeResponseHeaders(execReq.Conn, upgrade, connection, execReq.TTY)

	// Execute command and handle streaming
	result, err := h.service.Exec(req.Request.Context(), boxID, &execReq)
	if err != nil {
		if err == service.ErrBoxNotFound {
			// Cannot write standard error after hijack, maybe log?
			log.Errorf("Box not found for exec: %s", boxID)
			// Consider closing the connection? clientConn.Close()
			return
		}
		// Cannot write standard error after hijack, maybe log?
		log.Errorf("ExecBoxError for box %s: %v", boxID, err)
		// Consider closing the connection? clientConn.Close()
		return
	}

	// Cannot write entity after hijack. Result (exit code) handling might need adjustment.
	// Maybe log the exit code?
	if result != nil {
		log.Infof("Exec finished for box %s with exit code: %d", boxID, result.ExitCode)
	} else {
		log.Warnf("Exec finished for box %s but result was nil", boxID)
	}
	// resp.WriteEntity(result) // Cannot do this after hijack
}

// RunBox runs a command in a box
func (h *BoxHandler) RunBox(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	// Read request body directly into model.BoxRunParams
	var runReq model.BoxRunParams
	if err := req.ReadEntity(&runReq); err != nil {
		writeError(resp, http.StatusBadRequest, "InvalidRequest", err.Error())
		return
	}

	result, err := h.service.Run(req.Request.Context(), boxID, &runReq)
	if err != nil {
		if err == service.ErrBoxNotFound {
			writeError(resp, http.StatusNotFound, "BoxNotFound", err.Error())
			return
		}
		writeError(resp, http.StatusInternalServerError, "RunBoxError", err.Error())
		return
	}

	// Write the internal model.BoxRunResult directly
	resp.WriteEntity(result)
}

// StartBox starts a stopped box
func (h *BoxHandler) StartBox(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	result, err := h.service.Start(req.Request.Context(), boxID)
	if err != nil {
		if err == service.ErrBoxNotFound {
			writeError(resp, http.StatusNotFound, "BoxNotFound", err.Error())
			return
		}
		writeError(resp, http.StatusInternalServerError, "StartBoxError", err.Error())
		return
	}
	// Write the internal model.BoxStartResult directly
	resp.WriteHeaderAndEntity(http.StatusOK, result)
}

// StopBox stops a running box
func (h *BoxHandler) StopBox(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	result, err := h.service.Stop(req.Request.Context(), boxID)
	if err != nil {
		if err == service.ErrBoxNotFound {
			writeError(resp, http.StatusNotFound, "BoxNotFound", err.Error())
			return
		}
		writeError(resp, http.StatusInternalServerError, "StopBoxError", err.Error())
		return
	}
	// Write the internal model.BoxStopResult directly
	resp.WriteHeaderAndEntity(http.StatusOK, result)
}

// GetArchive gets files from box as tar archive
func (h *BoxHandler) GetArchive(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	path := req.QueryParameter("path")

	archiveReq := &model.BoxArchiveGetParams{
		Path: path,
	}

	archiveResp, archive, err := h.service.GetArchive(req.Request.Context(), boxID, archiveReq)
	if err != nil {
		if err == service.ErrBoxNotFound {
			writeError(resp, http.StatusNotFound, "BoxNotFound", err.Error())
			return
		}
		writeError(resp, http.StatusInternalServerError, "GetArchiveError", err.Error())
		return
	}
	defer archive.Close()

	// Convert archiveResp to JSON string
	statJSON, err := json.Marshal(archiveResp)
	if err != nil {
		writeError(resp, http.StatusInternalServerError, "GetArchiveError", fmt.Sprintf("Failed to marshal stat: %v", err))
		return
	}

	// Set response headers
	resp.Header().Set("Content-Type", "application/x-tar")
	resp.Header().Set("X-Gbox-Path-Stat", string(statJSON))
	resp.Header().Set("Content-Length", fmt.Sprintf("%d", archiveResp.Size))
	resp.Header().Set("Last-Modified", archiveResp.Mtime)

	// Copy the archive to response
	if _, err := io.Copy(resp.ResponseWriter, archive); err != nil {
		writeError(resp, http.StatusInternalServerError, "GetArchiveError", fmt.Sprintf("Failed to copy archive: %v", err))
		return
	}
}

// HeadArchive gets metadata about files in box
func (h *BoxHandler) HeadArchive(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	path := req.QueryParameter("path")

	headReq := &model.BoxArchiveHeadParams{
		Path: path,
	}

	stat, err := h.service.HeadArchive(req.Request.Context(), boxID, headReq)
	if err != nil {
		if err == service.ErrBoxNotFound {
			writeError(resp, http.StatusNotFound, "BoxNotFound", err.Error())
			return
		}
		writeError(resp, http.StatusInternalServerError, "HeadArchiveError", err.Error())
		return
	}

	// Convert stat to JSON string
	statJSON, err := json.Marshal(stat)
	if err != nil {
		writeError(resp, http.StatusInternalServerError, "HeadArchiveError", fmt.Sprintf("Failed to marshal stat: %v", err))
		return
	}

	// Set response headers
	resp.Header().Set("Content-Type", "application/json")
	resp.Header().Set("X-Gbox-Path-Stat", string(statJSON))
	resp.WriteHeader(http.StatusOK)
}

// ExtractArchive extracts tar archive to box
func (h *BoxHandler) ExtractArchive(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	path := req.QueryParameter("path")

	// Read request body
	content, err := io.ReadAll(req.Request.Body)
	if err != nil {
		writeError(resp, http.StatusBadRequest, "InvalidRequest", "Failed to read request body")
		return
	}

	extractParams := &model.BoxArchiveExtractParams{
		Path:    path,
		Content: content,
	}

	if err := h.service.ExtractArchive(req.Request.Context(), boxID, extractParams); err != nil {
		if err == service.ErrBoxNotFound {
			writeError(resp, http.StatusNotFound, "BoxNotFound", err.Error())
			return
		}
		writeError(resp, http.StatusInternalServerError, "ExtractArchiveError", err.Error())
		return
	}
	resp.WriteHeader(http.StatusOK)
}

// writeError writes an error response using local apiError struct
func writeError(resp *restful.Response, status int, code, message string) {
	resp.WriteHeaderAndEntity(status, &apiError{
		Code:    code,
		Message: message,
	})
}

// writeResponseHeaders writes HTTP response headers based on upgrade and TTY status
func writeResponseHeaders(w io.Writer, upgrade, connection string, tty bool) {
	if upgrade == "tcp" && connection == "Upgrade" {
		fmt.Fprintf(w, "HTTP/1.1 101 UPGRADED\r\n")
		fmt.Fprintf(w, "Content-Type: %s\r\n", getContentType(tty)) // Use local helper
		fmt.Fprintf(w, "Connection: Upgrade\r\n")
		fmt.Fprintf(w, "Upgrade: tcp\r\n")
		log.Debugf("Protocol upgrade requested, using %s", getStreamType(tty)) // Use local helper
	} else {
		fmt.Fprintf(w, "HTTP/1.1 200 OK\r\n")
		fmt.Fprintf(w, "Content-Type: %s\r\n", getContentType(tty))                  // Use local helper
		log.Debugf("No protocol upgrade requested, using %s", getStreamType(tty)) // Use local helper
	}
	fmt.Fprintf(w, "\r\n")

	// Flush response headers if possible
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// getContentType returns the appropriate content type based on TTY status using local constants
func getContentType(tty bool) string {
	if tty {
		return mediaTypeRawStream
	}
	return mediaTypeMultiplexedStream
}

// getStreamType returns a human-readable stream type description using local constants
func getStreamType(tty bool) string {
	if tty {
		return "raw stream (TTY mode)"
	}
	return "multiplexed stream (non-TTY mode)"
}
