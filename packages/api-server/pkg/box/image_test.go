package model_test

import (
	"testing"

	model "github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/stretchr/testify/assert"
)

func TestImageUpdateParams(t *testing.T) {
	params := model.ImageUpdateParams{
		ImageReference: "repo/image:tag",
		DryRun:        true,
		Force:         true,
	}

	assert.Equal(t, "repo/image:tag", params.ImageReference)
	assert.True(t, params.DryRun)
	assert.True(t, params.Force)
}

func TestImageUpdateResponse(t *testing.T) {
	images := []model.ImageInfo{
		{
			ImageID:    "sha256:123",
			Repository: "repo/image1",
			Tag:        "latest",
			Status:     model.ImageStatusUpToDate,
			Action:     "keep",
		},
	}

	response := model.ImageUpdateResponse{
		Images:       images,
		ErrorMessage: "test error",
	}

	assert.Len(t, response.Images, 1)
	assert.Equal(t, "test error", response.ErrorMessage)
	assert.Equal(t, "sha256:123", response.Images[0].ImageID)
}

func TestImageStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   model.ImageStatus
		expected string
	}{
		{
			name:     "Up to date status",
			status:   model.ImageStatusUpToDate,
			expected: "uptodate",
		},
		{
			name:     "Outdated status",
			status:   model.ImageStatusOutdated,
			expected: "outdated",
		},
		{
			name:     "Missing status",
			status:   model.ImageStatusMissing,
			expected: "missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestImageInfo(t *testing.T) {
	tests := []struct {
		name        string
		imageInfo   model.ImageInfo
		expectedStr string
	}{
		{
			name: "Complete image info",
			imageInfo: model.ImageInfo{
				ImageID:    "sha256:abc123",
				Repository: "babelcloud/gbox-playwright",
				Tag:        "b281f1a",
				Status:     model.ImageStatusUpToDate,
				Action:     "keep",
			},
			expectedStr: "sha256:abc123",
		},
		{
			name: "Missing image ID",
			imageInfo: model.ImageInfo{
				Repository: "babelcloud/gbox-playwright",
				Tag:        "latest",
				Status:     model.ImageStatusMissing,
				Action:     "pull",
			},
			expectedStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedStr, tt.imageInfo.ImageID)
			assert.NotEmpty(t, tt.imageInfo.Repository)
			assert.NotEmpty(t, tt.imageInfo.Tag)
			assert.NotEmpty(t, tt.imageInfo.Status)
		})
	}
}
