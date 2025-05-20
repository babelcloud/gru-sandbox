package model_test

import (
	"testing"

	"github.com/babelcloud/gbox/packages/api-server/pkg/browser"
	"github.com/stretchr/testify/assert"
)

func TestPageActionTypes(t *testing.T) {
	tests := []struct {
		name     string
		action   model.PageActionType
		expected string
	}{
		{"VisionClick", model.ActionVisionClick, "vision.click"},
		{"VisionDoubleClick", model.ActionVisionDoubleClick, "vision.doubleClick"},
		{"VisionType", model.ActionVisionType, "vision.type"},
		{"VisionDrag", model.ActionVisionDrag, "vision.drag"},
		{"VisionKeyPress", model.ActionVisionKeyPress, "vision.keyPress"},
		{"VisionMove", model.ActionVisionMove, "vision.move"},
		{"VisionScreenshot", model.ActionVisionScreenshot, "vision.screenshot"},
		{"VisionScroll", model.ActionVisionScroll, "vision.scroll"},
		{"SnapshotClick", model.ActionSnapshotClick, "snapshot.click"},
		{"SnapshotHover", model.ActionSnapshotHover, "snapshot.hover"},
		{"SnapshotDrag", model.ActionSnapshotDrag, "snapshot.drag"},
		{"SnapshotType", model.ActionSnapshotType, "snapshot.type"},
		{"SnapshotSelectOption", model.ActionSnapshotSelectOption, "snapshot.selectOption"},
		{"SnapshotCapture", model.ActionSnapshotCapture, "snapshot.capture"},
		{"SnapshotTakeScreenshot", model.ActionSnapshotTakeScreenshot, "snapshot.takeScreenshot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, model.PageActionType(tt.expected), tt.action)
		})
	}
}

func TestMouseButtonTypes(t *testing.T) {
	tests := []struct {
		name     string
		button   model.MouseButtonType
		expected string
	}{
		{"Left", model.MouseButtonLeft, "left"},
		{"Right", model.MouseButtonRight, "right"},
		{"Wheel", model.MouseButtonWheel, "wheel"},
		{"Back", model.MouseButtonBack, "back"},
		{"Forward", model.MouseButtonForward, "forward"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, model.MouseButtonType(tt.expected), tt.button)
		})
	}
}

func TestVisionScreenshotParams(t *testing.T) {
	fullPage := true
	imgType := "jpeg"
	quality := 75
	omitBg := true
	timeout := 5000.0
	scale := "css"
	animations := "disabled"
	caret := "initial"
	outputFormat := "url"

	params := model.VisionScreenshotParams{
		FullPage:      &fullPage,
		Type:          &imgType,
		Quality:       &quality,
		OmitBackground: &omitBg,
		Timeout:       &timeout,
		Clip: &model.Rect{
			X:      10,
			Y:      20,
			Width:  100,
			Height: 200,
		},
		Scale:        &scale,
		Animations:   &animations,
		Caret:       &caret,
		OutputFormat: &outputFormat,
	}

	assert.Equal(t, true, *params.FullPage)
	assert.Equal(t, "jpeg", *params.Type)
	assert.Equal(t, 75, *params.Quality)
	assert.Equal(t, true, *params.OmitBackground)
	assert.Equal(t, 5000.0, *params.Timeout)
	assert.Equal(t, 10.0, params.Clip.X)
	assert.Equal(t, 20.0, params.Clip.Y)
	assert.Equal(t, 100.0, params.Clip.Width)
	assert.Equal(t, 200.0, params.Clip.Height)
	assert.Equal(t, "css", *params.Scale)
	assert.Equal(t, "disabled", *params.Animations)
	assert.Equal(t, "initial", *params.Caret)
	assert.Equal(t, "url", *params.OutputFormat)
}

func TestVisionActionResults(t *testing.T) {
	t.Run("SuccessResults", func(t *testing.T) {
		clickResult := model.VisionClickResult{Success: true}
		assert.True(t, clickResult.Success)

		doubleClickResult := model.VisionDoubleClickResult{Success: true}
		assert.True(t, doubleClickResult.Success)

		typeResult := model.VisionTypeResult{Success: true}
		assert.True(t, typeResult.Success)

		dragResult := model.VisionDragResult{Success: true}
		assert.True(t, dragResult.Success)

		keyPressResult := model.VisionKeyPressResult{Success: true}
		assert.True(t, keyPressResult.Success)

		moveResult := model.VisionMoveResult{Success: true}
		assert.True(t, moveResult.Success)

		scrollResult := model.VisionScrollResult{Success: true}
		assert.True(t, scrollResult.Success)
	})

	t.Run("ScreenshotResults", func(t *testing.T) {
		screenshotResult := model.VisionScreenshotResult{
			Success:       true,
			URL:          "http://example.com/screenshot.png",
			Base64Content: "base64encodedcontent",
		}
		assert.True(t, screenshotResult.Success)
		assert.Equal(t, "http://example.com/screenshot.png", screenshotResult.URL)
		assert.Equal(t, "base64encodedcontent", screenshotResult.Base64Content)

		base64Result := model.VisionScreenshotBase64Result{
			Success:    true,
			Screenshot: "base64encodedscreenshot",
		}
		assert.True(t, base64Result.Success)
		assert.Equal(t, "base64encodedscreenshot", base64Result.Screenshot)
	})

	t.Run("ErrorResult", func(t *testing.T) {
		errorResult := model.VisionErrorResult{
			Success: false,
			Error:   "Operation failed",
		}
		assert.False(t, errorResult.Success)
		assert.Equal(t, "Operation failed", errorResult.Error)
	})
}

func TestCoordinateAndRect(t *testing.T) {
	t.Run("Coordinate", func(t *testing.T) {
		coord := model.Coordinate{X: 100, Y: 200}
		assert.Equal(t, 100, coord.X)
		assert.Equal(t, 200, coord.Y)
	})

	t.Run("Rect", func(t *testing.T) {
		rect := model.Rect{
			X:      10.5,
			Y:      20.5,
			Width:  100.5,
			Height: 200.5,
		}
		assert.Equal(t, 10.5, rect.X)
		assert.Equal(t, 20.5, rect.Y)
		assert.Equal(t, 100.5, rect.Width)
		assert.Equal(t, 200.5, rect.Height)
	})
}
