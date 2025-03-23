package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"

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

	// Pull image if not exists
	img := common.GetImage(boxReq.Image)
	logger.Info("Pulling image: %q", img)

	// Prepare pull options
	pullOptions := image.PullOptions{}
	if boxReq.ImagePullSecret != "" {
		pullOptions.RegistryAuth = boxReq.ImagePullSecret
	}

	reader, err := h.client.ImagePull(req.Request.Context(), img, pullOptions)
	if err != nil {
		logger.Error("Error pulling image: %v", err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
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
			return
		}

		if event.Error != "" {
			logger.Error("Error pulling image: %s", event.Error)
			resp.WriteError(http.StatusInternalServerError, fmt.Errorf("error pulling image: %s", event.Error))
			return
		}

		logger.Debug("Pull status: %s %s", event.Status, event.Progress)
	}

	logger.Info("Image pulled successfully: %q", img)

	boxID := uuid.New().String()
	containerName := fmt.Sprintf("gbox-%s", boxID)

	// Prepare labels
	labels := prepareLabels(boxID, &boxReq)

	// Create container
	containerResp, err := h.client.ContainerCreate(
		req.Request.Context(),
		&container.Config{
			Image:      img,
			Cmd:        common.GetCommand(boxReq.Cmd),
			Env:        common.GetEnvVars(boxReq.Env),
			WorkingDir: boxReq.WorkingDir,
			Labels:     labels,
		},
		&container.HostConfig{},
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
	if err := h.client.ContainerStart(req.Request.Context(), containerResp.ID, container.StartOptions{}); err != nil {
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
		GboxLabelCompose:   "gbox-boxes",
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
