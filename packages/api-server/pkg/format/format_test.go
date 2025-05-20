package format_test

import (
	"testing"

	"github.com/babelcloud/gbox/packages/api-server/pkg/format"
	"github.com/babelcloud/gbox/packages/api-server/pkg/logger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestFormatHTTPMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"GET method", "GET"},
		{"POST method", "POST"},
		{"PUT method", "PUT"},
		{"PATCH method", "PATCH"},
		{"DELETE method", "DELETE"},
		{"HEAD method", "HEAD"},
		{"OPTIONS method", "OPTIONS"},
		{"Unknown method", "CUSTOM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := format.FormatHTTPMethod(tt.method)
			assert.NotEmpty(t, result)
		})
	}
}

func TestFormatServerMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
	}{
		{"Docker mode", "docker"},
		{"Development mode", "development"},
		{"Production mode", "production"},
		{"Empty mode", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := format.FormatServerMode(tt.mode)
			assert.NotEmpty(t, result)
		})
	}
}

func TestLogAPIEndpoint(t *testing.T) {
	log := logrus.New()
	logger := &logger.Logger{Logger: log}

	tests := []struct {
		name     string
		endpoint format.APIEndpoint
	}{
		{
			name: "Basic endpoint",
			endpoint: format.APIEndpoint{
				Method:      "GET",
				Path:       "/api/test",
				Description: "Test endpoint",
			},
		},
		{
			name: "Empty fields",
			endpoint: format.APIEndpoint{
				Method:      "",
				Path:       "",
				Description: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format.LogAPIEndpoint(logger, tt.endpoint)
		})
	}
}

func TestLogAPIEndpoints(t *testing.T) {
	log := logrus.New()
	logger := &logger.Logger{Logger: log}

	tests := []struct {
		name      string
		endpoints []format.APIEndpoint
	}{
		{
			name: "Multiple endpoints",
			endpoints: []format.APIEndpoint{
				{
					Method:      "GET",
					Path:       "/api/test1",
					Description: "Test endpoint 1",
				},
				{
					Method:      "POST",
					Path:       "/api/test2",
					Description: "Test endpoint 2",
				},
			},
		},
		{
			name:      "Empty endpoints",
			endpoints: []format.APIEndpoint{},
		},
		{
			name:      "Nil endpoints",
			endpoints: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format.LogAPIEndpoints(logger, tt.endpoints)
		})
	}
}
