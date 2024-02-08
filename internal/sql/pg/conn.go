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

	"github.com/kwilteam/kwil-db/internal/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	pCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	// NOTE: we can consider changing the default exec mode at construction e.g.:
	// pCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	pCfg.ConnConfig.OnNotice = func(_ *pgconn.PgConn, n *pgconn.Notice) {
		logger.Infof("%v [%v]: %v / %v", n.Severity, n.Code, n.Message, n.Detail)
	}
	defaultOnPgError := pCfg.ConnConfig.OnPgError
	pCfg.ConnConfig.OnPgError = func(c *pgconn.PgConn, n *pgconn.PgError) bool {
		logger.Warnf("%v [%v]: %v / %v", n.Severity, n.Code, n.Message, n.Detail)
		return defaultOnPgError(c, n) // automatically close any fatal errors (default we are overridding)
	}
	db, err := pgxpool.NewWithConfig(ctx, pCfg)
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

// WARNING: The Execute method is for completeness and helping tests, but is not
// intended to be used with the DB type, which performs all such operations via
// the Tx returned from BeginTx.

func (p *Pool) Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return query(ctx, &cqWrapper{p.writer}, stmt, args...)
}

func (p *Pool) Get(ctx context.Context, kvTable string, key []byte, pending bool) ([]byte, error) {
	q := p.Query
	if pending {
		q = p.Execute
	}
	return Get(ctx, kvTable, key, q)
}

func (p *Pool) Set(ctx context.Context, kvTable string, key, value []byte) error {
	return Set(ctx, kvTable, key, value, WrapQueryFun(p.Execute))
}

func (p *Pool) Delete(ctx context.Context, kvTable string, key []byte) error {
	return Delete(ctx, kvTable, key, WrapQueryFun(p.Execute))
}

type QueryFn func(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error)

// Get is a plain function for kv table access using any QueryFn, such as a
// method from a DB or Pool, dealers choice.
func Get(ctx context.Context, kvTable string, key []byte, q QueryFn) ([]byte, error) {
	stmt := fmt.Sprintf(selectKvStmtTmpl, kvTable)
	res, err := q(ctx, stmt, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
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

// ExecFn is a function that executes a query without returning any results.
// Although most other methods require a trivial adaptor to satisfy this, it is
// easier to implement in general, such as from third party methods that cannot
// trivially create a ResultSet. See WrapQueryFun.
type ExecFn func(ctx context.Context, stmt string, args ...any) error

// WrapQueryFun adapts a QueryFn to an ExecFn for functions that require no
// results.
func WrapQueryFun(q QueryFn) ExecFn {
	return func(ctx context.Context, stmt string, args ...any) error {
		_, err := q(ctx, stmt, args...)
		return err
	}
}

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

// Execute is now identical to Query. We should consider removing Query as a
// transaction method since their is no semantic or syntactic difference
// (transactions generated from DB or Pool use the write connection).
func (ptx *poolTx) Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	// This method is now identical to Query, but we previously used pgx.Tx.Exec
	// 	res,_ := ptx.Tx.Exec(ctx, stmt, args...)
	// 	ptx.RowsAffected += res.RowsAffected()
	return query(ctx, ptx.Tx, stmt, args...)
}

func (ptx *poolTx) Query(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return query(ctx, ptx.Tx, stmt, args...)
}

func (ptx *poolTx) Precommit(context.Context) ([]byte, error) {
	return nil, errors.New("prepared transactions are not supported in this context")
}

// Begin starts a read-write transaction on the writer connection.
func (p *Pool) Begin(ctx context.Context) (sql.TxCloser, error) {
	tx, err := p.writer.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return nil, err
	}
	return &poolTx{tx, 0}, nil
}

// BeginTx starts a read-write transaction.
func (p *Pool) BeginTx(ctx context.Context) (sql.Tx, error) {
	tx, err := p.writer.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return nil, err
	}
	return &nestedTx{
		Tx:         tx,
		accessMode: sql.ReadWrite,
	}, nil
}

// BeginReadTx starts a read-only transaction.
func (p *Pool) BeginReadTx(ctx context.Context) (sql.Tx, error) {
	tx, err := p.pgxp.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &nestedTx{
		Tx:         tx,
		accessMode: sql.ReadOnly,
	}, nil
}
