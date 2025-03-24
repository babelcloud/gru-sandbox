package models

// BoxExecRequest represents a request to execute a command in a box
type BoxExecRequest struct {
	Cmd     []string `json:"cmd"`     // Command to execute
	Args    []string `json:"args,omitempty"`    // Arguments to pass to the command
	Stdin   bool     `json:"stdin,omitempty"`   // Whether to attach stdin
	Stdout  bool     `json:"stdout,omitempty"`  // Whether to attach stdout
	Stderr  bool     `json:"stderr,omitempty"`  // Whether to attach stderr
	TTY     bool     `json:"tty,omitempty"`     // Whether to allocate a TTY
}

// BoxExecResponse represents the response from an exec operation
type BoxExecResponse struct {
	ExitCode int    `json:"exit_code,omitempty"` // Exit code of the command
	Stdout   string `json:"stdout,omitempty"`    // Standard output from command execution
	Stderr   string `json:"stderr,omitempty"`    // Standard error from command execution
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