package store

import uuid "github.com/satori/go.uuid"

type ServicePutter interface {
	PutService(*Service) (*Service, error)
}

type ServiceStore interface {
	ServicePutter
}

type Service struct {
	Id          uuid.UUID
	LogicalName string
	Namespace   string
	Description string
}
