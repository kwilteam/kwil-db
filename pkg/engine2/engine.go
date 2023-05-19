package engine2

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine2/dataset"
	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
	"github.com/kwilteam/kwil-db/pkg/engine2/sqldb"
	"github.com/kwilteam/kwil-db/pkg/engine2/sqldb/sqlite"
	"github.com/kwilteam/kwil-db/pkg/engine2/utils"
	"github.com/kwilteam/kwil-db/pkg/log"
)

type Engine interface {
	// NewDataset creates a new dataset.
	NewDataset(ctx context.Context, dsCtx *dataset.DatasetContext) error

	// GetDatastore returns a datastore for the given dataset.
	GetDataset(dbid string) (Dataset, error)

	// Close closes the engine
	Close() error
}

type engine struct {
	db                datastore
	path              string
	log               log.Logger
	datasets          map[string]*dataset.Dataset
	deleteDatasetStmt sqldb.Statement
}

func Open(ctx context.Context, opts ...EngineOpt) (Engine, error) {
	e := &engine{
		path:     defaultPath,
		log:      log.NewNoOp(),
		datasets: make(map[string]*dataset.Dataset),
	}

	for _, opt := range opts {
		opt(e)
	}

	var err error
	e.db, err = e.openDB(e.path)
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

	return e, nil
}

// openDB opens a database connections and wraps it in a sqldb.DB.
// This should probably be done in a different package to avoid coupling to sqlite.
func (e *engine) openDB(name string) (*sqlite.SqliteStore, error) {
	return sqlite.NewSqliteStore(name,
		sqlite.WithPath(e.path),
		sqlite.WithLogger(e.log),
		sqlite.WithGlobalVariables(dto.GlobalVars),
	)
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

		e.datasets[dbid], err = dataset.NewDataset(ctx, &dataset.DatasetContext{
			Name:  storedDataset.name,
			Owner: storedDataset.owner,
		}, db)
		if err != nil {
			return fmt.Errorf("failed to open dataset: %w", err)
		}
	}

	return nil
}

func (e *engine) Close() error {
	for _, ds := range e.datasets {
		err := ds.Close()
		if err != nil {
			return err
		}
	}

	return e.db.Close()
}

func (e *engine) NewDataset(ctx context.Context, dsCtx *dataset.DatasetContext) error {
	dbid := utils.GenerateDBID(dsCtx.Name, dsCtx.Owner)

	_, ok := e.datasets[dbid]
	if ok {
		return fmt.Errorf("dataset %s already exists", dbid)
	}

	db, err := e.openDB(dbid)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	e.datasets[dbid], err = dataset.NewDataset(ctx, dsCtx, db)
	if err != nil {
		return fmt.Errorf("failed to open dataset: %w", err)
	}

	err = e.storeDataset(ctx, dsCtx)
	if err != nil {
		return fmt.Errorf("failed to store dataset: %w", err)
	}

	return nil
}

// getDataset retrieves a dataset if it exists
func (e *engine) GetDataset(dbid string) (Dataset, error) {
	ds, ok := e.datasets[dbid]
	if !ok {
		return nil, fmt.Errorf("dataset %s does not exist", dbid)
	}

	return ds, nil
}
