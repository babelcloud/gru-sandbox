package common

import "encoding/json"

const (
	DefaultImage = "ubuntu:latest"
	// DefaultWorkDirPath is the base path for all box-related directories
	DefaultWorkDirPath = "/var/gbox"
	// DefaultShareDirPath is the path for the shared directory within a box
	DefaultShareDirPath = DefaultWorkDirPath + "/share"
)

// GetImage returns the image to use, falling back to default if none specified
func GetImage(image string) string {
	if image == "" {
		return DefaultImage
	}
	return image
}

// GetCommand returns the command to run, falling back to default if none specified
func GetCommand(cmd string, args []string) []string {
	if cmd == "" {
		return []string{"sleep", "infinity"}
	}
	if len(args) == 0 {
		// If no args provided, use shell to parse the command string
		return []string{"/bin/sh", "-c", cmd}
	}
	// If args are provided, use direct command array
	return append([]string{cmd}, args...)
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
