package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/babelcloud/gbox/packages/api-server/pkg/box"
)

const (
	GboxExtraLabelPrefix = "gbox.extra"
)

// Get implements Service.Get
func (s *Service) Get(ctx context.Context, id string) (*model.Box, error) {
	container, err := s.getContainerByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return containerToBox(container), nil
}

// List implements Service.List
func (s *Service) List(ctx context.Context, params *model.BoxListParams) (*model.BoxListResult, error) {
	// Build filter for gbox containers
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("%s=gbox", labelName))

	// Apply filters from params
	for _, filter := range params.Filters {
		switch filter.Field {
		case "id":
			// Use name filter for box ID (container name is gbox-{id})
			filterArgs.Add("name", fmt.Sprintf("gbox-%s", filter.Value))
		case "label":
			// Check if the value contains an equals sign
			if strings.Contains(filter.Value, "=") {
				// Format: label=key=value
				// Split into key and value
				labelParts := strings.Split(filter.Value, "=")
				if len(labelParts) != 2 {
					continue
				}
				key, val := labelParts[0], labelParts[1]
				// Add GboxExtraLabelPrefix prefix to the label key for filtering
				filterArgs.Add("label", fmt.Sprintf("%s.%s=%s", GboxExtraLabelPrefix, key, val))
			} else {
				// Format: label=key
				// Just check if the label exists
				filterArgs.Add("label", fmt.Sprintf("%s.%s", GboxExtraLabelPrefix, filter.Value))
			}
		case "ancestor":
			filterArgs.Add("ancestor", filter.Value)
		}
	}

	containers, err := s.client.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	boxes := make([]model.Box, 0, len(containers))
	for i := range containers {
		boxes = append(boxes, *containerToBox(&containers[i]))
	}

	return &model.BoxListResult{
		Boxes: boxes,
		Count: len(boxes),
	}, nil
}
