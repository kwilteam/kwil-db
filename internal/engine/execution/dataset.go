package execution

import (
	"fmt"

	"context"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	sql "github.com/kwilteam/kwil-db/internal/sql"
)

type Dataset interface {
	Call(caller *ProcedureContext, method string, inputs []any) ([]any, error)
	Execute(ctx context.Context, stmt string, params map[string]any) (*sql.ResultSet, error)
	Query(ctx context.Context, stmt string, params map[string]any) (*sql.ResultSet, error)
	Schema() *types.Schema
}

// baseDataset is a deployed database schema.
// It implements the Dataset interface.
type baseDataset struct {
	// readWrite is a readWriter connection to the dataset.
	readWriter sql.ResultSetFunc
	// read is a read connection to the dataset.
	read sql.ResultSetFunc

	// schema is the schema of the dataset.
	schema *types.Schema

	// namespaces are the namespaces available for use in the dataset.
	namespaces map[string]ExtensionNamespace

	// procedures are the procedures that are available for use in the dataset.
	procedures map[string]*procedure
}

// Call calls a procedure from the dataset.
// If the procedure is not public, it will return an error.
// It implements the Namespace interface.
func (d *baseDataset) Call(caller *ProcedureContext, method string, inputs []any) ([]any, error) {
	proc, ok := d.procedures[method]
	if !ok {
		return nil, fmt.Errorf(`procedure "%s" not found`, method)
	}

	if !proc.public {
		return nil, fmt.Errorf(`procedure "%s" is not public`, method)
	}

	newCtx := caller.NewScope()
	newCtx.DBID = d.schema.DBID()
	newCtx.Procedure = method

	err := proc.call(newCtx, inputs)
	if err != nil {
		return nil, err
	}

	caller.Result = newCtx.Result

	// we currently do not support returning values from dataset procedures
	// if we do, then we will need to return the result here
	return nil, nil
}

func (d *baseDataset) Execute(ctx context.Context, stmt string, params map[string]any) (*sql.ResultSet, error) {
	return d.readWriter(ctx, stmt, params)
}

func (d *baseDataset) Query(ctx context.Context, stmt string, params map[string]any) (*sql.ResultSet, error) {
	return d.read(ctx, stmt, params)
}

func (d *baseDataset) Schema() *types.Schema {
	return d.schema
}

// protectedDataset is a deployed database schema.
// It parses incoming queries to ensure they are deterministic.
// It implements the Dataset interface.
type protectedDataset struct {
	*baseDataset
}

var _ Dataset = (*protectedDataset)(nil)

// Execute executes a statement on the dataset.
func (d *protectedDataset) Execute(ctx context.Context, stmt string, params map[string]any) (*sql.ResultSet, error) {
	// TODO: once we switch to postgres, we will have to switch the named parameters to positional parameters
	analyzed, err := sqlanalyzer.ApplyRules(stmt, sqlanalyzer.AllRules, d.schema.Tables, d.schema.DBID())
	if err != nil {
		return nil, fmt.Errorf("error analyzing statement: %w", err)
	}

	return d.readWriter(ctx, analyzed.Statement(), params)
}

// Query executes a read-only query on the dataset.
func (d *protectedDataset) Query(ctx context.Context, stmt string, params map[string]any) (*sql.ResultSet, error) {
	// TODO: once we switch to postgres, we will have to switch the named parameters to positional parameters
	// usually, we do not need to guarantee determinism for read-only queries.
	// however, we don't actually know if this is being called from a non-mutative context.
	// It is very possible that this is being called from an action that later mutates the database,
	// so we need to guarantee determinism here.
	analyzed, err := sqlanalyzer.ApplyRules(stmt, sqlanalyzer.AllRules, d.schema.Tables, d.schema.DBID())
	if err != nil {
		return nil, fmt.Errorf("error analyzing statement: %w", err)
	}
	if analyzed.Mutative() {
		return nil, fmt.Errorf("extension uses mutative statement in read-only query")
	}

	return d.read(ctx, analyzed.Statement(), params)
}
