package engine

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	"github.com/kwilteam/kwil-db/pkg/engine/dto"
	"github.com/kwilteam/kwil-db/pkg/engine/sqldb"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"go.uber.org/zap"
)

type CreateDatasetOpt func(*datasetMetadata)

type datasetMetadata struct {
	Tables     []*dto.Table
	Actions    []*dto.Action
	Extensions map[string]map[string]string // extension name -> extension metadata
	Owner      string
	Name       string
}

// WithTables adds tables to the dataset.
func WithTables(tables ...*dto.Table) CreateDatasetOpt {
	return func(e *datasetMetadata) {
		e.Tables = append(e.Tables, tables...)
	}
}

// WithActions adds actions to the dataset.
func WithActions(actions ...*dto.Action) CreateDatasetOpt {
	return func(e *datasetMetadata) {
		e.Actions = append(e.Actions, actions...)
	}
}

// WithExtensions adds extensions to the dataset.
func WithExtensions(extentionMetadata map[string]map[string]string) CreateDatasetOpt {
	return func(e *datasetMetadata) {
		e.Extensions = extentionMetadata
	}
}

// WithOwner sets the owner of the dataset.
func WithOwner(owner string) CreateDatasetOpt {
	return func(e *datasetMetadata) {
		e.Owner = owner
	}
}

// WithName sets the name of the dataset.
func WithName(name string) CreateDatasetOpt {
	return func(e *datasetMetadata) {
		e.Name = name
	}
}

// NewDataset creates a new dataset, caches it, and returns it
func (e *engine) NewDataset(ctx context.Context, opts ...CreateDatasetOpt) (Dataset, error) {
	metadata := &datasetMetadata{
		Tables:     []*dto.Table{},
		Actions:    []*dto.Action{},
		Extensions: map[string]map[string]string{},
		Owner:      "",
		Name:       "",
	}

	for _, opt := range opts {
		opt(metadata)
	}

	if metadata.Owner == "" {
		return nil, fmt.Errorf("owner must be specified")
	}

	if metadata.Name == "" {
		return nil, fmt.Errorf("name must be specified")
	}

	dbid := utils.GenerateDBID(metadata.Name, metadata.Owner)

	_, ok := e.datasets[dbid]
	if ok {
		return nil, fmt.Errorf("dataset %s already exists", dbid)
	}

	db, err := e.openDB(dbid)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	newDataset, err := e.createNewDataset(ctx, db, metadata)
	if err != nil {
		err2 := db.Delete()
		if err2 != nil {
			e.log.Error("failed to delete dataset while cleaning up failure", zap.Error(err2))
		}

		return nil, fmt.Errorf("failed to create dataset: %w", err)
	}

	e.datasets[dbid] = newDataset

	err = e.registerDataset(metadata.Name, metadata.Owner)
	if err != nil {
		e.log.Error("failed to register dataset", zap.Error(err))

		delete(e.datasets, dbid)

		err2 := db.Delete()
		if err2 != nil {
			e.log.Error("failed to delete dataset", zap.Error(err2))
		}

		return nil, fmt.Errorf("failed to store dataset: %w", err)
	}

	return newDataset, nil
}

// createNewDataset initializes the necessary extensions and creates the dataset.
func (e *engine) createNewDataset(ctx context.Context, db sqldb.DB, metadata *datasetMetadata) (internalDataset, error) {
	newDataset, err := dataset.Builder().
		Named(metadata.Name).
		OwnedBy(metadata.Owner).
		WithActions(metadata.Actions...).
		WithTables(metadata.Tables...).
		WithExtensions(metadata.Extensions).
		WithExtensionInitializers(e.extensionInitializers()).
		WithDatastore(db).
		Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open dataset: %w", err)
	}

	return newDataset, nil
}
