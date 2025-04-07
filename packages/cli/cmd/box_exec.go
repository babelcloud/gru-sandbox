package cmd

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/babelcloud/gru-sandbox/packages/cli/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// BoxExecOptions holds command options
type BoxExecOptions struct {
	Interactive bool
	Tty         bool
	BoxID       string
	Command     []string
}

// BoxExecRequest represents the request to execute a command in a box
type BoxExecRequest struct {
	Cmd      []string          `json:"cmd"`
	Args     []string          `json:"args,omitempty"`
	Stdin    bool              `json:"stdin"`
	Stdout   bool              `json:"stdout"`
	Stderr   bool              `json:"stderr"`
	Tty      bool              `json:"tty"`
	TermSize map[string]int    `json:"term_size,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	WorkDir  string            `json:"workdir,omitempty"`
}

// TerminalSize represents terminal dimensions
type TerminalSize struct {
	Height int
	Width  int
}

// GetTerminalSize returns the current terminal dimensions
func GetTerminalSize() (*TerminalSize, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil, fmt.Errorf("not a terminal")
	}

	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	return &TerminalSize{
		Height: height,
		Width:  width,
	}, nil
}

// NewBoxExecCommand creates a new box exec command
func NewBoxExecCommand() *cobra.Command {
	opts := &BoxExecOptions{}

	cmd := &cobra.Command{
		Use:   "exec [box-id] -- [command] [args...]",
		Short: "Execute a command in a box",
		Long: `usage: gbox-box-exec [-h] [-i] [-t] box_id

Execute a command in a box

positional arguments:
  box_id             ID of the box

options:
  -h, --help         show this help message and exit
  -i, --interactive  Enable interactive mode (with stdin)
  -t, --tty          Force TTY allocation`,
		Example: `    gbox box exec 550e8400-e29b-41d4-a716-446655440000 -- ls -l     # List files in box
    gbox box exec 550e8400-e29b-41d4-a716-446655440000 -t -- bash     # Run interactive bash
    gbox box exec 550e8400-e29b-41d4-a716-446655440000 -i -- cat       # Run cat with stdin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			argsLenAtDash := cmd.ArgsLenAtDash()
			if argsLenAtDash == -1 {
				return fmt.Errorf("command must be specified after '--'")
			}

			if len(args) == 0 || argsLenAtDash == 0 {
				cmd.Help()
				return fmt.Errorf("box ID is required")
			}

			// Get box ID (first argument before --)
			opts.BoxID = args[0]

			// Get command (all arguments after --)
			if argsLenAtDash >= len(args) {
				return fmt.Errorf("command must be specified after '--'")
			}
			opts.Command = args[argsLenAtDash:]

			// Run the command
			return runExec(opts)
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&opts.Interactive, "interactive", "i", false, "Enable interactive mode (with stdin)")
	cmd.Flags().BoolVarP(&opts.Tty, "tty", "t", false, "Force TTY allocation")

	return cmd
}

// runExec implements the exec command functionality
func runExec(opts *BoxExecOptions) error {
	debug := os.Getenv("DEBUG") == "true"
	apiBase := config.GetAPIURL()
	apiURL := fmt.Sprintf("%s/api/v1", strings.TrimSuffix(apiBase, "/"))

	// Debug log
	debugLog := func(msg string) {
		if debug {
			fmt.Fprintf(os.Stderr, "DEBUG: %s\n", msg)
		}
	}

	// Check if stdin is available and interactive mode is enabled
	stdinAvailable := !term.IsTerminal(int(os.Stdin.Fd())) && opts.Interactive

	// Get terminal size if in TTY mode
	var termSize *TerminalSize
	var err error
	if opts.Tty {
		termSize, err = GetTerminalSize()
		if err != nil {
			debugLog(fmt.Sprintf("Failed to get terminal size: %v", err))
		} else {
			debugLog(fmt.Sprintf("Terminal size: height=%d, width=%d", termSize.Height, termSize.Width))
		}
	}

	// Prepare request body
	request := BoxExecRequest{
		Cmd:    []string{opts.Command[0]},
		Args:   opts.Command[1:],
		Stdin:  stdinAvailable || opts.Tty, // Always enable stdin in TTY mode
		Stdout: true,
		Stderr: true,
		Tty:    opts.Tty,
	}

	if termSize != nil {
		request.TermSize = map[string]int{
			"height": termSize.Height,
			"width":  termSize.Width,
		}
	}

	// Encode request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to encode request: %v", err)
	}

	debugLog(fmt.Sprintf("Request body: %s", string(requestBody)))

	// Create HTTP request
	requestURL := fmt.Sprintf("%s/boxes/%s/exec", apiURL, opts.BoxID)
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Upgrade", "tcp")
	req.Header.Set("Connection", "Upgrade")

	// Set appropriate Accept header based on TTY mode
	if opts.Tty {
		// In TTY mode, use raw stream
		req.Header.Set("Accept", "application/vnd.gbox.raw-stream")
	} else {
		// In non-TTY mode, use multiplexed stream
		req.Header.Set("Accept", "application/vnd.gbox.multiplexed-stream")
	}

	debugLog(fmt.Sprintf("Sending request to: POST %s", requestURL))
	for k, v := range req.Header {
		debugLog(fmt.Sprintf("Header %s: %s", k, v))
	}

	// Perform HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	debugLog(fmt.Sprintf("Response status: %d", resp.StatusCode))
	for k, v := range resp.Header {
		debugLog(fmt.Sprintf("Response header %s: %s", k, v))
	}

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusSwitchingProtocols {
		var errMsg string
		body, _ := io.ReadAll(resp.Body)
		debugLog(fmt.Sprintf("Response body: %s", string(body)))

		// Try to parse error message from JSON
		var errorData map[string]interface{}
		if err := json.Unmarshal(body, &errorData); err == nil {
			if message, ok := errorData["message"].(string); ok {
				errMsg = message
			} else {
				errMsg = fmt.Sprintf("Server returned status code %d", resp.StatusCode)
			}
		} else {
			errMsg = fmt.Sprintf("Server returned status code %d: %s", resp.StatusCode, string(body))
		}

		return fmt.Errorf("%s", errMsg)
	}

	// Get hijacked connection
	hijacker, ok := resp.Body.(io.ReadWriteCloser)
	if !ok {
		return fmt.Errorf("response does not support hijacking")
	}

	// Handle communication based on TTY mode
	if opts.Tty {
		return handleRawStream(hijacker)
	} else {
		return handleMultiplexedStream(hijacker, stdinAvailable)
	}
}

// handleRawStream handles raw stream in TTY mode
func handleRawStream(conn io.ReadWriteCloser) error {
	// Save terminal state
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set terminal to raw mode: %v", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Handle terminal resize
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	defer signal.Stop(sigChan)

	// Start goroutine for reading from connection
	errChan := make(chan error, 2)
	go func() {
		_, err := io.Copy(os.Stdout, conn)
		errChan <- err
	}()

	// Start goroutine for writing to connection
	go func() {
		_, err := io.Copy(conn, os.Stdin)
		errChan <- err
	}()

	// Wait for either an error or for both goroutines to finish
	err = <-errChan
	if err != nil && err != io.EOF {
		return fmt.Errorf("stream error: %v", err)
	}

	return nil
}

// handleMultiplexedStream handles multiplexed stream in non-TTY mode
func handleMultiplexedStream(conn io.ReadWriter, stdinAvailable bool) error {
	// Start goroutine for handling stdin if available
	if stdinAvailable {
		go func() {
			stdin := bufio.NewReader(os.Stdin)
			for {
				data := make([]byte, 1024)
				n, err := stdin.Read(data)
				if err != nil {
					if err != io.EOF {
						fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
					}
					return
				}

				_, err = conn.Write(data[:n])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to connection: %v\n", err)
					return
				}
			}
		}()
	}

	// Read multiplexed stream
	buf := make([]byte, 8)
	for {
		// Read header (8 bytes)
		n, err := io.ReadFull(conn, buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("failed to read header: %v", err)
		}
		if n != 8 {
			return fmt.Errorf("short read: got %d bytes, expected 8", n)
		}

		// Parse header
		streamType := buf[0]
		size := binary.BigEndian.Uint32(buf[4:])

		// Read payload
		payload := make([]byte, size)
		n, err = io.ReadFull(conn, payload)
		if err != nil {
			return fmt.Errorf("failed to read payload: %v", err)
		}
		if uint32(n) != size {
			return fmt.Errorf("short read: got %d bytes, expected %d", n, size)
		}

		// Write to appropriate output
		switch streamType {
		case 1: // stdout
			os.Stdout.Write(payload)
		case 2: // stderr
			os.Stderr.Write(payload)
		default:
			fmt.Fprintf(os.Stderr, "Unknown stream type: %d\n", streamType)
		}
	}
}

// hijackConn attempts to hijack the HTTP connection to a raw TCP connection
func hijackConn(resp *http.Response) (net.Conn, error) {
	// Check if the connection can be hijacked
	hj, ok := resp.Body.(interface {
		Hijack() (net.Conn, *bufio.ReadWriter, error)
	})
	if !ok {
		return nil, fmt.Errorf("connection does not support hijacking")
	}

	conn, _, err := hj.Hijack()
	if err != nil {
		return nil, fmt.Errorf("failed to hijack connection: %v", err)
	}

	return conn, nil
}
