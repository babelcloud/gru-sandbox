package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewClusterCleanupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up box environment and remove all boxes",
		Long:  "Clean up box environment and remove all boxes created by gbox",
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")
			return cleanupCluster(force)
		},
	}

	// Add flags
	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}

// cleanupCluster cleans up the cluster environment
func cleanupCluster(force bool) error {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to get user home directory: %v", err)
	}

	// Define config file path
	gboxHome := filepath.Join(homeDir, ".gbox")
	configFile := filepath.Join(gboxHome, "config.yml")

	// If config file doesn't exist, return directly
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("Cluster has been cleaned up.")
		return nil
	}

	// Get current mode
	mode, err := getCurrentMode(configFile)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		// Error is not fatal, continue execution
	}

	// If not in force mode, request confirmation
	if !force {
		var confirmMsg string
		if mode != "" {
			confirmMsg = fmt.Sprintf("This will delete all containers in %s mode. Continue? (y/N) ", mode)
		} else {
			confirmMsg = "This will delete all containers. Continue? (y/N) "
		}

		fmt.Print(confirmMsg)
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" && confirm != "yes" && confirm != "Yes" {
			fmt.Println("Cleanup cancelled")
			return nil
		}
	}

	// Perform cleanup based on mode
	if mode != "" {
		if mode == "docker" {
			if err := cleanupDocker(); err != nil {
				return err
			}
		} else if mode == "k8s" {
			if err := cleanupK8s(); err != nil {
				return err
			}
		}
	} else {
		// Try to clean up all modes
		cleanupDocker()
		cleanupK8s()
	}

	// Delete config file after cleanup
	if err := os.Remove(configFile); err != nil {
		return fmt.Errorf("failed to delete config file: %v", err)
	}

	return nil
}

// cleanupDocker cleans up Docker environment
func cleanupDocker() error {
	fmt.Println("Cleaning up docker environment...")

	// Stop docker-compose services
	fmt.Println("Stopping docker-compose services...")
	scriptDir, err := getScriptDir()
	if err != nil {
		return err
	}

	// Execute docker-compose down command
	composePath := filepath.Join(scriptDir, "..", "manifests", "docker", "docker-compose.yml")
	cmd := exec.Command("docker", "compose", "-f", composePath, "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop docker-compose services: %v", err)
	}

	fmt.Println("Docker environment cleanup complete")
	return nil
}

// cleanupK8s cleans up Kubernetes environment
func cleanupK8s() error {
	fmt.Println("Cleaning up k8s environment...")

	// Define cluster name
	gboxCluster := "gbox"

	// Delete cluster
	cmd := exec.Command("kind", "delete", "cluster", "--name", gboxCluster)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete cluster: %v", err)
	}

	fmt.Println("K8s environment cleanup complete")
	return nil
}
