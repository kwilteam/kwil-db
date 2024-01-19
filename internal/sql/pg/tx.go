package pg

// This file defines the nestedTx and dbTx types; it's all sql.Tx to consumers.

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/sql/v2" // temporary v2 for refactoring

	"github.com/jackc/pgx/v5"
)

// nestedTx is returned from the BeginTx method of both dbTx or another
// nestedTx. The underlying pgx.Tx is embedded so we do not need to redefine the
// Commit and Rollback methods.
type nestedTx struct {
	pgx.Tx
}

var _ sql.Tx = (*nestedTx)(nil)

func (tx *nestedTx) BeginTx(ctx context.Context) (sql.Tx, error) {
	// Make the nested transaction (savepoint)
	pgtx, err := tx.Tx.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return &nestedTx{pgtx}, nil
}

func (tx *nestedTx) Query(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return query(ctx, tx.Tx.Query, stmt, args...)
}

func (tx *nestedTx) Execute(ctx context.Context, stmt string, args ...any) error {
	_, err := tx.Tx.Exec(ctx, stmt, args...)
	return err // fmt.Println(execRes.RowsAffected())
}

// Commit is direct from embedded pgx.Tx.
// func (tx *nestedTx) Commit(ctx context.Context) error { return tx.Tx.Commit(ctx) }

// Rollback is direct from embedded pgx.Tx. It is ok to call Rollback repeatedly
// and even after Commit with no error.
// func (tx *nestedTx) Rollback(ctx context.Context) error { return tx.Tx.Rollback(ctx) }

// dbTx is the type returned by (*DB).BeginTx. It embeds all the nestedTx
// methods (thus returning a *nestedTx from it's BeginTx), but shadows Commit
// and Rollback to allow the DB to begin a new transaction.
type dbTx struct {
	*nestedTx     // should embed pgx.Tx
	db        *DB // for top level DB lifetime mgmt
}

// Commit commits the transaction. This partly satisfies sql.Tx.
func (tx *dbTx) Commit(ctx context.Context) error {
	return tx.db.commit(ctx)
}

// Rollback rolls back the transaction. This partly satisfies sql.Tx.
func (tx *dbTx) Rollback(ctx context.Context) error {
	return tx.db.rollback(ctx)
}
