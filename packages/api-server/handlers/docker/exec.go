package docker

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/emicklei/go-restful/v3"

	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// writeError writes a standard error response
func writeError(resp *restful.Response, statusCode int, code, message string) {
	resp.WriteHeader(statusCode)
	resp.WriteAsJson(ErrorResponse{
		Message: message,
		Code:    code,
	})
}

// handleExecBox handles the exec box operation
func handleExecBox(h *DockerBoxHandler, req *restful.Request, resp *restful.Response) {
	boxID := req.PathParameter("id")
	log.Printf("Received exec request for box: %s", boxID)

	container, err := h.getContainerByID(req.Request.Context(), boxID)
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
	if container.State != "running" {
		log.Printf("Box %s is not running (current state: stopped)", boxID)
		writeError(resp, http.StatusConflict, "BOX_NOT_RUNNING",
			fmt.Sprintf("Box %s is not running (current state: stopped)", boxID))
		return
	}

	// Parse request body
	var execReq models.BoxExecRequest
	if err := req.ReadEntity(&execReq); err != nil {
		log.Printf("Error reading request body: %v", err)
		writeError(resp, http.StatusBadRequest, "INVALID_REQUEST", fmt.Sprintf("Error reading request body: %v", err))
		return
	}
	log.Printf("Exec request: cmd=%v, args=%v, tty=%v, stdin=%v, stdout=%v, stderr=%v",
		execReq.Cmd, execReq.Args, execReq.TTY, execReq.Stdin, execReq.Stdout, execReq.Stderr)

	// Check Accept header
	accept := req.HeaderParameter("Accept")
	if accept == "" {
		accept = models.MediaTypeMultiplexedStream // Default to multiplexed stream
	}
	log.Printf("Accept header: %s", accept)

	// Validate Accept header
	if accept != models.MediaTypeRawStream && accept != models.MediaTypeMultiplexedStream {
		log.Printf("Unsupported Accept header: %s", accept)
		writeError(resp, http.StatusNotAcceptable, "UNSUPPORTED_MEDIA_TYPE",
			fmt.Sprintf("Unsupported Accept header: %s", accept))
		return
	}

	// Handle command execution request
	err = h.handleCommandExecution(req.Request.Context(), container.ID, &execReq, resp, req)
	if err != nil {
		log.Printf("Error executing command: %v", err)
		writeError(resp, http.StatusInternalServerError, "EXEC_FAILED", fmt.Sprintf("Error executing command: %v", err))
	}
}

// handleCommandExecution handles command execution in a container
func (h *DockerBoxHandler) handleCommandExecution(ctx context.Context, containerID string, execReq *models.BoxExecRequest, resp *restful.Response, req *restful.Request) error {
	if len(execReq.Cmd) == 0 {
		return fmt.Errorf("command is required")
	}

	// Set default stream options if not specified
	if !execReq.Stdin && !execReq.Stdout && !execReq.Stderr {
		execReq.Stdout = true
		execReq.Stderr = true
	}

	// Create exec configuration
	execConfig := container.ExecOptions{
		User:         "", // Use default user
		Privileged:   false,
		Tty:          execReq.TTY,
		AttachStdin:  execReq.Stdin,
		AttachStderr: execReq.Stderr,
		AttachStdout: execReq.Stdout,
		Detach:       false,
		DetachKeys:   "",  // Use default detach keys
		Env:          nil, // No additional environment variables
		WorkingDir:   "",  // Use container's working directory
		Cmd:          append(execReq.Cmd, execReq.Args...),
	}

	log.Printf("Creating exec with config: %+v", execConfig)

	// Create exec instance
	execCreate, err := h.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %v", err)
	}
	log.Printf("Created exec instance: %s", execCreate.ID)

	// Attach to exec instance
	execAttach, err := h.client.ContainerExecAttach(ctx, execCreate.ID, container.ExecAttachOptions{
		Detach: false,
		Tty:    execReq.TTY,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to exec: %v", err)
	}
	// Don't defer execAttach.Close() here, we'll close it after streams are done
	log.Printf("Attached to exec instance")

	// Hijack the connection
	httpResp := resp.ResponseWriter
	clientConn, _, err := httpResp.(http.Hijacker).Hijack()
	if err != nil {
		execAttach.Close()
		return fmt.Errorf("failed to hijack connection: %v", err)
	}
	defer clientConn.Close()
	log.Printf("Hijacked connection")

	// Check if client requested protocol upgrade
	upgrade := req.HeaderParameter("Upgrade")
	connection := req.HeaderParameter("Connection")

	// Write HTTP response headers
	if upgrade == "tcp" && connection == "Upgrade" {
		// Client requested protocol upgrade
		fmt.Fprintf(clientConn, "HTTP/1.1 101 UPGRADED\r\n")
		fmt.Fprintf(clientConn, "Content-Type: %s\r\n", models.MediaTypeRawStream)
		fmt.Fprintf(clientConn, "Connection: Upgrade\r\n")
		fmt.Fprintf(clientConn, "Upgrade: tcp\r\n")
		log.Printf("Protocol upgrade requested, using raw stream")
	} else {
		// No protocol upgrade requested
		fmt.Fprintf(clientConn, "HTTP/1.1 200 OK\r\n")
		fmt.Fprintf(clientConn, "Content-Type: %s\r\n", models.MediaTypeMultiplexedStream)
		log.Printf("No protocol upgrade requested, using multiplexed stream")
	}
	fmt.Fprintf(clientConn, "\r\n")

	// Flush response headers
	if f, ok := clientConn.(http.Flusher); ok {
		f.Flush()
		log.Printf("Flushed response headers")
	}

	// Create error channels for stream handling
	stdinDone := make(chan struct{})
	stdoutDone := make(chan error, 1)

	// Start streaming
	if execReq.TTY || (upgrade == "tcp" && connection == "Upgrade") {
		// For TTY sessions or raw stream mode, directly copy the raw stream
		go func() {
			if _, err := io.Copy(clientConn, execAttach.Reader); err != nil {
				if err != io.EOF && !isConnectionClosed(err) {
					log.Printf("Error copying from container: %v", err)
					stdoutDone <- err
				}
			}
			// Close the client connection to signal EOF
			if closer, ok := clientConn.(interface{ CloseWrite() error }); ok {
				closer.CloseWrite()
			}
			close(stdoutDone)
		}()

		if execReq.Stdin {
			go func() {
				if _, err := io.Copy(execAttach.Conn, clientConn); err != nil {
					if err != io.EOF && !isConnectionClosed(err) {
						log.Printf("Error copying to container: %v", err)
					}
				}
				// Try to close write end of the connection if possible
				if closeWriter, ok := execAttach.Conn.(interface{ CloseWrite() error }); ok {
					if err := closeWriter.CloseWrite(); err != nil {
						log.Printf("Error closing write end: %v", err)
					}
				}
				close(stdinDone)
			}()
		} else {
			close(stdinDone)
		}
	} else {
		// For non-TTY sessions, use multiplexed streaming
		log.Printf("Starting multiplexed stream")
		go func() {
			h.streamMultiplexed(execAttach.Reader, clientConn)
			// Close the client connection to signal EOF
			if closer, ok := clientConn.(interface{ CloseWrite() error }); ok {
				closer.CloseWrite()
			}
			close(stdoutDone)
		}()

		if execReq.Stdin {
			go func() {
				h.handleStdin(clientConn, execAttach.Conn)
				close(stdinDone)
			}()
		} else {
			close(stdinDone)
		}
	}

	// Wait for stdin to finish first
	select {
	case <-ctx.Done():
		log.Printf("Context cancelled while waiting for stdin")
		execAttach.Close()
		return ctx.Err()
	case <-stdinDone:
		log.Printf("Stdin stream completed")
	}

	// Then wait for stdout/stderr to complete
	select {
	case <-ctx.Done():
		log.Printf("Context cancelled while waiting for stdout")
		execAttach.Close()
		return ctx.Err()
	case err := <-stdoutDone:
		if err != nil && !isConnectionClosed(err) {
			log.Printf("Stream error: %v", err)
			execAttach.Close()
			return err
		}
		log.Printf("Stdout stream completed")
	}

	// Get the exit status before closing the connection
	execInspect, err := h.client.ContainerExecInspect(ctx, execCreate.ID)
	if err != nil {
		log.Printf("Error inspecting exec: %v", err)
		execAttach.Close()
		return err
	}

	log.Printf("Exec completed with exit code: %d", execInspect.ExitCode)

	// Close all connections
	execAttach.Close()
	if closer, ok := clientConn.(io.Closer); ok {
		closer.Close()
	}

	if execInspect.ExitCode != 0 {
		return fmt.Errorf("command exited with code %d", execInspect.ExitCode)
	}

	return nil
}
