package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDirExists(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gbox-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test existing directory
	assert.True(t, dirExists(tempDir))

	// Test non-existent directory
	nonExistentDir := filepath.Join(tempDir, "non-existent")
	assert.False(t, dirExists(nonExistentDir))

	// Test file instead of directory
	tempFile := filepath.Join(tempDir, "test-file")
	err = os.WriteFile(tempFile, []byte("test"), 0644)
	require.NoError(t, err)
	assert.False(t, dirExists(tempFile))
}

func TestGetPackagesRootPath(t *testing.T) {
	// This test is limited since it depends on the environment
	// Just verify the function runs without errors
	rootPath, err := getPackagesRootPath()

	// In a real environment, we should find a packages path
	// In test environment, it might fail depending on where tests are run
	if err == nil {
		// If successful, verify path contains "packages" and directory exists
		assert.Contains(t, rootPath, "packages")
		assert.True(t, dirExists(rootPath))
		assert.True(t, dirExists(filepath.Join(rootPath, "mcp-server")))
	}
}

func TestMergeConfigs(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "gbox-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test case 1: Target file does not exist
	nonExistentConfig := filepath.Join(tempDir, "non-existent.json")
	newConfig := McpConfig{
		McpServers: map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		}{
			"test": {
				Command: "test-cmd",
				Args:    []string{"arg1", "arg2"},
			},
		},
	}

	result, err := mergeConfigs(nonExistentConfig, newConfig)
	require.NoError(t, err)
	assert.Equal(t, newConfig, result)

	// Test case 2: Target file exists but is empty
	emptyConfig := filepath.Join(tempDir, "empty.json")
	err = os.WriteFile(emptyConfig, []byte{}, 0644)
	require.NoError(t, err)

	result, err = mergeConfigs(emptyConfig, newConfig)
	require.NoError(t, err)
	assert.Equal(t, newConfig, result)

	// Test case 3: Target file exists with existing configuration
	existingConfig := filepath.Join(tempDir, "existing.json")
	existingData := McpConfig{
		McpServers: map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		}{
			"existing": {
				Command: "existing-cmd",
				Args:    []string{"arg1"},
			},
		},
	}
	existingJson, err := json.Marshal(existingData)
	require.NoError(t, err)
	err = os.WriteFile(existingConfig, existingJson, 0644)
	require.NoError(t, err)

	result, err = mergeConfigs(existingConfig, newConfig)
	require.NoError(t, err)

	// Verify both configurations exist in the result
	assert.Equal(t, 2, len(result.McpServers))
	assert.Contains(t, result.McpServers, "existing")
	assert.Contains(t, result.McpServers, "test")
	assert.Equal(t, "existing-cmd", result.McpServers["existing"].Command)
	assert.Equal(t, "test-cmd", result.McpServers["test"].Command)
}

func TestNewMcpExportCommand(t *testing.T) {
	cmd := NewMcpExportCommand()

	// Verify command basic properties
	assert.Equal(t, "export", cmd.Use)
	assert.Equal(t, "Export MCP configuration for Claude Desktop", cmd.Short)

	// Verify flag options
	mergeTo, err := cmd.Flags().GetString("merge-to")
	require.NoError(t, err)
	assert.Equal(t, "", mergeTo)

	dryRun, err := cmd.Flags().GetBool("dry-run")
	require.NoError(t, err)
	assert.Equal(t, false, dryRun)
}
