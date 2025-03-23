package models

// BoxCreateRequest represents a request to create a box
type BoxCreateRequest struct {
	Image           string            `json:"image,omitempty"`
	ImagePullSecret string            `json:"imagePullSecret,omitempty"` // For docker: base64 encoded auth string, for k8s: secret name
	Env             map[string]string `json:"env,omitempty"`
	Cmd             string            `json:"cmd,omitempty"`
	Args            []string          `json:"args,omitempty"`
	WorkingDir      string            `json:"workingDir,omitempty"`
	ExtraLabels     map[string]string `json:"labels,omitempty"`
}

// BoxCreateResponse represents the response from creating a box
type BoxCreateResponse struct {
	Box     Box    `json:"box"`
	Message string `json:"message,omitempty"`
}
