package execution

import (
	"fmt"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
)

// baseDataset is a deployed database schema.
// It implements the Dataset interface.
type baseDataset struct {
	// schema is the schema of the dataset.
	schema *common.Schema

	// namespaces are the namespaces available for use in the dataset.
	namespaces map[string]precompiles.Instance

	// procedures are the procedures that are available for use in the dataset.
	procedures map[string]*procedure

	// global is the global context.
	global *GlobalContext
}

// Call calls a procedure from the dataset.
// If the procedure is not public, it will return an error.
// It implements the Namespace interface.
func (d *baseDataset) Call(caller *precompiles.ProcedureContext, app *common.App, method string, inputs []any) ([]any, error) {
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

	err := proc.call(newCtx, d.global, app.DB, inputs)
	if err != nil {
		return nil, err
	}

	caller.Result = newCtx.Result

	// we currently do not support returning values from dataset procedures
	// if we do, then we will need to return the result here
	return nil, nil
}
