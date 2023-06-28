package engine

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset3"
	metadataDB "github.com/kwilteam/kwil-db/pkg/engine/db"
	"github.com/kwilteam/kwil-db/pkg/engine/master"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sql/client"
)

// theres gotta be a better way to return savepoints and prepares than to have a bunch of adapters

// DbOpener is a function that opens a database.
// It is exported so that it can be overridden in tests.
var DbOpener Opener = openerFunc(func(name, path string, log log.Logger) (Datastore, error) {
	clnt, err := client.NewSqliteStore(name,
		client.WithPath(path),
		client.WithLogger(log),
	)
	if err != nil {
		return nil, err
	}

	return &sqliteClientAdapter{clnt}, nil
})

type Opener interface {
	Open(name, path string, log log.Logger) (Datastore, error)
}

type openerFunc func(name, path string, log log.Logger) (Datastore, error)

func (o openerFunc) Open(name, path string, l log.Logger) (Datastore, error) {
	return o(name, path, l)
}

func (e *Engine) openDB(name string) (Datastore, error) {
	return DbOpener.Open(name, e.path, e.log)
}

type sqliteClientAdapter struct {
	*client.SqliteClient
}

func (s *sqliteClientAdapter) Savepoint() (Savepoint, error) {
	return s.SqliteClient.Savepoint()
}

func (s *sqliteClientAdapter) Prepare(query string) (Statement, error) {
	return s.SqliteClient.Prepare(query)
}

type masterDbAdapter struct {
	Datastore
}

func (m *masterDbAdapter) Savepoint() (master.Savepoint, error) {
	return m.Datastore.Savepoint()
}

type metadataDBAdapter struct {
	Datastore
}

func (m metadataDBAdapter) Savepoint() (metadataDB.Savepoint, error) {
	return m.Datastore.Savepoint()
}

func (m metadataDBAdapter) Prepare(query string) (metadataDB.Statement, error) {
	return m.Datastore.Prepare(query)
}

type datasetDBAdapter struct {
	*metadataDB.DB
}

func (d datasetDBAdapter) Savepoint() (dataset3.Savepoint, error) {
	return d.DB.Savepoint()
}

func (d datasetDBAdapter) Prepare(query string) (dataset3.Statement, error) {
	return d.DB.Prepare(query)
}

type extensionInitializerAdapter struct {
	ExtensionInitializer
}

func (e extensionInitializerAdapter) Initialize(ctx context.Context, meta map[string]string) (dataset3.InitializedExtension, error) {
	return e.ExtensionInitializer.CreateInstance(ctx, meta)
}
