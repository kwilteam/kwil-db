package schema

import (
	"context"
	"database/sql"
	"kwil/x/sqlx/errors"
	"kwil/x/sqlx/sqlclient"
	"strings"
)

type SchemaManager interface {
	CreateSchema(ctx context.Context, schemaName string) error
	SchemaExists(ctx context.Context, schemaName string) (bool, error)
	DropSchema(ctx context.Context, schemaName string) error
}

// SchemaManagerTxer is a SchemaManager that can be used with a transaction
type SchemaManagerTxer interface {
	SchemaManager
	WithTx(tx *sql.Tx) SchemaManagerTxer
}

type schemaManager struct {
	db sqlclient.DBTX
}

func New(db *sqlclient.DB) SchemaManagerTxer {
	return &schemaManager{
		db: db,
	}
}

func (q *schemaManager) WithTx(tx *sql.Tx) SchemaManagerTxer {
	return &schemaManager{
		db: tx,
	}
}

func (q *schemaManager) CreateSchema(ctx context.Context, schemaName string) error {
	sb := strings.Builder{}
	sb.WriteString("CREATE SCHEMA ")
	sb.WriteString(schemaName)
	sb.WriteString(";")
	_, err := q.db.ExecContext(ctx, sb.String())
	return err
}

func (q *schemaManager) SchemaExists(ctx context.Context, schemaName string) (bool, error) {
	var exists bool
	err := q.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_catalog.pg_namespace WHERE nspname = $1);", schemaName).Scan(&exists)
	if errors.IsNoRowsInResult(err) {
		return false, nil
	}

	return exists, err
}

func (q *schemaManager) DropSchema(ctx context.Context, schemaName string) error {
	sb := strings.Builder{}
	sb.WriteString("DROP SCHEMA ")
	sb.WriteString(schemaName)
	sb.WriteString(" CASCADE;")
	_, err := q.db.ExecContext(ctx, sb.String())
	return err
}
