package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// Show help information
func showHelp(helpType string) {
	switch helpType {
	case "short":
		fmt.Println("Box management tool")
		return
	case "all":
		fmt.Println("Usage: gbox <command> [arguments]")
		fmt.Println("\nAvailable Commands:")

		// Display alias commands in fixed order
		aliases := []string{"setup", "cleanup", "export"}
		for _, alias := range aliases {
			if cmd, ok := aliasMap[alias]; ok {
				parts := strings.Split(cmd, " ")
				scriptPath := filepath.Join(scriptDir, fmt.Sprintf("gbox-%s", parts[0]))
				if _, err := os.Stat(scriptPath); err == nil {
					description := getSubCommandDescription(parts[0], parts[1])
					fmt.Printf("    %-18s %s\n", alias, description)
				}
			}
		}
		fmt.Printf("    %-18s %s\n", "help", "Show help information")

		fmt.Println("\nSub Commands:")
		// Display main commands in fixed order
		for _, cmd := range []string{"box", "cluster", "mcp"} {
			scriptPath := filepath.Join(scriptDir, fmt.Sprintf("gbox-%s", cmd))
			if _, err := os.Stat(scriptPath); err == nil {
				description := getCommandDescription(cmd)
				fmt.Printf("    %-18s %s\n", cmd, description)
			}
		}

		fmt.Println("\nOptions:")
		fmt.Println("    --help [short|all]  Show this help message (default: all)")

		fmt.Println("\nExamples:")
		fmt.Println("    gbox setup                 # Initialize the environment")
		fmt.Println("    gbox box create mybox      # Create a new box")
		fmt.Println("    gbox box list              # List all boxes")
		fmt.Println("    gbox export                # Export MCP configuration")
		fmt.Println("    gbox cleanup               # Clean up everything")

		fmt.Println("\nUse \"gbox <command> --help\" for more information about a command.")
	default:
		fmt.Fprintf(os.Stderr, "Invalid help type: %s\n", helpType)
		fmt.Fprintln(os.Stderr, "Valid types are: short, all")
		os.Exit(1)
	}
}

// Get description for a subcommand
func getSubCommandDescription(cmdName, subCmd string) string {
	scriptPath := filepath.Join(scriptDir, fmt.Sprintf("gbox-%s", cmdName))
	cmd := exec.Command(scriptPath, subCmd, "--help", "short")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("%s %s", cmdName, subCmd)
	}
	return strings.TrimSpace(string(output))
}

// Get description for a command
func getCommandDescription(cmdName string) string {
	scriptPath := filepath.Join(scriptDir, fmt.Sprintf("gbox-%s", cmdName))
	cmd := exec.Command(scriptPath, "--help", "short")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("%s operations", cmdName)
	}
	return strings.TrimSpace(string(output))
}

// Setup help command
func setupHelpCommand(rootCmd *cobra.Command) {
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		showHelp("all")
	})
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "help",
		Short:  "Show help information",
		Hidden: false,
		Run: func(cmd *cobra.Command, args []string) {
			helpType := "all"
			if len(args) > 0 {
				helpType = args[0]
			}
			showHelp(helpType)
		},
	})
	rootCmd.PersistentFlags().BoolP("help", "", false, "")
	rootCmd.PersistentFlags().MarkHidden("help")
}
