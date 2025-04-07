package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test parsing box path functionality
func TestParseBoxPath(t *testing.T) {
	// Test valid path
	validPath := "box-id:/some/path"
	result, err := parseBoxPath(validPath)
	assert.NoError(t, err)
	assert.Equal(t, "box-id", result.BoxID)
	assert.Equal(t, "/some/path", result.Path)

	// Test invalid path
	invalidPath := "invalid-path-without-colon"
	_, err = parseBoxPath(invalidPath)
	assert.Error(t, err)
}

// Test box path validation
func TestIsBoxPath(t *testing.T) {
	assert.True(t, isBoxPath("box-id:/path"))
	assert.False(t, isBoxPath("/local/path"))
}

// Test copying from box to local
func TestCopyFromBoxToLocal(t *testing.T) {
	// Skip this test as it involves tar command and filesystem operations
	t.Skip("skip test that requires filesystem operations and tar commands")

	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create temporary directory as destination path
	tempDir, err := os.MkdirTemp("", "box-cp-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Define destination file path
	destFile := filepath.Join(tempDir, "test-file")

	// Create mock HTTP server
	mockContent := []byte("mock file content")
	mockArchive := createMockTarArchive(t, "test-file", mockContent)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/boxes/box-id/archive", r.URL.Path)
		assert.Equal(t, "path=/test/file", r.URL.RawQuery)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/x-tar")
		w.WriteHeader(http.StatusOK)
		w.Write(mockArchive)
	}))
	defer server.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	defer os.Setenv("API_ENDPOINT", origAPIURL)

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", server.URL)

	// Capture stdout and stderr
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	// Execute command
	cmd := NewBoxCpCommand()
	cmd.SetArgs([]string{
		"box-id:/test/file",
		destFile,
	})

	// Start a goroutine to execute command to avoid os.Exit interruption
	done := make(chan bool)
	go func() {
		defer close(done)
		err = cmd.Execute()
		stdoutW.Close()
		stderrW.Close()
	}()

	// Wait for command completion
	<-done

	// Read stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, stdoutR)
	io.Copy(&stderrBuf, stderrR)

	// Verify destination file exists
	_, err = os.Stat(destFile)
	assert.NoError(t, err, "Destination file should exist")

	// Verify file content
	content, err := os.ReadFile(destFile)
	assert.NoError(t, err)
	assert.Equal(t, mockContent, content, "File content should be correct")

	// Verify stderr output
	assert.Contains(t, stderrBuf.String(), "Copied from box")
}

// Test copying from local to box
func TestCopyFromLocalToBox(t *testing.T) {
	// Skip this test as it involves tar command and filesystem operations
	t.Skip("skip test that requires filesystem operations and tar commands")

	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create temporary source file
	tempFile, err := os.CreateTemp("", "box-cp-source")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write test content to source file
	testContent := []byte("test content for upload")
	_, err = tempFile.Write(testContent)
	assert.NoError(t, err)
	tempFile.Close()

	// Verify content uploaded to server
	var uploadedContent []byte

	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/boxes/box-id/archive", r.URL.Path)
		assert.Equal(t, "path=/dest/path", r.URL.RawQuery)
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "application/x-tar", r.Header.Get("Content-Type"))

		// Read uploaded content
		uploadedContent, err = io.ReadAll(r.Body)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	defer os.Setenv("API_ENDPOINT", origAPIURL)

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", server.URL)

	// Capture stdout and stderr
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	// Execute command
	cmd := NewBoxCpCommand()
	cmd.SetArgs([]string{
		tempFile.Name(),
		"box-id:/dest/path",
	})

	// Start a goroutine to execute command to avoid os.Exit interruption
	done := make(chan bool)
	go func() {
		defer close(done)
		err = cmd.Execute()
		stdoutW.Close()
		stderrW.Close()
	}()

	// Wait for command completion
	<-done

	// Read stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, stdoutR)
	io.Copy(&stderrBuf, stderrR)

	// Verify uploaded content is valid tar file
	assert.True(t, len(uploadedContent) > 0, "Should upload non-empty content")

	// Verify stderr output
	assert.Contains(t, stderrBuf.String(), "Copied from")
	assert.Contains(t, stderrBuf.String(), "to box")
}

// Test copying from box to stdout
func TestCopyFromBoxToStdout(t *testing.T) {
	// Skip this test as it involves os.Exit call
	t.Skip("skip test that calls os.Exit")

	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create mock HTTP server
	mockContent := []byte("mock file content for stdout")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/boxes/box-id/archive", r.URL.Path)
		assert.Equal(t, "path=/test/file-stdout", r.URL.RawQuery)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/x-tar")
		w.WriteHeader(http.StatusOK)
		w.Write(mockContent)
	}))
	defer server.Close()

	// Save original environment variables
	origAPIURL := os.Getenv("API_ENDPOINT")
	defer os.Setenv("API_ENDPOINT", origAPIURL)

	// Set API URL to mock server
	os.Setenv("API_ENDPOINT", server.URL)

	// Capture stdout and stderr
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	// Execute command
	cmd := NewBoxCpCommand()
	cmd.SetArgs([]string{
		"box-id:/test/file-stdout",
		"-",
	})

	// Start a goroutine to execute command to avoid os.Exit interruption
	done := make(chan bool)
	go func() {
		defer close(done)
		err := cmd.Execute()
		assert.NoError(t, err)
		stdoutW.Close()
		stderrW.Close()
	}()

	// Wait for command completion
	<-done

	// Read stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, stdoutR)
	io.Copy(&stderrBuf, stderrR)

	// Verify stdout
	assert.Equal(t, mockContent, stdoutBuf.Bytes(), "Should write content to stdout")
}

// Test copying from stdin to box
func TestCopyFromStdinToBox(t *testing.T) {
	// Skip: This test requires stdin simulation, which is complex
	t.Skip("stdin test temporarily skipped, requires complex stdin simulation")
}

// Test help message
func TestBoxCpHelp(t *testing.T) {
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
	cmd := NewBoxCpCommand()
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

	// Check if help message contains key parts
	assert.Contains(t, output, "usage: gbox-box-cp")
	assert.Contains(t, output, "positional arguments:")
	assert.Contains(t, output, "src")
	assert.Contains(t, output, "dst")
	assert.Contains(t, output, "options:")
	assert.Contains(t, output, "--help")
	assert.Contains(t, output, "Copy local file to box")
	assert.Contains(t, output, "Copy from box to local")
	assert.Contains(t, output, "Copy tar stream from stdin")
	assert.Contains(t, output, "Copy from box to stdout")
}

// Create mock tar archive for testing
func createMockTarArchive(t *testing.T, filename string, content []byte) []byte {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "tar-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test file
	testFilePath := filepath.Join(tempDir, filename)
	err = os.WriteFile(testFilePath, content, 0644)
	assert.NoError(t, err)

	// Create temporary tar file
	tarFile, err := os.CreateTemp("", "test-*.tar")
	assert.NoError(t, err)
	defer os.Remove(tarFile.Name())
	tarFile.Close()

	// Create tar archive
	cmd := exec.Command("tar", "-cf", tarFile.Name(), "-C", tempDir, filename)
	err = cmd.Run()
	assert.NoError(t, err)

	// Read tar content
	tarContent, err := os.ReadFile(tarFile.Name())
	assert.NoError(t, err)

	return tarContent
}

// Test invalid arguments
func TestBoxCpInvalidArgs(t *testing.T) {
	// Skip this test as os.Exit will abort the test process
	t.Skip("skip test that calls os.Exit")

	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Capture stdout and stderr
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	// Execute command with insufficient arguments
	cmd := NewBoxCpCommand()
	cmd.SetArgs([]string{
		"only-one-arg",
	})

	// Since command calls os.Exit, we can't execute it directly
	// Here we only verify it outputs help message
	_ = cmd.Execute()

	// Verify results
	stdoutW.Close()
	stderrW.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, stdoutR)
	io.Copy(&stderrBuf, stderrR)

	output := stdoutBuf.String()
	assert.Contains(t, output, "Usage: gbox box cp <src> <dst>", "Should display help message")
}

// Test invalid path combinations
func TestBoxCpInvalidPathCombination(t *testing.T) {
	// Skip this test as os.Exit will abort the test process
	t.Skip("skip test that calls os.Exit")

	// Save original stdout and stderr for later restoration
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Capture stdout and stderr
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	// Execute command with two box paths
	cmd := NewBoxCpCommand()
	cmd.SetArgs([]string{
		"box1:/path1",
		"box2:/path2",
	})

	// Start a goroutine to execute command to avoid os.Exit interruption
	done := make(chan bool)
	go func() {
		defer close(done)
		_ = cmd.Execute()
		stdoutW.Close()
		stderrW.Close()
	}()

	// Wait for command completion
	<-done

	// Read stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, stdoutR)
	io.Copy(&stderrBuf, stderrR)

	combined := stdoutBuf.String() + stderrBuf.String()
	assert.True(t, strings.Contains(combined, "Error") || strings.Contains(combined, "Invalid path format"),
		"Should display error message")
}
