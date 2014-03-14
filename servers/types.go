package servers

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

func (s Status) String() string {
	switch s {
	case Running:
		return "Running"
	case Stopping:
		return "Stopping"
	case Stopped:
		return "Stopped"
	default:
		return "Unknown"
	}
}
