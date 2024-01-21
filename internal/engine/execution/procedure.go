package execution

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/common"
	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/internal/conv"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/clean"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/parameters"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	actparser "github.com/kwilteam/kwil-db/parse/action"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
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

// procedure is a predefined procedure that can be executed.
// Unlike the procedure declared in the shared types, this
// procedure's statements are parsed into a set of instructions.
type procedure struct {
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

// prepareProcedure parses a procedure from a types.Procedure.
// It converts all procedure modifiers and statements into instructions.
// these instructions are then used to execute the procedure.
// It will convert modifiers first, since these should be checked immediately
// when the procedure is called. It will then convert the statements into
// instructions.
func prepareProcedure(unparsed *common.Procedure, global *GlobalContext, schema *common.Schema) (*procedure, error) {
	instructions := make([]instruction, 0)
	owner := make([]byte, len(schema.Owner))
	copy(owner, schema.Owner) // copy this here since caller may modify the passed schema. maybe not necessary

	// converting modifiers
	isViewProcedure := false // isViewAction tracks whether this procedure is a view
	for _, mod := range unparsed.Modifiers {
		switch mod {
		case common.ModifierOwner:
			instructions = append(instructions, instructionFunc(func(scope *precompiles.ProcedureContext, global *GlobalContext, db sql.DB) error {

				if !bytes.Equal(scope.Signer, owner) {
					return fmt.Errorf("cannot call owner procedure, not owner")
				}

				return nil
			}))
		case common.ModifierView:
			isViewProcedure = true
		}
	}
	// if not a view action, then the action can only be called from a blockchain tx.
	// This means that the DB connection needs to be readwrite. If not readwrite, we
	// need to return an error
	if !isViewProcedure {
		instructions = append(instructions, instructionFunc(func(scope *precompiles.ProcedureContext, global *GlobalContext, db sql.DB) error {
			tx, ok := db.(sql.AccessModer)
			if !ok {
				return errors.New("DB does not provide access mode needed for mutative action")
			}
			if tx.AccessMode() != sql.ReadWrite {
				return fmt.Errorf("cannot call non-view procedure, not in a chain transaction")
			}

			return nil
		}))
	}

	// converting statements
	// we need to parse each statement with the action parser
	// based on the type of statement, we will convert it into an instruction.
	// If the procedure is a view, then it can neither contain mutative statements
	// nor call non-view procedures.
	for _, stmt := range unparsed.Statements {
		parsedStmt, err := actparser.Parse(stmt)
		if err != nil {
			return nil, err
		}

		switch stmt := parsedStmt.(type) {
		default:
			return nil, fmt.Errorf("unknown statement type %T", stmt)
		case *actparser.ExtensionCallStmt:
			args, err := makeExecutables(stmt.Args)
			if err != nil {
				return nil, err
			}

			receivers := make([]string, len(stmt.Receivers))
			for i, receiver := range stmt.Receivers {
				receivers[i] = strings.ToLower(receiver)
			}

			i := &callMethod{
				Namespace: strings.ToLower(stmt.Extension),
				Method:    strings.ToLower(stmt.Method),
				Args:      args,
				Receivers: receivers,
			}
			instructions = append(instructions, i)
		case *actparser.DMLStmt:
			// apply schema to db name in statement
			deterministic, err := sqlanalyzer.ApplyRules(stmt.Statement, sqlanalyzer.AllRules,
				schema.Tables, dbidSchema(schema.DBID()))
			if err != nil {
				return nil, err
			}

			if deterministic.Mutative && isViewProcedure {
				return nil, fmt.Errorf("view procedure cannot contain mutative statements")
			}

			i := &dmlStmt{
				SQLStatement:      deterministic.Statement,
				OrderedParameters: deterministic.ParameterOrder,
			}
			instructions = append(instructions, i)
		case *actparser.ActionCallStmt:
			args, err := makeExecutables(stmt.Args)
			if err != nil {
				return nil, err
			}

			receivers := make([]string, len(stmt.Receivers))
			for i, receiver := range stmt.Receivers {
				receivers[i] = strings.ToLower(receiver)
			}

			// if calling external procedure, the procedure must be public and view.
			// if calling internal procedure, the procedure must be view.
			callingViewProcedure := false // callingViewProcedure tracks whether the called procedure is a view
			if stmt.Database == schema.DBID() || stmt.Database == "" {
				// internal
				var procedure *common.Procedure
				for _, p := range schema.Procedures {
					if p.Name == stmt.Method {
						procedure = p
						break
					}
				}
				if procedure == nil {
					return nil, fmt.Errorf(`procedure "%s" not found`, stmt.Method)
				}

				for _, mod := range procedure.Modifiers {
					if mod == common.ModifierView {
						callingViewProcedure = true
						break
					}
				}

			} else {
				// external
				dataset, ok := global.datasets[stmt.Database]
				if !ok {
					return nil, fmt.Errorf(`dataset "%s" not found`, stmt.Database)
				}

				proc, ok := dataset.procedures[stmt.Method]
				if !ok {
					return nil, fmt.Errorf(`procedure "%s" not found`, stmt.Method)
				}

				if !proc.public {
					return nil, fmt.Errorf(`%w: procedure "%s" is not public`, ErrPrivateProcedure, stmt.Method)
				}

				callingViewProcedure = proc.view
			}
			if isViewProcedure && !callingViewProcedure {
				return nil, fmt.Errorf("view procedures cannot call non-view procedures")
			}

			i := &callMethod{
				Namespace: strings.ToLower(stmt.Database),
				Method:    strings.ToLower(stmt.Method),
				Args:      args,
				Receivers: receivers,
			}
			instructions = append(instructions, i)
		}
	}

	return &procedure{
		name:         unparsed.Name,
		public:       unparsed.Public,
		parameters:   unparsed.Args, // map with $ bind names, no @caller etc. yet
		view:         unparsed.IsView(),
		instructions: instructions,
	}, nil
}

// Call executes a procedure.
func (p *procedure) call(scope *precompiles.ProcedureContext, global *GlobalContext, db sql.DB, inputs []any) error {
	if len(inputs) != len(p.parameters) {
		return fmt.Errorf(`%w: procedure "%s" requires %d arguments, but %d were provided`, ErrIncorrectNumberOfArguments, p.name, len(p.parameters), len(inputs))
	}

	// if procedure does not have view tag, then it can mutate state
	// this means that we must have a readwrite connection
	if !p.view {
		tx, ok := db.(sql.AccessModer)
		if !ok {
			return errors.New("DB does not provide access mode needed for mutative action")
		}
		if tx.AccessMode() != sql.ReadWrite {
			return fmt.Errorf(`%w: mutable procedure "%s" called with non-mutative scope`, ErrMutativeProcedure, p.name)
		}
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

	newScope := scope.NewScope()
	newScope.StackDepth++ // not done by NewScope since (*baseDataset).Call would do it again

	// if no namespace is specified, we call a local procedure.
	// this can access public and private procedures.
	if e.Namespace == "" {
		procedure, ok := dataset.procedures[e.Method]
		if !ok {
			return fmt.Errorf(`procedure "%s" not found`, e.Method)
		}

		err = procedure.call(newScope, global, db, inputs)
	} else {
		namespace, ok := dataset.namespaces[e.Namespace]
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
		return fmt.Errorf(`%w: procedure "%s" returned %d values, but only %d receivers were specified`, ErrIncorrectNumberOfArguments, e.Method, len(results), len(e.Receivers))
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
	args := append([]any{pg.QueryModeExec}, params...)
	results, err := db.Execute(scope.Ctx, e.SQLStatement, args...)
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

// makeExecutables converts a set of tree.Expression into a set of evaluatables.
// These are SQL statements that executed with arguments from previously bound
// values (either from the action call params or results from preceding
// instructions in the procedure), and whose results are used as the input
// arguments for action or extension calls.
//
// See their execution in (*callMethod).execute inside the `range e.Args` to
// collect the `inputs` passed to the call of a dataset method or other
// "namespace" method, such as an extension method.
func makeExecutables(exprs []tree.Expression) ([]evaluatable, error) {
	execs := make([]evaluatable, 0, len(exprs))

	for _, expr := range exprs {
		switch e := expr.(type) {
		case *tree.ExpressionBindParameter:
			// This could be a special one that returns an evaluatable that
			// ignores the passed ResultSetFunc since the value is
		case *tree.ExpressionLiteral, *tree.ExpressionUnary, *tree.ExpressionBinaryComparison, *tree.ExpressionFunction, *tree.ExpressionArithmetic:
			// Acceptable expression type.
		default:
			return nil, fmt.Errorf("unsupported expression type: %T", e)
		}

		// clean expression, since it is submitted by the user
		err := expr.Walk(clean.NewStatementCleaner())
		if err != nil {
			return nil, err
		}

		// The schema walker is not necessary for inline expressions, since
		// we do not support table references in inline expressions.
		walker := sqlanalyzer.NewWalkerRecoverer(expr)
		paramVisitor := parameters.NewParametersWalker()
		err = walker.Walk(paramVisitor)
		if err != nil {
			return nil, fmt.Errorf("error replacing parameters: %w", err)
		}

		// SELECT expr;  -- prepare new value in receivers for subsequent
		// statements This query needs to be run in "simple" execution mode
		// rather than "extended" execution mode, which asks the database for
		// OID (placeholder types) that it can't know since there's no FOR table.
		selectTree := &tree.Select{
			SelectStmt: &tree.SelectStmt{
				SelectCores: []*tree.SelectCore{
					{
						SelectType: tree.SelectTypeAll,
						Columns: []tree.ResultColumn{
							&tree.ResultColumnExpression{
								Expression: expr,
							},
						},
					},
				},
			},
		}

		stmt, err := tree.SafeToSQL(selectTree)
		if err != nil {
			return nil, err
		}

		// here, we need to prepare the passed values and order them according to the bind names
		// We also must indicate to the database that we are in inferred arg types mode.
		// This allows inline expressions, such as SELECT $1 + $2.
		execs = append(execs, func(ctx context.Context, exec dbQueryFn, values map[string]any) (any, error) {
			// we need to start with a slice of the mode key
			// for in-line expressions, we need to use the inferred arg types
			valSlice := []any{pg.QueryModeInferredArgTypes}

			// ordering the map values according to the bind names
			valSlice = append(valSlice, orderAndCleanValueMap(values, paramVisitor.OrderedParameters)...)

			result, err := exec(ctx, stmt, valSlice...) // more values than binds
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

			return record[0], nil
		})
	}

	return execs, nil
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
