package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetCurrentModeEmpty tests getting current mode when config file doesn't exist
func TestGetCurrentModeEmpty(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gbox-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set non-existent config file path
	configFile := filepath.Join(tempDir, "config.yml")

	// Get current mode
	mode, err := getCurrentMode(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "", mode, "For non-existent config file, should return empty mode")
}

// TestGetCurrentModeDocker tests reading docker mode from config file
func TestGetCurrentModeDocker(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gbox-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test config file
	configFile := filepath.Join(tempDir, "config.yml")
	err = os.WriteFile(configFile, []byte("cluster:\n  mode: docker"), 0644)
	assert.NoError(t, err)

	// Get current mode
	mode, err := getCurrentMode(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "docker", mode, "Should correctly read docker mode")
}

// TestGetCurrentModeK8s tests reading k8s mode from config file
func TestGetCurrentModeK8s(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gbox-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test config file
	configFile := filepath.Join(tempDir, "config.yml")
	err = os.WriteFile(configFile, []byte("cluster:\n  mode: k8s"), 0644)
	assert.NoError(t, err)

	// Get current mode
	mode, err := getCurrentMode(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "k8s", mode, "Should correctly read k8s mode")
}

// TestSaveMode tests saving mode to config file
func TestSaveMode(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gbox-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set config file path
	configFile := filepath.Join(tempDir, "config.yml")

	// Save docker mode
	err = saveMode(configFile, "docker")
	assert.NoError(t, err)

	// Verify saved result
	mode, err := getCurrentMode(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "docker", mode, "Should correctly save and read docker mode")

	// Change to k8s mode
	err = saveMode(configFile, "k8s")
	assert.NoError(t, err)

	// Verify change result
	mode, err = getCurrentMode(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "k8s", mode, "Should correctly update and read k8s mode")
}

// TestGetScriptDir tests getting script directory
func TestGetScriptDir(t *testing.T) {
	dir, err := getScriptDir()
	assert.NoError(t, err)
	assert.NotEmpty(t, dir, "Script directory should not be empty")
}
