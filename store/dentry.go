package store

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type DentryPutter interface {
	PutDentry(*Dentry) (*Dentry, error)
}

type DentryStore interface {
	DentryPutter
}

type Dentry struct {
	Id          uuid.UUID `db:"id"`
	CreateDate  time.Time `db:"create_date"`
	UpdateDate  time.Time `db:"update_date"`
	Dtab        string    `db:"dtab"`
	Prefix      string    `db:"prefix"`
	Destination string    `db:"destination"`
	Priority    int32     `db:"priority"`
	ServiceId   string    `db:"service_id"`
}
