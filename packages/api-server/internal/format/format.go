package format

import (
	"github.com/fatih/color"

	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
)

// APIEndpoint represents an API endpoint
type APIEndpoint struct {
	Method      string
	Path        string
	Description string
}

// FormatHTTPMethod returns a colored and bold HTTP method string
func FormatHTTPMethod(method string) string {
	switch method {
	case "GET":
		return color.New(color.Bold, color.FgGreen).Sprint(method) // Green for safe operations
	case "POST":
		return color.New(color.Bold, color.FgYellow).Sprint(method) // Yellow for creation
	case "PUT":
		return color.New(color.Bold, color.FgBlue).Sprint(method) // Blue for updates
	case "PATCH":
		return color.New(color.Bold, color.FgCyan).Sprint(method) // Cyan for partial updates
	case "DELETE":
		return color.New(color.Bold, color.FgRed).Sprint(method) // Red for deletion
	case "HEAD":
		return color.New(color.Bold, color.FgMagenta).Sprint(method) // Magenta for metadata
	case "OPTIONS":
		return color.New(color.Bold, color.FgWhite).Sprint(method) // White for capabilities
	default:
		return color.New(color.Bold).Sprint(method) // Default bold for unknown methods
	}
}

// FormatServerMode returns a colored and bold server mode string
func FormatServerMode(mode string) string {
	green := color.New(color.FgGreen)
	if mode == "docker" {
		return green.Sprint("Starting server in ") +
			color.New(color.Bold, color.FgCyan).Sprint("docker") +
			green.Sprint(" mode...")
	}
	return green.Sprintf("Starting server in %s mode...", mode)
}

// LogAPIEndpoint logs an API endpoint with consistent formatting
func LogAPIEndpoint(logger *log.Logger, endpoint APIEndpoint) {
	// Using tabs for alignment since ANSI color codes don't affect tab stops
	logger.Info("  %s\t\t%s\t\t%s",
		FormatHTTPMethod(endpoint.Method),
		endpoint.Path,
		endpoint.Description,
	)
}

// LogAPIEndpoints logs a header and a list of API endpoints
func LogAPIEndpoints(logger *log.Logger, endpoints []APIEndpoint) {
	logger.Info("API endpoints:")
	for _, endpoint := range endpoints {
		LogAPIEndpoint(logger, endpoint)
	}
}
