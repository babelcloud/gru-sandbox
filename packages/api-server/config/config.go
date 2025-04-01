package config

import (
	"fmt"
	"os"

	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
	"github.com/spf13/viper"
)

// Config is the interface that wraps the basic cluster configuration methods
type Config interface {
	// Initialize initializes the configuration
	Initialize(logger *log.Logger) error
}

var v *viper.Viper

func init() {
	v = viper.New()

	// Set default values
	v.SetDefault("server.port", 28080)
	v.SetDefault("cluster.mode", "docker")
	v.SetDefault("gbox.label_compose", "gbox-boxes")

	// Environment variables
	v.AutomaticEnv()
	v.BindEnv("cluster.mode", "CLUSTER_MODE")
	v.BindEnv("server.port", "PORT")
	v.BindEnv("cluster.docker.host", "DOCKER_HOST")
	v.BindEnv("cluster.k8s.cfg", "KUBECONFIG")
	v.BindEnv("gbox.home", "GBOX_HOME")
	v.BindEnv("gbox.share", "GBOX_SHARE")
	v.BindEnv("gbox.host_share", "GBOX_HOST_SHARE")
	v.BindEnv("gbox.label_compose", "GBOX_LABEL_COMPOSE")

	// Config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Look for config in the following paths
	configPaths := []string{
		".",
		"$HOME/.gbox",
		"/etc/gbox",
	}

	for _, path := range configPaths {
		expandedPath := os.ExpandEnv(path)
		v.AddConfigPath(expandedPath)
	}

	// Read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			panic(fmt.Sprintf("Fatal error reading config file: %s", err))
		}
		// Config file not found; ignore error and use defaults
	}
}

// GetConfig returns the appropriate config based on cluster mode
func GetConfig() (Config, error) {
	mode := v.GetString("cluster.mode")

	var cfg Config
	switch mode {
	case "docker":
		cfg = NewDockerConfig()
	case "k8s":
		cfg = NewK8sConfig()
	default:
		return nil, fmt.Errorf("unsupported cluster mode: %s", mode)
	}

	return cfg, nil
}

// GetServerPort returns the configured server port
func GetServerPort() int {
	return v.GetInt("server.port")
}

// GetGboxLabelCompose returns the configured gbox label compose value
func GetGboxLabelCompose() string {
	return v.GetString("gbox.label_compose")
}
