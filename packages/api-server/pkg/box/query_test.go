package model_test

import (
	"testing"

	model "github.com/babelcloud/gbox/packages/api-server/pkg/box"
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
			name: "empty field",
			filter: model.Filter{
				Field:    "",
				Operator: model.FilterOperatorEquals,
				Value:    "123",
			},
			expected: model.Filter{
				Field:    "",
				Operator: model.FilterOperatorEquals,
				Value:    "123",
			},
		},
		{
			name: "empty value",
			filter: model.Filter{
				Field:    "id",
				Operator: model.FilterOperatorEquals,
				Value:    "",
			},
			expected: model.Filter{
				Field:    "id",
				Operator: model.FilterOperatorEquals,
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
			name: "empty filters",
			params: model.BoxListParams{
				Filters: []model.Filter{},
			},
			expected: model.BoxListParams{
				Filters: []model.Filter{},
			},
		},
		{
			name: "single filter",
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
			name: "multiple filters",
			params: model.BoxListParams{
				Filters: []model.Filter{
					{
						Field:    "id",
						Operator: model.FilterOperatorEquals,
						Value:    "123",
					},
					{
						Field:    "label",
						Operator: model.FilterOperatorEquals,
						Value:    "test",
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
					{
						Field:    "label",
						Operator: model.FilterOperatorEquals,
						Value:    "test",
					},
				},
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
			name: "empty result",
			result: model.BoxListResult{
				Boxes:   []model.Box{},
				Count:   0,
				Message: "",
			},
			expected: model.BoxListResult{
				Boxes:   []model.Box{},
				Count:   0,
				Message: "",
			},
		},
		{
			name: "result with message",
			result: model.BoxListResult{
				Boxes:   []model.Box{},
				Count:   0,
				Message: "test message",
			},
			expected: model.BoxListResult{
				Boxes:   []model.Box{},
				Count:   0,
				Message: "test message",
			},
		},
		{
			name: "result with boxes",
			result: model.BoxListResult{
				Boxes: []model.Box{
					{},
					{},
				},
				Count:   2,
				Message: "",
			},
			expected: model.BoxListResult{
				Boxes: []model.Box{
					{},
					{},
				},
				Count:   2,
				Message: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result)
		})
	}
}
