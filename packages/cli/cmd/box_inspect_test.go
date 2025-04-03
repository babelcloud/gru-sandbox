package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test data
const mockBoxInspectResponse = `{
	"id": "test-box-id",
	"image": "ubuntu:latest",
	"status": "running",
	"created": "2023-05-01T12:00:00Z",
	"ports": [{"host": 8080, "container": 80}],
	"env": {"DEBUG": "true", "ENV": "test"}
}`

// TestBoxInspect tests getting box details
func TestBoxInspect(t *testing.T) {
	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/boxes/test-box-id", r.URL.Path)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxInspectResponse))
	}))
	defer server.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	origTESTING := os.Getenv("TESTING")
	defer func() {
		os.Setenv("API_ENDPOINT", origAPIURL)
		os.Setenv("TESTING", origTESTING)
	}()

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", server.URL)
	os.Setenv("TESTING", "true")

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewBoxInspectCommand()
	cmd.SetArgs([]string{"test-box-id"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "Box details:")
	assert.Contains(t, output, "id")
	assert.Contains(t, output, "test-box-id")
	assert.Contains(t, output, "image")
	assert.Contains(t, output, "ubuntu:latest")
	assert.Contains(t, output, "status")
	assert.Contains(t, output, "running")
}

// TestBoxInspectWithJsonOutput tests getting box details in JSON format
func TestBoxInspectWithJsonOutput(t *testing.T) {
	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/boxes/test-box-id", r.URL.Path)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxInspectResponse))
	}))
	defer server.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	origTESTING := os.Getenv("TESTING")
	defer func() {
		os.Setenv("API_ENDPOINT", origAPIURL)
		os.Setenv("TESTING", origTESTING)
	}()

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", server.URL)
	os.Setenv("TESTING", "true")

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewBoxInspectCommand()
	cmd.SetArgs([]string{"test-box-id", "--output", "json"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check if output is original JSON
	assert.JSONEq(t, mockBoxInspectResponse, strings.TrimSpace(output))
}

// TestBoxInspectNotFound tests the case when box does not exist
func TestBoxInspectNotFound(t *testing.T) {
	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return 404 error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Box not found"}`))
	}))
	defer server.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	origTESTING := os.Getenv("TESTING")
	defer func() {
		os.Setenv("API_ENDPOINT", origAPIURL)
		os.Setenv("TESTING", origTESTING)
	}()

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", server.URL)
	os.Setenv("TESTING", "true")

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewBoxInspectCommand()
	cmd.SetArgs([]string{"non-existent-box-id"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check output
	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)
	assert.Contains(t, output, "Box not found")
}

// TestBoxInspectHelp tests help information
func TestBoxInspectHelp(t *testing.T) {
	// Save original stdout and stderr for later restoration
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
	cmd := NewBoxInspectCommand()
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check if help message contains key sections
	assert.Contains(t, output, "Usage: gbox box inspect <id> [options]")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "json or text")
	assert.Contains(t, output, "Get box details")
	assert.Contains(t, output, "Get box details in JSON format")
}
