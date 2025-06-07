package model

// ActionResultScreenshot (common for many results)
type ActionResultScreenshot struct {
	// URI of the screenshot after the action
	After ActionResultScreenshotAfter `json:"after,required"`
	// URI of the screenshot before the action
	Before ActionResultScreenshotBefore `json:"before,required"`
	// URI of the screenshot before the action with highlight
	Highlight ActionResultScreenshotHighlight `json:"highlight,required"`
}

type ActionResultScreenshotAfter struct {
	// URI of the screenshot after the action
	Uri string `json:"uri,required"`
}

type ActionResultScreenshotBefore struct {
	// URI of the screenshot before the action
	Uri string `json:"uri,required"`
}

type ActionResultScreenshotHighlight struct {
	// URI of the screenshot before the action with highlight
	Uri string `json:"uri,required"`
}

// --- Click Action ---
type BoxActionClickParams struct {
	Type any `json:"type,omitzero,required"`
	// X coordinate of the click
	X float64 `json:"x,required"`
	// Y coordinate of the click
	Y float64 `json:"y,required"`
	// Whether to perform a double click
	Double bool `json:"double,omitzero"`
	// Mouse button to click
	// Any of "left", "right", "middle".
	Button string `json:"button,omitzero"`
	// Type of the URI
	// Any of "base64", "storageKey".
	OutputFormat string `json:"outputFormat,omitzero"`
}

type BoxActionClickResult struct {
	Screenshot ActionResultScreenshot `json:"screenshot,required"`
}

// --- Drag Action ---
type BoxActionDragParamsPath struct {
	// X coordinate of a point in the drag path
	X float64 `json:"x,required"`
	// Y coordinate of a point in the drag path
	Y float64 `json:"y,required"`
}

type BoxActionDragParams struct {
	// Path of the drag action as a series of coordinates
	Path []BoxActionDragParamsPath `json:"path,omitzero,required"`
	// Action type for drag interaction
	Type any `json:"type,omitzero,required"`
	// Time interval between points (e.g. "50ms")
	Duration string `json:"duration,omitzero"`
	// Type of the URI
	// Any of "base64", "storageKey".
	OutputFormat string `json:"outputFormat,omitzero"`
}

type BoxActionDragResult struct {
	Screenshot ActionResultScreenshot `json:"screenshot,required"`
}

// --- Move Action ---
type BoxActionMoveParams struct {
	// Action type for cursor movement
	Type any `json:"type,omitzero,required"`
	// X coordinate to move to
	X float64 `json:"x,required"`
	// Y coordinate to move to
	Y float64 `json:"y,required"`
	// Type of the URI
	// Any of "base64", "storageKey".
	OutputFormat string `json:"outputFormat,omitzero"`
}

type BoxActionMoveResult struct {
	Screenshot ActionResultScreenshot `json:"screenshot,required"`
}

// --- Press Action ---
type BoxActionPressParams struct {
	// Array of keys to press
	Keys []string `json:"keys,omitzero,required"`
	// Action type for keyboard key press
	Type any `json:"type,omitzero,required"`
	// Type of the URI
	// Any of "base64", "storageKey".
	OutputFormat string `json:"outputFormat,omitzero"`
}

type BoxActionPressResult struct {
	Screenshot ActionResultScreenshot `json:"screenshot,required"`
}

// --- Screenshot Action ---
type BoxActionScreenshotParamsClip struct {
	// Height of the clip
	Height float64 `json:"height,required"`
	// Width of the clip
	Width float64 `json:"width,required"`
	// X coordinate of the clip
	X float64 `json:"x,required"`
	// Y coordinate of the clip
	Y float64 `json:"y,required"`
}

type BoxActionScreenshotParams struct {
	// clip of the screenshot
	Clip BoxActionScreenshotParamsClip `json:"clip,omitzero"`
	// Type of the URI
	// Any of "base64", "storageKey".
	OutputFormat string `json:"outputFormat,omitzero"`
	// Action type for screenshot
	// Any of "png", "jpeg".
	Type string `json:"type,omitzero"`
}

type BoxActionScreenshotResult struct {
	// URL of the screenshot
	Uri string `json:"uri,required"`
}

// --- Scroll Action ---
type BoxActionScrollParams struct {
	// Horizontal scroll amount
	ScrollX float64 `json:"scrollX,required"`
	// Vertical scroll amount
	ScrollY float64 `json:"scrollY,required"`
	// Action type for scroll interaction
	Type any `json:"type,omitzero,required"`
	// X coordinate of the scroll position
	X float64 `json:"x,required"`
	// Y coordinate of the scroll position
	Y float64 `json:"y,required"`
	// Type of the URI
	// Any of "base64", "storageKey".
	OutputFormat string `json:"outputFormat,omitzero"`
}

type BoxActionScrollResult struct {
	Screenshot ActionResultScreenshot `json:"screenshot,required"`
}

// --- Touch Action ---
type BoxActionTouchParamsPointStart struct {
	// Starting X coordinate
	X float64 `json:"x,required"`
	// Starting Y coordinate
	Y float64 `json:"y,required"`
}

type BoxActionTouchParamsPoint struct {
	// Starting position for touch
	Start BoxActionTouchParamsPointStart `json:"start,omitzero,required"`
	// Sequence of actions to perform after initial touch
	Actions []any `json:"actions,omitzero"`
}

type BoxActionTouchParams struct {
	// Array of touch points and their actions
	Points []BoxActionTouchParamsPoint `json:"points,omitzero,required"`
	// Action type for touch interaction
	Type any `json:"type,omitzero,required"`
	// Type of the URI
	// Any of "base64", "storageKey".
	OutputFormat string `json:"outputFormat,omitzero"`
}

type BoxActionTouchResult struct {
	Screenshot ActionResultScreenshot `json:"screenshot,required"`
}

// --- Type Action ---
type BoxActionTypeParams struct {
	// Text to type
	Text string `json:"text,required"`
	// Action type for typing text
	Type any `json:"type,omitzero,required"`
	// Type of the URI
	// Any of "base64", "storageKey".
	OutputFormat string `json:"outputFormat,omitzero"`
}

type BoxActionTypeResult struct {
	Screenshot ActionResultScreenshot `json:"screenshot,required"`
}