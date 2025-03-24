package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/format"
	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
)

// DockerConfig holds Docker-specific configuration
type DockerConfig struct {
	Host string // Docker daemon socket/host
}

// NewDockerConfig creates a new DockerConfig
func NewDockerConfig() Config {
	return &DockerConfig{}
}

// Initialize validates and initializes the configuration
func (c *DockerConfig) Initialize(logger *log.Logger) error {
	logger.Info("%s", format.FormatServerMode("docker"))

	// Check configuration from Viper
	if host := v.GetString("docker.host"); host != "" {
		c.Host = host
		logger.Info("Using Docker configuration with host: \"%s\"", c.Host)
		return nil
	}

	// Try user's home directory socket first
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	// Check ~/.docker/run/docker.sock first
	userSocket := filepath.Join(homeDir, ".docker", "run", "docker.sock")
	if _, err := os.Stat(userSocket); err == nil {
		c.Host = fmt.Sprintf("unix://%s", userSocket)
		logger.Info("Using Docker configuration with host: \"%s\"", c.Host)
		return nil
	}

	// If user socket doesn't exist, try /var/run/docker.sock
	systemSocket := "/var/run/docker.sock"
	if _, err := os.Stat(systemSocket); err == nil {
		c.Host = fmt.Sprintf("unix://%s", systemSocket)
		logger.Info("Using Docker configuration with host: \"%s\"", c.Host)
		return nil
	}

	return fmt.Errorf("no Docker socket found in ~/.docker/run/docker.sock or /var/run/docker.sock")
}
