package model_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/stretchr/testify/assert"
)

func TestProgressStatus(t *testing.T) {
	assert.Equal(t, model.ProgressStatus("prepare"), model.ProgressStatusPrepare)
	assert.Equal(t, model.ProgressStatus("complete"), model.ProgressStatusComplete)
	assert.Equal(t, model.ProgressStatus("error"), model.ProgressStatusError)
}

func TestProgressUpdate(t *testing.T) {
	tests := []struct {
		name     string
		update   model.ProgressUpdate
		expected model.ProgressUpdate
	}{
		{
			name: "complete status",
			update: model.ProgressUpdate{
				Status:  model.ProgressStatusComplete,
				Message: "Operation completed",
				ImageID: "test-image",
			},
			expected: model.ProgressUpdate{
				Status:  model.ProgressStatusComplete,
				Message: "Operation completed",
				ImageID: "test-image",
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
	writer := &bytes.Buffer{}
	params := model.BoxCreateParams{
		Image:           "test-image:latest",
		ImagePullSecret: "secret123",
		Env: map[string]string{
			"KEY1": "value1",
			"KEY2": "value2",
		},
		Cmd:        "test-cmd",
		Args:       []string{"arg1", "arg2"},
		WorkingDir: "/app",
		ExtraLabels: map[string]string{
			"label1": "value1",
		},
		Volumes: []model.VolumeMount{
			{
				Source:      "/host/path",
				Target:      "/container/path",
				ReadOnly:    true,
				Propagation: "private",
			},
		},
		WaitForReady:               true,
		WaitForReadyTimeoutSeconds: 30,
		CreateTimeoutSeconds:       60,
		Timeout:                    5 * time.Minute,
		ProgressWriter:             writer,
	}

	assert.NotNil(t, params)
	assert.Equal(t, "test-image:latest", params.Image)
	assert.Equal(t, "secret123", params.ImagePullSecret)
	assert.Equal(t, 2, len(params.Env))
	assert.Equal(t, "test-cmd", params.Cmd)
	assert.Equal(t, 2, len(params.Args))
	assert.Equal(t, "/app", params.WorkingDir)
	assert.Equal(t, 1, len(params.ExtraLabels))
	assert.Equal(t, 1, len(params.Volumes))
	assert.True(t, params.WaitForReady)
	assert.Equal(t, 30, params.WaitForReadyTimeoutSeconds)
	assert.Equal(t, 60, params.CreateTimeoutSeconds)
	assert.Equal(t, 5*time.Minute, params.Timeout)
	assert.Equal(t, writer, params.ProgressWriter)
}

func TestVolumeMount(t *testing.T) {
	mount := model.VolumeMount{
		Source:      "/host/path",
		Target:      "/container/path",
		ReadOnly:    true,
		Propagation: "private",
	}

	assert.Equal(t, "/host/path", mount.Source)
	assert.Equal(t, "/container/path", mount.Target)
	assert.True(t, mount.ReadOnly)
	assert.Equal(t, "private", mount.Propagation)
}

func TestBoxCreateResult(t *testing.T) {
	result := model.BoxCreateResult{
		Box:     model.Box{},
		Message: "Box created successfully",
	}

	assert.NotNil(t, result)
	assert.Equal(t, "Box created successfully", result.Message)
}

func TestBoxDeleteParams(t *testing.T) {
	params := model.BoxDeleteParams{
		Force: true,
	}

	assert.True(t, params.Force)
}

func TestBoxDeleteResult(t *testing.T) {
	result := model.BoxDeleteResult{
		Message: "Box deleted successfully",
	}

	assert.Equal(t, "Box deleted successfully", result.Message)
}

func TestBoxesDeleteParams(t *testing.T) {
	params := model.BoxesDeleteParams{
		Force: true,
	}

	assert.True(t, params.Force)
}

func TestBoxesDeleteResult(t *testing.T) {
	result := model.BoxesDeleteResult{
		Count:   2,
		Message: "Boxes deleted successfully",
		IDs:     []string{"box1", "box2"},
	}

	assert.Equal(t, 2, result.Count)
	assert.Equal(t, "Boxes deleted successfully", result.Message)
	assert.Equal(t, []string{"box1", "box2"}, result.IDs)
}

func TestOperationResult(t *testing.T) {
	result := model.OperationResult{
		Success: true,
		Message: "Operation completed successfully",
	}

	assert.True(t, result.Success)
	assert.Equal(t, "Operation completed successfully", result.Message)
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
