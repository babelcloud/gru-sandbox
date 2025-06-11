package model_test

import (
	"testing"

	model "github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/stretchr/testify/assert"
)

func TestBoxExecParams(t *testing.T) {
	params := model.BoxExecParams{
		Commands:   []string{"ls", "-l"},
		Timeout:    "30s",
		WorkingDir: "/app",
		Envs: map[string]string{
			"FOO": "bar",
		},
	}

	assert.Equal(t, []string{"ls", "-l"}, params.Commands)
	assert.Equal(t, "30s", params.Timeout)
	assert.Equal(t, "/app", params.WorkingDir)
	assert.Equal(t, map[string]string{"FOO": "bar"}, params.Envs)
}

func TestBoxExecResult(t *testing.T) {
	result := model.BoxExecResult{
		ExitCode: 0,
		Stdout:   "output",
		Stderr:   "error",
	}

	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "output", result.Stdout)
	assert.Equal(t, "error", result.Stderr)
}

func TestBoxRunCodeParams(t *testing.T) {
	params := model.BoxRunCodeParams{
		Code:       "print('hello')",
		Language:   "python3",
		Argv:       []string{"arg1", "arg2"},
		Timeout:    "10s",
		WorkingDir: "/code",
		Envs: map[string]string{
			"PATH": "/usr/local/bin",
		},
	}

	assert.Equal(t, "print('hello')", params.Code)
	assert.Equal(t, "python3", params.Language)
	assert.Equal(t, []string{"arg1", "arg2"}, params.Argv)
	assert.Equal(t, "10s", params.Timeout)
	assert.Equal(t, "/code", params.WorkingDir)
	assert.Equal(t, map[string]string{"PATH": "/usr/local/bin"}, params.Envs)
}

func TestBoxRunCodeResult(t *testing.T) {
	result := model.BoxRunCodeResult{
		ExitCode: 1,
		Stdout:   "hello",
		Stderr:   "error occurred",
	}

	assert.Equal(t, 1, result.ExitCode)
	assert.Equal(t, "hello", result.Stdout)
	assert.Equal(t, "error occurred", result.Stderr)
}

func TestBoxExecWSParams(t *testing.T) {
	params := model.BoxExecWSParams{
		Cmd:        []string{"bash"},
		Args:       []string{"-c", "echo hello"},
		TTY:        true,
		WorkingDir: "/ws",
	}

	assert.Equal(t, []string{"bash"}, params.Cmd)
	assert.Equal(t, []string{"-c", "echo hello"}, params.Args)
	assert.True(t, params.TTY)
	assert.Equal(t, "/ws", params.WorkingDir)
}

func TestStreamConstants(t *testing.T) {
	assert.Equal(t, model.StreamType(0), model.StreamStdin)
	assert.Equal(t, model.StreamType(1), model.StreamStdout)
	assert.Equal(t, model.StreamType(2), model.StreamStderr)

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
