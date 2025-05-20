package model_test

import (
	"testing"

	model "github.com/babelcloud/gbox/packages/api-server/pkg/browser"
	"github.com/stretchr/testify/assert"
)

func TestCreateContextParams(t *testing.T) {
	params := model.CreateContextParams{
		ViewportWidth:  1920,
		ViewportHeight: 1080,
		UserAgent:      "Mozilla/5.0",
		Locale:         "en-US",
		Timezone:       "UTC",
		Permissions:    []string{"geolocation", "notifications"},
	}

	assert.Equal(t, 1920, params.ViewportWidth)
	assert.Equal(t, 1080, params.ViewportHeight)
	assert.Equal(t, "Mozilla/5.0", params.UserAgent)
	assert.Equal(t, "en-US", params.Locale)
	assert.Equal(t, "UTC", params.Timezone)
	assert.Equal(t, []string{"geolocation", "notifications"}, params.Permissions)
}

func TestCreateContextResult(t *testing.T) {
	result := model.CreateContextResult{
		ContextID: "test-context-id",
	}

	assert.Equal(t, "test-context-id", result.ContextID)
}

func TestNewCreateContextResult(t *testing.T) {
	tests := []struct {
		name      string
		contextID string
	}{
		{
			name:      "with valid context ID",
			contextID: "test-context-id",
		},
		{
			name:      "with empty context ID",
			contextID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.NewCreateContextResult(tt.contextID)
			assert.Equal(t, tt.contextID, result.ContextID)
		})
	}
}
