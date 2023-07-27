/*
this package will replace the current top level engine package
*/
package engine

import (
	"context"
	"errors"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/log"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	metadataDB "github.com/kwilteam/kwil-db/pkg/engine/db"
	"github.com/kwilteam/kwil-db/pkg/engine/master"
)

// TODO: this is a stub. delete it
type IEngine interface {
	CreateDataset(ctx context.Context, name string, owner string, schema *Schema) (finalErr error)
	ListDatasets(ctx context.Context, owner string) ([]string, error)
	Close() error
}

type Engine struct {
	master     MasterDB
	name       string
	path       string
	log        log.Logger
	datasets   map[string]Dataset
	extensions map[string]ExtensionInitializer

	curBlockHeight int64           // Tracks the current block height of the blockchain
	ModifiedDBs    map[string]bool // Tracks which databases have been modified in the current block, to be reset at the end of the block

	opener Opener
}

func Open(ctx context.Context, opts ...EngineOpt) (*Engine, error) {
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

func (e *Engine) openDB(name string) (Datastore, error) {
	return e.opener.Open(name, e.path, e.log)
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

	for _, datasetInfo := range datasets {
		datastore, err := e.openDB(datasetInfo.DBID)
		if err != nil {
			return err
		}

		db, err := metadataDB.NewDB(ctx, &metadataDBAdapter{datastore})
		if err != nil {
			return err
		}

		ds, err := dataset.OpenDataset(ctx, &datasetDBAdapter{db},
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

func (e *Engine) SetCurrentBlockHeight(height int64) {
	e.curBlockHeight = height
}

func (e *Engine) GetCurrentBlockHeight() int64 {
	return e.curBlockHeight
}

func (e *Engine) AddDbToModifiedList(dbid string) {
	e.ModifiedDBs[dbid] = true
}
