package service

import "errors"

var (
	// ErrUnknownImplementation is returned when trying to create a service with an unknown implementation
	ErrUnknownImplementation = errors.New("unknown box service implementation")

	// ErrInvalidConfig is returned when the provided configuration is invalid
	ErrInvalidConfig = errors.New("invalid box service configuration")

	// ErrBoxNotFound is returned when a box with the specified ID does not exist
	ErrBoxNotFound = errors.New("box not found")

	// ErrBoxNotRunning is returned when trying to execute a command in a box that is not running
	ErrBoxNotRunning = errors.New("box is not running")
)
