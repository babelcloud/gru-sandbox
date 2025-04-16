package api

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"

	"github.com/babelcloud/gbox/packages/api-server/pkg/browser"
)

// RegisterBrowserRoutes adds the browser API routes to the web service.
func RegisterBrowserRoutes(ws *restful.WebService, handler *Handler) {

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

	// --- Action Route ---

	ws.Route(ws.POST("/boxes/{id}/browser-contexts/{context_id}/pages/{page_id}/actions").To(handler.ExecuteAction).
		Doc("Execute an action on a page").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.PathParameter("context_id", "identifier of the context").DataType("string")).
		Param(ws.PathParameter("page_id", "identifier of the page").DataType("string")).
		Reads(model.PageActionParams{}).
		Returns(http.StatusOK, "OK", model.PageActionResult{}).
		Returns(http.StatusBadRequest, "Bad Request", nil).
		Returns(http.StatusNotFound, "Not Found", nil).
		Returns(http.StatusInternalServerError, "Internal Server Error", nil))
}
