package model_test

import (
	"testing"

	"github.com/babelcloud/gbox/packages/api-server/pkg/browser"
	"github.com/stretchr/testify/assert"
)

func TestPageActionTypes(t *testing.T) {
	assert.Equal(t, model.PageActionType("vision.click"), model.ActionVisionClick)
	assert.Equal(t, model.PageActionType("vision.doubleClick"), model.ActionVisionDoubleClick)
	assert.Equal(t, model.PageActionType("vision.type"), model.ActionVisionType)
	assert.Equal(t, model.PageActionType("vision.drag"), model.ActionVisionDrag)
	assert.Equal(t, model.PageActionType("vision.keyPress"), model.ActionVisionKeyPress)
	assert.Equal(t, model.PageActionType("vision.move"), model.ActionVisionMove)
	assert.Equal(t, model.PageActionType("vision.screenshot"), model.ActionVisionScreenshot)
	assert.Equal(t, model.PageActionType("vision.scroll"), model.ActionVisionScroll)

	assert.Equal(t, model.PageActionType("snapshot.click"), model.ActionSnapshotClick)
	assert.Equal(t, model.PageActionType("snapshot.hover"), model.ActionSnapshotHover)
	assert.Equal(t, model.PageActionType("snapshot.drag"), model.ActionSnapshotDrag)
	assert.Equal(t, model.PageActionType("snapshot.type"), model.ActionSnapshotType)
	assert.Equal(t, model.PageActionType("snapshot.selectOption"), model.ActionSnapshotSelectOption)
	assert.Equal(t, model.PageActionType("snapshot.capture"), model.ActionSnapshotCapture)
	assert.Equal(t, model.PageActionType("snapshot.takeScreenshot"), model.ActionSnapshotTakeScreenshot)
}

func TestMouseButtonTypes(t *testing.T) {
	assert.Equal(t, model.MouseButtonType("left"), model.MouseButtonLeft)
	assert.Equal(t, model.MouseButtonType("right"), model.MouseButtonRight)
	assert.Equal(t, model.MouseButtonType("wheel"), model.MouseButtonWheel)
	assert.Equal(t, model.MouseButtonType("back"), model.MouseButtonBack)
	assert.Equal(t, model.MouseButtonType("forward"), model.MouseButtonForward)
}

func TestCoordinate(t *testing.T) {
	coord := model.Coordinate{X: 100, Y: 200}
	assert.Equal(t, 100, coord.X)
	assert.Equal(t, 200, coord.Y)
}

func TestRect(t *testing.T) {
	rect := model.Rect{X: 10.5, Y: 20.5, Width: 100.5, Height: 200.5}
	assert.Equal(t, 10.5, rect.X)
	assert.Equal(t, 20.5, rect.Y)
	assert.Equal(t, 100.5, rect.Width)
	assert.Equal(t, 200.5, rect.Height)
}

func TestVisionClickParams(t *testing.T) {
	params := model.VisionClickParams{
		Button: model.MouseButtonLeft,
		X: 100,
		Y: 200,
	}
	assert.Equal(t, model.MouseButtonLeft, params.Button)
	assert.Equal(t, 100, params.X)
	assert.Equal(t, 200, params.Y)
}

func TestVisionDoubleClickParams(t *testing.T) {
	params := model.VisionDoubleClickParams{
		X: 100,
		Y: 200,
	}
	assert.Equal(t, 100, params.X)
	assert.Equal(t, 200, params.Y)
}

func TestVisionTypeParams(t *testing.T) {
	params := model.VisionTypeParams{
		Text: "test text",
	}
	assert.Equal(t, "test text", params.Text)
}

func TestVisionDragParams(t *testing.T) {
	path := []model.Coordinate{
		{X: 100, Y: 200},
		{X: 300, Y: 400},
	}
	params := model.VisionDragParams{
		Path: path,
	}
	assert.Equal(t, path, params.Path)
}

func TestVisionKeyPressParams(t *testing.T) {
	keys := []string{"Shift", "a", "Enter"}
	params := model.VisionKeyPressParams{
		Keys: keys,
	}
	assert.Equal(t, keys, params.Keys)
}

func TestVisionMoveParams(t *testing.T) {
	params := model.VisionMoveParams{
		X: 100,
		Y: 200,
	}
	assert.Equal(t, 100, params.X)
	assert.Equal(t, 200, params.Y)
}

func TestVisionScreenshotParams(t *testing.T) {
	fullPage := true
	imgType := "jpeg"
	quality := 80
	omitBg := true
	timeout := 5000.0
	clip := &model.Rect{X: 0, Y: 0, Width: 100, Height: 100}
	scale := "css"
	animations := "disabled"
	caret := "hide"
	format := "base64"

	params := model.VisionScreenshotParams{
		FullPage: &fullPage,
		Type: &imgType,
		Quality: &quality,
		OmitBackground: &omitBg,
		Timeout: &timeout,
		Clip: clip,
		Scale: &scale,
		Animations: &animations,
		Caret: &caret,
		OutputFormat: &format,
	}

	assert.True(t, *params.FullPage)
	assert.Equal(t, "jpeg", *params.Type)
	assert.Equal(t, 80, *params.Quality)
	assert.True(t, *params.OmitBackground)
	assert.Equal(t, 5000.0, *params.Timeout)
	assert.Equal(t, clip, params.Clip)
	assert.Equal(t, "css", *params.Scale)
	assert.Equal(t, "disabled", *params.Animations)
	assert.Equal(t, "hide", *params.Caret)
	assert.Equal(t, "base64", *params.OutputFormat)
}

func TestVisionScrollParams(t *testing.T) {
	params := model.VisionScrollParams{
		ScrollX: 100,
		ScrollY: 200,
		X: 300,
		Y: 400,
	}
	assert.Equal(t, 100, params.ScrollX)
	assert.Equal(t, 200, params.ScrollY)
	assert.Equal(t, 300, params.X)
	assert.Equal(t, 400, params.Y)
}

func TestVisionResults(t *testing.T) {
	// Test success results
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

	// Test screenshot results
	screenshotResult := model.VisionScreenshotResult{
		Success: true,
		URL: "http://example.com/screenshot.png",
		Base64Content: "base64data",
	}
	assert.True(t, screenshotResult.Success)
	assert.Equal(t, "http://example.com/screenshot.png", screenshotResult.URL)
	assert.Equal(t, "base64data", screenshotResult.Base64Content)

	screenshotBase64Result := model.VisionScreenshotBase64Result{
		Success: true,
		Screenshot: "base64data",
	}
	assert.True(t, screenshotBase64Result.Success)
	assert.Equal(t, "base64data", screenshotBase64Result.Screenshot)

	// Test error result
	errorResult := model.VisionErrorResult{
		Success: false,
		Error: "Test error message",
	}
	assert.False(t, errorResult.Success)
	assert.Equal(t, "Test error message", errorResult.Error)
}

func TestConnectURLResult(t *testing.T) {
	result := model.ConnectURLResult{
		CdpURL: "ws://localhost:9222/devtools/browser/123",
	}
	assert.Equal(t, "ws://localhost:9222/devtools/browser/123", result.CdpURL)
}
