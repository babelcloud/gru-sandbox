package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/emicklei/go-restful/v3"

	"github.com/babelcloud/gru-sandbox/packages/api-server/config"
	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/common"
	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
)

var logger = log.New()

// handleCreateBox handles the create box operation
func handleCreateBox(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	var boxReq models.BoxCreateRequest
	if err := req.ReadEntity(&boxReq); err != nil {
		logger.Error("Error reading request body: %v", err)
		resp.WriteError(http.StatusBadRequest, err)
		return
	}

	logger.Info("Received create request with image: %q", boxReq.Image)

	// Get image name
	img := common.GetImage(boxReq.Image)
	logger.Info("Checking image: %q", img)

	// Check if image exists
	_, _, err := h.client.ImageInspectWithRaw(req.Request.Context(), img)
	if err == nil {
		logger.Info("Using existing image: %q", img)
	} else {
		logger.Info("Image %q not found, pulling", img)
		if err := pullImage(h, req, resp, img, boxReq.ImagePullSecret); err != nil {
			return
		}
	}

	boxID := common.GenerateBoxID()
	containerName := fmt.Sprintf("gbox-%s", boxID)

	// Prepare labels
	labels := prepareLabels(boxID, &boxReq)

	// Get share directory from config
	fileConfig := config.NewFileConfig().(*config.FileConfig)
	if err := fileConfig.Initialize(logger); err != nil {
		logger.Error("Error initializing file config: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Create share directory for the box
	hostShareDir := filepath.Join(fileConfig.GetHostShareDir(), boxID)
	shareDir := filepath.Join(fileConfig.GetFileShareDir(), boxID)
	if err := os.MkdirAll(shareDir, 0755); err != nil {
		logger.Error("Error creating share directory: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Prepare volume mounts
	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: hostShareDir,
			Target: common.DefaultShareDirPath,
		},
	}

	// Create container
	containerResp, err := h.client.ContainerCreate(
		req.Request.Context(),
		&container.Config{
			Image:      img,
			Cmd:        common.GetCommand(boxReq.Cmd, boxReq.Args),
			Env:        common.GetEnvVars(boxReq.Env),
			WorkingDir: boxReq.WorkingDir,
			Labels:     labels,
		},
		&container.HostConfig{
			Mounts: mounts,
		},
		nil,
		nil,
		containerName,
	)
	if err != nil {
		logger.Error("Error creating container: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Start container
	if err := h.client.ContainerStart(req.Request.Context(), containerResp.ID, types.ContainerStartOptions{}); err != nil {
		logger.Error("Error starting container: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Wait for container to be running
	containerInfo, err := h.client.ContainerInspect(req.Request.Context(), containerResp.ID)
	if err != nil {
		logger.Error("Error inspecting container: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	// Convert container to box model
	box := containerToBox(containerInfo)
	resp.WriteHeaderAndEntity(http.StatusCreated, box)
}

// prepareLabels prepares container labels
func prepareLabels(boxID string, boxReq *models.BoxCreateRequest) map[string]string {
	labels := map[string]string{
		GboxLabelCompose:   config.GetGboxLabelCompose(),
		GboxLabelID:        boxID,
		GboxLabelName:      "gbox",
		GboxLabelVersion:   "v1",
		GboxLabelComponent: "sandbox",
		GboxLabelManagedBy: "gru-api-server",
	}

	// Add command configuration to labels if provided
	if boxReq.Cmd != "" {
		labels[GboxLabelPrefix+".cmd"] = boxReq.Cmd
	}
	if len(boxReq.Args) > 0 {
		labels[GboxLabelPrefix+".args"] = common.JoinArgs(boxReq.Args)
	}
	if boxReq.WorkingDir != "" {
		labels[GboxLabelPrefix+".working-dir"] = boxReq.WorkingDir
	}

	// Add custom labels with prefix
	if boxReq.ExtraLabels != nil {
		for k, v := range boxReq.ExtraLabels {
			labels[GboxExtraLabelPrefix+"."+k] = v
		}
	}

	return labels
}

// pullImage pulls the specified image
func pullImage(h *DockerBoxHandler, req *restful.Request, resp *restful.Response, img, imagePullSecret string) error {
	// Prepare pull options
	pullOptions := types.ImagePullOptions{}
	if imagePullSecret != "" {
		pullOptions.RegistryAuth = imagePullSecret
	}

	reader, err := h.client.ImagePull(req.Request.Context(), img, pullOptions)
	if err != nil {
		logger.Error("Error pulling image: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return err
	}
	defer reader.Close()

	// Wait for image pull to complete and decode the JSON stream
	decoder := json.NewDecoder(reader)
	for {
		var event struct {
			Status         string `json:"status"`
			Error          string `json:"error"`
			Progress       string `json:"progress"`
			ProgressDetail struct {
				Current int `json:"current"`
				Total   int `json:"total"`
			} `json:"progressDetail"`
		}

		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			logger.Error("Error decoding pull response: %v", err)
			resp.WriteError(http.StatusInternalServerError, err)
			return err
		}

		if event.Error != "" {
			logger.Error("Error pulling image: %s", event.Error)
			resp.WriteError(http.StatusInternalServerError, fmt.Errorf("error pulling image: %s", event.Error))
			return fmt.Errorf("error pulling image: %s", event.Error)
		}

		logger.Debug("Pull status: %s %s", event.Status, event.Progress)
	}

	logger.Info("Image pulled successfully: %q", img)
	return nil
}
