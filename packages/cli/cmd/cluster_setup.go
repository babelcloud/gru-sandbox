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

	// 添加标志
	cmd.Flags().String("mode", "docker", "Cluster mode (docker or k8s)")

	return cmd
}

// setupCluster 设置集群环境
func setupCluster(mode string) error {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("无法获取用户主目录: %v", err)
	}

	// 定义配置文件路径
	gboxHome := filepath.Join(homeDir, ".gbox")
	configFile := filepath.Join(gboxHome, "config.yml")

	// 获取当前模式（如果存在）
	currentMode, err := getCurrentMode(configFile)
	if err != nil {
		fmt.Printf("读取配置文件时出错: %v\n", err)
		// 错误不是致命的，继续执行
	}

	// 如果命令行没有指定模式且当前配置中有模式，使用当前配置中的模式
	if mode == "docker" && currentMode != "" {
		mode = currentMode
	}

	// 验证模式
	if mode != "docker" && mode != "k8s" {
		return fmt.Errorf("无效的模式: %s. 必须是 'docker' 或者 'k8s'", mode)
	}

	// 检查模式是否改变
	if currentMode != "" && currentMode != mode {
		return fmt.Errorf("错误: 不能在未清理的情况下从 '%s' 模式更改为 '%s' 模式\n请先运行 'gbox cluster cleanup'",
			currentMode, mode)
	}

	// 保存模式到配置文件
	if err := saveMode(configFile, mode); err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}

	// 为相应的模式执行设置
	if mode == "docker" {
		return setupDocker()
	} else {
		return setupK8s()
	}
}

// setupDocker 设置Docker环境
func setupDocker() error {
	fmt.Println("正在设置docker环境...")

	// 检查并创建Docker socket符号链接（如果需要）
	if _, err := os.Stat("/var/run/docker.sock"); os.IsNotExist(err) {
		fmt.Println("在/var/run/docker.sock未找到Docker socket符号链接")
		fmt.Println("此符号链接是Docker Desktop for Mac正常工作所必需的")
		fmt.Println("我们需要sudo权限在/var/run/docker.sock创建符号链接")
		fmt.Println("这是一次性操作，将被记住")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("无法获取用户主目录: %v", err)
		}

		dockerSock := filepath.Join(homeDir, ".docker/run/docker.sock")

		// 执行sudo ln -sf命令
		cmd := exec.Command("sudo", "ln", "-sf", dockerSock, "/var/run/docker.sock")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("创建Docker socket符号链接失败: %v", err)
		}
	}

	// 启动docker-compose服务
	fmt.Println("启动docker-compose服务...")
	scriptDir, err := getScriptDir()
	if err != nil {
		return err
	}

	composePath := filepath.Join(scriptDir, "..", "manifests", "docker", "docker-compose.yml")
	cmd := exec.Command("docker", "compose", "-f", composePath, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启动docker-compose服务失败: %v", err)
	}

	fmt.Println("Docker设置成功完成")
	return nil
}

// setupK8s 设置Kubernetes环境
func setupK8s() error {
	fmt.Println("正在设置k8s环境...")

	// 获取GBOX_HOME路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("无法获取用户主目录: %v", err)
	}
	gboxHome := filepath.Join(homeDir, ".gbox")

	// 创建GBOX_HOME目录
	if err := os.MkdirAll(gboxHome, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 获取脚本目录
	scriptDir, err := getScriptDir()
	if err != nil {
		return err
	}

	// 定义常量
	manifestDir := filepath.Join(scriptDir, "..", "manifests")
	gboxCluster := "gbox"
	gboxKubecfg := filepath.Join(gboxHome, "kubeconfig")

	// 检查集群是否已存在
	cmd := exec.Command("kind", "get", "clusters")
	output, err := cmd.Output()
	clusterExists := false
	if err == nil {
		clusterExists = contains(string(output), gboxCluster)
	}

	if !clusterExists {
		fmt.Println("创建新集群...")

		// 执行ytt命令生成集群配置
		yttCmd := exec.Command("ytt", "-f", filepath.Join(manifestDir, "k8s/cluster.yml"),
			"--data-value-yaml", "apiServerPort=41080",
			"--data-value", "home="+homeDir)

		// 创建管道
		r, w := io.Pipe()
		yttCmd.Stdout = w
		yttCmd.Stderr = os.Stderr

		// 使用kind创建集群
		kindCmd := exec.Command("kind", "create", "cluster",
			"--name", gboxCluster,
			"--kubeconfig", gboxKubecfg,
			"--config", "-")
		kindCmd.Stdin = r
		kindCmd.Stdout = os.Stdout
		kindCmd.Stderr = os.Stderr

		// 启动子进程
		if err := yttCmd.Start(); err != nil {
			return fmt.Errorf("启动ytt命令失败: %v", err)
		}

		// 启动kind进程
		if err := kindCmd.Start(); err != nil {
			yttCmd.Process.Kill()
			return fmt.Errorf("启动kind命令失败: %v", err)
		}

		// 等待进程完成
		go func() {
			yttCmd.Wait()
			w.Close()
		}()

		if err := kindCmd.Wait(); err != nil {
			return fmt.Errorf("创建集群失败: %v", err)
		}
	} else {
		fmt.Printf("集群 '%s' 已存在，跳过创建...\n", gboxCluster)
	}

	// 部署gbox应用
	fmt.Println("部署gbox应用...")

	// 使用ytt生成应用配置
	yttCmd := exec.Command("ytt", "-f", filepath.Join(manifestDir, "k8s/app/"))

	// 创建管道
	r, w := io.Pipe()
	yttCmd.Stdout = w
	yttCmd.Stderr = os.Stderr

	// 使用kapp部署应用
	kappCmd := exec.Command("kapp", "deploy", "-y",
		"--kubeconfig", gboxKubecfg,
		"--app", "gbox",
		"--file", "-")
	kappCmd.Stdin = r
	kappCmd.Stdout = os.Stdout
	kappCmd.Stderr = os.Stderr

	// 启动子进程
	if err := yttCmd.Start(); err != nil {
		return fmt.Errorf("启动ytt命令失败: %v", err)
	}

	// 启动kapp进程
	if err := kappCmd.Start(); err != nil {
		yttCmd.Process.Kill()
		return fmt.Errorf("启动kapp命令失败: %v", err)
	}

	// 等待进程完成
	go func() {
		yttCmd.Wait()
		w.Close()
	}()

	if err := kappCmd.Wait(); err != nil {
		return fmt.Errorf("部署应用失败: %v", err)
	}

	fmt.Println("K8s设置成功完成")
	return nil
}

// contains 检查字符串是否包含某个子字符串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
