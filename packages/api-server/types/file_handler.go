package types

import "github.com/emicklei/go-restful/v3"

// FileHandler defines the interface for file operations
type FileHandler interface {
	// HeadFile handles HEAD requests to get file metadata
	HeadFile(req *restful.Request, resp *restful.Response)

	// GetFile handles GET requests to retrieve file content
	GetFile(req *restful.Request, resp *restful.Response)

	// ReclaimFiles removes files that haven't been accessed for more than 14 days
	ReclaimFiles(req *restful.Request, resp *restful.Response)
}
