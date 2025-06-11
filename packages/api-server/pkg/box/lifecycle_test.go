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
		expected model.ProgressStatus
	}{
		{
			name: "prepare status",
			update: model.ProgressUpdate{
				Status:  model.ProgressStatusPrepare,
				Message: "Preparing...",
			},
			expected: model.ProgressStatusPrepare,
		},
		{
			name: "complete status",
			update: model.ProgressUpdate{
				Status:   model.ProgressStatusComplete,
				Message: "Completed",
				ImageID: "test-image",
			},
			expected: model.ProgressStatusComplete,
		},
		{
			name: "error status",
			update: model.ProgressUpdate{
				Status: model.ProgressStatusError,
				Error:  "Something went wrong",
			},
			expected: model.ProgressStatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.update.Status)
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
		Cmd:                        "test-cmd",
		Args:                       []string{"arg1", "arg2"},
		WorkingDir:                 "/app",
		WaitForReady:              true,
		WaitForReadyTimeoutSeconds: 30,
		CreateTimeoutSeconds:       60,
		Timeout:                    5 * time.Minute,
		ProgressWriter:             writer,
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
	}

	assert.NotNil(t, params)
	assert.Equal(t, "test-image:latest", params.Image)
	assert.Equal(t, "secret123", params.ImagePullSecret)
	assert.Equal(t, 2, len(params.Env))
	assert.Equal(t, "test-cmd", params.Cmd)
	assert.Equal(t, 2, len(params.Args))
	assert.Equal(t, "/app", params.WorkingDir)
	assert.True(t, params.WaitForReady)
	assert.Equal(t, 30, params.WaitForReadyTimeoutSeconds)
	assert.Equal(t, 60, params.CreateTimeoutSeconds)
	assert.Equal(t, 5*time.Minute, params.Timeout)
	assert.NotNil(t, params.ProgressWriter)
	assert.NotNil(t, params.LinuxAndroidBoxCreateParam)
	assert.Equal(t, "30s", params.LinuxAndroidBoxCreateParam.Timeout)
	assert.True(t, params.LinuxAndroidBoxCreateParam.Wait)
	assert.Equal(t, "1h", params.LinuxAndroidBoxCreateParam.Config.ExpiresIn)
}

func TestVolumeMount(t *testing.T) {
	mount := model.VolumeMount{
		Source:      "/host/path",
		Target:      "/container/path",
		ReadOnly:    true,
		Propagation: "rprivate",
	}

	assert.Equal(t, "/host/path", mount.Source)
	assert.Equal(t, "/container/path", mount.Target)
	assert.True(t, mount.ReadOnly)
	assert.Equal(t, "rprivate", mount.Propagation)
}

func TestBoxCreateResult(t *testing.T) {
	result := model.BoxCreateResult{
		Box: model.Box{
			ID:     "test-box",
			Status: "running",
		},
		Message: "Box created successfully",
	}

	assert.Equal(t, "test-box", result.Box.ID)
	assert.Equal(t, "running", result.Box.Status)
	assert.Equal(t, "Box created successfully", result.Message)
}

func TestBoxDeleteResult(t *testing.T) {
	result := model.BoxDeleteResult{
		Message: "Box deleted successfully",
	}
	assert.Equal(t, "Box deleted successfully", result.Message)
}

func TestBoxesDeleteResult(t *testing.T) {
	result := model.BoxesDeleteResult{
		Count:   5,
		Message: "5 boxes deleted",
		IDs:     []string{"box1", "box2", "box3", "box4", "box5"},
	}

	assert.Equal(t, 5, result.Count)
	assert.Equal(t, "5 boxes deleted", result.Message)
	assert.Equal(t, 5, len(result.IDs))
}

func TestBoxReclaimResult(t *testing.T) {
	result := model.BoxReclaimResult{
		StoppedCount: 3,
		DeletedCount: 2,
		StoppedIDs:   []string{"box1", "box2", "box3"},
		DeletedIDs:   []string{"box4", "box5"},
	}

	assert.Equal(t, 3, result.StoppedCount)
	assert.Equal(t, 2, result.DeletedCount)
	assert.Equal(t, 3, len(result.StoppedIDs))
	assert.Equal(t, 2, len(result.DeletedIDs))
}

func TestLegacyTypes(t *testing.T) {
	androidParam := model.AndroidBoxCreateParam{
		CreateAndroidBox: model.LinuxAndroidBoxCreateParam{
			Timeout: "60s",
			Wait:    true,
			Config: model.CreateBoxConfigParam{
				ExpiresIn: "2h",
				Envs:      map[string]string{"ANDROID_HOME": "/android"},
			},
		},
	}

	linuxParam := model.LinuxBoxCreateParam{
		CreateLinuxBox: model.LinuxAndroidBoxCreateParam{
			Timeout: "30s",
			Wait:    false,
			Config: model.CreateBoxConfigParam{
				ExpiresIn: "1h",
				Envs:      map[string]string{"PATH": "/usr/local/bin"},
			},
		},
	}

	assert.Equal(t, "60s", androidParam.CreateAndroidBox.Timeout)
	assert.True(t, androidParam.CreateAndroidBox.Wait)
	assert.Equal(t, "2h", androidParam.CreateAndroidBox.Config.ExpiresIn)

	assert.Equal(t, "30s", linuxParam.CreateLinuxBox.Timeout)
	assert.False(t, linuxParam.CreateLinuxBox.Wait)
	assert.Equal(t, "1h", linuxParam.CreateLinuxBox.Config.ExpiresIn)
}
