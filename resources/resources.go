package resources

type ResourceStatus struct {
}

type Resource interface {
	Start() error
	Stop() error
	Status() ResourceStatus
}
