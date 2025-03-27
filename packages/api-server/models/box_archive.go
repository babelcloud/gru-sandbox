package models

// BoxGetArchiveRequest represents the request for getting an archive from a container
type BoxGetArchiveRequest struct {
	Path string `json:"path" description:"resource in the container's filesystem to archive"`
}

// BoxHeadArchiveRequest represents the request for getting metadata about a resource in the container's filesystem
type BoxHeadArchiveRequest struct {
	Path string `json:"path" description:"resource in the container's filesystem to get metadata for"`
}

// BoxExtractArchiveRequest represents the request for extracting an archive to a container
type BoxExtractArchiveRequest struct {
	Path                 string `json:"path" description:"path to a directory in the container to extract the archive's contents into"`
	NoOverwriteDirNonDir bool   `json:"noOverwriteDirNonDir,omitempty" description:"if true, it will be an error if unpacking would cause an existing directory to be replaced with a non-directory and vice versa"`
	CopyUIDGID           bool   `json:"copyUIDGID,omitempty" description:"if true, it will copy UID/GID maps to the dest file or dir"`
}

// BoxExtractArchiveResponse represents the response for extract archive operation
type BoxExtractArchiveResponse struct {
	Success bool   `json:"success" description:"whether the operation was successful"`
	Error   string `json:"error,omitempty" description:"error message if operation failed"`
}
