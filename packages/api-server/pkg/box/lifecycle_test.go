package model_test

import (
	"testing"
	"time"
	"bytes"

	"github.com/stretchr/testify/assert"
	"github.com/babelcloud/gbox/packages/api-server/pkg/box"
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
				ImageID: "image123",
			},
			expected: model.ProgressUpdate{
				Status:  model.ProgressStatusComplete,
				Message: "Operation completed",
				ImageID: "image123",
			},
		},
		{
			name: "error status",
			update: model.ProgressUpdate{
				Status:  model.ProgressStatusError,
				Message: "Operation failed",
				Error:   "error details",
			},
			expected: model.ProgressUpdate{
				Status:  model.ProgressStatusError,
				Message: "Operation failed",
				Error:   "error details",
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
				Image: "nginx:latest",
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
			name: "with volumes",
			params: model.BoxCreateParams{
				Image: "ubuntu:latest",
				Volumes: []model.VolumeMount{
					{
						Source:      "/host/path",
						Target:      "/container/path",
						ReadOnly:    true,
						Propagation: "rprivate",
					},
				},
			},
		},
		{
			name: "with linux box params",
			params: model.BoxCreateParams{
				LinuxAndroidBoxCreateParam: &model.LinuxAndroidBoxCreateParam{
					Timeout: "30s",
					Wait:    true,
					Config: model.CreateBoxConfigParam{
						ExpiresIn: "1h",
						Envs: map[string]string{
							"ENV1": "val1",
						},
						Labels: map[string]string{
							"label1": "value1",
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

func TestOperationResult(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name   string
		result model.OperationResult
	}{
		{
			name: "success result",
			result: model.OperationResult{
				Success: true,
				Message: "Operation successful",
				ID:      "box-123",
				Config: model.CreateBoxConfigParam{
					ExpiresIn: "1h",
					Envs:      map[string]string{"KEY": "value"},
					Labels:    map[string]string{"app": "test"},
				},
				CreatedAt: now,
				ExpiresAt: now.Add(time.Hour),
				Status:    "running",
				Type:      "linux",
				UpdatedAt: now,
			},
		},
		{
			name: "failure result",
			result: model.OperationResult{
				Success: false,
				Message: "Operation failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.result)
			if tt.result.Success {
				assert.NotEmpty(t, tt.result.ID)
			}
		})
	}
}

func TestBoxReclaimResult(t *testing.T) {
	tests := []struct {
		name   string
		result model.BoxReclaimResult
	}{
		{
			name: "with stopped and deleted boxes",
			result: model.BoxReclaimResult{
				StoppedCount: 2,
				DeletedCount: 1,
				StoppedIDs:   []string{"box-1", "box-2"},
				DeletedIDs:   []string{"box-3"},
			},
		},
		{
			name: "empty result",
			result: model.BoxReclaimResult{
				StoppedCount: 0,
				DeletedCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.result)
			assert.Equal(t, len(tt.result.StoppedIDs), tt.result.StoppedCount)
			assert.Equal(t, len(tt.result.DeletedIDs), tt.result.DeletedCount)
		})
	}
}
