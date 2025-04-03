package docker

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"github.com/emicklei/go-restful/v3"

	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
)

// handleStartBox handles starting a stopped box
func handleStartBox(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	logger := log.New()
	boxID := req.PathParameter("id")
	logger.Info("Starting box: %s", boxID)

	// Parse request body
	var startReq models.BoxStartRequest
	if err := req.ReadEntity(&startReq); err != nil {
		// If no request body, continue with default empty mounts
		startReq.Mounts = []models.Mount{}
	}

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

	// Process mount configurations
	var mounts []mount.Mount
	for _, m := range startReq.Mounts {
		mountType := mount.TypeBind
		switch m.Type {
		case models.MountTypeBind:
			mountType = mount.TypeBind
		case models.MountTypeVolume:
			mountType = mount.TypeVolume
		case models.MountTypeTmpfs:
			mountType = mount.TypeTmpfs
		default:
			logger.Error("Invalid mount type: %s", m.Type)
			resp.WriteError(http.StatusBadRequest, fmt.Errorf("invalid mount type: %s", m.Type))
			return
		}

		// Validate source path for bind mounts
		if m.Type == models.MountTypeBind {
			if !filepath.IsAbs(m.Source) {
				logger.Error("Source path must be absolute: %s", m.Source)
				resp.WriteError(http.StatusBadRequest, fmt.Errorf("source path must be absolute: %s", m.Source))
				return
			}

			// Check if source path exists
			if _, err := os.Stat(m.Source); err != nil {
				logger.Error("Source path does not exist: %s", m.Source)
				resp.WriteError(http.StatusBadRequest, fmt.Errorf("source path does not exist: %s", m.Source))
				return
			}
		}

		// Add mount configuration
		mountConfig := mount.Mount{
			Type:     mountType,
			Source:   m.Source,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		}

		// Set consistency if specified
		if m.Consistency != "" {
			switch m.Consistency {
			case "default":
				mountConfig.Consistency = mount.ConsistencyDefault
			case "cached":
				mountConfig.Consistency = mount.ConsistencyCached
			case "delegated":
				mountConfig.Consistency = mount.ConsistencyDelegated
			default:
				logger.Error("Invalid mount consistency: %s", m.Consistency)
				resp.WriteError(http.StatusBadRequest, fmt.Errorf("invalid mount consistency: %s", m.Consistency))
				return
			}
		}

		mounts = append(mounts, mountConfig)
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
