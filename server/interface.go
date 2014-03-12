package server

// Status the status of a server
type Status int

const (
	// Running server is still alive and responding to requests
	Running Status = iota

	// Stopping server is shutting down
	Stopping

	// Stopped server is no longer running
	Stopped
)

// Server is the interface all servers implement.
type Server interface {
	// Start starts a server
	Start()

	// Stop stops a server
	Stop()

	// Status returns the servers status
	Status() Status
}
