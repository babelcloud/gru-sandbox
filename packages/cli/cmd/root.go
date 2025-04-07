package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/babelcloud/gru-sandbox/packages/cli/internal/version"
	"github.com/spf13/cobra"
)

var (
	// Command alias mapping
	aliasMap = map[string]string{
		"setup":   "cluster setup",
		"cleanup": "cluster cleanup",
		"export":  "mcp export",
	}

	// Script directory (relative to packages/cli)
	scriptDir = filepath.Clean("../../bin")

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "gru",
		Short: "Gru CLI Tool",
		Long: `Gru CLI is a command-line tool for managing and operating box, cluster, and mcp resources.
It provides a set of commands to create, manage, and operate these resources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If version flag is set, print the version and exit
			if cmd.Flag("version").Changed {
				info := version.ClientInfo()
				fmt.Printf("GBOX version %s, build %s\n", info["Version"], info["GitCommit"])
				return nil
			}
			return cmd.Help()
		},
	}
)

// Execute runs the root command and handles any errors
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add global version flag
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print version information and exit")

	// Add alias commands
	for alias, cmd := range aliasMap {
		createAliasCommand(alias, cmd)
	}

	// Add main commands
	rootCmd.AddCommand(NewBoxCommand())
	rootCmd.AddCommand(NewClusterCommand())
	rootCmd.AddCommand(NewMcpCommand())
	rootCmd.AddCommand(NewVersionCommand())
}

// createAliasCommand creates a new command that acts as an alias to another command
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

// executeScript runs an external script with the given arguments
func executeScript(cmdName string, args []string) error {
	scriptPath := filepath.Join(scriptDir, fmt.Sprintf("gbox-%s", cmdName))

	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("Script not found: %s", scriptPath)
	}

	// Prepare command
	cmd := exec.Command(scriptPath)
	if len(args) > 0 {
		cmd = exec.Command(scriptPath, args...)
	}

	// Set standard input/output
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute command
	return cmd.Run()
}
