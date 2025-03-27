package handlers

import (
	"net/http"

	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
	"github.com/emicklei/go-restful/v3"
)

// HandleFileOperation handles file operations like reclaim and share
func (h *FileHandler) HandleFileOperation(req *restful.Request, resp *restful.Response) {
	operation := req.QueryParameter("operation")
	logger := log.New()

	switch operation {
	case "reclaim":
		h.ReclaimFiles(req, resp)
	case "share":
		h.ShareFile(req, resp)
	default:
		logger.Error("Invalid operation: %s", operation)
		resp.WriteErrorString(http.StatusBadRequest, "invalid operation")
	}
}
