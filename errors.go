package mcfs

import (
	"fmt"
)

// ErrServerNotRunning Attempt to access a server that is not running.
var ErrServerNotRunning = fmt.Errorf("server not running")

// ErrDBLookupFailed Failed to lookup item in database.
var ErrDBLookupFailed = fmt.Errorf("lookup failed")

// ErrDBUpdateFailed Failed to update item in database.
var ErrDBUpdateFailed = fmt.Errorf("update failed")

// ErrDBInsertFailed Failed to Insert item in database.
var ErrDBInsertFailed = fmt.Errorf("insert failed")

// ErrDBRelatedUpdateFailed Failed to update related item(s) in database.
var ErrDBRelatedUpdateFailed = fmt.Errorf("related item update failed")
