package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetCurrentModeEmpty 测试当配置文件不存在时获取当前模式
func TestGetCurrentModeEmpty(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "gbox-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 设置不存在的配置文件路径
	configFile := filepath.Join(tempDir, "config.yml")

	// 获取当前模式
	mode, err := getCurrentMode(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "", mode, "对于不存在的配置文件，应该返回空模式")
}

// TestGetCurrentModeDocker 测试从配置文件读取docker模式
func TestGetCurrentModeDocker(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "gbox-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建测试配置文件
	configFile := filepath.Join(tempDir, "config.yml")
	err = os.WriteFile(configFile, []byte("cluster:\n  mode: docker"), 0644)
	assert.NoError(t, err)

	// 获取当前模式
	mode, err := getCurrentMode(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "docker", mode, "应该正确读取docker模式")
}

// TestGetCurrentModeK8s 测试从配置文件读取k8s模式
func TestGetCurrentModeK8s(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "gbox-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建测试配置文件
	configFile := filepath.Join(tempDir, "config.yml")
	err = os.WriteFile(configFile, []byte("cluster:\n  mode: k8s"), 0644)
	assert.NoError(t, err)

	// 获取当前模式
	mode, err := getCurrentMode(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "k8s", mode, "应该正确读取k8s模式")
}

// TestSaveMode 测试保存模式到配置文件
func TestSaveMode(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "gbox-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 设置配置文件路径
	configFile := filepath.Join(tempDir, "config.yml")

	// 保存docker模式
	err = saveMode(configFile, "docker")
	assert.NoError(t, err)

	// 验证保存结果
	mode, err := getCurrentMode(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "docker", mode, "应该正确保存和读取docker模式")

	// 更改为k8s模式
	err = saveMode(configFile, "k8s")
	assert.NoError(t, err)

	// 验证更改结果
	mode, err = getCurrentMode(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "k8s", mode, "应该正确更新和读取k8s模式")
}

// TestGetScriptDir 测试获取脚本目录
func TestGetScriptDir(t *testing.T) {
	dir, err := getScriptDir()
	assert.NoError(t, err)
	assert.NotEmpty(t, dir, "脚本目录不应为空")
}
