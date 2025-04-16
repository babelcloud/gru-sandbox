package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful/v3"

	browserSvc "github.com/babelcloud/gbox/packages/api-server/internal/browser/service"
	"github.com/babelcloud/gbox/packages/api-server/pkg/browser"
)

// Handler wraps the browser service to expose it via API endpoints.
type Handler struct {
	service *browserSvc.BrowserService
}

// NewHandler creates a new API handler for the browser service.
func NewHandler(svc *browserSvc.BrowserService) *Handler {
	if svc == nil {
		panic("BrowserService cannot be nil") // Or return an error
	}
	return &Handler{
		service: svc,
	}
}

// writeError is a helper to write standard error responses.
func writeError(resp *restful.Response, statusCode int, err error) {
	resp.WriteHeader(statusCode)
	_ = resp.WriteError(statusCode, err)
	// Log the error server-side as well?
	fmt.Printf("API Error (%d): %v\n", statusCode, err)
}

// --- Context Handlers ---

// CreateContext handles POST /boxes/{id}/browser-contexts
func (h *Handler) CreateContext(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	if boxID == "" {
		writeError(resp, http.StatusBadRequest, fmt.Errorf("box ID is required"))
		return
	}

	var params model.CreateContextParams
	if err := req.ReadEntity(&params); err != nil {
		writeError(resp, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	result, err := h.service.CreateContext(boxID, params)
	if err != nil {
		// TODO: Map service errors (e.g., ErrBoxNotFound) to specific HTTP status codes
		writeError(resp, http.StatusInternalServerError, fmt.Errorf("failed to create context: %w", err))
		return
	}

	resp.WriteHeader(http.StatusOK)
	_ = resp.WriteAsJson(result)
}

// CloseContext handles DELETE /boxes/{id}/browser-contexts/{context_id}
func (h *Handler) CloseContext(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	contextID := req.PathParameter("context_id")
	if boxID == "" || contextID == "" {
		writeError(resp, http.StatusBadRequest, fmt.Errorf("box ID and context ID are required"))
		return
	}

	err := h.service.CloseContext(boxID, contextID)
	if err != nil {
		if errors.Is(err, browserSvc.ErrContextNotFound) {
			writeError(resp, http.StatusNotFound, err)
		} else {
			// Consider other specific errors from the service if added
			writeError(resp, http.StatusInternalServerError, fmt.Errorf("failed to close context: %w", err))
		}
		return
	}

	resp.WriteHeader(http.StatusNoContent)
}

// --- Page Handlers ---

// CreatePage handles POST /boxes/{id}/browser-contexts/{context_id}/pages
func (h *Handler) CreatePage(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	contextID := req.PathParameter("context_id")
	if boxID == "" || contextID == "" {
		writeError(resp, http.StatusBadRequest, fmt.Errorf("box ID and context ID are required"))
		return
	}

	var params model.CreatePageParams
	if err := req.ReadEntity(&params); err != nil {
		writeError(resp, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	result, err := h.service.CreatePage(boxID, contextID, params)
	if err != nil {
		if errors.Is(err, browserSvc.ErrContextNotFound) {
			writeError(resp, http.StatusNotFound, err)
		} else {
			writeError(resp, http.StatusInternalServerError, fmt.Errorf("failed to create page: %w", err))
		}
		return
	}

	resp.WriteHeader(http.StatusOK)
	_ = resp.WriteAsJson(result)
}

// ListPages handles GET /boxes/{id}/browser-contexts/{context_id}/pages
func (h *Handler) ListPages(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	contextID := req.PathParameter("context_id")
	if boxID == "" || contextID == "" {
		writeError(resp, http.StatusBadRequest, fmt.Errorf("box ID and context ID are required"))
		return
	}

	result, err := h.service.ListPages(boxID, contextID)
	if err != nil {
		if errors.Is(err, browserSvc.ErrContextNotFound) {
			writeError(resp, http.StatusNotFound, err)
		} else {
			writeError(resp, http.StatusInternalServerError, fmt.Errorf("failed to list pages: %w", err))
		}
		return
	}

	resp.WriteHeader(http.StatusOK)
	_ = resp.WriteAsJson(result)
}

// ClosePage handles DELETE /boxes/{id}/browser-contexts/{context_id}/pages/{page_id}
func (h *Handler) ClosePage(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	contextID := req.PathParameter("context_id")
	pageID := req.PathParameter("page_id")
	if boxID == "" || contextID == "" || pageID == "" {
		writeError(resp, http.StatusBadRequest, fmt.Errorf("box ID, context ID and page ID are required"))
		return
	}

	err := h.service.ClosePage(boxID, contextID, pageID)
	if err != nil {
		if errors.Is(err, browserSvc.ErrContextNotFound) || errors.Is(err, browserSvc.ErrPageNotFound) {
			writeError(resp, http.StatusNotFound, err)
		} else {
			writeError(resp, http.StatusInternalServerError, fmt.Errorf("failed to close page: %w", err))
		}
		return
	}

	resp.WriteHeader(http.StatusNoContent)
}

// --- Action Handler ---

// ExecuteAction handles POST /boxes/{id}/browser-contexts/{context_id}/pages/{page_id}/actions
func (h *Handler) ExecuteAction(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	contextID := req.PathParameter("context_id")
	pageID := req.PathParameter("page_id")
	if boxID == "" || contextID == "" || pageID == "" {
		writeError(resp, http.StatusBadRequest, fmt.Errorf("box ID, context ID and page ID are required"))
		return
	}

	var params model.PageActionParams
	if err := req.ReadEntity(&params); err != nil {
		writeError(resp, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Basic validation for action type?
	if params.Action == "" {
		writeError(resp, http.StatusBadRequest, fmt.Errorf("action type is required"))
		return
	}

	result, err := h.service.ExecuteAction(boxID, contextID, pageID, params)
	if err != nil {
		if errors.Is(err, browserSvc.ErrContextNotFound) || errors.Is(err, browserSvc.ErrPageNotFound) {
			writeError(resp, http.StatusNotFound, err)
		} else if strings.Contains(err.Error(), "missing parameter") || strings.Contains(err.Error(), "unsupported action") {
			// Catch specific errors from service like missing params or unsupported actions
			writeError(resp, http.StatusBadRequest, err)
		} else {
			// Other errors are likely internal
			writeError(resp, http.StatusInternalServerError, fmt.Errorf("failed to execute action: %w", err))
		}
		return
	}

	resp.WriteHeader(http.StatusOK)
	_ = resp.WriteAsJson(result)
}
