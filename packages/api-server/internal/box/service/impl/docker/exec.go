package docker

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"

	"github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/babelcloud/gbox/packages/api-server/internal/common"
)

// Exec implements Service.Exec
func (s *Service) Exec(ctx context.Context, id string, req *model.BoxExecParams) (*model.BoxExecResult, error) {
	// Update access time on exec
	s.accessTracker.Update(id)

	containerInfo, err := s.getContainerByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check container status
	if containerInfo.State != "running" {
		return nil, fmt.Errorf("box %s is not running (current state: %s)", id, containerInfo.State)
	}

	// Create exec configuration
	execConfig := types.ExecConfig{
		User:         "", // Use default user
		Privileged:   false,
		Tty:          req.TTY,
		AttachStdin:  req.Stdin,
		AttachStdout: true,
		AttachStderr: true,
		Detach:       false,
		DetachKeys:   "",  // Use default detach keys
		Env:          nil, // No additional environment variables
		WorkingDir:   common.DefaultWorkDirPath,
		Cmd:          append(req.Cmd, req.Args...),
	}

	// Create exec instance
	execResp, err := s.client.ContainerExecCreate(ctx, containerInfo.ID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}

	// Attach to exec instance
	attachResp, err := s.client.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{
		Detach: false,
		Tty:    req.TTY,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer attachResp.Close()

	// Create channels for stream handling
	stdinDone := make(chan struct{})
	stdoutDone := make(chan error, 1)

	// Start streaming
	if req.TTY {
		// For TTY sessions, directly copy the raw stream
		go func() {
			if _, err := io.Copy(req.Conn, attachResp.Reader); err != nil {
				if err != io.EOF && !isConnectionClosed(err) {
					stdoutDone <- err
				}
			}
			// Close the client connection to signal EOF
			if closer, ok := req.Conn.(interface{ CloseWrite() error }); ok {
				closer.CloseWrite()
			}
			close(stdoutDone)
		}()

		if execConfig.AttachStdin {
			go func() {
				s.handleStdin(req.Conn, attachResp.Conn)
				// Try to close write end of the connection if possible
				if closeWriter, ok := attachResp.Conn.(interface{ CloseWrite() error }); ok {
					if err := closeWriter.CloseWrite(); err != nil {
						s.logger.Error("Error closing write end: %v", err)
					}
				}
				close(stdinDone)
			}()
		} else {
			close(stdinDone)
		}
	} else {
		// For non-TTY sessions, use multiplexed streaming
		s.logger.Debug("Starting multiplexed stream")
		go func() {
			s.streamMultiplexed(attachResp.Reader, req.Conn)
			// Close the client connection to signal EOF
			if closer, ok := req.Conn.(interface{ CloseWrite() error }); ok {
				closer.CloseWrite()
			}
			close(stdoutDone)
		}()

		if execConfig.AttachStdin {
			go func() {
				s.handleStdin(req.Conn, attachResp.Conn)
				// Try to close write end of the connection if possible
				if closeWriter, ok := attachResp.Conn.(interface{ CloseWrite() error }); ok {
					if err := closeWriter.CloseWrite(); err != nil {
						s.logger.Error("Error closing write end: %v", err)
					}
				}
				close(stdinDone)
			}()
		} else {
			close(stdinDone)
		}
	}

	// Wait for stdin to finish first
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-stdinDone:
	}

	// Then wait for stdout/stderr to complete
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-stdoutDone:
		if err != nil && !isConnectionClosed(err) {
			return nil, fmt.Errorf("stream error: %w", err)
		}
	}

	// Get exit code
	inspectResp, err := s.client.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect exec: %w", err)
	}

	return &model.BoxExecResult{
		ExitCode: inspectResp.ExitCode,
	}, nil
}

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
			return "", "", fmt.Errorf("error reading stream header: %w", err)
		}

		// Parse header
		streamType := header[0]
		// Skip 3 bytes reserved for future use
		size := binary.BigEndian.Uint32(header[4:])

		// Read payload
		payload := make([]byte, size)
		_, err = io.ReadFull(reader, payload)
		if err != nil {
			return "", "", fmt.Errorf("error reading stream payload: %w", err)
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
func (s *Service) collectOutput(reader io.Reader, stdoutLimit, stderrLimit int) (string, string) {
	stdout, stderr, err := readDockerStream(reader)
	if err != nil {
		s.logger.Error("Error reading Docker stream: %v", err)
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

// Run implements Service.Run
func (s *Service) Run(ctx context.Context, id string, req *model.BoxRunParams) (*model.BoxRunResult, error) {
	// Update access time on run
	s.accessTracker.Update(id)

	containerInfo, err := s.getContainerByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check container status
	if containerInfo.State != "running" {
		return nil, fmt.Errorf("box %s is not running (current state: %s)", id, containerInfo.State)
	}

	// Set default line limits if not specified
	if req.StdoutLineLimit == 0 {
		req.StdoutLineLimit = 100
	}
	if req.StderrLineLimit == 0 {
		req.StderrLineLimit = 100
	}

	execConfig := types.ExecConfig{
		User:         "", // Use default user
		Privileged:   false,
		Tty:          false, // Run commands typically don't need TTY
		AttachStdin:  req.Stdin != "",
		AttachStdout: true,
		AttachStderr: true,
		Detach:       false,
		DetachKeys:   "",  // Use default detach keys
		Env:          nil, // No additional environment variables
		WorkingDir:   common.DefaultWorkDirPath,
		Cmd:          append(req.Cmd, req.Args...),
	}

	execResp, err := s.client.ContainerExecCreate(ctx, containerInfo.ID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}

	attachResp, err := s.client.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer attachResp.Close()

	// Create channels for collecting output
	outputChan := make(chan struct {
		stdout string
		stderr string
	})
	exitCodeChan := make(chan int)

	// Start goroutine to collect output
	go func() {
		stdout, stderr := s.collectOutput(attachResp.Reader, req.StdoutLineLimit, req.StderrLineLimit)
		outputChan <- struct {
			stdout string
			stderr string
		}{stdout, stderr}
	}()

	// Write stdin if provided
	if req.Stdin != "" {
		go func() {
			_, err := io.WriteString(attachResp.Conn, req.Stdin)
			if err != nil {
				s.logger.Error("Error writing stdin: %v", err)
			}
			// Close write end of the connection to signal EOF
			if closer, ok := attachResp.Conn.(interface{ CloseWrite() error }); ok {
				closer.CloseWrite()
			}
		}()
	}

	// Wait for exec to complete and get exit code
	go func() {
		inspectResp, err := s.client.ContainerExecInspect(ctx, execResp.ID)
		if err != nil {
			s.logger.Error("Error inspecting exec: %v", err)
			exitCodeChan <- -1
			return
		}
		exitCodeChan <- inspectResp.ExitCode
	}()

	// Collect results
	output := <-outputChan
	exitCode := <-exitCodeChan

	return &model.BoxRunResult{
		ExitCode: exitCode,
		Stdout:   output.stdout,
		Stderr:   output.stderr,
	}, nil
}

// isConnectionClosed checks if the error is due to a closed connection
func isConnectionClosed(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "use of closed network connection") ||
		strings.Contains(err.Error(), "io: read/write on closed pipe")
}

// streamMultiplexed handles multiplexed streaming of stdout and stderr
func (s *Service) streamMultiplexed(reader io.Reader, writer io.Writer) {
	header := make([]byte, 8)
	for {
		// Read header
		_, err := io.ReadFull(reader, header)
		if err != nil {
			if err != io.EOF && !isConnectionClosed(err) {
				s.logger.Error("Error reading stream header: %v", err)
			}
			return
		}

		// Parse header
		streamType := header[0]
		frameSize := binary.BigEndian.Uint32(header[4:])

		// Read frame
		frame := make([]byte, frameSize)
		_, err = io.ReadFull(reader, frame)
		if err != nil {
			if err != io.EOF && !isConnectionClosed(err) {
				s.logger.Error("Error reading stream frame: %v", err)
			}
			return
		}

		// Write header and frame to client
		if _, err := writer.Write(header); err != nil {
			if !isConnectionClosed(err) {
				s.logger.Error("Error writing stream header: %v", err)
			}
			return
		}
		if _, err := writer.Write(frame); err != nil {
			if !isConnectionClosed(err) {
				s.logger.Error("Error writing stream frame: %v", err)
			}
			return
		}

		// Log stream type and size for debugging
		s.logger.Debug("Stream type: %d, size: %d", streamType, frameSize)
	}
}

// handleStdin handles stdin stream
func (s *Service) handleStdin(reader io.Reader, writer io.Writer) {
	_, err := io.Copy(writer, reader)
	if err != nil && !isConnectionClosed(err) {
		s.logger.Error("Error copying stdin: %v", err)
	}
	// Signal EOF to the container
	if closer, ok := writer.(interface{ CloseWrite() error }); ok {
		if err := closer.CloseWrite(); err != nil {
			s.logger.Error("Error closing write end: %v", err)
		}
	}
}
