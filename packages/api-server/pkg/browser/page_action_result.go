// packages/api-server/pkg/browser/page_action_result.go
package model

// TabState represents the state part of TabMetadata.
type PageState struct {
	Loading bool `json:"loading"` // Whether the tab is currently loading
	Active  bool `json:"active"`  // Whether the tab is currently active
}

// TabMetadata mirrors the TabMetadataSchema defined in the TS schema.
type Page struct {
	Title   string    `json:"title"`   // The title of the tab
	URL     string    `json:"url"`     // The URL of the tab
	Favicon string    `json:"favicon"` // The favicon URL of the tab (may be empty if unavailable)
	State   PageState `json:"state"`   // The state of the tab
}

// FullScreenshot represents before/after screenshots of the full page.
type FullScreenshot struct {
	Before *string `json:"before"` // Screenshot of the full page before the action (nullable)
	After  *string `json:"after"`  // Screenshot of the full page after the action (nullable)
}

// ContentViewScreenshot mirrors the contentView part of the screenshot schema.
type ContentViewScreenshot struct {
	// Base64 encoded PNG/JPEG screenshot strings.
	Before *string `json:"before"` // Screenshot of the content view before the action (nullable)
	After  *string `json:"after"`  // Screenshot of the content view after the action (nullable)
}

// ScreenshotResult mirrors the screenshot part of the SuccessResultSchema.
type ScreenshotResult struct {
	FullScreen  FullScreenshot        `json:"fullScreen"`
	ContentView ContentViewScreenshot `json:"contentView"`
}

// SuccessResult mirrors the SuccessResultSchema.
type SuccessResult struct {
	Status     string           `json:"status"` // Always "success"
	Screenshot ScreenshotResult `json:"screenshot"`
	Video      string           `json:"video,omitempty"`
	Pages      []Page           `json:"pages"` // Re-added: State of all tabs after the action
}

// FailureResult mirrors the FailureResultSchema.
type FailureResult struct {
	Status string `json:"status"` // Always "failure"
	Error  string `json:"error"`  // Error message describing the failure
}

// PageActionResult represents the combined result structure.
// Only one of Success or Failure fields should be non-nil.
type PageActionResult struct {
	Status  string         `json:"status"` // "success" or "failure"
	Success *SuccessResult `json:"success,omitempty"`
	Failure *FailureResult `json:"failure,omitempty"`
}

// --- Helper Functions ---

// NewSuccessPageActionResult creates a success result.
// Takes the 'after' content screenshot (if any) and current tab state.
func NewSuccessPageActionResult(afterScreenshotBase64 *string, pages []Page) PageActionResult { // Parameter type corrected
	var beforeScreenshotBase64 *string = nil
	// Placeholder for before full screenshot
	var beforeFullScreenshotBase64 *string = nil
	// Placeholder for after full screenshot - needs implementation if required
	var afterFullScreenshotBase64 *string = nil

	return PageActionResult{
		Status: "success",
		Success: &SuccessResult{
			Status: "success",
			Screenshot: ScreenshotResult{
				FullScreen: FullScreenshot{ // Added field
					Before: beforeFullScreenshotBase64,
					After:  afterFullScreenshotBase64,
				},
				ContentView: ContentViewScreenshot{
					Before: beforeScreenshotBase64, // Assuming content view before
					After:  afterScreenshotBase64,  // Assuming content view after
				},
			},
			Pages: pages, // Assign the pages parameter
		},
		Failure: nil,
	}
}

// NewFailurePageActionResult creates a failure result.
func NewFailurePageActionResult(errMsg string) PageActionResult {
	return PageActionResult{
		Status:  "failure",
		Success: nil,
		Failure: &FailureResult{
			Status: "failure",
			Error:  errMsg,
		},
	}
}
