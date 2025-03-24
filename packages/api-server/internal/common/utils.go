package common

import (
	"fmt"
)

// GetEnvVars converts environment variables map to string slice
func GetEnvVars(env map[string]string) []string {
	if env == nil {
		return nil
	}

	vars := make([]string, 0, len(env))
	for k, v := range env {
		vars = append(vars, fmt.Sprintf("%s=%s", k, v))
	}
	return vars
}
