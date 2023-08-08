package db

import (
	"context"
	"encoding/json"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

const (
	tableVersion     = 1
	procedureVersion = 2
	extensionVersion = 1
)

// persistTableMetadata persists the metadata for a table to the database
func (d *DB) persistTableMetadata(ctx context.Context, table *types.Table) error {
	bts, err := json.Marshal(table)
	if err != nil {
		return err
	}

	return d.persistVersionedMetadata(ctx, table.Name, metadataTypeTable, &VersionedMetadata{
		Version: tableVersion,
		Data:    bts,
	})
}

// ListTables lists all tables in the database
func (d *DB) ListTables(ctx context.Context) ([]*types.Table, error) {
	meta, err := d.getVersionedMetadata(ctx, metadataTypeTable)
	if err != nil {
		return nil, err
	}

	return decodeMetadata[types.Table](meta)
}

// StoreProcedure stores a procedure in the database
func (d *DB) StoreProcedure(ctx context.Context, procedure *types.Procedure) error {
	bts, err := json.Marshal(procedure)
	if err != nil {
		return err
	}

	return d.persistVersionedMetadata(ctx, procedure.Name, metadataTypeProcedure, &VersionedMetadata{
		Version: procedureVersion,
		Data:    bts,
	})
}

// ListProcedures lists all procedures in the database
func (d *DB) ListProcedures(ctx context.Context) ([]*types.Procedure, error) {
	meta, err := d.getVersionedMetadata(ctx, metadataTypeProcedure)
	if err != nil {
		return nil, err
	}

	return decodeVersionedProcedures(meta)
}

// StoreExtension stores an extension in the database
func (d *DB) StoreExtension(ctx context.Context, extension *types.Extension) error {
	bts, err := json.Marshal(extension)
	if err != nil {
		return err
	}

	return d.persistVersionedMetadata(ctx, extension.Alias, metadataTypeExtension, &VersionedMetadata{
		Version: extensionVersion,
		Data:    bts,
	})
}

// ListExtensions lists all extensions in the database
func (d *DB) ListExtensions(ctx context.Context) ([]*types.Extension, error) {
	meta, err := d.getVersionedMetadata(ctx, metadataTypeExtension)
	if err != nil {
		return nil, err
	}

	return decodeMetadata[types.Extension](meta)
}
