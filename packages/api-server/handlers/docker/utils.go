package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

// getContainerByID finds a container by its ID label
func (h *DockerBoxHandler) getContainerByID(ctx context.Context, boxID string) (*types.Container, error) {
	if boxID == "" {
		return nil, fmt.Errorf("box ID is required")
	}

	// Create a filter to find the container by ID label
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("%s=%s", GboxLabelID, boxID))

	boxes, err := h.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	if len(boxes) == 0 {
		return nil, fmt.Errorf("box not found")
	}

	return &boxes[0], nil
}

// getAllContainers finds all containers with gbox label
func (h *DockerBoxHandler) getAllContainers(ctx context.Context) ([]types.Container, error) {
	logger := log.New()
	logger.Debug("Getting all containers")

	// Create a filter to only list gbox containers
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("%s=gbox", GboxLabelName))
	logger.Debug("Added base filter for gbox label: %v", filterArgs)

	containers, err := h.client.ContainerList(ctx, types.ContainerListOptions{
		All:     true, // Include stopped containers
		Filters: filterArgs,
	})
	if err != nil {
		logger.Error("Failed to list containers: %v", err)
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	logger.Debug("Retrieved %d containers from Docker", len(containers))
	return containers, nil
}

// extractExtraLabels extracts extra labels from container labels
func extractExtraLabels(labels map[string]string) map[string]string {
	extraLabels := make(map[string]string)
	for k, v := range labels {
		if strings.HasPrefix(k, GboxExtraLabelPrefix+".") {
			// Remove the prefix to get the original key
			key := strings.TrimPrefix(k, GboxExtraLabelPrefix+".")
			extraLabels[key] = v
		}
	}
	return extraLabels
}

// mapContainerState maps container state to box status
func mapContainerState(state string) string {
	switch state {
	case "running":
		return "running"
	case "exited":
		return "stopped"
	case "created":
		return "created"
	case "paused":
		return "paused"
	default:
		return "unknown"
	}
}

// containerToBox converts a container to a box model
func containerToBox(c interface{}) models.Box {
	var id, status, image string
	var labels map[string]string

	logger := log.New()
	logger.Debug("Converting container to box, type: %T", c)

	switch c := c.(type) {
	case types.ContainerJSON:
		logger.Debug("Handling InspectResponse")
		id = c.Config.Labels[GboxLabelID]
		status = mapContainerState(c.State.Status)
		image = c.Config.Image
		labels = c.Config.Labels
		logger.Debug("ContainerJSON labels: %v", labels)
	case types.Container:
		logger.Debug("Handling Summary")
		id = c.Labels[GboxLabelID]
		status = mapContainerState(c.State)
		image = c.Image
		labels = c.Labels
		logger.Debug("Container labels: %v", labels)
	case *types.Container:
		logger.Debug("Handling *Summary")
		id = c.Labels[GboxLabelID]
		status = mapContainerState(c.State)
		image = c.Image
		labels = c.Labels
		logger.Debug("*Container labels: %v", labels)
	default:
		logger.Debug("Unknown type: %T", c)
	}

	logger.Debug("Converted container: id=%s, status=%s, image=%s, labels=%v", id, status, image, labels)

	return models.Box{
		ID:          id,
		Status:      status,
		Image:       image,
		ExtraLabels: extractExtraLabels(labels),
	}
}
