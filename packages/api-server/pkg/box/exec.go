package model

import (
	"io"
)

// BoxExecParams represents a request to execute a command in a box
type BoxExecParams struct {
	Cmd    []string           `json:"cmd"`              // Command to execute
	Args   []string           `json:"args,omitempty"`   // Arguments to pass to the command
	Stdin  bool               `json:"stdin,omitempty"`  // Whether to attach stdin
	Stdout bool               `json:"stdout,omitempty"` // Whether to attach stdout
	Stderr bool               `json:"stderr,omitempty"` // Whether to attach stderr
	TTY    bool               `json:"tty,omitempty"`    // Whether to allocate a TTY
	Conn   io.ReadWriteCloser `json:"-"`                // Connection for streaming
}

// BoxExecResult represents the response from an exec operation
type BoxExecResult struct {
	ExitCode int `json:"exitCode,omitempty"` // Exit code of the command
}

// BoxRunParams represents a request to run a command in a box
type BoxRunParams struct {
	Cmd             []string `json:"cmd,omitempty"`
	Args            []string `json:"args,omitempty"`
	Stdin           string   `json:"stdin,omitempty"`
	StdoutLineLimit int      `json:"stdoutLineLimit,omitempty"`
	StderrLineLimit int      `json:"stderrLineLimit,omitempty"`
}

// BoxRunResult represents the response from a run operation
type BoxRunResult struct {
	Box      Box    `json:"box"`                // Box where the command was executed
	ExitCode int    `json:"exitCode,omitempty"` // Exit code of the command
	Stdout   string `json:"stdout,omitempty"`   // Standard output from command execution
	Stderr   string `json:"stderr,omitempty"`   // Standard error from command execution
}

// StreamType represents the type of stream in multiplexed output
type StreamType byte

const (
	StreamStdin  StreamType = 0 // stdin (written to stdout)
	StreamStdout StreamType = 1 // stdout
	StreamStderr StreamType = 2 // stderr

	// MediaTypeRawStream is the MIME-Type for raw TTY streams
	MediaTypeRawStream = "application/vnd.gbox.raw-stream"

	// MediaTypeMultiplexedStream is the MIME-Type for stdin/stdout/stderr multiplexed streams
	MediaTypeMultiplexedStream = "application/vnd.gbox.multiplexed-stream"
)

// StreamHeader represents the header of a multiplexed stream frame
type StreamHeader struct {
	Type StreamType // Stream type (0: stdin, 1: stdout, 2: stderr)
	Size uint32     // Size of the frame payload
}
