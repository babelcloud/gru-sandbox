package api

import (
	"net/http"

	"github.com/babelcloud/gbox/packages/api-server/internal/misc/service"
	"github.com/emicklei/go-restful/v3"
)

// MiscHandler handles miscellaneous operations
type MiscHandler struct {
	service *service.MiscService
}

// NewMiscHandler creates a new MiscHandler
func NewMiscHandler(service *service.MiscService) *MiscHandler {
	return &MiscHandler{
		service: service,
	}
}

// GetVersion handles GET /version request
func (h *MiscHandler) GetVersion(req *restful.Request, resp *restful.Response) {
	version := h.service.GetVersion()
	resp.WriteHeaderAndJson(http.StatusOK, version, restful.MIME_JSON)
}
