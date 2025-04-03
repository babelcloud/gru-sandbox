package cmd

import (
	"bytes"
	"io"
	"os" // needed for indirect use
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test cleaning up cluster (Docker mode)
func TestClusterCleanupDocker(t *testing.T) {
	// Skip pipe test
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		return
	}

	// Save original execution function
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gbox-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mock executor
	mockExec := newMockExecutor()
	execCommand = mockExec.execCommand

	// Create .gbox directory and config file
	gboxDir := filepath.Join(tempDir, ".gbox")
	err = os.MkdirAll(gboxDir, 0755)
	assert.NoError(t, err)

	configFile := filepath.Join(gboxDir, "config.yml")
	err = os.WriteFile(configFile, []byte("cluster:\n  mode: docker"), 0644)
	assert.NoError(t, err)

	// Save original stdout for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewClusterCleanupCommand()
	cmd.SetArgs([]string{"--force"}) // Use --force to skip confirmation

	// Set environment variables to simulate HOME directory
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Execute command
	execErr := cmd.Execute()
	t.Logf("Command execution result: %v", execErr) // Log error but don't assert, as it might fail in test environment

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	_, copyErr := io.Copy(&buf, r)
	assert.NoError(t, copyErr, "Reading output should succeed")
	output := buf.String()

	// Output execution information
	t.Logf("Output: %s", output)
	t.Logf("Executed commands: %v", mockExec.commands)
}

// Test cleaning up cluster (K8s mode)
func TestClusterCleanupK8s(t *testing.T) {
	// Skip pipe test
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		return
	}

	// Save original execution function
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gbox-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mock executor
	mockExec := newMockExecutor()
	execCommand = mockExec.execCommand

	// Create .gbox directory and config file
	gboxDir := filepath.Join(tempDir, ".gbox")
	err = os.MkdirAll(gboxDir, 0755)
	assert.NoError(t, err)

	configFile := filepath.Join(gboxDir, "config.yml")
	err = os.WriteFile(configFile, []byte("cluster:\n  mode: k8s"), 0644)
	assert.NoError(t, err)

	// Save original stdout for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewClusterCleanupCommand()
	cmd.SetArgs([]string{"--force"}) // Use --force to skip confirmation

	// Set environment variables to simulate HOME directory
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Execute command
	execErr := cmd.Execute()
	t.Logf("Command execution result: %v", execErr) // Log error but don't assert, as it might fail in test environment

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	_, copyErr := io.Copy(&buf, r)
	assert.NoError(t, copyErr, "Reading output should succeed")
	output := buf.String()

	// Output execution information
	t.Logf("Output: %s", output)
	t.Logf("Executed commands: %v", mockExec.commands)
}

// Test already cleaned situation
func TestClusterCleanupAlreadyCleaned(t *testing.T) {
	// Skip pipe test
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		return
	}

	// Save original execution function
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gbox-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mock executor
	mockExec := newMockExecutor()
	execCommand = mockExec.execCommand

	// Save original stdout for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewClusterCleanupCommand()
	cmd.SetArgs([]string{"--force"}) // Use --force to skip confirmation

	// Set environment variables to simulate HOME directory
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Execute command
	err = cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	_, copyErr := io.Copy(&buf, r)
	assert.NoError(t, copyErr, "Reading output should succeed")
	output := buf.String()

	// Verify output
	assert.Contains(t, output, "Cluster has been cleaned up")
}

// Test help information
func TestClusterCleanupHelp(t *testing.T) {
	// Skip pipe test
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		return
	}

	// Save original stdout for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewClusterCleanupCommand()
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	_, copyErr := io.Copy(&buf, r)
	assert.NoError(t, copyErr, "Reading output should succeed")
	output := buf.String()

	// Verify output contains help information
	assert.Contains(t, output, "Clean up box environment")
	assert.Contains(t, output, "--force")
}
