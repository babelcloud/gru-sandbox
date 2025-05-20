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
		wantCmd []string
		wantArgs []string
		wantStdin bool
		wantStdout bool
		wantStderr bool
		wantTTY bool
	}{
		{
			name: "basic command",
			params: model.BoxExecParams{
				Cmd: []string{"ls"},
				Args: []string{"-l"},
				Stdin: true,
				Stdout: true,
				Stderr: true,
				TTY: false,
			},
			wantCmd: []string{"ls"},
			wantArgs: []string{"-l"},
			wantStdin: true,
			wantStdout: true,
			wantStderr: true,
			wantTTY: false,
		},
		{
			name: "empty args",
			params: model.BoxExecParams{
				Cmd: []string{"pwd"},
				Stdin: false,
				Stdout: true,
				Stderr: false,
				TTY: true,
			},
			wantCmd: []string{"pwd"},
			wantArgs: nil,
			wantStdin: false,
			wantStdout: true,
			wantStderr: false,
			wantTTY: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCmd, tt.params.Cmd)
			assert.Equal(t, tt.wantArgs, tt.params.Args)
			assert.Equal(t, tt.wantStdin, tt.params.Stdin)
			assert.Equal(t, tt.wantStdout, tt.params.Stdout)
			assert.Equal(t, tt.wantStderr, tt.params.Stderr)
			assert.Equal(t, tt.wantTTY, tt.params.TTY)
		})
	}
}

func TestBoxRunParams(t *testing.T) {
	tests := []struct {
		name string
		params model.BoxRunParams
		wantCmd []string
		wantArgs []string
		wantStdin string
		wantStdoutLimit int
		wantStderrLimit int
	}{
		{
			name: "complete params",
			params: model.BoxRunParams{
				Cmd: []string{"echo"},
				Args: []string{"hello"},
				Stdin: "input",
				StdoutLineLimit: 100,
				StderrLineLimit: 50,
			},
			wantCmd: []string{"echo"},
			wantArgs: []string{"hello"},
			wantStdin: "input",
			wantStdoutLimit: 100,
			wantStderrLimit: 50,
		},
		{
			name: "minimal params",
			params: model.BoxRunParams{
				Cmd: []string{"date"},
			},
			wantCmd: []string{"date"},
			wantArgs: nil,
			wantStdin: "",
			wantStdoutLimit: 0,
			wantStderrLimit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCmd, tt.params.Cmd)
			assert.Equal(t, tt.wantArgs, tt.params.Args)
			assert.Equal(t, tt.wantStdin, tt.params.Stdin)
			assert.Equal(t, tt.wantStdoutLimit, tt.params.StdoutLineLimit)
			assert.Equal(t, tt.wantStderrLimit, tt.params.StderrLineLimit)
		})
	}
}

func TestBoxExecWSParams(t *testing.T) {
	tests := []struct {
		name string
		params model.BoxExecWSParams
		wantCmd []string
		wantArgs []string
		wantTTY bool
		wantWorkingDir string
	}{
		{
			name: "full params",
			params: model.BoxExecWSParams{
				Cmd: []string{"bash"},
				Args: []string{"-c", "echo hello"},
				TTY: true,
				WorkingDir: "/app",
			},
			wantCmd: []string{"bash"},
			wantArgs: []string{"-c", "echo hello"},
			wantTTY: true,
			wantWorkingDir: "/app",
		},
		{
			name: "minimal params",
			params: model.BoxExecWSParams{
				Cmd: []string{"ls"},
			},
			wantCmd: []string{"ls"},
			wantArgs: nil,
			wantTTY: false,
			wantWorkingDir: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCmd, tt.params.Cmd)
			assert.Equal(t, tt.wantArgs, tt.params.Args)
			assert.Equal(t, tt.wantTTY, tt.params.TTY)
			assert.Equal(t, tt.wantWorkingDir, tt.params.WorkingDir)
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
	header := model.StreamHeader{
		Type: model.StreamStdout,
		Size: 1024,
	}

	assert.Equal(t, model.StreamStdout, header.Type)
	assert.Equal(t, uint32(1024), header.Size)
}
