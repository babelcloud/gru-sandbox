package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 修正测试服务器的响应格式
const mockResponse = `{"id": "mock-box-id", "status": "stopped", "image": "alpine:latest"}`

// TestNewBoxCreateCommand 用于测试 NewBoxCreateCommand 解析 CLI 参数并正确调用 API
func TestNewBoxCreateCommand(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建 mock 服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查请求方法
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)

		// 读取请求体
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		// 解析请求 JSON
		var req BoxCreateRequest
		err = json.Unmarshal(body, &req)
		assert.NoError(t, err)

		// 确保 image 字段被正确解析
		assert.Equal(t, "alpine:latest", req.Image)
		assert.Equal(t, "/bin/sh", req.Cmd)
		assert.Equal(t, []string{"-c", "echo Hello"}, req.Args)
		assert.Equal(t, map[string]string{"ENV_VAR": "value"}, req.Env)

		// 返回 mock 响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	defer os.Setenv("API_URL", origAPIURL)

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", mockServer.URL)

	// 创建管道以捕获标准输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 执行命令
	cmd := NewBoxCreateCommand()
	cmd.SetArgs([]string{
		"--image", "alpine:latest",
		"--env", "ENV_VAR=value",
		"--", "/bin/sh", "-c", "echo Hello",
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查 CLI 输出
	assert.Contains(t, output, "mock-box-id", "CLI 应该正确输出返回的 ID")
}

// TestNewBoxCreateCommandWithLabelsAndWorkDir 测试带有标签和工作目录的情况
func TestNewBoxCreateCommandWithLabelsAndWorkDir(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		var req BoxCreateRequest
		err = json.Unmarshal(body, &req)
		assert.NoError(t, err)

		// 验证标签和工作目录
		assert.Equal(t, "nginx:latest", req.Image)
		assert.Equal(t, "/app", req.WorkingDir)
		assert.Equal(t, map[string]string{"app": "web", "env": "test"}, req.Labels)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	defer os.Setenv("API_URL", origAPIURL)

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", mockServer.URL)

	// 创建管道以捕获标准输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 执行命令
	cmd := NewBoxCreateCommand()
	cmd.SetArgs([]string{
		"--image", "nginx:latest",
		"--work-dir", "/app",
		"--label", "app=web",
		"-l", "env=test",
		"--", "nginx", "-g", "daemon off;",
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查 CLI 输出
	assert.Contains(t, output, "mock-box-id", "CLI 应该正确输出返回的 ID")
}

// TestNewBoxCreateCommandWithJSONOutput 测试 JSON 输出格式
func TestNewBoxCreateCommandWithJSONOutput(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	defer os.Setenv("API_URL", origAPIURL)

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", mockServer.URL)

	// 创建管道以捕获标准输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 执行命令
	cmd := NewBoxCreateCommand()
	cmd.SetArgs([]string{
		"--output", "json",
		"--image", "alpine:latest",
		"--", "/bin/sh", "-c", "echo Hello",
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查 CLI 输出
	assert.True(t, strings.Contains(output, `"id": "mock-box-id"`), "JSON输出应包含mock-box-id")
}

// TestNewBoxCreateCommandWithError 测试API返回错误的情况
func TestNewBoxCreateCommandWithError(t *testing.T) {
	// 跳过本测试，因为它需要测试os.Exit行为
	// 在Go测试环境中，os.Exit会直接结束测试进程
	// 要正确测试这种情况，需要在子进程中运行
	t.Skip("需要在子进程中测试os.Exit行为")
}

// 注意: 如果需要测试os.Exit行为，可以使用以下方法:
// 1. 创建一个带有特殊参数的测试程序
// 2. 在子进程中运行该程序
// 3. 检查子进程的退出码
// 例如:
// func TestOsExitBehavior(t *testing.T) {
//     if os.Getenv("TEST_EXIT") == "1" {
//         // 这里放置会调用os.Exit的代码
//         return
//     }
//     cmd := exec.Command(os.Args[0], "-test.run=TestOsExitBehavior")
//     cmd.Env = append(os.Environ(), "TEST_EXIT=1")
//     err := cmd.Run()
//     if e, ok := err.(*exec.ExitError); ok && !e.Success() {
//         // 测试通过，程序如预期返回非零退出码
//         return
//     }
//     t.Fatalf("期望进程退出，实际没有")
// }

// TestNewBoxCreateHelp 测试帮助信息
func TestNewBoxCreateHelp(t *testing.T) {
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
	cmd := NewBoxCreateCommand()
	cmd.SetArgs([]string{
		"--help",
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查帮助信息中是否包含关键选项和用法说明
	assert.Contains(t, output, "用法:")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "--image")
	assert.Contains(t, output, "--env")
	assert.Contains(t, output, "--work-dir")
	assert.Contains(t, output, "--label")
}

// TestNewBoxCreateCommandWithMultipleOptions 测试多个环境变量和标签
func TestNewBoxCreateCommandWithMultipleOptions(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建 mock 服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查请求方法
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)

		// 读取请求体
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		// 解析请求 JSON
		var req BoxCreateRequest
		err = json.Unmarshal(body, &req)
		assert.NoError(t, err)

		// 验证多个选项
		assert.Equal(t, "python:3.9", req.Image)
		assert.Equal(t, "/app", req.WorkingDir)

		// 验证多个环境变量
		expectedEnv := map[string]string{
			"PATH":     "/usr/local/bin:/usr/bin:/bin",
			"DEBUG":    "true",
			"NODE_ENV": "production",
		}
		assert.Equal(t, expectedEnv, req.Env)

		// 验证多个标签
		expectedLabels := map[string]string{
			"project": "myapp",
			"env":     "prod",
			"version": "1.0",
		}
		assert.Equal(t, expectedLabels, req.Labels)

		// 返回 mock 响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	defer os.Setenv("API_URL", origAPIURL)

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", mockServer.URL)

	// 创建管道以捕获标准输出
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 执行命令
	cmd := NewBoxCreateCommand()
	cmd.SetArgs([]string{
		"--image", "python:3.9",
		"--work-dir", "/app",
		"--env", "PATH=/usr/local/bin:/usr/bin:/bin",
		"--env", "DEBUG=true",
		"--env", "NODE_ENV=production",
		"--label", "project=myapp",
		"--label", "env=prod",
		"--label", "version=1.0",
		"--", "python", "-m", "http.server", "8000",
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查 CLI 输出
	assert.Contains(t, output, "mock-box-id", "CLI 应该正确输出返回的 ID")
}
