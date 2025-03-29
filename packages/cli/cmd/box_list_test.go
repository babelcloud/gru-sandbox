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
const mockBoxListResponse = `{"boxes":[
	{"id":"box-1","image":"ubuntu:latest","status":"running"},
	{"id":"box-2","image":"nginx:1.19","status":"stopped"}
]}`

const mockEmptyBoxListResponse = `{"boxes":[]}`

// TestBoxListAll 测试列出所有盒子
func TestBoxListAll(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)
		assert.Empty(t, r.URL.RawQuery, "应该没有查询参数")

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxListResponse))
	}))
	defer server.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	defer os.Setenv("API_URL", origAPIURL)

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", server.URL)

	// 创建管道以捕获标准输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 执行命令
	cmd := NewBoxListCommand()
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
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "IMAGE")
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "box-1")
	assert.Contains(t, output, "box-2")
	assert.Contains(t, output, "ubuntu:latest")
	assert.Contains(t, output, "nginx:1.19")
	assert.Contains(t, output, "running")
	assert.Contains(t, output, "stopped")
}

// TestBoxListWithJsonOutput 测试以JSON格式列出盒子
func TestBoxListWithJsonOutput(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxListResponse))
	}))
	defer server.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	defer os.Setenv("API_URL", origAPIURL)

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", server.URL)

	// 创建管道以捕获标准输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 执行命令
	cmd := NewBoxListCommand()
	cmd.SetArgs([]string{"--output", "json"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出是否为原始JSON
	assert.JSONEq(t, mockBoxListResponse, strings.TrimSpace(output))
}

// TestBoxListWithLabelFilter 测试使用标签过滤盒子
func TestBoxListWithLabelFilter(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)

		// 检查查询参数
		query := r.URL.Query()
		filters := query["filter"]
		assert.Len(t, filters, 1)
		assert.Equal(t, "label=project=myapp", filters[0])

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxListResponse))
	}))
	defer server.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	defer os.Setenv("API_URL", origAPIURL)

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", server.URL)

	// 创建管道以捕获标准输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 执行命令
	cmd := NewBoxListCommand()
	cmd.SetArgs([]string{"-f", "label=project=myapp"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "box-1")
	assert.Contains(t, output, "box-2")
}

// TestBoxListWithAncestorFilter 测试使用镜像祖先过滤盒子
func TestBoxListWithAncestorFilter(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)

		// 检查查询参数
		query := r.URL.Query()
		filters := query["filter"]
		assert.Len(t, filters, 1)
		assert.Equal(t, "ancestor=ubuntu:latest", filters[0])

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxListResponse))
	}))
	defer server.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	defer os.Setenv("API_URL", origAPIURL)

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", server.URL)

	// 创建管道以捕获标准输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 执行命令
	cmd := NewBoxListCommand()
	cmd.SetArgs([]string{"--filter", "ancestor=ubuntu:latest"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "box-1")
	assert.Contains(t, output, "box-2")
}

// TestBoxListMultipleFilters 测试使用多个过滤器
func TestBoxListMultipleFilters(t *testing.T) {
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
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)

		// 检查查询参数
		query := r.URL.Query()
		filters := query["filter"]
		assert.Len(t, filters, 2)
		assert.Contains(t, filters, "label=project=myapp")
		assert.Contains(t, filters, "id=box-1")

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockBoxListResponse))
	}))
	defer server.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	defer os.Setenv("API_URL", origAPIURL)

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", server.URL)

	// 创建管道以捕获标准输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 执行命令
	cmd := NewBoxListCommand()
	cmd.SetArgs([]string{"-f", "label=project=myapp", "-f", "id=box-1"})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "box-1")
	assert.Contains(t, output, "box-2")
}

// TestBoxListEmpty 测试没有盒子的情况
func TestBoxListEmpty(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回空盒子列表
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockEmptyBoxListResponse))
	}))
	defer server.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	defer os.Setenv("API_URL", origAPIURL)

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", server.URL)

	// 创建管道以捕获标准输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 执行命令
	cmd := NewBoxListCommand()
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
	assert.Contains(t, output, "未找到盒子")
}

// TestBoxListHelp 测试帮助信息
func TestBoxListHelp(t *testing.T) {
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
	cmd := NewBoxListCommand()
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
	assert.Contains(t, output, "用法: gbox box list [选项]")
	assert.Contains(t, output, "列出所有盒子")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "--filter")
	assert.Contains(t, output, "id=abc123")
	assert.Contains(t, output, "label=project=myapp")
	assert.Contains(t, output, "ancestor=ubuntu:latest")
}
