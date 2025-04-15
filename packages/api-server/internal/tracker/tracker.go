package tracker

import "time"

// AccessTracker defines the interface for tracking last access times.
type AccessTracker interface {
	Update(id string)
	GetLastAccessed(id string) (time.Time, bool)
	Remove(id string)
}
