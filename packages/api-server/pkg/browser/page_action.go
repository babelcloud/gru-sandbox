package model

// PageActionType defines the type of action to execute on a page.
type PageActionType string

const (
	ActionClick   PageActionType = "click"
	ActionFill    PageActionType = "fill"
	ActionGetText PageActionType = "getText"
	// Add other supported actions here
)

// PageActionParams represents the request parameters for executing a page action.
type PageActionParams struct {
	Action PageActionType         `json:"action"`
	Params map[string]interface{} `json:"params"` // Parameters specific to the action, e.g., {"selector": "#id", "value": "text"}
}

// PageActionResult represents the response from executing a page action.
type PageActionResult struct {
	Result interface{} `json:"result,omitempty"` // Result of the action, e.g., text content
}

// NewPageActionResult creates a new PageActionResult.
func NewPageActionResult(result interface{}) PageActionResult {
	return PageActionResult{
		Result: result,
	}
}
