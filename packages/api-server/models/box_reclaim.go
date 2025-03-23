package models

// BoxReclaimResponse represents the response from reclaiming boxes
type BoxReclaimResponse struct {
	Message      string   `json:"message"`
	StoppedIDs   []string `json:"stoppedIds,omitempty"` // IDs of boxes that were stopped
	DeletedIDs   []string `json:"deletedIds,omitempty"` // IDs of boxes that were deleted
	StoppedCount int      `json:"stoppedCount"`         // Number of boxes stopped
	DeletedCount int      `json:"deletedCount"`         // Number of boxes deleted
}
