package handlers

import (
	"fmt"

	"github.com/babelcloud/gru-sandbox/packages/api-server/config"
	"github.com/babelcloud/gru-sandbox/packages/api-server/handlers/docker"
	"github.com/babelcloud/gru-sandbox/packages/api-server/types"
)

// InitBoxHandler initializes the appropriate box handler based on configuration
func InitBoxHandler(cfg config.Config) (types.BoxHandler, error) {
	switch c := cfg.(type) {
	case *config.K8sConfig:
		return NewK8sBoxHandler(c)
	case *config.DockerConfig:
		return docker.NewDockerBoxHandler(c)
	default:
		return nil, fmt.Errorf("unsupported config type: %T", cfg)
	}
}
