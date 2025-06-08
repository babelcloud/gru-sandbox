package model_test

import (
	"testing"

	"github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	tests := []struct {
		name     string
		filter   model.Filter
		expected model.Filter
	}{
		{
			name: "valid filter",
			filter: model.Filter{
				Field:    "id",
				Operator: model.FilterOperatorEquals,
				Value:    "123",
			},
			expected: model.Filter{
				Field:    "id",
				Operator: model.FilterOperatorEquals,
				Value:    "123",
			},
		},
		{
			name: "empty filter",
			filter: model.Filter{
				Field:    "",
				Operator: "",
				Value:    "",
			},
			expected: model.Filter{
				Field:    "",
				Operator: "",
				Value:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.filter)
		})
	}
}

func TestBoxListParams(t *testing.T) {
	tests := []struct {
		name     string
		params   model.BoxListParams
		expected model.BoxListParams
	}{
		{
			name: "params with filters",
			params: model.BoxListParams{
				Filters: []model.Filter{
					{
						Field:    "id",
						Operator: model.FilterOperatorEquals,
						Value:    "123",
					},
				},
			},
			expected: model.BoxListParams{
				Filters: []model.Filter{
					{
						Field:    "id",
						Operator: model.FilterOperatorEquals,
						Value:    "123",
					},
				},
			},
		},
		{
			name: "empty params",
			params: model.BoxListParams{
				Filters: nil,
			},
			expected: model.BoxListParams{
				Filters: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.params)
		})
	}
}

func TestBoxListResult(t *testing.T) {
	tests := []struct {
		name     string
		result   model.BoxListResult
		expected model.BoxListResult
	}{
		{
			name: "result with data",
			result: model.BoxListResult{
				Data:     []model.Box{{ID: "123"}},
				Total:    1,
				Message:  "Success",
				Page:     1.0,
				PageSize: 10.0,
			},
			expected: model.BoxListResult{
				Data:     []model.Box{{ID: "123"}},
				Total:    1,
				Message:  "Success",
				Page:     1.0,
				PageSize: 10.0,
			},
		},
		{
			name: "empty result",
			result: model.BoxListResult{
				Data:     nil,
				Total:    0,
				Message:  "",
				Page:     0,
				PageSize: 0,
			},
			expected: model.BoxListResult{
				Data:     nil,
				Total:    0,
				Message:  "",
				Page:     0,
				PageSize: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result)
		})
	}
}
