package model

import (
	"time"
)

// Box represents a sandbox box
type Box struct {
	ID          string            `json:"id"`
	Status      string            `json:"status"`
	Image       string            `json:"image"`
	CreatedAt   time.Time         `json:"created_at"`
	ExtraLabels map[string]string `json:"extra_labels,omitempty"`
}
