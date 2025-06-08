package model_test

import (
	"io"
	"testing"

	"github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/stretchr/testify/assert"
)

type mockReadWriteCloser struct {
	io.Reader
	io.Writer
	io.Closer
}

func TestBoxExecParams(t *testing.T) {
	tests := []struct {
		name string
		params model.BoxExecParams
		want model.BoxExecParams
	}{
		{
			name: "basic params",
			params: model.BoxExecParams{
				Commands: []string{"ls", "-l"},
				Args:    []string{"-a"},
				Stdin:   true,
				Stdout:  true,
				Stderr:  true,
				TTY:     true,
			},
			want: model.BoxExecParams{
				Commands: []string{"ls", "-l"},
				Args:    []string{"-a"},
				Stdin:   true,
				Stdout:  true,
				Stderr:  true,
				TTY:     true,
			},
		},
		{
			name: "with timeout and working dir",
			params: model.BoxExecParams{
				Commands:   []string{"python", "script.py"},
				Timeout:   "30s",
				WorkingDir: "/app",
				Envs: map[string]string{
					"FOO": "bar",
				},
			},
			want: model.BoxExecParams{
				Commands:   []string{"python", "script.py"},
				Timeout:   "30s",
				WorkingDir: "/app",
				Envs: map[string]string{
					"FOO": "bar",
				},
			},
		},
		{
			name: "empty params",
			params: model.BoxExecParams{},
			want: model.BoxExecParams{},
		},
		{
			name: "with connection",
			params: model.BoxExecParams{
				Commands: []string{"echo", "test"},
				Conn: &mockReadWriteCloser{},
			},
			want: model.BoxExecParams{
				Commands: []string{"echo", "test"},
				Conn: &mockReadWriteCloser{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.params)
		})
	}
}

func TestBoxExecResult(t *testing.T) {
	tests := []struct {
		name string
		result model.BoxExecResult
		want model.BoxExecResult
	}{
		{
			name: "success result",
			result: model.BoxExecResult{
				ExitCode: 0,
				Stdout: "success output",
				Stderr: "",
			},
			want: model.BoxExecResult{
				ExitCode: 0,
				Stdout: "success output",
				Stderr: "",
			},
		},
		{
			name: "error result",
			result: model.BoxExecResult{
				ExitCode: 1,
				Stdout: "",
				Stderr: "error message",
			},
			want: model.BoxExecResult{
				ExitCode: 1,
				Stdout: "",
				Stderr: "error message",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.result)
		})
	}
}

func TestBoxRunParams(t *testing.T) {
	tests := []struct {
		name string
		params model.BoxRunParams
		want model.BoxRunParams
	}{
		{
			name: "basic run params",
			params: model.BoxRunParams{
				Cmd:  []string{"echo", "hello"},
				Argv: []string{"world"},
				Stdin: "test input",
				StdoutLineLimit: 100,
				StderrLineLimit: 50,
			},
			want: model.BoxRunParams{
				Cmd:  []string{"echo", "hello"},
				Argv: []string{"world"},
				Stdin: "test input",
				StdoutLineLimit: 100,
				StderrLineLimit: 50,
			},
		},
		{
			name: "cloud version params",
			params: model.BoxRunParams{
				Code:     "print('hello')",
				Language: "python3",
				Timeout:  "10s",
				WorkingDir: "/scripts",
				Envs: map[string]string{
					"PYTHONPATH": "/usr/local/lib/python3.9",
				},
			},
			want: model.BoxRunParams{
				Code:     "print('hello')",
				Language: "python3",
				Timeout:  "10s",
				WorkingDir: "/scripts",
				Envs: map[string]string{
					"PYTHONPATH": "/usr/local/lib/python3.9",
				},
			},
		},
		{
			name: "empty params",
			params: model.BoxRunParams{},
			want: model.BoxRunParams{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.params)
		})
	}
}

func TestBoxExecWSParams(t *testing.T) {
	tests := []struct {
		name string
		params model.BoxExecWSParams
		want model.BoxExecWSParams
	}{
		{
			name: "basic websocket params",
			params: model.BoxExecWSParams{
				Cmd: []string{"bash"},
				Args: []string{"-c", "echo hello"},
				TTY: true,
				WorkingDir: "/app",
			},
			want: model.BoxExecWSParams{
				Cmd: []string{"bash"},
				Args: []string{"-c", "echo hello"},
				TTY: true,
				WorkingDir: "/app",
			},
		},
		{
			name: "minimal params",
			params: model.BoxExecWSParams{
				Cmd: []string{"ls"},
			},
			want: model.BoxExecWSParams{
				Cmd: []string{"ls"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.params)
		})
	}
}

func TestStreamTypes(t *testing.T) {
	assert.Equal(t, model.StreamType(0), model.StreamStdin)
	assert.Equal(t, model.StreamType(1), model.StreamStdout)
	assert.Equal(t, model.StreamType(2), model.StreamStderr)
}

func TestMediaTypes(t *testing.T) {
	assert.Equal(t, "application/vnd.gbox.raw-stream", model.MediaTypeRawStream)
	assert.Equal(t, "application/vnd.gbox.multiplexed-stream", model.MediaTypeMultiplexedStream)
}

func TestStreamHeader(t *testing.T) {
	tests := []struct {
		name string
		header model.StreamHeader
		wantType model.StreamType
		wantSize uint32
	}{
		{
			name: "stdout header",
			header: model.StreamHeader{
				Type: model.StreamStdout,
				Size: 1024,
			},
			wantType: model.StreamStdout,
			wantSize: 1024,
		},
		{
			name: "stderr header",
			header: model.StreamHeader{
				Type: model.StreamStderr,
				Size: 512,
			},
			wantType: model.StreamStderr,
			wantSize: 512,
		},
		{
			name: "stdin header",
			header: model.StreamHeader{
				Type: model.StreamStdin,
				Size: 256,
			},
			wantType: model.StreamStdin,
			wantSize: 256,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.header.Type)
			assert.Equal(t, tt.wantSize, tt.header.Size)
		})
	}
}
