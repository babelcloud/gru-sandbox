package model

// BoxCreateParams represents a request to create a box
type BoxCreateParams struct {
	Image           string            `json:"image,omitempty"`
	ImagePullSecret string            `json:"image_pull_secret,omitempty"` // For docker: base64 encoded auth string, for k8s: secret name
	Env             map[string]string `json:"env,omitempty"`
	Cmd             string            `json:"cmd,omitempty"`
	Args            []string          `json:"args,omitempty"`
	WorkingDir      string            `json:"working_dir,omitempty"`
	ExtraLabels     map[string]string `json:"extra_labels,omitempty"`
	Volumes         []VolumeMount     `json:"volumes,omitempty"` // Volume mounts for the container
}

// VolumeMount represents a volume mount configuration
type VolumeMount struct {
	Source      string `json:"source"`      // Host path
	Target      string `json:"target"`      // Container path
	ReadOnly    bool   `json:"read_only"`   // Whether the mount is read-only
	Propagation string `json:"propagation"` // Mount propagation (private, rprivate, shared, rshared, slave, rslave)
}

// BoxCreateResult represents the response from creating a box
type BoxCreateResult struct {
	Box     Box    `json:"box"`
	Message string `json:"message,omitempty"`
}

// BoxDeleteParams represents a request to delete a box
type BoxDeleteParams struct {
	Force bool `json:"force,omitempty"` // Whether to force delete the box
}

// BoxDeleteResult represents a response from deleting a box
type BoxDeleteResult struct {
	Message string `json:"message"`
}

// BoxesDeleteParams represents a request to delete multiple boxes
type BoxesDeleteParams struct {
	Force bool `json:"force,omitempty"` // Whether to force delete the boxes
}

// BoxesDeleteResult represents a response from deleting multiple boxes
type BoxesDeleteResult struct {
	Count   int      `json:"count"`         // Number of boxes deleted
	Message string   `json:"message"`       // Response message
	IDs     []string `json:"ids,omitempty"` // IDs of deleted boxes
}

// BoxStartResult represents a response from starting a box
type BoxStartResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// BoxStopResult represents a response from stopping a box
type BoxStopResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// BoxReclaimResult represents a response from reclaiming boxes
type BoxReclaimResult struct {
	StoppedCount int      `json:"stopped_count"`         // Number of boxes stopped
	DeletedCount int      `json:"deleted_count"`         // Number of boxes deleted
	StoppedIDs   []string `json:"stopped_ids,omitempty"` // IDs of stopped boxes
	DeletedIDs   []string `json:"deleted_ids,omitempty"` // IDs of deleted boxes
}
