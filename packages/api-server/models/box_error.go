package models

// BoxError represents a standard error response for box operations
type BoxError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
