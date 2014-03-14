package mcfs

import (
	"fmt"
)

// ErrServerNotRunning Attempt to access a server that is not running.
var ErrServerNotRunning = fmt.Errorf("server not running")
