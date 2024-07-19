package execution

import (
	"bytes"
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	costtypes "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
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

	// These support the Catalog. Probably move these up to global context.
	// stats  map[string]*datatypes.Statistics
	// fields map[string]*datatypes.Schema
}

var _ precompiles.Instance = (*baseDataset)(nil)

var (
	ErrPrivate   = fmt.Errorf("procedure/action is not public")
	ErrOwnerOnly = fmt.Errorf("procedure/action is owner only")
)

// buildStats refreshes/builds statistics for the tables in the dataset. These
// statistics are used to implement the Catalog used by the query planner for
// cost estimation.
//
// After initial population of the statistics from a full table scan, the
// statistics must be updated for efficiently. TODO: figure out what column
// stats can be updated, and which if any need to use the DB transaction's
// changeset to reestablish completely accurate stats
func (d *baseDataset) buildStats(ctx context.Context, db sql.Executor) (map[string]*costtypes.Statistics, error) {
	return buildStats(ctx, d.schema, db)
}

func buildStats(ctx context.Context, schema *types.Schema, db sql.Executor) (map[string]*costtypes.Statistics, error) {
	// Statistics. Check statistics tables? Recompute full on start?
	stats := map[string]*costtypes.Statistics{}
	pgSchema := dbidSchema(schema.DBID())
	for _, table := range schema.Tables {
		qualifiedTable := pgSchema + "." + table.Name
		tblStats, err := pg.TableStats(ctx, qualifiedTable, db)
		if err != nil {
			return nil, err
		}
		stats[table.Name] = tblStats
	}
	return stats, nil
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
