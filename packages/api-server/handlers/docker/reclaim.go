package docker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
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
	logger := log.New()
	now := time.Now()
	logger.Info("Starting box reclamation at %v", now)

	// Get all existing boxes
	var boxes []models.Box
	listReq := restful.NewRequest(&http.Request{
		URL: &url.URL{
			Path: "/api/v1/boxes",
		},
	})
	writer := &responseWriter{}
	listResp := restful.NewResponse(writer)
	h.ListBoxes(listReq, listResp)

	// Get boxes from response
	var listResponse models.BoxListResponse
	if err := json.Unmarshal(writer.body, &listResponse); err != nil {
		logger.Error("Failed to get boxes list: %v, body: %s", err, string(writer.body))
		resp.WriteError(http.StatusInternalServerError, fmt.Errorf("failed to get boxes list: %v", err))
		return
	}
	boxes = listResponse.Boxes
	logger.Info("Found %d existing boxes", len(boxes))

	// Get the wrapped handler to access the reclaimer
	wrapped, ok := h.(*wrappedHandler)
	if !ok {
		logger.Error("Handler is not properly wrapped")
		resp.WriteError(http.StatusInternalServerError, fmt.Errorf("handler is not properly wrapped"))
		return
	}

	// Create a map of existing box IDs and their status for quick lookup
	existingBoxes := make(map[string]bool)
	stoppedBoxes := make(map[string]bool)
	for _, box := range boxes {
		existingBoxes[box.ID] = true
		if box.Status == "stopped" {
			stoppedBoxes[box.ID] = true
		}
	}

	// Track stopped and deleted boxes
	stoppedIDs := make([]string, 0)
	deletedIDs := make([]string, 0)

	// Iterate through all boxes and check their last access time
	wrapped.reclaimer.lastAccess.Range(func(key, value interface{}) bool {
		boxID := key.(string)
		accessTime := value.(time.Time)

		// Remove record if box no longer exists
		if !existingBoxes[boxID] {
			logger.Debug("Removing access record for non-existent box %s", boxID)
			wrapped.reclaimer.lastAccess.Delete(boxID)
			return true
		}

		// Calculate inactivity duration
		inactiveDuration := now.Sub(accessTime)
		logger.Debug("Box %s: last access at %v, inactive for %v", boxID, accessTime, inactiveDuration)

		// Stop inactive boxes to free up resources
		if inactiveDuration >= wrapped.reclaimer.config.StopInterval {
			logger.Info("Stopping box %s: inactive for %v", boxID, inactiveDuration)
			// Create a new request for stopping the box
			stopReq := restful.NewRequest(req.Request)
			stopReq.PathParameters()["id"] = boxID
			stopWriter := &responseWriter{}
			stopResp := restful.NewResponse(stopWriter)
			wrapped.reclaimer.handler.StopBox(stopReq, stopResp)
			if stopResp.StatusCode() == http.StatusOK {
				stoppedIDs = append(stoppedIDs, boxID)
				logger.Info("Successfully stopped box %s", boxID)
				// Update access time after stopping to prevent repeated attempts
				wrapped.reclaimer.lastAccess.Store(boxID, now)
			} else if stopResp.StatusCode() == http.StatusBadRequest && string(stopWriter.body) == "box is already stopped" {
				// If box is already stopped, just update its access time
				logger.Debug("Box %s is already stopped, updating access time", boxID)
				wrapped.reclaimer.lastAccess.Store(boxID, now)
			} else {
				logger.Error("Failed to stop box %s: status code %d, body: %s", boxID, stopResp.StatusCode(), string(stopWriter.body))
			}
		}

		// Delete very inactive boxes to completely free resources
		if inactiveDuration >= wrapped.reclaimer.config.DeleteInterval {
			logger.Info("Deleting box %s: inactive for %v", boxID, inactiveDuration)
			// Create a new request for deleting the box
			deleteReq := restful.NewRequest(req.Request)
			deleteReq.PathParameters()["id"] = boxID
			deleteWriter := &responseWriter{}
			deleteResp := restful.NewResponse(deleteWriter)
			wrapped.reclaimer.handler.DeleteBox(deleteReq, deleteResp)
			if deleteResp.StatusCode() == http.StatusOK {
				deletedIDs = append(deletedIDs, boxID)
				logger.Info("Successfully deleted box %s", boxID)
				// Remove access record after successful deletion
				wrapped.reclaimer.lastAccess.Delete(boxID)
			} else {
				logger.Error("Failed to delete box %s: status code %d, body: %s", boxID, deleteResp.StatusCode(), string(deleteWriter.body))
			}
		}

		return true
	})

	// Initialize access records for new boxes
	for _, box := range boxes {
		if _, exists := wrapped.reclaimer.lastAccess.Load(box.ID); !exists {
			logger.Debug("Initializing access time for new box %s", box.ID)
			wrapped.reclaimer.lastAccess.Store(box.ID, now)
		}
	}

	logger.Info("Reclamation completed: stopped %d boxes, deleted %d boxes",
		len(stoppedIDs), len(deletedIDs))

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
	logger := log.New()
	boxID := req.PathParameter("id")
	logger.Debug("Updating access time for box %s (ExecBox)", boxID)
	w.reclaimer.lastAccess.Store(boxID, time.Now())
	w.reclaimer.handler.ExecBox(req, resp)
}

// RunBox implements BoxHandler interface
func (w *wrappedHandler) RunBox(req *restful.Request, resp *restful.Response) {
	logger := log.New()
	boxID := req.PathParameter("id")
	logger.Debug("Updating access time for box %s (RunBox)", boxID)
	w.reclaimer.lastAccess.Store(boxID, time.Now())
	w.reclaimer.handler.RunBox(req, resp)
}

// StartBox implements BoxHandler interface
func (w *wrappedHandler) StartBox(req *restful.Request, resp *restful.Response) {
	logger := log.New()
	boxID := req.PathParameter("id")
	logger.Debug("Updating access time for box %s (StartBox)", boxID)
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
