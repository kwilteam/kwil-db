package execution

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/clean"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/parameters"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/schema"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	sql "github.com/kwilteam/kwil-db/internal/sql"
	actparser "github.com/kwilteam/kwil-db/parse/action"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

var (
	ErrIncorrectNumberOfArguments = fmt.Errorf("incorrect number of arguments")
	ErrPrivateProcedure           = fmt.Errorf("procedure is private")
	ErrMutativeProcedure          = fmt.Errorf("procedure is mutative")
)

// instruction is an instruction that can be executed.
// It is used to define the behavior of a procedure.
type instruction interface { // i.e. dmlStmt, callMethod, or instructionFunc
	execute(scope *ProcedureContext, dataset *baseDataset) error
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

	// mutable indicates whether the procedure is mutable.
	mutable bool

	// instructions are the instructions that the procedure executes when called.
	instructions []instruction

	// dataset is the dataset that the procedure is defined in.
	dataset *baseDataset
}

// prepareProcedure parses a procedure from a types.Procedure.
func prepareProcedure(unparsed *types.Procedure, datasetCtx *baseDataset) (*procedure, error) {
	instructions := make([]instruction, 0)

	for _, mod := range unparsed.Modifiers {
		instr, err := convertModifier(mod)
		if err != nil {
			return nil, err
		}

		instructions = append(instructions, instr)
	}

	for _, stmt := range unparsed.Statements {
		// pass datasetCtx.schema.DBID() to be used in sql rewrite with dbid as
		// schema.table
		//
		// also, future support for cross dataset queries might require more
		// context, unless such queries would explicitly specify the schema of
		// the other dataset in the SQL statement
		instr, err := prepareStmt(stmt, !unparsed.IsMutative(), datasetCtx.schema.Tables, datasetCtx.schema.DBID())
		if err != nil {
			return nil, err
		}

		instructions = append(instructions, instr)
	}

	return &procedure{
		name:         unparsed.Name,
		public:       unparsed.Public,
		parameters:   unparsed.Args, // map with $ bind names, no @caller etc. yet
		mutable:      unparsed.IsMutative(),
		instructions: instructions,
		dataset:      datasetCtx,
	}, nil
}

// Call executes a procedure.
func (p *procedure) call(scope *ProcedureContext, inputs []any) error {
	if len(inputs) != len(p.parameters) {
		return fmt.Errorf(`%w: procedure "%s" requires %d arguments, but %d were provided`, ErrIncorrectNumberOfArguments, p.name, len(p.parameters), len(inputs))
	}

	if p.mutable && !scope.Mutative {
		return fmt.Errorf(`%w: mutable procedure "%s" called with non-mutative scope`, ErrMutativeProcedure, p.name)
	}

	for i, param := range p.parameters {
		scope.values[param] = inputs[i]
	}

	for _, inst := range p.instructions {
		if err := inst.execute(scope, p.dataset); err != nil {
			return err
		}
	}

	return nil
}

// covertModifier converts a types.Modifier to an instruction.
func convertModifier(mod types.Modifier) (instruction, error) {
	switch mod {
	case types.ModifierOwner:
		return instructionFunc(func(scope *ProcedureContext, dataset *baseDataset) error {
			if !bytes.Equal(scope.Signer, dataset.schema.Owner) {
				return fmt.Errorf("cannot call owner procedure, not owner")
			}

			return nil
		}), nil
	}

	// we do not necessarily have an instruction for every modifier type, but we do not want to return an error
	return instructionFunc(func(scope *ProcedureContext, dataset *baseDataset) error {
		return nil
	}), nil
}

// prepareStmt parses a statement into an instruction.
// if immutable (aka a VIEW procedure), then the function will
// return an error if the statement is attempting to mutate state.
func prepareStmt(stmt string, immutable bool, tables []*types.Table, dbid string) (instruction, error) {
	parsedStmt, err := actparser.Parse(stmt)
	if err != nil {
		return nil, err
	}

	var instr instruction

	switch stmt := parsedStmt.(type) {
	case *actparser.ExtensionCallStmt:
		args, err := makeExecutables(stmt.Args, types.DBIDSchema(dbid))
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
		instr = i
	case *actparser.DMLStmt:
		// apply schema to db name in statement
		deterministic, err := sqlanalyzer.ApplyRules(stmt.Statement, sqlanalyzer.AllRules,
			tables, types.DBIDSchema(dbid))
		if err != nil {
			return nil, err
		}

		nonDeterministic, err := sqlanalyzer.ApplyRules(stmt.Statement,
			sqlanalyzer.NoCartesianProduct|sqlanalyzer.ReplaceNamedParameters,
			tables, types.DBIDSchema(dbid))
		if err != nil {
			return nil, err
		}

		i := &dmlStmt{
			DeterministicStatement:    deterministic.Statement(),
			NonDeterministicStatement: nonDeterministic.Statement(),
			Mutative:                  deterministic.Mutative(),
		}
		instr = i

		if immutable && i.Mutative {
			return nil, fmt.Errorf("cannot mutate state in immutable procedure")
		}

	case *actparser.ActionCallStmt:
		args, err := makeExecutables(stmt.Args, types.DBIDSchema(dbid))
		if err != nil {
			return nil, err
		}

		receivers := make([]string, len(stmt.Receivers))
		for i, receiver := range stmt.Receivers {
			receivers[i] = strings.ToLower(receiver)
		}

		i := &callMethod{
			Namespace: strings.ToLower(stmt.Database),
			Method:    strings.ToLower(stmt.Method),
			Args:      args,
			Receivers: receivers,
		}
		instr = i
	default:
		return nil, fmt.Errorf("unknown statement type %T", stmt)
	}

	return instr, nil
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

// Execute calls a method from a namespace that is accessible within this dataset.
// If no namespace is specified, the local namespace is used.
// It will pass all arguments to the method, and assign the return values to the receivers.
func (e *callMethod) execute(scope *ProcedureContext, dataset *baseDataset) error {
	var exec types.ResultSetFunc
	if scope.Mutative {
		exec = dataset.readWriter
		// how do we know we are in a db transaction / session and that this is
		// ok? This scope/context seems to have mutative set according to the statement.
	} else {
		exec = dataset.read
		// what if we are in a session and want uncommitted data for this procedure exec?

		// only acct and val stores use the QueryPending method (previously,
		// used Execute to get results).  I think it also matters for engine
		// procedure execution, and it depends on call context (rpc call vs. block-tx)
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
		val, err := arg(scope.Ctx, exec, vals)
		if err != nil {
			return err
		}

		inputs = append(inputs, val)
	}

	var results []any
	var err error

	newScope := scope.NewScope()

	// if no namespace is specified, we call a local procedure.
	// this can access public and private procedures.
	if e.Namespace == "" {
		procedure, ok := dataset.procedures[e.Method]
		if !ok {
			return fmt.Errorf(`procedure "%s" not found`, e.Method)
		}

		err = procedure.call(newScope, inputs)
	} else {
		namespace, ok := dataset.namespaces[e.Namespace]
		if !ok {
			return fmt.Errorf(`namespace "%s" not found`, e.Namespace)
		}

		// new scope since we are calling a namespace
		results, err = namespace.Call(newScope, e.Method, inputs)
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

		scope.values[e.Receivers[i]] = result
	}

	return nil
}

// dmlStmt is a DML statement, we leave the parsing to sqlparser
type dmlStmt struct {
	// DeterministicStatement is the deterministic version of the statement.
	// This is almost always more expensive to execute, but it is guaranteed to return the same results.
	DeterministicStatement string

	// NonDeterministicStatement is the non-deterministic version of the statement.
	// This is almost always cheaper to execute, but it is not guaranteed to return the same results.
	NonDeterministicStatement string

	// Mutative is whether the statement mutates state.
	Mutative bool
}

func (e *dmlStmt) execute(scope *ProcedureContext, dataset *baseDataset) error {
	// this might be redundant
	if !scope.Mutative && e.Mutative {
		return fmt.Errorf("cannot mutate state in immutable procedure")
	}

	// Inject environment variables like @caller into args with scope.Values()
	var results *sql.ResultSet
	var err error
	if scope.Mutative {
		results, err = dataset.readWriter(scope.Ctx, e.DeterministicStatement, scope.Values())
	} else {
		results, err = dataset.read(scope.Ctx, e.NonDeterministicStatement, scope.Values())
	}
	if err != nil {
		return err
	}

	scope.Result = results

	return nil
}

type instructionFunc func(scope *ProcedureContext, dataset *baseDataset) error

// implement instruction
func (f instructionFunc) execute(scope *ProcedureContext, dataset *baseDataset) error {
	return f(scope, dataset)
}

// evaluatable is an expression that can be evaluated to a scalar value.
// It is used to handle inline expressions, such as within action calls.
type evaluatable func(ctx context.Context, exec types.ResultSetFunc, values map[string]any) (any, error)

// makeExecutables converts a set of tree.Expression into a set of evaluatables.
// These are SQL statements that executed with arguments from previously bound
// values (either from the action call params or results from preceding
// instructions in the procedure), and whose results are used as the input
// arguments for action or extension calls.
//
// See their execution in (*callMethod).execute inside the `range e.Args` to
// collect the `inputs` passed to the call of a dataset method or other
// "namespace" method, such as an extension method.
func makeExecutables(exprs []tree.Expression, pgSchemaName string) ([]evaluatable, error) {
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
		err := expr.Accept(clean.NewStatementCleaner())
		if err != nil {
			return nil, err
		}

		accept := sqlanalyzer.NewAcceptRecoverer(expr)
		paramVisitor := parameters.NewNamedParametersVisitor()
		err = accept.Accept(paramVisitor)
		if err != nil {
			return nil, fmt.Errorf("error replacing named parameters: %w", err)
		}
		// Is the schema walker necessary too? Yes, if any of these action/extension
		// call argument expressions reference a table column...
		err = accept.Accept(schema.NewSchemaWalker(pgSchemaName))
		if err != nil {
			return nil, fmt.Errorf("error applying schema rules: %w", err)
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
								Alias:      concatKeys(paramVisitor.Binds), // not critical, just for column name other than ?column?
							},
						},
					},
				},
			},
		}

		stmt, err := selectTree.ToSQL()
		if err != nil {
			return nil, err
		}
		// fmt.Println("selectTree.ToSQL(): ", stmt)

		execs = append(execs, func(ctx context.Context, exec types.ResultSetFunc, values map[string]any) (any, error) {
			// exec must be created by queryor() to prepare the values map to
			// match the rewritten statement from NamedParametersVisitor, and to
			// make the query run in "simple" execution mode.
			result, err := exec(ctx, stmt, values) // more values than binds
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

func concatKeys[T any](m map[string]T) string {
	if len(m) == 0 {
		return "unnamed"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return strings.Join(keys, "_")
}
