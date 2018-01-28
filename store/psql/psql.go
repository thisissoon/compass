package psql

// postgresql driver
import (
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

// Store embeds the various other stores for a single store interface
type Store struct {
	*ServiceStore
	*DentryStore
}

// New connects to a new Store
func New(db *sqlx.DB) *Store {
	return &Store{
		NewServiceStore(db),
		NewDentryStore(db),
	}
}

// A RowHandler handles struct scanning a sqlx.Row into a destination inerface
type RowHandler interface {
	Scan(*sqlx.Row) error
}

// The RowHandlerFunc type is an adapter to allow the use of ordinary
// functions as row handlers
type RowHandlerFunc func(*sqlx.Row) error

// Scan implements the RowHandler interface
func (fn RowHandlerFunc) Scan(row *sqlx.Row) error {
	return fn(row)
}

// QueryRowx wraps sqlx.QueryRowx, executing the query
// and passing the resulting row to a RowHandler
func QueryRowx(db sqlx.Queryer, handler RowHandler, qry string, args ...interface{}) error {
	return handler.Scan(db.QueryRowx(qry, args...))
}
