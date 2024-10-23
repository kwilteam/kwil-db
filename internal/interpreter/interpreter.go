package interpreter

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
)

// TODO: move this into the SQL package once pg supports this
type DB interface {
	sql.TxMaker
	Execute(ctx context.Context, stmt string, args ...any) (Rows, error)
}

type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	Err() error
}

// TODO: delete this
// simply using this for planning
type IInterpeter interface {
	// ExecuteRaw executes a raw Postgres SQL statement against the database.
	ExecuteRaw(ctx context.Context, db DB, statement string) (Rows, error)
	// Execute executes Kwil SQL statements against the database.
	ExecuteAdmin(ctx context.Context, db DB, statement string) (Rows, error)
	// Call executes an action against the database.
	Call(ctx *common.TxContext, db DB, action string, args []any) (Rows, error)
}

type Interpreter struct {
	availableFunctions map[string]*executable
}

// NewInterpreter creates a new interpreter.
func NewInterpreter(ctx context.Context, db DB, log log.Logger) (*Interpreter, error) {
	availableFuncs := make(map[string]*executable)

	// we will convert all built-in functions to be executables
	for funcName, impl := range parse.Functions {
		if scalarImpl, ok := impl.(*parse.ScalarFunctionDefinition); ok {
			funcName := funcName // avoid shadowing

			exec := &executable{
				Name: funcName,
				ReturnType: func(v []Value) (*types.ProcedureReturn, error) {
					dataTypes := make([]*types.DataType, len(v))
					for i, arg := range v {
						dataTypes[i] = arg.Type()
					}

					retTyp, err := scalarImpl.ValidateArgsFunc(dataTypes)
					if err != nil {
						return nil, err
					}

					return &types.ProcedureReturn{
						IsTable: false,
						Fields:  []*types.NamedType{{Name: funcName, Type: retTyp}},
					}, nil
				},
				Func: func(ctx context.Context, e *executionContext, args []Value) (Cursor, error) {

				},
			}

			availableFuncs[funcName] = exec
		}
	}

	// we now convert all actions to be executables
	schema, err := getSchema(ctx, db)
	if err != nil {
		return nil, err
	}

	actionSet := make(map[string]struct{})
	for _, action := range schema.Procedures {
		parseRes, err := parse.ParseProcedure(action, schema)
		if err != nil {
			return nil, err
		}
		if parseRes.ParseErrs.Err() != nil {
			return nil, parseRes.ParseErrs.Err()
		}

		exec := makeActionToExecutable(action.Name, parseRes.AST, action.Parameters, action.Returns)

		// we perform this check in case of some incorrectly modified Postgres DB.
		// Not sure if it is needed, might end up deleting this.
		_, ok := actionSet[exec.Name]
		if ok {
			return nil, fmt.Errorf("duplicate action detected on startup: %s", exec.Name)
		}

		_, ok = availableFuncs[exec.Name]
		if ok {
			log.Warnf("built-in PostgreSQL function %s is shadowed by action", exec.Name)
		}
		availableFuncs[exec.Name] = exec
		actionSet[exec.Name] = struct{}{}
	}

	return &Interpreter{
		availableFunctions: availableFuncs,
	}, nil
}

func getSchema(ctx context.Context, db DB) (*types.Schema, error) {
	panic("TODO: implement")
}
