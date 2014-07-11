package mcfserr

import (
	"fmt"
)

// ErrServerNotRunning Attempt to access a server that is not running.
var ErrServerNotRunning = fmt.Errorf("server not running")

// ErrDBLookupFailed Failed to lookup item in database.
var ErrDB = fmt.Errorf("db err")
