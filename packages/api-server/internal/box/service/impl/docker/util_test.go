package docker

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/babelcloud/gbox/packages/api-server/internal/box/service"
	model "github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestContainerName(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "normal id",
			id:       "test-123",
			expected: "gbox-test-123",
		},
		{
			name:     "empty id",
			id:       "",
			expected: "gbox-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containerName(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapContainerState(t *testing.T) {
	tests := []struct {
		name     string
		state    string
		expected string
	}{
		{"running", "running", "running"},
		{"created", "created", "created"},
		{"restarting", "restarting", "restarting"},
		{"removing", "removing", "removing"},
		{"paused", "paused", "paused"},
		{"exited", "exited", "stopped"},
		{"dead", "dead", "dead"},
		{"unknown", "unknown", "unknown"},
		{"empty", "", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapContainerState(tt.state)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainerToBox(t *testing.T) {
	now := time.Now()
	createdStr := now.Format(time.RFC3339)

	tests := []struct {
		name     string
		input    interface{}
		expected *model.Box
	}{
		{
			name: "ContainerJSON input",
			input: types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{
						Status: "running",
					},
					Created: createdStr,
				},
				Config: &container.Config{
					Image: "test:latest",
					Labels: map[string]string{
						"gbox.id":         "test-123",
						"gbox.extra.key1": "value1",
					},
					Env: []string{"KEY=VALUE"},
				},
			},
			expected: &model.Box{
				ID:     "test-123",
				Status: "running",
				Image:  "test:latest",
				Type:   "linux",
				ExtraLabels: map[string]string{
					"key1": "value1",
				},
				Config: model.LinuxAndroidBoxConfig{
					Resolution: model.LinuxAndroidBoxConfigResolution{
						Width:  1920,
						Height: 1080,
					},
				},
			},
		},
		{
			name: "Container input",
			input: types.Container{
				ID:      "container-id",
				Image:   "test:latest",
				State:   "running",
				Created: now.Unix(),
				Labels: map[string]string{
					"gbox.id":         "test-123",
					"gbox.extra.key1": "value1",
				},
			},
			expected: &model.Box{
				ID:     "test-123",
				Status: "running",
				Image:  "test:latest",
				Type:   "linux",
				ExtraLabels: map[string]string{
					"key1": "value1",
				},
				Config: model.LinuxAndroidBoxConfig{
					Resolution: model.LinuxAndroidBoxConfigResolution{
						Width:  1920,
						Height: 1080,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containerToBox(tt.input)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Status, result.Status)
			assert.Equal(t, tt.expected.Image, result.Image)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.ExtraLabels, result.ExtraLabels)
			assert.Equal(t, tt.expected.Config.Resolution, result.Config.Resolution)
		})
	}
}

func TestPrepareLabels(t *testing.T) {
	tests := []struct {
		name     string
		boxID    string
		params   *model.BoxCreateParams
		expected map[string]string
	}{
		{
			name:  "basic labels",
			boxID: "test-123",
			params: &model.BoxCreateParams{
				Cmd:        "test-cmd",
				Args:       []string{"arg1", "arg2"},
				WorkingDir: "/test",
				ExtraLabels: map[string]string{
					"custom": "value",
				},
			},
			expected: map[string]string{
				"gbox.id":         "test-123",
				"gbox.cmd":        "test-cmd",
				"gbox.args":       `["arg1","arg2"]`,
				"gbox.working-dir": "/test",
				"gbox.extra.custom": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrepareLabels(tt.boxID, tt.params)
			for k, v := range tt.expected {
				assert.Equal(t, v, result[k])
			}
		})
	}
}

func TestJoinArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "empty args",
			args:     []string{},
			expected: "",
		},
		{
			name:     "single arg",
			args:     []string{"test"},
			expected: `["test"]`,
		},
		{
			name:     "multiple args",
			args:     []string{"arg1", "arg2", "arg3"},
			expected: `["arg1","arg2","arg3"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinArgs(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		args     []string
		expected []string
	}{
		{
			name:     "empty command",
			cmd:      "",
			args:     nil,
			expected: []string{"sleep", "infinity"},
		},
		{
			name:     "command without args",
			cmd:      "test",
			args:     []string{},
			expected: []string{"/bin/sh", "-c", "test"},
		},
		{
			name:     "command with args",
			cmd:      "test",
			args:     []string{"arg1", "arg2"},
			expected: []string{"test", "arg1", "arg2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCommand(tt.cmd, tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected []string
	}{
		{
			name:     "nil env",
			env:      nil,
			expected: nil,
		},
		{
			name:     "empty env",
			env:      map[string]string{},
			expected: []string{},
		},
		{
			name: "multiple vars",
			env: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			expected: []string{"KEY1=value1", "KEY2=value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEnvVars(tt.env)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestGetImage(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		expected string
	}{
		{
			name:     "empty image",
			image:    "",
			expected: "babelcloud/gbox-playwright:latest",
		},
		{
			name:     "image without tag",
			image:    "nginx",
			expected: "nginx:latest",
		},
		{
			name:     "image with tag",
			image:    "nginx:1.19",
			expected: "nginx:1.19",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetImage(tt.image)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWaitForResponse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    string
	}{
		{
			name:        "valid response",
			input:       "{\"status\":\"Downloading\"}\n{\"status\":\"Complete\"}",
			expectError: false,
			expected:    "Downloading\nComplete\n",
		},
		{
			name:        "error response",
			input:       "{\"error\":\"Failed to download\"}",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err := WaitForResponse(reader)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(result))
			}
		})
	}
}

type mockWriter struct {
	strings.Builder
	flushed bool
}

func (m *mockWriter) Flush() {
	m.flushed = true
}

func TestProcessPullProgress(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name: "valid progress",
			input: `{"status":"Downloading","progressDetail":{},"id":"layer1"}
{"status":"Complete","progressDetail":{},"id":"layer1"}`,
			expectError: false,
		},
		{
			name:        "error progress",
			input:       `{"error":"Failed to pull image"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			writer := &mockWriter{}

			err := ProcessPullProgress(reader, writer)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, writer.String())
			}
		})
	}
}

func TestMapToEnv(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected []string
	}{
		{
			name:     "nil map",
			env:      nil,
			expected: nil,
		},
		{
			name:     "empty map",
			env:      map[string]string{},
			expected: []string{},
		},
		{
			name: "populated map",
			env: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			expected: []string{"KEY1=value1", "KEY2=value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapToEnv(tt.env)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestHandleContainerError(t *testing.T) {
	tests := []struct {
		name          string
		err          error
		id           string
		expectedType error
	}{
		{
			name:          "nil error",
			err:          nil,
			id:           "test-id",
			expectedType: nil,
		},
		{
			name:          "not found error",
			err:          fmt.Errorf("no such container"),
			id:           "test-id",
			expectedType: service.ErrBoxNotFound,
		},
		{
			name:          "generic error",
			err:          fmt.Errorf("some other error"),
			id:           "test-id",
			expectedType: fmt.Errorf("docker error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleContainerError(tt.err, tt.id)
			if tt.expectedType == nil {
				assert.Nil(t, result)
			} else {
				assert.Error(t, result)
				if tt.expectedType == service.ErrBoxNotFound {
					assert.True(t, strings.Contains(result.Error(), "not found"))
				} else {
					assert.True(t, strings.Contains(result.Error(), "docker error"))
				}
			}
		})
	}
}
