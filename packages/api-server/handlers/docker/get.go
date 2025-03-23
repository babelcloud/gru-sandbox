package docker

import (
	"fmt"
	"log"
	"net/http"

	"github.com/emicklei/go-restful/v3"
)

// handleGetBox handles getting a single box
func handleGetBox(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	log.Printf("Received get request for box: %s", boxID)

	box, err := h.getContainerByID(req.Request.Context(), boxID)
	if err != nil {
		if err.Error() == "box not found" {
			log.Printf("Box not found: %s", boxID)
			writeError(resp, http.StatusNotFound, "BOX_NOT_FOUND", fmt.Sprintf("Box not found: %s", boxID))
		} else if err.Error() == "box ID is required" {
			log.Printf("Invalid request: box ID is required")
			writeError(resp, http.StatusBadRequest, "INVALID_REQUEST", "Box ID is required")
		} else {
			log.Printf("Error getting container: %v", err)
			writeError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error getting container: %v", err))
		}
		return
	}

	// Convert container to box model
	boxModel := containerToBox(box)
	resp.WriteAsJson(boxModel)
}
