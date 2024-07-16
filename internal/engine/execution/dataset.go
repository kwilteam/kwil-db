package execution

import (
	"bytes"
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
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

	stats  map[string]*datatypes.Statistics
	fields map[string]*datatypes.Schema
}

var _ precompiles.Instance = (*baseDataset)(nil)

var (
	ErrPrivate   = fmt.Errorf("procedure/action is not public")
	ErrOwnerOnly = fmt.Errorf("procedure/action is owner only")
)

func (d *baseDataset) buildStats(ctx context.Context, db sql.Executor) error {
	// Statistics. Check statistics tables? Recompute full on start?
	pgSchema := dbidSchema(d.schema.DBID())
	for _, table := range d.schema.Tables {
		res, err := db.Execute(ctx, `SELECT count(*) FROM %s.%s`, pgSchema, table.Name)
		if err != nil {
			return err
		}
		count, ok := sql.Int64(res.Rows[0][0])
		if !ok {
			return fmt.Errorf("no row count for %s.%s", pgSchema, table.Name)
		}
		// We needs a schema-table stats database so we don't ever have to do a
		// full table scan for column stats.
		//   datatypes.ColumnStatistics{min, max, nullcount, ... }
		d.stats[table.Name] = &datatypes.Statistics{
			RowCount: count,
			// ColumnStatistics: ,
		}
	}
	return nil
}

func (d *baseDataset) extCall(scope *precompiles.ProcedureContext,
	app *common.App, ext, method string, inputs []any) ([]any, error) {
	ex, ok := d.extensions[ext]
	if !ok {
		return nil, fmt.Errorf(`extension "%s" not found`, ext)
	}
	return ex.Call(scope, app, method, inputs)
}

// Call calls a procedure from the dataset.
// If the procedure is not public, it will return an error.
// It satisfies precompiles.Instance.
func (d *baseDataset) Call(caller *precompiles.ProcedureContext, app *common.App, method string, inputs []any) ([]any, error) {
	// check if it is a procedure
	proc, ok := d.procedures[method]
	if ok {
		if !proc.public {
			return nil, fmt.Errorf(`%w: "%s"`, ErrPrivate, method)
		}
		if proc.ownerOnly && !bytes.Equal(caller.Signer, d.schema.Owner) {
			return nil, fmt.Errorf(`%w: "%s"`, ErrOwnerOnly, method)
		}
		if !proc.view && app.DB.(sql.AccessModer).AccessMode() == sql.ReadOnly {
			return nil, fmt.Errorf(`%w: "%s"`, ErrMutativeProcedure, method)
		}

		// this is not a strictly necessary check, as postgres will throw an error, but this gives a more
		// helpful error message
		if len(inputs) != len(proc.parameters) {
			return nil, fmt.Errorf(`procedure "%s" expects %d argument(s), got %d`, method, len(proc.parameters), len(inputs))
		}

		res, err := app.DB.Execute(caller.Ctx, proc.callString(d.schema.DBID()), append([]any{pg.QueryModeExec}, inputs...)...)
		if err != nil {
			return nil, err
		}

		err = proc.shapeReturn(res)
		if err != nil {
			return nil, err
		}

		caller.Result = res
		return nil, nil
	}

	// otherwise, it is an action
	act, ok := d.actions[method]
	if !ok {
		return nil, fmt.Errorf(`action "%s" not found`, method)
	}

	if !act.public {
		return nil, fmt.Errorf(`%w: "%s"`, ErrPrivate, method)
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
