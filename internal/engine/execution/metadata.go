package execution

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/metadata"
	"github.com/kwilteam/kwil-db/internal/engine/types"
)

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
	dbid := schema.DBID()
	kv := &metadataKv{
		registry: datastore,
		dbid:     dbid,
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

	exec := func(ctx context.Context, stmt string, params map[string]any) error {
		// NOTE: caller must have prefixed tables in stmt with pg schema!
		var err error
		if len(params) == 0 { // create tables always passes nil, maybe just discard the map input to this closure?
			_, err = datastore.Execute(ctx, dbid, stmt)
		} else {
			fmt.Println("storeSchema: unexpected Execute with non-nil params: ", params)
			_, err = datastore.Execute(ctx, dbid, stmt, params)
		}
		return err
	}
	err = metadata.CreateTables(ctx, dbid, schema.Tables, kv, exec)
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
