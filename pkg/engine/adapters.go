package engine

import (
	"context"
	"io"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	metadataDB "github.com/kwilteam/kwil-db/pkg/engine/db"
	"github.com/kwilteam/kwil-db/pkg/engine/master"
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

	return &sqliteClientAdapter{clnt}, nil
})

type Opener interface {
	Open(name, path string, log log.Logger) (Datastore, error)
}

type openerFunc func(name, path string, log log.Logger) (Datastore, error)

func (o openerFunc) Open(name, path string, l log.Logger) (Datastore, error) {
	return o(name, path, l)
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

func (s *sqliteClientAdapter) CreateSession() (Session, error) {
	return s.SqliteClient.CreateSession()
}

func (s *sqliteClientAdapter) ApplyChangeset(changeset io.Reader) error {
	return s.SqliteClient.ApplyChangeset(changeset)
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

func (m metadataDBAdapter) CreateSession() (metadataDB.Session, error) {
	return m.Datastore.CreateSession()
}

func (m metadataDBAdapter) ApplyChangeset(changeset io.Reader) error {
	return m.Datastore.ApplyChangeset(changeset)
}

type datasetDBAdapter struct {
	*metadataDB.DB
}

func (d datasetDBAdapter) Savepoint() (dataset.Savepoint, error) {
	return d.DB.Savepoint()
}

func (d datasetDBAdapter) CreateSession() (dataset.Session, error) {
	return d.DB.CreateSession()
}

func (d datasetDBAdapter) ApplyChangeset(changeset io.Reader) error {
	return d.DB.ApplyChangeset(changeset)
}
func (d datasetDBAdapter) Prepare(query string) (dataset.Statement, error) {
	return d.DB.Prepare(query)
}

type extensionInitializerAdapter struct {
	ExtensionInitializer
}

func (e extensionInitializerAdapter) Initialize(ctx context.Context, meta map[string]string) (dataset.InitializedExtension, error) {
	return e.ExtensionInitializer.CreateInstance(ctx, meta)
}
