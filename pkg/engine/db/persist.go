package db

import (
	"context"
	"fmt"

	sqlddlgenerator "github.com/kwilteam/kwil-db/pkg/engine/db/sql-ddl-generator"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

// CreateTable creates a new table and persists the metadata to the database
func (d *DB) CreateTable(ctx context.Context, table *types.Table) error {
	savepoint, err := d.sqldb.Savepoint()
	if err != nil {
		return fmt.Errorf("failed to create savepoint: %w", err)
	}
	defer savepoint.Rollback()

	err = d.deployTable(table)
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
func (d *DB) deployTable(table *types.Table) error {
	ddlStmts, err := sqlddlgenerator.GenerateDDL(table)
	if err != nil {
		return err
	}

	savepoint, err := d.sqldb.Savepoint()
	if err != nil {
		return err
	}
	defer savepoint.Rollback()

	for _, ddlStmt := range ddlStmts {
		err = d.sqldb.Execute(ddlStmt)
		if err != nil {
			return err
		}
	}

	return savepoint.Commit()
}

// persistTableMetadata persists the metadata for a table to the database
func (d *DB) persistTableMetadata(ctx context.Context, table *types.Table) error {
	return serdes[types.Table]{
		db: d,
	}.persistSerializable(ctx, table)
}

// ListTables lists all tables in the database
func (d *DB) ListTables(ctx context.Context) ([]*types.Table, error) {
	return serdes[types.Table]{
		db: d,
	}.listDeserialized(ctx)
}

// StoreProcedure stores a procedure in the database
func (d *DB) StoreProcedure(ctx context.Context, procedure *types.Procedure) error {
	return serdes[types.Procedure]{
		db: d,
	}.persistSerializable(ctx, procedure)
}

// ListProcedures lists all procedures in the database
func (d *DB) ListProcedures(ctx context.Context) ([]*types.Procedure, error) {
	return serdes[types.Procedure]{db: d}.listDeserialized(ctx)
}
