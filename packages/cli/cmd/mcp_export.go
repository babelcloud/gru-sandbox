package cmd

import (
	"github.com/spf13/cobra"
)

func NewMcpExportCommand() *cobra.Command {
	return &cobra.Command{
		Use:                "export",
		Short:              "mcp export operation",
		DisableFlagParsing: true,
		DisableAutoGenTag:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			allArgs := append([]string{"export"}, args...)
			return executeScript("mcp", allArgs)
		},
	}
}
