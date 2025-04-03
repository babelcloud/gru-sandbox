package models

// BoxStartRequest represents a request to start a box
type BoxStartRequest struct {
	Mounts []Mount `json:"mounts,omitempty"` // Additional mount configurations
}

// BoxStartResponse represents the response for starting a box
type BoxStartResponse struct {
	Success bool   `json:"success"` // Whether the operation was successful
	Message string `json:"message,omitempty"` // Human-readable message
}
