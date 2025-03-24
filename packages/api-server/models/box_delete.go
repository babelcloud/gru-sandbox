package models

// BoxDeleteRequest represents a request to delete a box
type BoxDeleteRequest struct {
	ID    string `json:"id,omitempty"`
	Force bool   `json:"force,omitempty"`
}

// BoxesDeleteRequest represents a request to delete all boxes
type BoxesDeleteRequest struct {
	Force bool `json:"force,omitempty"`
}

// BoxDeleteResponse represents the response from deleting a box
type BoxDeleteResponse struct {
	Message string `json:"message"`
}

// BoxesDeleteResponse represents the response from deleting all boxes
type BoxesDeleteResponse struct {
	Count   int      `json:"count"`
	Message string   `json:"message"`
	IDs     []string `json:"ids,omitempty"`
}
