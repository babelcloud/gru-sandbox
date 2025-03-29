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
const mockBoxStartSuccessResponse = `{"message":"Box started successfully"}`
const mockBoxStartAlreadyRunningResponse = `{"error":"Box is already running"}`
const mockBoxStartInvalidRequestResponse = `{"error":"Invalid request"}`

// TestBoxStartSuccess 测试成功启动盒子
func TestBoxStartSuccess(t *testing.T) {
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
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/boxes/test-box-id/start", r.URL.Path)

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxStartSuccessResponse))
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
	cmd := NewBoxStartCommand()
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
	assert.Contains(t, output, "Box started successfully")
}

// TestBoxStartWithJsonOutput 测试以JSON格式启动盒子
func TestBoxStartWithJsonOutput(t *testing.T) {
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
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/boxes/test-box-id/start", r.URL.Path)

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxStartSuccessResponse))
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
	cmd := NewBoxStartCommand()
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
	assert.JSONEq(t, mockBoxStartSuccessResponse, strings.TrimSpace(output))
}

// TestBoxStartNotFound 测试盒子不存在的情况
func TestBoxStartNotFound(t *testing.T) {
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
	cmd := NewBoxStartCommand()
	cmd.SetArgs([]string{"non-existent-box-id"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "盒子未找到")
}

// TestBoxStartAlreadyRunning 测试盒子已经在运行的情况
func TestBoxStartAlreadyRunning(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回400错误，盒子已在运行
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(mockBoxStartAlreadyRunningResponse))
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
	cmd := NewBoxStartCommand()
	cmd.SetArgs([]string{"already-running-box-id"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "盒子已在运行")
}

// TestBoxStartInvalidRequest 测试无效请求的情况
func TestBoxStartInvalidRequest(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回400错误，但不是"盒子已在运行"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(mockBoxStartInvalidRequestResponse))
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
	cmd := NewBoxStartCommand()
	cmd.SetArgs([]string{"invalid-request-box-id"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "错误: 无效的请求")
}

// TestBoxStartHelp 测试帮助信息
func TestBoxStartHelp(t *testing.T) {
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
	cmd := NewBoxStartCommand()
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
	assert.Contains(t, output, "用法: gbox box start <id> [选项]")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "json或text")
	assert.Contains(t, output, "启动一个盒子")
	assert.Contains(t, output, "启动盒子并输出JSON")
}
