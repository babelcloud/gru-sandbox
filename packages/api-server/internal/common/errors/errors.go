package errors

import (
	"fmt"
	"net/http"
)

// Error represents a common error type
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.Message
}

// New creates a new error with the given code and message
func New(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Newf creates a new error with the given code and formatted message
func Newf(code int, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Common error codes
const (
	CodeInternalError = http.StatusInternalServerError
	CodeBadRequest    = http.StatusBadRequest
	CodeNotFound      = http.StatusNotFound
	CodeConflict      = http.StatusConflict
)

// Common error messages
var (
	ErrInternalError = New(CodeInternalError, "Internal server error")
	ErrBadRequest    = New(CodeBadRequest, "Bad request")
	ErrNotFound      = New(CodeNotFound, "Resource not found")
	ErrConflict      = New(CodeConflict, "Resource conflict")
)
