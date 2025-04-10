package cmd

import (
	"github.com/spf13/cobra"
)

func NewClusterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "cluster",
		Short:              "Manage clusters (setup/cleanup)",
		DisableFlagParsing: true,
		DisableAutoGenTag:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeScript("cluster", args)
		},
	}

	return cmd
}
