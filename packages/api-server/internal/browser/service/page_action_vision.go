// packages/api-server/internal/browser/service/page_action_vision.go
package service

import (
	// "encoding/json" // No longer needed for unmarshalling here
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/playwright-community/playwright-go"

	"github.com/babelcloud/gbox/packages/api-server/config" // Import config package
	model "github.com/babelcloud/gbox/packages/api-server/pkg/browser"
)

// --- Helper ---
// Helper to map our MouseButtonType enum to Playwright MouseButton type pointer
func mapMouseButton(button model.MouseButtonType) *playwright.MouseButton {
	switch button { // Use the enum directly
	case model.MouseButtonRight:
		return playwright.MouseButtonRight
	case model.MouseButtonWheel: // Use the enum constant
		return playwright.MouseButtonMiddle // Map wheel to middle
	case model.MouseButtonBack:
		return playwright.MouseButtonLeft // Defaulting
	case model.MouseButtonForward:
		return playwright.MouseButtonLeft // Defaulting
	case model.MouseButtonLeft:
		fallthrough
	default: // Includes empty string case if button wasn't required/provided correctly
		return playwright.MouseButtonLeft
	}
}

// --- Vision Actions ---

// ExecuteVisionClick handles the vision.click action.
func (s *BrowserService) ExecuteVisionClick(boxID, contextID, pageID string, params model.VisionClickParams) interface{} {
	targetPage, err := s.GetPageInstance(boxID, contextID, pageID)
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.click failed: %v", err)}
	}
	mouse := targetPage.Mouse()
	clickOpts := playwright.MouseClickOptions{Button: mapMouseButton(params.Button)}
	err = mouse.Click(float64(params.X), float64(params.Y), clickOpts)
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.click failed: %v", err)}
	}
	return model.VisionClickResult{Success: true}
}

// ExecuteVisionDoubleClick handles the vision.doubleClick action.
func (s *BrowserService) ExecuteVisionDoubleClick(boxID, contextID, pageID string, params model.VisionDoubleClickParams) interface{} {
	targetPage, err := s.GetPageInstance(boxID, contextID, pageID)
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.doubleClick failed: %v", err)}
	}
	mouse := targetPage.Mouse()
	dblClickOpts := playwright.MouseDblclickOptions{}
	err = mouse.Dblclick(float64(params.X), float64(params.Y), dblClickOpts)
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.doubleClick failed: %v", err)}
	}
	return model.VisionDoubleClickResult{Success: true}
}

// ExecuteVisionType handles the vision.type action.
func (s *BrowserService) ExecuteVisionType(boxID, contextID, pageID string, params model.VisionTypeParams) interface{} {
	targetPage, err := s.GetPageInstance(boxID, contextID, pageID)
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.type failed: %v", err)}
	}
	keyboard := targetPage.Keyboard()
	err = keyboard.Type(params.Text)
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.type failed: %v", err)}
	}
	return model.VisionTypeResult{Success: true}
}

// ExecuteVisionDrag handles the vision.drag action.
func (s *BrowserService) ExecuteVisionDrag(boxID, contextID, pageID string, params model.VisionDragParams) interface{} {
	targetPage, err := s.GetPageInstance(boxID, contextID, pageID)
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.drag failed: %v", err)}
	}
	mouse := targetPage.Mouse()
	if len(params.Path) == 0 {
		return model.VisionErrorResult{Success: false, Error: "vision.drag requires at least one point in the path"}
	}

	startX, startY := float64(params.Path[0].X), float64(params.Path[0].Y)
	err = mouse.Move(startX, startY)
	if err == nil {
		err = mouse.Down()
	}
	if err == nil {
		var moveErr error
		for i := 1; i < len(params.Path); i++ {
			moveErr = mouse.Move(float64(params.Path[i].X), float64(params.Path[i].Y), playwright.MouseMoveOptions{
				Steps: playwright.Int(5),
			})
			if moveErr != nil {
				err = moveErr
				break
			}
		}
		upErr := mouse.Up()
		if err == nil {
			err = upErr
		}
	}

	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.drag failed: %v", err)}
	}
	return model.VisionDragResult{Success: true}
}

// ExecuteVisionKeyPress handles the vision.keyPress action.
func (s *BrowserService) ExecuteVisionKeyPress(boxID, contextID, pageID string, params model.VisionKeyPressParams) interface{} {
	targetPage, err := s.GetPageInstance(boxID, contextID, pageID)
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.keyPress failed: %v", err)}
	}
	keyboard := targetPage.Keyboard()
	if len(params.Keys) == 0 {
		return model.VisionErrorResult{Success: false, Error: "keys array cannot be empty for vision.keyPress"}
	}
	var pressErr error // Changed variable name to avoid shadowing outer err
	for _, key := range params.Keys {
		if pressErr = keyboard.Press(key); pressErr != nil {
			break
		}
	}
	if pressErr != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.keyPress failed: %v", pressErr)}
	}
	return model.VisionKeyPressResult{Success: true}
}

// ExecuteVisionMove handles the vision.move action.
func (s *BrowserService) ExecuteVisionMove(boxID, contextID, pageID string, params model.VisionMoveParams) interface{} {
	targetPage, err := s.GetPageInstance(boxID, contextID, pageID)
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.move failed: %v", err)}
	}
	mouse := targetPage.Mouse()
	err = mouse.Move(float64(params.X), float64(params.Y))
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.move failed: %v", err)}
	}
	return model.VisionMoveResult{Success: true}
}

// ExecuteVisionScreenshot takes a screenshot and returns either its base64 content or a URL.
func (s *BrowserService) ExecuteVisionScreenshot(boxID, contextID, pageID string, params model.VisionScreenshotParams) interface{} {
	targetPage, err := s.GetPageInstance(boxID, contextID, pageID)
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.screenshot failed to get page: %v", err)}
	}

	cfg := config.GetInstance() // Get config instance once
	screenshotOpts := playwright.PageScreenshotOptions{}
	outputFormat := "base64" // Default output format
	if params.OutputFormat != nil && *params.OutputFormat == "url" {
		outputFormat = "url"
	}

	// --- Map common options from params to screenshotOpts ---
	if params.Type != nil {
		if *params.Type == "png" {
			screenshotOpts.Type = playwright.ScreenshotTypePng
		} else if *params.Type == "jpeg" || *params.Type == "jpg" {
			screenshotOpts.Type = playwright.ScreenshotTypeJpeg
		}
	}
	if params.FullPage != nil {
		screenshotOpts.FullPage = params.FullPage
	}
	var isJpeg bool
	if params.Type != nil && (*params.Type == "jpeg" || *params.Type == "jpg") {
		isJpeg = true
	}
	if params.Quality != nil && isJpeg {
		screenshotOpts.Quality = params.Quality
	}
	if params.OmitBackground != nil {
		screenshotOpts.OmitBackground = params.OmitBackground
	}
	if params.Timeout != nil {
		screenshotOpts.Timeout = params.Timeout
	}
	if params.Clip != nil {
		screenshotOpts.Clip = &playwright.Rect{
			X:      params.Clip.X,
			Y:      params.Clip.Y,
			Width:  params.Clip.Width,
			Height: params.Clip.Height,
		}
	}
	if params.Scale != nil {
		if *params.Scale == "css" {
			screenshotOpts.Scale = playwright.ScreenshotScaleCss
		} else if *params.Scale == "device" {
			screenshotOpts.Scale = playwright.ScreenshotScaleDevice
		}
	}
	if params.Animations != nil {
		if *params.Animations == "disabled" {
			screenshotOpts.Animations = playwright.ScreenshotAnimationsDisabled
		} else if *params.Animations == "allow" {
			screenshotOpts.Animations = playwright.ScreenshotAnimationsAllow
		}
	}
	if params.Caret != nil {
		if *params.Caret == "hide" {
			screenshotOpts.Caret = playwright.ScreenshotCaretHide
		} else if *params.Caret == "initial" {
			screenshotOpts.Caret = playwright.ScreenshotCaretInitial
		}
	}
	// --------------------------------------------------------

	fmt.Printf("DEBUG: Executing screenshot with options: %+v, OutputFormat: %s\n", screenshotOpts, outputFormat)

	if outputFormat == "url" {
		// --- Handle URL Output ---
		fileExt := "png" // Default extension
		if screenshotOpts.Type == playwright.ScreenshotTypeJpeg {
			fileExt = "jpeg"
		}

		// Define base dir and generate filename
		baseScreenshotDir := filepath.Join(cfg.File.Share, boxID, "screenshot")
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("screenshot_%s.%s", timestamp, fileExt)
		finalPath := filepath.Join(baseScreenshotDir, filename)
		relativeSavePath := filepath.Join("screenshot", filename) // Path relative to box share dir

		// Ensure the target directory exists
		if err := os.MkdirAll(baseScreenshotDir, 0755); err != nil {
			return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("failed to create target directory '%s': %v", baseScreenshotDir, err)}
		}

		// Set the path for saving the file
		screenshotOpts.Path = &finalPath

		// Execute screenshot to file
		_, err = targetPage.Screenshot(screenshotOpts)
		if err != nil {
			return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.screenshot (url mode) failed: %v", err)}
		}

		// Generate the access URL (adjust format as needed)
		// Assuming /api/v1/files/shared/{boxID}/{relativePath}
		accessURL := fmt.Sprintf("/api/v1/files/shared/%s/%s", boxID, relativeSavePath)

		return model.VisionScreenshotResult{Success: true, URL: accessURL}

	} else {
		// --- Handle Base64 Output (Default) ---
		// Ensure Path is not set when getting buffer
		screenshotOpts.Path = nil

		// Execute screenshot to buffer
		buffer, err := targetPage.Screenshot(screenshotOpts)
		if err != nil {
			return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.screenshot (base64 mode) failed: %v", err)}
		}

		// Encode buffer to base64
		encodedString := base64.StdEncoding.EncodeToString(buffer)

		return model.VisionScreenshotResult{Success: true, Base64Content: encodedString}
	}
}

// ExecuteVisionScroll handles the vision.scroll action.
func (s *BrowserService) ExecuteVisionScroll(boxID, contextID, pageID string, params model.VisionScrollParams) interface{} {
	targetPage, err := s.GetPageInstance(boxID, contextID, pageID)
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.scroll failed to get page instance: %v", err)}
	}

	// Always scroll the window
	// Use page.Mouse().Wheel() for potentially more reliable scrolling simulation
	err = targetPage.Mouse().Wheel(float64(params.ScrollX), float64(params.ScrollY))
	if err != nil {
		// Improve error reporting
		errMsg := fmt.Sprintf("Mouse.Wheel failed: %v", err)
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.scroll failed: %s", errMsg)}
	}

	// If successful, return the success result
	return model.VisionScrollResult{Success: true}
}
