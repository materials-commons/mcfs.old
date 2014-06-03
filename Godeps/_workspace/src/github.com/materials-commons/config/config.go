package config

import (
	"github.com/materials-commons/config/cfg"
	"time"
)

// A Configer is a configuration object that can store and retrieve key/value pairs.
type Configer interface {
	cfg.Initer
	cfg.Getter
	cfg.TypeGetterErr
	cfg.TypeGetterDefault
	cfg.Setter
	SetHandler(handler cfg.Handler)
	SetHandlerInit(handler cfg.Handler) error
}

// config is a private type for storing configuration information.
type config struct {
	handler   cfg.Handler
	lastError error         // Last error see on get
	efunc     cfg.ErrorFunc // Error function to call see SetErrorHandler
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
	value, err := c.handler.Get(key, args...)
	c.lastError = err
	return value, err
}

// GetIntErr returns an integer value for a key. See TypeGetter interface for
// error codes.
func (c *config) GetIntErr(key string, args ...interface{}) (int, error) {
	val, err := c.Get(key, args...)
	if err != nil {
		return 0, err
	}
	return cfg.ToInt(val)
}

// GetStringErr returns an string value for a key. See TypeGetter interface for
// error codes.
func (c *config) GetStringErr(key string, args ...interface{}) (string, error) {
	val, err := c.Get(key, args...)
	if err != nil {
		return "", err
	}
	return cfg.ToString(val)
}

// GetTimeErr returns an time.Time value for a key. See TypeGetter interface for
// error codes.
func (c *config) GetTimeErr(key string, args ...interface{}) (time.Time, error) {
	val, err := c.Get(key, args...)
	if err != nil {
		return time.Time{}, err
	}
	return cfg.ToTime(val)
}

// GetBoolErr returns an bool value for a key. See TypeGetter interface for
// error codes.
func (c *config) GetBoolErr(key string, args ...interface{}) (bool, error) {
	val, err := c.Get(key, args...)
	if err != nil {
		return false, err
	}
	return cfg.ToBool(val)
}

// GetInt gets an integer key. It returns the default value of 0 if
// there is an error. GetLastError can be called to see the error.
// If a function is set with SetErrorHandler then the function will
// be called when an error occurs.
func (c *config) GetInt(key string, args ...interface{}) int {
	val, err := c.GetIntErr(key, args...)
	if err != nil {
		c.errorHandler(key, err, args...)
	}
	return val
}

// GetString gets an integer key. It returns the default value of "" if
// there is an error. GetLastError can be called to see the error.
// If a function is set with SetErrorHandler then the function will
// be called when an error occurs.
func (c *config) GetString(key string, args ...interface{}) string {
	val, err := c.GetStringErr(key, args...)
	if err != nil {
		c.errorHandler(key, err, args...)
	}
	return val
}

// GetTime gets an integer key. It returns the default value of an
// empty time.Time if there is an error. GetLastError can be
// called to see the error. If a function is set with
// SetErrorHandler then the function will be called when an error occurs.
func (c *config) GetTime(key string, args ...interface{}) time.Time {
	val, err := c.GetTimeErr(key, args...)
	if err != nil {
		c.errorHandler(key, err, args...)
	}
	return val
}

// GetBool gets an integer key. It returns the default value of false if
// there is an error. GetLastError can be called to see the error. if a
// function is set with SetErrorHandler then the function will be called
// when an error occurs.
func (c *config) GetBool(key string, args ...interface{}) bool {
	val, err := c.GetBoolErr(key, args...)
	if err != nil {
		c.errorHandler(key, err, args...)
	}
	return val
}

// GetLastError returns any error that occured when GetInt, GetString,
// GetBool, or GetTime are called. It will return nil if there was
// no error.
func (c *config) GetLastError() error {
	return c.lastError
}

// SetErrorHandler sets a function to call when GetInt, GetString,
// GetBool, or GetTime return an error. You can use this function
// to handle error in an application specific way. For example if
// an error is fatal you can have this function call os.Exit() or
// panic. Alternatively you can easily log errors with this.
func (c *config) SetErrorHandler(f cfg.ErrorFunc) {
	c.efunc = f
}

// errorHandler calls the error function set with SetErrorHandler.
func (c *config) errorHandler(key string, err error, args ...interface{}) {
	if c.efunc != nil {
		c.efunc(key, err, args...)
	}
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
