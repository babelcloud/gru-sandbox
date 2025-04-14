package api

import (
	"github.com/babelcloud/gbox/packages/api-server/pkg/file"
	"github.com/emicklei/go-restful/v3"
)

// RegisterRoutes registers the file-related routes
func RegisterRoutes(ws *restful.WebService, handler *FileHandler) {
	// File routes
	ws.Route(ws.HEAD("/files/{path:*}").To(handler.HeadFile).
		Doc("get file metadata").
		Param(ws.PathParameter("path", "path to the file").DataType("string")).
		Returns(200, "OK", model.FileStat{}).
		Returns(400, "Bad Request", model.FileError{}).
		Returns(404, "Not Found", model.FileError{}).
		Returns(500, "Internal Server Error", model.FileError{}))

	ws.Route(ws.GET("/files/{path:*}").To(handler.GetFile).
		Doc("get file content").
		Param(ws.PathParameter("path", "path to the file").DataType("string")).
		Returns(200, "OK", nil).
		Notes("The response Content-Type will be set according to the file's MIME type. "+
			"For example: text/plain for text files, image/jpeg for JPEG images, etc. "+
			"Directories will return application/x-directory.").
		Returns(400, "Bad Request", model.FileError{}).
		Returns(404, "Not Found", model.FileError{}).
		Returns(500, "Internal Server Error", model.FileError{}))

	ws.Route(ws.POST("/files").To(handler.HandleFileOperation).
		Doc("handle file operations like reclaim and share").
		Param(ws.QueryParameter("operation", "operation to perform (reclaim or share)").DataType("string").Required(true)).
		Reads(model.FileShareParams{}).
		Returns(200, "OK", model.FileShareResult{}).
		Returns(400, "Bad Request", model.FileError{}).
		Returns(404, "Not Found", model.FileError{}).
		Returns(500, "Internal Server Error", model.FileError{}))
}
