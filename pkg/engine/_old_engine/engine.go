package engine

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	"github.com/kwilteam/kwil-db/pkg/engine/dto"
	"github.com/kwilteam/kwil-db/pkg/engine/extensions"
	"github.com/kwilteam/kwil-db/pkg/engine/sqldb"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"github.com/kwilteam/kwil-db/pkg/log"
	"go.uber.org/zap"
)

type Engine interface {
	// NewDataset creates a new dataset.
	// TODO: this should probably be a builder, since many of the options are required.
	NewDataset(ctx context.Context, opts ...CreateDatasetOpt) (Dataset, error)

	// GetDatastore returns a datastore for the given dataset.
	GetDataset(dbid string) (Dataset, error)

	// Close closes the engine.
	// If closeAll is true, it will also close all datasets.
	Close(closeAll bool) error

	// DeleteDataset deletes a dataset.  The caller of the txCtx must be the owner of the dataset.
	DeleteDataset(ctx context.Context, txCtx *dto.TxContext, dbid string) error

	// Delete deletes the master database.  If true is passed, it will also delete all deployed datasets.
	Delete(deleteAll bool) error

	// ListDatasets lists the datasets for the given owner.
	ListDatasets(ctx context.Context, owner string) ([]string, error)
}

type engine struct {
	db          datastore
	name        string
	path        string
	log         log.Logger
	datasets    map[string]internalDataset
	wipeOnStart bool
	extensions  map[string]*extensions.Extension
}

func Open(ctx context.Context, opts ...EngineOpt) (Engine, error) {
	e := &engine{
		name:        defaultName,
		log:         log.NewNoOp(),
		datasets:    make(map[string]internalDataset),
		wipeOnStart: false,
		extensions:  map[string]*extensions.Extension{},
	}

	for _, opt := range opts {
		opt(e)
	}

	err := e.openMasterDB()
	if err != nil {
		return nil, err
	}

	err = e.initTables(ctx)
	if err != nil {
		return nil, err
	}

	err = e.openStoredDatasets(ctx)
	if err != nil {
		return nil, err
	}

	err = e.connectExtensions(ctx)
	if err != nil {
		return nil, err
	}

	return e, nil
}

// openDB opens a database connections and wraps it in a sqldb.DB.
// This should probably be done in a different package to avoid coupling to sqldb.
func (e *engine) openDB(name string) (*sqldb.SqliteStore, error) {
	opts := []sqldb.SqliteOpts{
		sqldb.WithGlobalVariables(dto.GlobalVars),
		sqldb.WithLogger(e.log),
	}
	if e.path != "" {
		opts = append(opts, sqldb.WithPath(e.path))
	}

	return sqldb.NewSqliteStore(name,
		opts...,
	)
}

// openMasterDB opens the master database.
// if wipeOnStart is true, it will open the database, delete it, and then reopen it.
func (e *engine) openMasterDB() error {
	if e.wipeOnStart {
		db, err := e.openDB(e.name)
		if err != nil {
			return err
		}

		err = db.Delete()
		if err != nil {
			return err
		}
	}

	db, err := e.openDB(e.name)
	if err != nil {
		return err
	}

	e.db = db
	return nil
}

func (e *engine) openStoredDatasets(ctx context.Context) error {
	storedDatasets, err := e.listDatasets(ctx)
	if err != nil {
		return err
	}

	for _, storedDataset := range storedDatasets {
		dbid := utils.GenerateDBID(storedDataset.name, storedDataset.owner)

		db, err := e.openDB(dbid)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		e.datasets[dbid], err = dataset.OpenDataset(ctx, db,
			dataset.WithName(storedDataset.name),
			dataset.WithOwner(storedDataset.owner),
			dataset.WithExtensionInitializers(e.extensionInitializers()),
		)

		if err != nil {
			return fmt.Errorf("failed to open dataset: %w", err)
		}
	}

	return nil
}

func (e *engine) connectExtensions(ctx context.Context) error {
	for name, ext := range e.extensions {
		err := ext.Connect(ctx)
		if err != nil {
			return fmt.Errorf("failed to connect extension %s: %w", name, err)
		}
	}

	return nil
}

func (e *engine) Close(closeAll bool) error {
	if closeAll {
		for _, ds := range e.datasets {
			err := ds.Close()
			if err != nil {
				return err
			}
		}
	}

	return e.db.Close()
}

// GetDataset retrieves a dataset if it exists
func (e *engine) GetDataset(dbid string) (Dataset, error) {
	ds, ok := e.datasets[dbid]
	if !ok {
		return nil, fmt.Errorf("dataset %s does not exist", dbid)
	}

	return ds, nil
}

// DeleteDataset deletes a dataset
func (e *engine) DeleteDataset(ctx context.Context, txCtx *dto.TxContext, dbid string) error {
	ds, ok := e.datasets[dbid]
	if !ok {
		return fmt.Errorf("dataset %s does not exist", dbid)
	}

	err := ds.Delete(txCtx)
	if err != nil {
		return err
	}

	err = e.unregisterDataset(ctx, dbid)
	if err != nil {
		e.log.Error("failed to unregister dataset after deletion", zap.Error(err), zap.String("dbid", dbid))
		return err
	}

	delete(e.datasets, dbid)

	return nil
}

func (d *engine) Delete(deleteAll bool) error {
	if deleteAll {
		for dbid, ds := range d.datasets {
			err := ds.Delete(&dto.TxContext{
				Caller: ds.Owner(),
			})
			if err != nil {
				d.log.Error("failed to delete dataset", zap.Error(err), zap.String("dbid", dbid))
			}

			err = d.unregisterDataset(context.Background(), dbid)
			if err != nil {
				d.log.Error("failed to unregister dataset after deletion", zap.Error(err), zap.String("dbid", dbid))
			}

			delete(d.datasets, dbid)
		}
	}

	return d.db.Delete()
}

func (e *engine) ListDatasets(ctx context.Context, owner string) ([]string, error) {
	return e.listDatasetsByOwner(ctx, owner)
}

// function that implements dataset.Initializer
type extenstionInitializer func(ctx context.Context, metadata map[string]string) (*extensions.Instance, error)

// method for implementing dataset.Initializer
func (f extenstionInitializer) Initialize(ctx context.Context, metadata map[string]string) (dataset.InitializedExtension, error) {
	return f(ctx, metadata)
}

func (e *engine) extensionInitializers() map[string]dataset.Initializer {
	initializers := make(map[string]dataset.Initializer)

	for name, ext := range e.extensions {
		initializers[name] = extenstionInitializer(ext.CreateInstance)
	}

	return initializers
}
