package service

import (
	"fmt"

	model "github.com/babelcloud/gbox/packages/api-server/pkg/browser"
)

// ExecuteAction executes an action on a specific managed page.
func (s *BrowserService) ExecuteAction(boxID, contextID, pageID string, params model.PageActionParams) (*model.PageActionResult, error) {
	mp, err := s.findManagedPage(pageID)
	if err != nil {
		return nil, err // Includes ErrPageNotFound
	}

	// Verify ownership and state
	if mp.ParentContext == nil || mp.ParentContext.ID != contextID || mp.ParentContext.ParentBrowser.BoxID != boxID {
		return nil, fmt.Errorf("page %s does not belong to context %s or box %s", pageID, contextID, boxID)
	}
	targetPage := mp.Instance
	if targetPage.IsClosed() {
		// Should have been removed by listener, but double-check
		s.handlePageClose(mp) // Attempt cleanup if somehow missed
		return nil, ErrPageNotFound
	}
	if !mp.ParentContext.Instance.Browser().IsConnected() {
		return nil, fmt.Errorf("browser is disconnected")
	}

	var result interface{}
	var actionErr error

	getStringParam := func(key string) (string, error) {
		val, ok := params.Params[key]
		if !ok {
			return "", fmt.Errorf("missing parameter: %s", key)
		}
		strVal, ok := val.(string)
		if !ok {
			return "", fmt.Errorf("parameter '%s' is not a string", key)
		}
		return strVal, nil
	}

	switch params.Action {
	case model.ActionClick:
		selector, err := getStringParam("selector")
		if err != nil {
			actionErr = err
			break
		}
		// Use Locator API instead of deprecated Page method
		// TODO: Expose LocatorClickOptions (button, count, delay, timeout, etc.) via params.Params map
		actionErr = targetPage.Locator(selector).Click()
	case model.ActionFill:
		selector, err := getStringParam("selector")
		if err != nil {
			actionErr = err
			break
		}
		value, err := getStringParam("value")
		if err != nil {
			actionErr = err
			break
		}
		// Use Locator API instead of deprecated Page method
		// TODO: Expose LocatorFillOptions (force, noWaitAfter, timeout)
		actionErr = targetPage.Locator(selector).Fill(value)
	case model.ActionGetText:
		selector, err := getStringParam("selector")
		if err != nil {
			actionErr = err
			break
		}
		// Use Locator API instead of deprecated Page method
		// TODO: Expose LocatorTextContentOptions (timeout)
		textContent, err := targetPage.Locator(selector).TextContent()
		if err == nil {
			result = textContent
		}
		actionErr = err
	default:
		actionErr = fmt.Errorf("unsupported action: %s", params.Action)
	}

	if actionErr != nil {
		return nil, fmt.Errorf("action '%s' failed on page %s: %w", params.Action, pageID, actionErr)
	}

	actionResult := model.NewPageActionResult(result)
	return &actionResult, nil
}
