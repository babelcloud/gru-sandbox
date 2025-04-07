package models

// BoxStartResponse represents the response for starting a box
type BoxStartResponse struct {
	Success bool   `json:"success"` // Whether the operation was successful
	Message string `json:"message"` // Human-readable message
}
