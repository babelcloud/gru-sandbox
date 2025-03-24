package config

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/format"
	"github.com/babelcloud/gru-sandbox/packages/api-server/internal/log"
)

// K8sConfig holds Kubernetes-specific configuration
type K8sConfig struct {
	KubeConfig string                // Path to kubeconfig file
	Config     *rest.Config          // Kubernetes REST config
	Client     *kubernetes.Clientset // Kubernetes client
}

// NewK8sConfig creates a new K8sConfig
func NewK8sConfig() Config {
	// Try configuration from Viper first
	if kubeconfig := v.GetString("k8s.cfg"); kubeconfig != "" {
		return &K8sConfig{
			KubeConfig: kubeconfig,
		}
	}

	// Try user's home directory
	home := os.Getenv("GBOX_HOME")
	if home == "" {
		if userHome, err := os.UserHomeDir(); err == nil {
			home = filepath.Join(userHome, ".gbox")
		}
	}
	kubeconfig := filepath.Join(home, "kubeconfig")

	// If .gbox/kubeconfig doesn't exist, try default location
	if _, err := os.Stat(kubeconfig); err != nil {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	return &K8sConfig{
		KubeConfig: kubeconfig,
	}
}

// Initialize validates and initializes the configuration
func (c *K8sConfig) Initialize(logger *log.Logger) error {
	logger.Info("%s", format.FormatServerMode("k8s"))

	// Check configuration from Viper
	if kubeConfig := v.GetString("k8s.cfg"); kubeConfig != "" {
		c.KubeConfig = kubeConfig
		logger.Info("Using Kubernetes configuration from: %s", c.KubeConfig)
		return nil
	}

	// Try default kubeconfig paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	// Check ~/.kube/config first
	userConfig := filepath.Join(homeDir, ".kube", "config")
	if _, err := os.Stat(userConfig); err == nil {
		c.KubeConfig = userConfig
		logger.Info("Using Kubernetes configuration from: %s", c.KubeConfig)
		return nil
	}

	// If user config doesn't exist, try /etc/kubernetes/admin.conf
	systemConfig := "/etc/kubernetes/admin.conf"
	if _, err := os.Stat(systemConfig); err == nil {
		c.KubeConfig = systemConfig
		logger.Info("Using Kubernetes configuration from: %s", c.KubeConfig)
		return nil
	}

	return fmt.Errorf("no Kubernetes configuration found in ~/.kube/config or /etc/kubernetes/admin.conf")
}
