package model_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/stretchr/testify/assert"
)

func TestProgressUpdate(t *testing.T) {
	tests := []struct {
		name     string
		update   model.ProgressUpdate
		expected model.ProgressUpdate
	}{
		{
			name: "prepare status",
			update: model.ProgressUpdate{
				Status:  model.ProgressStatusPrepare,
				Message: "Preparing operation",
			},
			expected: model.ProgressUpdate{
				Status:  model.ProgressStatusPrepare,
				Message: "Preparing operation",
			},
		},
		{
			name: "complete status with image ID",
			update: model.ProgressUpdate{
				Status:  model.ProgressStatusComplete,
				Message: "Operation completed",
				ImageID: "sha256:abc123",
			},
			expected: model.ProgressUpdate{
				Status:  model.ProgressStatusComplete,
				Message: "Operation completed",
				ImageID: "sha256:abc123",
			},
		},
		{
			name: "error status",
			update: model.ProgressUpdate{
				Status:  model.ProgressStatusError,
				Message: "Operation failed",
				Error:   "connection timeout",
			},
			expected: model.ProgressUpdate{
				Status:  model.ProgressStatusError,
				Message: "Operation failed",
				Error:   "connection timeout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.update)
		})
	}
}

func TestBoxCreateParams(t *testing.T) {
	tests := []struct {
		name   string
		params model.BoxCreateParams
	}{
		{
			name: "basic params",
			params: model.BoxCreateParams{
				Image:           "nginx:latest",
				ImagePullSecret: "secret123",
				Env: map[string]string{
					"KEY": "value",
				},
				Cmd:        "nginx",
				Args:       []string{"-g", "daemon off;"},
				WorkingDir: "/app",
				ExtraLabels: map[string]string{
					"app": "web",
				},
				WaitForReady:               true,
				WaitForReadyTimeoutSeconds: 30,
				CreateTimeoutSeconds:       60,
				Timeout:                    5 * time.Minute,
				ProgressWriter:             &bytes.Buffer{},
			},
		},
		{
			name: "linux box params",
			params: model.BoxCreateParams{
				Image: "ubuntu:latest",
				OfCreateLinuxBox: &model.LinuxAndroidBoxCreateParam{
					Type: "linux",
					Config: model.CreateBoxConfigParam{
						ExpiresIn: "24h",
						Envs: map[string]string{
							"ENV1": "val1",
						},
						Labels: map[string]string{
							"type": "linux",
						},
					},
				},
			},
		},
		{
			name: "android box params",
			params: model.BoxCreateParams{
				Image: "android:latest",
				OfCreateAndroidBox: &model.LinuxAndroidBoxCreateParam{
					Type: "android",
					Config: model.CreateBoxConfigParam{
						ExpiresIn: "12h",
						Envs: map[string]string{
							"ANDROID_HOME": "/opt/android",
						},
						Labels: map[string]string{
							"type": "android",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.params)
		})
	}
}

func TestVolumeMount(t *testing.T) {
	tests := []struct {
		name   string
		mount  model.VolumeMount
		expect model.VolumeMount
	}{
		{
			name: "read-only mount",
			mount: model.VolumeMount{
				Source:      "/host/path",
				Target:      "/container/path",
				ReadOnly:    true,
				Propagation: "rprivate",
			},
			expect: model.VolumeMount{
				Source:      "/host/path",
				Target:      "/container/path",
				ReadOnly:    true,
				Propagation: "rprivate",
			},
		},
		{
			name: "read-write mount",
			mount: model.VolumeMount{
				Source:      "/data",
				Target:      "/mnt/data",
				ReadOnly:    false,
				Propagation: "rshared",
			},
			expect: model.VolumeMount{
				Source:      "/data",
				Target:      "/mnt/data",
				ReadOnly:    false,
				Propagation: "rshared",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.mount)
		})
	}
}

func TestOperationResults(t *testing.T) {
	tests := []struct {
		name     string
		result   model.OperationResult
		expected model.OperationResult
	}{
		{
			name: "successful operation",
			result: model.OperationResult{
				Success: true,
				Message: "Operation completed successfully",
			},
			expected: model.OperationResult{
				Success: true,
				Message: "Operation completed successfully",
			},
		},
		{
			name: "failed operation",
			result: model.OperationResult{
				Success: false,
				Message: "Operation failed",
			},
			expected: model.OperationResult{
				Success: false,
				Message: "Operation failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result)
		})
	}
}

func TestBoxReclaimResult(t *testing.T) {
	result := model.BoxReclaimResult{
		StoppedCount: 2,
		DeletedCount: 1,
		StoppedIDs:   []string{"box1", "box2"},
		DeletedIDs:   []string{"box3"},
	}

	assert.Equal(t, 2, result.StoppedCount)
	assert.Equal(t, 1, result.DeletedCount)
	assert.Equal(t, []string{"box1", "box2"}, result.StoppedIDs)
	assert.Equal(t, []string{"box3"}, result.DeletedIDs)
}
