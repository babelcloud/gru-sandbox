package models

// Box represents a box entity
type Box struct {
	ID          string            `json:"id"`
	Status      string            `json:"status"`
	Image       string            `json:"image"`
	ExtraLabels map[string]string `json:"labels,omitempty"`
}
