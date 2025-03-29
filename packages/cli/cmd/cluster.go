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

	// 设置自定义帮助函数
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		showHelp("all")
	})

	// 添加所有cluster相关的子命令
	cmd.AddCommand(
		NewClusterSetupCommand(),
		NewClusterCleanupCommand(),
	)

	return cmd
}
