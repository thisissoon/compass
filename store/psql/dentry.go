package psql

import (
	"fmt"

	"compass/logger"
	"compass/namerd"
	"compass/store"

	"github.com/jmoiron/sqlx"
)

// Dentry table name
const DentryTableName = "dentry"

// Service upsert query
var UpsertDentryQry = fmt.Sprintf(`
	INSERT INTO public.%s (
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

var DentryListByDtab = fmt.Sprintf(`
	SELECT *
	FROM public.%s
	WHERE dtab = $1
	ORDER BY priority ASC`, DentryTableName)

// DentryStore manages dentry CRUD ops
type DentryStore struct {
	db           *sqlx.DB
	dtabUpdatesC chan namerd.Dtab
}

// TODO: could be a subscription style model
func (store *DentryStore) DtabUpdates() <-chan namerd.Dtab {
	return (<-chan namerd.Dtab)(store.dtabUpdatesC)
}

// PutDentry creates or updates a dentry
func (store *DentryStore) PutDentry(dentry *store.Dentry) (*store.Dentry, error) {
	return upsertDentry(store.db, store.dtabUpdatesC, dentry)
}

// upsertDentry implements the underlying database logic for inserting
// or updating a dentry
func upsertDentry(db sqlx.Queryer, C chan namerd.Dtab, dentry *store.Dentry) (*store.Dentry, error) {
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
	if err == nil && C != nil {
		C <- namerd.Dtab(dentry.Dtab)
	}
	return dentry, err
}

// DentriesByDtab returns a slice of Dentry for the Delegation table
func (store *DentryStore) DentriesByDtab(dtab string) (<-chan *store.Dentry, error) {
	return dentriesByDtab(store.db, dtab)
}

// dentriesByDtab
func dentriesByDtab(db sqlx.Queryer, dtab string) (<-chan *store.Dentry, error) {
	log := logger.New()
	rows, err := db.Queryx(DentryListByDtab, dtab)
	if err != nil {
		return nil, err
	}
	C := make(chan *store.Dentry, 1)
	go func() {
		defer close(C)
		defer rows.Close()
		for rows.Next() {
			var dentry store.Dentry
			if err := rows.StructScan(&dentry); err != nil {
				log.Error().Err(err).Msg("failed to scan row")
			}
			C <- &dentry
		}
	}()
	return (<-chan *store.Dentry)(C), nil
}

// NewDentryStore returns a new DentryStore
func NewDentryStore(db *sqlx.DB) *DentryStore {
	return &DentryStore{
		db:           db,
		dtabUpdatesC: make(chan namerd.Dtab, 1),
	}
}
