package model

// FilterOperator represents the type of filter operation
type FilterOperator string

const (
	FilterOperatorEquals FilterOperator = "="
)

// Filter represents a single filter condition
type Filter struct {
	Field    string         `json:"field"`    // Field to filter on (id, label, ancestor)
	Operator FilterOperator `json:"operator"` // Operation to perform (only supports =)
	Value    string         `json:"value"`    // Value to compare against
}

// BoxListParams represents a request to list boxes
type BoxListParams struct {
	Filters []Filter `json:"filters,omitempty"` // List of filter conditions
}

// BoxListResult represents a response from listing boxes
type BoxListResult struct {
	Boxes   []Box  `json:"boxes"`             // List of boxes
	Count   int    `json:"count"`             // Total number of boxes
	Message string `json:"message,omitempty"` // Response message
}
