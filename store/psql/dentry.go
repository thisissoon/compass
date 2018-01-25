package psql

import (
	"fmt"

	"compass/store"

	"github.com/jmoiron/sqlx"
)

// Dentry table name
const DentryTableName = "dentry"

// Service upsert query
var UpsertDentryQry = fmt.Sprintf(`
	INSERT INTO "%s" (
		"id",
		"prefix",
		"destination",
		"namespace",
		"priority")
	VALUES ($1,$2,$3,$4)
	ON CONFLICT ON CONSTRAINT dentry_id DO
	UPDATE SET
		update_date=timezone('UTC'::text, now()),
		prefix=excluded.prefix,
		destination=excluded.destination,
		namesapce=excluded.namespace,
		priority=excluded.priority
	RETURNING *;`, DentryTableName)

// DentryStore manages dentry CRUD ops
type DentryStore struct {
	db *sqlx.DB
}

// PutDentry creates or updates a dentry
func (store *DentryStore) PutDentry(dentry *store.Dentry) (*store.Dentry, error) {
	return upsertDentry(store.db, dentry)
}

// upsertDentry implements the underlying database logic for inserting
// or updating a dentry
func upsertDentry(db sqlx.Queryer, dentry *store.Dentry) (*store.Dentry, error) {
	err := QueryRowx(
		db,
		RowHandlerFunc(func(row *sqlx.Row) error {
			return row.StructScan(dentry)
		}),
		UpsertDentryQry,
		dentry.Id,
		dentry.Prefix,
		dentry.Destination,
		dentry.Namespace,
		dentry.Priority)
	return dentry, err
}

// NewDentryStore returns a new DentryStore
func NewDentryStore(db *sqlx.DB) *DentryStore {
	return &DentryStore{
		db: db,
	}
}
