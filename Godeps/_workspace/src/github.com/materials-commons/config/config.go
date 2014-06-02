package config

import (
	"github.com/materials-commons/config/cfg"
	"time"
)

// A Configer is a configuration object that can store and retrieve key/value pairs.
type Configer interface {
	cfg.Initer
	cfg.Getter
	cfg.TypeGetter
	cfg.Setter
	SetHandler(handler cfg.Handler)
	SetHandlerInit(handler cfg.Handler) error
}

// config is a private type for storing configuration information.
type config struct {
	handler cfg.Handler
}

// New creates a new Configer instance that uses the specified Handler for
// key/value retrieval and storage.
func New(handler cfg.Handler) Configer {
	return &config{handler: handler}
}

// Init initializes the Configer. It should be called before retrieving
// or setting keys.
func (c *config) Init() error {
	return c.handler.Init()
}

// Get returns the value for a key. It can return any value type.
func (c *config) Get(key string, args ...interface{}) (interface{}, error) {
	return c.handler.Get(key, args...)
}

// GetInt returns an integer value for a key. See TypeGetter interface for
// error codes.
func (c *config) GetInt(key string, args ...interface{}) (int, error) {
	val, err := c.Get(key, args...)
	if err != nil {
		return 0, err
	}
	return cfg.ToInt(val)
}

// GetString returns an string value for a key. See TypeGetter interface for
// error codes.
func (c *config) GetString(key string, args ...interface{}) (string, error) {
	val, err := c.Get(key, args...)
	if err != nil {
		return "", err
	}
	return cfg.ToString(val)
}

// GetTime returns an time.Time value for a key. See TypeGetter interface for
// error codes.
func (c *config) GetTime(key string, args ...interface{}) (time.Time, error) {
	val, err := c.Get(key, args...)
	if err != nil {
		return time.Time{}, err
	}
	return cfg.ToTime(val)
}

// GetBool returns an bool value for a key. See TypeGetter interface for
// error codes.
func (c *config) GetBool(key string, args ...interface{}) (bool, error) {
	val, err := c.Get(key, args...)
	if err != nil {
		return false, err
	}
	return cfg.ToBool(val)
}

// SetHandler changes the handler for a Configer. If this method is called
// then you must call Init before accessing any of the keys.
func (c *config) SetHandler(handler cfg.Handler) {
	c.handler = handler
}

// SetHandlerInit changes the handler for a Configer. It also immediately calls
// Init and returns the error from this call.
func (c *config) SetHandlerInit(handler cfg.Handler) error {
	c.handler = handler
	return c.Init()
}

// Set sets key to value. See Setter interface for error codes.
func (c *config) Set(key string, value interface{}, args ...interface{}) error {
	return c.handler.Set(key, value, args...)
}
