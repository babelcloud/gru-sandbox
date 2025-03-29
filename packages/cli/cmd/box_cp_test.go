package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试解析盒子路径的功能
func TestParseBoxPath(t *testing.T) {
	// 测试有效路径
	validPath := "box-id:/some/path"
	result, err := parseBoxPath(validPath)
	assert.NoError(t, err)
	assert.Equal(t, "box-id", result.BoxID)
	assert.Equal(t, "/some/path", result.Path)

	// 测试无效路径
	invalidPath := "invalid-path-without-colon"
	_, err = parseBoxPath(invalidPath)
	assert.Error(t, err)
}

// 测试判断是否为盒子路径的功能
func TestIsBoxPath(t *testing.T) {
	assert.True(t, isBoxPath("box-id:/path"))
	assert.False(t, isBoxPath("/local/path"))
}

// 测试从盒子复制到本地
func TestCopyFromBoxToLocal(t *testing.T) {
	// 跳过此测试，因为它涉及执行tar命令和文件系统操作
	t.Skip("skip test that requires filesystem operations and tar commands")

	// 保存原始标准输出和错误以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建临时目录作为目标路径
	tempDir, err := os.MkdirTemp("", "box-cp-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 定义目标文件路径
	destFile := filepath.Join(tempDir, "test-file")

	// 创建模拟HTTP服务器
	mockContent := []byte("mock file content")
	mockArchive := createMockTarArchive(t, "test-file", mockContent)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/boxes/box-id/archive", r.URL.Path)
		assert.Equal(t, "path=/test/file", r.URL.RawQuery)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/x-tar")
		w.WriteHeader(http.StatusOK)
		w.Write(mockArchive)
	}))
	defer server.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	defer os.Setenv("API_URL", origAPIURL)

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", server.URL)

	// 捕获标准输出和错误
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	// 执行命令
	cmd := NewBoxCpCommand()
	cmd.SetArgs([]string{
		"box-id:/test/file",
		destFile,
	})

	// 启动一个goroutine来执行命令，避免被os.Exit中断测试
	done := make(chan bool)
	go func() {
		defer close(done)
		err = cmd.Execute()
		stdoutW.Close()
		stderrW.Close()
	}()

	// 等待命令完成
	<-done

	// 读取标准输出和错误
	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, stdoutR)
	io.Copy(&stderrBuf, stderrR)

	// 验证目标文件存在
	_, err = os.Stat(destFile)
	assert.NoError(t, err, "目标文件应该存在")

	// 验证文件内容
	content, err := os.ReadFile(destFile)
	assert.NoError(t, err)
	assert.Equal(t, mockContent, content, "文件内容应该正确")

	// 验证标准错误输出
	assert.Contains(t, stderrBuf.String(), "已复制从盒子")
}

// 测试从本地复制到盒子
func TestCopyFromLocalToBox(t *testing.T) {
	// 跳过此测试，因为它涉及执行tar命令和文件系统操作
	t.Skip("skip test that requires filesystem operations and tar commands")

	// 保存原始标准输出和错误以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建临时源文件
	tempFile, err := os.CreateTemp("", "box-cp-source")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// 写入测试内容到源文件
	testContent := []byte("test content for upload")
	_, err = tempFile.Write(testContent)
	assert.NoError(t, err)
	tempFile.Close()

	// 验证上传到服务器的内容
	var uploadedContent []byte

	// 创建模拟HTTP服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/boxes/box-id/archive", r.URL.Path)
		assert.Equal(t, "path=/dest/path", r.URL.RawQuery)
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "application/x-tar", r.Header.Get("Content-Type"))

		// 读取上传内容
		uploadedContent, err = io.ReadAll(r.Body)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	defer os.Setenv("API_URL", origAPIURL)

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", server.URL)

	// 捕获标准输出和错误
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	// 执行命令
	cmd := NewBoxCpCommand()
	cmd.SetArgs([]string{
		tempFile.Name(),
		"box-id:/dest/path",
	})

	// 启动一个goroutine来执行命令，避免被os.Exit中断测试
	done := make(chan bool)
	go func() {
		defer close(done)
		err = cmd.Execute()
		stdoutW.Close()
		stderrW.Close()
	}()

	// 等待命令完成
	<-done

	// 读取标准输出和错误
	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, stdoutR)
	io.Copy(&stderrBuf, stderrR)

	// 验证上传的内容是有效的tar文件
	assert.True(t, len(uploadedContent) > 0, "应该上传非空内容")

	// 验证标准错误输出
	assert.Contains(t, stderrBuf.String(), "已复制从")
	assert.Contains(t, stderrBuf.String(), "到盒子")
}

// 测试从盒子复制到标准输出
func TestCopyFromBoxToStdout(t *testing.T) {
	// 跳过此测试，因为它涉及os.Exit调用
	t.Skip("skip test that calls os.Exit")

	// 保存原始标准输出和错误以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 创建模拟HTTP服务器
	mockContent := []byte("mock file content for stdout")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/boxes/box-id/archive", r.URL.Path)
		assert.Equal(t, "path=/test/file-stdout", r.URL.RawQuery)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/x-tar")
		w.WriteHeader(http.StatusOK)
		w.Write(mockContent)
	}))
	defer server.Close()

	// 保存原始环境变量
	origAPIURL := os.Getenv("API_URL")
	defer os.Setenv("API_URL", origAPIURL)

	// 设置 API 地址为 mock 服务器
	os.Setenv("API_URL", server.URL)

	// 捕获标准输出和错误
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	// 执行命令
	cmd := NewBoxCpCommand()
	cmd.SetArgs([]string{
		"box-id:/test/file-stdout",
		"-",
	})

	// 启动一个goroutine来执行命令，避免被os.Exit中断测试
	done := make(chan bool)
	go func() {
		defer close(done)
		err := cmd.Execute()
		assert.NoError(t, err)
		stdoutW.Close()
		stderrW.Close()
	}()

	// 等待命令完成
	<-done

	// 读取标准输出和错误
	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, stdoutR)
	io.Copy(&stderrBuf, stderrR)

	// 验证标准输出
	assert.Equal(t, mockContent, stdoutBuf.Bytes(), "应该将内容写入标准输出")
}

// 测试从标准输入复制到盒子
func TestCopyFromStdinToBox(t *testing.T) {
	// 跳过：这个测试需要模拟标准输入，比较复杂
	t.Skip("标准输入测试暂时跳过，需要复杂的标准输入模拟")
}

// 测试帮助信息
func TestBoxCpHelp(t *testing.T) {
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
	cmd := NewBoxCpCommand()
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

	// 检查帮助信息中是否包含关键部分
	assert.Contains(t, output, "用法: gbox box cp <src> <dst>")
	assert.Contains(t, output, "本地文件/目录路径")
	assert.Contains(t, output, "盒子路径，格式为 BOX_ID:SRC_PATH")
	assert.Contains(t, output, "\"-\" 表示从标准输入读取")
	assert.Contains(t, output, "\"-\" 表示写入标准输出")
	assert.Contains(t, output, "复制本地文件到盒子")
	assert.Contains(t, output, "从盒子复制到本地")
	assert.Contains(t, output, "从标准输入复制tar流到盒子")
	assert.Contains(t, output, "从盒子复制到标准输出")
	assert.Contains(t, output, "复制目录从本地到盒子")
	assert.Contains(t, output, "复制目录从盒子到本地")
}

// 创建测试用的tar归档文件
func createMockTarArchive(t *testing.T, filename string, content []byte) []byte {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "tar-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFilePath := filepath.Join(tempDir, filename)
	err = os.WriteFile(testFilePath, content, 0644)
	assert.NoError(t, err)

	// 创建临时tar文件
	tarFile, err := os.CreateTemp("", "test-*.tar")
	assert.NoError(t, err)
	defer os.Remove(tarFile.Name())
	tarFile.Close()

	// 创建tar归档
	cmd := exec.Command("tar", "-cf", tarFile.Name(), "-C", tempDir, filename)
	err = cmd.Run()
	assert.NoError(t, err)

	// 读取tar内容
	tarContent, err := os.ReadFile(tarFile.Name())
	assert.NoError(t, err)

	return tarContent
}

// 测试无效参数
func TestBoxCpInvalidArgs(t *testing.T) {
	// 跳过此测试，因为os.Exit会中止测试进程
	t.Skip("skip test that calls os.Exit")

	// 保存原始标准输出和错误以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 捕获标准输出和错误
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	// 执行命令，参数不足
	cmd := NewBoxCpCommand()
	cmd.SetArgs([]string{
		"only-one-arg",
	})

	// 由于命令会调用os.Exit，我们不能直接执行它
	// 这里我们只验证它会输出帮助信息
	_ = cmd.Execute()

	// 验证结果
	stdoutW.Close()
	stderrW.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, stdoutR)
	io.Copy(&stderrBuf, stderrR)

	output := stdoutBuf.String()
	assert.Contains(t, output, "用法: gbox box cp <src> <dst>", "应该显示帮助信息")
}

// 测试无效的路径组合
func TestBoxCpInvalidPathCombination(t *testing.T) {
	// 跳过此测试，因为os.Exit会中止测试进程
	t.Skip("skip test that calls os.Exit")

	// 保存原始标准输出和错误以便后续恢复
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// 捕获标准输出和错误
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	// 执行命令，两个都是盒子路径
	cmd := NewBoxCpCommand()
	cmd.SetArgs([]string{
		"box1:/path1",
		"box2:/path2",
	})

	// 启动一个goroutine来执行命令，避免被os.Exit中断测试
	done := make(chan bool)
	go func() {
		defer close(done)
		_ = cmd.Execute()
		stdoutW.Close()
		stderrW.Close()
	}()

	// 等待命令完成
	<-done

	// 读取标准输出和错误
	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, stdoutR)
	io.Copy(&stderrBuf, stderrR)

	combined := stdoutBuf.String() + stderrBuf.String()
	assert.True(t, strings.Contains(combined, "错误") || strings.Contains(combined, "无效的路径格式"),
		"应该显示错误信息")
}
