package docker

import (
	"fmt"

	"github.com/docker/docker/client"
	"github.com/emicklei/go-restful/v3"

	"github.com/babelcloud/gru-sandbox/packages/api-server/config"
	"github.com/babelcloud/gru-sandbox/packages/api-server/types"
)

// DockerBoxHandler handles box-related operations in Docker
type DockerBoxHandler struct {
	client *client.Client
	config *config.DockerConfig
}

// NewDockerBoxHandler creates a new Docker box handler
func NewDockerBoxHandler(cfg *config.DockerConfig) (types.BoxHandler, error) {
	// Initialize Docker client
	client, err := client.NewClientWithOpts(client.WithHost(cfg.Host))
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}

	handler := &DockerBoxHandler{
		client: client,
		config: cfg,
	}

	// Create a new handler instance for the reclaimer to avoid recursion
	reclaimerHandler := &DockerBoxHandler{
		client: client,
		config: cfg,
	}

	// Wrap the handler with reclamation functionality
	reclaimer := NewBoxReclaimer(reclaimerHandler, DefaultBoxReclaimConfig())
	return reclaimer.WrapHandler(handler), nil
}

// ListBoxes returns all boxes
func (h *DockerBoxHandler) ListBoxes(req *restful.Request, resp *restful.Response) {
	handleListBoxes(h, req, resp)
}

// CreateBox creates a new box
func (h *DockerBoxHandler) CreateBox(req *restful.Request, resp *restful.Response) {
	handleCreateBox(h, req, resp)
}

// DeleteBox deletes a box by ID
func (h *DockerBoxHandler) DeleteBox(req *restful.Request, resp *restful.Response) {
	handleDeleteBox(h, req, resp)
}

// DeleteBoxes deletes all boxes
func (h *DockerBoxHandler) DeleteBoxes(req *restful.Request, resp *restful.Response) {
	handleDeleteAllBoxes(h, req, resp)
}

// ExecBox executes a command in a box
func (h *DockerBoxHandler) ExecBox(req *restful.Request, resp *restful.Response) {
	handleExecBox(h, req, resp)
}

// RunBox handles the run box operation
func (h *DockerBoxHandler) RunBox(req *restful.Request, resp *restful.Response) {
	handleRunBox(h, req, resp)
}

// StartBox starts a stopped box
func (h *DockerBoxHandler) StartBox(req *restful.Request, resp *restful.Response) {
	handleStartBox(h, req, resp)
}

// StopBox stops a running box
func (h *DockerBoxHandler) StopBox(req *restful.Request, resp *restful.Response) {
	handleStopBox(h, req, resp)
}

// GetBox gets a box by ID
func (h *DockerBoxHandler) GetBox(req *restful.Request, resp *restful.Response) {
	handleGetBox(h, req, resp)
}

// ReclaimBoxes performs the cleanup of inactive boxes
func (h *DockerBoxHandler) ReclaimBoxes(req *restful.Request, resp *restful.Response) {
	handleReclaimBoxes(h, req, resp)
}
