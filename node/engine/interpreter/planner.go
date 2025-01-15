// package interpreter provides a basic interpreter for Kuneiform procedures.
// It allows running procedures as standalone programs (instead of generating
// PL/pgSQL code).
package interpreter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/node/engine"
	"github.com/kwilteam/kwil-db/node/engine/parse"
	pggenerate "github.com/kwilteam/kwil-db/node/engine/pg_generate"
)

// makeActionToExecutable creates an executable from an action
func makeActionToExecutable(namespace string, act *Action) *executable {
	planner := &interpreterPlanner{}
	stmtFns := make([]stmtFunc, len(act.Body))
	for j, stmt := range act.Body {
		stmtFns[j] = stmt.Accept(planner).(stmtFunc)
	}

	validateArgs := func(v []precompiles.Value) error {
		if len(v) != len(act.Parameters) {
			return fmt.Errorf("expected %d arguments, got %d", len(act.Parameters), len(v))
		}

		for i, arg := range v {
			if !act.Parameters[i].Type.EqualsStrict(arg.Type()) {
				return fmt.Errorf("%w: expected argument %d to be %s, got %s", engine.ErrType, i+1, act.Parameters[i].Type, arg.Type())
			}
		}

		return nil
	}

	return &executable{
		Name: act.Name,
		Func: func(exec *executionContext, args []precompiles.Value, fn resultFunc) error {
			if err := exec.canExecute(namespace, act.Name, act.Modifiers); err != nil {
				return err
			}

			// validate the args
			err := validateArgs(args)
			if err != nil {
				return err
			}

			// get the expected return col names
			var returnColNames []string
			if act.Returns != nil {
				for _, f := range act.Returns.Fields {
					cName := f.Name
					if cName == "" {
						cName = unknownColName
					}
					returnColNames = append(returnColNames, cName)
				}
			}

			exec2 := exec.subscope(namespace)

			for j, param := range act.Parameters {
				err = exec2.allocateVariable(param.Name, args[j])
				if err != nil {
					return err
				}
			}

			// execute the statements
			for _, stmt := range stmtFns {
				err := stmt(exec2, func(row *row) error {
					row.columns = returnColNames
					err := fn(row)
					if err != nil {
						return err
					}

					return nil
				})
				switch err {
				case nil:
					// do nothing
				case errReturn:
					// the procedure is done, exit early
					return nil
				default:
					return err
				}
			}

			return nil
		},
		Type: executableTypeAction,
	}
}

// interpreterPlanner creates functions for running Kuneiform logic.
type interpreterPlanner struct{}

var (

	// errBreak is an error returned when a break statement is encountered.
	errBreak = errors.New("break")
	// errContinue is an error returned when a continue statement is encountered.
	errContinue = errors.New("continue")
	// errReturn is an error returned when a return statement is encountered.
	errReturn = errors.New("return")
)

func makeRow(v []precompiles.Value) *row {
	return &row{
		Values: v,
	}
}

// row represents a row of values.
type row struct {
	// columns is a list of column names.
	// It can be nil and/or not match the length of values.
	// The Columns() method should always be used.
	columns []string
	// precompiles.Values is a list of values.
	Values []precompiles.Value
}

func (r *row) record() (*precompiles.RecordValue, error) {
	rec := precompiles.EmptyRecordValue()
	for i, name := range r.Columns() {
		if name == unknownColName {
			continue
		}

		err := rec.AddValue(name, r.Values[i])
		if err != nil {
			return nil, err
		}
	}

	return rec, nil
}

const unknownColName = "?column?"

func (r *row) Columns() []string {
	switch len(r.columns) {
	case 0:
		for range r.Values {
			r.columns = append(r.columns, unknownColName)
		}
		return r.columns
	case len(r.Values):
		return r.columns
	default:
		panic(fmt.Errorf("columns and values do not match: %d columns, %d values", len(r.columns), len(r.Values)))
	}
}

// fillUnnamed fills all empty strings in the columns with the unknown column name.
func (r *row) fillUnnamed() {
	r.Columns() // make sure the columns are initialized
	for i, col := range r.columns {
		if col == "" {
			r.columns[i] = unknownColName
		}
	}
}

type resultFunc func(*row) error

type stmtFunc func(exec *executionContext, fn resultFunc) error

func (i *interpreterPlanner) VisitActionStmtDeclaration(p0 *parse.ActionStmtDeclaration) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		nv, err := precompiles.MakeNull(p0.Type)
		if err != nil {
			return err
		}

		return exec.allocateVariable(p0.Variable.Name, nv)
	})
}

func (i *interpreterPlanner) VisitActionStmtAssignment(p0 *parse.ActionStmtAssign) any {
	valFn := p0.Value.Accept(i).(exprFunc)

	var arrFn exprFunc
	// index in case of array access
	var indexFn exprFunc
	// to and from in case of slice
	var toFn exprFunc
	var fromFn exprFunc
	if a, ok := p0.Variable.(*parse.ExpressionArrayAccess); ok {
		arrFn = a.Array.Accept(i).(exprFunc)
		if a.Index != nil {
			indexFn = a.Index.Accept(i).(exprFunc)
		}
		if a.FromTo != nil {
			if a.FromTo[0] != nil {
				fromFn = a.FromTo[0].Accept(i).(exprFunc)
			}
			if a.FromTo[1] != nil {
				toFn = a.FromTo[1].Accept(i).(exprFunc)
			}
		}
	}
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		val, err := valFn(exec)
		if err != nil {
			return err
		}

		switch a := p0.Variable.(type) {
		case *parse.ExpressionVariable:
			// if p0 has a type, then a variable must either already exist of that type, OR it must be new
			if p0.Type != nil {
				v, err := exec.getVariable(a.Name) // this will error if it does not exist
				// if unknown variable, assign it.
				// if other error, return.
				// if nil error, a var exists, so ensure it is of this type
				switch {
				case errors.Is(err, engine.ErrUnknownVariable):
					err2 := exec.allocateNullVariable(a.Name, p0.Type)
					if err2 != nil {
						return err2
					}
				case err != nil:
					return err
				default:
					if !v.Type().EqualsStrict(p0.Type) {
						return fmt.Errorf(`%w: cannot assign new type "%s" to variable "%s" of type "%s"`, engine.ErrType, p0.Type.String(), a.Name, v.Type().String())
					}
					// then we do nothing, since it is already allocated.
				}
			}

			return exec.setVariable(a.Name, val)
		case *parse.ExpressionArrayAccess:
			if p0.Type != nil {
				return fmt.Errorf(`%w: cannot assign to array element with type assignment "%s"`, engine.ErrType, p0.Type.String())
			}

			arrVal, err := arrFn(exec)
			if err != nil {
				return err
			}
			arr, ok := arrVal.(precompiles.ArrayValue)
			if !ok {
				return fmt.Errorf("%w: expected array, got %T", engine.ErrType, arrVal)
			}

			// if array access, then there are two cases.
			// one is that we are assigning a scalar value to an array element:
			// arr[1] = 5
			// the other is that we are assigning an array to a slice of an array:
			// arr[2:3] = [1, 2]
			// arr[2:] = [1, 2]
			// arr[:3] = [1, 2]

			if indexFn != nil {
				// we are assigning a scalar value to an array element
				scalarVal, ok := val.(precompiles.ScalarValue)
				if !ok {
					return fmt.Errorf("%w: expected scalar value, got %T", engine.ErrType, val)
				}

				index, err := indexFn(exec)
				if err != nil {
					return err
				}

				if index.Null() {
					return fmt.Errorf("%w: array index cannot be null when assigning to array", engine.ErrInvalidNull)
				}

				if !index.Type().EqualsStrict(types.IntType) {
					return fmt.Errorf("array index must be integer, got %s", index.Type())
				}

				err = arr.Set(int32(index.RawValue().(int64)), scalarVal)
				if err != nil {
					return err
				}

				return nil
			}

			evaluateSliceIdx := func(fn exprFunc, defaultVal int32) (int32, error) {
				if fn == nil {
					return defaultVal, nil
				}

				val, err := fn(exec)
				if err != nil {
					return 0, err
				}

				if val.Null() {
					return 0, fmt.Errorf("%w: slice index cannot be null when assigning to array", engine.ErrInvalidNull)
				}

				if !val.Type().EqualsStrict(types.IntType) {
					return 0, fmt.Errorf("array index must be integer, got %s", val.Type())
				}

				return int32(val.RawValue().(int64)), nil
			}

			// we are assigning an array to a slice of an array
			// We will start by evaluating the from and to indices.
			// From there, we will ensure that our new value is of the right length.
			// Finally, we will assign the values.
			from, err := evaluateSliceIdx(fromFn, 1) // default 1 in case of arr[1:]
			if err != nil {
				return err
			}
			to, err := evaluateSliceIdx(toFn, arr.Len()) // default arr.Len() in case of arr[:2]
			if err != nil {
				return err
			}

			if from < 1 {
				return fmt.Errorf("%w: slice from index must be greater than 0, got %d", engine.ErrIndexOutOfBounds, from)
			}

			if to < from {
				return fmt.Errorf("%w: slice to index must be greater than or equal to from index, got %d", engine.ErrIndexOutOfBounds, to)
			}

			// now, we can get the new array and check its length
			newArrVal, err := valFn(exec)
			if err != nil {
				return err
			}

			newArr, ok := newArrVal.(precompiles.ArrayValue)
			if !ok {
				return fmt.Errorf("%w: expected array, got %T", engine.ErrType, newArrVal)
			}

			// to match postgres:
			// if the receiving array is too small, we truncate the new array so that it fits.
			// if the receiving array is too large, we return an error.
			receiveLen := to - from + 1
			newArrLen := newArr.Len()
			if receiveLen > newArrLen {
				return fmt.Errorf("%w: expected slice to have length %d, got %d", engine.ErrArrayTooSmall, receiveLen, newArrLen)
			}

			j := int32(1)
			// finally, we can assign the values
			for i := from; i <= to; i++ {
				newVal, err := newArr.Get(j)
				if err != nil {
					return err
				}

				err = arr.Set(i, newVal)
				if err != nil {
					return err
				}

				j++
			}

			return nil
		default:
			panic(fmt.Errorf("unexpected assignable variable type: %T", p0.Variable))
		}
	})
}

func (i *interpreterPlanner) VisitActionStmtCall(p0 *parse.ActionStmtCall) any {

	// we cannot simply use the same visitor as the expression function call, because expression function
	// calls always return exactly one value. Here, we can return 0 values, many values, or a table.

	receivers := make([]*string, len(p0.Receivers))
	for j, r := range p0.Receivers {
		// if r is nil, we can ignore the receiver.
		if r == nil {
			receivers[j] = nil
			continue
		}
		receivers[j] = &r.Name
	}

	args := make([]exprFunc, len(p0.Call.Args))
	for j, arg := range p0.Call.Args {
		args[j] = arg.Accept(i).(exprFunc)
	}

	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		ns, err := exec.getNamespace(p0.Call.Namespace)
		if err != nil {
			return err
		}

		funcDef, ok := ns.availableFunctions[p0.Call.Name]
		if !ok {
			return fmt.Errorf(`unknown action "%s" in namespace "%s"`, p0.Call.Name, p0.Call.Namespace)
		}

		vals := make([]precompiles.Value, len(args))
		for j, valFn := range args {
			val, err := valFn(exec)
			if err != nil {
				return err
			}

			vals[j] = val
		}

		iter := 0
		err = funcDef.Func(exec, vals, func(row *row) error {
			// if there are receivers and this returns more than 1 value, we should return an error.
			if iter > 0 && len(receivers) > 0 {
				return fmt.Errorf(`%w: expected function or action "%s" to return a single record, but it returned a record set`, engine.ErrReturnShape, funcDef.Name)
			}
			iter++

			// re-verify the returns, since the above checks only for what the function signature
			// says, but this checks what the function actually returns.
			if len(receivers) > len(row.Values) {
				return fmt.Errorf(`%w: expected function or action "%s" to return at least %d values, but it returned %d`, engine.ErrReturnShape, funcDef.Name, len(receivers), len(row.Values))
			}

			for j, r := range receivers {
				if r == nil {
					continue
				}
				err = exec.setVariable(*r, row.Values[j])
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
				return fmt.Errorf(`%w: expected function or action "%s" to return a single record, but it returned nothing`, engine.ErrReturnShape, funcDef.Name)
			}
		}

		return nil
	})
}

// executeBlock executes a block of statements with their own sub-scope.
// It takes a list of statements, and a list of variable allocations that will be made in the sub-scope.
func executeBlock(exec *executionContext, fn resultFunc,
	stmtFuncs []stmtFunc) error {
	exec.scope.child()
	defer exec.scope.popScope()

	for _, stmt := range stmtFuncs {
		err := stmt(exec, fn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *interpreterPlanner) VisitActionStmtForLoop(p0 *parse.ActionStmtForLoop) any {
	stmtFns := make([]stmtFunc, len(p0.Body))
	for j, stmt := range p0.Body {
		stmtFns[j] = stmt.Accept(i).(stmtFunc)
	}

	loopFn := p0.LoopTerm.Accept(i).(loopTermFunc)

	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		err := loopFn(exec, func(term precompiles.Value) error {
			exec.scope.child()
			defer exec.scope.popScope()
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
		if errors.Is(err, errBreak) {
			return nil // swallow break errors and exit
		}
		return err
	})
}

// loopTermFunc is a function that allows iterating over a loop term.
// It calls the function passed to it with each value.
type loopTermFunc func(exec *executionContext, fn func(precompiles.Value) error) (err error)

// handleLoopTermErr is a helper function that handles the error returned by a loop term.
// If it is a continue, it will return nil. If it is a break, it will bubble it up.
// Otherwise, it will return the error.
func handleLoopTermErr(err error) error {
	if errors.Is(err, errContinue) {
		return nil
	}
	return err
}

func (i *interpreterPlanner) VisitLoopTermRange(p0 *parse.LoopTermRange) any {
	startFn := p0.Start.Accept(i).(exprFunc)
	endFn := p0.End.Accept(i).(exprFunc)

	return loopTermFunc(func(exec *executionContext, fn func(precompiles.Value) error) (err error) {
		start, err := startFn(exec)
		if err != nil {
			return err
		}

		end, err := endFn(exec)
		if err != nil {
			return err
		}

		if start.Null() || end.Null() {
			return nil
		}

		if !start.Type().EqualsStrict(types.IntType) {
			return fmt.Errorf("%w: expected integer, got %s", engine.ErrType, start.Type())
		}

		if !end.Type().EqualsStrict(types.IntType) {
			return fmt.Errorf("%w: expected integer, got %s", engine.ErrType, end.Type())
		}

		for i := start.RawValue().(int64); i <= end.RawValue().(int64); i++ {
			err = handleLoopTermErr(fn(precompiles.MakeInt8(i)))
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (i *interpreterPlanner) VisitLoopTermExpression(p0 *parse.LoopTermExpression) any {
	expr := p0.Expression.Accept(i).(exprFunc)
	return loopTermFunc(func(exec *executionContext, fn func(precompiles.Value) error) error {
		// there are two cases for expressions here.
		// The first is that the expression is calling a table-returning function.
		// The second is that the expression returns an array.
		// In the second case, we should verify that p0.Array is true.

		// check if the expression is a function call
		functionCall, ok := p0.Expression.(*parse.ExpressionFunctionCall)
		// Even if it is a function call, if p0.Array is true, then we should expect it to return a single array.
		if ok && !p0.Array {
			// user did not specify IN ARRAY and it is a function.

			ns, err := exec.getNamespace(functionCall.Namespace)
			if err != nil {
				return err
			}

			funcDef, ok := ns.availableFunctions[functionCall.Name]
			if !ok {
				return fmt.Errorf(`unknown function "%s" in namespace "%s"`, functionCall.Name, functionCall.Namespace)
			}

			vals := make([]precompiles.Value, len(functionCall.Args))
			for j, arg := range functionCall.Args {
				val, err := arg.Accept(i).(exprFunc)(exec)
				if err != nil {
					return err
				}

				vals[j] = val
			}

			err = funcDef.Func(exec, vals, func(row *row) error {
				rec, err := row.record()
				if err != nil {
					return err
				}

				return handleLoopTermErr(fn(rec))
			})
			if err != nil {
				return err
			}

			return nil
		}

		// expect the expression to return a single array
		// If the user did not specify this, we should return an error.
		if !p0.Array {
			return fmt.Errorf("%w: must use IN ARRAY when looping over anything that is not a function, integer range, or SQL statement", engine.ErrLoop)
		}

		val, err := expr(exec)
		if err != nil {
			return err
		}

		if val.Null() {
			return nil
		}

		arr, ok := val.(precompiles.ArrayValue)
		if !ok {
			return fmt.Errorf("%w: expected array, got %T", engine.ErrType, val)
		}

		for i := range arr.Len() {
			scalar, err := arr.Get(i + 1) // all arrays are 1-indexed
			if err != nil {
				return err
			}

			err = handleLoopTermErr(fn(scalar))
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (i *interpreterPlanner) VisitLoopTermSQL(p0 *parse.LoopTermSQL) any {
	return loopTermFunc(func(exec *executionContext, fn func(precompiles.Value) error) error {
		raw, err := p0.Statement.Raw()
		if err != nil {
			return err
		}

		// query executes a Kuneiform query and returns a cursor.
		return exec.query(raw, func(r *row) error {
			rec, err := r.record()
			if err != nil {
				return err
			}

			return handleLoopTermErr(fn(rec))
		})
	})
}

func (i *interpreterPlanner) VisitActionStmtIf(p0 *parse.ActionStmtIf) any {
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

	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		branchRun := false // tracks if any IF branch has been run
		for _, ifThen := range ifThenFns {
			if branchRun {
				break
			}

			cond, err := ifThen.If(exec)
			if err != nil {
				return err
			}

			if cond.Null() {
				continue
			}
			if boolVal, ok := cond.(*precompiles.BoolValue); ok {
				if boolVal.Null() {
					continue
				}
				if !boolVal.Bool.Bool {
					continue
				}
			} else {
				return fmt.Errorf("%w: IF clause expects type bool, got %s", engine.ErrType, cond.Type())
			}

			branchRun = true

			err = executeBlock(exec, fn, ifThen.Then)
			if err != nil {
				return err
			}
		}

		if !branchRun && p0.Else != nil {
			err := executeBlock(exec, fn, elseFns)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (i *interpreterPlanner) VisitActionStmtSQL(p0 *parse.ActionStmtSQL) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		raw, err := p0.SQL.Raw()
		if err != nil {
			return err
		}

		// query executes any arbitrary SQL.
		err = exec.query(raw, func(rv *row) error {
			// we ignore results here since we are not returning anything.
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func (i *interpreterPlanner) VisitActionStmtLoopControl(p0 *parse.ActionStmtLoopControl) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		switch p0.Type {
		case parse.LoopControlTypeBreak:
			return errBreak
		case parse.LoopControlTypeContinue:
			return errContinue
		default:
			panic(fmt.Errorf("unexpected loop control type: %s", p0.Type))
		}
	})
}

func (i *interpreterPlanner) VisitActionStmtReturn(p0 *parse.ActionStmtReturn) any {
	var valFns []exprFunc
	var sqlStmt stmtFunc

	if len(p0.Values) > 0 {
		for _, v := range p0.Values {
			valFns = append(valFns, v.Accept(i).(exprFunc))
		}
	} else {
		sqlStmt = p0.SQL.Accept(i).(stmtFunc)
	}

	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		if len(valFns) > 0 {
			vals := make([]precompiles.Value, len(p0.Values))
			for j, valFn := range valFns {
				val, err := valFn(exec)
				if err != nil {
					return err
				}

				vals[j] = val
			}

			err := fn(makeRow(vals))
			if err != nil {
				return err
			}

			// we return a special error to indicate that the procedure is done.
			return errReturn
		}

		// otherwise, we execute the SQL statement.
		return sqlStmt(exec, func(row *row) error {
			row.fillUnnamed()
			return fn(row)
		})
	})
}

func (i *interpreterPlanner) VisitActionStmtReturnNext(p0 *parse.ActionStmtReturnNext) any {
	valFns := make([]exprFunc, len(p0.Values))
	for j, v := range p0.Values {
		valFns[j] = v.Accept(i).(exprFunc)
	}

	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		vals := make([]precompiles.Value, len(p0.Values))
		for j, valFn := range valFns {
			val, err := valFn(exec)
			if err != nil {
				return err
			}

			vals[j] = val
		}

		err := fn(makeRow(vals))
		if err != nil {
			return err
		}

		// we don't return an errReturn or mark done here because return next is not the last statement in a procedure.
		return nil
	})
}

// everything in this section is for expressions, which evaluate to exactly one value.

// handleTypeCast is a helper function that handles type casting.
func cast(t parse.Typecasted, s exprFunc) exprFunc {
	if t.GetTypeCast() == nil {
		return s
	}

	return exprFunc(func(exec *executionContext) (precompiles.Value, error) {
		val, err := s(exec)
		if err != nil {
			return nil, err
		}

		return val.Cast(t.GetTypeCast())
	})
}

// exprFunc is a function that returns a value.
type exprFunc func(exec *executionContext) (precompiles.Value, error)

func (i *interpreterPlanner) VisitExpressionLiteral(p0 *parse.ExpressionLiteral) any {
	return cast(p0, func(exec *executionContext) (precompiles.Value, error) {
		return precompiles.NewValue(p0.Value)
	})
}

func (i *interpreterPlanner) VisitExpressionFunctionCall(p0 *parse.ExpressionFunctionCall) any {
	args := make([]exprFunc, len(p0.Args))
	for j, arg := range p0.Args {
		args[j] = arg.Accept(i).(exprFunc)
	}

	return cast(p0, func(exec *executionContext) (precompiles.Value, error) {
		ns, err := exec.getNamespace(p0.Namespace)
		if err != nil {
			return nil, err
		}

		execute, ok := ns.availableFunctions[p0.Name]
		if !ok {
			return nil, fmt.Errorf(`unknown function "%s" in namespace "%s"`, p0.Name, p0.Namespace)
		}

		vals := make([]precompiles.Value, len(args))
		for j, arg := range args {
			val, err := arg(exec)
			if err != nil {
				return nil, err
			}

			vals[j] = val
		}

		var val precompiles.Value
		iters := 0
		err = execute.Func(exec, vals, func(received *row) error {
			iters++
			if len(received.Values) != 1 {
				return fmt.Errorf(`%w: expected function or action "%s" to return 1 value, but it returned %d`, engine.ErrReturnShape, p0.Name, len(received.Values))
			}

			val = received.Values[0]

			return nil
		})
		if err != nil {
			return nil, err
		}

		if iters == 0 {
			return nil, fmt.Errorf(`%w: expected function or action "%s" to return a single value, but it returned nothing`, engine.ErrReturnShape, p0.Name)
		} else if iters > 1 {
			return nil, fmt.Errorf(`%w: expected function or action "%s" to return a single value, but it returned %d values`, engine.ErrReturnShape, p0.Name, iters)
		}

		return val, nil
	})
}

func (i *interpreterPlanner) VisitExpressionVariable(p0 *parse.ExpressionVariable) any {
	return cast(p0, func(exec *executionContext) (precompiles.Value, error) {
		val, err := exec.getVariable(p0.Name)
		if err != nil {
			return nil, err
		}

		return val, nil
	})
}

func (i *interpreterPlanner) VisitExpressionArrayAccess(p0 *parse.ExpressionArrayAccess) any {
	arrFn := p0.Array.Accept(i).(exprFunc)
	var indexFn exprFunc
	var fromFn exprFunc
	var toFn exprFunc
	if p0.Index != nil {
		indexFn = p0.Index.Accept(i).(exprFunc)
	} else if p0.FromTo != nil {
		if p0.FromTo[0] != nil {
			fromFn = p0.FromTo[0].Accept(i).(exprFunc)
		}
		if p0.FromTo[1] != nil {
			toFn = p0.FromTo[1].Accept(i).(exprFunc)
		}
	} else {
		panic("unexpected array access statement")
	}

	return cast(p0, func(exec *executionContext) (precompiles.Value, error) {
		arrVal, err := arrFn(exec)
		if err != nil {
			return nil, err
		}

		if arrVal.Null() {
			if arrVal.Type().EqualsStrict(types.UnknownType) {
				return nil, fmt.Errorf("%w: cannot access array element of unknown type", engine.ErrInvalidNull)
			}
			arrType := arrVal.Type().Copy()
			arrType.IsArray = false
			return precompiles.MakeNull(arrType)
		}

		arr, ok := arrVal.(precompiles.ArrayValue)
		if !ok {
			return nil, fmt.Errorf("%w: expected array, got %T", engine.ErrType, arrVal)
		}

		checkArrIdx := func(v precompiles.Value) error {
			if !v.Type().EqualsStrict(types.IntType) {
				return fmt.Errorf("array index must be integer, got %s", v.Type())
			}

			return nil
		}

		if indexFn != nil {
			index, err := indexFn(exec)
			if err != nil {
				return nil, err
			}

			// if null, it should return a null value
			// of the scalar type of the array.
			// e.g. pg_typeof(text_array_val[nil]) = text
			if index.Null() {
				arrType := arr.Type().Copy()
				arrType.IsArray = false
				return precompiles.MakeNull(arrType)
			}

			if err := checkArrIdx(index); err != nil {
				return nil, err
			}

			return arr.Get(int32(index.RawValue().(int64)))
		}

		// 1-indexed
		start := int32(1)
		end := arr.Len()
		if fromFn != nil {
			fromVal, err := fromFn(exec)
			if err != nil {
				return nil, err
			}

			// if a null slice, it should return a null array.
			// e.g. pg_typeof(text_array_val[nil:nil]) = text[]
			if fromVal.Null() {
				return precompiles.MakeNull(arr.Type())
			}

			if err := checkArrIdx(fromVal); err != nil {
				return nil, err
			}

			start = int32(fromVal.RawValue().(int64))
		}
		if toFn != nil {
			toVal, err := toFn(exec)
			if err != nil {
				return nil, err
			}

			// if a null slice, it should return a null array.
			// e.g. pg_typeof(text_array_val[nil:nil]) = text[]
			if toVal.Null() {
				return precompiles.MakeNull(arr.Type())
			}

			if err := checkArrIdx(toVal); err != nil {
				return nil, err
			}

			end = int32(toVal.RawValue().(int64))
		}

		if start > end {
			// in Postgres, if the start is greater than the end, it returns an empty array.
			return precompiles.NewZeroValue(arr.Type())
		}
		// in Postgres, if the start is less than 1, it is treated as 1.
		if start < 1 {
			start = 1
		}
		// in Postgres, if the end is greater than the length of the array, it is treated as the length of the array.
		if end > arr.Len() {
			end = arr.Len()
		}

		zv, err := precompiles.NewZeroValue(arr.Type())
		if err != nil {
			return nil, err
		}

		arrZv, ok := zv.(precompiles.ArrayValue)
		if !ok {
			// should never happen
			return nil, fmt.Errorf("%w: expected array, got %T", engine.ErrType, zv)
		}

		j := int32(1)
		for i := start; i <= end; i++ {
			idx, err := arr.Get(i)
			if err != nil {
				return nil, err
			}
			err = arrZv.Set(j, idx)
			if err != nil {
				return nil, err
			}

			j++
		}

		return arrZv, nil
	})
}

func (i *interpreterPlanner) VisitExpressionMakeArray(p0 *parse.ExpressionMakeArray) any {
	valFns := make([]exprFunc, len(p0.Values))
	for j, v := range p0.Values {
		valFns[j] = v.Accept(i).(exprFunc)
	}

	return cast(p0, func(exec *executionContext) (precompiles.Value, error) {
		vals := make([]precompiles.ScalarValue, len(valFns))
		for j, valFn := range valFns {
			val, err := valFn(exec)
			if err != nil {
				return nil, err
			}

			scal, ok := val.(precompiles.ScalarValue)
			if !ok {
				return nil, fmt.Errorf("%w: expected scalar value, got %T", engine.ErrType, val)
			}

			vals[j] = scal
		}

		return precompiles.MakeArray(vals, p0.TypeCast)
	})
}

func (i *interpreterPlanner) VisitExpressionFieldAccess(p0 *parse.ExpressionFieldAccess) any {
	recordFn := p0.Record.Accept(i).(exprFunc)

	return cast(p0, func(exec *executionContext) (precompiles.Value, error) {
		objVal, err := recordFn(exec)
		if err != nil {
			return nil, err
		}

		obj, ok := objVal.(*precompiles.RecordValue)
		if !ok {
			return nil, fmt.Errorf("%w: expected object, got %T", engine.ErrType, objVal)
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
	cmpOps, negate := engine.ConvertComparisonOps(p0.Operator)

	left := p0.Left.Accept(i).(exprFunc)
	right := p0.Right.Accept(i).(exprFunc)

	retFn := makeComparisonFunc(left, right, cmpOps[0])

	for _, op := range cmpOps[1:] {
		retFn = makeLogicalFunc(retFn, makeComparisonFunc(left, right, op), false)
	}

	if negate {
		return makeUnaryFunc(retFn, engine.NOT)
	}

	return retFn
}

// makeComparisonFunc returns a function that compares two values.
func makeComparisonFunc(left, right exprFunc, cmpOps engine.ComparisonOp) exprFunc {
	return func(exec *executionContext) (precompiles.Value, error) {
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
	return func(exec *executionContext) (precompiles.Value, error) {
		leftVal, err := left(exec)
		if err != nil {
			return nil, err
		}

		rightVal, err := right(exec)
		if err != nil {
			return nil, err
		}

		if leftVal.Null() {
			return precompiles.MakeNull(types.BoolType)
		}

		if rightVal.Null() {
			return precompiles.MakeNull(types.BoolType)
		}

		if leftVal.Type() != types.BoolType || rightVal.Type() != types.BoolType {
			return nil, fmt.Errorf("%w: AND and OR operands must be of type bool, got %s and %s", engine.ErrType, leftVal.Type(), rightVal.Type())
		}

		if and {
			return precompiles.MakeBool(leftVal.RawValue().(bool) && rightVal.RawValue().(bool)), nil
		}

		return precompiles.MakeBool(leftVal.RawValue().(bool) || rightVal.RawValue().(bool)), nil
	}
}

func (i *interpreterPlanner) VisitExpressionArithmetic(p0 *parse.ExpressionArithmetic) any {
	op := engine.ConvertArithmeticOp(p0.Operator)

	leftFn := p0.Left.Accept(i).(exprFunc)
	rightFn := p0.Right.Accept(i).(exprFunc)
	return exprFunc(func(exec *executionContext) (precompiles.Value, error) {
		left, err := leftFn(exec)
		if err != nil {
			return nil, err
		}

		right, err := rightFn(exec)
		if err != nil {
			return nil, err
		}

		leftScalar, ok := left.(precompiles.ScalarValue)
		if !ok {
			return nil, fmt.Errorf("%w: expected scalar, got %T", engine.ErrType, left)
		}

		rightScalar, ok := right.(precompiles.ScalarValue)
		if !ok {
			return nil, fmt.Errorf("%w: expected scalar, got %T", engine.ErrType, right)
		}

		return leftScalar.Arithmetic(rightScalar, op)
	})
}

func (i *interpreterPlanner) VisitExpressionUnary(p0 *parse.ExpressionUnary) any {
	op := engine.ConvertUnaryOp(p0.Operator)
	val := p0.Expression.Accept(i).(exprFunc)
	return makeUnaryFunc(val, op)
}

// makeUnaryFunc returns a function that performs a unary operation.
func makeUnaryFunc(val exprFunc, op engine.UnaryOp) exprFunc {
	return exprFunc(func(exec *executionContext) (precompiles.Value, error) {
		v, err := val(exec)
		if err != nil {
			return nil, err
		}

		vScalar, ok := v.(precompiles.ScalarValue)
		if !ok {
			return nil, fmt.Errorf("%w: unary operations can only be performed on scalars, got %T", engine.ErrType, v)
		}

		return vScalar.Unary(op)
	})
}

func (i *interpreterPlanner) VisitExpressionIs(p0 *parse.ExpressionIs) any {
	left := p0.Left.Accept(i).(exprFunc)
	right := p0.Right.Accept(i).(exprFunc)

	op := engine.IS
	if p0.Distinct {
		op = engine.IS_DISTINCT_FROM
	}

	retFn := makeComparisonFunc(left, right, op)

	if p0.Not {
		return makeUnaryFunc(retFn, engine.NOT)
	}

	return retFn
}

/*
Role management
*/
func (i *interpreterPlanner) VisitGrantOrRevokeStatement(p0 *parse.GrantOrRevokeStatement) any {
	var varExprFn exprFunc
	if p0.ToVariable != nil {
		varExprFn = p0.ToVariable.Accept(i).(exprFunc)
	}
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		if err := exec.checkPrivilege(RolesPrivilege); err != nil {
			return err
		}

		// if we are granting ownership, then we need to check if the caller is an owner.
		if p0.GrantRole == ownerRole && !exec.engineCtx.OverrideAuthz {
			if !exec.isOwner() {
				return fmt.Errorf("%w: %s", engine.ErrDoesNotHavePrivilege, `only an owner can grant the "owner" role`)
			}
		}
		if p0.GrantRole == defaultRole {
			return fmt.Errorf("cannot grant or revoke the default role")
		}

		switch {
		case len(p0.Privileges) > 0 && p0.ToRole != "":
			fn := exec.interpreter.accessController.GrantPrivileges
			if !p0.IsGrant {
				fn = exec.interpreter.accessController.RevokePrivileges
			}

			convPrivs, err := validatePrivileges(p0.Privileges...)
			if err != nil {
				return err
			}

			if p0.Namespace != nil {
				err = canBeNamespaced(convPrivs...)
				if err != nil {
					return err
				}
			}

			return fn(exec.engineCtx.TxContext.Ctx, exec.db, p0.ToRole, convPrivs, p0.Namespace, p0.If)
		case p0.GrantRole != "" && p0.ToUser != "":
			fn := exec.interpreter.accessController.AssignRole
			if !p0.IsGrant {
				fn = exec.interpreter.accessController.UnassignRole
			}

			if p0.Namespace != nil {
				return fmt.Errorf("role assignment is global and cannot be namespaced")
			}

			return fn(exec.engineCtx.TxContext.Ctx, exec.db, p0.GrantRole, p0.ToUser, p0.If)
		case p0.GrantRole != "" && p0.ToVariable != nil:
			fn := exec.interpreter.accessController.AssignRole
			if !p0.IsGrant {
				fn = exec.interpreter.accessController.UnassignRole
			}

			if p0.Namespace != nil {
				return fmt.Errorf("role assignment is global and cannot be namespaced")
			}

			val, err := varExprFn(exec)
			if err != nil {
				return err
			}

			if val.Type() != types.TextType {
				return fmt.Errorf("%w: expected text, got %s", engine.ErrType, val.Type())
			}

			strVal, ok := val.RawValue().(string)
			if !ok {
				if val.Null() {
					return fmt.Errorf("cannot assign role to null user")
				}
				return fmt.Errorf("%w: expected text, got %T", engine.ErrType, val.RawValue())
			}

			return fn(exec.engineCtx.TxContext.Ctx, exec.db, p0.GrantRole, strVal, p0.If)
		default:
			// failure to hit these cases should have been caught by the parser, where better error
			// messages can be generated. This is a catch-all for any other invalid cases.
			return fmt.Errorf("invalid grant/revoke statement")
		}
	})
}

func (i *interpreterPlanner) VisitCreateRoleStatement(p0 *parse.CreateRoleStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		if err := exec.checkPrivilege(RolesPrivilege); err != nil {
			return err
		}

		if p0.IfNotExists {
			if exec.interpreter.accessController.RoleExists(p0.Role) {
				return nil
			}
		}

		return exec.interpreter.accessController.CreateRole(exec.engineCtx.TxContext.Ctx, exec.db, p0.Role)
	})
}

func (i *interpreterPlanner) VisitDropRoleStatement(p0 *parse.DropRoleStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		if err := exec.checkPrivilege(RolesPrivilege); err != nil {
			return err
		}

		if p0.IfExists {
			if !exec.interpreter.accessController.RoleExists(p0.Role) {
				return nil
			}
		}

		return exec.interpreter.accessController.DeleteRole(exec.engineCtx.TxContext.Ctx, exec.db, p0.Role)
	})
}

/*
	top-level adhoc
*/

// handleNamespaced is a helper function that handles statements namespaced with curly braces.
func handleNamespaced(exec *executionContext, stmt parse.Namespaceable) (reset func(), err error) {
	// if no special namespace is set, we can just return a no-op function
	if stmt.GetNamespacePrefix() == "" {
		return func() {}, nil
	}

	// otherwise, we need to set the current namespace
	oldNs := exec.scope.namespace

	// ensure the new namespace exists
	_, err = exec.getNamespace(stmt.GetNamespacePrefix())
	if err != nil {
		return nil, err
	}

	// set the new namespace
	exec.scope.namespace = stmt.GetNamespacePrefix()

	return func() {
		exec.scope.namespace = oldNs
	}, nil
}

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
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		reset, err := handleNamespaced(exec, p0)
		if err != nil {
			return err
		}
		defer reset()

		if mutatesState {
			if err := exec.checkNamespaceMutatbility(); err != nil {
				return err
			}
		}

		if err := exec.checkPrivilege(privilege); err != nil {
			return err
		}

		// if the query is trying to mutate state but the exec ctx cant then we should error
		if mutatesState && !exec.canMutateState {
			return fmt.Errorf("%w: SQL statement mutates state, but the execution context is read-only: %s", engine.ErrCannotMutateState, raw)
		}

		return exec.query(raw, fn)
	})
}

// here, we other top-level statements that are not covered by the other visitors.

// genAndExec generates and executes a DML statement.
// It should only be used for DDL statements, which do not bind or return values.
func genAndExec(exec *executionContext, stmt parse.TopLevelStatement) error {
	sql, _, err := pggenerate.GenerateSQL(stmt, exec.scope.namespace)
	if err != nil {
		return fmt.Errorf("%w: %w", engine.ErrPGGen, err)
	}

	return execute(exec.engineCtx.TxContext.Ctx, exec.db, sql)
}

func (i *interpreterPlanner) VisitAlterTableStatement(p0 *parse.AlterTableStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		reset, err := handleNamespaced(exec, p0)
		if err != nil {
			return err
		}
		defer reset()

		if err := exec.checkNamespaceMutatbility(); err != nil {
			return err
		}

		// ensure that the caller has the necessary privileges
		if err := exec.checkPrivilege(AlterPrivilege); err != nil {
			return err
		}

		// ensure the table exists
		_, err = exec.getTable("", p0.Table)
		if err != nil {
			return err
		}

		// instead of handling every case and how it should change the in-memory objects, we just
		// generate the SQL and execute it, and then completely refresh the in-memory objects for this schema.
		// This isn't the most efficient way to do it, but it's the easiest to implement, and since DDL isn't
		// really a hotpath, it's fine.
		err = genAndExec(exec, p0)
		if err != nil {
			return err
		}

		return exec.reloadTables()
	})
}

func (i *interpreterPlanner) VisitCreateTableStatement(p0 *parse.CreateTableStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		reset, err := handleNamespaced(exec, p0)
		if err != nil {
			return err
		}
		defer reset()

		if err := exec.checkNamespaceMutatbility(); err != nil {
			return err
		}

		// ensure that the caller has the necessary privileges
		if err := exec.checkPrivilege(CreatePrivilege); err != nil {
			return err
		}

		// ensure the table does not already exist
		_, err = exec.getTable("", p0.Name)
		if err == nil {
			// the table already exists
			if p0.IfNotExists {
				return nil
			}

			return fmt.Errorf(`table "%s" already exists`, p0.Name)
		} else if !errors.Is(err, engine.ErrUnknownTable) {
			return err
		}

		err = genAndExec(exec, p0)
		if err != nil {
			return err
		}

		return exec.reloadTables()
	})
}

func (i *interpreterPlanner) VisitDropTableStatement(p0 *parse.DropTableStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		reset, err := handleNamespaced(exec, p0)
		if err != nil {
			return err
		}
		defer reset()

		if err := exec.checkNamespaceMutatbility(); err != nil {
			return err
		}

		// ensure that the caller has the necessary privileges
		if err := exec.checkPrivilege(DropPrivilege); err != nil {
			return err
		}

		for _, table := range p0.Tables {
			// ensure the table exists
			_, err := exec.getTable("", table)
			if err != nil {
				if errors.Is(err, engine.ErrUnknownTable) {
					if p0.IfExists {
						continue
					}

					return fmt.Errorf(`table "%s" does not exist`, table)
				}

				return err
			}
		}

		if err := genAndExec(exec, p0); err != nil {
			return err
		}

		return exec.reloadTables()
	})
}

func (i *interpreterPlanner) VisitCreateIndexStatement(p0 *parse.CreateIndexStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		reset, err := handleNamespaced(exec, p0)
		if err != nil {
			return err
		}
		defer reset()

		if err := exec.checkNamespaceMutatbility(); err != nil {
			return err
		}

		// ensure that the caller has the necessary privileges
		if err := exec.checkPrivilege(CreatePrivilege); err != nil {
			return err
		}

		// ensure the table exists
		tbl, err := exec.getTable("", p0.On)
		if err != nil {
			return err
		}

		// ensure the columns exist
		tblCols := make(map[string]struct{}, len(tbl.Columns))
		for _, col := range tbl.Columns {
			tblCols[col.Name] = struct{}{}
		}

		for _, col := range p0.Columns {
			if _, found := tblCols[col]; !found {
				return fmt.Errorf(`column "%s" does not exist in table "%s"`, col, p0.On)
			}
		}

		if err := genAndExec(exec, p0); err != nil {
			return err
		}

		// we reload tables here because we track indexes in the table object
		return exec.reloadTables()
	})
}

func (i *interpreterPlanner) VisitDropIndexStatement(p0 *parse.DropIndexStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		reset, err := handleNamespaced(exec, p0)
		if err != nil {
			return err
		}
		defer reset()

		if err := exec.checkNamespaceMutatbility(); err != nil {
			return err
		}

		// ensure that the caller has the necessary privileges
		if err := exec.checkPrivilege(DropPrivilege); err != nil {
			return err
		}

		if err := genAndExec(exec, p0); err != nil {
			return err
		}

		// we reload tables here because we track indexes in the table object
		return exec.reloadTables()
	})
}

func (i *interpreterPlanner) VisitUseExtensionStatement(p0 *parse.UseExtensionStatement) any {
	configValues := make([]exprFunc, len(p0.Config))
	for j, config := range p0.Config {
		configValues[j] = config.Value.Accept(i).(exprFunc)
	}

	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		// ensure that the caller has the necessary privileges
		if err := exec.checkPrivilege(UsePrivilege); err != nil {
			return err
		}

		// see if the extension is already initialized
		if existing, exists := exec.interpreter.namespaces[p0.Alias]; exists {
			if existing.namespaceType == namespaceTypeExtension {
				if p0.IfNotExists {
					return nil
				} else {
					return fmt.Errorf(`extension initialized with alias "%s" already exists`, p0.Alias)
				}
			}

			return fmt.Errorf(`namespace "%s" already exists and is not an extension`, p0.Alias)
		}

		config := make(map[string]precompiles.Value, len(p0.Config))

		for j, configValue := range configValues {
			val, err := configValue(exec)
			if err != nil {
				return err
			}

			config[p0.Config[j].Key] = val
		}

		initializer, ok := precompiles.RegisteredPrecompiles()[strings.ToLower(p0.ExtName)]
		if !ok {
			return fmt.Errorf(`extension "%s" does not exist`, p0.ExtName)
		}

		extNamespace, inst, err := initializeExtension(exec.engineCtx.TxContext.Ctx, exec.interpreter.service, exec.db, initializer, p0.Alias, config)
		if err != nil {
			return err
		}

		if err := inst.OnStart(exec.engineCtx.TxContext.Ctx, exec.app()); err != nil {
			return err
		}

		err = registerExtensionInitialization(exec.engineCtx.TxContext.Ctx, exec.db, p0.Alias, p0.ExtName, config)
		if err != nil {
			return err
		}

		exec.interpreter.namespaces[p0.Alias] = extNamespace
		exec.interpreter.accessController.registerNamespace(p0.Alias)

		err = extNamespace.onDeploy(exec)
		if err != nil {
			delete(exec.interpreter.namespaces, p0.Alias)
			exec.interpreter.accessController.unregisterNamespace(p0.Alias)
			return err
		}

		return nil
	})
}

func (i *interpreterPlanner) VisitUnuseExtensionStatement(p0 *parse.UnuseExtensionStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		// ensure that the caller has the necessary privileges
		if err := exec.checkPrivilege(UsePrivilege); err != nil {
			return err
		}

		ns, exists := exec.interpreter.namespaces[p0.Alias]
		if !exists {
			if p0.IfExists {
				return nil
			}

			return fmt.Errorf(`extension initialized with alias "%s" does not exist`, p0.Alias)
		}

		if ns.namespaceType != namespaceTypeExtension {
			return fmt.Errorf(`namespace "%s" is not an extension`, p0.Alias)
		}

		err := ns.onUndeploy(exec)
		if err != nil {
			return err
		}

		err = unregisterExtensionInitialization(exec.engineCtx.TxContext.Ctx, exec.db, p0.Alias)
		if err != nil {
			return err
		}

		delete(exec.interpreter.namespaces, p0.Alias)
		exec.interpreter.accessController.unregisterNamespace(p0.Alias)

		return nil
	})
}

func (i *interpreterPlanner) VisitCreateActionStatement(p0 *parse.CreateActionStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		reset, err := handleNamespaced(exec, p0)
		if err != nil {
			return err
		}
		defer reset()

		if err := exec.checkNamespaceMutatbility(); err != nil {
			return err
		}

		if err := exec.checkPrivilege(CreatePrivilege); err != nil {
			return err
		}
		namespace := exec.interpreter.namespaces[exec.scope.namespace]

		// we check in the available functions map because there is a chance that the user is overwriting an existing function.
		if existingExec, exists := namespace.availableFunctions[p0.Name]; exists {
			if p0.IfNotExists {
				return nil
			} else if p0.OrReplace {
				// we delete the existing function.
				// If it is an action, we need to unstore it
				// If it is a built-in function, we just remove it from the map.
				if existingExec.Type == executableTypeAction {
					err = deleteAction(exec.engineCtx.TxContext.Ctx, exec.db, exec.scope.namespace, p0.Name)
					if err != nil {
						return err
					}
				}

				delete(namespace.availableFunctions, p0.Name)
			} else {
				return fmt.Errorf(`action/function "%s" already exists`, p0.Name)
			}
		}

		act := Action{}
		if err := act.FromAST(p0); err != nil {
			return err
		}

		err = storeAction(exec.engineCtx.TxContext.Ctx, exec.db, exec.scope.namespace, &act)
		if err != nil {
			return err
		}

		execute := makeActionToExecutable(exec.scope.namespace, &act)
		namespace.availableFunctions[p0.Name] = execute

		return nil
	})
}

func (i *interpreterPlanner) VisitDropActionStatement(p0 *parse.DropActionStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		reset, err := handleNamespaced(exec, p0)
		if err != nil {
			return err
		}
		defer reset()

		if err := exec.checkNamespaceMutatbility(); err != nil {
			return err
		}

		if err := exec.checkPrivilege(DropPrivilege); err != nil {
			return err
		}

		namespace := exec.interpreter.namespaces[exec.scope.namespace]

		// we check that the referenced executable is an action
		executable, exists := namespace.availableFunctions[p0.Name]
		if !exists {
			if p0.IfExists {
				return nil
			}

			return fmt.Errorf(`action "%s" does not exist`, p0.Name)
		}
		if executable.Type != executableTypeAction {
			return fmt.Errorf(`cannot drop executable "%s" of type %s`, p0.Name, executable.Type)
		}

		delete(namespace.availableFunctions, p0.Name)

		err = deleteAction(exec.engineCtx.TxContext.Ctx, exec.db, exec.scope.namespace, p0.Name)
		if err != nil {
			return err
		}

		// there are two cases we need to watch out for.
		// One is where the action originally overwrote a function. We should restore the function if it exists.
		// The second is if the action overwrote a method on an extension namespace, which we need to restore.
		// If it overwrote a method that overwrote a function, we should restore the method
		if funcDef, ok := engine.Functions[p0.Name]; ok {
			if scalarFunc, ok := funcDef.(*engine.ScalarFunctionDefinition); ok {
				namespace.availableFunctions[p0.Name] = funcDefToExecutable(p0.Name, scalarFunc)
			}
		}

		// if it is an extension, see if a corresponding method exists.
		// This will overwrite the function we just restored.
		if namespace.namespaceType == namespaceTypeExtension {
			method, ok := namespace.methods[p0.Name]
			if ok {
				namespace.availableFunctions[p0.Name] = method
			}
		}

		return nil
	})
}

func (i *interpreterPlanner) VisitCreateNamespaceStatement(p0 *parse.CreateNamespaceStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		if err := exec.checkPrivilege(CreatePrivilege); err != nil {
			return err
		}

		if _, exists := exec.interpreter.namespaces[p0.Namespace]; exists {
			if p0.IfNotExists {
				return nil
			}

			return fmt.Errorf(`%w: "%s"`, engine.ErrNamespaceExists, p0.Namespace)
		}

		if _, err := createNamespace(exec.engineCtx.TxContext.Ctx, exec.db, p0.Namespace, namespaceTypeUser); err != nil {
			return err
		}

		exec.interpreter.namespaces[p0.Namespace] = &namespace{
			availableFunctions: make(map[string]*executable),
			tables:             make(map[string]*engine.Table),
			onDeploy:           func(*executionContext) error { return nil },
			onUndeploy:         func(*executionContext) error { return nil },
		}
		exec.interpreter.accessController.registerNamespace(p0.Namespace)

		return nil
	})
}

func (i *interpreterPlanner) VisitDropNamespaceStatement(p0 *parse.DropNamespaceStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		if err := exec.checkPrivilege(DropPrivilege); err != nil {
			return err
		}

		ns, exists := exec.interpreter.namespaces[p0.Namespace]
		if !exists {
			if p0.IfExists {
				return nil
			}

			return fmt.Errorf(`%w: namespace "%s" does not exist`, engine.ErrNamespaceNotFound, p0.Namespace)
		}

		if p0.Namespace == DefaultNamespace || p0.Namespace == infoNamespace {
			return fmt.Errorf(`%w: "%s"`, engine.ErrCannotDropBuiltinNamespace, p0.Namespace)
		}
		if ns.namespaceType == namespaceTypeExtension {
			return fmt.Errorf(`%w: cannot drop extension namespace "%s" using DROP NAMESPACE. use UNUSE instead`, engine.ErrCannotMutateExtension, p0.Namespace)
		}

		if err := dropNamespace(exec.engineCtx.TxContext.Ctx, exec.db, p0.Namespace); err != nil {
			return err
		}

		delete(exec.interpreter.namespaces, p0.Namespace)
		exec.interpreter.accessController.unregisterNamespace(p0.Namespace)

		return nil
	})
}

// below are the alter table statements

func (i *interpreterPlanner) VisitAddColumn(p0 *parse.AddColumn) any {
	panic("intepreter planner should not be called for alter table statements")
}

func (i *interpreterPlanner) VisitDropColumn(p0 *parse.DropColumn) any {
	panic("intepreter planner should not be called for alter table statements")
}

func (i *interpreterPlanner) VisitRenameColumn(p0 *parse.RenameColumn) any {
	panic("intepreter planner should not be called for alter table statements")
}

func (i *interpreterPlanner) VisitRenameTable(p0 *parse.RenameTable) any {
	panic("intepreter planner should not be called for alter table statements")
}

func (i *interpreterPlanner) VisitAddTableConstraint(p0 *parse.AddTableConstraint) any {
	panic("intepreter planner should not be called for alter table statements")
}

func (i *interpreterPlanner) VisitDropTableConstraint(p0 *parse.DropTableConstraint) any {
	panic("intepreter planner should not be called for alter table statements")
}

func (i *interpreterPlanner) VisitColumn(p0 *parse.Column) any {
	panic("intepreter planner should not be called for alter table statements")
}

func (i *interpreterPlanner) VisitAlterColumnSet(p0 *parse.AlterColumnSet) any {
	panic("intepreter planner should not be called for alter table statements")
}

func (i *interpreterPlanner) VisitAlterColumnDrop(p0 *parse.AlterColumnDrop) any {
	panic("intepreter planner should not be called for alter table statements")
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

func (i *interpreterPlanner) VisitIfThen(p0 *parse.IfThen) any {
	// we handle this directly in VisitActionStmtIf
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

func (i *interpreterPlanner) VisitPrimaryKeyInlineConstraint(p0 *parse.PrimaryKeyInlineConstraint) any {
	panic("interpreter planner should never be called for table constraints")
}

func (i *interpreterPlanner) VisitPrimaryKeyOutOfLineConstraint(p0 *parse.PrimaryKeyOutOfLineConstraint) any {
	panic("interpreter planner should never be called for table constraints")
}

func (i *interpreterPlanner) VisitUniqueInlineConstraint(p0 *parse.UniqueInlineConstraint) any {
	panic("interpreter planner should never be called for table constraints")
}

func (i *interpreterPlanner) VisitUniqueOutOfLineConstraint(p0 *parse.UniqueOutOfLineConstraint) any {
	panic("interpreter planner should never be called for table constraints")
}

func (i *interpreterPlanner) VisitDefaultConstraint(p0 *parse.DefaultConstraint) any {
	panic("interpreter planner should never be called for table constraints")
}

func (i *interpreterPlanner) VisitNotNullConstraint(p0 *parse.NotNullConstraint) any {
	panic("interpreter planner should never be called for table constraints")
}

func (i *interpreterPlanner) VisitCheckConstraint(p0 *parse.CheckConstraint) any {
	panic("interpreter planner should never be called for table constraints")
}

func (i *interpreterPlanner) VisitForeignKeyReferences(p0 *parse.ForeignKeyReferences) any {
	panic("interpreter planner should never be called for table constraints")
}

func (i *interpreterPlanner) VisitForeignKeyOutOfLineConstraint(p0 *parse.ForeignKeyOutOfLineConstraint) any {
	panic("interpreter planner should never be called for table constraints")
}
