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

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"

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
	readers  *pgxpool.Pool
	writer   *pgxpool.Pool // a pool for auto-reconnect on Acquire, but should only be one actual writer
	reserved *pgxpool.Pool // reserved readers so validator operations (block proposal) aren't blocked by RPC readers
	// The above is a reserved/superuser connection pool in a manner similar to
	// how postgres itself reserves connections with the reserved_connections
	// and superuser_reserved_connections system settings.
	// https://www.postgresql.org/docs/current/runtime-config-connection.html#GUC-RESERVED-CONNECTIONS
}

// PoolConfig combines a connection config with additional options for a pool of
// read connections and a single write connection, as required for kwild.MaxConns
type PoolConfig struct {
	ConnConfig

	// MaxConns is the maximum number of allowable connections in the read pool.
	// This does not include the reserved connections for the consensus thread,
	// one of which is a writer.
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
		level := log.InfoLevel
		if n.Code == "42710" || strings.HasPrefix(n.Code, "42P") || // duplicate something ignored: https://www.postgresql.org/docs/16/errcodes-appendix.html
			strings.HasPrefix(n.SchemaName, "ds_") { // user query error
			level = log.DebugLevel
		}
		if n.Detail == "" {
			logger.Logf(level, "%v [%v]: %v", n.Severity, n.Code, n.Message)
		} else {
			logger.Logf(level, "%v [%v]: %v / %v", n.Severity, n.Code, n.Message, n.Detail)
		}
	}
	defaultOnPgError := pCfg.ConnConfig.OnPgError
	pCfg.ConnConfig.OnPgError = func(c *pgconn.PgConn, n *pgconn.PgError) bool {
		level := log.WarnLevel
		if strings.HasPrefix(n.SchemaName, "ds_") { // user query error
			level = log.DebugLevel
		}
		switch sev := strings.ToUpper(n.Severity); sev {
		case "FATAL", "PANIC":
			level = log.ErrorLevel
		} // otherwise it would be "ERROR"
		if n.Detail == "" {
			logger.Logf(level, "%v [%v]: %v", n.Severity, n.Code, n.Message)
		} else {
			logger.Logf(level, "%v [%v]: %v / %v", n.Severity, n.Code, n.Message, n.Detail)
		}
		return defaultOnPgError(c, n) // automatically close any fatal errors (default we are overridding)
	}

	pCfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return registerTypes(ctx, conn)
	}

	db, err := pgxpool.NewWithConfig(ctx, pCfg)
	if err != nil {
		return nil, err
	}

	writerCfg := pCfg.Copy()
	writerCfg.MaxConns = 2 // just one should be fine, but keep a pair for faster reconnect if it needs reconnect
	writer, err := pgxpool.NewWithConfig(ctx, writerCfg)
	if err != nil {
		return nil, err
	}

	reservedCfg := pCfg.Copy()
	reservedCfg.MaxConns = 2 // just one should be fine, but keep a pair for faster reconnect if it needs reconnect
	reserved, err := pgxpool.NewWithConfig(ctx, reservedCfg)
	if err != nil {
		return nil, err
	}

	pool := &Pool{
		readers:  db,
		writer:   writer,
		reserved: reserved,
	}

	return pool, db.Ping(ctx)
}

func registerTypes(ctx context.Context, conn *pgx.Conn) error {
	err := ensureUint256Domain(ctx, conn)
	if err != nil {
		return err
	}

	// PostgreSQL "domains" will use codec of the underlying type, but the
	// dynamic OID of the custom domain needs to be registered with pgx.

	pt, err := conn.LoadType(ctx, "uint256")
	if err != nil {
		return err
	}

	conn.TypeMap().RegisterType(pt)

	pt, err = conn.LoadType(ctx, "uint256[]")
	if err != nil {
		return err
	}

	conn.TypeMap().RegisterType(pt)
	return nil
}

// Query performs a read-only query using the read connection pool. It is
// executed in a transaction with read only access mode to ensure there can be
// no modifications.
func (p *Pool) Query(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return queryTx(ctx, p.readers, stmt, args...)
}

// WARNING: The Execute method is for completeness and helping tests, but is not
// intended to be used with the DB type, which performs all such operations via
// the Tx returned from BeginTx.

// Execute performs a read-write query on the writer connection.
func (p *Pool) Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	var res *sql.ResultSet
	err := p.writer.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		var err error
		res, err = query(ctx, &cqWrapper{c.Conn()}, stmt, args...)
		return err
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (p *Pool) Close() error {
	p.readers.Close()
	p.reserved.Close()
	p.writer.Close()
	return nil
}

// BeginTx starts a read-write transaction. It is an error to call this twice
// without first closing the initial transaction.
func (p *Pool) BeginTx(ctx context.Context) (sql.Tx, error) {
	tx, err := p.writer.BeginTx(ctx, pgx.TxOptions{
		AccessMode: pgx.ReadWrite,
		IsoLevel:   pgx.ReadCommitted,
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
	tx, err := p.readers.BeginTx(ctx, pgx.TxOptions{
		AccessMode: pgx.ReadOnly,
		IsoLevel:   pgx.RepeatableRead,
	})
	if err != nil {
		return nil, err
	}
	return &nestedTx{
		Tx:         tx,
		accessMode: sql.ReadOnly,
	}, nil
}
