package models

// BoxRunRequest represents a request to run a command in a box
type BoxRunRequest struct {
	// Container creation options
	Image           string            `json:"image,omitempty"`           // Image to use for the container
	ImagePullSecret string            `json:"imagePullSecret,omitempty"` // For docker: base64 encoded auth string
	Env             map[string]string `json:"env,omitempty"`            // Environment variables
	WorkingDir      string            `json:"workingDir,omitempty"`     // Working directory
	ExtraLabels     map[string]string `json:"labels,omitempty"`         // Additional labels
	Mounts          []Mount           `json:"mounts,omitempty"`         // Volume mounts

	// Command execution options
	Cmd             []string `json:"cmd,omitempty"`             // Command to run
	Args            []string `json:"args,omitempty"`            // Command arguments
	Stdin           string   `json:"stdin,omitempty"`           // Standard input to send
	StdoutLineLimit int      `json:"stdoutLineLimit,omitempty"` // Max lines to return from stdout
	StderrLineLimit int      `json:"stderrLineLimit,omitempty"` // Max lines to return from stderr
}

// BoxRunResponse represents the response from a run operation
type BoxRunResponse struct {
	Box      Box    `json:"box"`                // Box where the command was executed
	ExitCode int    `json:"exitCode,omitempty"` // Exit code of the command
	Stdout   string `json:"stdout,omitempty"`   // Standard output from command execution
	Stderr   string `json:"stderr,omitempty"`   // Standard error from command execution
}
