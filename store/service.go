package store

type ServicePutter interface {
	PutService(*Service) (*Service, error)
}

type Service struct {
	LogicalName string
	Namespace   string
	Description string
}
