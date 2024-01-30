package execution

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/clean"
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
type instruction interface {
	execute(scope *ScopeContext, dataset *dataset) error
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
	dataset *dataset
}

// prepareProcedure parses a procedure from a types.Procedure.
func prepareProcedure(unparsed *types.Procedure, datasetCtx *dataset) (*procedure, error) {
	instructions := make([]instruction, 0)

	for _, mod := range unparsed.Modifiers {
		instr, err := convertModifier(mod)
		if err != nil {
			return nil, err
		}

		instructions = append(instructions, instr)
	}

	for _, stmt := range unparsed.Statements {
		instr, err := prepareStmt(stmt, !unparsed.IsMutative(), datasetCtx.schema.Tables)
		if err != nil {
			return nil, err
		}

		instructions = append(instructions, instr)
	}

	return &procedure{
		name:         unparsed.Name,
		public:       unparsed.Public,
		parameters:   unparsed.Args,
		mutable:      unparsed.IsMutative(),
		instructions: instructions,
		dataset:      datasetCtx,
	}, nil
}

// Call executes a procedure.
func (p *procedure) call(scope *ScopeContext, inputs []any) error {
	if len(inputs) != len(p.parameters) {
		return fmt.Errorf(`%w: procedure "%s" requires %d arguments, but %d were provided`, ErrIncorrectNumberOfArguments, p.name, len(p.parameters), len(inputs))
	}

	if p.mutable && !scope.Mutative() {
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
		return instructionFunc(func(scope *ScopeContext, dataset *dataset) error {
			if !bytes.Equal(scope.execution.signer, dataset.schema.Owner) {
				return fmt.Errorf("cannot call owner procedure, not owner")
			}

			return nil
		}), nil
	}

	// we do not necessarily have an instruction for every modifier type, but we do not want to return an error
	return instructionFunc(func(scope *ScopeContext, dataset *dataset) error {
		return nil
	}), nil
}

// prepareStmt parses a statement into an instruction.
// if immutable (aka a VIEW procedure), then the function will
// return an error if the statement is attempting to mutate state.
func prepareStmt(stmt string, immutable bool, tables []*types.Table) (instruction, error) {
	parsedStmt, err := actparser.Parse(stmt)
	if err != nil {
		return nil, err
	}

	var instr instruction

	switch stmt := parsedStmt.(type) {
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
		instr = i
	case *actparser.DMLStmt:
		deterministic, err := sqlanalyzer.ApplyRules(stmt.Statement, sqlanalyzer.AllRules, tables)
		if err != nil {
			return nil, err
		}

		nonDeterministic, err := sqlanalyzer.ApplyRules(stmt.Statement, sqlanalyzer.NoCartesianProduct, tables)
		if err != nil {
			return nil, err
		}
		//
		//var stubSchema *types.Schema
		//costCalculator, err := cost.GenCostCalculator(stmt.Statement, tables, stubSchema)
		//if err != nil {
		//	return nil, err
		//}

		i := &dmlStmt{
			DeterministicStatement:    deterministic.Statement(),
			NonDeterministicStatement: nonDeterministic.Statement(),
			Mutative:                  deterministic.Mutative(),
			//CostCalculator:            costCalculator,
		}
		instr = i

		if immutable && i.Mutative {
			return nil, fmt.Errorf("cannot mutate state in immutable procedure")
		}

	case *actparser.ActionCallStmt:
		args, err := makeExecutables(stmt.Args)
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

	// Receivers are the variables that the return values are assigned to.
	Receivers []string
}

// Execute calls a method from a namespace that is accessible within this dataset.
// If no namespace is specified, the local namespace is used.
// It will pass all arguments to the method, and assign the return values to the receivers.
func (e *callMethod) execute(scope *ScopeContext, dataset *dataset) error {
	var exec sql.ResultSetFunc
	if scope.Mutative() {
		exec = dataset.readWriter
	} else {
		exec = dataset.read
	}

	var inputs []any
	vals := scope.Values() // declare here since scope.Values() is expensive
	for _, arg := range e.Args {
		val, err := arg(scope.Ctx(), exec, vals)
		if err != nil {
			return err
		}

		inputs = append(inputs, val)
	}

	var results []any
	var err error

	// if no namespace is specified, we call a local procedure.
	// this can access public and private procedures.
	if e.Namespace == "" {
		procedure, ok := dataset.procedures[e.Method]
		if !ok {
			return fmt.Errorf(`procedure "%s" not found`, e.Method)
		}

		err = procedure.call(scope.NewScope(scope.DBID(), scope.Procedure()), inputs)
	} else {
		namespace, ok := dataset.namespaces[e.Namespace]
		if !ok {
			return fmt.Errorf(`namespace "%s" not found`, e.Namespace)
		}

		// new scope since we are calling a namespace
		results, err = namespace.Call(scope.NewScope(scope.DBID(), scope.Procedure()), e.Method, inputs)
	}
	if err != nil {
		return err
	}

	if len(e.Receivers) > len(results) {
		return fmt.Errorf(`%w: procedure "%s" returned %d values, but only %d receivers were specified`, ErrIncorrectNumberOfArguments, e.Method, len(results), len(e.Receivers))
	}

	for i, result := range results {
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

	//// CostCalculator is the cost calculator for the statement.
	//CostCalculator cost.Calculator
}

func (e *dmlStmt) execute(scope *ScopeContext, dataset *dataset) error {
	// this might be redundant
	if !scope.Mutative() && e.Mutative {
		return fmt.Errorf("cannot mutate state in immutable procedure")
	}

	var results *sql.ResultSet
	var err error
	if scope.Mutative() {
		results, err = dataset.readWriter(scope.Ctx(), e.DeterministicStatement, scope.Values())
	} else {
		results, err = dataset.read(scope.Ctx(), e.NonDeterministicStatement, scope.Values())
	}
	if err != nil {
		return err
	}

	scope.SetResult(results)

	return nil
}

type instructionFunc func(scope *ScopeContext, dataset *dataset) error

// implement instruction
func (f instructionFunc) execute(scope *ScopeContext, dataset *dataset) error {
	return f(scope, dataset)
}

// evaluatable is an expression that can be evaluated to a scalar value.
// It is used to handle inline expressions, such as within action calls.
type evaluatable func(ctx context.Context, exec sql.ResultSetFunc, values map[string]any) (any, error)

// makeExecutables converts a set of tree.Expression into a set of evaluatables.
func makeExecutables(exprs []tree.Expression) ([]evaluatable, error) {
	execs := make([]evaluatable, 0)

	for _, expr := range exprs {
		switch e := expr.(type) {
		case *tree.ExpressionLiteral, *tree.ExpressionBindParameter, *tree.ExpressionUnary, *tree.ExpressionBinaryComparison, *tree.ExpressionFunction, *tree.ExpressionArithmetic:
			// do nothing
		default:
			return nil, fmt.Errorf("unsupported expression type: %T", e)
		}

		// clean expression, since it is submitted by the user
		err := expr.Walk(clean.NewStatementCleaner())
		if err != nil {
			return nil, err
		}

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

		execs = append(execs, func(ctx context.Context, exec sql.ResultSetFunc, values map[string]any) (any, error) {
			result, err := exec(ctx, stmt, values)
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
