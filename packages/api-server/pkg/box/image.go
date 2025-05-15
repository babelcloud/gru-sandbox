package model

// ImageUpdateParams represents parameters for updating docker images
type ImageUpdateParams struct {
	ImageReference string `json:"imageReference,omitempty"` // Image reference to update (format: repo/image or repo/image:tag)
	DryRun         bool   `json:"dryRun,omitempty"`         // If true, only report operations without executing them
	Force          bool   `json:"force,omitempty"`          // If true, force remove images
}

// ImageUpdateResponse represents the response for the image update operation.
type ImageUpdateResponse struct {
	Images       []ImageInfo `json:"images"`                 // List of relevant images
	ErrorMessage string      `json:"errorMessage,omitempty"` // General error message for the update process if needed.
}

// ImageStatus represents the status of an image
type ImageStatus string

const (
	// ImageStatusUpToDate means this image is current and exists locally
	ImageStatusUpToDate ImageStatus = "uptodate"
	// ImageStatusOutdated means this image is outdated and exists locally
	ImageStatusOutdated ImageStatus = "outdated"
	// ImageStatusMissing means this image is current but does not exist locally
	ImageStatusMissing ImageStatus = "missing"
)

// ImageInfo represents information about an image in the update context
type ImageInfo struct {
	ImageID    string      `json:"imageId,omitempty"` // Image ID (if exists locally)
	Repository string      `json:"repository"`        // Image repository (e.g., "babelcloud/gbox-playwright")
	Tag        string      `json:"tag"`               // Image tag (e.g., "b281f1a")
	Status     ImageStatus `json:"status"`            // "uptodate", "outdated", or "missing"
	Action     string      `json:"action,omitempty"`  // What will be done: "keep", "delete", "pull"
}
