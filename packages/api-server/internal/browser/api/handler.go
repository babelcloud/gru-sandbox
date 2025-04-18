package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful/v3"

	browserSvc "github.com/babelcloud/gbox/packages/api-server/internal/browser/service"
	model "github.com/babelcloud/gbox/packages/api-server/pkg/browser"
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

// GetPage handles GET /boxes/{id}/browser-contexts/{context_id}/pages/{page_id}
func (h *Handler) GetPage(req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	contextID := req.PathParameter("context_id")
	pageID := req.PathParameter("page_id")
	if boxID == "" || contextID == "" || pageID == "" {
		writeError(resp, http.StatusBadRequest, fmt.Errorf("box ID, context ID and page ID are required"))
		return
	}

	// Query parameters
	withContentStr := req.QueryParameter("withContent")   // Defaults to "false" by route definition
	contentTypeParam := req.QueryParameter("contentType") // Defaults to "html" by route definition

	withContent := false
	if strings.ToLower(withContentStr) == "true" {
		withContent = true
	}

	// Determine requested MIME type
	mimeType := "text/html" // Default
	if withContent {
		switch strings.ToLower(contentTypeParam) {
		case "markdown":
			mimeType = "text/markdown"
		case "html":
			mimeType = "text/html"
		default:
			writeError(resp, http.StatusBadRequest, fmt.Errorf("invalid contentType: '%s', must be 'html' or 'markdown'", contentTypeParam))
			return
		}
	}

	// Call service
	result, err := h.service.GetPage(boxID, contextID, pageID, withContent, mimeType)
	if err != nil {
		if errors.Is(err, browserSvc.ErrContextNotFound) || errors.Is(err, browserSvc.ErrPageNotFound) {
			writeError(resp, http.StatusNotFound, err)
		} else if strings.Contains(err.Error(), "conversion failed") { // Check for markdown conversion error
			writeError(resp, http.StatusInternalServerError, fmt.Errorf("failed to process page content: %w", err))
		} else {
			writeError(resp, http.StatusInternalServerError, fmt.Errorf("failed to get page details: %w", err))
		}
		return
	}

	resp.WriteHeader(http.StatusOK)
	_ = resp.WriteAsJson(result)
}

// --- Vision Action Handlers ---

func (h *Handler) ExecuteVisionClickAction(req *restful.Request, resp *restful.Response) {
	executeSpecificAction[model.VisionClickParams, model.VisionClickResult](h, req, resp, h.service.ExecuteVisionClick)
}

func (h *Handler) ExecuteVisionDoubleClickAction(req *restful.Request, resp *restful.Response) {
	executeSpecificAction[model.VisionDoubleClickParams, model.VisionDoubleClickResult](h, req, resp, h.service.ExecuteVisionDoubleClick)
}

func (h *Handler) ExecuteVisionTypeAction(req *restful.Request, resp *restful.Response) {
	executeSpecificAction[model.VisionTypeParams, model.VisionTypeResult](h, req, resp, h.service.ExecuteVisionType)
}

func (h *Handler) ExecuteVisionDragAction(req *restful.Request, resp *restful.Response) {
	executeSpecificAction[model.VisionDragParams, model.VisionDragResult](h, req, resp, h.service.ExecuteVisionDrag)
}

func (h *Handler) ExecuteVisionKeyPressAction(req *restful.Request, resp *restful.Response) {
	executeSpecificAction[model.VisionKeyPressParams, model.VisionKeyPressResult](h, req, resp, h.service.ExecuteVisionKeyPress)
}

func (h *Handler) ExecuteVisionMoveAction(req *restful.Request, resp *restful.Response) {
	executeSpecificAction[model.VisionMoveParams, model.VisionMoveResult](h, req, resp, h.service.ExecuteVisionMove)
}

func (h *Handler) ExecuteVisionScreenshotAction(req *restful.Request, resp *restful.Response) {
	executeSpecificAction[model.VisionScreenshotParams, model.VisionScreenshotResult](h, req, resp, h.service.ExecuteVisionScreenshot)
}

func (h *Handler) ExecuteVisionScrollAction(req *restful.Request, resp *restful.Response) {
	executeSpecificAction[model.VisionScrollParams, model.VisionScrollResult](h, req, resp, h.service.ExecuteVisionScroll)
}

// --- Specific Action Handlers ---

// executeSpecificAction is a generic helper to reduce boilerplate in action handlers
func executeSpecificAction[P any, R any]( // P: Params type, R: Result type
	h *Handler,
	req *restful.Request,
	resp *restful.Response,
	// Changed signature: Executor now takes IDs + Params
	actionExecutor func(boxID, contextID, pageID string, params P) interface{},
) {
	boxID := req.PathParameter("id")
	contextID := req.PathParameter("context_id")
	pageID := req.PathParameter("page_id")
	if boxID == "" || contextID == "" || pageID == "" {
		writeError(resp, http.StatusBadRequest, fmt.Errorf("box ID, context ID and page ID are required"))
		return
	}

	// 1. Read specific parameters
	var params P
	if err := req.ReadEntity(&params); err != nil {
		writeError(resp, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// 2. Execute the specific action via the provided executor function, passing IDs
	result := actionExecutor(boxID, contextID, pageID, params)

	// 4. Handle the result
	if errResult, ok := result.(model.VisionErrorResult); ok {
		// It's an error result from the service. Check the message content.
		errMsg := errResult.Error // Get the actual error message string
		if strings.Contains(errMsg, browserSvc.ErrBoxNotFound.Error()) ||
			strings.Contains(errMsg, browserSvc.ErrContextNotFound.Error()) ||
			strings.Contains(errMsg, browserSvc.ErrPageNotFound.Error()) {
			// If the message indicates a Not Found error from GetPageInstance
			writeError(resp, http.StatusNotFound, errors.New(errMsg))
		} else {
			// Assume other errors are client-actionable (e.g., bad params within the action) or internal
			// Treat as Bad Request for now.
			writeError(resp, http.StatusBadRequest, errors.New(errMsg))
		}
		return
	}

	// Check if the result type matches the expected successful result type R
	if _, ok := result.(R); !ok {
		writeError(resp, http.StatusInternalServerError, fmt.Errorf("internal error: unexpected result type %T", result))
		return
	}

	// Success
	resp.WriteHeader(http.StatusOK)
	_ = resp.WriteAsJson(result)
}

// --- TODO: Add Snapshot Action Handlers ---
