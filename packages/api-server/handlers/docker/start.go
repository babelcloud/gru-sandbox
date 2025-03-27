package docker

import (
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/emicklei/go-restful/v3"

	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
)

// handleStartBox handles starting a stopped box
func handleStartBox(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	logger := log.New()
	boxID := req.PathParameter("id")
	logger.Info("Starting box: %s", boxID)

	containerSummary, err := h.getContainerByID(req.Request.Context(), boxID)
	if err != nil {
		if err.Error() == "box not found" {
			logger.Error("Box not found: %s", boxID)
			resp.WriteErrorString(http.StatusNotFound, err.Error())
		} else if err.Error() == "box ID is required" {
			logger.Error("Invalid request: box ID is required")
			resp.WriteErrorString(http.StatusBadRequest, err.Error())
		} else {
			logger.Error("Error getting container: %v", err)
			resp.WriteError(http.StatusInternalServerError, err)
		}
		return
	}

	// Check if container is already running
	if containerSummary.State == "running" {
		logger.Info("Box is already running: %s", boxID)
		resp.WriteErrorString(http.StatusBadRequest, "box is already running")
		return
	}

	// Log container details before starting
	logger.Debug("Container details - ID: %s, State: %s",
		containerSummary.ID, containerSummary.State)

	// Start the container
	logger.Debug("Starting container with ID: %s", containerSummary.ID)
	err = h.client.ContainerStart(req.Request.Context(), containerSummary.ID, types.ContainerStartOptions{})
	if err != nil {
		logger.Error("Error starting container: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Verify container is running
	inspect, err := h.client.ContainerInspect(req.Request.Context(), containerSummary.ID)
	if err != nil {
		logger.Error("Error inspecting container after start: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	if !inspect.State.Running {
		logger.Error("Container failed to start - State: %s, ExitCode: %d, Error: %s",
			inspect.State.Status, inspect.State.ExitCode, inspect.State.Error)
		resp.WriteErrorString(http.StatusInternalServerError, "container failed to start")
		return
	}

	logger.Info("Box started successfully: %s", boxID)
	resp.WriteHeaderAndEntity(http.StatusOK, models.BoxStartResponse{
		Success: true,
		Message: fmt.Sprintf("Box %s started successfully", boxID),
	})
}
