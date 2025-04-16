package service

import (
	"fmt"
	"strings"

	model "github.com/babelcloud/gbox/packages/api-server/pkg/browser"
)

// ExecuteAction executes an action on a specific managed page.
// It dispatches the execution to vision or snapshot handlers based on the action prefix.
// Returns the new PageActionResult structure.
func (s *BrowserService) ExecuteAction(boxID, contextID, pageID string, params model.PageActionParams) (*model.PageActionResult, error) {
	mp, err := s.findManagedPage(pageID)
	if err != nil {
		// Return specific error structure if page not found or other initial errors
		result := model.NewFailurePageActionResult(err.Error())
		return &result, nil // Return nil error, as the failure is captured in the result struct
	}

	// Verify ownership and state
	if mp.ParentContext == nil || mp.ParentContext.ID != contextID || mp.ParentContext.ParentBrowser.BoxID != boxID {
		errMsg := fmt.Sprintf("page %s does not belong to context %s or box %s", pageID, contextID, boxID)
		result := model.NewFailurePageActionResult(errMsg)
		return &result, nil
	}
	targetPage := mp.Instance
	if targetPage.IsClosed() {
		s.handlePageClose(mp)
		result := model.NewFailurePageActionResult(ErrPageNotFound.Error()) // Use defined error
		return &result, nil
	}
	if !mp.ParentContext.Instance.Browser().IsConnected() {
		errMsg := "browser is disconnected"
		result := model.NewFailurePageActionResult(errMsg)
		return &result, nil
	}

	var actionResultData interface{} // Holds data returned by sub-actions (e.g., screenshot base64)
	var actionErr error

	// --- TODO: Implement Pre-Action Screenshot (if needed) ---
	// var beforeScreenshotStr *string = nil
	// beforeBytes, beforeErr := targetPage.Screenshot()
	// if beforeErr == nil {
	// 	 str := base64.StdEncoding.EncodeToString(beforeBytes)
	// 	 beforeScreenshotStr = &str
	// } // else log error?

	// Dispatch based on action prefix
	actionStr := string(params.Action)
	switch {
	case strings.HasPrefix(actionStr, "vision."):
		actionResultData, actionErr = s.executeVisionAction(targetPage, params.Action, params.Params)
	case strings.HasPrefix(actionStr, "snapshot."):
		actionResultData, actionErr = s.executeSnapshotAction(targetPage, params.Action, params.Params)
	default:
		actionErr = fmt.Errorf("unsupported or unrecognized action prefix for action: %s", params.Action)
	}

	// --- Result Handling ---
	if actionErr != nil {
		// Action failed, return FailureResult
		wrappedErrMsg := fmt.Sprintf("action '%s' failed: %s", params.Action, actionErr.Error())
		result := model.NewFailurePageActionResult(wrappedErrMsg)
		return &result, nil // Error is handled within the result
	}

	// Action succeeded, return SuccessResult
	// --- TODO: Get current tab state ---
	// This needs a function to get all TabMetadata for the current context/browser
	currentPages := s.getCurrentPages(mp.ParentContext) // Renamed function call

	// Handle specific result data (e.g., placing screenshot data)
	var afterScreenshotStr *string = nil
	if base64Str, ok := actionResultData.(string); ok && params.Action == model.ActionVisionScreenshot {
		// Only use result data if it's a string (expected base64) AND from screenshot action
		afterScreenshotStr = &base64Str
	}
	// --- TODO: Optionally take an 'after' screenshot even if action wasn't screenshot ---
	// if afterScreenshotStr == nil { // If not already taken by the action itself
	//    afterBytes, afterErr := targetPage.Screenshot(...)
	//    if afterErr == nil { ... set afterScreenshotStr ... }
	// }

	result := model.NewSuccessPageActionResult(afterScreenshotStr, currentPages)
	return &result, nil // Action succeeded
}

// Generic parameter helper functions are removed as parameters are now handled by specific structs.
