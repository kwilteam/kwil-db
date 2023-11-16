/*
Package metadata wraps a SQL connection with metadata operations.
These operations are used to store and retrieve metadata about a database.
*/
package metadata

import (
	"context"
	"encoding/json"

	ddl "github.com/kwilteam/kwil-db/internal/engine/metadata/ddl"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	sql "github.com/kwilteam/kwil-db/internal/sql"
)

// KV is a key/value store.
type KV interface {
	Set(ctx context.Context, key []byte, value []byte) error
	Get(ctx context.Context, key []byte) ([]byte, error)
}

// RunMigration runs a migration against the metadata store.
// This is used in case of an update to Kwil's schema structure.
// It will read out the current schema version, update them, and store them, all if necessary.
func RunMigration(ctx context.Context, kv KV) error {
	// currently, we have no upgrades to the metadata store
	return nil
}

// CreateTables creates tables and stores them
func CreateTables(ctx context.Context, tables []*types.Table, kv KV, exec sql.ResultSetFunc) error {
	for _, table := range tables {
		ddl, err := ddl.GenerateDDL(table)
		if err != nil {
			return err
		}

		for _, stmt := range ddl {
			_, err := exec(ctx, stmt, nil)
			if err != nil {
				return err
			}
		}
	}

	// we can use json here since metadata storage is an implementation detail and not part of consensus
	bts, err := json.Marshal(tables)
	if err != nil {
		return err
	}

	err = storeMetadata(ctx, kv, &metadata{
		Type:    metadataTypeTable,
		Content: bts,
	})
	if err != nil {
		return err
	}

	return nil
}

// ListTables lists all tables in the database.
func ListTables(ctx context.Context, kv KV) ([]*types.Table, error) {
	metadata, err := getMetadata(ctx, kv, metadataTypeTable)
	if err != nil {
		return nil, err
	}

	tables := []*types.Table{}
	err = json.Unmarshal(metadata, &tables)

	return tables, err
}

// StoreProcedures stores a procedure in the metadata store.
func StoreProcedures(ctx context.Context, procedures []*types.Procedure, kv KV) error {
	bts, err := json.Marshal(procedures)
	if err != nil {
		return err
	}

	err = storeMetadata(ctx, kv, &metadata{
		Type:    metadataTypeProcedure,
		Content: bts,
	})
	if err != nil {
		return err
	}

	return nil
}

// ListProcedures lists all procedures in the database.
func ListProcedures(ctx context.Context, kv KV) ([]*types.Procedure, error) {
	bts, err := getMetadata(ctx, kv, metadataTypeProcedure)
	if err != nil {
		return nil, err
	}

	procedures := []*types.Procedure{}
	err = json.Unmarshal(bts, &procedures)

	return procedures, err
}

// StoreExtension stores an extension in the metadata store.
func StoreExtensions(ctx context.Context, extensions []*types.Extension, kv KV) error {
	bts, err := json.Marshal(extensions)
	if err != nil {
		return err
	}

	return storeMetadata(ctx, kv, &metadata{
		Type:    metadataTypeExtension,
		Content: bts,
	})
}

// ListExtensions lists all extensions in the database.
func ListExtensions(ctx context.Context, kv KV) ([]*types.Extension, error) {
	bts, err := getMetadata(ctx, kv, metadataTypeExtension)
	if err != nil {
		return nil, err
	}

	extensions := []*types.Extension{}
	err = json.Unmarshal(bts, &extensions)

	return extensions, err
}
