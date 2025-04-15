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
const mockBoxReclaimSuccessResponse = `{"stopped_count":2,"deleted_count":1}`
const mockBoxReclaimEmptyResponse = `{"stopped_count":0,"deleted_count":0}`

// TestBoxReclaimSuccess tests successful reclamation of all box resources
func TestBoxReclaimSuccess(t *testing.T) {
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
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/boxes/reclaim", r.URL.Path)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxReclaimSuccessResponse))
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
	cmd := NewBoxReclaimCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "Box resources successfully reclaimed")
	assert.Contains(t, output, "Stopped 2 boxes")
	assert.Contains(t, output, "Deleted 1 boxes")
}

// TestBoxReclaimWithJsonOutput tests JSON format output for resource reclamation
func TestBoxReclaimWithJsonOutput(t *testing.T) {
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
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/boxes/reclaim", r.URL.Path)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxReclaimSuccessResponse))
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
	cmd := NewBoxReclaimCommand()
	cmd.SetArgs([]string{"--output", "json"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check if output is original JSON
	assert.JSONEq(t, mockBoxReclaimSuccessResponse, strings.TrimSpace(output))
}

// TestBoxReclaimWithForce tests using force parameter to reclaim box resources
func TestBoxReclaimWithForce(t *testing.T) {
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
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/boxes/reclaim", r.URL.Path)

		// Check force parameter
		assert.Equal(t, "force=true", r.URL.RawQuery)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxReclaimSuccessResponse))
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
	cmd := NewBoxReclaimCommand()
	cmd.SetArgs([]string{"--force"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "Box resources successfully reclaimed")
}

// TestBoxReclaimNoResourcesFound tests the case when no resources are found to reclaim
func TestBoxReclaimNoResourcesFound(t *testing.T) {
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
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/boxes/reclaim", r.URL.Path)

		// Return mock response - no resources reclaimed
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxReclaimEmptyResponse))
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
	cmd := NewBoxReclaimCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "Box resources successfully reclaimed")
	assert.NotContains(t, output, "Stopped")
	assert.NotContains(t, output, "Deleted")
}

// TestBoxReclaimHelp tests help information
func TestBoxReclaimHelp(t *testing.T) {
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
	cmd := NewBoxReclaimCommand()
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
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "reclaim [flags]")
	assert.Contains(t, output, "Reclaim resources for all inactive boxes based on configured idle time")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "-f, --force")
	assert.Contains(t, output, "Force resource reclamation, even if box is running")
	assert.Contains(t, output, "gbox box reclaim")
}

// TestBoxReclaimNotFound tests the case when API endpoint is not found
func TestBoxReclaimNotFound(t *testing.T) {
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
		w.Write([]byte(`{"error": "API endpoint not found"}`))
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
	cmd := NewBoxReclaimCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "No inactive boxes found to reclaim or API endpoint not found")
}
