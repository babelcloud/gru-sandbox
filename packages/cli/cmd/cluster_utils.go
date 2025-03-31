package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ClusterConfig cluster configuration file structure
type ClusterConfig struct {
	Cluster struct {
		Mode string `yaml:"mode"`
	} `yaml:"cluster"`
}

// getCurrentMode gets current mode from config file
func getCurrentMode(configFile string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return "", nil
	}

	// Read config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return "", err
	}

	// Parse YAML
	var config ClusterConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", err
	}

	return config.Cluster.Mode, nil
}

// saveMode saves mode to config file
func saveMode(configFile string, mode string) error {
	// Ensure directory exists
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Prepare config data
	config := ClusterConfig{}
	config.Cluster.Mode = mode

	// If file already exists, read existing content first
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return err
		}

		if err := yaml.Unmarshal(data, &config); err != nil {
			return err
		}

		// Update mode
		config.Cluster.Mode = mode
	}

	// Save to file
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

// getScriptDir gets script directory
func getScriptDir() (string, error) {
	// Get binary file path
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("unable to get executable path: %v", err)
	}

	// Return directory containing the binary file
	return filepath.Dir(exePath), nil
}
