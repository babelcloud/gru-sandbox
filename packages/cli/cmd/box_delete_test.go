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

// 模拟响应
const mockDeleteResponse = `{"status":"success"}`
const mockListResponse = `{"boxes":[{"id":"box-1"},{"id":"box-2"}]}`
const mockEmptyListResponse = `{"boxes":[]}`

// TestDeleteSingleBox 测试删除单个盒子
func TestDeleteSingleBox(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建模拟服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查路径和方法
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/v1/boxes/test-box-id", r.URL.Path)

		// 检查请求内容
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		var req map[string]interface{}
		err = json.Unmarshal(body, &req)
		assert.NoError(t, err)
		assert.Equal(t, true, req["force"])

		// 返回成功响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockDeleteResponse))
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
	cmd := NewBoxDeleteCommand()
	cmd.SetArgs([]string{
		"test-box-id",
		"--force", // 强制删除，避免确认提示
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "盒子删除成功")
}

// TestDeleteAllBoxes 测试删除所有盒子
func TestDeleteAllBoxes(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建模拟服务器
	var requestCount int
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// 第一个请求应该是列出所有盒子
		if requestCount == 1 {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/v1/boxes", r.URL.Path)

			// 返回盒子列表
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockListResponse))
			return
		}

		// 后续请求应该是删除单个盒子
		assert.Equal(t, "DELETE", r.Method)
		assert.True(t, strings.HasPrefix(r.URL.Path, "/api/v1/boxes/box-"))

		// 检查请求内容
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		var req map[string]interface{}
		err = json.Unmarshal(body, &req)
		assert.NoError(t, err)
		assert.Equal(t, true, req["force"])

		// 返回成功响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockDeleteResponse))
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
	cmd := NewBoxDeleteCommand()
	cmd.SetArgs([]string{
		"--all",
		"--force", // 强制删除，避免确认提示
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "以下盒子将被删除")
	assert.Contains(t, output, "所有盒子删除成功")

	// 验证发送了正确数量的请求
	// 1个GET请求列出盒子 + 2个DELETE请求删除盒子 = 3个请求
	assert.Equal(t, 3, requestCount)
}

// TestDeleteAllBoxesEmpty 测试当没有盒子时删除所有盒子
func TestDeleteAllBoxesEmpty(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建模拟服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/boxes", r.URL.Path)

		// 返回空盒子列表
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockEmptyListResponse))
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
	cmd := NewBoxDeleteCommand()
	cmd.SetArgs([]string{
		"--all",
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "没有盒子需要删除")
}

// TestDeleteBoxWithJSONOutput 测试JSON输出格式
func TestDeleteBoxWithJSONOutput(t *testing.T) {
	// 保存原始标准输出以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建模拟服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/v1/boxes/test-box-id", r.URL.Path)

		// 返回成功响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockDeleteResponse))
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
	cmd := NewBoxDeleteCommand()
	cmd.SetArgs([]string{
		"test-box-id",
		"--force",
		"--output", "json",
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, `{"status":"success"`)
	assert.Contains(t, output, `"message":"盒子删除成功"`)
}

// TestDeleteInvalidInput 测试无效的输入
func TestDeleteInvalidInput(t *testing.T) {
	// 跳过本测试，因为它需要测试os.Exit行为
	// 在Go测试环境中，os.Exit会直接结束测试进程
	t.Skip("需要在子进程中测试os.Exit行为")

	// 注意: 如需测试错误情况，可以按照box_create_test.go中的注释方法，
	// 创建子进程来测试os.Exit行为
}

// TestDeleteAllBoxesWithConfirmation 测试带确认的删除所有盒子功能
func TestDeleteAllBoxesWithConfirmation(t *testing.T) {
	// 保存原始标准输出和输入以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldStdin := os.Stdin
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		os.Stdin = oldStdin
	}()

	// 创建模拟服务器
	var requestCount int
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// 第一个请求应该是列出所有盒子
		if requestCount == 1 {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/v1/boxes", r.URL.Path)

			// 返回盒子列表
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockListResponse))
			return
		}

		// 后续请求应该是删除单个盒子
		assert.Equal(t, "DELETE", r.Method)
		assert.True(t, strings.HasPrefix(r.URL.Path, "/api/v1/boxes/box-"))

		// 检查请求内容
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		var req map[string]interface{}
		err = json.Unmarshal(body, &req)
		assert.NoError(t, err)
		assert.Equal(t, true, req["force"])

		// 返回成功响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockDeleteResponse))
	}))
	defer mockServer.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	defer os.Setenv("API_URL", origAPIURL)

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", mockServer.URL)

	// 模拟用户输入"y"进行确认
	r, w, _ := os.Pipe()
	os.Stdin = r
	// 写入"y"到标准输入
	go func() {
		w.Write([]byte("y\n"))
		w.Close()
	}()

	// 捕获标准输出
	outR, outW, _ := os.Pipe()
	os.Stdout = outW
	os.Stderr = outW

	// 执行命令
	cmd := NewBoxDeleteCommand()
	cmd.SetArgs([]string{
		"--all", // 删除所有盒子，会提示确认
	})
	err := cmd.Execute()
	assert.NoError(t, err)

	// 读取捕获的输出
	outW.Close()
	var buf bytes.Buffer
	io.Copy(&buf, outR)
	output := buf.String()

	fmt.Fprintf(oldStdout, "捕获的输出: %s\n", output)

	// 检查输出
	assert.Contains(t, output, "以下盒子将被删除")
	assert.Contains(t, output, "您确定要删除所有盒子吗?")
	assert.Contains(t, output, "所有盒子删除成功")

	// 验证发送了正确数量的请求
	// 1个GET请求列出盒子 + 2个DELETE请求删除盒子 = 3个请求
	assert.Equal(t, 3, requestCount)
}

// TestDeleteHelp 测试帮助信息
func TestDeleteHelp(t *testing.T) {
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
	cmd := NewBoxDeleteCommand()
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

	// 检查帮助信息中是否包含关键选项和用法
	assert.Contains(t, output, "用法:")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "--all")
	assert.Contains(t, output, "--force")
}
