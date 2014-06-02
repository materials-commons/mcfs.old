package handler

import (
	"github.com/materials-commons/config/cfg"
	"sync"
)

// syncHandler holds all the attributes needed to provide
// safe, synchronized access to a handler.
type syncHandler struct {
	handler cfg.Handler
	loaded  bool
	mutex   sync.Mutex
}

// Sync creates a Handler that can be safely accessed by multiple threads. It
// ensures that the Init method only initializes a handler one time, regardless
// of the number of threads that call it.
func Sync(handler cfg.Handler) cfg.Handler {
	return &syncHandler{handler: handler}
}

// Init safely initializes the handler once. If Init has already been successfully called
// additional calls to it don't do anything.
func (h *syncHandler) Init() error {
	defer h.mutex.Unlock()
	h.mutex.Lock()

	switch {
	case h.loaded:
		return nil
	default:
		if err := h.handler.Init(); err != nil {
			return err
		}
	}

	h.loaded = true
	return nil
}

// Get provides synchronized access to key retrieval.
func (h *syncHandler) Get(key string, args ...interface{}) (interface{}, error) {
	defer h.mutex.Unlock()
	h.mutex.Lock()
	return h.handler.Get(key, args...)
}

// Set provides synchronized access to setting a key.
func (h *syncHandler) Set(key string, value interface{}, args ...interface{}) error {
	defer h.mutex.Unlock()
	h.mutex.Lock()
	return h.handler.Set(key, value, args...)
}

// Args returns true if the handler takes additional arguments.
func (h *syncHandler) Args() bool {
	return h.handler.Args()
}
