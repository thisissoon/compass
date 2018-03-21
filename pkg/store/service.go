package store

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type ServicePutter interface {
	PutService(*Service) (*Service, error)
}

type ServiceSelector interface {
	GetByLogicalName(logicalName string) (*Service, error)
}

type ServiceStore interface {
	ServicePutter
	ServiceSelector
}

type Service struct {
	Id          uuid.UUID `db:"id"`
	CreateDate  time.Time `db:"create_date"`
	UpdateDate  time.Time `db:"update_date"`
	LogicalName string    `db:"logical_name"`
	Namespace   string    `db:"namespace"`
	Description string    `db:"description"`
}
