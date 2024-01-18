package execution

import (
	"fmt"

	"context"

	"github.com/kwilteam/kwil-db/internal/engine/types"
	sql "github.com/kwilteam/kwil-db/internal/sql"
)

// Dataset is a deployed database schema.
type Dataset struct {
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
func (d *Dataset) Call(caller *ProcedureContext, method string, inputs []any) ([]any, error) {
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

func (d *Dataset) Execute(ctx context.Context, stmt string, params map[string]any) (*sql.ResultSet, error) {
	return d.readWriter(ctx, stmt, params)
}

func (d *Dataset) Query(ctx context.Context, stmt string, params map[string]any) (*sql.ResultSet, error) {
	return d.read(ctx, stmt, params)
}

func (d *Dataset) Schema() *types.Schema {
	return d.schema
}
