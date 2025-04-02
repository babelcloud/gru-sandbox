package models

// MountType represents the type of mount
type MountType string

const (
	// MountTypeBind represents a bind mount from the host filesystem
	MountTypeBind MountType = "bind"
	// MountTypeVolume represents a Docker volume mount
	MountTypeVolume MountType = "volume"
	// MountTypeTmpfs represents a tmpfs mount
	MountTypeTmpfs MountType = "tmpfs"
)

// Mount represents a mount configuration for a box
type Mount struct {
	Type        MountType `json:"type"`                  // Type of mount (bind, volume, tmpfs)
	Source      string    `json:"source"`                // Source path on host (for bind mounts) or volume name
	Target      string    `json:"target"`                // Target path in container
	ReadOnly    bool      `json:"readOnly,omitempty"`    // Whether the mount is read-only
	Consistency string    `json:"consistency,omitempty"` // Mount consistency (default, cached, delegated)
}