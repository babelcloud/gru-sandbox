package cmd

import (
	"github.com/spf13/cobra"
)

func NewMcpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "mcp",
		Short:              getCommandDescription("mcp"),
		DisableFlagParsing: true,
		DisableAutoGenTag:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeScript("mcp", args)
		},
	}

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		showHelp("all")
	})

	cmd.AddCommand(
		NewMcpExportCommand(),
	)

	return cmd
}
