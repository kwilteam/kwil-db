package engine

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	metadataDB "github.com/kwilteam/kwil-db/pkg/engine/db"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"go.uber.org/zap"
)

type Schema struct {
	Extensions []*types.Extension
	Tables     []*types.Table
	Procedures []*types.Procedure
}

// CreateDataset creates a new dataset
func (e *Engine) CreateDataset(ctx context.Context, name string, owner string, schema *Schema) (dbid string, finalErr error) {
	dbid = GenerateDBID(name, owner)
	_, ok := e.datasets[dbid]
	if ok {
		return dbid, fmt.Errorf("%w: %s", ErrDatasetExists, dbid)
	}

	err := e.master.RegisterDataset(ctx, name, owner)
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

	ds, err := e.buildNewDataset(ctx, name, owner, schema)
	if err != nil {
		return dbid, fmt.Errorf("failed to build new dataset: %w", err)
	}

	e.datasets[dbid] = ds

	return dbid, nil

}

// buildNewDataset builds a new datastore, and puts it in a dataset
func (e *Engine) buildNewDataset(ctx context.Context, name string, owner string, schema *Schema) (ds *dataset.Dataset, finalErr error) {
	dbid := GenerateDBID(name, owner)
	datastore, err := e.openDB(dbid)
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

	db, err := metadataDB.NewDB(ctx, &metadataDBAdapter{datastore})
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata db: %w", err)
	}

	ds, err = dataset.Builder().
		WithDatastore(&datasetDBAdapter{db}).
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

	return ds, nil
}

func (e *Engine) GetDataset(ctx context.Context, dbid string) (Dataset, error) {
	ds, ok := e.datasets[dbid]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDatasetNotFound, dbid)
	}

	return ds, nil
}

func (e *Engine) DropDataset(ctx context.Context, sender, dbid string) error {
	ds, ok := e.datasets[dbid]
	if !ok {
		return fmt.Errorf("%w: %s", ErrDatasetNotFound, dbid)
	}

	_, owner := ds.Metadata()
	if !strings.EqualFold(owner, sender) {
		return fmt.Errorf("%w: %s", ErrDatasetNotOwned, dbid)
	}

	err := ds.Delete()
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

func (e *Engine) BlockDBSavepoint(dbid string) error {
	ds, ok := e.datasets[dbid]
	if !ok {
		return fmt.Errorf("%w: %s", ErrDatasetNotFound, dbid)
	}

	begin, err := ds.BlockSavepoint(e.curBlockHeight)
	if begin && err == nil {
		e.AddDbToModifiedList(dbid)
		return nil
	}
	return err
}
