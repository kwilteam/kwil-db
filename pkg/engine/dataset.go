package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	metadataDB "github.com/kwilteam/kwil-db/pkg/engine/db"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"go.uber.org/zap"
)

// CreateDataset creates a new dataset
func (e *Engine) CreateDataset(ctx context.Context, schema *types.Schema, owner types.UserIdentifier) (dbid string, finalErr error) {
	user, err := newDatasetUser(owner)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	err = schema.Clean()
	if err != nil {
		return "", fmt.Errorf("failed to clean schema: %w", err)
	}

	schema.Owner = user.PubKey()

	dbid = GenerateDBID(schema.Name, schema.Owner)
	_, ok := e.datasets[dbid]
	if ok {
		return dbid, fmt.Errorf("%w: %s", ErrDatasetExists, dbid)
	}

	err = e.master.RegisterDataset(ctx, schema.Name, owner)
	if err != nil {
		return dbid, fmt.Errorf("failed to register dataset: %w", err)
	}
	defer func() {
		if err != nil {
			err2 := e.master.UnregisterDataset(ctx, dbid)
			if err2 != nil {
				e.log.Warn("failed to unregister dataset: %s", zap.String("dbid", dbid))
			}

			finalErr = errors.Join(err, err2)
		}
	}()

	ds, err := e.buildNewDataset(ctx, schema.Name, user, schema)
	if err != nil {
		return dbid, fmt.Errorf("failed to build new dataset: %w", err)
	}

	e.datasets[dbid] = ds
	return dbid, nil

}

// buildNewDataset builds a new datastore, and puts it in a dataset
func (e *Engine) buildNewDataset(ctx context.Context, name string, owner *datasetUser, schema *types.Schema) (ds *dataset.Dataset, finalErr error) {

	dbid := GenerateDBID(name, owner.PubKey())

	// we do not use the private open method since that registers the dataset
	// this is because we need to execute all ddl before registering the dataset
	// this needs to get fixed in the new engine upgrade.
	datastore, err := e.opener.Open(dbid, e.log)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	defer func() {
		if err != nil {
			err2 := datastore.Delete()
			if err2 != nil {
				e.log.Warn("failed to delete database: %s", zap.String("dbid", dbid))
			}

			finalErr = errors.Join(err, err2)
		}
	}()

	db, err := metadataDB.NewDB(ctx, datastore)
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata db: %w", err)
	}

	ds, err = dataset.Builder().
		WithDatastore(db).
		WithProcedures(schema.Procedures...).
		WithTables(schema.Tables...).
		WithInitializers(e.getInitializers()).
		WithExtensions(schema.Extensions...).
		OwnedBy(owner).
		Named(name).
		WithLogger(e.log).
		Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build dataset: %w", err)
	}

	// now we register
	err = e.commitRegister.Register(ctx, dbid, datastore)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func (e *Engine) GetDataset(ctx context.Context, dbid string) (Dataset, error) {
	ds, ok := e.datasets[dbid]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDatasetNotFound, dbid)
	}

	return ds, nil
}

func (e *Engine) DropDataset(ctx context.Context, dbid string, sender types.UserIdentifier) error {
	ds, ok := e.datasets[dbid]
	if !ok {
		return fmt.Errorf("%w: %s", ErrDatasetNotFound, dbid)
	}

	senderPub, err := sender.PubKey()
	if err != nil {
		return fmt.Errorf("failed to get sender public key: %w", err)
	}

	_, owner := ds.Metadata()
	if !bytes.Equal(owner.PubKey(), senderPub.Bytes()) {
		return fmt.Errorf("%w: %s", ErrDatasetNotOwned, dbid)
	}

	// we call unregister first so the session can be canceled, before the database is deleted
	err = e.commitRegister.Unregister(ctx, dbid)
	if err != nil {
		return fmt.Errorf("failed to unregister dataset: %w", err)
	}

	err = ds.Delete()
	if err != nil {
		return fmt.Errorf("failed to close dataset: %w", err)
	}

	delete(e.datasets, dbid)

	err = e.master.UnregisterDataset(ctx, dbid)
	if err != nil {
		return fmt.Errorf("failed to unregister dataset: %w", err)
	}

	return nil
}
