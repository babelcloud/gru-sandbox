// packages/api-server/internal/browser/service/page_action_vision.go
package service

import (
	// "encoding/json" // No longer needed for unmarshalling here
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
			moveErr = mouse.Move(float64(params.Path[i].X), float64(params.Path[i].Y))
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

// ExecuteVisionScreenshot handles the vision.screenshot action.
// It always saves the screenshot to a file. If no path is specified,
// it saves to the default shared directory with a timestamped filename.
func (s *BrowserService) ExecuteVisionScreenshot(boxID, contextID, pageID string, params model.VisionScreenshotParams) interface{} {
	targetPage, err := s.GetPageInstance(boxID, contextID, pageID)
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.screenshot failed to get page: %v", err)}
	}

	cfg := config.GetInstance() // Get config instance once
	screenshotOpts := playwright.PageScreenshotOptions{}
	savePath := ""
	baseScreenshotDir := filepath.Join(cfg.File.Share, boxID, "screenshot") // Define base dir for defaults/relatives

	// --- Determine Save Path ---
	var fileExt string = "png"
	if params.Type != nil && (*params.Type == "jpeg" || *params.Type == "jpg") {
		fileExt = "jpeg"
	}

	if params.Path != nil && *params.Path != "" {
		providedPath := *params.Path
		if filepath.IsAbs(providedPath) {
			// Use the user-provided absolute path
			savePath = providedPath
			fmt.Printf("DEBUG: Using provided absolute path: %s\n", savePath)
		} else {
			// Join the relative path with the base screenshot directory
			savePath = filepath.Join(baseScreenshotDir, providedPath)
			fmt.Printf("DEBUG: Relative path '%s' resolved to '%s'\n", providedPath, savePath)
		}
	} else {
		// Generate default filename and join with base screenshot directory
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("screenshot_%s.%s", timestamp, fileExt)
		savePath = filepath.Join(baseScreenshotDir, filename)
		fmt.Printf("DEBUG: Using default generated path: %s\n", savePath)
	}

	// --- Ensure Directory Exists ---
	// Get the directory part of the final save path
	finalDir := filepath.Dir(savePath)
	// Ensure the final target directory exists before attempting to save
	if err := os.MkdirAll(finalDir, 0755); err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("failed to create target directory '%s': %v", finalDir, err)}
	}
	// -----------------------------

	screenshotOpts.Path = &savePath // Set the final path
	// -------------------------

	// --- Map other options from params to screenshotOpts ---
	// Only set options if they are provided in params, otherwise let Playwright use defaults.
	if params.Type != nil {
		if *params.Type == "png" {
			screenshotOpts.Type = playwright.ScreenshotTypePng
		} else if *params.Type == "jpeg" || *params.Type == "jpg" {
			screenshotOpts.Type = playwright.ScreenshotTypeJpeg
		}
		// Add warning or error if value is invalid?
	}

	if params.FullPage != nil {
		screenshotOpts.FullPage = params.FullPage
	}

	// Check params.Type for JPEG before applying Quality
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

	// Map Rect (Clip)
	if params.Clip != nil {
		screenshotOpts.Clip = &playwright.Rect{
			X:      params.Clip.X,
			Y:      params.Clip.Y,
			Width:  params.Clip.Width,
			Height: params.Clip.Height,
		}
	}

	// Map Scale
	if params.Scale != nil {
		if *params.Scale == "css" {
			screenshotOpts.Scale = playwright.ScreenshotScaleCss
		} else if *params.Scale == "device" {
			screenshotOpts.Scale = playwright.ScreenshotScaleDevice
		}
		// Add warning or error if value is invalid?
	}

	// Map Animations
	if params.Animations != nil {
		if *params.Animations == "disabled" {
			screenshotOpts.Animations = playwright.ScreenshotAnimationsDisabled
		} else if *params.Animations == "allow" {
			screenshotOpts.Animations = playwright.ScreenshotAnimationsAllow
		}
		// Add warning or error if value is invalid?
	}

	// Map Caret
	if params.Caret != nil {
		if *params.Caret == "hide" {
			screenshotOpts.Caret = playwright.ScreenshotCaretHide
		} else if *params.Caret == "initial" {
			screenshotOpts.Caret = playwright.ScreenshotCaretInitial
		}
		// Add warning or error if value is invalid?
	}
	// --------------------------------------------------------

	// Log the options being used
	fmt.Printf("DEBUG: Executing screenshot with options: %+v\n", screenshotOpts)

	// Execute the screenshot command
	_, err = targetPage.Screenshot(screenshotOpts)
	if err != nil {
		// Consider checking for specific errors like permission denied vs other playwright errors
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.screenshot failed: %v", err)}
	}

	// Always return the path where the file was saved
	return model.VisionScreenshotResult{Success: true, SavedPath: savePath}
}

// ExecuteVisionScroll handles the vision.scroll action.
func (s *BrowserService) ExecuteVisionScroll(boxID, contextID, pageID string, params model.VisionScrollParams) interface{} {
	targetPage, err := s.GetPageInstance(boxID, contextID, pageID)
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.scroll failed: %v", err)}
	}
	_, err = targetPage.Evaluate("window.scrollBy(arguments[0], arguments[1])", params.ScrollX, params.ScrollY)
	if err != nil {
		return model.VisionErrorResult{Success: false, Error: fmt.Sprintf("vision.scroll failed: %v", err)}
	}
	return model.VisionScrollResult{Success: true}
}
