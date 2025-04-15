package api

import (
	"github.com/babelcloud/gbox/packages/api-server/internal/misc/model"
	"github.com/emicklei/go-restful/v3"
)

// RegisterRoutes registers the miscellaneous routes
func RegisterRoutes(ws *restful.WebService, handler *MiscHandler) {
	// Version route
	ws.Route(ws.GET("/version").To(handler.GetVersion).
		Doc("get server version information").
		Returns(200, "OK", model.VersionInfo{}).
		Returns(500, "Internal Server Error", nil))
}
