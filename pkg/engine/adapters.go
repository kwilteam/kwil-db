package engine

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sql/client"
)

// theres gotta be a better way to return savepoints and prepares than to have a bunch of adapters

// dbOpener is a function that opens a database.
// it is the default opener for the engine
var dbOpener Opener = openerFunc(func(name, path string, log log.Logger) (Datastore, error) {
	clnt, err := client.NewSqliteStore(name,
		client.WithPath(path),
		client.WithLogger(log),
	)
	if err != nil {
		return nil, err
	}

	return clnt, nil
})

type Opener interface {
	Open(name, path string, log log.Logger) (Datastore, error)
}

type openerFunc func(name, path string, log log.Logger) (Datastore, error)

func (o openerFunc) Open(name, path string, l log.Logger) (Datastore, error) {
	return o(name, path, l)
}

type extensionInitializerAdapter struct {
	ExtensionInitializer
}

func (e extensionInitializerAdapter) Initialize(ctx context.Context, meta map[string]string) (dataset.InitializedExtension, error) {
	return e.ExtensionInitializer.CreateInstance(ctx, meta)
}
