package store

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type ServicePutter interface {
	PutService(*Service) (*Service, error)
}

type ServiceStore interface {
	ServicePutter
}

type Service struct {
	Id          uuid.UUID
	CreateDate  time.Time
	UpdateDate  time.Time
	LogicalName string
	Namespace   string
	Description string
}
