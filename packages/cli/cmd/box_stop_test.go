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
const mockBoxStopSuccessResponse = `{"message":"Box stopped successfully"}`

// TestBoxStopSuccess 测试成功停止盒子
func TestBoxStopSuccess(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes/test-box-id/stop", r.URL.Path)

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxStopSuccessResponse))
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
	cmd := NewBoxStopCommand()
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
	assert.Contains(t, output, "盒子已成功停止")
}

// TestBoxStopWithJsonOutput 测试以JSON格式停止盒子
func TestBoxStopWithJsonOutput(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes/test-box-id/stop", r.URL.Path)

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxStopSuccessResponse))
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
	cmd := NewBoxStopCommand()
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
	expectedJSON := `{"status":"success","message":"盒子已成功停止"}`
	assert.JSONEq(t, expectedJSON, strings.TrimSpace(output))
}

// TestBoxStopNotFound 测试盒子不存在的情况
func TestBoxStopNotFound(t *testing.T) {
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
	cmd := NewBoxStopCommand()
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

// TestBoxStopServerError 测试服务器错误的情况
func TestBoxStopServerError(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回服务器错误
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
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
	cmd := NewBoxStopCommand()
	cmd.SetArgs([]string{"server-error-box-id"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "错误: 停止盒子失败")
}

// TestBoxStopHelp 测试帮助信息
func TestBoxStopHelp(t *testing.T) {
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
	cmd := NewBoxStopCommand()
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
	assert.Contains(t, output, "用法: gbox box stop <id> [选项]")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "json或text")
	assert.Contains(t, output, "停止一个盒子")
	assert.Contains(t, output, "停止盒子并输出JSON")
}
