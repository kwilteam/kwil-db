package engine

import (
	"context"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset3"
	metadataDB "github.com/kwilteam/kwil-db/pkg/engine/db"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"go.uber.org/zap"
)

type Schema struct {
	Extensions []*types.Extension
	Tables     []*types.Table
	Procedures []*types.Procedure
}

// CreateDataset creates a new dataset
func (e *Engine) CreateDataset(ctx context.Context, name string, owner string, schema *Schema) (finalErr error) {
	dbid := utils.GenerateDBID(name, owner)
	_, ok := e.datasets[dbid]
	if ok {
		return fmt.Errorf("%w: %s", ErrDatasetExists, dbid)
	}

	err := e.master.RegisterDataset(ctx, name, owner)
	if err != nil {
		return fmt.Errorf("failed to register dataset: %w", err)
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

	dataset, err := e.buildNewDataset(ctx, name, owner, schema)
	if err != nil {
		return fmt.Errorf("failed to build new dataset: %w", err)
	}

	e.datasets[dbid] = dataset

	return nil

}

// buildNewDataset builds a new datastore, and puts it in a dataset
func (e *Engine) buildNewDataset(ctx context.Context, name string, owner string, schema *Schema) (dataset *dataset3.Dataset, finalErr error) {
	dbid := utils.GenerateDBID(name, owner)
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

	return dataset3.Builder().
		WithDatastore(&datasetDBAdapter{db}).
		WithProcedures(schema.Procedures...).
		WithTables(schema.Tables...).
		WithInitializers(e.getInitializers()).
		WithExtensions(schema.Extensions...).
		OwnedBy(owner).
		Named(name).
		Build(ctx)
}

func (e *Engine) GetDataset(ctx context.Context, dbid string) (Dataset, error) {
	dataset, ok := e.datasets[dbid]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDatasetNotFound, dbid)
	}

	return dataset, nil
}
