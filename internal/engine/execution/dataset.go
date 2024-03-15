package execution

import (
	"fmt"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
)

// baseDataset is a deployed database schema.
// It implements the precompiles.Instance interface.
type baseDataset struct {
	// schema is the schema of the dataset.
	schema *types.Schema

	// extensions are the extensions available for use in the dataset.
	extensions map[string]precompiles.Instance

	// actions are the actions that are available for use in the dataset.
	actions map[string]*preparedAction

	// procedures are the procedures that are available for use in the dataset.
	// It only includes public procedures.
	procedures map[string]*preparedProcedure

	// global is the global context.
	global *GlobalContext
}

var _ precompiles.Instance = (*baseDataset)(nil)

// Call calls a procedure from the dataset.
// If the procedure is not public, it will return an error.
// It satisfies precompiles.Instance.
func (d *baseDataset) Call(caller *precompiles.ProcedureContext, app *common.App, method string, inputs []any) ([]any, error) {
	// check if it is a procedure
	proc, ok := d.procedures[method]
	if ok {
		inputs, err := proc.coerceInputs(inputs)
		if err != nil {
			return nil, err
		}

		res, err := app.DB.Execute(caller.Ctx, proc.callString(caller.Procedure), inputs...)
		if err != nil {
			return nil, err
		}

		caller.Result = res
		return nil, nil
	}

	// otherwise, it is an action
	act, ok := d.actions[method]
	if !ok {
		return nil, fmt.Errorf(`procedure "%s" not found`, method)
	}

	if !act.public {
		return nil, fmt.Errorf(`procedure "%s" is not public`, method)
	}

	newCtx := caller.NewScope()
	newCtx.DBID = d.schema.DBID()
	newCtx.Procedure = method

	err := act.call(newCtx, d.global, app.DB, inputs)
	if err != nil {
		return nil, err
	}

	caller.Result = newCtx.Result

	// we currently do not support returning values from dataset procedures
	// if we do, then we will need to return the result here
	return nil, nil
}
