package docker

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/emicklei/go-restful/v3"

	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
)

// readDockerStream reads from a Docker stream and returns stdout and stderr content
func readDockerStream(reader io.Reader) (string, string, error) {
	header := make([]byte, 8)
	var stdout, stderr strings.Builder

	for {
		// Read header
		_, err := io.ReadFull(reader, header)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", "", fmt.Errorf("error reading stream header: %v", err)
		}

		// Parse header
		streamType := header[0]
		// Skip 3 bytes reserved for future use
		size := binary.BigEndian.Uint32(header[4:])

		// Read payload
		payload := make([]byte, size)
		_, err = io.ReadFull(reader, payload)
		if err != nil {
			return "", "", fmt.Errorf("error reading stream payload: %v", err)
		}

		// Write to appropriate output based on stream type
		switch streamType {
		case 1: // stdout
			stdout.Write(payload)
		case 2: // stderr
			stderr.Write(payload)
		}
	}

	return stdout.String(), stderr.String(), nil
}

// collectOutput collects output from a reader with line limit
func collectOutput(reader io.Reader, stdoutLimit, stderrLimit int) (string, string) {
	stdout, stderr, err := readDockerStream(reader)
	if err != nil {
		log.Printf("Error reading Docker stream: %v", err)
		return "", ""
	}

	// Process stdout with line limit
	var stdoutLines []string
	if stdoutLimit >= 0 {
		scanner := bufio.NewScanner(strings.NewReader(stdout))
		for scanner.Scan() && len(stdoutLines) < stdoutLimit {
			stdoutLines = append(stdoutLines, scanner.Text())
		}
		stdout = strings.Join(stdoutLines, "\n")
	}

	// Process stderr with line limit
	var stderrLines []string
	if stderrLimit >= 0 {
		scanner := bufio.NewScanner(strings.NewReader(stderr))
		for scanner.Scan() && len(stderrLines) < stderrLimit {
			stderrLines = append(stderrLines, scanner.Text())
		}
		stderr = strings.Join(stderrLines, "\n")
	}

	return stdout, stderr
}

// handleRunBox handles the run box operation
func handleRunBox(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	log.Printf("Received run request for box: %s", boxID)

	box, err := h.getContainerByID(req.Request.Context(), boxID)
	if err != nil {
		if err.Error() == "box not found" {
			log.Printf("Box not found: %s", boxID)
			writeError(resp, http.StatusNotFound, "BOX_NOT_FOUND", fmt.Sprintf("Box not found: %s", boxID))
		} else if err.Error() == "box ID is required" {
			log.Printf("Invalid request: box ID is required")
			writeError(resp, http.StatusBadRequest, "INVALID_REQUEST", "Box ID is required")
		} else {
			log.Printf("Error getting container: %v", err)
			writeError(resp, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error getting container: %v", err))
		}
		return
	}

	// Check container status
	if box.State != "running" {
		log.Printf("Box %s is not running (current state: stopped)", boxID)
		writeError(resp, http.StatusConflict, "BOX_NOT_RUNNING",
			fmt.Sprintf("Box %s is not running (current state: stopped)", boxID))
		return
	}

	// Parse request body
	var runReq models.BoxRunRequest
	if err := req.ReadEntity(&runReq); err != nil {
		log.Printf("Error reading request body: %v", err)
		writeError(resp, http.StatusBadRequest, "INVALID_REQUEST", fmt.Sprintf("Error reading request body: %v", err))
		return
	}

	// Set default line limits if not specified
	if runReq.StdoutLineLimit == 0 {
		runReq.StdoutLineLimit = 100
	}
	if runReq.StderrLineLimit == 0 {
		runReq.StderrLineLimit = 100
	}

	// Create exec configuration
	execConfig := types.ExecConfig{
		User:         "", // Use default user
		Privileged:   false,
		Tty:          false,
		AttachStdin:  runReq.Stdin != "", // Only attach stdin if stdin string is provided
		AttachStderr: true,
		AttachStdout: true,
		Detach:       false,
		DetachKeys:   "",  // Use default detach keys
		Env:          nil, // No additional environment variables
		WorkingDir:   "",  // Use container's working directory
		Cmd:          append(runReq.Cmd, runReq.Args...),
	}

	log.Printf("Creating exec with config: %+v", execConfig)

	// Create exec instance
	execCreate, err := h.client.ContainerExecCreate(req.Request.Context(), box.ID, execConfig)
	if err != nil {
		log.Printf("Error creating exec: %v", err)
		writeError(resp, http.StatusInternalServerError, "EXEC_FAILED", fmt.Sprintf("Error creating exec: %v", err))
		return
	}

	// Attach to exec instance
	execAttach, err := h.client.ContainerExecAttach(req.Request.Context(), execCreate.ID, types.ExecStartCheck{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		log.Printf("Error attaching to exec: %v", err)
		writeError(resp, http.StatusInternalServerError, "EXEC_FAILED", fmt.Sprintf("Error attaching to exec: %v", err))
		return
	}
	defer execAttach.Close()

	// Create channels for collecting output
	outputChan := make(chan struct {
		stdout string
		stderr string
	})
	exitCodeChan := make(chan int)

	// Start goroutine to collect output
	go func() {
		stdout, stderr := collectOutput(execAttach.Reader, runReq.StdoutLineLimit, runReq.StderrLineLimit)
		outputChan <- struct {
			stdout string
			stderr string
		}{stdout, stderr}
	}()

	// Write stdin if provided
	if runReq.Stdin != "" {
		go func() {
			_, err := io.WriteString(execAttach.Conn, runReq.Stdin)
			if err != nil {
				log.Printf("Error writing stdin: %v", err)
			}
			// Close write end of the connection to signal EOF
			if closer, ok := execAttach.Conn.(interface{ CloseWrite() error }); ok {
				closer.CloseWrite()
			}
		}()
	}

	// Wait for exec to complete and get exit code
	go func() {
		execInspect, err := h.client.ContainerExecInspect(req.Request.Context(), execCreate.ID)
		if err != nil {
			log.Printf("Error inspecting exec: %v", err)
			exitCodeChan <- -1
			return
		}
		exitCodeChan <- execInspect.ExitCode
	}()

	// Collect results
	output := <-outputChan
	exitCode := <-exitCodeChan

	// Prepare response
	result := models.BoxRunResponse{
		BoxID:    boxID,
		ExitCode: exitCode,
		Stdout:   output.stdout,
		Stderr:   output.stderr,
	}

	resp.WriteAsJson(result)
}
