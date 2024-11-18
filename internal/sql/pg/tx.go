package pg

// This file defines the nestedTx and dbTx types; it's all sql.Tx to consumers.

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	common "github.com/kwilteam/kwil-db/common/sql"
)

type releaser interface {
	Release()
}

// nestedTx is returned from the BeginTx method of both dbTx or another
// nestedTx. The underlying pgx.Tx is embedded so we do not need to redefine the
// Commit and Rollback methods.
type nestedTx struct {
	pgx.Tx
	accessMode common.AccessMode
}

var _ common.Tx = (*nestedTx)(nil)

// BeginTx creates a new transaction with the same access mode as the parent.
// Internally this is savepoint, which allows rollback to the innermost
// savepoint rather than the entire outer transaction.
func (tx *nestedTx) BeginTx(ctx context.Context) (common.Tx, error) {
	// Make the nested transaction (savepoint)
	pgtx, err := tx.Tx.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return &nestedTx{
		Tx:         pgtx,
		accessMode: tx.accessMode,
	}, nil
}

func (tx *nestedTx) Query(ctx context.Context, stmt string, args ...any) (*common.ResultSet, error) {
	resSet, err := query(ctx, tx.Tx, stmt, args...)
	if errors.Is(err, common.ErrDBFailure) {
		tx.Tx.Conn().Close(ctx) // do not allow the outer txn to commit!
	}
	return resSet, err
}

// Execute is now literally identical to Query in both semantics and syntax. We
// might remove one or the other in this context (transaction methods).
func (tx *nestedTx) Execute(ctx context.Context, stmt string, args ...any) (*common.ResultSet, error) {
	resSet, err := query(ctx, tx.Tx, stmt, args...)
	if errors.Is(err, common.ErrDBFailure) {
		tx.Tx.Conn().Close(ctx) // do not allow the outer txn to commit!
	}
	return resSet, err
}

// AccessMode returns the access mode of the transaction.
func (tx *nestedTx) AccessMode() common.AccessMode {
	return tx.accessMode
}

// dbTx is the type returned by (*DB).BeginTx. It embeds all the nestedTx
// methods (thus returning a *nestedTx from it's BeginTx), but shadows Commit
// and Rollback to allow the DB to begin a subsequent transaction, and to
// coordinate the two-phase commit process using a "prepared transaction".
type dbTx struct {
	*nestedTx      // should embed pgx.Tx
	db         *DB // for top level DB lifetime mgmt
	accessMode common.AccessMode
}

// Precommit creates a prepared transaction for a two-phase commit. An ID
// derived from the updates is return. This must be called before Commit. Either
// Commit or Rollback must follow.
func (tx *dbTx) Precommit(ctx context.Context) ([]byte, error) {
	return tx.db.precommit(ctx)
}

// Commit commits the transaction. This partly satisfies sql.Tx.
func (tx *dbTx) Commit(ctx context.Context) error {
	if rel, ok := tx.nestedTx.Tx.(releaser); ok {
		defer rel.Release()
	}
	return tx.db.commit(ctx)
}

// Rollback rolls back the transaction. This partly satisfies sql.Tx.
func (tx *dbTx) Rollback(ctx context.Context) error {
	if rel, ok := tx.nestedTx.Tx.(releaser); ok {
		defer rel.Release()
	}
	return tx.db.rollback(ctx)
}

// AccessMode returns the access mode of the transaction.
func (tx *dbTx) AccessMode() common.AccessMode {
	return tx.accessMode
}

// readTx is a tx that handles a read-only transaction.
// It will release the connection back to the reader pool
// when it is committed or rolled back.
type readTx struct {
	*nestedTx
	release func()
}

// Commit is a no-op for read-only transactions.
// It will unconditionally return the connection to the pool.
func (tx *readTx) Commit(ctx context.Context) error {
	defer tx.release()

	return tx.nestedTx.Commit(ctx)
}

// Rollback will unconditionally return the connection to the pool.
func (tx *readTx) Rollback(ctx context.Context) error {
	defer tx.release()

	return tx.nestedTx.Rollback(ctx)
}
