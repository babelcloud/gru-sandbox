package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试数据
const mockBoxInspectResponse = `{
	"id": "test-box-id",
	"image": "ubuntu:latest",
	"status": "running",
	"created": "2023-05-01T12:00:00Z",
	"ports": [{"host": 8080, "container": 80}],
	"env": {"DEBUG": "true", "ENV": "test"}
}`

// TestBoxInspect 测试获取盒子详情
func TestBoxInspect(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查请求方法和路径
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/boxes/test-box-id", r.URL.Path)

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxInspectResponse))
	}))
	defer server.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	origTESTING := os.Getenv("TESTING")
	defer func() {
		os.Setenv("API_URL", origAPIURL)
		os.Setenv("TESTING", origTESTING)
	}()

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", server.URL)
	os.Setenv("TESTING", "true")

	// 创建管道以捕获标准输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 执行命令
	cmd := NewBoxInspectCommand()
	cmd.SetArgs([]string{"test-box-id"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "盒子详情:")
	assert.Contains(t, output, "id")
	assert.Contains(t, output, "test-box-id")
	assert.Contains(t, output, "image")
	assert.Contains(t, output, "ubuntu:latest")
	assert.Contains(t, output, "status")
	assert.Contains(t, output, "running")
}

// TestBoxInspectWithJsonOutput 测试以JSON格式获取盒子详情
func TestBoxInspectWithJsonOutput(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查请求方法和路径
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/boxes/test-box-id", r.URL.Path)

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxInspectResponse))
	}))
	defer server.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	origTESTING := os.Getenv("TESTING")
	defer func() {
		os.Setenv("API_URL", origAPIURL)
		os.Setenv("TESTING", origTESTING)
	}()

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", server.URL)
	os.Setenv("TESTING", "true")

	// 创建管道以捕获标准输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 执行命令
	cmd := NewBoxInspectCommand()
	cmd.SetArgs([]string{"test-box-id", "--output", "json"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出是否为原始JSON
	assert.JSONEq(t, mockBoxInspectResponse, strings.TrimSpace(output))
}

// TestBoxInspectNotFound 测试盒子不存在的情况
func TestBoxInspectNotFound(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回404错误
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Box not found"}`))
	}))
	defer server.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	origTESTING := os.Getenv("TESTING")
	defer func() {
		os.Setenv("API_URL", origAPIURL)
		os.Setenv("TESTING", origTESTING)
	}()

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", server.URL)
	os.Setenv("TESTING", "true")

	// 创建管道以捕获标准输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 执行命令
	cmd := NewBoxInspectCommand()
	cmd.SetArgs([]string{"non-existent-box-id"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// 检查输出
	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)
	assert.Contains(t, output, "盒子未找到")
}

// TestBoxInspectHelp 测试帮助信息
func TestBoxInspectHelp(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建管道以捕获标准输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 执行命令
	cmd := NewBoxInspectCommand()
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查帮助信息中是否包含关键部分
	assert.Contains(t, output, "用法: gbox box inspect <id> [选项]")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "json或text")
	assert.Contains(t, output, "获取盒子详情")
	assert.Contains(t, output, "获取JSON格式的盒子详情")
}
