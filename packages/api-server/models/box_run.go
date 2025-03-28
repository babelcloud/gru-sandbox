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
	BoxID    string `json:"boxId,omitempty"`    // ID of the box where the command was executed
	ExitCode int    `json:"exitCode,omitempty"` // Exit code of the command
	Stdout   string `json:"stdout,omitempty"`   // Standard output from command execution
	Stderr   string `json:"stderr,omitempty"`   // Standard error from command execution
}
