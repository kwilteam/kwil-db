/*
	Package registry abstracts multiple datasets via postgresql "schemata",
	which are like namespaces to have multiple sets of tables in the same
	database connection.

	This is potentially an engine package, but it could be applied for other
	non-user datasets such as the accounts and validators stores.

	It implements Set/Get for sessions.Committable to store an idempotency key
	within the registry to track what changes have been committed.

	If a database already contains the idempotent key, it will return nil for
	any incoming operation.
*/

// NOTE: the requirements for Registry:
//  1. satisfy sessions.Committable (Begin/Commit/Cancel) for the MultiCommitter
//  2. satisfy engine/execution.Registry
//	   a. dbid-specific Query/Execute/Set/Get
//	   b. dataset Create/Delete/List (by dbid)

package registry

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/sql"

	"go.uber.org/zap"
)

const (
	sqlSchemaExists = `SELECT 1
		FROM information_schema.schemata
		WHERE schema_name = $1;` // for datasets: 'ds_' + dbid

	datasetSchemaPrefix = `ds_`
	// dsspEscaped         = `ds\_` // underscore requires escaping in LIKE?
	sqlListDatasets = `SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name LIKE '` + datasetSchemaPrefix + `%';`

	sqlCreateSchemaTmpl = `CREATE SCHEMA IF NOT EXISTS %s;`   // cant do $1
	sqlDeleteSchemaTmpl = `DROP SCHEMA IF EXISTS %s CASCADE;` // and all objects (tables, indexes, etc.)
)

func datasetSchema(dbid string) string {
	return datasetSchemaPrefix + dbid
}

// DB is the main dependency of a Registry. It provides the query and kv
// functions, and the ability to create transactions (and nested transactions).
type DB interface {
	sql.Queryer
	sql.Executor
	sql.KV
	sql.TxMaker // the special kind that can make nested txns
}

// Registry is used to register databases.
type Registry struct {
	log log.Logger
	db  DB

	// Starting a session creates a transaction and stores a session key
	// provided by the caller.
	mu         sync.RWMutex
	sessionKey []byte
	tx         sql.Tx
	commitID   []byte
}

// New opens a registry.
func New(ctx context.Context, db DB, opts ...RegistryOpt) (*Registry, error) {
	r := &Registry{
		db:  db,
		log: log.NewNoOp(),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

func (r *Registry) schemaExists(ctx context.Context, schema string, sync bool) (bool, error) {
	q := r.db.Query
	if sync {
		q = r.db.Execute
	}
	res, err := q(ctx, sqlSchemaExists, schema)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return len(res.Rows) > 0, nil
}

func (r *Registry) datasetExists(ctx context.Context, dbid string, sync bool) (bool, error) {
	return r.schemaExists(ctx, datasetSchema(dbid), sync)
}

func (r *Registry) assertDatasetExists(ctx context.Context, dbid string, sync bool) error {
	exists, err := r.datasetExists(ctx, dbid, sync)
	if err != nil {
		return err
	}
	if !exists {
		return ErrDatabaseNotFound
	}
	return nil
}

func (r *Registry) inSession() bool {
	return len(r.sessionKey) > 0
}

// Create creates a new database.
// If the database already exists, it returns an error.
func (r *Registry) Create(ctx context.Context, dbid string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.inSession() {
		return ErrRegistryNotWritable
	}

	// check if the dataset already exists
	exists, err := r.datasetExists(ctx, dbid, true)
	if err != nil {
		return err
	}
	if exists {
		return ErrDatabaseExists
	}

	// No need to isolate this in a nested tx as this just creates the dataset's
	// postgres schema. The tables in the schema are created via Execute (see
	// storeSchema in engine.).
	_, err = r.tx.Execute(ctx, fmt.Sprintf(sqlCreateSchemaTmpl, datasetSchema(dbid)))
	return err
}

// Delete deletes a database.
// If the database does not exist, it returns an error.
func (r *Registry) Delete(ctx context.Context, dbid string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.inSession() {
		return ErrRegistryNotWritable
	}

	// check if the database exists
	err := r.assertDatasetExists(ctx, dbid, true)
	if err != nil {
		return err
	}
	_, err = r.tx.Execute(ctx, fmt.Sprintf(sqlDeleteSchemaTmpl, datasetSchema(dbid)))
	return err
}

// List lists all databases. (committed only)
func (r *Registry) List(ctx context.Context) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// instead maintain a r.dbs map field that we update on create/delete? that
	// would be easy and cheaper than this (but would include uncommitted):

	res, err := r.db.Query(ctx, sqlListDatasets)
	if err != nil {
		return nil, err
	}

	dbids := make([]string, len(res.Rows))

	for i, row := range res.Rows {
		if len(row) != 1 {
			return nil, errors.New("not one row")
		}
		dbid, ok := row[0].(string)
		if !ok {
			return nil, errors.New("dbid not a string")
		}
		dbids[i], ok = strings.CutPrefix(dbid, datasetSchemaPrefix)
		if !ok {
			return nil, fmt.Errorf("incorrect schema prefix for dataset %q", dbid)
		}
	}

	return dbids, nil
}

func namespaceKey(dbid string, key []byte) []byte {
	return append([]byte(dbid+";"), key...)
}

// Set sets the value for a key in a database's key value store.
// NOTE: engine also uses this to store schema meta data... alternatives???
func (r *Registry) Set(ctx context.Context, dbid string, key, value []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.inSession() {
		return ErrRegistryNotWritable
	}

	// Check that the schema/namespace for dbid exists
	err := r.assertDatasetExists(ctx, dbid, true)
	if err != nil {
		return err
	}

	// concat dbid and key for the actual db.Set call
	key = namespaceKey(dbid, key)
	return r.db.Set(ctx, key, value) // alt: consider making the statement and using a per-dataset kv table
}

// Get gets the value for a key in a database's key value store.
func (r *Registry) Get(ctx context.Context, dbid string, key []byte, sync bool) ([]byte, error) {
	err := r.assertDatasetExists(ctx, dbid, true)
	if err != nil {
		return nil, err
	}
	key = namespaceKey(dbid, key)
	return r.db.Get(ctx, key, sync)
}

// Execute executes a statement on a database. The statement must already be
// written with the dbid in the schema for the table. The dbid is only used to
// check the existence of the schema before executing the query. If the database
// does not exist, it returns an error. An DB session (outer transaction) must
// already have been started with Begin. The statement is executed within a
// nested transaction so that an error does no rollback the whole session.
func (r *Registry) Execute(ctx context.Context, dbid, stmt string, args ...any) (*sql.ResultSet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.inSession() {
		return nil, ErrRegistryNotWritable
	}

	if err := r.assertDatasetExists(ctx, dbid, true); err != nil {
		return nil, err
	}

	// for i := range args {
	// 	if named, isNamed := args[i].(map[string]any); isNamed {
	// 		args[i] = pgx.NamedArgs(named)
	// 		r.log.Infof("converting named args %v \n%v", args[0], stmt)
	// 		break
	// 	}
	// }

	if len(args) == 1 {
		// if named, isNamed := args[0].(map[string]any); isNamed {
		// 	args[0] = pgx.NamedArgs(named)
		// 	r.log.Infof("converting named args %v \n%v", args[0], stmt)
		// }
		if args[0] == nil { // []any{[]any(nil)} indicates caller probably forgot the "..."
			fmt.Println("**************** ALMOST CERTAINLY CALLER BUG (missing ...?)!!!!")
			args = nil // to a nil []any
		}
	}

	// Execute in a nested transaction a.k.a. savepoint.
	var res *sql.ResultSet
	err := txFn(ctx, r.tx, func(tx sql.Tx) error {
		var err error
		res, err = tx.Execute(ctx, stmt, args...)
		return err
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func txFn(ctx context.Context, tm sql.TxMaker, fn func(tx sql.Tx) error) error {
	tx, err := tm.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx) // note: consumes and replaces CommitID
}

// Query executes a query on a database.
// If the database does not exist, it returns an error.
// If in a session, it will reflect uncommitted state.
// ^^^^^ UNFORTUNATELY THAT DEPENDS ON CONTEXT IF DESIRABLE TO READ UNCOMMITTED
func (r *Registry) Query(ctx context.Context, dbid, stmt string, args ...any) (*sql.ResultSet, error) {
	err := r.assertDatasetExists(ctx, dbid, false)
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		// if named, isNamed := args[0].(map[string]any); isNamed {
		// 	args[0] = pgx.NamedArgs(named)
		// 	r.log.Info("converting named args")
		// }
		if args[0] == nil { // []any{[]any(nil)} indicates caller probably forgot the "..."
			fmt.Println("**************** ALMOST CERTAINLY CALLER BUG (missing ...?)!!!!")
			args = nil // to a nil []any
		}
	}
	if r.inSession() { // new behavior, but not always correct
		// if in session, use a nested transaction
		var res *sql.ResultSet
		err := txFn(ctx, r.tx, func(tx sql.Tx) error {
			res, err = tx.Query(ctx, stmt, args...)
			return err
		})
		return res, err
	}
	return r.db.Query(ctx, stmt, args...)
}

// Close closes the registry.
func (r *Registry) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.inSession() {
		return r.tx.Rollback(ctx) // would be done anyway on DB.Close?
	} // maybe just call r.Cancel unconditionally?

	return nil
}

// Begin signals the start of a session. A session is a series of operations
// across many datasets that are executed atomically with a transaction on the
// parent database, where the individual datasets are in different postgresql
// "schema".
func (r *Registry) Begin(ctx context.Context, idempotencyKey []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.inSession() {
		return ErrAlreadyInSession
	}

	// First check the latest idempotency key in the kv.  If it's the same,
	// block stuff until commit (or next begin).

	if lastKey, err := getIdempotencyKey(ctx, r.db); err != nil {
		return err
	} else if bytes.Equal(lastKey, idempotencyKey) {
		return fmt.Errorf("registry received duplicate idempotency key, recovery not possible")
	}

	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return err
	}

	r.sessionKey = idempotencyKey
	r.tx = tx

	return nil
}

func (r *Registry) Precommit(ctx context.Context) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.inSession() {
		return nil, ErrRegistryNotWritable
	}

	commitID, err := r.tx.Precommit(ctx)
	if err != nil {
		return nil, err
	}
	r.commitID = commitID // hang on to it to save it in Commit (if we even need to)

	return commitID, nil
}

// Commit signals the end of a session. It returns the apphash.
// All databases will be committed, in order.
func (r *Registry) Commit(ctx context.Context, idempotencyKey []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.inSession() {
		return ErrRegistryNotWritable
	}

	// Ensure this is committing what was begun with begin.
	if !bytes.Equal(r.sessionKey, idempotencyKey) {
		return fmt.Errorf("%w: expected %x, got %x", ErrIdempotencyKeyMismatch,
			r.sessionKey, idempotencyKey)
	}

	err := setIdempotencyKey(ctx, r.db, idempotencyKey)
	if err != nil {
		return err
	}

	// commit the prepared transaction
	if err := r.tx.Commit(ctx); err != nil {
		return err // don't need to explicitly rollback, right? (check)
	}
	r.sessionKey = nil
	r.tx = nil

	// can't do this outside of a transaction, but can't do it after PREPARE
	// TRANSACTION either. We will probably remove this since loading it was
	// only needed to power the old recover mode
	if err = r.adHocTx(ctx, func() error { return setAppHash(ctx, r.db, r.commitID) }); err != nil {
		r.log.Error("couldn't save commit id to kv store", zap.Error(err))
	}

	r.commitID = nil

	return nil
}

// adHocTx is a hack so we can store the apphash returned from the main tx
// Commit (i.e. outside of Begin/Commit). This method must not be used while
// there is already an active tx. It's also a problem if we crash before saving
// the apphash with this.
func (r *Registry) adHocTx(ctx context.Context, fn func() error) error {
	if r.inSession() {
		return ErrAlreadyInSession
	}
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	if err = fn(); err != nil {
		tx.Rollback(ctx)
		return err
	}
	if _, err := tx.Precommit(ctx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Cancel signals that the session should be cancelled.
// If no session is in progress, it returns nil.
func (r *Registry) Cancel(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.inSession() {
		return nil
	}

	err := r.tx.Rollback(ctx)

	r.sessionKey = nil

	return err
}

var (
	// idempotencyKeyKey is the key for the idempotency key.
	idempotencyKeyKey = []byte("idempotency_key")
	// appHashKey is the key for the app hash.
	appHashKey = []byte("app_hash")
)

// getIdempotencyKey gets the most recently persisted idempotency key for a database.
func getIdempotencyKey(ctx context.Context, conn KVGetter) ([]byte, error) {
	return conn.Get(ctx, idempotencyKeyKey, false)
}

// setIdempotencyKey sets the idempotency key for a database.
func setIdempotencyKey(ctx context.Context, conn KV, idempotencyKey []byte) error {
	return conn.Set(ctx, idempotencyKeyKey, idempotencyKey)
}

// getAppHash gets the most recently persisted app hash for a database.
func getAppHash(ctx context.Context, conn KVGetter) ([]byte, error) { //nolint:unused
	return conn.Get(ctx, appHashKey, false)
}

// setAppHash sets the app hash for a database.
func setAppHash(ctx context.Context, conn KV, appHash []byte) error {
	return conn.Set(ctx, appHashKey, appHash)
}
