package dataset

import (
	"context"
	"io"

	"github.com/kwilteam/kwil-db/pkg/engine/execution"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
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
	Prepare(stmt string) (Statement, error)
	CreateTable(ctx context.Context, table *types.Table) error
	ListTables(ctx context.Context) ([]*types.Table, error)
	StoreProcedure(ctx context.Context, procedure *types.Procedure) error
	ListProcedures(ctx context.Context) ([]*types.Procedure, error)
	StoreExtension(ctx context.Context, extension *types.Extension) error
	ListExtensions(ctx context.Context) ([]*types.Extension, error)
	Close() error
	Delete() error
	Query(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error)
	Savepoint() (Savepoint, error)
	CreateSession() (Session, error)
	ApplyChangeset(changeset io.Reader) error
}

type Statement interface {
	Execute(ctx context.Context, args map[string]any) ([]map[string]any, error)
	Close() error
}

type Savepoint interface {
	Rollback() error
	Commit() error
	CommitAndCheckpoint() error
}

type Session interface {
	GenerateChangeset() ([]byte, error)
	Delete()
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

func (d datastoreWrapper) Prepare(stmt string) (execution.PreparedStatement, error) {
	return d.Datastore.Prepare(stmt)
}
