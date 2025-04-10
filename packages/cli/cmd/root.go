package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/babelcloud/gru-sandbox/packages/cli/config"
	"github.com/babelcloud/gru-sandbox/packages/cli/internal/version"
	"github.com/spf13/cobra"
)

var (
	aliasMap = map[string]string{
		"setup":   "cluster setup",
		"cleanup": "cluster cleanup",
		"export":  "mcp export",
	}

	scriptDir string

	rootCmd = &cobra.Command{
		Use:   "gbox",
		Short: "Gru CLI Tool",
		Long: `Gru CLI is a command-line tool for managing and operating box, cluster, and mcp resources.
It provides a set of commands to create, manage, and operate these resources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flag("version").Changed {
				info := version.ClientInfo()
				fmt.Printf("GBOX version %s, build %s\n", info["Version"], info["GitCommit"])
				return nil
			}
			return cmd.Help()
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting executable path: %v\n", err)
		os.Exit(1)
	}

	_, isGoRun := os.LookupEnv("GORUN")

	if isGoRun {
		projectRoot := filepath.Clean(getProjectRoot())
		scriptDir = filepath.Join(projectRoot, "packages/cli/cmd/script")
	} else {
		realExePath, err := filepath.EvalSymlinks(exePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving symlink: %v\n", err)
			os.Exit(1)
		}
		exeDir := filepath.Dir(realExePath)
		scriptDir = filepath.Join(exeDir, "cmd", "script")
	}

	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print version information and exit")

	for alias, cmd := range aliasMap {
		createAliasCommand(alias, cmd)
	}

	rootCmd.AddCommand(NewBoxCommand())
	rootCmd.AddCommand(NewClusterCommand())
	rootCmd.AddCommand(NewMcpCommand())
	rootCmd.AddCommand(NewVersionCommand())
}

func createAliasCommand(alias, targetCmd string) {
	parts := strings.Split(targetCmd, " ")
	aliasCmd := &cobra.Command{
		Use:   alias,
		Short: fmt.Sprintf("Alias for '%s'", targetCmd),
		RunE: func(cmd *cobra.Command, args []string) error {
			allArgs := append(parts[1:], args...)
			return executeScript(parts[0], allArgs)
		},
	}

	rootCmd.AddCommand(aliasCmd)
}

func executeScript(cmdName string, args []string) error {
	scriptPath := filepath.Join(scriptDir, fmt.Sprintf("gbox-%s", cmdName))

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("Script not found: %s", scriptPath)
	}

	cmd := exec.Command(scriptPath)
	if len(args) > 0 {
		cmd = exec.Command(scriptPath, args...)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// 获取项目根目录，优先使用配置文件中的设置
func getProjectRoot() string {
	// 使用config包中的GetProjectRoot函数
	projectRoot := config.GetProjectRoot()
	if projectRoot != "" {
		return projectRoot
	}

	// 如果配置中没有设置，尝试从当前目录向上查找
	currentDir, err := os.Getwd()
	if err == nil {
		return currentDir
	}

	// 最后的回退方案
	return "."
}
