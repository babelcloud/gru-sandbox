package types

import (
	"github.com/emicklei/go-restful/v3"
)

// BoxHandler defines the interface for box operations
type BoxHandler interface {
	ListBoxes(req *restful.Request, resp *restful.Response)
	CreateBox(req *restful.Request, resp *restful.Response)
	DeleteBox(req *restful.Request, resp *restful.Response)
	DeleteBoxes(req *restful.Request, resp *restful.Response)
	ExecBox(req *restful.Request, resp *restful.Response)
	RunBox(req *restful.Request, resp *restful.Response)
	StartBox(req *restful.Request, resp *restful.Response)
	StopBox(req *restful.Request, resp *restful.Response)
	GetBox(req *restful.Request, resp *restful.Response)
	GetArchive(req *restful.Request, resp *restful.Response)
	HeadArchive(req *restful.Request, resp *restful.Response)
	ExtractArchive(req *restful.Request, resp *restful.Response)
	ReclaimBoxes(req *restful.Request, resp *restful.Response)
}
