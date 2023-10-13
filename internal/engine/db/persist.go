package db

import (
	"context"

	"github.com/kwilteam/kwil-db/core/types/serialize"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/utils/order"
)

const (
	tableVersion     = 1
	procedureVersion = 2
	extensionVersion = 1
)

// persistTableMetadata persists the metadata for a table to the database
func (d *DB) persistTableMetadata(ctx context.Context, table *types.Table) error {
	bts, err := serialize.Encode(table)
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
	bts, err := serialize.Encode(procedure)
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
	bts, err := serialize.Encode(&encodeableExtension{
		Name:           extension.Name,
		Initialization: order.OrderMap(extension.Initialization),
		Alias:          extension.Alias,
	})
	if err != nil {
		return err
	}

	return d.persistVersionedMetadata(ctx, extension.Alias, metadataTypeExtension, &VersionedMetadata{
		Version: extensionVersion,
		Data:    bts,
	})
}

// encodeableExtension is a modification of the extension struct that can be encoded
// using rlp. This is because the extension struct contains a map[string]string and
// since maps cannot be rlp encoded, we need to convert the map[string]string to a slice
// of key value pairs
type encodeableExtension struct {
	Name           string
	Initialization []*order.KVPair[string, string]
	Alias          string
}

// ListExtensions lists all extensions in the database
func (d *DB) ListExtensions(ctx context.Context) ([]*types.Extension, error) {
	meta, err := d.getVersionedMetadata(ctx, metadataTypeExtension)
	if err != nil {
		return nil, err
	}

	encodeable, err := decodeMetadata[encodeableExtension](meta)
	if err != nil {
		return nil, err
	}

	var extensions []*types.Extension
	for _, ext := range encodeable {
		extensions = append(extensions, &types.Extension{
			Name:           ext.Name,
			Initialization: order.ToMap(ext.Initialization),
			Alias:          ext.Alias,
		})
	}

	return extensions, nil
}
