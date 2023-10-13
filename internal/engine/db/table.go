package db

import (
	"context"
	"fmt"

	sqlddlgenerator "github.com/kwilteam/kwil-db/internal/engine/db/sql-ddl-generator"
	"github.com/kwilteam/kwil-db/internal/engine/types"
)

// CreateTable creates a new table and persists the metadata to the database
func (d *DB) CreateTable(ctx context.Context, table *types.Table) error {
	savepoint, err := d.Sqldb.Savepoint()
	if err != nil {
		return fmt.Errorf("failed to create savepoint: %w", err)
	}
	defer savepoint.Rollback()

	err = d.deployTable(ctx, table)
	if err != nil {
		return fmt.Errorf("failed to deploy table: %w", err)
	}

	err = d.persistTableMetadata(ctx, table)
	if err != nil {
		return fmt.Errorf("failed to persist table metadata: %w", err)
	}

	return savepoint.Commit()
}

// deployTable deploys a new table to the database
func (d *DB) deployTable(ctx context.Context, table *types.Table) error {
	ddlStmts, err := sqlddlgenerator.GenerateDDL(table)
	if err != nil {
		return err
	}

	savepoint, err := d.Sqldb.Savepoint()
	if err != nil {
		return err
	}
	defer savepoint.Rollback()

	for _, ddlStmt := range ddlStmts {
		err = d.Sqldb.Execute(ctx, ddlStmt, nil)
		if err != nil {
			return err
		}
	}

	return savepoint.Commit()
}
