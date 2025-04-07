package models

// VolumeMount represents a volume mount configuration
type VolumeMount struct {
	Source      string `json:"source"`      // Host path
	Target      string `json:"target"`      // Container path
	ReadOnly    bool   `json:"readOnly"`    // Whether the mount is read-only
	Propagation string `json:"propagation"` // Mount propagation (private, rprivate, shared, rshared, slave, rslave)
}

// BoxCreateRequest represents a request to create a box
type BoxCreateRequest struct {
	Image           string            `json:"image,omitempty"`
	ImagePullSecret string            `json:"imagePullSecret,omitempty"` // For docker: base64 encoded auth string, for k8s: secret name
	Env             map[string]string `json:"env,omitempty"`
	Cmd             string            `json:"cmd,omitempty"`
	Args            []string          `json:"args,omitempty"`
	WorkingDir      string            `json:"workingDir,omitempty"`
	ExtraLabels     map[string]string `json:"labels,omitempty"`
	Volumes         []VolumeMount     `json:"volumes,omitempty"` // Volume mounts for the container
}

// BoxCreateResponse represents the response from creating a box
type BoxCreateResponse struct {
	Box     Box    `json:"box"`
	Message string `json:"message,omitempty"`
}
