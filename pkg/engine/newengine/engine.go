/*
this package will replace the current top level engine package
*/
package engine

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/log"

	"github.com/kwilteam/kwil-db/pkg/engine/extensions"
)

type Engine struct {
	db          datastore
	name        string
	path        string
	log         log.Logger
	datasets    map[string]Dataset
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
