package docker

import (
	"fmt"
	"net/http"

	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
	"github.com/docker/docker/api/types/container"
	"github.com/emicklei/go-restful/v3"
)

// handleStopBox handles stopping a running box
func handleStopBox(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	logger := log.New()
	logger.Info("Received stop request for box: %s", boxID)

	containerSummary, err := h.getContainerByID(req.Request.Context(), boxID)
	if err != nil {
		if err.Error() == "box not found" {
			logger.Info("Box not found: %s", boxID)
			resp.WriteErrorString(http.StatusNotFound, err.Error())
		} else if err.Error() == "box ID is required" {
			logger.Info("Invalid request: box ID is required")
			resp.WriteErrorString(http.StatusBadRequest, err.Error())
		} else {
			logger.Error("Error getting container: %v", err)
			resp.WriteError(http.StatusInternalServerError, err)
		}
		return
	}

	// Check if container is already stopped
	if containerSummary.State == "exited" {
		logger.Info("Box is already stopped: %s", boxID)
		resp.WriteErrorString(http.StatusBadRequest, "box is already stopped")
		return
	}

	// Stop the container with a timeout of 10 seconds
	timeout := 10
	err = h.client.ContainerStop(req.Request.Context(), containerSummary.ID, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		logger.Error("Error stopping container: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	logger.Info("Box stopped successfully: %s", boxID)
	resp.WriteAsJson(models.BoxStopResponse{
		Success: true,
		Message: fmt.Sprintf("Box %s stopped successfully", boxID),
	})
}
