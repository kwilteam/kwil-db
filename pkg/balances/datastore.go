package balances

import (
	"context"
	"io"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sql/client"
)

type Datastore interface {
	Savepoint() (Savepoint, error)
	Close() error
	// Execute executes a statement.
	Execute(ctx context.Context, stmt string, args map[string]any) error

	// Query executes a query and returns the result.
	Query(ctx context.Context, query string, args map[string]any) ([]map[string]any, error)
	TableExists(ctx context.Context, table string) (bool, error)
	Prepare(stmt string) (PreparedStatement, error)
	ApplyChangeset(reader io.Reader) error
	CreateSession() (Session, error)
}

type Savepoint interface {
	Rollback() error
	Commit() error
}

// Opener is an interface that opens a database.
type Opener interface {
	Open(name, path string, log log.Logger) (Datastore, error)
}

type openerFunc func(name, path string, log log.Logger) (Datastore, error)

func (o openerFunc) Open(name, path string, l log.Logger) (Datastore, error) {
	return o(name, path, l)
}

// DbOpener is a function that opens a database.
// it is the default opener
var dbOpener Opener = openerFunc(func(name, path string, log log.Logger) (Datastore, error) {
	clnt, err := client.NewSqliteStore(name,
		client.WithPath(path),
		client.WithLogger(log),
	)
	if err != nil {
		return nil, err
	}

	return &dbAdapter{clnt}, nil
})

// TODO: this should get deleted once we merge in main

type dbAdapter struct {
	*client.SqliteClient
}

func (s *dbAdapter) Savepoint() (Savepoint, error) {
	return s.SqliteClient.Savepoint()
}

func (s *dbAdapter) Prepare(stmt string) (PreparedStatement, error) {
	return s.SqliteClient.Prepare(stmt)
}

func (s *dbAdapter) CreateSession() (Session, error) {
	return s.SqliteClient.CreateSession()
}

// Session is a session that can be used to execute multiple statements.
type Session interface {
	GenerateChangeset() ([]byte, error)
	Delete() error
}

type PreparedStatement interface {
	Close() error
	Execute(ctx context.Context, args map[string]any) ([]map[string]any, error)
}
