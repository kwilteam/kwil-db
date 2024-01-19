package pg

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/kwilteam/kwil-db/internal/sql/v2" // temporary v2 for refactoring

	"github.com/jackc/pgx/v5"
)

// DB is a session-aware wrapper that creates and stores a write Tx on request,
// and provides top level Exec/Set methods that error if no Tx exists. It also
// implements a "QueryPending" method that uses tx.Query if such a Tx has been
// created and stored for the lifetime of a session. This design prevents any
// out-of-session write statements from executing, and makes uncommitted reads
// explicit (and impossible in the absence of an active transaction).
//
// This type is tailored to use in kwild in the following ways:
//
//  1. Controlled transactional interaction that requires beginning a
//     transaction before using the Exec method, unless put in "autocommit" mode
//     using the AutoCommit method. Use of the write connection when not
//     executing a block's transactions is prevented.
//
//  2. Using an underlying connection pool, with multiple readers and a single
//     write connection to ensure all uses of Execute operate on the active
//     transaction.
//
//  3. Emulating SQLite changesets by collecting WAL data for updates from a
//     dedicated logical replication connection and slot. When called after
//     Commit, the CommitID method returns a digest of the updates in that
//     transaction.
//     NOTE: this may need to switch to lots of triggers on every table...
//
// IMPORTANT: This type must be the exclusive database user. If any other type
// or even external process like psql changes the database, transactions with
// this DB type may fail.
type DB struct {
	// dev note: satisfies Datastore / poolAdapter and registry.DB

	pool *Pool    // raw connection pool
	repl *replMon // logical replication monitor for collecting commit IDs

	// Guarantee that we are in-session by tracking and using a Tx for the write methods.
	mtx        sync.Mutex
	autoCommit bool   // skip the explicit transaction (begin/commit automatically)
	tx         pgx.Tx // interface
	commitID   []byte
}

// DBConfig is the configuration for the Kwil DB backend, which includes the
// connection parameters and a schema filter used to selectively include WAL
// data for certain PostgreSQL schemas in commit ID calculation.
type DBConfig struct {
	PoolConfig

	// SchemaFilter is used to include WAL data for certain *postgresq* schema
	// (not Kwil schema). If nil, the default is to include updates to tables in
	// any schema prefixed by "ds_".
	SchemaFilter func(string) bool
}

var defaultSchemaFilter = func(schema string) bool {
	return strings.HasPrefix(schema, "ds_")
}

// [dev note] Transaction sequencing flow:
// - when ready to commit a tx, increment (UPDATE) the seq int8 in kwild_internal.sentry table
// - request from the repl monitor a promise for the commit ID for that seq
// - commit the tx
// - repl captures the ordered updates for the transaction
// - in repl receiver, decode and record the seq row update from WAL data (the final update before the commit message)
// - send complete commit digest back to the consumer via the promise channel for that seq value
// - ensure it matches the seq in the exec just prior
//
// To prepare for the above, initialize as follows:
// - create kwild_internal.sentry table if not exists
// - insert row with seq=0, if no rows

// NewDB creates a new Kwil DB instance. On creation, it will connect to the
// configured postgres process, creating as many connections as specified by the
// PoolConfig plus a special connection for a logical replication slot receiver.
func NewDB(ctx context.Context, cfg *DBConfig) (*DB, error) {
	// Create the unrestricted connection pool.
	pool, err := NewPool(ctx, &cfg.PoolConfig)
	if err != nil {
		return nil, err
	}

	// Ensure all tables that are created with no primary key or unique index
	// are altered to have "full replication identity" for UPDATE and DELETES.
	if err = ensureTriggerReplIdentity(ctx, pool.writer); err != nil {
		return nil, err
	}

	if err = ensureKvTable(ctx, pool.writer); err != nil {
		return nil, err
	}

	okSchema := cfg.SchemaFilter
	if okSchema == nil {
		okSchema = defaultSchemaFilter
	}

	repl, err := newReplMon(ctx, cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.DBName, okSchema)
	if err != nil {
		return nil, err
	}

	// Create the tx sequence table with single row if it doesn't exists.
	if err = ensureSentryTable(ctx, pool.writer); err != nil {
		return nil, err
	}

	// Register the error function so a statement like `SELECT error('boom');`
	// will raise an exception and cause the query to error.
	if err = ensureErrorPLFunc(ctx, pool.writer); err != nil { // not sure this is the place to do this
		return nil, err
	}

	return &DB{
		pool: pool,
		repl: repl,
	}, nil
}

// Close shuts down the Kwil DB. This stops all connections and the WAL data
// receiver.
func (db *DB) Close() error {
	db.repl.stop()
	return db.pool.Close()
}

// AutoCommit toggles auto-commit mode, in which the Execute method may be used
// without having to begin/commit. This is to support startup and initialization
// tasks that occur prior to the start of the atomic commit process used while
// executing blocks.
func (db *DB) AutoCommit(auto bool) {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	if db.tx != nil {
		panic("already in a tx")
	}
	db.autoCommit = auto
}

// For {accounts,validators}.Datasets / registry.DB
var _ sql.Executor = (*DB)(nil)
var _ sql.Queryer = (*DB)(nil)
var _ sql.PendingQueryer = (*DB)(nil)
var _ sql.KV = (*DB)(nil)

var _ sql.TxMaker = (*DB)(nil) // for dataset Registry

// BeginTx makes the DB's singular transaction, which is used automatically by
// consumers of the Query and Execute methods. This is the mode of operation
// used by Kwil to have one system coordinating transaction lifetime, with one
// or more other systems implicitly using the transaction for their queries.
//
// The returned transaction is also capable of creating nested transactions.
// This functionality is used to prevent user dataset query errors from rolling
// back the outermost transaction.
func (db *DB) BeginTx(ctx context.Context) (sql.Tx, error) {
	tx, err := db.beginTx(ctx)
	if err != nil {
		return nil, err
	}

	ntx := &nestedTx{tx}
	return &dbTx{ntx, db}, nil
}

var _ sql.TxBeginner = (*DB)(nil) // for CommittableStore => MultiCommitter

// Begin is for consumers that require a smaller interface on the return but
// same instance of the concrete type, a case which annoyingly creates
// incompatible interfaces in Go.
func (db *DB) Begin(ctx context.Context) (sql.TxCloser, error) {
	return db.BeginTx(ctx) // just slice down sql.Tx
}

// beginTx is the critical section of BeginTx
func (db *DB) beginTx(ctx context.Context) (pgx.Tx, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	if db.tx != nil {
		return nil, errors.New("tx exists")
	}

	tx, err := db.pool.writer.BeginTx(ctx, pgx.TxOptions{
		AccessMode: pgx.ReadWrite,
		IsoLevel:   pgx.RepeatableRead,
	})
	if err != nil {
		return nil, err
	}

	// Make the tx available to Execute and QueryPending.
	db.tx = tx
	db.commitID = nil

	return tx, nil
}

// commit is called from the Commit method of the sql.Tx (or sql.TxCloser)
// returned from BeginTx (or Begin). See tx.go.
func (db *DB) commit(ctx context.Context) error {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	if db.tx == nil && !db.autoCommit {
		return errors.New("no tx exists")
	}

	defer db.tx.Rollback(ctx) // yes, safe even on non-error Commit!

	// Do the seq update in sentry table. This ensures a replication message
	// sequence is emitted from this transaction, and that it the data returned
	// from it includes the expected seq value.
	seq, err := incrementSeq(ctx, db.tx)
	if err != nil {
		return err
	}
	logger.Debugf("updated seq to %d", seq)
	resChan := db.repl.recvID(seq)

	err = db.tx.Commit(ctx)
	db.tx = nil
	if err != nil {
		return err
	}

	// Get and store the "commit id". NOTE: could return the channel to the dbTx
	// for it to receive async since the channel is linked to this seq.
	select {
	case db.commitID = <-resChan:
		logger.Infof("received commit ID %x", db.commitID)
	case err = <-db.repl.errChan: // the replMon has died, so probably DB should close too...
	case <-ctx.Done():
		return ctx.Err()
	}

	return err
}

// rollback is called from the Rollback method of the sql.Tx (or sql.TxCloser)
// returned from BeginTx (or Begin). See tx.go.
func (db *DB) rollback(ctx context.Context) error {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	if db.tx == nil {
		return errors.New("no tx exists")
	}

	err := db.tx.Rollback(ctx)
	db.tx = nil
	return err
}

// CommitID returns the commit ID from the last committed transaction.
func (db *DB) CommitID() []byte {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	// if db.tx != nil { // remove
	// 	panic("called commitID with uncommitted transaction")
	// }
	return db.commitID
}

// Query performs a read-only query on a read connection.
func (db *DB) Query(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	// Pass through to the read pool, isolated from any active transactions on
	// the write connection.
	return db.pool.Query(ctx, stmt, args...)
}

func (db *DB) QueryPending(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	if db.tx == nil && !db.autoCommit {
		return nil, sql.ErrNoTransaction
	}
	if db.autoCommit { // with no txn, this is the same as using the read pool
		if db.tx != nil {
			return nil, errors.New("tx already created, cannot use auto commit")
		}
		// query(ctx, db.pool.writer.Query, stmt, args...)
		return db.pool.QueryPending(ctx, stmt, args...)
	} // actually I'm not certain that query pending makes sense in autocommit

	return query(ctx, db.tx.Query, stmt, args...)
}

// discardCommitID is for Execute when in auto-commit mode.
func (db *DB) discardCommitID(ctx context.Context, resChan chan []byte) {
	select {
	case cid := <-resChan:
		logger.Infof("discarding commit ID %x", cid)
	case <-db.repl.done:
	case <-ctx.Done():
	}
}

// Execute runs a statement on an existing transaction, or on a short lived
// transaction from the write connection if in auto-commit mode.
func (db *DB) Execute(ctx context.Context, stmt string, args ...any) error {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	if db.tx != nil {
		if db.autoCommit {
			return errors.New("tx already created, cannot use auto commit")
		}
		_, err := db.tx.Exec(ctx, stmt, args...)
		return err // fmt.Println(execRes.RowsAffected())
	}
	if !db.autoCommit {
		return sql.ErrNoTransaction
	}

	// if db.tx == nil {
	// 	if !db.autoCommit {
	// 		return sql.ErrNoTransaction
	// 	}
	// } else if !db.autoCommit {
	// 	_, err := db.tx.Exec(ctx, stmt, args...)
	// 	return err // fmt.Println(execRes.RowsAffected())
	// } else {
	// 	return errors.New("tx already created, cannot use auto commit")
	// }

	// We do manual autocommit since postgresql will skip it for some
	// statements, plus we are also injecting the seq update query.
	var resChan chan []byte
	err := pgx.BeginTxFunc(ctx, db.pool.writer,
		pgx.TxOptions{
			AccessMode: pgx.ReadWrite,
			IsoLevel:   pgx.RepeatableRead,
		},
		func(tx pgx.Tx) error {
			seq, err := incrementSeq(ctx, tx)
			if err != nil {
				return err
			}
			resChan = db.repl.recvID(seq)
			_, err = tx.Exec(ctx, stmt, args...)
			return err
		},
	)
	if err != nil {
		return err
	}
	db.discardCommitID(ctx, resChan)
	return nil
}

// Get retrieves the value for a key using Query (read-only), optionally using
// QueryPending if the write connection should be used to get uncommitted
// (pending) data if currently in a transaction. If there is no stored value for
// the key, both the returned slice and error are nil.
//
// NOTE: This DB type is not aware of a user dataset "dbid", so there is just
// one global kv table. It might be preferable to implement Get/Set via the
// other methods using statements crafted at a higher level, which would
// facilitate separate kv tables for different Kwil user datasets.
func (db *DB) Get(ctx context.Context, key []byte, pending bool) ([]byte, error) {
	queryFun := db.Query
	if pending {
		queryFun = db.QueryPending
	}
	res, err := queryFun(ctx, selectKvStmt, key)
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

// Set
func (db *DB) Set(ctx context.Context, key []byte, value []byte) error {
	return db.Execute(ctx, insertKvStmt, key, value)
}

// TODO: require rw with target_session_attrs=read-write ?
