package db

import (
	"context"
	"fmt"

	sqlddlgenerator "github.com/kwilteam/kwil-db/pkg/engine/db/sql-ddl-generator"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/mitchellh/mapstructure"
)

const (
	tableVersion = 1
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
	return d.persistVersionedMetadata(ctx, &versionedMetadata{
		Version: tableVersion,
		Data:    table,
	})
}

// ListTables lists all tables in the database
func (d *DB) ListTables(ctx context.Context) ([]*types.Table, error) {
	meta, err := d.getVersionedMetadata(ctx, metadataTypeTable)
	if err != nil {
		return nil, err
	}

	var tables []*types.Table

	for _, value := range meta {
		tbl := types.Table{}
		err = mapstructure.Decode(value.Data, &tbl)
		if err != nil {
			return nil, err
		}

		tables = append(tables, &tbl)
	}

	return tables, nil
}
