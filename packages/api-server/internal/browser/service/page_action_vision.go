// packages/api-server/internal/browser/service/page_action_vision.go
package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/playwright-community/playwright-go"

	model "github.com/babelcloud/gbox/packages/api-server/pkg/browser"
)

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

// executeVisionAction handles the execution of page actions operating in Vision mode.
func (s *BrowserService) executeVisionAction(targetPage playwright.Page, action model.PageActionType, paramsRaw json.RawMessage) (result interface{}, err error) {
	mouse := targetPage.Mouse()       // Get mouse instance
	keyboard := targetPage.Keyboard() // Get keyboard instance

	switch action {
	case model.ActionVisionClick:
		var props model.VisionClickParams
		if err := json.Unmarshal(paramsRaw, &props); err != nil {
			return nil, fmt.Errorf("invalid params for %s: %w", action, err)
		}
		clickOpts := playwright.MouseClickOptions{
			Button: mapMouseButton(props.Button), // Pass the enum value
		}
		err = mouse.Click(float64(props.X), float64(props.Y), clickOpts)
		return nil, err

	case model.ActionVisionDoubleClick:
		var props model.VisionDoubleClickParams
		if err := json.Unmarshal(paramsRaw, &props); err != nil {
			return nil, fmt.Errorf("invalid params for %s: %w", action, err)
		}
		// Vision double click uses coordinates directly
		// Schema does not specify button, Playwright defaults to Left
		dblClickOpts := playwright.MouseDblclickOptions{
			// Button: playwright.MouseButtonLeft, // Default
			// TODO: Expose Delay from props if added
		}
		err = mouse.Dblclick(float64(props.X), float64(props.Y), dblClickOpts)
		return nil, err

	case model.ActionVisionType:
		var props model.VisionTypeParams
		if err := json.Unmarshal(paramsRaw, &props); err != nil {
			return nil, fmt.Errorf("invalid params for %s: %w", action, err)
		}
		if props.Text == "" {
			// Allow empty string type? Or return error? Let's allow for now.
			// return nil, fmt.Errorf("text cannot be empty for %s", action)
		}
		// Vision type assumes typing into the currently focused element
		// TODO: Expose Delay option
		err = keyboard.Type(props.Text)
		return nil, err

	case model.ActionVisionDrag:
		var props model.VisionDragParams
		if err := json.Unmarshal(paramsRaw, &props); err != nil {
			return nil, fmt.Errorf("invalid params for %s: %w", action, err)
		}
		if len(props.Path) == 0 {
			return nil, fmt.Errorf("vision.drag action requires at least one point in the path")
		}

		startX, startY := float64(props.Path[0].X), float64(props.Path[0].Y)
		err = mouse.Move(startX, startY)
		if err != nil {
			return nil, err
		}
		err = mouse.Down() // Assuming left button drag
		if err != nil {
			return nil, err
		}
		var moveErr error
		for i := 1; i < len(props.Path); i++ {
			moveErr = mouse.Move(float64(props.Path[i].X), float64(props.Path[i].Y))
			if moveErr != nil {
				err = moveErr // Store the move error
				break         // Stop moving
			}
		}
		upErr := mouse.Up()
		if err == nil {
			err = upErr
		} // Prioritize the move error
		return nil, err

	case model.ActionVisionKeypress:
		var props model.VisionKeypressParams
		if err := json.Unmarshal(paramsRaw, &props); err != nil {
			return nil, fmt.Errorf("invalid params for %s: %w", action, err)
		}
		if len(props.Keys) == 0 {
			return nil, fmt.Errorf("keys array cannot be empty for %s", action)
		}
		for _, key := range props.Keys {
			pressErr := keyboard.Press(key)
			if pressErr != nil {
				err = pressErr
				break
			}
		}
		return nil, err

	case model.ActionVisionMove:
		var props model.VisionMoveParams
		if err := json.Unmarshal(paramsRaw, &props); err != nil {
			return nil, fmt.Errorf("invalid params for %s: %w", action, err)
		}
		// TODO: Expose Steps option
		err = mouse.Move(float64(props.X), float64(props.Y))
		return nil, err

	case model.ActionVisionScreenshot:
		var props model.VisionScreenshotParams // Parse empty props for consistency
		if err := json.Unmarshal(paramsRaw, &props); err != nil {
			return nil, fmt.Errorf("invalid params for %s: %w", action, err)
		}
		// TODO: Expose PageScreenshotOptions based on props fields when added
		screenshotBytes, err := targetPage.Screenshot()
		if err != nil {
			return nil, err
		}
		return base64.StdEncoding.EncodeToString(screenshotBytes), nil

	case model.ActionVisionScroll:
		var props model.VisionScrollParams
		if err := json.Unmarshal(paramsRaw, &props); err != nil {
			return nil, fmt.Errorf("invalid params for %s: %w", action, err)
		}
		// Use window.scrollBy for relative scrolling based on scrollX, scrollY
		// The props.X and props.Y indicating origin are ignored here.
		_, err = targetPage.Evaluate("window.scrollBy(arguments[0], arguments[1])", props.ScrollX, props.ScrollY)
		return nil, err

	case model.ActionVisionWait:
		var props model.VisionWaitParams
		if err := json.Unmarshal(paramsRaw, &props); err != nil {
			return nil, fmt.Errorf("invalid params for %s: %w", action, err)
		}
		targetPage.WaitForTimeout(float64(props.Duration)) // Playwright expects float64 milliseconds
		return nil, nil

	default:
		return nil, fmt.Errorf("unknown vision action: %s", action)
	}
}
