package api

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"

	model "github.com/babelcloud/gbox/packages/api-server/pkg/browser"
)

// RegisterBrowserRoutes adds the browser API routes to the web service.
func RegisterBrowserRoutes(ws *restful.WebService, handler *Handler) {

	// Base path for actions
	actionsPath := "/boxes/{id}/browser-contexts/{context_id}/pages/{page_id}/actions"

	// --- Context Routes ---

	ws.Route(ws.POST("/boxes/{id}/browser-contexts").To(handler.CreateContext).
		Doc("Create a new browser context").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Reads(model.CreateContextParams{}).
		Returns(http.StatusOK, "OK", model.CreateContextResult{}).
		Returns(http.StatusBadRequest, "Bad Request", nil).
		Returns(http.StatusInternalServerError, "Internal Server Error", nil))

	ws.Route(ws.DELETE("/boxes/{id}/browser-contexts/{context_id}").To(handler.CloseContext).
		Doc("Close a browser context").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.PathParameter("context_id", "identifier of the context").DataType("string")).
		Returns(http.StatusNoContent, "No Content", nil).
		Returns(http.StatusNotFound, "Not Found", nil).
		Returns(http.StatusInternalServerError, "Internal Server Error", nil))

	// --- Page Routes ---

	ws.Route(ws.POST("/boxes/{id}/browser-contexts/{context_id}/pages").To(handler.CreatePage).
		Doc("Create a new page in a context and navigate to URL").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.PathParameter("context_id", "identifier of the context").DataType("string")).
		Reads(model.CreatePageParams{}).
		Returns(http.StatusOK, "OK", model.CreatePageResult{}).
		Returns(http.StatusBadRequest, "Bad Request", nil).
		Returns(http.StatusNotFound, "Not Found", nil).
		Returns(http.StatusInternalServerError, "Internal Server Error", nil))

	ws.Route(ws.GET("/boxes/{id}/browser-contexts/{context_id}/pages").To(handler.ListPages).
		Doc("List all pages in a context").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.PathParameter("context_id", "identifier of the context").DataType("string")).
		Returns(http.StatusOK, "OK", model.ListPagesResult{}).
		Returns(http.StatusNotFound, "Not Found", nil).
		Returns(http.StatusInternalServerError, "Internal Server Error", nil))

	ws.Route(ws.DELETE("/boxes/{id}/browser-contexts/{context_id}/pages/{page_id}").To(handler.ClosePage).
		Doc("Close a page in a context").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.PathParameter("context_id", "identifier of the context").DataType("string")).
		Param(ws.PathParameter("page_id", "identifier of the page").DataType("string")).
		Returns(http.StatusNoContent, "No Content", nil).
		Returns(http.StatusNotFound, "Not Found", nil).
		Returns(http.StatusInternalServerError, "Internal Server Error", nil))

	ws.Route(ws.GET("/boxes/{id}/browser-contexts/{context_id}/pages/{page_id}").To(handler.GetPage).
		Doc("Get details for a specific page, optionally including content").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.PathParameter("context_id", "identifier of the context").DataType("string")).
		Param(ws.PathParameter("page_id", "identifier of the page").DataType("string")).
		Param(ws.QueryParameter("withContent", "Whether to include page content (HTML or Markdown)").DataType("boolean").Required(false).DefaultValue("false")).
		Param(ws.QueryParameter("contentType", "Format for content if included ('html' or 'markdown')").DataType("string").Required(false).DefaultValue("html")).
		Writes(model.GetPageResult{}).
		Returns(http.StatusOK, "OK", model.GetPageResult{}).
		Returns(http.StatusBadRequest, "Bad Request", model.VisionErrorResult{}).
		Returns(http.StatusNotFound, "Not Found", model.VisionErrorResult{}).
		Returns(http.StatusInternalServerError, "Internal Server Error", model.VisionErrorResult{}))

	// --- Vision Action Routes ---

	// Vision Click
	ws.Route(ws.POST(actionsPath+"/vision-click").To(handler.ExecuteVisionClickAction).
		Doc("Execute vision.click action").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.PathParameter("context_id", "identifier of the context").DataType("string")).
		Param(ws.PathParameter("page_id", "identifier of the page").DataType("string")).
		Reads(model.VisionClickParams{}).
		Returns(http.StatusOK, "OK", model.VisionClickResult{}).
		Returns(http.StatusBadRequest, "Bad Request", model.VisionErrorResult{}).
		Returns(http.StatusNotFound, "Not Found", model.VisionErrorResult{}).
		Returns(http.StatusInternalServerError, "Internal Server Error", model.VisionErrorResult{}))

	// Vision Double Click
	ws.Route(ws.POST(actionsPath+"/vision-doubleClick").To(handler.ExecuteVisionDoubleClickAction).
		Doc("Execute vision.doubleClick action").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.PathParameter("context_id", "identifier of the context").DataType("string")).
		Param(ws.PathParameter("page_id", "identifier of the page").DataType("string")).
		Reads(model.VisionDoubleClickParams{}).
		Returns(http.StatusOK, "OK", model.VisionDoubleClickResult{}).
		Returns(http.StatusBadRequest, "Bad Request", model.VisionErrorResult{}).
		Returns(http.StatusNotFound, "Not Found", model.VisionErrorResult{}).
		Returns(http.StatusInternalServerError, "Internal Server Error", model.VisionErrorResult{}))

	// Vision Type
	ws.Route(ws.POST(actionsPath+"/vision-type").To(handler.ExecuteVisionTypeAction).
		Doc("Execute vision.type action").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.PathParameter("context_id", "identifier of the context").DataType("string")).
		Param(ws.PathParameter("page_id", "identifier of the page").DataType("string")).
		Reads(model.VisionTypeParams{}).
		Returns(http.StatusOK, "OK", model.VisionTypeResult{}).
		Returns(http.StatusBadRequest, "Bad Request", model.VisionErrorResult{}).
		Returns(http.StatusNotFound, "Not Found", model.VisionErrorResult{}).
		Returns(http.StatusInternalServerError, "Internal Server Error", model.VisionErrorResult{}))

	// Vision Drag
	ws.Route(ws.POST(actionsPath+"/vision-drag").To(handler.ExecuteVisionDragAction).
		Doc("Execute vision.drag action").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.PathParameter("context_id", "identifier of the context").DataType("string")).
		Param(ws.PathParameter("page_id", "identifier of the page").DataType("string")).
		Reads(model.VisionDragParams{}).
		Returns(http.StatusOK, "OK", model.VisionDragResult{}).
		Returns(http.StatusBadRequest, "Bad Request", model.VisionErrorResult{}).
		Returns(http.StatusNotFound, "Not Found", model.VisionErrorResult{}).
		Returns(http.StatusInternalServerError, "Internal Server Error", model.VisionErrorResult{}))

	// Vision KeyPress
	ws.Route(ws.POST(actionsPath+"/vision-keyPress").To(handler.ExecuteVisionKeyPressAction).
		Doc("Execute vision.keyPress action").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.PathParameter("context_id", "identifier of the context").DataType("string")).
		Param(ws.PathParameter("page_id", "identifier of the page").DataType("string")).
		Reads(model.VisionKeyPressParams{}).
		Returns(http.StatusOK, "OK", model.VisionKeyPressResult{}).
		Returns(http.StatusBadRequest, "Bad Request", model.VisionErrorResult{}).
		Returns(http.StatusNotFound, "Not Found", model.VisionErrorResult{}).
		Returns(http.StatusInternalServerError, "Internal Server Error", model.VisionErrorResult{}))

	// Vision Move
	ws.Route(ws.POST(actionsPath+"/vision-move").To(handler.ExecuteVisionMoveAction).
		Doc("Execute vision.move action").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.PathParameter("context_id", "identifier of the context").DataType("string")).
		Param(ws.PathParameter("page_id", "identifier of the page").DataType("string")).
		Reads(model.VisionMoveParams{}).
		Returns(http.StatusOK, "OK", model.VisionMoveResult{}).
		Returns(http.StatusBadRequest, "Bad Request", model.VisionErrorResult{}).
		Returns(http.StatusNotFound, "Not Found", model.VisionErrorResult{}).
		Returns(http.StatusInternalServerError, "Internal Server Error", model.VisionErrorResult{}))

	// Vision Screenshot
	ws.Route(ws.POST(actionsPath+"/vision-screenshot").To(handler.ExecuteVisionScreenshotAction).
		Doc("Execute vision.screenshot action").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.PathParameter("context_id", "identifier of the context").DataType("string")).
		Param(ws.PathParameter("page_id", "identifier of the page").DataType("string")).
		Reads(model.VisionScreenshotParams{}).
		Returns(http.StatusOK, "OK", model.VisionScreenshotResult{}).
		Returns(http.StatusBadRequest, "Bad Request", model.VisionErrorResult{}).
		Returns(http.StatusNotFound, "Not Found", model.VisionErrorResult{}).
		Returns(http.StatusInternalServerError, "Internal Server Error", model.VisionErrorResult{}))

	// Vision Scroll
	ws.Route(ws.POST(actionsPath+"/vision-scroll").To(handler.ExecuteVisionScrollAction).
		Doc("Execute vision.scroll action").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.PathParameter("context_id", "identifier of the context").DataType("string")).
		Param(ws.PathParameter("page_id", "identifier of the page").DataType("string")).
		Reads(model.VisionScrollParams{}).
		Returns(http.StatusOK, "OK", model.VisionScrollResult{}).
		Returns(http.StatusBadRequest, "Bad Request", model.VisionErrorResult{}).
		Returns(http.StatusNotFound, "Not Found", model.VisionErrorResult{}).
		Returns(http.StatusInternalServerError, "Internal Server Error", model.VisionErrorResult{}))

	// TODO: Add routes for Snapshot actions
}
