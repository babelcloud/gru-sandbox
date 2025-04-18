package model

// --- Types and Constants ---

// PageActionType defines the type of action to execute on a page.
type PageActionType string

// MouseButtonType defines the allowed mouse buttons based on the schema.
type MouseButtonType string

//goland:noinspection GoSnakeCaseUsage
const (
	// --- Action Types ---
	ActionVisionClick       PageActionType = "vision.click"
	ActionVisionDoubleClick PageActionType = "vision.doubleClick"
	ActionVisionType        PageActionType = "vision.type"
	ActionVisionDrag        PageActionType = "vision.drag"
	ActionVisionKeyPress    PageActionType = "vision.keyPress"
	ActionVisionMove        PageActionType = "vision.move"
	ActionVisionScreenshot  PageActionType = "vision.screenshot"
	ActionVisionScroll      PageActionType = "vision.scroll"
	// REMOVED: ActionVisionWait

	ActionSnapshotClick          PageActionType = "snapshot.click"
	ActionSnapshotHover          PageActionType = "snapshot.hover"
	ActionSnapshotDrag           PageActionType = "snapshot.drag"
	ActionSnapshotType           PageActionType = "snapshot.type"
	ActionSnapshotSelectOption   PageActionType = "snapshot.selectOption"
	ActionSnapshotCapture        PageActionType = "snapshot.capture"
	ActionSnapshotTakeScreenshot PageActionType = "snapshot.takeScreenshot"

	// --- Mouse Buttons ---
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

// Rect represents a rectangular area on the page using float pixels.
// Based on Playwright's Rect type.
type Rect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
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

// VisionKeyPressParams corresponds to the props for the vision.keyPress action based on KeyPressActionSchema.
type VisionKeyPressParams struct {
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
// Aligned with relevant options from playwright.PageScreenshotOptions.
type VisionScreenshotParams struct {
	// Path specifies the file path to save the screenshot to.
	// If omitted, the screenshot is returned as a base64 encoded string.
	Path *string `json:"path,omitempty"`

	// When true, takes a screenshot of the full scrollable page. Defaults to false.
	FullPage *bool `json:"fullPage,omitempty"`

	// Specifies the screenshot type, e.g., "png" or "jpeg". Defaults to "png".
	Type *string `json:"type,omitempty"` // Consider using an enum like ScreenshotType if strict validation is needed.

	// The quality of the image (0-100) for "jpeg" type. Not applicable to "png".
	Quality *int `json:"quality,omitempty"`

	// Hides the default white background, allowing transparency (for "png"). Defaults to false.
	OmitBackground *bool `json:"omitBackground,omitempty"`

	// Maximum time in milliseconds to wait. Defaults to 30000 (30s). 0 disables timeout.
	Timeout *float64 `json:"timeout,omitempty"`

	// An object specifying clipping of the resulting image.
	Clip *Rect `json:"clip,omitempty"`

	// Scale of the screenshot ('css' or 'device'). Defaults to 'device'.
	Scale *string `json:"scale,omitempty"` // Consider enum ScreenshotScale

	// Animation handling ('allow', 'disabled'). Defaults to 'allow'.
	Animations *string `json:"animations,omitempty"` // Consider enum ScreenshotAnimations

	// Caret visibility ('hide', 'initial'). Defaults to 'hide'.
	Caret *string `json:"caret,omitempty"` // Consider enum ScreenshotCaret

	// TODO: Consider adding Mask, MaskColor, Style if needed later.
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

// --- Vision Mode Action Result Structs ---

// VisionClickResult represents the result of a successful vision.click action.
type VisionClickResult struct {
	Success bool `json:"success"` // Always true for this type
}

// VisionDoubleClickResult represents the result of a successful vision.doubleClick action.
type VisionDoubleClickResult struct {
	Success bool `json:"success"` // Always true for this type
}

// VisionTypeResult represents the result of a successful vision.type action.
type VisionTypeResult struct {
	Success bool `json:"success"` // Always true for this type
}

// VisionDragResult represents the result of a successful vision.drag action.
type VisionDragResult struct {
	Success bool `json:"success"` // Always true for this type
}

// VisionKeyPressResult represents the result of a successful vision.keyPress action.
type VisionKeyPressResult struct {
	Success bool `json:"success"` // Always true for this type
}

// VisionMoveResult represents the result of a successful vision.move action.
type VisionMoveResult struct {
	Success bool `json:"success"` // Always true for this type
}

// VisionScrollResult represents the result of a successful vision.scroll action.
type VisionScrollResult struct {
	Success bool `json:"success"` // Always true for this type
}

// VisionScreenshotResult is returned when ActionVisionScreenshot saves to a file.
type VisionScreenshotResult struct {
	Success   bool   `json:"success"`   // Always true for this type
	SavedPath string `json:"savedPath"` // The absolute path where the screenshot was saved.
}

// VisionScreenshotBase64Result is returned when ActionVisionScreenshot returns base64 data.
type VisionScreenshotBase64Result struct {
	Success    bool   `json:"success"`    // Always true for this type
	Screenshot string `json:"screenshot"` // Base64 encoded screenshot data
}

// VisionErrorResult represents the result of a failed vision action.
type VisionErrorResult struct {
	Success bool   `json:"success"` // Always false for this type
	Error   string `json:"error"`   // Error message describing the failure
}

// --- Snapshot Mode Action Parameter Structs (Placeholders) ---
// TODO: Define structs for Snapshot action parameters when implemented.

// --- Snapshot Mode Action Result Structs (Placeholders) ---
// TODO: Define structs for Snapshot action results when implemented.
