package models

// BoxRunRequest represents a request to run a command in a box
type BoxRunRequest struct {
	Cmd             []string `json:"cmd,omitempty"`
	Args            []string `json:"args,omitempty"`
	Stdin           string   `json:"stdin,omitempty"`
	StdoutLineLimit int      `json:"stdoutLineLimit,omitempty"`
	StderrLineLimit int      `json:"stderrLineLimit,omitempty"`
}

// BoxRunResponse represents the response from a run operation
type BoxRunResponse struct {
	ExitCode int    `json:"exit_code,omitempty"` // Exit code of the command
	Stdout   string `json:"stdout,omitempty"`    // Standard output from command execution
	Stderr   string `json:"stderr,omitempty"`    // Standard error from command execution
}
