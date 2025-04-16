// packages/api-server/pkg/browser/page_action_params.go
package model

// MouseButtonType defines the allowed mouse buttons based on the schema.
type MouseButtonType string

const (
	MouseButtonLeft    MouseButtonType = "left"
	MouseButtonRight   MouseButtonType = "right"
	MouseButtonWheel   MouseButtonType = "wheel" // or "middle" if preferred for mapping
	MouseButtonBack    MouseButtonType = "back"
	MouseButtonForward MouseButtonType = "forward"
)

// Coordinate represents a point on the page using integer pixels.
type Coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// --- Vision Mode Action Parameter Structs ---

// VisionClickParams corresponds to the props for the vision.click action based on ClickActionSchema.
type VisionClickParams struct {
	// Button specifies the mouse button.
	Button MouseButtonType `json:"button"` // Use the enum type
	// X and Y are the integer coordinates to click at.
	X int `json:"x"` // Required
	Y int `json:"y"` // Required
	// TODO: Add Playwright options if needed (e.g., Delay, ClickCount)
}

// VisionDoubleClickParams corresponds to the props for the vision.doubleClick action based on DoubleClickActionSchema.
type VisionDoubleClickParams struct {
	// X and Y are the integer coordinates to double click at.
	X int `json:"x"` // Required
	Y int `json:"y"` // Required
	// TODO: Add Playwright options if needed (e.g., Button, Delay) - Note: Schema doesn't specify button for doubleClick
}

// VisionTypeParams corresponds to the props for the vision.type action based on TypeActionSchema.
// Note: No coordinate/selector target. Assumes typing into the currently focused element.
type VisionTypeParams struct {
	// Text is the text to type.
	Text string `json:"text"` // Required
	// TODO: Add Playwright options if needed (e.g., Delay)
}

// VisionDragParams corresponds to the props for the vision.drag action based on DragActionSchema.
type VisionDragParams struct {
	// Path defines the sequence of points (integer coordinates) to drag the mouse through.
	Path []Coordinate `json:"path"` // Required, must have at least one point
}

// VisionKeypressParams corresponds to the props for the vision.keypress action based on KeypressActionSchema.
type VisionKeypressParams struct {
	// Keys is an array of key names to press (e.g., "Shift", "a", "Enter").
	Keys []string `json:"keys"` // Required
	// TODO: Add playwright PressOptions like Delay.
}

// VisionMoveParams corresponds to the props for the vision.move action based on MoveActionSchema.
type VisionMoveParams struct {
	// X and Y are the integer coordinates to move the mouse to.
	X int `json:"x"` // Required
	Y int `json:"y"` // Required
	// TODO: Add playwright MouseMoveOptions like Steps.
}

// VisionScreenshotParams corresponds to the props for the vision.screenshot action based on ScreenshotActionSchema.
// Schema defines empty props.
type VisionScreenshotParams struct {
	// TODO: Add playwright PageScreenshotOptions if needed, although schema is empty.
}

// VisionScrollParams corresponds to the props for the vision.scroll action based on ScrollActionSchema.
type VisionScrollParams struct {
	// ScrollX is the horizontal scroll amount in pixels.
	ScrollX int `json:"scrollX"` // Required
	// ScrollY is the vertical scroll amount in pixels.
	ScrollY int `json:"scrollY"` // Required
	// X and Y coordinates from schema - represent the position *where* the scroll event originates.
	// Not directly used by the current `window.scrollBy` implementation.
	X int `json:"x"` // Required
	Y int `json:"y"` // Required
}

// VisionWaitParams corresponds to the props for the vision.wait action based on WaitActionSchema.
type VisionWaitParams struct {
	// Duration is the time to wait in milliseconds.
	Duration int `json:"duration"` // Required
}

// --- Snapshot Mode Action Parameter Structs (Placeholders) ---

// TODO: Define structs for Snapshot action parameters when implemented.
