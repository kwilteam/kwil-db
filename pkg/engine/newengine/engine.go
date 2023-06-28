/*
this package will replace the current top level engine package
*/
package engine

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/log"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset3"
	metadataDB "github.com/kwilteam/kwil-db/pkg/engine/db"
	"github.com/kwilteam/kwil-db/pkg/engine/master"
)

// TODO: this is a stub. delete it
type IEngine interface {
	CreateDataset(ctx context.Context, name string, owner string, schema *Schema) (finalErr error)
	ListDatasets(ctx context.Context, owner string) ([]string, error)
}

type Engine struct {
	master     MasterDB
	name       string
	path       string
	log        log.Logger
	datasets   map[string]Dataset
	extensions map[string]ExtensionInitializer
}

func Open(ctx context.Context, opts ...EngineOpt) (*Engine, error) {
	e := &Engine{
		name:       masterDBName,
		log:        log.NewNoOp(),
		datasets:   make(map[string]Dataset),
		extensions: make(map[string]ExtensionInitializer),
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
	ds, err := e.openDB(masterDBName)
	if err != nil {
		return err
	}

	e.master, err = master.New(ctx, &masterDbAdapter{ds})
	return err
}

// openStoredDatasets opens all of the datasets that are stored in the master
func (e *Engine) openStoredDatasets(ctx context.Context) error {
	datasets, err := e.master.ListDatasets(ctx)
	if err != nil {
		return err
	}

	for _, dataset := range datasets {
		datastore, err := e.openDB(dataset.DBID)
		if err != nil {
			return err
		}

		db, err := metadataDB.NewDB(ctx, &metadataDBAdapter{datastore})
		if err != nil {
			return err
		}

		ds, err := dataset3.OpenDataset(ctx, &datasetDBAdapter{db},
			dataset3.WithAvailableExtensions(e.getInitializers()),
			dataset3.Named(dataset.Name),
			dataset3.OwnedBy(dataset.Owner),
		)
		if err != nil {
			return err
		}

		e.datasets[dataset.DBID] = ds
	}

	return nil
}

// getInitializers gets all of the initializers for extensions that have been
// added to the engine.
func (e *Engine) getInitializers() map[string]dataset3.Initializer {
	initializers := make(map[string]dataset3.Initializer)
	for name, ext := range e.extensions {
		initializers[name] = &extensionInitializerAdapter{ext}
	}

	return initializers
}
