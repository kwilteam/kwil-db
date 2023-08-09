package dataset

import (
	"context"
	"io"

	"github.com/kwilteam/kwil-db/pkg/engine/db"
	"github.com/kwilteam/kwil-db/pkg/engine/execution"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/sql"
)

type Engine interface {
	Close() error
	ExecuteProcedure(ctx context.Context, name string, args []any, opts ...execution.ExecutionOpt) ([]map[string]any, error)
}

type InitializedExtension interface {
	Execute(ctx context.Context, method string, args ...any) ([]any, error)
}

type Initializer interface {
	Initialize(context.Context, map[string]string) (InitializedExtension, error)
}

type Datastore interface {
	Prepare(ctx context.Context, query string) (*db.PreparedStatement, error)
	CreateTable(ctx context.Context, table *types.Table) error
	ListTables(ctx context.Context) ([]*types.Table, error)
	StoreProcedure(ctx context.Context, procedure *types.Procedure) error
	ListProcedures(ctx context.Context) ([]*types.Procedure, error)
	StoreExtension(ctx context.Context, extension *types.Extension) error
	ListExtensions(ctx context.Context) ([]*types.Extension, error)
	Close() error
	Delete() error
	Query(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error)
	Savepoint() (sql.Savepoint, error)
	CreateSession() (sql.Session, error)
	ApplyChangeset(changeset io.Reader) error
}

type initializerWrapper struct {
	Initializer
}

func (i initializerWrapper) Initialize(ctx context.Context, initializeVars map[string]string) (execution.InitializedExtension, error) {
	return i.Initializer.Initialize(ctx, initializeVars)
}

type datastoreWrapper struct {
	Datastore
}

func (d datastoreWrapper) Prepare(ctx context.Context, stmt string) (execution.PreparedStatement, error) {
	return d.Datastore.Prepare(ctx, stmt)
}
