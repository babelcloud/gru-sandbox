package model_test

import (
	"testing"
	"time"

	"github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/stretchr/testify/assert"
)

func TestBox(t *testing.T) {
	now := time.Now()

	box := model.Box{
		ID:     "test-box-1",
		Status: "running",
		Image:  "ubuntu:latest",
		ExtraLabels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Config: model.LinuxAndroidBoxConfig{
			Browser: model.LinuxAndroidBoxConfigBrowser{
				Type:    "chrome",
				Version: "100.0",
			},
			CPU: 2.0,
			Envs: map[string]string{
				"ENV1": "value1",
			},
			Labels: map[string]string{
				"label1": "value1",
			},
			Memory: 4096,
			Os: model.LinuxAndroidBoxConfigOs{
				Version: "20.04",
			},
			Storage:    100,
			WorkingDir: "/workspace",
		},
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
		Type:      "linux",
		UpdatedAt: now,
	}

	assert.Equal(t, "test-box-1", box.ID)
	assert.Equal(t, "running", box.Status)
	assert.Equal(t, "ubuntu:latest", box.Image)
	assert.Equal(t, map[string]string{"key1": "value1", "key2": "value2"}, box.ExtraLabels)
	assert.Equal(t, "chrome", box.Config.Browser.Type)
	assert.Equal(t, "100.0", box.Config.Browser.Version)
	assert.Equal(t, 2.0, box.Config.CPU)
	assert.Equal(t, map[string]string{"ENV1": "value1"}, box.Config.Envs)
	assert.Equal(t, map[string]string{"label1": "value1"}, box.Config.Labels)
	assert.Equal(t, 4096.0, box.Config.Memory)
	assert.Equal(t, "20.04", box.Config.Os.Version)
	assert.Equal(t, 100.0, box.Config.Storage)
	assert.Equal(t, "/workspace", box.Config.WorkingDir)
	assert.Equal(t, now, box.CreatedAt)
	assert.Equal(t, now.Add(24*time.Hour), box.ExpiresAt)
	assert.Equal(t, "linux", box.Type)
	assert.Equal(t, now, box.UpdatedAt)
}

func TestBoxFile(t *testing.T) {
	now := time.Now()

	file := model.BoxFile{
		LastModified: now,
		Name:         "test.txt",
		Path:         "/workspace/test.txt",
		Size:         "1024",
		Type:         "file",
	}

	assert.Equal(t, now, file.LastModified)
	assert.Equal(t, "test.txt", file.Name)
	assert.Equal(t, "/workspace/test.txt", file.Path)
	assert.Equal(t, "1024", file.Size)
	assert.Equal(t, "file", file.Type)
}

func TestBoxFileListParams(t *testing.T) {
	params := model.BoxFileListParams{
		Path:  "/workspace",
		Depth: 2,
	}

	assert.Equal(t, "/workspace", params.Path)
	assert.Equal(t, 2.0, params.Depth)
}

func TestBoxFileListResult(t *testing.T) {
	now := time.Now()

	result := model.BoxFileListResult{
		Data: []model.BoxFile{
			{
				LastModified: now,
				Name:         "test1.txt",
				Path:         "/workspace/test1.txt",
				Size:         "1024",
				Type:         "file",
			},
			{
				LastModified: now,
				Name:         "test2.txt",
				Path:         "/workspace/test2.txt",
				Size:         "2048",
				Type:         "file",
			},
		},
	}

	assert.Len(t, result.Data, 2)
	assert.Equal(t, "test1.txt", result.Data[0].Name)
	assert.Equal(t, "test2.txt", result.Data[1].Name)
}

func TestBoxFileReadParams(t *testing.T) {
	params := model.BoxFileReadParams{
		Path: "/workspace/test.txt",
	}

	assert.Equal(t, "/workspace/test.txt", params.Path)
}

func TestBoxFileReadResult(t *testing.T) {
	result := model.BoxFileReadResult{
		Content: "file content",
	}

	assert.Equal(t, "file content", result.Content)
}

func TestBoxFileWriteParams(t *testing.T) {
	params := model.BoxFileWriteParams{
		Path:    "/workspace/test.txt",
		Content: "new content",
	}

	assert.Equal(t, "/workspace/test.txt", params.Path)
	assert.Equal(t, "new content", params.Content)
}

func TestBoxFileWriteResult(t *testing.T) {
	result := model.BoxFileWriteResult{
		Message: "success",
	}

	assert.Equal(t, "success", result.Message)
}
