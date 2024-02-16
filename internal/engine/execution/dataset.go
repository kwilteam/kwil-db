package execution

import (
	"fmt"

	"context"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/sql"
)

// Dataset is a deployed database schema.
// It has a schema, procedures that are ready to be called, and can execute statements.
type Dataset interface {
	// Call calls an action from the dataset.
	// If the action is not public, it will return an error.
	Call(caller *ProcedureContext, method string, inputs []any) ([]any, error)
	// Execute executes a statement on the dataset.
	// It understands Kwil's SQL syntax, and gives the same determinism guarantees
	// as any SQL written in Kuneiform.
	// It will use the exec interface to execute the statement.
	Execute(ctx context.Context, exec sql.Executor, stmt string, params map[string]any) (*sql.ResultSet, error)
	// Schema returns the schema of the dataset.
	Schema() *types.Schema
}

// baseDataset is a deployed database schema.
// It implements the Dataset interface.
type baseDataset struct {
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

func (d *baseDataset) Execute(ctx context.Context, exec sql.Executor, stmt string, params map[string]any) (*sql.ResultSet, error) {
	analyzed, err := sqlanalyzer.ApplyRules(stmt, sqlanalyzer.AllRules, d.schema.Tables,
		dbidSchema(d.schema.DBID()))
	if err != nil {
		return nil, fmt.Errorf("error analyzing statement: %w", err)
	}

	orderedParams, err := orderAndCleanValueMap(params, analyzed.ParameterOrder)
	if err != nil {
		return nil, fmt.Errorf("error ordering parameters: %w", err)
	}

	return exec.Execute(ctx, analyzed.Statement, orderedParams)
}

func (d *baseDataset) Schema() *types.Schema {
	return d.schema
}
