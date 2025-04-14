package model

// BoxError represents an error response from the box service
type BoxError struct {
	Code    string `json:"code"`              // Error code
	Message string `json:"message"`           // Error message
	Details string `json:"details,omitempty"` // Additional error details
}
