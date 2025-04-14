package tracker

import (
	"sync"
	"time"
)

// InMemoryAccessTracker implements AccessTracker using an in-memory map.
type InMemoryAccessTracker struct {
	mu          sync.RWMutex
	accessTimes map[string]time.Time
}

// NewInMemoryAccessTracker creates a new InMemoryAccessTracker.
func NewInMemoryAccessTracker() *InMemoryAccessTracker {
	return &InMemoryAccessTracker{accessTimes: make(map[string]time.Time)}
}

// Update sets the last access time for the given ID to now.
func (t *InMemoryAccessTracker) Update(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.accessTimes[id] = time.Now()
}

// GetLastAccessed retrieves the last access time for the given ID.
// If the ID is not found (e.g., after a server restart), it initializes
// the access time to now and returns true.
func (t *InMemoryAccessTracker) GetLastAccessed(id string) (time.Time, bool) {
	t.mu.RLock()
	ts, found := t.accessTimes[id]
	t.mu.RUnlock()

	if !found {
		// If not found, acquire write lock to potentially add it
		t.mu.Lock()
		// Double-check if another goroutine added it while waiting for the lock
		ts, found = t.accessTimes[id]
		if !found {
			// If still not found, add it with the current time
			now := time.Now()
			t.accessTimes[id] = now
			ts = now
			found = true // Treat this as the first access
		}
		t.mu.Unlock()
	}
	return ts, found
}

// Remove deletes the tracking information for the given ID.
func (t *InMemoryAccessTracker) Remove(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.accessTimes, id)
}
