package model_test

import (
	"testing"

	model "github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	t.Run("create filter with valid values", func(t *testing.T) {
		filter := model.Filter{
			Field:    "id",
			Operator: model.FilterOperatorEquals,
			Value:    "123",
		}

		assert.Equal(t, "id", filter.Field)
		assert.Equal(t, model.FilterOperatorEquals, filter.Operator)
		assert.Equal(t, "123", filter.Value)
	})
}

func TestBoxListParams(t *testing.T) {
	t.Run("create params with no filters", func(t *testing.T) {
		params := model.BoxListParams{}
		assert.Empty(t, params.Filters)
	})

	t.Run("create params with filters", func(t *testing.T) {
		filter := model.Filter{
			Field:    "label",
			Operator: model.FilterOperatorEquals,
			Value:    "test",
		}

		params := model.BoxListParams{
			Filters: []model.Filter{filter},
		}

		assert.Len(t, params.Filters, 1)
		assert.Equal(t, filter, params.Filters[0])
	})
}

func TestBoxListResult(t *testing.T) {
	t.Run("create empty result", func(t *testing.T) {
		result := model.BoxListResult{}
		assert.Empty(t, result.Data)
		assert.Equal(t, 0, result.Total)
		assert.Empty(t, result.Message)
		assert.Equal(t, float64(0), result.Page)
		assert.Equal(t, float64(0), result.PageSize)
	})

	t.Run("create result with data", func(t *testing.T) {
		boxes := []model.Box{{ID: "1"}, {ID: "2"}}
		result := model.BoxListResult{
			Data:     boxes,
			Total:    2,
			Message:  "Success",
			Page:     1,
			PageSize: 10,
		}

		assert.Equal(t, boxes, result.Data)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, "Success", result.Message)
		assert.Equal(t, float64(1), result.Page)
		assert.Equal(t, float64(10), result.PageSize)
	})
}
