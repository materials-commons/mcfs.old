package inprogress

import (
	"sync"
)

// Tracker holds the list of items being tracked as in progress. Access to this object
// is thread safe.
type Tracker struct {
	tracking map[string]bool
	mutex    sync.RWMutex
}

var gTracker = NewTracker()

// NewTracker creates a new instance of a Tracker.
func NewTracker() *Tracker {
	return &Tracker{
		tracking: make(map[string]bool),
	}
}

// Mark marks a particular item as inprogress.
func (t *Tracker) Mark(id string) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	val := t.tracking[id]
	if !val {
		t.tracking[id] = true
	}

	return val
}

// Unmark marks an item as untracked.
func (t *Tracker) Unmark(id string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.tracking, id)
}

// Is returns true if a item is being tracked, false otherwise.
func (t *Tracker) Is(id string) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.tracking[id]
}

// Mark uses the global tracking list. It marks a particular item in inprogress.
func Mark(id string) bool {
	return gTracker.Mark(id)
}

// Unmark uses the global tracking list. It marks an item as untracked.
func Unmark(id string) {
	gTracker.Unmark(id)
}

// Is uses the global tracking list. It returns truck if an item is tracked, and false otherwise.
func Is(id string) bool {
	return gTracker.Is(id)
}
