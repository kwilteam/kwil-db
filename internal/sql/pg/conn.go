// Package pg defines the primary PostgreSQL-powered DB and Pool types used to
// support Kwil DB.
//
// See the [DB] type for more information.
package pg

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/sql/v2" // temporary v2 for refactoring

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func connString(host, port, user, pass, dbName string, repl bool) string {
	// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
	connStr := fmt.Sprintf("host=%s user=%s database=%s sslmode=disable",
		host, user, dbName)

	if pass != "" {
		connStr += fmt.Sprintf(" password=%s", pass)
	}

	// Only add port for TCP connections, not UNIX domain sockets.
	if !strings.HasPrefix(host, "/") {
		connStr += fmt.Sprintf(" port=%s", port)
	}

	if repl {
		connStr += " replication=database"
	}

	return connStr
}

// ConnConfig groups the basic connection settings used to construct the DSN
// "connection string" used to open a new connection to a postgres host.
// TODO: use this in the various constructors for DB, Pool, etc.
type ConnConfig struct {
	// Host, Port, User, Pass, and DBName are used verbatim to create a
	// connection string in DSN format.
	// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
	Host, Port string
	User, Pass string
	DBName     string
}

// Pool is a simple read connection pool with one dedicated writer connection.
// This type is relatively low level, and Kwil will generally use the DB type to
// manage sessions instead of this type directly. It is exported primarily for
// testing and reuse in more general use cases.
//
// Pool supports Kwil's single transactional DB writer model:
//   - a single writer connection, on which a transaction is created by a top
//     level system during block execution (i.e. the AbciApp),
//     and from which reads of uncommitted DB records may be performed.
//   - multiple readers, which may service other asynchronous operations such as
//     a gRPC user service.
//
// The write methods from the Tx returned from the BeginTx method should be
// preferred over directly using the Pool's write methods. The DB type is the
// session-aware wrapper that creates and stores the write Tx, and provides top
// level Exec/Set methods that error if no Tx exists. Only use Pool as a
// building block or for testing individual systems outside of the context of a
// session.
type Pool struct {
	pgxp   *pgxpool.Pool
	writer *pgx.Conn // hijacked from the pool
}

var _ sql.Queryer = (*Pool)(nil)

// PoolConfig combines a connection config with additional options for a pool of
// read connections and a single write connection, as required for kwild.MaxConns
type PoolConfig struct {
	ConnConfig

	// MaxConns is the maximum number of allowable connections, including the
	// one write connection. Thus there will be MaxConns-1 readers.
	MaxConns uint32
}

// TODO: update connStr with more pool options
//   - pool_max_conns: integer greater than 0
//   - pool_min_conns: integer 0 or greater
//   - pool_max_conn_lifetime: duration string
//   - pool_max_conn_idle_time: duration string
//   - pool_health_check_period: duration string
//   - pool_max_conn_lifetime_jitter: duration string

// NewPool creates a connection pool to a PostgreSQL database.
func NewPool(ctx context.Context, cfg *PoolConfig) (*Pool, error) {
	if cfg.User == "" {
		return nil, errors.New("db user must not be empty")
	}
	if cfg.MaxConns < 2 {
		return nil, errors.New("at least two total connections are required")
	}
	const repl = false
	connStr := connString(cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.DBName, repl)
	connStr += fmt.Sprintf(" pool_max_conns=%d", cfg.MaxConns)

	db, err := pgxpool.New(ctx, connStr) // sql.Open("pgx/v5", connStr)
	if err != nil {
		return nil, err
	}

	writer, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	pool := &Pool{
		pgxp:   db,
		writer: writer.Hijack(),
	}

	return pool, db.Ping(ctx)
}

// Query performs a read-only query using the read connection pool. It is
// executed in a transaction with read only access mode to ensure there can be
// no modifications.
func (p *Pool) Query(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return queryTx(ctx, p.pgxp, stmt, args...)
}

// WARNING: The Execute and QueryPending are for completeness and helping tests,
// but are not intended to be used with the DB type, which performs all such
// operations via the Tx returned from BeginTx.

func (p *Pool) QueryPending(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return query(ctx, p.writer.Query, stmt, args...)
}

func (p *Pool) Execute(ctx context.Context, stmt string, args ...any) error {
	_, err := p.writer.Exec(ctx, stmt, args...)
	return err
}

func (p *Pool) Get(ctx context.Context, kvTable string, key []byte, pending bool) ([]byte, error) {
	q := p.Query
	if pending {
		q = p.QueryPending
	}
	return Get(ctx, kvTable, key, q)
}

func (p *Pool) Set(ctx context.Context, kvTable string, key, value []byte) error {
	return Set(ctx, kvTable, key, value, p.Execute)
}

func (p *Pool) Delete(ctx context.Context, kvTable string, key []byte) error {
	return Delete(ctx, kvTable, key, p.Execute)
}

type QueryFn func(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error)

// Get is a plain function for kv table access using any QueryFn, such as a
// method from a DB or Pool, dealers choice.
func Get(ctx context.Context, kvTable string, key []byte, q QueryFn) ([]byte, error) {
	stmt := fmt.Sprintf(selectKvStmtTmpl, kvTable)
	res, err := q(ctx, stmt, key)
	if err != nil {
		return nil, err
	}

	switch len(res.Rows) {
	case 0: // this is fine
		return nil, nil
	case 1:
	default:
		return nil, errors.New("exactly one row not returned")
	}

	row := res.Rows[0]
	if len(row) != 1 {
		return nil, errors.New("exactly one value not in row")
	}
	val, ok := row[0].([]byte)
	if !ok {
		return nil, errors.New("value not a bytea")
	}
	return val, nil
}

type ExecFn func(ctx context.Context, stmt string, args ...any) error

// Set is a plain function for setting values in a kv using any ExecFn, such as
// the method from a DB or Pool.
func Set(ctx context.Context, kvTable string, key, value []byte, e ExecFn) error {
	stmt := fmt.Sprintf(insertKvStmtTmpl, kvTable)
	return e(ctx, stmt, key, value)
}

// Delete is a plain function for deleting values from a kv table using any
// ExecFn, such as the method from a DB or Pool.
func Delete(ctx context.Context, kvTable string, key []byte, e ExecFn) error {
	stmt := fmt.Sprintf(deleteKvStmtTmpl, kvTable)
	return e(ctx, stmt, key)
}

// CreateKVTable is a helper to create a table compatible with the Get and Set
// functions.
func CreateKVTable(ctx context.Context, tableName string, e ExecFn) error {
	createStmt := fmt.Sprintf(createKvStmtTmpl, tableName)
	return e(ctx, createStmt)
}

func (p *Pool) Close() error {
	p.pgxp.Close()
	return p.writer.Close(context.TODO())
}

type poolTx struct {
	pgx.Tx
	RowsAffected int64 // for debugging and testing
}

func (ptx *poolTx) Execute(ctx context.Context, stmt string, args ...any) error {
	res, err := ptx.Exec(ctx, stmt, args...)
	if err != nil {
		return err
	}
	ptx.RowsAffected += res.RowsAffected()
	return err
}

func (ptx *poolTx) Query(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return query(ctx, ptx.Tx.Query, stmt, args...)
}

func (p *Pool) BeginTx(ctx context.Context) (sql.TxCloser, error) {
	tx, err := p.writer.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return nil, err
	}
	return &poolTx{tx, 0}, nil
}
