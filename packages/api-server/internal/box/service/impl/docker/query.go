package docker

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/go-connections/nat"

	model "github.com/babelcloud/gbox/packages/api-server/pkg/box"
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

// GetExternalPort implements Service.GetExternalPort
func (s *Service) GetExternalPort(ctx context.Context, id string, internalPort int) (int, error) {
	// Use the new helper function to get container details
	containerJSON, err := s.inspectContainerByID(ctx, id)
	if err != nil {
		// inspectContainerByID already uses handleContainerError
		return 0, err
	}

	if containerJSON.NetworkSettings == nil || containerJSON.NetworkSettings.Ports == nil {
		return 0, fmt.Errorf("no network settings or ports found for box %s", id)
	}

	// Construct the nat.Port object (defaulting to tcp)
	internalNatPort, err := nat.NewPort("tcp", strconv.Itoa(internalPort))
	if err != nil {
		// This error occurs if the port string is invalid, which shouldn't happen with strconv.Itoa
		return 0, fmt.Errorf("invalid internal port %d: %w", internalPort, err)
	}

	// Look up the port binding in the map
	portBindings, ok := containerJSON.NetworkSettings.Ports[internalNatPort]
	if !ok || len(portBindings) == 0 {
		// Port not found or not published
		// Check if the port *was* exposed but just not published
		if _, exposed := containerJSON.Config.ExposedPorts[internalNatPort]; exposed {
			return 0, fmt.Errorf("internal port %d is exposed but not published for box %s", internalPort, id)
		}
		return 0, fmt.Errorf("internal port %d not exposed or mapped for box %s", internalPort, id)
	}

	// Use the first available binding.
	hostPortStr := portBindings[0].HostPort
	hostPort, err := strconv.Atoi(hostPortStr)
	if err != nil {
		// This should ideally not happen if Docker API returns valid data
		return 0, fmt.Errorf("failed to parse host port \"%s\" for box %s: %w", hostPortStr, id, err)
	}

	return hostPort, nil
}
