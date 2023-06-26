package dataset3

import (
	"context"
	"io"

	"github.com/kwilteam/kwil-db/pkg/engine/eng"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

type Engine interface {
	// Close closes the engine.
	Close() error
}

type InitializedExtension interface {
	Execute(ctx context.Context, method string, args ...any) ([]any, error)
}

type Initializer interface {
	Initialize(context.Context, map[string]string) (InitializedExtension, error)
}

type Datastore interface {
	eng.Datastore
	CreateTable(ctx context.Context, table *types.Table) error
	ListTables(ctx context.Context) ([]*types.Table, error)
	StoreProcedure(ctx context.Context, procedure *types.Procedure) error
	ListProcedures(ctx context.Context) ([]*types.Procedure, error)
	Close() error
	Delete() error
	Query(ctx context.Context, stmt string, args map[string]any) (io.Reader, error)
}

type PreparedStatement interface {
	// Execute executes a prepared statement with the given arguments.
	Execute(context.Context, map[string]any) (io.Reader, error)

	// Close closes the statement.
	Close() error
}

type Savepoint interface {
	Rollback() error
	Commit() error
}
