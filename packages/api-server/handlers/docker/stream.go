package docker

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"strings"
	"syscall"
)

// streamMultiplexed copies the container's output stream to the client with Docker's multiplexing protocol
func (h *DockerBoxHandler) streamMultiplexed(reader io.Reader, writer io.Writer) {
	header := make([]byte, 8)
	for {
		// Read header
		_, err := io.ReadFull(reader, header)
		if err != nil {
			if err != io.EOF && !isConnectionClosed(err) {
				log.Printf("Error reading stream header: %v", err)
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
				log.Printf("Error reading stream frame: %v", err)
			}
			return
		}

		// Write header and frame to client
		if _, err := writer.Write(header); err != nil {
			if !isConnectionClosed(err) {
				log.Printf("Error writing stream header: %v", err)
			}
			return
		}
		if _, err := writer.Write(frame); err != nil {
			if !isConnectionClosed(err) {
				log.Printf("Error writing stream frame: %v", err)
			}
			return
		}

		// Log stream type and size for debugging
		log.Printf("Stream type: %d, size: %d", streamType, frameSize)
	}
}

// handleStdin reads from client and writes to container's stdin with Docker's multiplexing protocol
func (h *DockerBoxHandler) handleStdin(reader io.Reader, writer io.Writer) {
	header := make([]byte, 8)
	for {
		// Read header
		_, err := io.ReadFull(reader, header)
		if err != nil {
			if err != io.EOF && !isConnectionClosed(err) {
				log.Printf("Error reading stdin header: %v", err)
			}
			return
		}

		// Parse header
		frameSize := binary.BigEndian.Uint32(header[4:])

		// Read frame
		frame := make([]byte, frameSize)
		_, err = io.ReadFull(reader, frame)
		if err != nil {
			if err != io.EOF && !isConnectionClosed(err) {
				log.Printf("Error reading stdin frame: %v", err)
			}
			return
		}

		// Write frame to container
		if _, err := writer.Write(frame); err != nil {
			if !isConnectionClosed(err) {
				log.Printf("Error writing to container stdin: %v", err)
			}
			return
		}

		log.Printf("Wrote %d bytes to container stdin", frameSize)
	}
}

// isConnectionClosed checks if the error indicates a closed connection
func isConnectionClosed(err error) bool {
	if err == nil {
		return false
	}

	if err == io.EOF {
		return true
	}

	if strings.Contains(err.Error(), "use of closed network connection") {
		return true
	}

	if strings.Contains(err.Error(), "connection reset by peer") {
		return true
	}

	if strings.Contains(err.Error(), "broken pipe") {
		return true
	}

	if netErr, ok := err.(*net.OpError); ok {
		if netErr.Err == syscall.EPIPE || netErr.Err == syscall.ECONNRESET {
			return true
		}
	}

	return false
}
