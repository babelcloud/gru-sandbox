package model

// BoxArchiveGetParams represents the request for getting an archive from a container
type BoxArchiveGetParams struct {
	Path string `json:"path" description:"resource in the container's filesystem to archive"`
}

// BoxArchiveHeadParams represents the request for getting metadata about a resource in the container's filesystem
type BoxArchiveHeadParams struct {
	Path string `json:"path" description:"resource in the container's filesystem to get metadata for"`
}

// BoxArchiveExtractParams represents the request for extracting an archive to a container
type BoxArchiveExtractParams struct {
	Path                 string `json:"path" description:"path to a directory in the container to extract the archive's contents into"`
	NoOverwriteDirNonDir bool   `json:"noOverwriteDirNonDir,omitempty" description:"if true, it will be an error if unpacking would cause an existing directory to be replaced with a non-directory and vice versa"`
	CopyUIDGID           bool   `json:"copyUIDGID,omitempty" description:"if true, it will copy UID/GID maps to the dest file or dir"`
	Content              []byte `json:"-" description:"the content of the archive to extract"`
}

// BoxArchiveResult represents the response for getting an archive
type BoxArchiveResult struct {
	Name  string `json:"name" description:"name of the file or directory"`
	Size  int64  `json:"size" description:"size of the file or directory"`
	Mode  uint32 `json:"mode" description:"file mode bits"`
	Mtime string `json:"mtime" description:"modification time in RFC3339 format"`
}

// BoxArchiveHeadResult represents the response for getting metadata about a resource
type BoxArchiveHeadResult struct {
	Name  string `json:"name" description:"name of the file or directory"`
	Size  int64  `json:"size" description:"size of the file or directory"`
	Mode  uint32 `json:"mode" description:"file mode bits"`
	Mtime string `json:"mtime" description:"modification time in RFC3339 format"`
}

// BoxArchiveExtractResult represents the response for extract archive operation
type BoxArchiveExtractResult struct {
	Success bool   `json:"success" description:"whether the operation was successful"`
	Error   string `json:"error,omitempty" description:"error message if operation failed"`
}
