package log

import (
	"fmt"
	"github.com/inconshreveable/log15"
	"os"
)

var (
	// Global log variable.
	L = log15.New()

	// Default handler used in the package.
	defaultHandler log15.Handler
)

func init() {
	stdoutHandler := log15.StreamHandler(os.Stdout, log15.LogfmtFormat())
	SetDefaultHandler(log15.LvlFilterHandler(log15.LvlInfo, stdoutHandler))
	L.SetHandler(defaultHandler)
}

// New creates a new instance of the logger using the current default handler
// for its output.
func New(ctx ...interface{}) log15.Logger {
	l := log15.New(ctx...)
	l.SetHandler(defaultHandler)
	return l
}

// Msg is short hand to create a message string using fmt.Sprintf.
func Msg(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

// SetDefaultHandler sets the handler for the logger. It wraps handlers in a SyncHandler. You
// should not pass in handlers that are already wrapped in a SyncHandler.
func SetDefaultHandler(handler log15.Handler) {
	defaultHandler = log15.SyncHandler(handler)
	L.SetHandler(defaultHandler)
}

// DefaultHandler returns the current handler. It can be used to create additional
// logger instances that all use the same handler for output.
func DefaultHandler() log15.Handler {
	return defaultHandler
}
