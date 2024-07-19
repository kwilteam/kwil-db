package execution

import (
	"bytes"
	"context"
	"errors"
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
		// *datatypes.ColumnStatistics{min, max, nullcount, ... }
		tblStats, err := buildTableStats(ctx, pgSchema, table.Name, db)
		if err != nil {
			return nil, err
		}
		stats[table.Name] = tblStats
	}
	return stats, nil
}

func buildTableStats(ctx context.Context, pgSchema, table string, db sql.Executor) (*costtypes.Statistics, error) {
	// table stats:
	//  1. row count
	//  2. per-column stats
	//		a. min and max
	//		b. null count
	//		c. unique value count ?
	//		d. average record size ?
	//		e. ???

	qualifiedTable := pgSchema + "." + table

	// row count
	res, err := db.Execute(ctx, `SELECT count(*) FROM %s`, qualifiedTable)
	if err != nil {
		return nil, err
	}
	count, ok := sql.Int64(res.Rows[0][0])
	if !ok {
		return nil, fmt.Errorf("no row count for %s", qualifiedTable)
	}
	// TODO: We needs a schema-table stats database so we don't ever have to do
	// a full table scan for column stats.

	colInfo, err := pg.ColumnInfo(ctx, db, qualifiedTable)
	if err != nil {
		return nil, err
	}
	numCols := len(colInfo)
	colStats := make([]costtypes.ColumnStatistics, numCols)

	// NOTE: this code is not going to be here.  I'm just coding it here so the
	// goal is in focus.

	// iterate over all rows (select *)
	// var scans []any
	// for _, col := range colInfo {
	// 	scans = append(scans, col.ScanVal()) // for QueryRowFunc
	// }
	err = pg.QueryRowFuncAny(ctx, db, `SELECT * FROM `+qualifiedTable,
		func(_ []pg.FieldDesc, vals []any) error {
			for i, val := range vals {
				stat := &colStats[i]
				if val == nil {
					stat.NullCount++
					continue
				}

				if colInfo[i].IsInt() {
					valInt, ok := sql.Int64(val)
					if !ok {
						return errors.New("not int")
					}
					if stat.Min == nil {
						stat.Min = valInt
						stat.Max = valInt
						continue
					}
					if valInt < stat.Min.(int64) {
						stat.Min = valInt
					}
					if valInt > stat.Max.(int64) {
						stat.Max = valInt
					}
					continue
				}

				if colInfo[i].IsText() {
					valStr, ok := val.(string)
					if !ok {
						return errors.New("not string")
					}
					if stat.Min == nil {
						stat.Min = valStr
						stat.Max = valStr
						continue
					}
					if valStr < stat.Min.(string) {
						stat.Min = valStr
					}
					if valStr > stat.Max.(string) {
						stat.Max = valStr
					}
					continue
				}

				// if colInfo[i].IsNumeric() { // TODO

				if colInfo[i].IsByteA() {
					valBytea, ok := val.([]byte)
					if !ok {
						return errors.New("not string")
					}
					if stat.Min == nil {
						stat.Min = valBytea
						stat.Max = valBytea
						continue
					}
					switch bytes.Compare(valBytea, stat.Min.([]byte)) {
					case -1:
						stat.Min = valBytea
					case 1:
						stat.Max = valBytea
					}
					continue
				}

			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	//   datatypes.ColumnStatistics{min, max, nullcount, ... }
	return &costtypes.Statistics{
		RowCount:         count,
		ColumnStatistics: colStats,
	}, nil
}

type ColInfo struct {
	Pos      int
	Name     string
	DataType string
	Nullable bool
	Default  any
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
