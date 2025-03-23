package models

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

// BoxListRequest represents a request to list boxes
type BoxListRequest struct {
	Filters []Filter `json:"filters,omitempty"` // List of filter conditions
}

// BoxListResponse represents the response from listing boxes
type BoxListResponse struct {
	Boxes []Box `json:"boxes"`
}
