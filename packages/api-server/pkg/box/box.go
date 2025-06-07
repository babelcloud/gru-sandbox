package model

import (
	"time"
)

// Box represents a sandbox box
type Box struct {
	ID          string            `json:"id"`
	Status      string            `json:"status"`
	Image       string            `json:"image"`
	ExtraLabels map[string]string `json:"extra_labels,omitempty"`
	// This field is a union of [LinuxBoxConfig], [AndroidBoxConfig]
	Config    LinuxAndroidBoxConfig `json:"config"`
	CreatedAt time.Time             `json:"createdAt"`
	ExpiresAt time.Time             `json:"expiresAt"`
	Type      string                `json:"type"`
	UpdatedAt time.Time             `json:"updatedAt"`
}

type LinuxAndroidBoxConfig struct {
	// This field is a union of [LinuxBoxConfigBrowser], [AndroidBoxConfigBrowser]
	Browser LinuxAndroidBoxConfigBrowser `json:"browser"`
	CPU     float64                      `json:"cpu"`
	Envs    map[string]string            `json:"envs"`
	Labels  map[string]string            `json:"labels"`
	Memory  float64                      `json:"memory"`
	// This field is a union of [LinuxBoxConfigOs], [AndroidBoxConfigOs]
	Os         LinuxAndroidBoxConfigOs `json:"os"`
	Storage    float64                 `json:"storage"`
	WorkingDir string                  `json:"workingDir"`
}

type LinuxAndroidBoxConfigBrowser struct {
	Type    string `json:"type"`
	Version string `json:"version"`
}

type LinuxAndroidBoxConfigOs struct {
	Version string `json:"version"`
}

type BoxFile struct {
	LastModified time.Time `json:"lastModified"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Size         string    `json:"size"`
	Type         string    `json:"type"`
}

type BoxFileListParams struct {
	// Path to the directory
	Path string `json:"-"`
	// Depth of the directory
	Depth float64 `json:"-"`
}

type BoxFileListResult struct {
	Data []BoxFile `json:"data"`
}

type BoxFileReadParams struct {
	Path string `json:"-"`
}

type BoxFileReadResult struct {
	Content string `json:"content"`
}

type BoxFileWriteParams struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type BoxFileWriteResult struct {
	Message string `json:"message"`
}
