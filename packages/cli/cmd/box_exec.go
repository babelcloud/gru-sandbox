package cmd

import (
	"github.com/spf13/cobra"
)

type BoxExecRequest struct {
	Cmd     []string          `json:"cmd"`
	Env     map[string]string `json:"env,omitempty"`
	WorkDir string            `json:"workdir,omitempty"`
}

type BoxExecResponse struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

func NewBoxExecCommand() *cobra.Command {
	return &cobra.Command{
		Use:                "exec",
		Short:              "Execute a command in a box",
		DisableFlagParsing: true,
		DisableAutoGenTag:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			allArgs := append([]string{"exec"}, args...)
			return executeScript("box", allArgs)
		},
	}
}
