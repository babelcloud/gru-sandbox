// packages/api-server/pkg/browser/page_action_types.go
package model

// PageActionType defines the type of action to execute on a page.
type PageActionType string

//goland:noinspection GoSnakeCaseUsage
const (
	// --- Vision Mode Actions (Using Selectors/Coordinates) ---
	ActionVisionClick       PageActionType = "vision.click"       // Clicks an element matching the selector or at coordinates.
	ActionVisionDoubleClick PageActionType = "vision.doubleClick" // Double clicks an element matching the selector or at coordinates.
	ActionVisionType        PageActionType = "vision.type"        // Types text into an element matching the selector. Needs focus.
	ActionVisionDrag        PageActionType = "vision.drag"        // Drags the mouse along a path (coordinates).
	ActionVisionKeyPress    PageActionType = "vision.keyPress"    // Presses keyboard keys.
	ActionVisionMove        PageActionType = "vision.move"        // Moves the mouse cursor to specified coordinates.
	ActionVisionScreenshot  PageActionType = "vision.screenshot"  // Takes a screenshot of the page.
	ActionVisionScroll      PageActionType = "vision.scroll"      // Scrolls the page by a specified amount (coordinate based).
	ActionVisionWait        PageActionType = "vision.wait"        // Pauses execution for a specified duration.

	// --- Snapshot Mode Actions (Using Accessibility/DOM Snapshot References - Future Implementation) ---
	ActionSnapshotClick          PageActionType = "snapshot.click"          // Placeholder: Clicks an element based on a snapshot reference.
	ActionSnapshotHover          PageActionType = "snapshot.hover"          // Placeholder: Hovers an element based on a snapshot reference.
	ActionSnapshotDrag           PageActionType = "snapshot.drag"           // Placeholder: Drags between elements based on snapshot references.
	ActionSnapshotType           PageActionType = "snapshot.type"           // Placeholder: Types into an element based on a snapshot reference.
	ActionSnapshotSelectOption   PageActionType = "snapshot.selectOption"   // Placeholder: Selects dropdown option based on snapshot reference.
	ActionSnapshotCapture        PageActionType = "snapshot.capture"        // Placeholder: Captures an accessibility snapshot.
	ActionSnapshotTakeScreenshot PageActionType = "snapshot.takeScreenshot" // Placeholder: Takes a screenshot (potentially different options than vision).

	// TODO: Add other snapshot actions as needed.
)
