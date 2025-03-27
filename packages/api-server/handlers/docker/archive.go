package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/emicklei/go-restful/v3"
)

// handleHeadArchive handles getting metadata about a resource in the container's filesystem
func handleHeadArchive(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	path := req.QueryParameter("path")

	log.Printf("Received head archive request for box: %s, path: %s", boxID, path)

	if path == "" {
		log.Printf("Invalid request: path is required")
		writeError(resp, http.StatusBadRequest, "INVALID_REQUEST", "Path is required")
		return
	}

	// Get the container
	container, err := h.getContainerByID(req.Request.Context(), boxID)
	if err != nil {
		if err.Error() == "box not found" {
			log.Printf("Box not found: %s", boxID)
			writeError(resp, http.StatusNotFound, "BOX_NOT_FOUND", fmt.Sprintf("Box not found: %s", boxID))
		} else {
			log.Printf("Error getting container: %v", err)
			writeError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error getting container: %v", err))
		}
		return
	}

	// Get the file/directory metadata
	stat, err := h.client.ContainerStatPath(req.Request.Context(), container.ID, path)
	if err != nil {
		log.Printf("Error getting path metadata: %v", err)
		writeError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error getting path metadata: %v", err))
		return
	}

	// Convert stat to JSON string
	statJSON, err := json.Marshal(stat)
	if err != nil {
		log.Printf("Error marshaling stat: %v", err)
		writeError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error marshaling stat: %v", err))
		return
	}

	// Set response headers
	resp.Header().Set("Content-Type", "application/json")
	resp.Header().Set("X-Gbox-Path-Stat", string(statJSON))
	resp.WriteHeader(http.StatusOK)
}

// handleGetArchive handles getting a tar archive of a resource in the container's filesystem
func handleGetArchive(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	path := req.QueryParameter("path")

	log.Printf("Received get archive request for box: %s, path: %s", boxID, path)

	if path == "" {
		log.Printf("Invalid request: path is required")
		writeError(resp, http.StatusBadRequest, "INVALID_REQUEST", "Path is required")
		return
	}

	// Get the container
	container, err := h.getContainerByID(req.Request.Context(), boxID)
	if err != nil {
		if err.Error() == "box not found" {
			log.Printf("Box not found: %s", boxID)
			writeError(resp, http.StatusNotFound, "BOX_NOT_FOUND", fmt.Sprintf("Box not found: %s", boxID))
		} else {
			log.Printf("Error getting container: %v", err)
			writeError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error getting container: %v", err))
		}
		return
	}

	// Get the archive
	archive, stat, err := h.client.CopyFromContainer(req.Request.Context(), container.ID, path)
	if err != nil {
		log.Printf("Error getting archive: %v", err)
		writeError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error getting archive: %v", err))
		return
	}
	defer archive.Close()

	// Convert stat to JSON string
	statJSON, err := json.Marshal(stat)
	if err != nil {
		log.Printf("Error marshaling stat: %v", err)
		writeError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error marshaling stat: %v", err))
		return
	}

	// Set response headers
	resp.Header().Set("Content-Type", "application/x-tar")
	resp.Header().Set("X-Gbox-Path-Stat", string(statJSON))

	// Copy the archive to response
	if _, err := io.Copy(resp, archive); err != nil {
		log.Printf("Error copying archive to response: %v", err)
		writeError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error copying archive: %v", err))
		return
	}
}

// handleExtractArchive handles extracting a tar archive to a directory in the container
func handleExtractArchive(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	path := req.QueryParameter("path")
	noOverwriteDirNonDir := req.QueryParameter("noOverwriteDirNonDir") == "1"
	copyUIDGID := req.QueryParameter("copyUIDGID") == "1"

	log.Printf("Received extract archive request for box: %s, path: %s", boxID, path)

	if path == "" {
		log.Printf("Invalid request: path is required")
		writeError(resp, http.StatusBadRequest, "INVALID_REQUEST", "Path is required")
		return
	}

	// Get the container
	container, err := h.getContainerByID(req.Request.Context(), boxID)
	if err != nil {
		if err.Error() == "box not found" {
			log.Printf("Box not found: %s", boxID)
			writeError(resp, http.StatusNotFound, "BOX_NOT_FOUND", fmt.Sprintf("Box not found: %s", boxID))
		} else {
			log.Printf("Error getting container: %v", err)
			writeError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error getting container: %v", err))
		}
		return
	}

	// Extract the archive
	err = h.client.CopyToContainer(req.Request.Context(), container.ID, path, req.Request.Body, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: !noOverwriteDirNonDir,
		CopyUIDGID:                copyUIDGID,
	})
	if err != nil {
		log.Printf("Error extracting archive: %v", err)
		writeError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error extracting archive: %v", err))
		return
	}

	resp.WriteHeader(http.StatusOK)
}
