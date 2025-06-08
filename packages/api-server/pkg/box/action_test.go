package model_test

import (
	"testing"

	model "github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/stretchr/testify/assert"
)

func TestActionResultScreenshot(t *testing.T) {
	screenshot := model.ActionResultScreenshot{
		After: model.ActionResultScreenshotAfter{
			Uri: "after.png",
		},
		Before: model.ActionResultScreenshotBefore{
			Uri: "before.png",
		},
		Highlight: model.ActionResultScreenshotHighlight{
			Uri: "highlight.png",
		},
	}

	assert.Equal(t, "after.png", screenshot.After.Uri)
	assert.Equal(t, "before.png", screenshot.Before.Uri)
	assert.Equal(t, "highlight.png", screenshot.Highlight.Uri)
}

func TestBoxActionClickParams(t *testing.T) {
	params := model.BoxActionClickParams{
		Type:         "click",
		X:            100.5,
		Y:            200.5,
		Double:       true,
		Button:       "left",
		OutputFormat: "base64",
	}

	assert.Equal(t, "click", params.Type)
	assert.Equal(t, 100.5, params.X)
	assert.Equal(t, 200.5, params.Y)
	assert.True(t, params.Double)
	assert.Equal(t, "left", params.Button)
	assert.Equal(t, "base64", params.OutputFormat)

	// Test with different button types
	params.Button = "right"
	assert.Equal(t, "right", params.Button)
	params.Button = "middle"
	assert.Equal(t, "middle", params.Button)

	// Test with storageKey output format
	params.OutputFormat = "storageKey"
	assert.Equal(t, "storageKey", params.OutputFormat)
}

func TestBoxActionDragParams(t *testing.T) {
	path := []model.BoxActionDragParamsPath{
		{X: 10, Y: 20},
		{X: 30, Y: 40},
	}

	params := model.BoxActionDragParams{
		Path:         path,
		Type:         "drag",
		Duration:     "50ms",
		OutputFormat: "storageKey",
	}

	assert.Len(t, params.Path, 2)
	assert.Equal(t, 10.0, params.Path[0].X)
	assert.Equal(t, 20.0, params.Path[0].Y)
	assert.Equal(t, "drag", params.Type)
	assert.Equal(t, "50ms", params.Duration)
	assert.Equal(t, "storageKey", params.OutputFormat)

	// Test empty path
	params.Path = []model.BoxActionDragParamsPath{}
	assert.Len(t, params.Path, 0)
}

func TestBoxActionMoveParams(t *testing.T) {
	params := model.BoxActionMoveParams{
		Type:         "move",
		X:            150.0,
		Y:            250.0,
		OutputFormat: "base64",
	}

	assert.Equal(t, "move", params.Type)
	assert.Equal(t, 150.0, params.X)
	assert.Equal(t, 250.0, params.Y)
	assert.Equal(t, "base64", params.OutputFormat)

	// Test negative coordinates
	params.X = -10.0
	params.Y = -20.0
	assert.Equal(t, -10.0, params.X)
	assert.Equal(t, -20.0, params.Y)
}

func TestBoxActionPressParams(t *testing.T) {
	params := model.BoxActionPressParams{
		Keys:         []string{"Control", "A"},
		Type:         "press",
		OutputFormat: "base64",
	}

	assert.Len(t, params.Keys, 2)
	assert.Equal(t, "Control", params.Keys[0])
	assert.Equal(t, "A", params.Keys[1])
	assert.Equal(t, "press", params.Type)
	assert.Equal(t, "base64", params.OutputFormat)

	// Test empty keys array
	params.Keys = []string{}
	assert.Len(t, params.Keys, 0)
}

func TestBoxActionScreenshotParams(t *testing.T) {
	params := model.BoxActionScreenshotParams{
		Clip: model.BoxActionScreenshotParamsClip{
			Height: 100,
			Width:  200,
			X:      50,
			Y:      75,
		},
		OutputFormat: "base64",
		Type:         "png",
	}

	assert.Equal(t, 100.0, params.Clip.Height)
	assert.Equal(t, 200.0, params.Clip.Width)
	assert.Equal(t, 50.0, params.Clip.X)
	assert.Equal(t, 75.0, params.Clip.Y)
	assert.Equal(t, "base64", params.OutputFormat)
	assert.Equal(t, "png", params.Type)

	// Test jpeg type
	params.Type = "jpeg"
	assert.Equal(t, "jpeg", params.Type)
}

func TestBoxActionScrollParams(t *testing.T) {
	params := model.BoxActionScrollParams{
		ScrollX:      100,
		ScrollY:      200,
		Type:         "scroll",
		X:            50,
		Y:            75,
		OutputFormat: "base64",
	}

	assert.Equal(t, 100.0, params.ScrollX)
	assert.Equal(t, 200.0, params.ScrollY)
	assert.Equal(t, "scroll", params.Type)
	assert.Equal(t, 50.0, params.X)
	assert.Equal(t, 75.0, params.Y)
	assert.Equal(t, "base64", params.OutputFormat)

	// Test negative scroll values
	params.ScrollX = -100
	params.ScrollY = -200
	assert.Equal(t, -100.0, params.ScrollX)
	assert.Equal(t, -200.0, params.ScrollY)
}

func TestBoxActionTouchParams(t *testing.T) {
	points := []model.BoxActionTouchParamsPoint{
		{
			Start: model.BoxActionTouchParamsPointStart{
				X: 10,
				Y: 20,
			},
			Actions: []any{"tap", "swipe"},
		},
	}

	params := model.BoxActionTouchParams{
		Points:       points,
		Type:         "touch",
		OutputFormat: "base64",
	}

	assert.Len(t, params.Points, 1)
	assert.Equal(t, 10.0, params.Points[0].Start.X)
	assert.Equal(t, 20.0, params.Points[0].Start.Y)
	assert.Len(t, params.Points[0].Actions, 2)
	assert.Equal(t, "touch", params.Type)
	assert.Equal(t, "base64", params.OutputFormat)

	// Test multiple points
	params.Points = append(params.Points, model.BoxActionTouchParamsPoint{
		Start: model.BoxActionTouchParamsPointStart{
			X: 30,
			Y: 40,
		},
		Actions: []any{"tap"},
	})
	assert.Len(t, params.Points, 2)
}

func TestBoxActionTypeParams(t *testing.T) {
	params := model.BoxActionTypeParams{
		Text:         "Hello World",
		Type:         "type",
		OutputFormat: "base64",
	}

	assert.Equal(t, "Hello World", params.Text)
	assert.Equal(t, "type", params.Type)
	assert.Equal(t, "base64", params.OutputFormat)

	// Test empty text
	params.Text = ""
	assert.Empty(t, params.Text)
}

func TestBoxActionClickResult(t *testing.T) {
	result := model.BoxActionClickResult{
		Screenshot: model.ActionResultScreenshot{
			After:     model.ActionResultScreenshotAfter{Uri: "after.png"},
			Before:    model.ActionResultScreenshotBefore{Uri: "before.png"},
			Highlight: model.ActionResultScreenshotHighlight{Uri: "highlight.png"},
		},
	}
	assert.Equal(t, "after.png", result.Screenshot.After.Uri)
}

func TestBoxActionScreenshotResult(t *testing.T) {
	result := model.BoxActionScreenshotResult{
		Uri: "screenshot.png",
	}
	assert.Equal(t, "screenshot.png", result.Uri)
}

func TestBoxActionDragResult(t *testing.T) {
	result := model.BoxActionDragResult{
		Screenshot: model.ActionResultScreenshot{
			After:     model.ActionResultScreenshotAfter{Uri: "after.png"},
			Before:    model.ActionResultScreenshotBefore{Uri: "before.png"},
			Highlight: model.ActionResultScreenshotHighlight{Uri: "highlight.png"},
		},
	}
	assert.Equal(t, "after.png", result.Screenshot.After.Uri)
	assert.Equal(t, "before.png", result.Screenshot.Before.Uri)
	assert.Equal(t, "highlight.png", result.Screenshot.Highlight.Uri)
}

func TestBoxActionMoveResult(t *testing.T) {
	result := model.BoxActionMoveResult{
		Screenshot: model.ActionResultScreenshot{
			After:     model.ActionResultScreenshotAfter{Uri: "after.png"},
			Before:    model.ActionResultScreenshotBefore{Uri: "before.png"},
			Highlight: model.ActionResultScreenshotHighlight{Uri: "highlight.png"},
		},
	}
	assert.Equal(t, "after.png", result.Screenshot.After.Uri)
	assert.Equal(t, "before.png", result.Screenshot.Before.Uri)
	assert.Equal(t, "highlight.png", result.Screenshot.Highlight.Uri)
}

func TestBoxActionPressResult(t *testing.T) {
	result := model.BoxActionPressResult{
		Screenshot: model.ActionResultScreenshot{
			After:     model.ActionResultScreenshotAfter{Uri: "after.png"},
			Before:    model.ActionResultScreenshotBefore{Uri: "before.png"},
			Highlight: model.ActionResultScreenshotHighlight{Uri: "highlight.png"},
		},
	}
	assert.Equal(t, "after.png", result.Screenshot.After.Uri)
	assert.Equal(t, "before.png", result.Screenshot.Before.Uri)
	assert.Equal(t, "highlight.png", result.Screenshot.Highlight.Uri)
}

func TestBoxActionScrollResult(t *testing.T) {
	result := model.BoxActionScrollResult{
		Screenshot: model.ActionResultScreenshot{
			After:     model.ActionResultScreenshotAfter{Uri: "after.png"},
			Before:    model.ActionResultScreenshotBefore{Uri: "before.png"},
			Highlight: model.ActionResultScreenshotHighlight{Uri: "highlight.png"},
		},
	}
	assert.Equal(t, "after.png", result.Screenshot.After.Uri)
	assert.Equal(t, "before.png", result.Screenshot.Before.Uri)
	assert.Equal(t, "highlight.png", result.Screenshot.Highlight.Uri)
}

func TestBoxActionTouchResult(t *testing.T) {
	result := model.BoxActionTouchResult{
		Screenshot: model.ActionResultScreenshot{
			After:     model.ActionResultScreenshotAfter{Uri: "after.png"},
			Before:    model.ActionResultScreenshotBefore{Uri: "before.png"},
			Highlight: model.ActionResultScreenshotHighlight{Uri: "highlight.png"},
		},
	}
	assert.Equal(t, "after.png", result.Screenshot.After.Uri)
	assert.Equal(t, "before.png", result.Screenshot.Before.Uri)
	assert.Equal(t, "highlight.png", result.Screenshot.Highlight.Uri)
}

func TestBoxActionTypeResult(t *testing.T) {
	result := model.BoxActionTypeResult{
		Screenshot: model.ActionResultScreenshot{
			After:     model.ActionResultScreenshotAfter{Uri: "after.png"},
			Before:    model.ActionResultScreenshotBefore{Uri: "before.png"},
			Highlight: model.ActionResultScreenshotHighlight{Uri: "highlight.png"},
		},
	}
	assert.Equal(t, "after.png", result.Screenshot.After.Uri)
	assert.Equal(t, "before.png", result.Screenshot.Before.Uri)
	assert.Equal(t, "highlight.png", result.Screenshot.Highlight.Uri)
}
