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

	// 添加标志
	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}

// cleanupCluster 清理集群环境
func cleanupCluster(force bool) error {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("无法获取用户主目录: %v", err)
	}

	// 定义配置文件路径
	gboxHome := filepath.Join(homeDir, ".gbox")
	configFile := filepath.Join(gboxHome, "config.yml")

	// 如果配置文件不存在，则直接返回
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("集群已清理完毕。")
		return nil
	}

	// 获取当前模式
	mode, err := getCurrentMode(configFile)
	if err != nil {
		fmt.Printf("读取配置文件时出错: %v\n", err)
		// 错误不是致命的，继续执行
	}

	// 如果不是强制模式，则请求确认
	if !force {
		var confirmMsg string
		if mode != "" {
			confirmMsg = fmt.Sprintf("这将删除%s模式下的所有容器。继续？(y/N) ", mode)
		} else {
			confirmMsg = "这将删除所有容器。继续？(y/N) "
		}

		fmt.Print(confirmMsg)
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" && confirm != "yes" && confirm != "Yes" {
			fmt.Println("清理已取消")
			return nil
		}
	}

	// 根据模式执行清理
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
		// 尝试清理所有模式
		cleanupDocker()
		cleanupK8s()
	}

	// 清理完成后删除配置文件
	if err := os.Remove(configFile); err != nil {
		return fmt.Errorf("删除配置文件失败: %v", err)
	}

	return nil
}

// cleanupDocker 清理Docker环境
func cleanupDocker() error {
	fmt.Println("正在清理docker环境...")

	// 停止docker-compose服务
	fmt.Println("停止docker-compose服务...")
	scriptDir, err := getScriptDir()
	if err != nil {
		return err
	}

	// 执行docker-compose down命令
	composePath := filepath.Join(scriptDir, "..", "manifests", "docker", "docker-compose.yml")
	cmd := exec.Command("docker", "compose", "-f", composePath, "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("停止docker-compose服务失败: %v", err)
	}

	fmt.Println("Docker环境清理完成")
	return nil
}

// cleanupK8s 清理Kubernetes环境
func cleanupK8s() error {
	fmt.Println("正在清理k8s环境...")

	// 定义集群名称
	gboxCluster := "gbox"

	// 删除集群
	cmd := exec.Command("kind", "delete", "cluster", "--name", gboxCluster)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("删除集群失败: %v", err)
	}

	fmt.Println("K8s环境清理完成")
	return nil
}
