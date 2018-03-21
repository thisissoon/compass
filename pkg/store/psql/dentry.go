package psql

import (
	"context"
	"fmt"

	"compass/pkg/namerd"
	"compass/pkg/store"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
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

var SelectDentryById = fmt.Sprintf(`
	SELECT *
	FROM public.%s
	WHERE id=$1`, DentryTableName)

var DeleteDentryByIdQry = fmt.Sprintf(`
	DELETE
	FROM public.%s
	WHERE id=$1`, DentryTableName)

var DeleteDentryByPrefixQry = fmt.Sprintf(`
	DELETE
	FROM public.%s
	WHERE
		dtab=$1
		AND prefix=$2`, DentryTableName)

// DentryStore manages dentry CRUD ops
type DentryStore struct {
	db           *sqlx.DB
	dtabUpdatesC chan namerd.Dtab
}

// TODO: could be a subscription style model
func (store *DentryStore) DtabUpdates() <-chan namerd.Dtab {
	return (<-chan namerd.Dtab)(store.dtabUpdatesC)
}

// DelegationTables returns a list of distinct delgation tables
func (store *DentryStore) DelegationTables() ([]string, error) {
	var dtabs []string
	qry := fmt.Sprintf("SELECT DISTINCT(dtab) FROM public.%s ORDER BY dtab ASC", DentryTableName)
	rows, err := store.db.Queryx(qry)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var dtab string
		if err := rows.Scan(&dtab); err != nil {
			return nil, err
		}
		dtabs = append(dtabs, dtab)
	}
	return dtabs, nil
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
func (store *DentryStore) DentriesByDtab(ctx context.Context, dtab string) (<-chan *store.Dentry, error) {
	return dentriesByDtab(ctx, store.db, dtab)
}

// dentriesByDtab
func dentriesByDtab(ctx context.Context, db sqlx.Queryer, dtab string) (<-chan *store.Dentry, error) {
	log := zerolog.Ctx(ctx)
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

// DeleteDentryByPrefix deletes a dentry by it's prefix in a specified dtab
func (store *DentryStore) DeleteDentryByPrefix(dtab, prefix string) (int64, error) {
	return deleteDentryByPrefix(store.db, dtab, prefix, store.dtabUpdatesC)
}

// deleteDentryByPrefix executes a quert to delete a dentry within a specified dtab
// by dentry prefix
func deleteDentryByPrefix(db sqlx.Execer, dtab, prefix string, C chan namerd.Dtab) (int64, error) {
	res, err := db.Exec(DeleteDentryByPrefixQry, dtab, prefix)
	if err != nil {
		return 0, err
	}
	if err == nil && C != nil {
		C <- namerd.Dtab(dtab)
	}
	return res.RowsAffected()
}

// DeleteDentryById deletes a dentry by it's UUID
func (store *DentryStore) DeleteDentryById(id uuid.UUID) (int64, error) {
	return deleteDentryById(store.db, id, store.dtabUpdatesC)
}

// deleteDentryById executes a query to delete a dentry by it's UUID
func deleteDentryById(db QueryExecer, id uuid.UUID, C chan namerd.Dtab) (int64, error) {
	var dentry store.Dentry
	row := db.QueryRowx(SelectDentryById, id)
	if err := row.StructScan(&dentry); err != nil {
		return 0, err
	}
	res, err := db.Exec(DeleteDentryByIdQry, id)
	if err != nil {
		return 0, err
	}
	if err == nil && C != nil {
		C <- namerd.Dtab(dentry.Dtab)
	}
	return res.RowsAffected()
}

// NewDentryStore returns a new DentryStore
func NewDentryStore(db *sqlx.DB) *DentryStore {
	return &DentryStore{
		db:           db,
		dtabUpdatesC: make(chan namerd.Dtab, 1),
	}
}
