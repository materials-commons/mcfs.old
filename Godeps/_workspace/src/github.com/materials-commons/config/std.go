package config

import (
	"github.com/materials-commons/config/cfg"
	"github.com/materials-commons/config/handler"
	"time"
)

// Store configuration in environment as specified for 12 Factor Applications:
// http://12factor.net/config
var TwelveFactor = handler.Env()

// Store configuration in environment, but allow overrides, either by the
// application setting them internally as defaults, or setting them from
// the command line. See http://12factor.net/config. Overrides are an
// extension to the specification. This handler is thread safe and can
// safely be used across multiple go routines.
var TwelveFactorWithOverride = handler.Multi(handler.Sync(handler.Map()), handler.Env())

var std Configer

// Init initializes the standard Configer using the specified handler. The
// standard configer is a global config that can be conveniently accessed
// from the config package.
func Init(handler cfg.Handler) error {
	std = New(handler)
	return std.Init()
}

// Get gets a key from the standard Configer.
func Get(key string, args ...interface{}) (interface{}, error) {
	return std.Get(key, args...)
}

// GetInt gets an integer key from the standard Configer.
func GetInt(key string, args ...interface{}) (int, error) {
	return std.GetInt(key, args...)
}

// GetString gets an string key from the standard Configer.
func GetString(key string, args ...interface{}) (string, error) {
	return std.GetString(key, args...)
}

// GetTime gets an time key from the standard Configer.
func GetTime(key string, args ...interface{}) (time.Time, error) {
	return std.GetTime(key, args...)
}

// GetBool gets an bool key from the standard Configer.
func GetBool(key string, args ...interface{}) (bool, error) {
	return std.GetBool(key, args...)
}

// Set sets key to value in the standard Configer.
func Set(key string, value interface{}, args ...interface{}) error {
	return std.Set(key, value, args...)
}
