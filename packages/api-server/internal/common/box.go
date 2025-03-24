package common

import "encoding/json"

const (
	DefaultImage = "ubuntu:latest"
)

// GetImage returns the image to use, falling back to default if none specified
func GetImage(image string) string {
	if image == "" {
		return DefaultImage
	}
	return image
}

// GetCommand returns the command to run, falling back to default if none specified
func GetCommand(cmd string) []string {
	if cmd == "" {
		return []string{"sleep", "infinity"}
	}
	return []string{"/bin/sh", "-c", cmd}
}

// JoinArgs converts a string array to a JSON string
func JoinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	// Convert args array to JSON string to preserve spaces and special characters
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return ""
	}
	return string(argsJSON)
}
