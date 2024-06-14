package execution

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/kwilteam/kwil-db/common"
	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/internal/conv"
	"github.com/kwilteam/kwil-db/internal/engine/generate"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
)

// MaxStackDepth is the limit on the number of nested procedure calls allowed.
// This is different from the Go call stack depth, which may be much higher as
// it depends on the program design. The value 1,000 was empirically selected to
// be a call stack size of about 1MB and to provide a very high limit that no
// reasonable schema would exceed (even 100 would suggest a poorly designed
// schema).
//
// In addition to exorbitant memory required to support a call stack 1 million
// deep (>1GB), the execution of that many calls can take seconds, even if they
// do nothing else.
//
// Progressive gas metering may be used in the future to limit resources used by
// abusive recursive calls, but a hard upper limit will likely be necessary
// unless the price of an action call is extremely expensive or rises
// exponentially at each level of the call stack.
const MaxStackDepth = 1000

var (
	ErrIncorrectNumberOfArguments = errors.New("incorrect number of arguments")
	ErrPrivateProcedure           = errors.New("procedure is private")
	ErrMutativeProcedure          = errors.New("procedure is mutative")
	ErrMaxStackDepth              = errors.New("max call stack depth reached")
)

// instruction is an instruction that can be executed.
// It is used to define the behavior of a procedure.
type instruction interface { // i.e. dmlStmt, callMethod, or instructionFunc
	execute(scope *precompiles.ProcedureContext, global *GlobalContext, db sql.DB) error
}

// preparedAction is a predefined action that can be executed.
// Unlike the action declared in the shared types,
// preparedAction's statements are parsed into a set of instructions.
type preparedAction struct {
	// name is the name of the procedure.
	name string

	// public indicates whether the procedure is public or privately scoped.
	public bool

	// parameters are the parameters of the procedure.
	parameters []string

	// view indicates whether the procedure has a `view` tag.
	view bool

	// instructions are the instructions that the procedure executes when called.
	instructions []instruction
}

// prepareActions parses all actions.
// It converts all modifiers and statements into instructions.
// these instructions are then used to execute the action.
// It will convert modifiers first, since these should be checked immediately
// when the action is called. It will then convert the statements into
// instructions.
func prepareActions(schema *types.Schema) ([]*preparedAction, error) {
	owner := make([]byte, len(schema.Owner))
	copy(owner, schema.Owner) // copy this here since caller may modify the passed schema. maybe not necessary

	preparedActions := make([]*preparedAction, len(schema.Actions))

	for idx, action := range schema.Actions {
		instructions := make([]instruction, 0)

		actionStmt, err := generate.GenerateActionBody(action, schema, dbidSchema(schema.DBID()))
		if err != nil {
			return nil, err
		}

		// add instructions for both owner only and view procedures
		if action.IsOwnerOnly() {
			instructions = append(instructions, instructionFunc(func(scope *precompiles.ProcedureContext, global *GlobalContext, db sql.DB) error {
				if !bytes.Equal(scope.Signer, owner) {
					return fmt.Errorf("cannot call owner action, not owner")
				}

				return nil
			}))
		}

		if !action.IsView() {
			instructions = append(instructions, instructionFunc(func(scope *precompiles.ProcedureContext, global *GlobalContext, db sql.DB) error {
				tx, ok := db.(sql.AccessModer)
				if !ok {
					return errors.New("DB does not provide access mode needed for mutative action")
				}
				if tx.AccessMode() != sql.ReadWrite {
					return fmt.Errorf("%w, not in a chain transaction", ErrMutativeProcedure)
				}

				return nil
			}))
		}

		for _, parsedStmt := range actionStmt {
			switch stmt := parsedStmt.(type) {
			default:
				return nil, fmt.Errorf("unknown statement type %T", stmt)
			case *generate.ActionExtensionCall:
				i := &callMethod{
					Namespace: stmt.Extension,
					Method:    stmt.Method,
					Args:      makeExecutables(stmt.Params),
					Receivers: stmt.Receivers,
				}
				instructions = append(instructions, i)
			case *generate.ActionSQL:
				i := &dmlStmt{
					SQLStatement:      stmt.Statement,
					OrderedParameters: stmt.ParameterOrder,
				}
				instructions = append(instructions, i)
			case *generate.ActionCall:

				var calledAction *types.Action
				for _, p := range schema.Actions {
					if p.Name == stmt.Action {
						calledAction = p
						break
					}
				}
				if calledAction == nil {
					return nil, fmt.Errorf(`action "%s" not found`, stmt.Action)
				}

				// we leave the namespace and receivers empty, since action calls can only
				// call actions within the same schema, and actions cannot return values.
				i := &callMethod{
					Method: stmt.Action,
					Args:   makeExecutables(stmt.Params),
				}
				instructions = append(instructions, i)
			}
		}

		preparedActions[idx] = &preparedAction{
			name:         action.Name,
			public:       action.Public,
			parameters:   action.Parameters,
			view:         action.IsView(),
			instructions: instructions,
		}
	}

	return preparedActions, nil
}

// Call executes an action.
func (p *preparedAction) call(scope *precompiles.ProcedureContext, global *GlobalContext, db sql.DB, inputs []any) error {
	if len(inputs) != len(p.parameters) {
		return fmt.Errorf(`%w: action "%s" requires %d arguments, but %d were provided`, ErrIncorrectNumberOfArguments, p.name, len(p.parameters), len(inputs))
	}

	for i, param := range p.parameters {
		scope.SetValue(param, inputs[i])
	}

	for _, inst := range p.instructions {
		if err := inst.execute(scope, global, db); err != nil {
			return err
		}
	}

	return nil
}

// callMethod is a statement that calls a method.
// This can be a local method, or a method from a namespace.
type callMethod struct {
	// Namespace is the namespace that the method is in.
	// If no namespace is specified, the local namespace is used.
	Namespace string

	// Method is the name of the method.
	Method string

	// Args are the arguments to the method.
	// They are evaluated in order, and passed to the method.
	Args []evaluatable
	// for Args we might consider some literals to avoid pointless and error
	// prone evaluation of certain trivial in-line expressions such as `SELECT @arg`;

	// Receivers are the variables that the return values are assigned to.
	Receivers []string
}

var _ instructionFunc = (&callMethod{}).execute

// Execute calls a method from a namespace that is accessible within this dataset.
// If no namespace is specified, the local namespace is used.
// It will pass all arguments to the method, and assign the return values to the receivers.
func (e *callMethod) execute(scope *precompiles.ProcedureContext, global *GlobalContext, db sql.DB) error {
	// This instruction is about to call into another procedure in this dataset
	// or another baseDataset. Check current call stack depth first.
	if scope.StackDepth >= MaxStackDepth {
		// NOTE: the actual Go call stack depth can be much more (e.g. more than
		// double) the procedure call depth depending on program design and the
		// number of Go function calls for each procedure. As of writing, it is
		// approximately double plus a handful from the caller:
		//
		// var pcs [4096]uintptr; fmt.Println("call stack depth", runtime.Callers(0, pcs[:]))
		return ErrMaxStackDepth
	}

	dataset, ok := global.datasets[scope.DBID]
	if !ok {
		return fmt.Errorf("%w: %s", ErrDatasetNotFound, scope.DBID)
	}

	// getting these types to match the type required by the the ultimate DML
	// statement is tricky. stuff like `SELECT $1;` breaks extended query
	// protocol mechanisms or ends up with the return as a string even if it's
	// input as an int like 1. If we decide to be more type-strict, we should
	// consider special Arg types that are literals (pass through functions?)
	// that avoid the round trip to the database. Expressions with arithmetic,
	// unary, binary, etc. operators still need to go through the DB.
	var inputs []any
	vals := scope.Values() // declare here since scope.Values() is expensive
	for _, arg := range e.Args {
		val, err := arg(scope.Ctx, db.Execute, vals)
		if err != nil {
			return err
		}

		inputs = append(inputs, val)
	}

	var results []any
	var err error

	scope.UsedGas += 10
	if scope.UsedGas >= 10000000 {
		return fmt.Errorf("out of gas")
	}

	newScope := scope.NewScope()
	newScope.StackDepth++ // not done by NewScope since (*baseDataset).Call would do it again

	// if no namespace is specified, we call a local procedure.
	// this can access public and private procedures.
	if e.Namespace == "" {
		procedure, ok := dataset.actions[e.Method]
		if !ok {
			return fmt.Errorf(`action "%s" not found`, e.Method)
		}

		err = procedure.call(newScope, global, db, inputs)
	} else {
		namespace, ok := dataset.extensions[e.Namespace]
		if !ok {
			return fmt.Errorf(`namespace "%s" not found`, e.Namespace)
		}

		// new scope since we are calling a namespace
		results, err = namespace.Call(newScope, &common.App{
			Service: global.service,
			DB:      db,
			Engine:  global,
		}, e.Method, inputs)
	}
	if err != nil {
		return err
	}

	scope.Result = newScope.Result

	if len(e.Receivers) > len(results) {
		return fmt.Errorf(`%w: action "%s" returned %d values, but only %d receivers were specified`, ErrIncorrectNumberOfArguments, e.Method, len(results), len(e.Receivers))
	}

	// Make the result available to either subsequent instructions or as the FinalResult.
	for i, result := range results { // fmt.Println("res::", i, e.Receivers[i], result)
		// make sure there is a receiver for the result
		if i >= len(e.Receivers) {
			break
		}

		scope.SetValue(e.Receivers[i], result)
	}

	return nil
}

// dmlStmt is a DML statement, we leave the parsing to sqlparser
type dmlStmt struct {
	// SQLStatement is the transformed, deterministic, Postgres compatible SQL statement.
	SQLStatement string

	// OrderedParameters is the named parameters in the order they need to be passed to the database.
	// Since Postgres doesn't support named parameters, we parse them to positional params, and then
	// pass them to the database in the order they are expected.
	OrderedParameters []string
}

var _ instructionFunc = (&dmlStmt{}).execute

func (e *dmlStmt) execute(scope *precompiles.ProcedureContext, _ *GlobalContext, db sql.DB) error {
	// Expend the arguments based on the ordered parameters for the DML statement.
	params := orderAndCleanValueMap(scope.Values(), e.OrderedParameters)
	// args := append([]any{pg.QueryModeExec}, params...)
	results, err := db.Execute(scope.Ctx, e.SQLStatement, append([]any{pg.QueryModeExec}, params...)...)
	if err != nil {
		return err
	}

	// we need to check for any pg numeric types returned, and convert them to int64
	for i, row := range results.Rows {
		for j, val := range row {
			int64Val, ok := sql.Int64(val)
			if ok {
				results.Rows[i][j] = int64Val
			}
		}
	}

	scope.Result = results

	return nil
}

type instructionFunc func(scope *precompiles.ProcedureContext, global *GlobalContext, db sql.DB) error

// implement instruction
func (f instructionFunc) execute(scope *precompiles.ProcedureContext, global *GlobalContext, db sql.DB) error {
	return f(scope, global, db)
}

// evaluatable is an expression that can be evaluated to a scalar value.
// It is used to handle inline expressions, such as within action calls.
type evaluatable func(ctx context.Context, exec dbQueryFn, values map[string]any) (any, error)

// makeExecutables converts inline expressions into a set of evaluatables.
// These are SQL statements that executed with arguments from previously bound
// values (either from the action call params or results from preceding
// instructions in the procedure), and whose results are used as the input
// arguments for action or extension calls.
//
// See their execution in (*callMethod).execute inside the `range e.Args` to
// collect the `inputs` passed to the call of a dataset method or other
// "namespace" method, such as an extension method.
func makeExecutables(params []*generate.InlineExpression) []evaluatable {
	var evaluatables []evaluatable

	for _, param := range params {
		// copy the param to avoid loop variable capture
		param2 := &generate.InlineExpression{
			Statement:     param.Statement,
			OrderedParams: param.OrderedParams,
		}
		evaluatables = append(evaluatables, func(ctx context.Context, exec dbQueryFn, values map[string]any) (any, error) {
			// we need to start with a slice of the mode key
			// for in-line expressions, we need to use the inferred arg types
			valSlice := []any{pg.QueryModeInferredArgTypes}

			// ordering the map values according to the bind names
			valSlice = append(valSlice, orderAndCleanValueMap(values, param2.OrderedParams)...)

			result, err := exec(ctx, param2.Statement, valSlice...) // more values than binds
			if err != nil {
				return nil, err
			}

			if len(result.Rows) == 0 {
				return nil, nil
			}
			if len(result.Rows) > 1 {
				return nil, fmt.Errorf("expected max 1 row for in-line expression, got %d", len(result.Rows))
			}

			record := result.Rows[0]
			if len(record) != 1 {
				return nil, fmt.Errorf("expected 1 value for in-line expression, got %d", len(record))
			}

			// Kwil supports nils in in-line expressions, so we need to check for nils
			if record[0] == nil {
				return nil, nil
			}
			// there is an edge case here where if the value is an array, it needs to be of the exact array type.
			// For example, pgx only understands []string, and not []any, however it will return arrays to us as
			// []any. If the returned type here is an array, we need to convert it to an array of the correct type.
			typeOf := reflect.TypeOf(record[0])
			if typeOf.Kind() == reflect.Slice && typeOf.Elem().Kind() != reflect.Uint8 {
				// if it is an array, we need to convert it to the correct type.
				// if of length 0, we can simply set it to a text array
				if len(record[0].([]any)) == 0 {
					return []string{}, nil
				}

				switch v := record[0].([]any)[0].(type) {
				case string:
					textArr := make([]string, len(record[0].([]any)))
					for i, val := range record[0].([]any) {
						textArr[i] = val.(string)
					}
					return textArr, nil
				case int64:
					intArr := make([]int64, len(record[0].([]any)))
					for i, val := range record[0].([]any) {
						intArr[i] = val.(int64)
					}
					return intArr, nil
				case []byte:
					blobArr := make([][]byte, len(record[0].([]any)))
					for i, val := range record[0].([]any) {
						blobArr[i] = val.([]byte)
					}
					return blobArr, nil
				case bool:
					boolArr := make([]bool, len(record[0].([]any)))
					for i, val := range record[0].([]any) {
						boolArr[i] = val.(bool)
					}
					return boolArr, nil
				case *types.UUID:
					uuidArr := make(types.UUIDArray, len(record[0].([]any)))
					for i, val := range record[0].([]any) {
						uuidArr[i] = val.(*types.UUID)
					}
					return uuidArr, nil
				case *types.Uint256:
					uint256Arr := make(types.Uint256Array, len(record[0].([]any)))
					for i, val := range record[0].([]any) {
						uint256Arr[i] = val.(*types.Uint256)
					}
					return uint256Arr, nil
				case *decimal.Decimal:
					decArr := make(decimal.DecimalArray, len(record[0].([]any)))
					for i, val := range record[0].([]any) {
						decArr[i] = val.(*decimal.Decimal)
					}
					return decArr, nil
				default:
					return nil, fmt.Errorf("unsupported in-line array type %T", v)
				}
			}

			return record[0], nil
		})
	}

	return evaluatables
}

// orderAndCleanValueMap takes a map of values and a slice of keys, and returns
// a slice of values in the order of the keys. If a value can be converted to an
// int, it will be. If a value does not exist, it will be set to nil.
func orderAndCleanValueMap(values map[string]any, keys []string) []any {
	ordered := make([]any, 0, len(keys))
	for _, key := range keys {
		val, ok := values[key]
		if ok {
			val = cleanseIntValue(val)
		} // leave nil if it doesn't exist, still append

		ordered = append(ordered, val)
	}

	return ordered
}

// cleanseIntValue attempts to coerce a value to an int64.
// bools are not converted.
//
// Client tooling sends everything as a string, and we don't have typing in any
// action arguments or variables. So we have no choice but to attempt to coerce
// a string or other value into an int so that the inline expression, which is
// basically always expecting integer arguments, does not bomb. I don't like
// this a lot, but it's essentially what SQLite did although maybe more
// judiciously depending on the needs of the query?
func cleanseIntValue(val any) any {
	if _, isBool := val.(bool); isBool {
		return val
	}
	intVal, err := conv.Int(val)
	if err == nil {
		return intVal
	}

	return val
}

func prepareProcedure(proc *types.Procedure) (*preparedProcedure, error) {
	return &preparedProcedure{
		name:       proc.Name,
		public:     proc.Public,
		parameters: proc.Parameters,
		ownerOnly:  proc.IsOwnerOnly(),
		view:       proc.IsView(),
		returns:    proc.Returns,
	}, nil
}

// preparedProcedure is a predefined procedure that can be executed.
type preparedProcedure struct {
	// name is the name of the procedure.
	name string

	// public indicates whether the procedure is public or privately scoped.
	public bool
	// ownerOnly indicates whether the procedure is owner only.
	ownerOnly bool

	// parameters are the parameters of the procedure.
	parameters []*types.ProcedureParameter

	// view indicates whether the procedure has a `view` tag.
	view bool

	returns *types.ProcedureReturn
}

func (p *preparedProcedure) callString(schema string) string {
	str := strings.Builder{}
	str.WriteString("SELECT * FROM ")
	str.WriteString(dbidSchema(schema))
	str.WriteString(".")
	str.WriteString(p.name)
	str.WriteString("(")
	for i := range p.parameters {
		if i != 0 {
			str.WriteString(", ")
		}
		str.WriteString(fmt.Sprintf("$%d", i+1))
	}
	str.WriteString(");")

	return str.String()
}

// shapeReturn takes a sql result and ensures it matches the expected return shape
// of the procedure. It will modify the passed result to match the expected shape.
func (p *preparedProcedure) shapeReturn(result *sql.ResultSet) error {
	// in postgres, `select * from proc()`, where proc() returns nothing,
	// will return a single empty column and row. We need to remove this.
	if p.returns == nil {
		result.Columns = nil
		result.Rows = nil
		return nil
	}

	if len(p.returns.Fields) != len(result.Columns) {
		// I'm quite positive this will get caught before the schema is even deployed,
		// but just in case, we should check here.
		return fmt.Errorf("shapeReturn: procedure definition expects result %d columns, but returned %d", len(p.returns.Fields), len(result.Columns))
	}

	for i, col := range p.returns.Fields {
		result.Columns[i] = col.Name

		// if the column is a decimal or a decimal array, we need to convert the values to
		// the specified scale and precision
		if col.Type.Name == types.DecimalStr {
			// if it is an array, we need to convert each value in the array
			if col.Type.IsArray {
				for _, row := range result.Rows {
					if row[i] == nil {
						continue
					}

					arr, ok := row[i].([]any)
					if !ok {
						return fmt.Errorf("shapeReturn: expected decimal array, got %T", row[i])
					}

					for _, v := range arr {
						if v == nil {
							continue
						}
						dec, ok := v.(*decimal.Decimal)
						if !ok {
							return fmt.Errorf("shapeReturn: expected decimal, got %T", dec)
						}
						err := dec.SetPrecisionAndScale(col.Type.Metadata[0], col.Type.Metadata[1])
						if err != nil {
							return err
						}
					}
				}
			} else {
				for _, row := range result.Rows {
					if row[i] == nil {
						continue
					}

					dec, ok := row[i].(*decimal.Decimal)
					if !ok {
						return fmt.Errorf("shapeReturn: expected decimal, got %T", row[i])
					}

					err := dec.SetPrecisionAndScale(col.Type.Metadata[0], col.Type.Metadata[1])
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
