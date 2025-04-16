package model

import "encoding/json"

// PageActionParams represents the request parameters for executing a page action.
type PageActionParams struct {
	Action PageActionType  `json:"action"` // The specific action type (defined in page_action_types.go)
	Params json.RawMessage `json:"params"` // Parameters specific to the action, parsed based on Action type
}

// NOTE: PageActionResult is now defined in page_action_result.go
