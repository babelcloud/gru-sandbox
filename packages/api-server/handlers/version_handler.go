package handlers

import (
	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/version"
	"github.com/emicklei/go-restful/v3"
)

// VersionHandler
type VersionHandler struct{}

// NewVersionHandler
func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

// GetVersion
func (h *VersionHandler) GetVersion(req *restful.Request, resp *restful.Response) {
	info := version.ServerInfo()
	resp.WriteEntity(info)
}
