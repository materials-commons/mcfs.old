package inprogress

import (
	"sync"
)

type Tracker struct {
	tracking map[string]bool
	mutex    sync.RWMutex
}

var gTracker = NewTracker()

func NewTracker() *Tracker {
	return &Tracker{
		tracking: make(map[string]bool),
	}
}

func (t *Tracker) Mark(id string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.tracking[id] = true
}

func (t *Tracker) Unmark(id string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.tracking, id)
}

func (t *Tracker) Is(id string) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.tracking[id]
}

func Mark(id string) {
	gTracker.Mark(id)
}

func Unmark(id string) {
	gTracker.Unmark(id)
}

func Is(id string) bool {
	return gTracker.Is(id)
}
