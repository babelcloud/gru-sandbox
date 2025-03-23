package docker

import (
	"log"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/emicklei/go-restful/v3"
)

// handleStartBox handles starting a stopped box
func handleStartBox(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	log.Printf("Received start request for box: %s", boxID)

	containerSummary, err := h.getContainerByID(req.Request.Context(), boxID)
	if err != nil {
		if err.Error() == "box not found" {
			log.Printf("Box not found: %s", boxID)
			resp.WriteErrorString(http.StatusNotFound, err.Error())
		} else if err.Error() == "box ID is required" {
			log.Printf("Invalid request: box ID is required")
			resp.WriteErrorString(http.StatusBadRequest, err.Error())
		} else {
			log.Printf("Error getting container: %v", err)
			resp.WriteError(http.StatusInternalServerError, err)
		}
		return
	}

	// Check if container is already running
	if containerSummary.State == "running" {
		log.Printf("Box is already running: %s", boxID)
		resp.WriteErrorString(http.StatusBadRequest, "box is already running")
		return
	}

	// Start the container
	err = h.client.ContainerStart(req.Request.Context(), containerSummary.ID, container.StartOptions{})
	if err != nil {
		log.Printf("Error starting container: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	log.Printf("Box started successfully: %s", boxID)
	resp.WriteHeaderAndEntity(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Box started successfully",
	})
}
