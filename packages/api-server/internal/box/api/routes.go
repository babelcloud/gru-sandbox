package api

import (
	"github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/emicklei/go-restful/v3"
)

// RegisterRoutes registers all box-related routes to the WebService
func RegisterRoutes(ws *restful.WebService, boxHandler *BoxHandler) {
	// Box Lifecycle Operations
	ws.Route(ws.GET("/boxes").To(boxHandler.ListBoxes).
		Doc("list all boxes").
		Returns(200, "OK", []model.Box{}).
		Returns(500, "Internal Server Error", model.BoxError{}))

	ws.Route(ws.GET("/boxes/{id}").To(boxHandler.GetBox).
		Doc("get a box by ID").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Returns(200, "OK", model.Box{}).
		Returns(404, "Not Found", model.BoxError{}).
		Returns(500, "Internal Server Error", model.BoxError{}))

	ws.Route(ws.POST("/boxes").To(boxHandler.CreateBox).
		Doc("create a box").
		Reads(model.BoxCreateParams{}).
		Returns(201, "Created", model.Box{}).
		Returns(400, "Bad Request", model.BoxError{}).
		Returns(500, "Internal Server Error", model.BoxError{}))

	ws.Route(ws.DELETE("/boxes/{id}").To(boxHandler.DeleteBox).
		Doc("delete a box").
		Reads(model.BoxDeleteParams{}).
		Returns(200, "OK", model.BoxDeleteResult{}).
		Returns(404, "Not Found", model.BoxError{}).
		Returns(500, "Internal Server Error", model.BoxError{}))

	ws.Route(ws.DELETE("/boxes").To(boxHandler.DeleteBoxes).
		Doc("delete all boxes").
		Reads(model.BoxesDeleteParams{}).
		Returns(200, "OK", model.BoxesDeleteResult{}).
		Returns(500, "Internal Server Error", model.BoxError{}))

	ws.Route(ws.POST("/boxes/reclaim").To(boxHandler.ReclaimBoxes).
		Doc("reclaim inactive boxes").
		Returns(200, "OK", model.BoxReclaimResult{}).
		Returns(500, "Internal Server Error", model.BoxError{}))

	// Box Runtime Operations
	ws.Route(ws.POST("/boxes/{id}/exec").To(boxHandler.ExecBox).
		Doc("execute a command in a box").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Reads(model.BoxExecParams{}).
		Consumes(restful.MIME_JSON).
		Produces(model.MediaTypeMultiplexedStream, model.MediaTypeRawStream).
		Returns(200, "OK", model.BoxExecResult{}).
		Returns(400, "Bad Request", model.BoxError{}).
		Returns(404, "Not Found", model.BoxError{}).
		Returns(500, "Internal Server Error", model.BoxError{}))

	ws.Route(ws.POST("/boxes/{id}/run").To(boxHandler.RunBox).
		Doc("run a command in a box").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Reads(model.BoxRunParams{}).
		Returns(200, "OK", model.BoxRunResult{}).
		Returns(400, "Bad Request", model.BoxError{}).
		Returns(404, "Not Found", model.BoxError{}).
		Returns(500, "Internal Server Error", model.BoxError{}))

	ws.Route(ws.POST("/boxes/{id}/start").To(boxHandler.StartBox).
		Doc("start a stopped box").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Returns(200, "OK", model.BoxStartResult{}).
		Returns(400, "Bad Request", model.BoxError{}).
		Returns(404, "Not Found", model.BoxError{}).
		Returns(500, "Internal Server Error", model.BoxError{}))

	ws.Route(ws.POST("/boxes/{id}/stop").To(boxHandler.StopBox).
		Doc("stop a running box").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Returns(200, "OK", model.BoxStopResult{}).
		Returns(400, "Bad Request", model.BoxError{}).
		Returns(404, "Not Found", model.BoxError{}).
		Returns(500, "Internal Server Error", model.BoxError{}))

	// Box Archive Operations
	ws.Route(ws.HEAD("/boxes/{id}/archive").To(boxHandler.HeadArchive).
		Doc("get metadata about files in box").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.QueryParameter("path", "path to get metadata from").DataType("string").Required(true)).
		Returns(200, "OK", nil).
		Returns(400, "Bad Request", model.BoxError{}).
		Returns(404, "Not Found", model.BoxError{}).
		Returns(500, "Internal Server Error", model.BoxError{}))

	ws.Route(ws.GET("/boxes/{id}/archive").To(boxHandler.GetArchive).
		Doc("get files from box as tar archive").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.QueryParameter("path", "path to get files from").DataType("string").Required(true)).
		Produces("application/x-tar").
		Returns(200, "OK", nil).
		Returns(400, "Bad Request", model.BoxError{}).
		Returns(404, "Not Found", model.BoxError{}).
		Returns(500, "Internal Server Error", model.BoxError{}))

	ws.Route(ws.PUT("/boxes/{id}/archive").To(boxHandler.ExtractArchive).
		Doc("extract tar archive to box").
		Param(ws.PathParameter("id", "identifier of the box").DataType("string")).
		Param(ws.QueryParameter("path", "path to extract files to").DataType("string").Required(true)).
		Consumes("application/x-tar").
		Returns(200, "OK", nil).
		Returns(400, "Bad Request", model.BoxError{}).
		Returns(404, "Not Found", model.BoxError{}).
		Returns(500, "Internal Server Error", model.BoxError{}))
}
