package docker

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
	"github.com/babelcloud/gru-sandbox/packages/api-server/types"
	"github.com/emicklei/go-restful/v3"
)

const (
	// Default reclaim intervals
	defaultStopInterval   = 30 * time.Minute // Stop boxes inactive for 30 minutes
	defaultDeleteInterval = 24 * time.Hour   // Delete boxes inactive for 24 hours
)

// BoxReclaimConfig defines the configuration for box resource reclamation
type BoxReclaimConfig struct {
	StopInterval   time.Duration
	DeleteInterval time.Duration
}

// DefaultBoxReclaimConfig returns the default reclamation configuration
func DefaultBoxReclaimConfig() BoxReclaimConfig {
	return BoxReclaimConfig{
		StopInterval:   defaultStopInterval,
		DeleteInterval: defaultDeleteInterval,
	}
}

// BoxReclaimer manages the reclamation of inactive boxes
type BoxReclaimer struct {
	config     BoxReclaimConfig
	handler    types.BoxHandler
	lastAccess sync.Map // map[string]time.Time
}

// NewBoxReclaimer creates a new box reclaimer
func NewBoxReclaimer(handler types.BoxHandler, config BoxReclaimConfig) *BoxReclaimer {
	return &BoxReclaimer{
		config:  config,
		handler: handler,
	}
}

// handleReclaimBoxes handles the reclamation of inactive boxes
func handleReclaimBoxes(h types.BoxHandler, req *restful.Request, resp *restful.Response) {
	now := time.Now()

	// Get all existing boxes
	var boxes []models.Box
	listReq := restful.NewRequest(&http.Request{
		URL: &url.URL{
			Path: "/api/v1/boxes",
		},
	})
	listResp := restful.NewResponse(&responseWriter{})
	h.ListBoxes(listReq, listResp)

	// Get boxes from response
	var listResponse models.BoxListResponse
	if err := listResp.WriteEntity(&listResponse); err != nil {
		resp.WriteError(http.StatusInternalServerError, fmt.Errorf("failed to get boxes list: %v", err))
		return
	}
	boxes = listResponse.Boxes

	// Create a map of existing box IDs for quick lookup
	existingBoxes := make(map[string]bool)
	for _, box := range boxes {
		existingBoxes[box.ID] = true
	}

	// Get the wrapped handler to access the reclaimer
	wrapped, ok := h.(*wrappedHandler)
	if !ok {
		resp.WriteError(http.StatusInternalServerError, fmt.Errorf("handler is not properly wrapped"))
		return
	}

	// Track stopped and deleted boxes
	stoppedIDs := make([]string, 0)
	deletedIDs := make([]string, 0)

	// First, add current time as access record for boxes without access history
	for _, box := range boxes {
		if _, exists := wrapped.reclaimer.lastAccess.Load(box.ID); !exists {
			wrapped.reclaimer.lastAccess.Store(box.ID, now)
		}
	}

	// Iterate through all boxes and check their last access time
	wrapped.reclaimer.lastAccess.Range(func(key, value interface{}) bool {
		boxID := key.(string)
		accessTime := value.(time.Time)

		// Remove record if box no longer exists
		if !existingBoxes[boxID] {
			wrapped.reclaimer.lastAccess.Delete(boxID)
			return true
		}

		// Calculate inactivity duration
		inactiveDuration := now.Sub(accessTime)

		// Stop inactive boxes to free up resources
		if inactiveDuration >= wrapped.reclaimer.config.StopInterval {
			// Create a new request for stopping the box
			stopReq := restful.NewRequest(req.Request)
			stopReq.PathParameters()["id"] = boxID
			stopResp := restful.NewResponse(resp.ResponseWriter)
			h.StopBox(stopReq, stopResp)
			stoppedIDs = append(stoppedIDs, boxID)
		}

		// Delete very inactive boxes to completely free resources
		if inactiveDuration >= wrapped.reclaimer.config.DeleteInterval {
			// Create a new request for deleting the box
			deleteReq := restful.NewRequest(req.Request)
			deleteReq.PathParameters()["id"] = boxID
			deleteResp := restful.NewResponse(resp.ResponseWriter)
			h.DeleteBox(deleteReq, deleteResp)
			deletedIDs = append(deletedIDs, boxID)
		}

		return true
	})

	// Write response
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	resp.WriteEntity(models.BoxReclaimResponse{
		Message:      "Box reclamation completed successfully",
		StoppedIDs:   stoppedIDs,
		DeletedIDs:   deletedIDs,
		StoppedCount: len(stoppedIDs),
		DeletedCount: len(deletedIDs),
	})
}

// wrappedHandler wraps a BoxHandler to track last access times
type wrappedHandler struct {
	reclaimer *BoxReclaimer
}

// WrapHandler wraps a BoxHandler to track last access times for exec, run, and start operations
func (r *BoxReclaimer) WrapHandler(handler types.BoxHandler) types.BoxHandler {
	return &wrappedHandler{reclaimer: r}
}

// ListBoxes implements BoxHandler interface
func (w *wrappedHandler) ListBoxes(req *restful.Request, resp *restful.Response) {
	w.reclaimer.handler.ListBoxes(req, resp)
}

// CreateBox implements BoxHandler interface
func (w *wrappedHandler) CreateBox(req *restful.Request, resp *restful.Response) {
	w.reclaimer.handler.CreateBox(req, resp)
}

// DeleteBox implements BoxHandler interface
func (w *wrappedHandler) DeleteBox(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	w.reclaimer.handler.DeleteBox(req, resp)
	w.reclaimer.lastAccess.Delete(boxID)
}

// DeleteBoxes implements BoxHandler interface
func (w *wrappedHandler) DeleteBoxes(req *restful.Request, resp *restful.Response) {
	w.reclaimer.handler.DeleteBoxes(req, resp)
	w.reclaimer.lastAccess = sync.Map{}
}

// ExecBox implements BoxHandler interface
func (w *wrappedHandler) ExecBox(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	w.reclaimer.lastAccess.Store(boxID, time.Now())
	w.reclaimer.handler.ExecBox(req, resp)
}

// RunBox implements BoxHandler interface
func (w *wrappedHandler) RunBox(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	w.reclaimer.lastAccess.Store(boxID, time.Now())
	w.reclaimer.handler.RunBox(req, resp)
}

// StartBox implements BoxHandler interface
func (w *wrappedHandler) StartBox(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	w.reclaimer.lastAccess.Store(boxID, time.Now())
	w.reclaimer.handler.StartBox(req, resp)
}

// StopBox implements BoxHandler interface
func (w *wrappedHandler) StopBox(req *restful.Request, resp *restful.Response) {
	w.reclaimer.handler.StopBox(req, resp)
}

// GetBox implements BoxHandler interface
func (w *wrappedHandler) GetBox(req *restful.Request, resp *restful.Response) {
	w.reclaimer.handler.GetBox(req, resp)
}

// ReclaimBoxes implements BoxHandler interface
func (w *wrappedHandler) ReclaimBoxes(req *restful.Request, resp *restful.Response) {
	handleReclaimBoxes(w, req, resp)
}

// responseWriter implements http.ResponseWriter for capturing responses
type responseWriter struct {
	header http.Header
	body   []byte
}

func (w *responseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return len(b), nil
}

func (w *responseWriter) WriteHeader(statusCode int) {
	// We don't need to do anything with the status code
}
