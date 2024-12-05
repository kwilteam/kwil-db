// package interpreter provides a basic interpreter for Kuneiform procedures.
// It allows running procedures as standalone programs (instead of generating
// PL/pgSQL code).
package interpreter

import (
	"context"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
)

func Run(ctx context.Context, proc *types.Procedure, schema *types.Schema, args []any) (*ProcedureRunResult, error) {
	panic("not implemented")
	// parseResult, err := parse.ParseProcedure(proc, schema)
	// if err != nil {
	// 	return nil, err
	// }
	// if parseResult.ParseErrs.Err() != nil {
	// 	return nil, parseResult.ParseErrs.Err()
	// }

	// i := &interpreterPlanner{}

	// exec := &executionContext{
	// 	scope: newScope(),
	// }

	// if len(proc.Parameters) != len(args) {
	// 	return nil, fmt.Errorf("expected %d arguments, got %d", len(proc.Parameters), len(args))
	// }

	// for j, arg := range args {
	// 	val, err := NewVariable(arg)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	if !proc.Parameters[j].Type.EqualsStrict(val.Type()) {
	// 		return nil, fmt.Errorf("expected argument %d to be %s, got %s", j+1, proc.Parameters[j].Type, val.Type())
	// 	}

	// 	err = exec.allocateVariable(proc.Parameters[j].Name, val)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	// var expectedShape []*types.DataType
	// if proc.Returns != nil {
	// 	for _, f := range proc.Returns.Fields {
	// 		expectedShape = append(expectedShape, f.Type)
	// 	}
	// }

	// res := newReturnableCursor(expectedShape)
	// procRes := &ProcedureRunResult{}

	// go func() {
	// 	for _, stmt := range parseResult.AST {
	// 		stmt.Accept(i).(stmtFunc)(exec, res)
	// 		if err != nil {
	// 			res.Err() <- err
	// 			return
	// 		}
	// 	}

	// 	err := res.Close()
	// 	if err != nil {
	// 		// TODO: use a logger
	// 		// Currently this close method's implementation cannot return an error
	// 		// so it's safe to ignore this error while developing.
	// 		panic(fmt.Errorf("error closing cursor: %w", err))
	// 	}
	// }()

	// // a procedure can return 3 things:
	// // - nothing
	// // - a single row
	// // - any number of rows (a table)
	// if proc.Returns == nil {
	// 	// should return nothing
	// 	_, done, err := res.Next(ctx)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if !done {
	// 		return nil, fmt.Errorf("unexpected return value")
	// 	}
	// } else if proc.Returns.IsTable {
	// 	// should return 0 to many rows
	// 	for {
	// 		vals, done, err := res.Next(ctx)
	// 		if err != nil && err != errReturn {
	// 			return nil, err
	// 		}
	// 		if done {
	// 			break
	// 		}

	// 		named, err := makeNamedReturns(proc.Returns.Fields, vals)
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 		procRes.Values = append(procRes.Values, named)

	// 		if err == errReturn {
	// 			break
	// 		}
	// 	}
	// } else {
	// 	// should return a single row
	// 	vals, done, err := res.Next(ctx)
	// 	// should always be errReturn?
	// 	if err != nil && err != errReturn {
	// 		return nil, err
	// 	}
	// 	if done {
	// 		return nil, fmt.Errorf("expected return value")
	// 	}

	// 	named, err := makeNamedReturns(proc.Returns.Fields, vals)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	procRes.Values = append(procRes.Values, named)

	// 	// check if there are more return values
	// 	_, done, err = res.Next(ctx)
	// 	if err != nil && err != errReturn {
	// 		return nil, err
	// 	}
	// 	if !done {
	// 		return nil, fmt.Errorf("unexpected return value")
	// 	}
	// }

	// return procRes, nil
}

func makeNamedReturns(expected []*types.NamedType, record RecordValue) ([]*NamedValue, error) {
	if len(expected) != len(record.Fields) {
		return nil, fmt.Errorf("expected %d return fields, got %d", len(expected), len(record.Fields))
	}
	if len(expected) != len(record.Order) {
		return nil, fmt.Errorf("expected %d return ordered fields, got %d", len(expected), len(record.Order))
	}

	named := make([]*NamedValue, len(expected))
	for i, e := range expected {
		fieldName := record.Order[i]

		val, ok := record.Fields[fieldName]
		if !ok {
			return nil, fmt.Errorf("expected return value %s not found", e.Name)
		}

		named[i] = &NamedValue{
			Name:  e.Name,
			Value: val,
		}
	}

	return named, nil
}

type NamedValue struct {
	Name  string
	Value Value
}

type ProcedureRunResult struct {
	Values [][]*NamedValue
}

// makeActionToExecutable creates an executable from an action
func makeActionToExecutable(act *Action) *executable {
	planner := &interpreterPlanner{}
	stmtFns := make([]stmtFunc, len(act.Body))
	for j, stmt := range act.Body {
		stmtFns[j] = stmt.Accept(planner).(stmtFunc)
	}

	validateArgs := func(v []Value) error {
		if len(v) != len(act.Parameters) {
			return fmt.Errorf("expected %d arguments, got %d", len(act.Parameters), len(v))
		}

		for i, arg := range v {
			if !act.Parameters[i].Type.EqualsStrict(arg.Type()) {
				return fmt.Errorf("expected argument %d to be %s, got %s", i+1, act.Parameters[i].Type, arg.Type())
			}
		}

		return nil
	}

	return &executable{
		Name: act.Name,
		ReturnType: func(v []Value) (*ActionReturn, error) {
			err := validateArgs(v)
			if err != nil {
				return nil, err
			}

			return act.Returns, nil
		},
		Func: func(exec *executionContext, args []Value, fn func([]Value) error) error {
			// if this is not a view action and the execution is trying to mutate state, then return an error
			if !act.IsView() && exec.mutatingState {
				return fmt.Errorf("%w: cannot execute action %s in a read-only transaction", ErrActionMutatesState, act.Name)
			}

			// if the action is owner only, then check if the user is the owner
			if act.OwnerOnly() && !exec.accessController.IsOwner(exec.txCtx.Caller) {
				return fmt.Errorf("%w: action %s can only be executed by the owner", ErrActionOwnerOnly, act.Name)
			}

			// validate the args
			err := validateArgs(args)
			if err != nil {
				return err
			}

			for j, param := range act.Parameters {
				err = exec.allocateVariable(param.Name, args[j])
				if err != nil {
					return err
				}
			}

			// execute the statements
			for _, stmt := range stmtFns {
				err := stmt(exec, fn)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// interpreterPlanner creates functions for running Kuneiform logic.
type interpreterPlanner struct{}

// FunctionSignature is the signature for either a user-defined PL/pgSQL function, a built-in function,
// or an action.
type FunctionSignature struct {
	// Name is the name of the function.
	Name string
	// Parameters are the parameters of the function.
	Parameters []*types.NamedType
	// Returns are the return values of the function.
	Returns *types.ProcedureReturn
}

var (

	// errBreak is an error returned when a break statement is encountered.
	errBreak = errors.New("break")
	// errReturn is an error returned when a return statement is encountered.
	errReturn = errors.New("return")
)

type stmtFunc func(exec *executionContext, fn func([]Value) error) error

func (i *interpreterPlanner) VisitProcedureStmtDeclaration(p0 *parse.ProcedureStmtDeclaration) any {
	return stmtFunc(func(exec *executionContext, fn func([]Value) error) error {
		return exec.allocateVariable(p0.Variable.Name, NewNullValue(p0.Type))
	})
}

func (i *interpreterPlanner) VisitProcedureStmtAssignment(p0 *parse.ProcedureStmtAssign) any {
	valFn := p0.Value.Accept(i).(exprFunc)

	var arrFn exprFunc
	var indexFn exprFunc
	if a, ok := p0.Variable.(*parse.ExpressionArrayAccess); ok {
		arrFn = a.Array.Accept(i).(exprFunc)
		indexFn = a.Index.Accept(i).(exprFunc)
	}
	return stmtFunc(func(exec *executionContext, fn func([]Value) error) error {
		val, err := valFn(exec)
		if err != nil {
			return err
		}

		switch a := p0.Variable.(type) {
		case *parse.ExpressionVariable:
			return exec.setVariable(a.Name, val)
		case *parse.ExpressionArrayAccess:
			scalarVal, ok := val.(ScalarValue)
			if !ok {
				return fmt.Errorf("expected scalar value, got %T", val)
			}

			arrVal, err := arrFn(exec)
			if err != nil {
				return err
			}

			arr, ok := arrVal.(ArrayValue)
			if !ok {
				return fmt.Errorf("expected array, got %T", arrVal)
			}

			index, err := indexFn(exec)
			if err != nil {
				return err
			}

			if !index.Type().EqualsStrict(types.IntType) {
				return fmt.Errorf("array index must be integer, got %s", index.Type())
			}

			err = arr.Set(index.RawValue().(int64), scalarVal)
			if err != nil {
				return err
			}

			// TODO: do I need to re-set the array? I dont think so b/c the implementation is a pointer

			return nil
		default:
			panic(fmt.Errorf("unexpected assignable variable type: %T", p0.Variable))
		}
	})
}

func (i *interpreterPlanner) VisitProcedureStmtCall(p0 *parse.ProcedureStmtCall) any {

	// we cannot simply use the same visitor as the expression function call, because expression function
	// calls always return exactly one value. Here, we can return 0 values, many values, or a table.

	receivers := make([]string, len(p0.Receivers))
	for j, r := range p0.Receivers {
		receivers[j] = r.Name
	}

	args := make([]exprFunc, len(p0.Call.Args))
	for j, arg := range p0.Call.Args {
		args[j] = arg.Accept(i).(exprFunc)
	}

	return stmtFunc(func(exec *executionContext, fn func([]Value) error) error {
		funcDef, ok := exec.availableFunctions[p0.Call.Name]
		if !ok {
			return fmt.Errorf(`action "%s" no longer exists`, p0.Call.Name)
		}

		vals := make([]Value, len(args))
		for j, valFn := range args {
			val, err := valFn(exec)
			if err != nil {
				return err
			}

			vals[j] = val
		}

		returns, err := funcDef.ReturnType(vals)
		if err != nil {
			return err
		}

		// verify the returns.
		// If the user expects values, then there must be enough values to assign to the receivers.
		// If the user does not expect values, then the function can return anything / return nothing.
		if len(receivers) != 0 {
			if returns == nil {
				return fmt.Errorf(`expected function "%s" to return %d values, but it does not return anything`, funcDef.Name, len(receivers))
			}

			if len(receivers) > len(returns.Fields) {
				return fmt.Errorf(`expected function "%s" to return at least %d values, but it returns %d`, funcDef.Name, len(receivers), len(returns.Fields))
			}

			if returns.IsTable {
				return fmt.Errorf(`expected function "%s" to return a single record, but it returns a table`, funcDef.Name)
			}
		}

		oldScope := exec.scope
		defer func() {
			exec.scope = oldScope
		}()
		// we create an entirely new scope because the new procedure should not be able to access any
		// variables from the current scope.
		exec.scope = newScope(oldScope.namespace)

		iter := 0
		err = funcDef.Func(exec, vals, func(received []Value) error {
			iter++

			// re-verify the returns, since the above checks only for what the function signature
			// says, but this checks what the function actually returns.
			if len(receivers) > len(received) {
				return fmt.Errorf(`expected action "%s" to return at least %d values, but it returned %d`, funcDef.Name, len(receivers), len(received))
			}

			for j, r := range receivers {
				if !returns.Fields[j].Type.EqualsStrict(received[j].Type()) {
					return fmt.Errorf(`expected action "%s" to return %s at position %d, but it returned %s`, funcDef.Name, returns.Fields[j].Type, j+1, received[j].Type())
				}

				err := exec.setVariable(r, received[j])
				if err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return err
		}
		if len(receivers) > 0 {
			if iter == 0 {
				return fmt.Errorf(`expected action "%s" to return a single record, but it returned nothing`, funcDef.Name)
			}
			if iter > 1 {
				return fmt.Errorf(`expected action "%s" to return a single record, but it returned %d records`, funcDef.Name, iter)
			}
		}

		return nil
	})
}

// executeBlock executes a block of statements with their own sub-scope.
// It takes a list of statements, and a list of variable allocations that will be made in the sub-scope.
func executeBlock(exec *executionContext, fn func([]Value) error,
	allocs []*NamedValue, stmtFuncs []stmtFunc) error {
	oldScope := exec.scope
	defer func() {
		exec.scope = oldScope
	}()

	exec.scope = exec.scope.subScope()

	for _, alloc := range allocs {
		err := exec.allocateVariable(alloc.Name, alloc.Value)
		if err != nil {
			return err
		}
	}

	for _, stmt := range stmtFuncs {
		err := stmt(exec, fn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *interpreterPlanner) VisitProcedureStmtForLoop(p0 *parse.ProcedureStmtForLoop) any {
	stmtFns := make([]stmtFunc, len(p0.Body))
	for j, stmt := range p0.Body {
		stmtFns[j] = stmt.Accept(i).(stmtFunc)
	}

	loopFn := p0.LoopTerm.Accept(i).(loopTermFunc)

	return stmtFunc(func(exec *executionContext, fn func([]Value) error) error {
		oldScope := exec.scope
		defer func() {
			exec.scope = oldScope
		}()

		err := loopFn(exec, func(term Value) error {
			exec.scope = oldScope.subScope()
			err := exec.allocateVariable(p0.Receiver.Name, term)
			if err != nil {
				return err
			}

			for _, stmt := range stmtFns {
				err := stmt(exec, fn)
				if err != nil {
					return err
				}
			}

			return nil
		})
		switch err {
		case nil, errBreak:
			// swallow break errors since we are breaking out of the loop
			return nil
		default:
			return err
		}
	})
}

// loopTermFunc is a function that allows iterating over a loop term.
// It calls the function passed to it with each value.
type loopTermFunc func(exec *executionContext, fn func(Value) error) (err error)

func (i *interpreterPlanner) VisitLoopTermRange(p0 *parse.LoopTermRange) any {
	startFn := p0.Start.Accept(i).(exprFunc)
	endFn := p0.End.Accept(i).(exprFunc)

	return loopTermFunc(func(exec *executionContext, fn func(Value) error) (err error) {
		start, err := startFn(exec)
		if err != nil {
			return err
		}

		end, err := endFn(exec)
		if err != nil {
			return err
		}

		if !start.Type().EqualsStrict(types.IntType) {
			return fmt.Errorf("expected integer, got %s", start.Type())
		}

		if !end.Type().EqualsStrict(types.IntType) {
			return fmt.Errorf("expected integer, got %s", end.Type())
		}

		for i := start.RawValue().(int64); i <= end.RawValue().(int64); i++ {
			err = fn(&IntValue{
				Val: i,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (i *interpreterPlanner) VisitLoopTermSQL(p0 *parse.LoopTermSQL) any {
	return loopTermFunc(func(exec *executionContext, fn func(Value) error) error {
		raw, err := p0.Statement.Raw()
		if err != nil {
			return err
		}

		// query executes a Kuneiform query and returns a cursor.
		return exec.query(raw, func(rv *RecordValue) error {
			return fn(rv)
		})
	})
}

func (i *interpreterPlanner) VisitLoopTermVariable(p0 *parse.LoopTermVariable) any {
	return loopTermFunc(func(exec *executionContext, fn func(Value) error) (err error) {
		val, found := exec.getVariable(p0.Variable.Name)
		if !found {
			return fmt.Errorf("%w: %s", ErrVariableNotFound, p0.Variable.Name)
		}

		arr, ok := val.(ArrayValue)
		if !ok {
			return fmt.Errorf("expected array, got %T", val)
		}

		for i := 0; i < arr.Len(); i++ {
			scalar, err := arr.Index(int64(i) + 1) // all arrays are 1-indexed
			if err != nil {
				return err
			}

			err = fn(scalar)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (i *interpreterPlanner) VisitProcedureStmtIf(p0 *parse.ProcedureStmtIf) any {
	var ifThenFns []struct {
		If   exprFunc
		Then []stmtFunc
	}

	for _, ifThen := range p0.IfThens {
		ifFn := ifThen.If.Accept(i).(exprFunc)
		var thenFns []stmtFunc
		for _, stmt := range ifThen.Then {
			thenFns = append(thenFns, stmt.Accept(i).(stmtFunc))
		}

		ifThenFns = append(ifThenFns, struct {
			If   exprFunc
			Then []stmtFunc
		}{
			If:   ifFn,
			Then: thenFns,
		})
	}

	var elseFns []stmtFunc
	if p0.Else != nil {
		for _, stmt := range p0.Else {
			elseFns = append(elseFns, stmt.Accept(i).(stmtFunc))
		}
	}

	return stmtFunc(func(exec *executionContext, fn func([]Value) error) error {
		branchRun := false // tracks if any IF branch has been run
		for _, ifThen := range ifThenFns {
			if branchRun {
				break
			}

			cond, err := ifThen.If(exec)
			if err != nil {
				return err
			}

			switch c := cond.(type) {
			case *BoolValue:
				if !c.Val {
					continue
				}
			case *NullValue:
				continue
			default:
				return fmt.Errorf("expected bool, got %s", c.Type())
			}

			branchRun = true

			err = executeBlock(exec, fn, nil, ifThen.Then)
			if err != nil {
				return err
			}
		}

		if !branchRun && p0.Else != nil {
			err := executeBlock(exec, fn, nil, elseFns)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (i *interpreterPlanner) VisitProcedureStmtSQL(p0 *parse.ProcedureStmtSQL) any {
	return stmtFunc(func(exec *executionContext, fn func([]Value) error) error {
		raw, err := p0.SQL.Raw()
		if err != nil {
			return err
		}

		// query executes any arbitrary SQL.
		err = exec.query(raw, func(rv *RecordValue) error {
			// we ignore results here since we are not returning anything.
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func (i *interpreterPlanner) VisitProcedureStmtBreak(p0 *parse.ProcedureStmtBreak) any {
	return stmtFunc(func(exec *executionContext, fn func([]Value) error) error {
		return errBreak
	})
}

func (i *interpreterPlanner) VisitProcedureStmtReturn(p0 *parse.ProcedureStmtReturn) any {
	valFns := make([]exprFunc, len(p0.Values))
	for j, v := range p0.Values {
		valFns[j] = v.Accept(i).(exprFunc)
	}

	return stmtFunc(func(exec *executionContext, fn func([]Value) error) error {
		vals := make([]Value, len(p0.Values))
		for j, valFn := range valFns {
			val, err := valFn(exec)
			if err != nil {
				return err
			}

			vals[j] = val
		}

		err := fn(vals)
		if err != nil {
			return err
		}

		// we return a special error to indicate that the procedure is done.
		return errReturn
	})
}

func (i *interpreterPlanner) VisitProcedureStmtReturnNext(p0 *parse.ProcedureStmtReturnNext) any {
	valFns := make([]exprFunc, len(p0.Values))
	for j, v := range p0.Values {
		valFns[j] = v.Accept(i).(exprFunc)
	}

	return stmtFunc(func(exec *executionContext, fn func([]Value) error) error {
		vals := make([]Value, len(p0.Values))
		for j, valFn := range valFns {
			val, err := valFn(exec)
			if err != nil {
				return err
			}

			vals[j] = val
		}

		err := fn(vals)
		if err != nil {
			return err
		}

		// we don't return an errReturn or mark done here because return next is not the last statement in a procedure.
		return nil
	})
}

// everything in this section is for expressions, which evaluate to exactly one value.

// exprFunc is a function that returns a value.
type exprFunc func(exec *executionContext) (Value, error)

func (i *interpreterPlanner) VisitExpressionLiteral(p0 *parse.ExpressionLiteral) any {
	return exprFunc(func(exec *executionContext) (Value, error) {
		return NewValue(p0.Value)
	})
}

func (i *interpreterPlanner) VisitExpressionFunctionCall(p0 *parse.ExpressionFunctionCall) any {
	args := make([]exprFunc, len(p0.Args))
	for j, arg := range p0.Args {
		args[j] = arg.Accept(i).(exprFunc)
	}

	return exprFunc(func(exec *executionContext) (Value, error) {
		// we check again because the action might have been dropped
		funcDef, ok := exec.availableFunctions[p0.Name]
		if !ok {
			return nil, fmt.Errorf(`function "%s" no longer exists`, p0.Name)
		}

		vals := make([]Value, len(args))
		for j, arg := range args {
			val, err := arg(exec)
			if err != nil {
				return nil, err
			}

			vals[j] = val
		}

		returns, err := funcDef.ReturnType(vals)
		if err != nil {
			return nil, err
		}

		if returns == nil {
			return nil, fmt.Errorf(`cannot call function "%s" in an expression because it returns nothing`, p0.Name)
		}
		if returns.IsTable {
			return nil, fmt.Errorf(`cannot call function "%s" in an expression because it returns a table`, p0.Name)
		}
		if len(returns.Fields) != 1 {
			return nil, fmt.Errorf(`cannot call function "%s" in an expression because it returns multiple values`, p0.Name)
		}

		var val Value
		iters := 0
		err = funcDef.Func(exec, vals, func(received []Value) error {
			iters++
			if len(received) != 1 {
				return fmt.Errorf(`expected function "%s" to return 1 value, but it returned %d`, p0.Name, len(received))
			}
			if !returns.Fields[0].Type.EqualsStrict(received[0].Type()) {
				return fmt.Errorf(`expected function "%s" to return %s, but it returned %s`, p0.Name, returns.Fields[0].Type, received[0].Type())
			}

			val = received[0]

			return nil
		})
		if err != nil {
			return nil, err
		}

		if iters == 0 {
			return nil, fmt.Errorf(`expected function "%s" to return a single value, but it returned nothing`, p0.Name)
		} else if iters > 1 {
			return nil, fmt.Errorf(`expected function "%s" to return a single value, but it returned %d values`, p0.Name, iters)
		}

		return val, nil
	})
}

func (i *interpreterPlanner) VisitExpressionVariable(p0 *parse.ExpressionVariable) any {
	return exprFunc(func(exec *executionContext) (Value, error) {
		val, found := exec.getVariable(p0.Name)
		if !found {
			return nil, fmt.Errorf("%w: %s", ErrVariableNotFound, p0.Name)
		}

		return val, nil
	})
}

func (i *interpreterPlanner) VisitExpressionArrayAccess(p0 *parse.ExpressionArrayAccess) any {
	arrFn := p0.Array.Accept(i).(exprFunc)
	indexFn := p0.Index.Accept(i).(exprFunc)

	return exprFunc(func(exec *executionContext) (Value, error) {
		arrVal, err := arrFn(exec)
		if err != nil {
			return nil, err
		}

		arr, ok := arrVal.(ArrayValue)
		if !ok {
			return nil, fmt.Errorf("expected array, got %T", arrVal)
		}

		index, err := indexFn(exec)
		if err != nil {
			return nil, err
		}

		if !index.Type().EqualsStrict(types.IntType) {
			return nil, fmt.Errorf("array index must be integer, got %s", index.Type())
		}

		return arr.Index(index.RawValue().(int64))
	})
}

func (i *interpreterPlanner) VisitExpressionMakeArray(p0 *parse.ExpressionMakeArray) any {
	valFns := make([]exprFunc, len(p0.Values))
	for j, v := range p0.Values {
		valFns[j] = v.Accept(i).(exprFunc)
	}

	return exprFunc(func(exec *executionContext) (Value, error) {
		if len(valFns) == 0 {
			return nil, fmt.Errorf("array must have at least one element")
		}

		val0, err := valFns[0](exec)
		if err != nil {
			return nil, err
		}

		scal, ok := val0.(ScalarValue)
		if !ok {
			return nil, fmt.Errorf("expected scalar value, got %T", val0)
		}

		var vals []ScalarValue
		for j, valFn := range valFns {
			if j == 0 {
				continue
			}

			val, err := valFn(exec)
			if err != nil {
				return nil, err
			}

			scal, ok := val.(ScalarValue)
			if !ok {
				return nil, fmt.Errorf("expected scalar value, got %T", val)
			}

			vals = append(vals, scal)
		}

		return scal.Array(vals...)
	})
}

func (i *interpreterPlanner) VisitExpressionFieldAccess(p0 *parse.ExpressionFieldAccess) any {
	recordFn := p0.Record.Accept(i).(exprFunc)

	return exprFunc(func(exec *executionContext) (Value, error) {
		objVal, err := recordFn(exec)
		if err != nil {
			return nil, err
		}

		obj, ok := objVal.(*RecordValue)
		if !ok {
			return nil, fmt.Errorf("expected object, got %T", objVal)
		}

		f, ok := obj.Fields[p0.Field]
		if !ok {
			return nil, fmt.Errorf("field %s not found in object", p0.Field)
		}

		return f, nil
	})
}

func (i *interpreterPlanner) VisitExpressionParenthesized(p0 *parse.ExpressionParenthesized) any {
	return p0.Inner.Accept(i)
}

func (i *interpreterPlanner) VisitExpressionComparison(p0 *parse.ExpressionComparison) any {
	cmpOps, negate := getComparisonOps(p0.Operator)

	left := p0.Left.Accept(i).(exprFunc)
	right := p0.Right.Accept(i).(exprFunc)

	retFn := makeComparisonFunc(left, right, cmpOps[0])

	for _, op := range cmpOps[1:] {
		retFn = makeLogicalFunc(retFn, makeComparisonFunc(left, right, op), false)
	}

	if negate {
		return makeUnaryFunc(retFn, not)
	}

	return retFn
}

// makeComparisonFunc returns a function that compares two values.
func makeComparisonFunc(left, right exprFunc, cmpOps ComparisonOp) exprFunc {
	return func(exec *executionContext) (Value, error) {
		leftVal, err := left(exec)
		if err != nil {
			return nil, err
		}

		rightVal, err := right(exec)
		if err != nil {
			return nil, err
		}

		return leftVal.Compare(rightVal, cmpOps)
	}
}

func (i *interpreterPlanner) VisitExpressionLogical(p0 *parse.ExpressionLogical) any {
	left := p0.Left.Accept(i).(exprFunc)
	right := p0.Right.Accept(i).(exprFunc)
	and := p0.Operator == parse.LogicalOperatorAnd

	return makeLogicalFunc(left, right, and)
}

// makeLogicalFunc returns a function that performs a logical operation.
// If and is true, it performs an AND operation, otherwise it performs an OR operation.
func makeLogicalFunc(left, right exprFunc, and bool) exprFunc {
	return func(exec *executionContext) (Value, error) {
		leftVal, err := left(exec)
		if err != nil {
			return nil, err
		}

		rightVal, err := right(exec)
		if err != nil {
			return nil, err
		}

		if leftVal.Type() != types.BoolType || rightVal.Type() != types.BoolType {
			return nil, fmt.Errorf("expected bools, got %s and %s", leftVal.Type(), rightVal.Type())
		}

		if _, ok := leftVal.(*NullValue); ok {
			return leftVal, nil
		}

		if _, ok := rightVal.(*NullValue); ok {
			return rightVal, nil
		}

		if and {
			return &BoolValue{
				Val: leftVal.RawValue().(bool) && rightVal.RawValue().(bool),
			}, nil
		}

		return &BoolValue{
			Val: leftVal.RawValue().(bool) || rightVal.RawValue().(bool),
		}, nil
	}
}

func (i *interpreterPlanner) VisitExpressionArithmetic(p0 *parse.ExpressionArithmetic) any {
	op := convertArithmeticOp(p0.Operator)

	leftFn := p0.Left.Accept(i).(exprFunc)
	rightFn := p0.Right.Accept(i).(exprFunc)
	return exprFunc(func(exec *executionContext) (Value, error) {
		left, err := leftFn(exec)
		if err != nil {
			return nil, err
		}

		right, err := rightFn(exec)
		if err != nil {
			return nil, err
		}

		leftScalar, ok := left.(ScalarValue)
		if !ok {
			return nil, fmt.Errorf("expected scalar, got %T", left)
		}

		rightScalar, ok := right.(ScalarValue)
		if !ok {
			return nil, fmt.Errorf("expected scalar, got %T", right)
		}

		return leftScalar.Arithmetic(rightScalar, op)
	})
}

func (i *interpreterPlanner) VisitExpressionUnary(p0 *parse.ExpressionUnary) any {
	op := convertUnaryOp(p0.Operator)
	val := p0.Expression.Accept(i).(exprFunc)
	return makeUnaryFunc(val, op)
}

// makeUnaryFunc returns a function that performs a unary operation.
func makeUnaryFunc(val exprFunc, op UnaryOp) exprFunc {
	return exprFunc(func(exec *executionContext) (Value, error) {
		v, err := val(exec)
		if err != nil {
			return nil, err
		}

		vScalar, ok := v.(ScalarValue)
		if !ok {
			return nil, fmt.Errorf("%w: expected scalar, got %T", ErrUnaryOnNonScalar, v)
		}

		return vScalar.Unary(op)
	})
}

func (i *interpreterPlanner) VisitExpressionIs(p0 *parse.ExpressionIs) any {
	left := p0.Left.Accept(i).(exprFunc)
	right := p0.Right.Accept(i).(exprFunc)

	op := is
	if p0.Distinct {
		op = isDistinctFrom
	}

	retFn := makeComparisonFunc(left, right, op)

	if p0.Not {
		return makeUnaryFunc(retFn, not)
	}

	return retFn
}

/*
Role management
*/
func (i *interpreterPlanner) VisitGrantOrRevokeStatement(p0 *parse.GrantOrRevokeStatement) any {
	return stmtFunc(func(exec *executionContext, fn func([]Value) error) error {
		if !exec.accessController.HasPrivilege(exec.txCtx.Caller, nil, RolesPrivilege) {
			return fmt.Errorf("%w: %s", ErrDoesNotHavePriv, RolesPrivilege)
		}

		switch {
		case len(p0.Privileges) > 0 && p0.ToRole != "":
			fn := exec.accessController.GrantPrivileges
			if !p0.IsGrant {
				fn = exec.accessController.RevokePrivileges
			}
			return fn(exec.txCtx.Ctx, exec.db, p0.ToRole, p0.Privileges, p0.Namespace)
		case p0.GrantRole != "" && p0.ToUser != "":
			fn := exec.accessController.AssignRole
			if !p0.IsGrant {
				fn = exec.accessController.UnassignRole
			}
			return fn(exec.txCtx.Ctx, exec.db, p0.ToUser, p0.GrantRole)
		default:
			// failure to hit these cases should have been caught by the parser, where better error
			// messages can be generated. This is a catch-all for any other invalid cases.
			return fmt.Errorf("invalid grant/revoke statement")
		}
	})
}

func (i *interpreterPlanner) VisitCreateRoleStatement(p0 *parse.CreateRoleStatement) any {
	return stmtFunc(func(exec *executionContext, fn func([]Value) error) error {
		if !exec.accessController.HasPrivilege(exec.txCtx.Caller, nil, RolesPrivilege) {
			return fmt.Errorf("%w: %s", ErrDoesNotHavePriv, RolesPrivilege)
		}

		if p0.IfNotExists {
			if exec.accessController.RoleExists(p0.Role) {
				return nil
			}
		}

		return exec.accessController.CreateRole(exec.txCtx.Ctx, exec.db, p0.Role)
	})
}

func (i *interpreterPlanner) VisitDropRoleStatement(p0 *parse.DropRoleStatement) any {
	return stmtFunc(func(exec *executionContext, fn func([]Value) error) error {
		if !exec.accessController.HasPrivilege(exec.txCtx.Caller, nil, RolesPrivilege) {
			return fmt.Errorf("%w: %s", ErrDoesNotHavePriv, RolesPrivilege)
		}

		if p0.IfExists {
			if !exec.accessController.RoleExists(p0.Role) {
				return nil
			}
		}

		return exec.accessController.DeleteRole(exec.txCtx.Ctx, exec.db, p0.Role)
	})
}

func (i *interpreterPlanner) VisitTransferOwnershipStatement(p0 *parse.TransferOwnershipStatement) any {
	return stmtFunc(func(exec *executionContext, fn func([]Value) error) error {
		if !exec.accessController.IsOwner(exec.txCtx.Caller) {
			return fmt.Errorf("%w: %s", ErrDoesNotHavePriv, "caller must be owner")
		}

		return exec.accessController.TransferOwnership(exec.txCtx.Ctx, exec.db, p0.To)
	})
}

/*
	top-level adhoc
*/

func (i *interpreterPlanner) VisitSQLStatement(p0 *parse.SQLStatement) any {
	mutatesState := true
	var privilege privilege
	switch p0.SQL.(type) {
	case *parse.InsertStatement:
		privilege = InsertPrivilege
	case *parse.UpdateStatement:
		privilege = UpdatePrivilege
	case *parse.DeleteStatement:
		privilege = DeletePrivilege
	case *parse.SelectStatement:
		privilege = SelectPrivilege
		mutatesState = false
	default:
		panic(fmt.Errorf("unexpected SQL statement type: %T", p0.SQL))
	}
	raw, err := p0.Raw()
	if err != nil {
		panic(err)
	}
	return stmtFunc(func(exec *executionContext, fn func([]Value) error) error {
		if !exec.hasPrivilege(privilege) {
			return fmt.Errorf("%w: %s", ErrDoesNotHavePriv, privilege)
		}

		// if the query is trying to mutate state but the exec ctx cant then we should error
		if mutatesState && !exec.mutatingState {
			return fmt.Errorf("%w: SQL statement mutates state, but the execution context is read-only: %s", ErrStatementMutatesState, raw)
		}

		return exec.query(raw, func(rv *RecordValue) error {
			vals := make([]Value, len(rv.Order))
			for i, field := range rv.Order {
				vals[i] = rv.Fields[field]
			}

			return fn(vals)
		})
	})
}

// below this, I have all visitors that are SQL specific. We don't need to implement them,
// since we will have separate handling for SQL statements at a later stage.

func (i *interpreterPlanner) VisitExpressionColumn(p0 *parse.ExpressionColumn) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitExpressionCollate(p0 *parse.ExpressionCollate) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitExpressionStringComparison(p0 *parse.ExpressionStringComparison) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitExpressionIn(p0 *parse.ExpressionIn) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitExpressionBetween(p0 *parse.ExpressionBetween) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitExpressionSubquery(p0 *parse.ExpressionSubquery) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitExpressionCase(p0 *parse.ExpressionCase) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitCommonTableExpression(p0 *parse.CommonTableExpression) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitSelectStatement(p0 *parse.SelectStatement) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitSelectCore(p0 *parse.SelectCore) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitResultColumnExpression(p0 *parse.ResultColumnExpression) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitResultColumnWildcard(p0 *parse.ResultColumnWildcard) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitRelationTable(p0 *parse.RelationTable) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitRelationSubquery(p0 *parse.RelationSubquery) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitJoin(p0 *parse.Join) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitUpdateStatement(p0 *parse.UpdateStatement) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitUpdateSetClause(p0 *parse.UpdateSetClause) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitDeleteStatement(p0 *parse.DeleteStatement) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitInsertStatement(p0 *parse.InsertStatement) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitUpsertClause(p0 *parse.OnConflict) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitOrderingTerm(p0 *parse.OrderingTerm) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitActionStmtSQL(p0 *parse.ActionStmtSQL) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitActionStmtExtensionCall(p0 *parse.ActionStmtExtensionCall) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitActionStmtActionCall(p0 *parse.ActionStmtActionCall) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitIfThen(p0 *parse.IfThen) any {
	// we handle this directly in VisitProcedureStmtIf
	panic("VisitIfThen should never be called by the interpreter")
}

func (i *interpreterPlanner) VisitWindowImpl(p0 *parse.WindowImpl) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitWindowReference(p0 *parse.WindowReference) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitExpressionWindowFunctionCall(p0 *parse.ExpressionWindowFunctionCall) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitAlterTableStatement(p0 *parse.AlterTableStatement) any {
	panic("intepreter planner should not be called for SQL expressions")
}

func (i *interpreterPlanner) VisitCreateTableStatement(p0 *parse.CreateTableStatement) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitDropTableStatement(p0 *parse.DropTableStatement) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitCreateIndexStatement(p0 *parse.CreateIndexStatement) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitDropIndexStatement(p0 *parse.DropIndexStatement) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitAddColumnStatement(p0 *parse.AddColumn) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitSetColumnConstraint(p0 *parse.AlterColumnSet) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitDropColumnConstraint(p0 *parse.AlterColumnDrop) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitAddColumn(p0 *parse.AddColumn) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitDropColumn(p0 *parse.DropColumn) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitRenameColumn(p0 *parse.RenameColumn) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitRenameTable(p0 *parse.RenameTable) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitAddTableConstraint(p0 *parse.AddTableConstraint) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitDropTableConstraint(p0 *parse.DropTableConstraint) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitTableIndex(p0 *parse.TableIndex) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitColumn(p0 *parse.Column) any {
	panic("TODO: Implement")
}
