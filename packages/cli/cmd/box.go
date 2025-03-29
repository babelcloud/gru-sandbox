package cmd

import (
	"github.com/spf13/cobra"
)

func NewBoxCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "box",
		Short:              getCommandDescription("box"),
		DisableFlagParsing: true,
		DisableAutoGenTag:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeScript("box", args)
		},
	}

	// 添加所有box相关的子命令
	cmd.AddCommand(
		NewBoxCreateCommand(),
		NewBoxDeleteCommand(),
		NewBoxListCommand(),
		NewBoxExecCommand(),
		NewBoxInspectCommand(),
		NewBoxStartCommand(),
		NewBoxStopCommand(),
		NewBoxReclaimCommand(),
		NewBoxCpCommand(),
	)

	return cmd
}
