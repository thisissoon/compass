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
	Id          uuid.UUID `db:"id"`
	CreateDate  time.Time `db:"create_date"`
	UpdateDate  time.Time `db:"update_date"`
	LogicalName string    `db:"logical_name"`
	Dtab        string    `db:"dtab"`
	Namespace   string    `db:"namespace"`
	Description string    `db:"description"`
}
