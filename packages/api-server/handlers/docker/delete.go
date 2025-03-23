package docker

import (
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/emicklei/go-restful/v3"

	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
)

// handleDeleteBox handles the delete box operation
func handleDeleteBox(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	if boxID == "" {
		resp.WriteHeader(http.StatusBadRequest)
		resp.WriteAsJson(models.BoxDeleteResponse{
			Message: "Box ID is required",
		})
		return
	}

	// Get box by ID
	box, err := h.getContainerByID(req.Request.Context(), boxID)
	if err != nil {
		if err.Error() == "box not found" {
			resp.WriteHeader(http.StatusNotFound)
			resp.WriteAsJson(models.BoxDeleteResponse{
				Message: "Box not found",
			})
		} else {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.WriteAsJson(models.BoxDeleteResponse{
				Message: err.Error(),
			})
		}
		return
	}

	// Parse request body for force option
	var deleteReq models.BoxDeleteRequest
	if err := req.ReadEntity(&deleteReq); err != nil {
		// If no request body, continue with default force=false
		deleteReq.Force = false
	}

	// Stop box if running and not force
	if box.State == "running" && !deleteReq.Force {
		if err := h.client.ContainerStop(req.Request.Context(), box.ID, container.StopOptions{}); err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.WriteAsJson(models.BoxDeleteResponse{
				Message: "Failed to stop box: " + err.Error(),
			})
			return
		}
	}

	// Remove box with force option
	if err := h.client.ContainerRemove(req.Request.Context(), box.ID, container.RemoveOptions{
		Force: deleteReq.Force,
	}); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.WriteAsJson(models.BoxDeleteResponse{
			Message: "Failed to remove box: " + err.Error(),
		})
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.WriteAsJson(models.BoxDeleteResponse{
		Message: "Box deleted successfully",
	})
}

// handleDeleteAllBoxes handles the delete all boxes operation
func handleDeleteAllBoxes(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	logger := log.New()
	logger.Info("Starting to delete all boxes")

	// Parse request body
	var deleteReq models.BoxesDeleteRequest
	if err := req.ReadEntity(&deleteReq); err != nil {
		logger.Error("Failed to parse request body: %v", err)
		resp.WriteHeader(http.StatusBadRequest)
		resp.WriteAsJson(models.BoxesDeleteResponse{
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Build Docker filter args
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("%s=gbox", GboxLabelName))
	logger.Debug("Added base filter for gbox label: %v", filterArgs)

	// Get containers with filters
	logger.Debug("Querying Docker with filters: %v", filterArgs)
	containerList, err := h.client.ContainerList(req.Request.Context(), container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		logger.Error("Failed to list containers: %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.WriteAsJson(models.BoxesDeleteResponse{
			Message: "Failed to list boxes: " + err.Error(),
		})
		return
	}
	logger.Debug("Retrieved %d containers from Docker", len(containerList))

	if len(containerList) == 0 {
		logger.Info("No boxes to delete")
		resp.WriteHeader(http.StatusOK)
		resp.WriteAsJson(models.BoxesDeleteResponse{
			Count:   0,
			Message: "No boxes to delete",
		})
		return
	}

	// Collect box IDs
	boxIDs := make([]string, len(containerList))
	for i, box := range containerList {
		boxIDs[i] = box.ID
	}

	// Stop and remove all boxes
	for _, box := range containerList {
		logger.Debug("Processing box %s: State=%s", box.ID, box.State)
		// Stop box if running and not force
		if box.State == "running" && !deleteReq.Force {
			logger.Debug("Stopping box %s", box.ID)
			if err := h.client.ContainerStop(req.Request.Context(), box.ID, container.StopOptions{}); err != nil {
				logger.Error("Failed to stop box %s: %v", box.ID, err)
				resp.WriteHeader(http.StatusInternalServerError)
				resp.WriteAsJson(models.BoxesDeleteResponse{
					Message: "Failed to stop box " + box.ID + ": " + err.Error(),
				})
				return
			}
		}

		// Remove box with force option
		logger.Debug("Removing box %s", box.ID)
		if err := h.client.ContainerRemove(req.Request.Context(), box.ID, container.RemoveOptions{
			Force: deleteReq.Force,
		}); err != nil {
			logger.Error("Failed to remove box %s: %v", box.ID, err)
			resp.WriteHeader(http.StatusInternalServerError)
			resp.WriteAsJson(models.BoxesDeleteResponse{
				Message: "Failed to remove box " + box.ID + ": " + err.Error(),
			})
			return
		}
	}

	logger.Info("Successfully deleted %d boxes", len(containerList))
	resp.WriteHeader(http.StatusOK)
	resp.WriteAsJson(models.BoxesDeleteResponse{
		Count:   len(containerList),
		Message: "All boxes deleted successfully",
		IDs:     boxIDs,
	})
}
