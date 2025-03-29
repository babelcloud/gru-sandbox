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
const mockBoxReclaimSuccessResponse = `{"status":"success","message":"资源回收成功","stoppedCount":2,"deletedCount":1}`
const mockBoxReclaimEmptyResponse = `{"status":"success","message":"没有发现可回收的资源","stoppedCount":0,"deletedCount":0}`

// TestBoxReclaimSuccess 测试成功回收特定盒子资源
func TestBoxReclaimSuccess(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes/test-box-id/reclaim", r.URL.Path)

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxReclaimSuccessResponse))
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
	cmd := NewBoxReclaimCommand()
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
	assert.Contains(t, output, "资源回收成功")
	assert.Contains(t, output, "已停止 2 个盒子")
	assert.Contains(t, output, "已删除 1 个盒子")
}

// TestBoxReclaimWithJsonOutput 测试以JSON格式输出回收结果
func TestBoxReclaimWithJsonOutput(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes/test-box-id/reclaim", r.URL.Path)

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxReclaimSuccessResponse))
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
	cmd := NewBoxReclaimCommand()
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
	assert.JSONEq(t, mockBoxReclaimSuccessResponse, strings.TrimSpace(output))
}

// TestBoxReclaimWithForce 测试使用强制参数回收盒子资源
func TestBoxReclaimWithForce(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes/test-box-id/reclaim", r.URL.Path)

		// 检查强制参数
		assert.Equal(t, "force=true", r.URL.RawQuery)

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxReclaimSuccessResponse))
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
	cmd := NewBoxReclaimCommand()
	cmd.SetArgs([]string{"test-box-id", "--force"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "资源回收成功")
}

// TestBoxReclaimAll 测试回收所有盒子资源
func TestBoxReclaimAll(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查请求方法和路径 - 全局回收
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/boxes/reclaim", r.URL.Path)

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxReclaimSuccessResponse))
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

	// 执行命令 - 不指定盒子ID
	cmd := NewBoxReclaimCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "资源回收成功")
}

// TestBoxReclaimNoResourcesFound 测试没有找到可回收资源的情况
func TestBoxReclaimNoResourcesFound(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes/reclaim", r.URL.Path)

		// 返回模拟响应 - 没有资源被回收
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxReclaimEmptyResponse))
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
	cmd := NewBoxReclaimCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "没有发现可回收的资源")
	assert.NotContains(t, output, "已停止")
	assert.NotContains(t, output, "已删除")
}

// TestBoxReclaimHelp 测试帮助信息
func TestBoxReclaimHelp(t *testing.T) {
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
	cmd := NewBoxReclaimCommand()
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
	assert.Contains(t, output, "用法: gbox box reclaim <id> [选项]")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "-f, --force")
	assert.Contains(t, output, "回收盒子资源")
	assert.Contains(t, output, "强制回收盒子资源")
	assert.Contains(t, output, "回收所有符合条件的盒子资源")
}

// TestBoxReclaimNotFound 测试盒子不存在的情况
func TestBoxReclaimNotFound(t *testing.T) {
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
	cmd := NewBoxReclaimCommand()
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
