package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

	rootCmd = &cobra.Command{
		Use:                "gru",
		Short:              "Gru CLI tool",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			showHelp("all")
		},
		// Completely disable Cobra's default help
		DisableAutoGenTag:          true,
		DisableFlagsInUseLine:      true,
		DisableSuggestions:         true,
		SuggestionsMinimumDistance: 1000,
		// Override default help
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && args[0] == "--help" {
				showHelp("all")
				return nil
			}
			showHelp("all")
			return nil
		},
	}
)

func init() {
	cobra.EnableCommandSorting = false

	// Setup help command
	setupHelpCommand(rootCmd)

	// Add alias commands
	for alias, cmd := range aliasMap {
		createAliasCommand(alias, cmd)
	}

	// Add main commands
	rootCmd.AddCommand(NewBoxCommand())
	rootCmd.AddCommand(NewClusterCommand())
	rootCmd.AddCommand(NewMcpCommand())
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// Create alias command
func createAliasCommand(alias, targetCmd string) {
	parts := strings.Split(targetCmd, " ")
	aliasCmd := &cobra.Command{
		Use:                alias,
		Short:              fmt.Sprintf("Alias for '%s'", targetCmd),
		DisableFlagParsing: true,
		DisableAutoGenTag:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && args[0] == "--help" {
				allArgs := append([]string{parts[1], "--help"}, args[1:]...)
				return executeScript(parts[0], allArgs)
			}
			allArgs := append(parts[1:], args...)
			return executeScript(parts[0], allArgs)
		},
	}
	// Set alias command help function
	aliasCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		allArgs := append([]string{parts[1], "--help"}, args...)
		executeScript(parts[0], allArgs)
	})
	rootCmd.AddCommand(aliasCmd)
}

// Execute script
func executeScript(cmdName string, args []string) error {
	scriptPath := filepath.Join(scriptDir, fmt.Sprintf("gbox-%s", cmdName))

	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script not found: %s", scriptPath)
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
