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

	// 设置自定义帮助函数
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		showHelp("all")
	})

	// 添加所有mcp相关的子命令
	cmd.AddCommand(
		NewMcpExportCommand(),
	)

	return cmd
}
