package resource

// ResourceStatus represents the current status of a given resource.
type ResourceStatus struct {
}

// Resource is the interface that all resources must implement. A Resource can
// be started, stopped and queried for access.
type Resource interface {
	Start() error
	Stop() error
	Status() ResourceStatus
}
