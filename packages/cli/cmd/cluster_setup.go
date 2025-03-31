package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func NewClusterSetupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup box environment",
		Long:  "Setup the box environment with specified cluster mode (docker or k8s)",
		RunE: func(cmd *cobra.Command, args []string) error {
			mode, _ := cmd.Flags().GetString("mode")
			return setupCluster(mode)
		},
	}

	// Add flags
	cmd.Flags().String("mode", "docker", "Cluster mode (docker or k8s)")

	return cmd
}

// setupCluster sets up the cluster environment
func setupCluster(mode string) error {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to get user home directory: %v", err)
	}

	// Define config file path
	gboxHome := filepath.Join(homeDir, ".gbox")
	configFile := filepath.Join(gboxHome, "config.yml")

	// Get current mode (if exists)
	currentMode, err := getCurrentMode(configFile)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		// Error is not fatal, continue execution
	}

	// If mode not specified in command line and current config has a mode, use current config mode
	if mode == "docker" && currentMode != "" {
		mode = currentMode
	}

	// Validate mode
	if mode != "docker" && mode != "k8s" {
		return fmt.Errorf("invalid mode: %s. Must be 'docker' or 'k8s'", mode)
	}

	// Check if mode has changed
	if currentMode != "" && currentMode != mode {
		return fmt.Errorf("error: cannot change from '%s' mode to '%s' mode without cleanup\nPlease run 'gbox cluster cleanup' first",
			currentMode, mode)
	}

	// Save mode to config file
	if err := saveMode(configFile, mode); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	// Execute setup for the respective mode
	if mode == "docker" {
		return setupDocker()
	} else {
		return setupK8s()
	}
}

// setupDocker sets up Docker environment
func setupDocker() error {
	fmt.Println("Setting up docker environment...")

	// Check and create Docker socket symlink (if needed)
	if _, err := os.Stat("/var/run/docker.sock"); os.IsNotExist(err) {
		fmt.Println("Docker socket symlink not found at /var/run/docker.sock")
		fmt.Println("This symlink is required for Docker Desktop for Mac to work properly")
		fmt.Println("We need sudo permission to create symlink at /var/run/docker.sock")
		fmt.Println("This is a one-time operation and will be remembered")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to get user home directory: %v", err)
		}

		dockerSock := filepath.Join(homeDir, ".docker/run/docker.sock")

		// Execute sudo ln -sf command
		cmd := exec.Command("sudo", "ln", "-sf", dockerSock, "/var/run/docker.sock")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create Docker socket symlink: %v", err)
		}
	}

	// Start docker-compose services
	fmt.Println("Starting docker-compose services...")
	scriptDir, err := getScriptDir()
	if err != nil {
		return err
	}

	composePath := filepath.Join(scriptDir, "..", "manifests", "docker", "docker-compose.yml")
	cmd := exec.Command("docker", "compose", "-f", composePath, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start docker-compose services: %v", err)
	}

	fmt.Println("Docker setup completed successfully")
	return nil
}

// setupK8s sets up Kubernetes environment
func setupK8s() error {
	fmt.Println("Setting up k8s environment...")

	// Get GBOX_HOME path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to get user home directory: %v", err)
	}
	gboxHome := filepath.Join(homeDir, ".gbox")

	// Create GBOX_HOME directory
	if err := os.MkdirAll(gboxHome, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Get script directory
	scriptDir, err := getScriptDir()
	if err != nil {
		return err
	}

	// Define constants
	manifestDir := filepath.Join(scriptDir, "..", "manifests")
	gboxCluster := "gbox"
	gboxKubecfg := filepath.Join(gboxHome, "kubeconfig")

	// Check if cluster already exists
	cmd := exec.Command("kind", "get", "clusters")
	output, err := cmd.Output()
	clusterExists := false
	if err == nil {
		clusterExists = contains(string(output), gboxCluster)
	}

	if !clusterExists {
		fmt.Println("Creating new cluster...")

		// Execute ytt command to generate cluster config
		yttCmd := exec.Command("ytt", "-f", filepath.Join(manifestDir, "k8s/cluster.yml"),
			"--data-value-yaml", "apiServerPort=41080",
			"--data-value", "home="+homeDir)

		// Create pipe
		r, w := io.Pipe()
		yttCmd.Stdout = w
		yttCmd.Stderr = os.Stderr

		// Create cluster with kind
		kindCmd := exec.Command("kind", "create", "cluster",
			"--name", gboxCluster,
			"--kubeconfig", gboxKubecfg,
			"--config", "-")
		kindCmd.Stdin = r
		kindCmd.Stdout = os.Stdout
		kindCmd.Stderr = os.Stderr

		// Start child process
		if err := yttCmd.Start(); err != nil {
			return fmt.Errorf("failed to start ytt command: %v", err)
		}

		// Start kind process
		if err := kindCmd.Start(); err != nil {
			yttCmd.Process.Kill()
			return fmt.Errorf("failed to start kind command: %v", err)
		}

		// Wait for process to complete
		go func() {
			yttCmd.Wait()
			w.Close()
		}()

		if err := kindCmd.Wait(); err != nil {
			return fmt.Errorf("failed to create cluster: %v", err)
		}
	} else {
		fmt.Printf("Cluster '%s' already exists, skipping creation...\n", gboxCluster)
	}

	// Deploy gbox application
	fmt.Println("Deploying gbox application...")

	// Generate application config with ytt
	yttCmd := exec.Command("ytt", "-f", filepath.Join(manifestDir, "k8s/app/"))

	// Create pipe
	r, w := io.Pipe()
	yttCmd.Stdout = w
	yttCmd.Stderr = os.Stderr

	// Deploy application with kapp
	kappCmd := exec.Command("kapp", "deploy", "-y",
		"--kubeconfig", gboxKubecfg,
		"--app", "gbox",
		"--file", "-")
	kappCmd.Stdin = r
	kappCmd.Stdout = os.Stdout
	kappCmd.Stderr = os.Stderr

	// Start child process
	if err := yttCmd.Start(); err != nil {
		return fmt.Errorf("failed to start ytt command: %v", err)
	}

	// Start kapp process
	if err := kappCmd.Start(); err != nil {
		yttCmd.Process.Kill()
		return fmt.Errorf("failed to start kapp command: %v", err)
	}

	// Wait for process to complete
	go func() {
		yttCmd.Wait()
		w.Close()
	}()

	if err := kappCmd.Wait(); err != nil {
		return fmt.Errorf("failed to deploy application: %v", err)
	}

	fmt.Println("K8s setup completed successfully")
	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
