package service

type ServiceStatus struct {
	
}

type Service interface {
	Start() error
	Stop() error
	Status() ServiceStatus
}
