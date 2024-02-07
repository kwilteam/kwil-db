package pg

// This file defines the nestedTx and dbTx types; it's all sql.Tx to consumers.

import (
	"context"
	"errors"

	"github.com/kwilteam/kwil-db/internal/sql"

	"github.com/jackc/pgx/v5"
)

// nestedTx is returned from the BeginTx method of both dbTx or another
// nestedTx. The underlying pgx.Tx is embedded so we do not need to redefine the
// Commit and Rollback methods.
type nestedTx struct {
	pgx.Tx
	accessMode sql.AccessMode
}

var _ sql.OuterTx = (*nestedTx)(nil)

// TODO: switch this to be BeginTx
func (tx *nestedTx) BeginSavepoint(ctx context.Context) (sql.Tx, error) {
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

func (tx *nestedTx) Query(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return query(ctx, tx.Tx, stmt, args...)
}

// Execute is now literally identical to Query in both semantics and syntax. We
// might remove one or the other in this context (transaction methods).
func (tx *nestedTx) Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return query(ctx, tx.Tx, stmt, args...)
}

func (tx *nestedTx) Precommit(context.Context) ([]byte, error) {
	// only the outer transaction does the prepared transaction
	return nil, errors.New("cannot prepare transaction from a nested transaction")
}

// AccessMode returns the access mode of the transaction.
func (tx *nestedTx) AccessMode() sql.AccessMode {
	return tx.accessMode
}

// Commit is direct from embedded pgx.Tx.
// func (tx *nestedTx) Commit(ctx context.Context) error { return tx.Tx.Commit(ctx) }

// Rollback is direct from embedded pgx.Tx. It is ok to call Rollback repeatedly
// and even after Commit with no error.
// func (tx *nestedTx) Rollback(ctx context.Context) error { return tx.Tx.Rollback(ctx) }

// dbTx is the type returned by (*DB).BeginTx. It embeds all the nestedTx
// methods (thus returning a *nestedTx from it's BeginTx), but shadows Commit
// and Rollback to allow the DB to begin a subsequent transaction, and to
// coordinate the two-phase commit process using a "prepared transaction".
type dbTx struct {
	*nestedTx      // should embed pgx.Tx
	db         *DB // for top level DB lifetime mgmt
	accessMode sql.AccessMode
}

// Precommit creates a prepared transaction for a two-phase commit. An ID
// derived from the updates is return. This must be called before Commit. Either
// Commit or Rollback must follow.
func (tx *dbTx) Precommit(ctx context.Context) ([]byte, error) {
	return tx.db.precommit(ctx)
}

// Commit commits the transaction. This partly satisfies sql.Tx.
func (tx *dbTx) Commit(ctx context.Context) error {
	return tx.db.commit(ctx)
}

// Rollback rolls back the transaction. This partly satisfies sql.Tx.
func (tx *dbTx) Rollback(ctx context.Context) error {
	return tx.db.rollback(ctx)
}

// AccessMode returns the access mode of the transaction.
func (tx *dbTx) AccessMode() sql.AccessMode {
	return tx.accessMode
}

// pgxAccessLevel converts a sql.AccessMode to a pgx.TxAccessMode.
func pgxAccessLevel(a sql.AccessMode) pgx.TxAccessMode {
	switch a {
	case sql.ReadOnly:
		return pgx.ReadOnly
	case sql.ReadWrite:
		return pgx.ReadWrite
	default:
		panic("unknown access level")
	}
}
