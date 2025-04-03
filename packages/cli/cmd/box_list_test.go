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
const mockBoxListResponse = `{"boxes":[
	{"id":"box-1","image":"ubuntu:latest","status":"running"},
	{"id":"box-2","image":"nginx:1.19","status":"stopped"}
]}`

const mockEmptyBoxListResponse = `{"boxes":[]}`

// TestBoxListAll tests listing all boxes
func TestBoxListAll(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)
		assert.Empty(t, r.URL.RawQuery, "There should be no query parameters")

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxListResponse))
	}))
	defer server.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	defer os.Setenv("API_ENDPOINT", origAPIURL)

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", server.URL)

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewBoxListCommand()
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
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "IMAGE")
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "box-1")
	assert.Contains(t, output, "box-2")
	assert.Contains(t, output, "ubuntu:latest")
	assert.Contains(t, output, "nginx:1.19")
	assert.Contains(t, output, "running")
	assert.Contains(t, output, "stopped")
}

// TestBoxListWithJsonOutput tests listing boxes in JSON format
func TestBoxListWithJsonOutput(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxListResponse))
	}))
	defer server.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	defer os.Setenv("API_ENDPOINT", origAPIURL)

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", server.URL)

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewBoxListCommand()
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
	assert.JSONEq(t, mockBoxListResponse, strings.TrimSpace(output))
}

// TestBoxListWithLabelFilter tests filtering boxes by label
func TestBoxListWithLabelFilter(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)

		// Check query parameters
		query := r.URL.Query()
		filters := query["filter"]
		assert.Len(t, filters, 1)
		assert.Equal(t, "label=project=myapp", filters[0])

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxListResponse))
	}))
	defer server.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	defer os.Setenv("API_ENDPOINT", origAPIURL)

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", server.URL)

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewBoxListCommand()
	cmd.SetArgs([]string{"-f", "label=project=myapp"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "box-1")
	assert.Contains(t, output, "box-2")
}

// TestBoxListWithAncestorFilter tests filtering boxes by image ancestor
func TestBoxListWithAncestorFilter(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)

		// Check query parameters
		query := r.URL.Query()
		filters := query["filter"]
		assert.Len(t, filters, 1)
		assert.Equal(t, "ancestor=ubuntu:latest", filters[0])

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxListResponse))
	}))
	defer server.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	defer os.Setenv("API_ENDPOINT", origAPIURL)

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", server.URL)

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewBoxListCommand()
	cmd.SetArgs([]string{"--filter", "ancestor=ubuntu:latest"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "box-1")
	assert.Contains(t, output, "box-2")
}

// TestBoxListMultipleFilters tests using multiple filters
func TestBoxListMultipleFilters(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)

		// Check query parameters
		query := r.URL.Query()
		filters := query["filter"]
		assert.Len(t, filters, 2)
		assert.Contains(t, filters, "label=project=myapp")
		assert.Contains(t, filters, "id=box-1")

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxListResponse))
	}))
	defer server.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	defer os.Setenv("API_ENDPOINT", origAPIURL)

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", server.URL)

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewBoxListCommand()
	cmd.SetArgs([]string{"-f", "label=project=myapp", "-f", "id=box-1"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "box-1")
	assert.Contains(t, output, "box-2")
}

// TestBoxListEmpty tests the case when no boxes are found
func TestBoxListEmpty(t *testing.T) {
	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty box list
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockEmptyBoxListResponse))
	}))
	defer server.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	defer os.Setenv("API_ENDPOINT", origAPIURL)

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", server.URL)

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewBoxListCommand()
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
	assert.Contains(t, output, "No boxes found")
}

// TestBoxListHelp tests help information
func TestBoxListHelp(t *testing.T) {
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
	cmd := NewBoxListCommand()
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
	assert.Contains(t, output, "Usage: gbox box list [options]")
	assert.Contains(t, output, "List all boxes")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "--filter")
	assert.Contains(t, output, "id=abc123")
	assert.Contains(t, output, "label=project=myapp")
	assert.Contains(t, output, "ancestor=ubuntu:latest")
}
