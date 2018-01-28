package psql

import (
	"fmt"

	"compass/namerd"
	"compass/store"
	"compass/sync"

	"github.com/jmoiron/sqlx"
)

// Dentry table name
const DentryTableName = "dentry"

// Service upsert query
var UpsertDentryQry = fmt.Sprintf(`
	INSERT INTO "%s" (
		"dtab",
		"prefix",
		"destination",
		"priority")
	VALUES ($1,$2,$3,$4)
	ON CONFLICT ON CONSTRAINT uq_dentry_dtab_prefix DO
	UPDATE SET
		update_date=timezone('UTC'::text, now()),
		dtab=excluded.dtab,
		prefix=excluded.prefix,
		destination=excluded.destination,
		priority=excluded.priority
	RETURNING *;`, DentryTableName)

// DentryStore manages dentry CRUD ops
type DentryStore struct {
	db     *sqlx.DB
	syncer *sync.Sync
}

// PutDentry creates or updates a dentry
func (store *DentryStore) PutDentry(dentry *store.Dentry) (*store.Dentry, error) {
	return upsertDentry(store.db, store.syncer, dentry)
}

// upsertDentry implements the underlying database logic for inserting
// or updating a dentry
func upsertDentry(db sqlx.Queryer, syncr sync.Syncer, dentry *store.Dentry) (*store.Dentry, error) {
	err := QueryRowx(
		db,
		RowHandlerFunc(func(row *sqlx.Row) error {
			return row.StructScan(dentry)
		}),
		UpsertDentryQry,
		dentry.Dtab,
		dentry.Prefix,
		dentry.Destination,
		dentry.Priority)
	if err == nil {
		syncr.Sync(namerd.Dtab(dentry.Dtab))
	}
	return dentry, err
}

func (store *DentryStore) DentryList(dtab string) ([]store.Dentry, error) {
	return nil, nil
}

// NewDentryStore returns a new DentryStore
func NewDentryStore(db *sqlx.DB, syncr *sync.Sync) *DentryStore {
	return &DentryStore{
		db:     db,
		syncer: syncr,
	}
}
