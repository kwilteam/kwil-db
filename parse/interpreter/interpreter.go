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
	"github.com/kwilteam/kwil-db/parse/common"
)

func Run(ctx context.Context, proc *types.Procedure, schema *types.Schema, args []any, maxCost int64, costTable *CostTable) (*ProcedureRunResult, error) {
	parseResult, err := parse.ParseProcedure(proc, schema)
	if err != nil {
		return nil, err
	}
	if parseResult.ParseErrs.Err() != nil {
		return nil, parseResult.ParseErrs.Err()
	}

	i := &interpreterPlanner{}

	exec := &executionContext{
		maxCost:   maxCost,
		scope:     newScope(),
		costTable: costTable,
	}

	if len(proc.Parameters) != len(args) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(proc.Parameters), len(args))
	}

	for j, arg := range args {
		val, err := common.NewVariable(arg)
		if err != nil {
			return nil, err
		}

		if !proc.Parameters[j].Type.EqualsStrict(val.Type()) {
			return nil, fmt.Errorf("expected argument %d to be %s, got %s", j+1, proc.Parameters[j].Type, val.Type())
		}

		err = exec.allocateVariable(proc.Parameters[j].Name, val)
		if err != nil {
			return nil, err
		}
	}

	var expectedShape []*types.DataType
	if proc.Returns != nil {
		for _, f := range proc.Returns.Fields {
			expectedShape = append(expectedShape, f.Type)
		}
	}

	res := newReturnableCursor(expectedShape)
	procRes := &ProcedureRunResult{}

	defer func() {
		procRes.Cost = exec.currentCost
	}()

	go func() {
		for _, stmt := range parseResult.AST {
			err := stmt.Accept(i).(stmtFunc)(ctx, exec, res)
			if err != nil {
				res.Err() <- err
				return
			}
		}

		err := res.Close()
		if err != nil {
			// TODO: use a logger
			// Currently this close method's implementation cannot return an error
			// so it's safe to ignore this error while developing.
			panic(fmt.Errorf("error closing cursor: %w", err))
		}
	}()

	// a procedure can return 3 things:
	// - nothing
	// - a single row
	// - any number of rows (a table)
	if proc.Returns == nil {
		// should return nothing
		_, done, err := res.Next(ctx)
		if err != nil {
			return nil, err
		}
		if !done {
			return nil, fmt.Errorf("unexpected return value")
		}
	} else if proc.Returns.IsTable {
		// should return 0 to many rows
		for {
			vals, done, err := res.Next(ctx)
			if err != nil && err != errReturn {
				return nil, err
			}
			if done {
				break
			}

			named, err := makeNamedReturns(proc.Returns.Fields, vals)
			if err != nil {
				return nil, err
			}

			procRes.Values = append(procRes.Values, named)

			if err == errReturn {
				break
			}
		}
	} else {
		// should return a single row
		vals, done, err := res.Next(ctx)
		// should always be errReturn?
		if err != nil && err != errReturn {
			return nil, err
		}
		if done {
			return nil, fmt.Errorf("expected return value")
		}

		named, err := makeNamedReturns(proc.Returns.Fields, vals)
		if err != nil {
			return nil, err
		}

		procRes.Values = append(procRes.Values, named)

		// check if there are more return values
		_, done, err = res.Next(ctx)
		if err != nil && err != errReturn {
			return nil, err
		}
		if !done {
			return nil, fmt.Errorf("unexpected return value")
		}
	}

	return procRes, nil
}

func makeNamedReturns(expected []*types.NamedType, record common.RecordValue) ([]*NamedValue, error) {
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
	Value common.Value
}

type ProcedureRunResult struct {
	Cost   int64
	Values [][]*NamedValue
}

// procedureCallFunc is a function that is generated for a procedure call.
// It is used to call a procedure.
type procedureCallFunc func(ctx context.Context, exec *executionContext, args []common.Value) (Cursor, error)

func (i *interpreterPlanner) makeProcedureCallFunc(procAst []parse.ProcedureStmt, procParams []*types.NamedType, procReturns *types.ProcedureReturn) procedureCallFunc {
	stmtFns := make([]stmtFunc, len(procAst))
	for j, stmt := range procAst {
		stmtFns[j] = stmt.Accept(i).(stmtFunc)
	}

	var expectedShape []*types.DataType
	if procReturns != nil {
		for _, f := range procReturns.Fields {
			expectedShape = append(expectedShape, f.Type)
		}
	}

	return func(ctx context.Context, exec *executionContext, args []common.Value) (Cursor, error) {
		if len(procParams) != len(args) {
			return nil, fmt.Errorf("expected %d arguments, got %d", len(procParams), len(args))
		}

		ret := newReturnableCursor(expectedShape)

		oldScope := exec.scope
		defer func() {
			exec.scope = oldScope
		}()

		// procedures cannot access variables from the parent scope, so we create a new scope
		// TODO: handle @foreign_caller
		exec.scope = newScope()

		for j, arg := range args {
			if !procParams[j].Type.EqualsStrict(arg.Type()) {
				return nil, fmt.Errorf("expected argument %d to be %s, got %s", j+1, procParams[j].Type, arg.Type())
			}

			err := exec.allocateVariable(procParams[j].Name, arg)
			if err != nil {
				return nil, err
			}
		}

		err := executeBlock(ctx, exec, ret, nil, stmtFns)
		if err != nil {
			return nil, err
		}

		return ret, nil
	}
}

// interpreterPlanner is a basic interpreterPlanner for Kuneiform procedures.
type interpreterPlanner struct {
	// schema is the database schema.
	schema *types.Schema
	// procedures are the asts
	// TODO: make a map
}

var (

	// errBreak is an error returned when a break statement is encountered.
	errBreak = errors.New("break")
	// errReturn is an error returned when a return statement is encountered.
	errReturn = errors.New("return")
)

type stmtFunc func(ctx context.Context, exec *executionContext, ret returnChans) error

func (i *interpreterPlanner) VisitProcedureStmtDeclaration(p0 *parse.ProcedureStmtDeclaration) any {
	return stmtFunc(func(ctx context.Context, exec *executionContext, ret returnChans) error {
		return exec.allocateVariable(p0.Variable.Name, common.NewNullValue(p0.Type))
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
	return stmtFunc(func(ctx context.Context, exec *executionContext, ret returnChans) error {
		val, err := valFn(ctx, exec)
		if err != nil {
			return err
		}

		switch a := p0.Variable.(type) {
		case *parse.ExpressionVariable:
			return exec.setVariable(a.Name, val)
		case *parse.ExpressionArrayAccess:
			scalarVal, ok := val.(common.ScalarValue)
			if !ok {
				return fmt.Errorf("expected scalar value, got %T", val)
			}

			arrVal, err := arrFn(ctx, exec)
			if err != nil {
				return err
			}

			arr, ok := arrVal.(common.ArrayValue)
			if !ok {
				return fmt.Errorf("expected array, got %T", arrVal)
			}

			err = exec.Spend(exec.costTable.ArrayAccessCost + exec.costTable.SetVariableCost)
			if err != nil {
				return err
			}

			index, err := indexFn(ctx, exec)
			if err != nil {
				return err
			}

			if !index.Type().EqualsStrict(types.IntType) {
				return fmt.Errorf("array index must be integer, got %s", index.Type())
			}

			return arr.Set(index.Value().(int64), scalarVal)
		default:
			return fmt.Errorf("unexpected assignable variable type: %T", p0.Variable)
		}
	})
}

func (i *interpreterPlanner) VisitProcedureStmtCall(p0 *parse.ProcedureStmtCall) any {
	panic("TODO: Implement")
}

// executeBlock executes a block of statements with their own sub-scope.
// It takes a list of statements, and a list of variable allocations that will be made in the sub-scope.
func executeBlock(ctx context.Context, exec *executionContext, ret returnChans,
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
		if err := stmt(ctx, exec, ret); err != nil {
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

	return stmtFunc(func(ctx context.Context, exec *executionContext, ret returnChans) error {
		looper, err := loopFn(ctx, exec)
		if err != nil {
			return err
		}
		defer looper.Close()

		for {
			term, done, err := looper.Next(ctx)
			if err != nil {
				return err
			}
			if done {
				break
			}

			err = exec.Spend(exec.costTable.LoopCost)
			if err != nil {
				return err
			}

			err = executeBlock(ctx, exec, ret, []*NamedValue{
				{
					Name:  p0.Receiver.Name,
					Value: term,
				},
			}, stmtFns)
			if err != nil {
				if err == errBreak {
					break
				} else {
					return err
				}
			}
		}

		return nil
	})
}

// loopTermFunc is a function that allows iterating over a loop term.
// It returns a function that returns the next value in the loop term.
type loopTermFunc func(ctx context.Context, exec *executionContext) (loop loopReturn, err error)

func (i *interpreterPlanner) VisitLoopTermRange(p0 *parse.LoopTermRange) any {
	startFn := p0.Start.Accept(i).(exprFunc)
	endFn := p0.End.Accept(i).(exprFunc)

	return loopTermFunc(func(ctx context.Context, exec *executionContext) (loop loopReturn, err error) {
		start, err := startFn(ctx, exec)
		if err != nil {
			return nil, err
		}

		end, err := endFn(ctx, exec)
		if err != nil {
			return nil, err
		}

		if !start.Type().EqualsStrict(types.IntType) {
			return nil, fmt.Errorf("expected integer, got %s", start.Type())
		}

		if !end.Type().EqualsStrict(types.IntType) {
			return nil, fmt.Errorf("expected integer, got %s", end.Type())
		}

		return &rangeLooper{
			end:     end.Value().(int64),
			current: start.Value().(int64),
		}, nil
	})
}

type rangeLooper struct {
	end     int64
	current int64
}

func (r *rangeLooper) Next(ctx context.Context) (common.Value, bool, error) {
	if r.current > r.end {
		return nil, true, nil
	}

	ret := r.current
	r.current++
	return &common.IntValue{
		Val: ret,
	}, false, nil
}

func (r *rangeLooper) Close() error {
	return nil
}

func (i *interpreterPlanner) VisitLoopTermSQL(p0 *parse.LoopTermSQL) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitLoopTermVariable(p0 *parse.LoopTermVariable) any {
	return loopTermFunc(func(ctx context.Context, exec *executionContext) (loop loopReturn, err error) {
		val, err := exec.getVariable(p0.Variable.Name)
		if err != nil {
			return nil, err
		}

		arr, ok := val.(common.ArrayValue)
		if !ok {
			return nil, fmt.Errorf("expected array, got %T", val)
		}

		return &arrayLooper{
			arr:   arr,
			index: 1, // all arrays are 1-indexed
		}, nil
	})
}

// loopReturn is an interface for iterating over the result of a loop term.
type loopReturn interface {
	Next(ctx context.Context) (common.Value, bool, error)
	Close() error
}

type arrayLooper struct {
	arr   common.ArrayValue
	index int64
}

func (a *arrayLooper) Next(ctx context.Context) (common.Value, bool, error) {
	ret, err := a.arr.Index(a.index)
	if err != nil {
		if err == common.ErrIndexOutOfBounds {
			return nil, true, nil
		}
		return nil, false, err
	}

	a.index++
	return ret, false, nil
}

func (a *arrayLooper) Close() error {
	return nil
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

	return stmtFunc(func(ctx context.Context, exec *executionContext, ret returnChans) error {
		branchRun := false // tracks if a branch has been run
		for _, ifThen := range ifThenFns {
			if branchRun {
				break
			}

			cond, err := ifThen.If(ctx, exec)
			if err != nil {
				return err
			}

			switch c := cond.(type) {
			case *common.BoolValue:
				if !c.Val {
					continue
				}
			case *common.NullValue:
				continue
			default:
				return fmt.Errorf("expected bool, got %s", c.Type())
			}

			branchRun = true

			err = executeBlock(ctx, exec, ret, nil, ifThen.Then)
			if err != nil {
				return err
			}
		}

		if !branchRun && p0.Else != nil {
			err := executeBlock(ctx, exec, ret, nil, elseFns)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (i *interpreterPlanner) VisitProcedureStmtSQL(p0 *parse.ProcedureStmtSQL) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitProcedureStmtBreak(p0 *parse.ProcedureStmtBreak) any {
	return stmtFunc(func(ctx context.Context, exec *executionContext, ret returnChans) error {
		if err := exec.Spend(exec.costTable.BreakCost); err != nil {
			return err
		}

		return errBreak
	})
}

func (i *interpreterPlanner) VisitProcedureStmtReturn(p0 *parse.ProcedureStmtReturn) any {
	valFns := make([]exprFunc, len(p0.Values))
	for j, v := range p0.Values {
		valFns[j] = v.Accept(i).(exprFunc)
	}

	return stmtFunc(func(ctx context.Context, exec *executionContext, ret returnChans) error {
		if err := exec.Spend(exec.costTable.ReturnCost); err != nil {
			return err
		}

		vals := make([]common.Value, len(p0.Values))
		for j, valFn := range valFns {
			val, err := valFn(ctx, exec)
			if err != nil {
				ret.Err() <- err
				return err
			}

			vals[j] = val
		}

		ret.Record() <- vals
		return errReturn
	})
}

func (i *interpreterPlanner) VisitProcedureStmtReturnNext(p0 *parse.ProcedureStmtReturnNext) any {
	valFns := make([]exprFunc, len(p0.Values))
	for j, v := range p0.Values {
		valFns[j] = v.Accept(i).(exprFunc)
	}

	return stmtFunc(func(ctx context.Context, exec *executionContext, ret returnChans) error {
		if err := exec.Spend(exec.costTable.ReturnCost); err != nil {
			return err
		}

		vals := make([]common.Value, len(p0.Values))
		for j, valFn := range valFns {
			val, err := valFn(ctx, exec)
			if err != nil {
				ret.Err() <- err
				return err
			}

			vals[j] = val
		}

		ret.Record() <- vals

		// we don't return an errReturn or mark done here because return next is not the last statement in a procedure.
		return nil
	})
}

// everything in this section is for expressions, which evaluate to exactly one value.

// exprFunc is a function that returns a value.
type exprFunc func(ctx context.Context, exec *executionContext) (common.Value, error)

func (i *interpreterPlanner) VisitExpressionLiteral(p0 *parse.ExpressionLiteral) any {
	return exprFunc(func(ctx context.Context, exec *executionContext) (common.Value, error) {
		return common.NewVariable(p0.Value)
	})
}

func (i *interpreterPlanner) VisitExpressionFunctionCall(p0 *parse.ExpressionFunctionCall) any {
	args := make([]exprFunc, len(p0.Args))
	for j, arg := range p0.Args {
		args[j] = arg.Accept(i).(exprFunc)
	}

	// can be a built-in function or a user-defined procedure
	funcDef, ok := common.Functions[p0.Name]
	if ok {
		scalarFunc, ok := funcDef.(*common.ScalarFunctionDefinition)
		if !ok {
			panic("cannot call non-scalar function in procedure")
		}

		return makeBuiltInFunctionCall(scalarFunc, args)
	}
	panic("TODO: Implement")
	// // otherwise, it must be a user defined procedure.
	// proc, ok := i.schema.FindProcedure(p0.Name)
	// if !ok {
	// 	panic("procedure not found")
	// }

	// return exprFunc(func(ctx context.Context, exec *executionContext) (common.Value, error) {

	// })
}

func makeBuiltInFunctionCall(funcDef *common.ScalarFunctionDefinition, args []exprFunc) exprFunc {
	return exprFunc(func(ctx context.Context, exec *executionContext) (common.Value, error) {
		err := exec.Spend(exec.costTable.CallBuiltInFunctionCost)
		if err != nil {
			return nil, err
		}

		vals := make([]common.Value, len(args))
		for i, arg := range args {
			val, err := arg(ctx, exec)
			if err != nil {
				return nil, err
			}

			vals[i] = val
		}

		return funcDef.EvaluateFunc(exec, vals)
	})
}

func (i *interpreterPlanner) VisitExpressionForeignCall(p0 *parse.ExpressionForeignCall) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitExpressionVariable(p0 *parse.ExpressionVariable) any {
	return exprFunc(func(ctx context.Context, exec *executionContext) (common.Value, error) {
		return exec.getVariable(p0.Name)
	})
}

func (i *interpreterPlanner) VisitExpressionArrayAccess(p0 *parse.ExpressionArrayAccess) any {
	arrFn := p0.Array.Accept(i).(exprFunc)
	indexFn := p0.Index.Accept(i).(exprFunc)

	return exprFunc(func(ctx context.Context, exec *executionContext) (common.Value, error) {
		err := exec.Spend(exec.costTable.ArrayAccessCost)
		if err != nil {
			return nil, err
		}

		arrVal, err := arrFn(ctx, exec)
		if err != nil {
			return nil, err
		}

		arr, ok := arrVal.(common.ArrayValue)
		if !ok {
			return nil, fmt.Errorf("expected array, got %T", arrVal)
		}

		index, err := indexFn(ctx, exec)
		if err != nil {
			return nil, err
		}

		if !index.Type().EqualsStrict(types.IntType) {
			return nil, fmt.Errorf("array index must be integer, got %s", index.Type())
		}

		return arr.Index(index.Value().(int64))
	})
}

func (i *interpreterPlanner) VisitExpressionMakeArray(p0 *parse.ExpressionMakeArray) any {
	valFns := make([]exprFunc, len(p0.Values))
	for j, v := range p0.Values {
		valFns[j] = v.Accept(i).(exprFunc)
	}

	return exprFunc(func(ctx context.Context, exec *executionContext) (common.Value, error) {
		err := exec.Spend(exec.costTable.MakeArrayCost)
		if err != nil {
			return nil, err
		}

		if len(valFns) == 0 {
			return nil, fmt.Errorf("array must have at least one element")
		}

		val0, err := valFns[0](ctx, exec)
		if err != nil {
			return nil, err
		}

		scal, ok := val0.(common.ScalarValue)
		if !ok {
			return nil, fmt.Errorf("expected scalar value, got %T", val0)
		}

		var vals []common.ScalarValue
		for j, valFn := range valFns {
			if j == 0 {
				continue
			}

			val, err := valFn(ctx, exec)
			if err != nil {
				return nil, err
			}

			scal, ok := val.(common.ScalarValue)
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

	return exprFunc(func(ctx context.Context, exec *executionContext) (common.Value, error) {
		err := exec.Spend(exec.costTable.GetVariableCost)
		if err != nil {
			return nil, err
		}

		objVal, err := recordFn(ctx, exec)
		if err != nil {
			return nil, err
		}

		obj, ok := objVal.(*common.RecordValue)
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
	cmpOps, negate := parse.GetComparisonOps(p0.Operator)

	left := p0.Left.Accept(i).(exprFunc)
	right := p0.Right.Accept(i).(exprFunc)

	retFn := makeComparisonFunc(left, right, cmpOps[0])

	for _, op := range cmpOps[1:] {
		retFn = makeLogicalFunc(retFn, makeComparisonFunc(left, right, op), false)
	}

	if negate {
		return makeUnaryFunc(retFn, common.Not)
	}

	return retFn
}

// makeComparisonFunc returns a function that compares two values.
func makeComparisonFunc(left, right exprFunc, cmpOps common.ComparisonOp) exprFunc {
	return func(ctx context.Context, exec *executionContext) (common.Value, error) {
		if err := exec.Spend(exec.costTable.ComparisonCost); err != nil {
			return nil, err
		}

		leftVal, err := left(ctx, exec)
		if err != nil {
			return nil, err
		}

		rightVal, err := right(ctx, exec)
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
	return func(ctx context.Context, exec *executionContext) (common.Value, error) {
		if err := exec.Spend(exec.costTable.LogicalCost); err != nil {
			return nil, err
		}

		leftVal, err := left(ctx, exec)
		if err != nil {
			return nil, err
		}

		rightVal, err := right(ctx, exec)
		if err != nil {
			return nil, err
		}

		if leftVal.Type() != types.BoolType || rightVal.Type() != types.BoolType {
			return nil, fmt.Errorf("expected bools, got %s and %s", leftVal.Type(), rightVal.Type())
		}

		if _, ok := leftVal.(*common.NullValue); ok {
			return leftVal, nil
		}

		if _, ok := rightVal.(*common.NullValue); ok {
			return rightVal, nil
		}

		if and {
			return &common.BoolValue{
				Val: leftVal.Value().(bool) && rightVal.Value().(bool),
			}, nil
		}

		return &common.BoolValue{
			Val: leftVal.Value().(bool) || rightVal.Value().(bool),
		}, nil
	}
}

func (i *interpreterPlanner) VisitExpressionArithmetic(p0 *parse.ExpressionArithmetic) any {

	op := parse.ConvertArithmeticOp(p0.Operator)

	leftFn := p0.Left.Accept(i).(exprFunc)
	rightFn := p0.Right.Accept(i).(exprFunc)
	return exprFunc(func(ctx context.Context, exec *executionContext) (common.Value, error) {
		if err := exec.Spend(exec.costTable.ArithmeticCost); err != nil {
			return nil, err
		}

		left, err := leftFn(ctx, exec)
		if err != nil {
			return nil, err
		}

		right, err := rightFn(ctx, exec)
		if err != nil {
			return nil, err
		}

		leftScalar, ok := left.(common.ScalarValue)
		if !ok {
			return nil, fmt.Errorf("expected scalar, got %T", left)
		}

		rightScalar, ok := right.(common.ScalarValue)
		if !ok {
			return nil, fmt.Errorf("expected scalar, got %T", right)
		}

		return leftScalar.Arithmetic(rightScalar, op)
	})
}

func (i *interpreterPlanner) VisitExpressionUnary(p0 *parse.ExpressionUnary) any {
	op := parse.ConvertUnaryOp(p0.Operator)
	val := p0.Expression.Accept(i).(exprFunc)
	return makeUnaryFunc(val, op)
}

// makeUnaryFunc returns a function that performs a unary operation.
func makeUnaryFunc(val exprFunc, op common.UnaryOp) exprFunc {
	return exprFunc(func(ctx context.Context, exec *executionContext) (common.Value, error) {
		if err := exec.Spend(exec.costTable.UnaryCost); err != nil {
			return nil, err
		}

		v, err := val(ctx, exec)
		if err != nil {
			return nil, err
		}

		vScalar, ok := v.(common.ScalarValue)
		if !ok {
			return nil, fmt.Errorf("%w: expected scalar, got %T", ErrUnaryOnNonScalar, v)
		}

		return vScalar.Unary(op)
	})
}

func (i *interpreterPlanner) VisitExpressionIs(p0 *parse.ExpressionIs) any {
	left := p0.Left.Accept(i).(exprFunc)
	right := p0.Right.Accept(i).(exprFunc)

	op := common.Is
	if p0.Distinct {
		op = common.IsDistinctFrom
	}

	retFn := makeComparisonFunc(left, right, op)

	if p0.Not {
		return makeUnaryFunc(retFn, common.Not)
	}

	return retFn
}

// below this, I have all visitors that are SQL specific. We don't need to implement them,
// since we will have separate handling for SQL statements at a later stage.

func (i *interpreterPlanner) VisitExpressionColumn(p0 *parse.ExpressionColumn) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitExpressionCollate(p0 *parse.ExpressionCollate) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitExpressionStringComparison(p0 *parse.ExpressionStringComparison) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitExpressionIn(p0 *parse.ExpressionIn) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitExpressionBetween(p0 *parse.ExpressionBetween) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitExpressionSubquery(p0 *parse.ExpressionSubquery) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitExpressionCase(p0 *parse.ExpressionCase) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitCommonTableExpression(p0 *parse.CommonTableExpression) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitSQLStatement(p0 *parse.SQLStatement) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitSelectStatement(p0 *parse.SelectStatement) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitSelectCore(p0 *parse.SelectCore) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitResultColumnExpression(p0 *parse.ResultColumnExpression) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitResultColumnWildcard(p0 *parse.ResultColumnWildcard) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitRelationTable(p0 *parse.RelationTable) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitRelationSubquery(p0 *parse.RelationSubquery) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitRelationFunctionCall(p0 *parse.RelationFunctionCall) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitJoin(p0 *parse.Join) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitUpdateStatement(p0 *parse.UpdateStatement) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitUpdateSetClause(p0 *parse.UpdateSetClause) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitDeleteStatement(p0 *parse.DeleteStatement) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitInsertStatement(p0 *parse.InsertStatement) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitUpsertClause(p0 *parse.UpsertClause) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitOrderingTerm(p0 *parse.OrderingTerm) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitActionStmtSQL(p0 *parse.ActionStmtSQL) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitActionStmtExtensionCall(p0 *parse.ActionStmtExtensionCall) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitActionStmtActionCall(p0 *parse.ActionStmtActionCall) any {
	panic("TODO: Implement")
}

func (i *interpreterPlanner) VisitIfThen(p0 *parse.IfThen) any {
	// we handle this directly in VisitProcedureStmtIf
	panic("VisitIfThen should never be called by the interpreter")
}
