/*
this package will replace the current top level engine package
*/
package engine

import (
	"context"
	"errors"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sql"

	"math/big"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	metadataDB "github.com/kwilteam/kwil-db/pkg/engine/db"
	"github.com/kwilteam/kwil-db/pkg/engine/master"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

// The engine is the top level object that manages all of the datasets and
// extensions. It is responsible for opening, closing, and executing against the master database,
// and all of the stored datasets.
type Engine struct {
	// master is the master database, which stores all of the metadata about created datasets
	master MasterDB

	// name is the file name of the master database
	name string

	// log is the logger for the engine
	log log.Logger

	// datasets is a map of all of the datasets that are currently stored in the engine
	datasets map[string]Dataset

	// extensions is a map of all of the extensions that have been added to the engine
	extensions map[string]ExtensionInitializer

	// opener is the function that is used to open sqlite databases
	opener sql.Opener
}

// Open opens a new engine with the provided options.
// It will also open any stored datasets.
func Open(ctx context.Context, dbOpener sql.Opener, opts ...EngineOpt) (*Engine, error) {
	e := &Engine{
		name:       masterDBName,
		log:        log.NewNoOp(),
		datasets:   make(map[string]Dataset),
		extensions: make(map[string]ExtensionInitializer),
		opener:     dbOpener,
	}

	for _, opt := range opts {
		opt(e)
	}

	err := e.openMasterDB(ctx)
	if err != nil {
		return nil, err
	}

	err = e.openStoredDatasets(ctx)
	if err != nil {
		return nil, err
	}

	return e, nil
}

// openMasterDB opens the master database
func (e *Engine) openMasterDB(ctx context.Context) error {
	ds, err := e.opener.Open(e.name, e.log)
	if err != nil {
		return err
	}

	e.master, err = master.New(ctx, ds)
	return err
}

// openStoredDatasets opens all of the datasets that are stored in the master
func (e *Engine) openStoredDatasets(ctx context.Context) error {
	datasets, err := e.master.ListDatasets(ctx)
	if err != nil {
		return err
	}

	for _, datasetInfo := range datasets {
		datastore, err := e.opener.Open(datasetInfo.DBID, e.log)
		if err != nil {
			return err
		}

		db, err := metadataDB.NewDB(ctx, datastore)
		if err != nil {
			return err
		}

		ds, err := dataset.OpenDataset(ctx, db,
			dataset.WithAvailableExtensions(e.getInitializers()),
			dataset.Named(datasetInfo.Name),
			dataset.OwnedBy(datasetInfo.Owner),
			dataset.OpenWithMissingExtensions(),
			dataset.WithLogger(e.log),
		)
		if err != nil {
			return err
		}

		e.datasets[datasetInfo.DBID] = ds
	}

	return nil
}

// getInitializers gets all of the initializers for extensions that have been
// added to the engine.
func (e *Engine) getInitializers() map[string]dataset.Initializer {
	initializers := make(map[string]dataset.Initializer)
	for name, ext := range e.extensions {
		initializers[name] = &extensionInitializerAdapter{ext}
	}

	return initializers
}

// Execute executes a procedure on a database.
// It returns the result of the procedure.
// It takes a context, the database id, the procedure name, the arguments, and optionally some options.
func (e *Engine) Execute(ctx context.Context, dbid string, procedure string, args [][]any, opts ...ExecutionOpt) ([]map[string]any, error) {
	options := &executionConfig{}
	for _, opt := range opts {
		opt(options)
	}

	ds, ok := e.datasets[dbid]
	if !ok {
		return nil, ErrDatasetNotFound
	}

	if len(args) == 0 {
		args = append(args, []any{})
	}

	if options.ReadOnly {

		return ds.Call(ctx, procedure, args[0], &dataset.TxOpts{
			Caller: options.Sender,
		})
	}

	return ds.Execute(ctx, procedure, args, &dataset.TxOpts{
		Caller: options.Sender,
	})
}

// Close closes the engine and all of the stored datasets.
func (e *Engine) Close() error {
	var errs []error
	for _, ds := range e.datasets {
		err := ds.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	err := e.master.Close()
	if err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (e *Engine) GetAllDatasets() ([]string, error) {
	var datasets []string
	for dbid := range e.datasets {
		datasets = append(datasets, dbid)
	}

	return datasets, nil
}

// ListDatasets lists all of the datasets that were deployed by the provided owner.
func (e *Engine) ListDatasets(ctx context.Context, owner string) ([]string, error) {
	dsInfo, err := e.master.ListDatasets(ctx)
	if err != nil {
		return nil, err
	}

	var datasets []string
	for _, info := range dsInfo {
		if strings.EqualFold(info.Owner, owner) {
			datasets = append(datasets, info.Name)
		}
	}

	return datasets, nil
}

// Query executes a query on a database.
// It returns the result of the query.
func (e *Engine) Query(ctx context.Context, dbid string, query string) ([]map[string]any, error) {
	ds, ok := e.datasets[dbid]
	if !ok {
		return nil, ErrDatasetNotFound
	}

	return ds.Query(ctx, query, nil)
}

// GetSchema gets the schema of a database.
func (e *Engine) GetSchema(ctx context.Context, dbid string) (*types.Schema, error) {
	ds, ok := e.datasets[dbid]
	if !ok {
		return nil, ErrDatasetNotFound
	}

	name, owner := ds.Metadata()

	tables, err := ds.ListTables(ctx)
	if err != nil {
		return nil, err
	}

	extensions, err := ds.ListExtensions(ctx)
	if err != nil {
		return nil, err
	}

	return &types.Schema{
		Name:       name,
		Owner:      owner,
		Tables:     tables,
		Extensions: extensions,
		Procedures: ds.ListProcedures(),
	}, nil
}

func (e *Engine) PriceDeploy(ctx context.Context, schema *types.Schema) (price *big.Int, err error) {
	return big.NewInt(0), nil
}

func (e *Engine) PriceDrop(ctx context.Context, dbid string) (price *big.Int, err error) {
	return big.NewInt(0), nil
}

func (e *Engine) PriceExecute(ctx context.Context, dbid string, action string, params []map[string]any) (price *big.Int, err error) {
	return big.NewInt(0), nil
}
