package cmd

import (
	"github.com/spf13/cobra"
)

func NewMcpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP configuration operations",
	}

	cmd.AddCommand(
		NewMcpExportCommand(),
	)

	return cmd
}
