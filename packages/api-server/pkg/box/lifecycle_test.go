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
	update := model.ProgressUpdate{
		Status:  model.ProgressStatusPrepare,
		Message: "Preparing operation",
		Error:   "",
		ImageID: "test-image",
	}

	assert.Equal(t, model.ProgressStatusPrepare, update.Status)
	assert.Equal(t, "Preparing operation", update.Message)
	assert.Empty(t, update.Error)
	assert.Equal(t, "test-image", update.ImageID)
}

func TestBoxCreateParams(t *testing.T) {
	writer := bytes.NewBuffer(nil)
	timeout := 30 * time.Second

	params := model.BoxCreateParams{
		Image:           "test/image:latest",
		ImagePullSecret: "secret123",
		Env: map[string]string{
			"KEY": "value",
		},
		Cmd:        "test command",
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
		WaitForReadyTimeoutSeconds: 60,
		CreateTimeoutSeconds:       120,
		Timeout:                    timeout,
		ProgressWriter:            writer,
		LinuxAndroidBoxCreateParam: &model.LinuxAndroidBoxCreateParam{
			Type:    "linux",
			Timeout: "30s",
			Wait:    true,
			Config: model.CreateBoxConfigParam{
				ExpiresIn: "1000s",
				Envs: map[string]string{
					"ENV1": "val1",
				},
				Labels: map[string]string{
					"label1": "val1",
				},
			},
		},
	}

	assert.Equal(t, "test/image:latest", params.Image)
	assert.Equal(t, "secret123", params.ImagePullSecret)
	assert.Equal(t, "value", params.Env["KEY"])
	assert.Equal(t, "test command", params.Cmd)
	assert.Equal(t, []string{"arg1", "arg2"}, params.Args)
	assert.Equal(t, "/app", params.WorkingDir)
	assert.Equal(t, "value1", params.ExtraLabels["label1"])
	assert.Len(t, params.Volumes, 1)
	assert.Equal(t, timeout, params.Timeout)
	assert.Equal(t, writer, params.ProgressWriter)
	assert.Equal(t, "linux", params.LinuxAndroidBoxCreateParam.Type)
}

func TestOperationResult(t *testing.T) {
	now := time.Now()
	result := model.OperationResult{
		Success: true,
		Message: "Operation successful",
		ID:      "test-id",
		Config: model.CreateBoxConfigParam{
			ExpiresIn: "1000s",
			Envs:      map[string]string{"ENV": "value"},
			Labels:    map[string]string{"label": "value"},
		},
		CreatedAt: now,
		ExpiresAt: now.Add(1000 * time.Second),
		Status:    "running",
		Type:      "linux",
		UpdatedAt: now,
	}

	assert.True(t, result.Success)
	assert.Equal(t, "Operation successful", result.Message)
	assert.Equal(t, "test-id", result.ID)
	assert.Equal(t, "1000s", result.Config.ExpiresIn)
	assert.Equal(t, now, result.CreatedAt)
	assert.Equal(t, "running", result.Status)
	assert.Equal(t, "linux", result.Type)
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
