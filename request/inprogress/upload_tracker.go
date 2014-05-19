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

var tracker = NewTracker()

// NewTracker creates a new instance of a Tracker.
func NewTracker() *Tracker {
	return &Tracker{
		tracking: make(map[string]bool),
	}
}

// Mark marks a particular item as in progress. Mark returns
// the previous value. This means that if Mark returns true,
// the file was already in progress.
func (t *Tracker) Mark(id string) bool {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	val := t.tracking[id]
	if !val {
		t.tracking[id] = true
	}

	return val
}

// Unmark marks an item as untracked.
func (t *Tracker) Unmark(id string) {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	delete(t.tracking, id)
}

// Is returns true if a item is being tracked, false otherwise.
func (t *Tracker) Is(id string) bool {
	defer t.mutex.RUnlock()
	t.mutex.RLock()
	return t.tracking[id]
}

// Mark uses the global tracking list. It marks a particular item as in progress.
// Mark returns the previous value. This means that if Mark returns true, the
// file was already in progress.
func Mark(id string) bool {
	return tracker.Mark(id)
}

// Unmark uses the global tracking list. It marks an item as untracked.
func Unmark(id string) {
	tracker.Unmark(id)
}

// Is uses the global tracking list. It returns truck if an item is tracked, and false otherwise.
func Is(id string) bool {
	return tracker.Is(id)
}
