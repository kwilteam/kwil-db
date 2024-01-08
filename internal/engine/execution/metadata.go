package execution

import (
	"context"
	"fmt"

	coreTypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/metadata"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	sql "github.com/kwilteam/kwil-db/internal/sql"
)

// executor makes a registry `Execute` or `Query` method into a sql.Executor
func executor(dbid string, fn func(ctx context.Context, dbid string, stmt string, params map[string]any) (coreTypes.ResultSet, error)) sql.ResultSetFunc {
	return func(ctx context.Context, stmt string, params map[string]any) (coreTypes.ResultSet, error) {
		return fn(ctx, dbid, stmt, params)
	}
}

// this file contains the metadata store for the execution engine
// metadata is not synchronized as part of consensus, so it can be non-deterministic

// metadataKv is a key-value store for metadata.
type metadataKv struct {
	// registry is the registry to use for metadata.
	registry Registry
	// dbid is the database id to use for metadata.
	dbid string
	// sync indicates whether it should read uncommitted data.
	// if false, it cannot Set values.
	sync bool
}

func (m *metadataKv) Set(ctx context.Context, key, value []byte) error {
	if !m.sync {
		return fmt.Errorf("cannot set metadata in a async-sync context")
	}
	return m.registry.Set(ctx, m.dbid, key, value)
}

// Get gets a value for a key.
func (m *metadataKv) Get(ctx context.Context, key []byte) ([]byte, error) {
	return m.registry.Get(ctx, m.dbid, key, m.sync)
}

var (
	ownerKey  = []byte("owner")
	dbNameKey = []byte("name")
)

// storeSchema stores a schema in the datastore.
func storeSchema(ctx context.Context, schema *types.Schema, datastore Registry) error {
	kv := &metadataKv{
		registry: datastore,
		dbid:     schema.DBID(),
		sync:     true,
	}

	err := kv.Set(ctx, ownerKey, schema.Owner)
	if err != nil {
		return err
	}

	err = kv.Set(ctx, dbNameKey, []byte(schema.Name))
	if err != nil {
		return err
	}

	err = metadata.CreateTables(ctx, schema.Tables, kv, executor(schema.DBID(), datastore.Execute))
	if err != nil {
		return err
	}

	err = metadata.StoreProcedures(ctx, schema.Procedures, kv)
	if err != nil {
		return err
	}

	err = metadata.StoreExtensions(ctx, schema.Extensions, kv)
	if err != nil {
		return err
	}

	return nil
}

// getSchema gets a schema from the datastore.
func getSchema(ctx context.Context, dbid string, datastore Registry) (*types.Schema, error) {
	kv := &metadataKv{
		registry: datastore,
		dbid:     dbid,
		sync:     false,
	}

	owner, err := kv.Get(ctx, ownerKey)
	if err != nil {
		return nil, err
	}

	name, err := kv.Get(ctx, dbNameKey)
	if err != nil {
		return nil, err
	}

	tables, err := metadata.ListTables(ctx, kv)
	if err != nil {
		return nil, err
	}

	procedures, err := metadata.ListProcedures(ctx, kv)
	if err != nil {
		return nil, err
	}

	extensions, err := metadata.ListExtensions(ctx, kv)
	if err != nil {
		return nil, err
	}

	return &types.Schema{
		Owner:      owner,
		Name:       string(name),
		Tables:     tables,
		Procedures: procedures,
		Extensions: extensions,
	}, nil
}

// runMigration runs a migration on a schema.
// this should be called whenever an existing datastore is loaded.
func runMigration(ctx context.Context, dbid string, datastore Registry) error {
	kv := &metadataKv{
		registry: datastore,
		dbid:     dbid,
		sync:     true,
	}

	return metadata.RunMigration(ctx, kv)
}
