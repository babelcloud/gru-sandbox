package models

// BoxStopResponse represents the response for stopping a box
type BoxStopResponse struct {
	Success bool   `json:"success"` // Whether the operation was successful
	Message string `json:"message"` // Human-readable message
}
