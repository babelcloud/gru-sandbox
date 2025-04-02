package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock responses
const mockDeleteResponse = `{"status":"success"}`
const mockListResponse = `{"boxes":[{"id":"box-1"},{"id":"box-2"}]}`
const mockEmptyListResponse = `{"boxes":[]}`

// TestDeleteSingleBox tests deleting a single box
func TestDeleteSingleBox(t *testing.T) {
	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check path and method
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/v1/boxes/test-box-id", r.URL.Path)

		// Check request content
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		var req map[string]interface{}
		err = json.Unmarshal(body, &req)
		assert.NoError(t, err)
		assert.Equal(t, true, req["force"])

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockDeleteResponse))
	}))
	defer mockServer.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	defer os.Setenv("API_ENDPOINT", origAPIURL)

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", mockServer.URL)

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewBoxDeleteCommand()
	cmd.SetArgs([]string{
		"test-box-id",
		"--force", // Force deletion to avoid confirmation prompt
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "Box deleted successfully")
}

// TestDeleteAllBoxes tests deleting all boxes
func TestDeleteAllBoxes(t *testing.T) {
	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create mock server
	var requestCount int
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// First request should be to list all boxes
		if requestCount == 1 {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/v1/boxes", r.URL.Path)

			// Return box list
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockListResponse))
			return
		}

		// Subsequent requests should be to delete individual boxes
		assert.Equal(t, "DELETE", r.Method)
		assert.True(t, strings.HasPrefix(r.URL.Path, "/api/v1/boxes/box-"))

		// Check request content
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		var req map[string]interface{}
		err = json.Unmarshal(body, &req)
		assert.NoError(t, err)
		assert.Equal(t, true, req["force"])

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockDeleteResponse))
	}))
	defer mockServer.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	defer os.Setenv("API_ENDPOINT", origAPIURL)

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", mockServer.URL)

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewBoxDeleteCommand()
	cmd.SetArgs([]string{
		"--all",
		"--force", // Force deletion to avoid confirmation prompt
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "The following boxes will be deleted")
	assert.Contains(t, output, "All boxes deleted successfully")

	// Verify correct number of requests were sent
	// 1 GET request to list boxes + 2 DELETE requests to delete boxes = 3 requests
	assert.Equal(t, 3, requestCount)
}

// TestDeleteAllBoxesEmpty tests deleting all boxes when there are no boxes
func TestDeleteAllBoxesEmpty(t *testing.T) {
	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)

		// Return empty box list
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockEmptyListResponse))
	}))
	defer mockServer.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	defer os.Setenv("API_ENDPOINT", origAPIURL)

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", mockServer.URL)

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewBoxDeleteCommand()
	cmd.SetArgs([]string{
		"--all",
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "No boxes to delete")
}

// TestDeleteBoxWithJSONOutput tests JSON output format
func TestDeleteBoxWithJSONOutput(t *testing.T) {
	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/v1/boxes/test-box-id", r.URL.Path)

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockDeleteResponse))
	}))
	defer mockServer.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	defer os.Setenv("API_ENDPOINT", origAPIURL)

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", mockServer.URL)

	// Create pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute command
	cmd := NewBoxDeleteCommand()
	cmd.SetArgs([]string{
		"test-box-id",
		"--force",
		"--output", "json",
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, `{"status":"success"`)
	assert.Contains(t, output, `"message":"Box deleted successfully"`)
}

// TestDeleteInvalidInput tests invalid input
func TestDeleteInvalidInput(t *testing.T) {
	// Skip this test as it requires testing os.Exit behavior
	// In Go test environment, os.Exit will directly end the test process
	t.Skip("Need to test os.Exit behavior in a subprocess")

	// Note: To test error cases, you can follow the method described in box_create_test.go
	// to create a subprocess to test os.Exit behavior
}

// TestDeleteAllBoxesWithConfirmation tests deleting all boxes with confirmation
func TestDeleteAllBoxesWithConfirmation(t *testing.T) {
	// Save original stdout, stderr, and stdin for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldStdin := os.Stdin
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		os.Stdin = oldStdin
	}()

	// Create mock server
	var requestCount int
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// First request should be to list all boxes
		if requestCount == 1 {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/v1/boxes", r.URL.Path)

			// Return box list
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockListResponse))
			return
		}

		// Subsequent requests should be to delete individual boxes
		assert.Equal(t, "DELETE", r.Method)
		assert.True(t, strings.HasPrefix(r.URL.Path, "/api/v1/boxes/box-"))

		// Check request content
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		var req map[string]interface{}
		err = json.Unmarshal(body, &req)
		assert.NoError(t, err)
		assert.Equal(t, true, req["force"])

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockDeleteResponse))
	}))
	defer mockServer.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	defer os.Setenv("API_ENDPOINT", origAPIURL)

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", mockServer.URL)

	// Simulate user typing "y" for confirmation
	r, w, _ := os.Pipe()
	os.Stdin = r
	// Write "y" to stdin
	go func() {
		w.Write([]byte("y\n"))
		w.Close()
	}()

	// Capture stdout
	outR, outW, _ := os.Pipe()
	os.Stdout = outW
	os.Stderr = outW

	// Execute command
	cmd := NewBoxDeleteCommand()
	cmd.SetArgs([]string{
		"--all", // Delete all boxes, will prompt for confirmation
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	outW.Close()
	var buf bytes.Buffer
	io.Copy(&buf, outR)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check output
	assert.Contains(t, output, "The following boxes will be deleted")
	assert.Contains(t, output, "Are you sure you want to delete all boxes?")
	assert.Contains(t, output, "All boxes deleted successfully")

	// Verify correct number of requests were sent
	// 1 GET request to list boxes + 2 DELETE requests to delete boxes = 3 requests
	assert.Equal(t, 3, requestCount)
}

// TestDeleteHelp tests help information
func TestDeleteHelp(t *testing.T) {
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
	cmd := NewBoxDeleteCommand()
	cmd.SetArgs([]string{
		"--help",
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Read captured output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "Captured output: %s\n", output)

	// Check if help message contains key options and usage
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "--all")
	assert.Contains(t, output, "--force")
}
