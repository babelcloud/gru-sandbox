package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ClusterConfig 集群配置文件结构
type ClusterConfig struct {
	Cluster struct {
		Mode string `yaml:"mode"`
	} `yaml:"cluster"`
}

// getCurrentMode 从配置文件获取当前模式
func getCurrentMode(configFile string) (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return "", nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return "", err
	}

	// 解析YAML
	var config ClusterConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", err
	}

	return config.Cluster.Mode, nil
}

// saveMode 保存模式到配置文件
func saveMode(configFile string, mode string) error {
	// 确保目录存在
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 准备配置数据
	config := ClusterConfig{}
	config.Cluster.Mode = mode

	// 如果文件已存在，先读取现有内容
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return err
		}

		if err := yaml.Unmarshal(data, &config); err != nil {
			return err
		}

		// 更新模式
		config.Cluster.Mode = mode
	}

	// 保存到文件
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

// getScriptDir 获取脚本目录
func getScriptDir() (string, error) {
	// 获取二进制文件的路径
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("无法获取可执行文件路径: %v", err)
	}

	// 返回包含二进制文件的目录
	return filepath.Dir(exePath), nil
}
