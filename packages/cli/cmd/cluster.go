package cmd

import (
	"github.com/spf13/cobra"
)

func NewClusterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "cluster",
		Short:              getCommandDescription("cluster"),
		DisableFlagParsing: true,
		DisableAutoGenTag:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeScript("cluster", args)
		},
	}

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		showHelp("all")
	})

	cmd.AddCommand(
		NewClusterSetupCommand(),
		NewClusterCleanupCommand(),
	)

	return cmd
}
