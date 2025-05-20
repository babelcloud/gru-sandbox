package model_test

import (
	"encoding/json"
	"testing"

	"github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/stretchr/testify/assert"
)

func TestBoxError(t *testing.T) {
	tests := []struct {
		name     string
		err      model.BoxError
		expected string
	}{
		{
			name: "complete error",
			err: model.BoxError{
				Code:    "ERR001",
				Message: "Test error message",
				Details: "Error details",
			},
			expected: `{"code":"ERR001","message":"Test error message","details":"Error details"}`,
		},
		{
			name: "error without details",
			err: model.BoxError{
				Code:    "ERR002",
				Message: "Another error",
			},
			expected: `{"code":"ERR002","message":"Another error"}`,
		},
		{
			name: "empty error",
			err:  model.BoxError{},
			expected: `{"code":"","message":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			jsonBytes, err := json.Marshal(tt.err)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(jsonBytes))

			// Test JSON unmarshaling
			var unmarshalledErr model.BoxError
			err = json.Unmarshal([]byte(tt.expected), &unmarshalledErr)
			assert.NoError(t, err)
			assert.Equal(t, tt.err, unmarshalledErr)
		})
	}
}
