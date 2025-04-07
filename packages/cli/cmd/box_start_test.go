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
const mockBoxStartSuccessResponse = `{"message":"Box started successfully"}`
const mockBoxStartAlreadyRunningResponse = `{"error":"Box is already running"}`
const mockBoxStartInvalidRequestResponse = `{"error":"Invalid request"}`

// TestBoxStartSuccess tests successfully starting a box
func TestBoxStartSuccess(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes/test-box-id/start", r.URL.Path)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxStartSuccessResponse))
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
	cmd := NewBoxStartCommand()
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
	assert.Contains(t, output, "Box started successfully")
}

// TestBoxStartWithJsonOutput tests starting a box with JSON output format
func TestBoxStartWithJsonOutput(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes/test-box-id/start", r.URL.Path)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxStartSuccessResponse))
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
	cmd := NewBoxStartCommand()
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
	assert.JSONEq(t, mockBoxStartSuccessResponse, strings.TrimSpace(output))
}

// TestBoxStartNotFound tests the case when box is not found
func TestBoxStartNotFound(t *testing.T) {
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
	cmd := NewBoxStartCommand()
	cmd.SetArgs([]string{"non-existent-box-id"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "Box not found")
}

// TestBoxStartAlreadyRunning tests the case when box is already running
func TestBoxStartAlreadyRunning(t *testing.T) {
	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return 400 error, box already running
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(mockBoxStartAlreadyRunningResponse))
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
	cmd := NewBoxStartCommand()
	cmd.SetArgs([]string{"already-running-box-id"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "Box is already running")
}

// TestBoxStartInvalidRequest tests the case of an invalid request
func TestBoxStartInvalidRequest(t *testing.T) {
	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return 400 error, but not "box already running"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(mockBoxStartInvalidRequestResponse))
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
	cmd := NewBoxStartCommand()
	cmd.SetArgs([]string{"invalid-request-box-id"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "Error: Invalid request")
}

// TestBoxStartHelp tests help information
func TestBoxStartHelp(t *testing.T) {
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
	cmd := NewBoxStartCommand()
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
	assert.Contains(t, output, "start [box-id]")
	assert.Contains(t, output, "Start a stopped box")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "json or text")
	assert.Contains(t, output, "gbox box start")
}
