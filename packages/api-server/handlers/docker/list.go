package docker

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/emicklei/go-restful/v3"

	"github.com/babelcloud/gru-sandbox/packages/api-server/config"
	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
)

// handleListBoxes handles the list boxes operation
func handleListBoxes(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	logger := log.New()
	logger.Info("Starting to list boxes")

	// Build Docker filter args
	filterArgs := filters.NewArgs()
	// Add base filter for gbox containers
	filterArgs.Add("label", fmt.Sprintf("%s=gbox", GboxLabelName))
	filterArgs.Add("label", fmt.Sprintf("%s=%s", GboxLabelCompose, config.GetGboxLabelCompose()))
	logger.Debug("Initialized filter args with base filter: %v", filterArgs)

	// Get filters from query parameters
	queryFilters := req.QueryParameters("filter")
	logger.Debug("Received query filters: %v", queryFilters)

	for _, filter := range queryFilters {
		// Parse filter format: field=value
		// For label filters, value might contain multiple equals signs
		firstEquals := strings.Index(filter, "=")
		if firstEquals == -1 {
			logger.Debug("Invalid filter format (no equals sign), skipping: %s", filter)
			continue
		}
		field := filter[:firstEquals]
		value := filter[firstEquals+1:]

		switch field {
		case "id":
			// Use name filter for box ID (container name is gbox-{id})
			filterArgs.Add("name", fmt.Sprintf("gbox-%s", value))
			logger.Debug("Added ID filter using name: gbox-%s", value)
		case "label":
			// Check if the value contains an equals sign
			if strings.Contains(value, "=") {
				// Format: label=key=value
				// Split into key and value
				labelParts := strings.Split(value, "=")
				if len(labelParts) != 2 {
					logger.Debug("Invalid label format, skipping: %s", value)
					continue
				}
				key, val := labelParts[0], labelParts[1]
				// Add GboxExtraLabelPrefix prefix to the label key for filtering
				filterArgs.Add("label", fmt.Sprintf("%s.%s=%s", GboxExtraLabelPrefix, key, val))
				logger.Debug("Added label filter with value: %s.%s=%s", GboxExtraLabelPrefix, key, val)
			} else {
				// Format: label=key
				// Just check if the label exists
				filterArgs.Add("label", fmt.Sprintf("%s.%s", GboxExtraLabelPrefix, value))
				logger.Debug("Added label existence filter: %s.%s", GboxExtraLabelPrefix, value)
			}
		case "ancestor":
			filterArgs.Add("ancestor", value)
			logger.Debug("Added ancestor filter: %s", value)
		default:
			logger.Debug("Unknown filter field: %s", field)
		}
	}

	// Get containers with filters
	logger.Debug("Querying Docker with filters: %v", filterArgs)
	containerList, err := h.getContainers(req.Request.Context(), &filterArgs)
	if err != nil {
		logger.Error("Failed to list containers: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}
	logger.Debug("Retrieved %d containers from Docker", len(containerList))

	// Convert containers to boxes
	boxes := make([]models.Box, 0, len(containerList))
	for i, c := range containerList {
		logger.Debug("Processing container %d: ID=%s, State=%s, Image=%s, Labels=%v", i, c.ID, c.State, c.Image, c.Labels)
		box := containerToBox(c)
		logger.Debug("Converted container %s to box: %+v", c.ID, box)
		boxes = append(boxes, box)
	}

	// Return BoxListResponse
	response := models.BoxListResponse{
		Boxes: boxes,
	}
	logger.Debug("Sending response with %d boxes", len(boxes))
	resp.WriteAsJson(response)
}
