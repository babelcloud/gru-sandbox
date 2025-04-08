package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/babelcloud/gru-sandbox/packages/cli/config"
	"github.com/spf13/cobra"
)

type McpConfig struct {
	McpServers map[string]struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	} `json:"mcpServers"`
}

func NewMcpExportCommand() *cobra.Command {
	var mergeTo string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export MCP configuration for Claude Desktop",
		RunE: func(cmd *cobra.Command, args []string) error {
			return exportConfig(mergeTo, dryRun)
		},
	}

	cmd.Flags().StringVar(&mergeTo, "merge-to", "", "Merge configuration into target config file (claude|cursor)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview merge result without applying changes")

	return cmd
}

func getPackagesRootPath() (string, error) {
	projectRoot := config.GetProjectRoot()
	if projectRoot != "" {
		packagesDir := filepath.Join(projectRoot, "packages")
		if dirExists(packagesDir) {
			return packagesDir, nil
		}
	}

	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	standardSubPath := filepath.Join("packages", "cli", "build")

	if !strings.Contains(execPath, standardSubPath) {
		return "", fmt.Errorf("unexpected binary location: %s; expected to contain %q", execPath, standardSubPath)
	}

	packagesIndex := strings.Index(execPath, standardSubPath)
	if packagesIndex == -1 {
		return "", fmt.Errorf("could not find %q in path: %s", standardSubPath, execPath)
	}

	packagesDir := execPath[:packagesIndex]
	packagesDir = filepath.Clean(filepath.Join(packagesDir, "packages"))

	if dirExists(filepath.Join(packagesDir, "mcp-server")) {
		return packagesDir, nil
	}

	return "", fmt.Errorf("could not determine project root from binary location: %s", execPath)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func exportConfig(mergeTo string, dryRun bool) error {
	packagesRoot, err := getPackagesRootPath()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	mcpServerDir := filepath.Join(packagesRoot, "mcp-server")
	serverScript := filepath.Join(mcpServerDir, "dist", "index.js")

	if _, err := os.Stat(serverScript); os.IsNotExist(err) {
		return fmt.Errorf("server script not found at %s\nPlease build the MCP server first by running:\n  cd %s && pnpm build", serverScript, mcpServerDir)
	}

	serverScriptAbs, err := filepath.Abs(serverScript)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for server script: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeConfig := filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	cursorConfig := filepath.Join(homeDir, ".cursor", "mcp.json")

	config := McpConfig{
		McpServers: map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		}{
			"gbox": {},
		},
	}

	if os.Getenv("DEBUG") == "true" {
		config.McpServers["gbox"] = struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		}{
			Command: "bash",
			Args: []string{
				"-c",
				fmt.Sprintf("cd %s && pnpm --silent dev", mcpServerDir),
			},
		}
	} else {
		config.McpServers["gbox"] = struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		}{
			Command: "node",
			Args:    []string{serverScriptAbs},
		}
	}

	if mergeTo != "" {
		if mergeTo != "claude" && mergeTo != "cursor" {
			return fmt.Errorf("--merge-to target must be either 'claude' or 'cursor'")
		}

		targetConfig := claudeConfig
		if mergeTo == "cursor" {
			targetConfig = cursorConfig
		}

		if err := os.MkdirAll(filepath.Dir(targetConfig), 0755); err != nil {
			return fmt.Errorf("failed to create target directory: %w", err)
		}

		if dryRun {
			fmt.Println("Preview of merged configuration:")
			fmt.Println("----------------------------------------")

			merged, err := mergeConfigs(targetConfig, config)
			if err != nil {
				return fmt.Errorf("failed to merge configurations: %w", err)
			}
			output, _ := json.MarshalIndent(merged, "", "  ")
			fmt.Println(string(output))
		} else {
			merged, err := mergeConfigs(targetConfig, config)
			if err != nil {
				return fmt.Errorf("failed to merge configurations: %w", err)
			}
			output, _ := json.MarshalIndent(merged, "", "  ")
			if err := os.WriteFile(targetConfig, output, 0644); err != nil {
				return fmt.Errorf("failed to write configuration: %w", err)
			}
			fmt.Printf("Configuration merged into %s\n", targetConfig)
		}
	} else {
		output, _ := json.MarshalIndent(config, "", "  ")
		fmt.Println(string(output))
		fmt.Println()
		fmt.Println("To merge this configuration, run:")
		fmt.Println("  gbox mcp export --merge-to claude   # For Claude Desktop")
		fmt.Println("  gbox mcp export --merge-to cursor   # For Cursor")
	}

	return nil
}

func mergeConfigs(targetPath string, newConfig McpConfig) (McpConfig, error) {
	var existing McpConfig

	content, err := os.ReadFile(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return newConfig, nil
		}
		return McpConfig{}, fmt.Errorf("failed to read target config: %w", err)
	}

	if len(content) == 0 {
		return newConfig, nil
	}

	if err := json.Unmarshal(content, &existing); err != nil {
		return McpConfig{}, fmt.Errorf("invalid JSON in target config: %w", err)
	}

	// Merge mcpServers
	if existing.McpServers == nil {
		existing.McpServers = make(map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		})
	}

	for k, v := range newConfig.McpServers {
		existing.McpServers[k] = v
	}

	return existing, nil
}
