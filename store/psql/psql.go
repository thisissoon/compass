package psql

// postgresql driver
import (
	"net/url"
	"path"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

type QueryExecer interface {
	sqlx.Queryer
	sqlx.Execer
}

// A DSN is a postgresql Data source name used for database connection
type DSN struct {
	Host     string
	Name     string
	Username string
	Password string
	SSLMode  string
}

// String bulds the DSN as a URL returning the constructed URL
func (dsn DSN) String() string {
	v := url.Values{}
	v.Add("sslmode", dsn.SSLMode)
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(dsn.Username, dsn.Password),
		Host:     dsn.Host,
		Path:     path.Join(dsn.Name),
		RawQuery: v.Encode(),
	}
	return u.String()
}

// Store embeds the various other stores for a single store interface
type Store struct {
	*ServiceStore
	*DentryStore
}

// Open opens a new postgrtes connection using the provided DSN
func Open(dsn DSN) (*sqlx.DB, error) {
	return sqlx.Open("postgres", dsn.String())
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
