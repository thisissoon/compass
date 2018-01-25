package store

type ServicePutter interface {
	PutService(*Service) (*Service, error)
}

type ServiceStore interface {
	ServicePutter
}

type Service struct {
	LogicalName string
	Namespace   string
	Description string
}
