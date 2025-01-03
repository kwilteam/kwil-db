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
		Func: func(exec *executionContext, args []Value, fn resultFunc) error {
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

			exec2 := exec.child(namespace)

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

func makeRow(v []Value) *row {
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
	// Values is a list of values.
	Values []Value
}

func (r *row) record() (*RecordValue, error) {
	rec := newRecordValue()
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
		return exec.allocateVariable(p0.Variable.Name, newNull(p0.Type))
	})
}

func (i *interpreterPlanner) VisitActionStmtAssignment(p0 *parse.ActionStmtAssign) any {
	valFn := p0.Value.Accept(i).(exprFunc)

	var arrFn exprFunc
	var indexFn exprFunc
	if a, ok := p0.Variable.(*parse.ExpressionArrayAccess); ok {
		arrFn = a.Array.Accept(i).(exprFunc)
		indexFn = a.Index.Accept(i).(exprFunc)
	}
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
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

			err = arr.Set(int32(index.RawValue().(int64)), scalarVal)
			if err != nil {
				return err
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

		vals := make([]Value, len(args))
		for j, valFn := range args {
			val, err := valFn(exec)
			if err != nil {
				return err
			}

			vals[j] = val
		}

		iter := 0
		err = funcDef.Func(exec, vals, func(row *row) error {
			iter++

			// re-verify the returns, since the above checks only for what the function signature
			// says, but this checks what the function actually returns.
			if len(receivers) > len(row.Values) {
				return fmt.Errorf(`expected action "%s" to return at least %d values, but it returned %d`, funcDef.Name, len(receivers), len(row.Values))
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
func executeBlock(exec *executionContext, fn resultFunc,
	stmtFuncs []stmtFunc) error {
	exec.scope.subScope()
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
		err := loopFn(exec, func(term Value) error {
			exec.scope.subScope()
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
type loopTermFunc func(exec *executionContext, fn func(Value) error) (err error)

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
			err = handleLoopTermErr(fn(newInt(i)))
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
		return exec.query(raw, func(r *row) error {
			rec, err := r.record()
			if err != nil {
				return err
			}

			return handleLoopTermErr(fn(rec))
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

		for i := range arr.Len() {
			scalar, err := arr.Index(i + 1) // all arrays are 1-indexed
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

func (i *interpreterPlanner) VisitLoopTermFunctionCall(p0 *parse.LoopTermFunctionCall) any {
	// we cannot simply use the expression function call because we enforce that expression function
	// calls do not return tables. Here, we can return tables.

	args := make([]exprFunc, len(p0.Call.Args))
	for j, arg := range p0.Call.Args {
		args[j] = arg.Accept(i).(exprFunc)
	}
	return loopTermFunc(func(exec *executionContext, fn func(Value) error) error {
		// the function call here can either return a table or a single array value.
		ns, err := exec.getNamespace(p0.Call.Namespace)
		if err != nil {
			return err
		}

		funcDef, ok := ns.availableFunctions[p0.Call.Name]
		if !ok {
			return fmt.Errorf(`unknown function "%s" in namespace "%s"`, p0.Call.Name, p0.Call.Namespace)
		}

		vals := make([]Value, len(args))
		for j, valFn := range args {
			val, err := valFn(exec)
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

			if boolVal, ok := cond.(*BoolValue); ok {
				if boolVal.Null() {
					continue
				}
				if !boolVal.Bool.Bool {
					continue
				}
			} else {
				return fmt.Errorf("expected bool, got %s", cond.Type())
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
			vals := make([]Value, len(p0.Values))
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
		vals := make([]Value, len(p0.Values))
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

	return exprFunc(func(exec *executionContext) (Value, error) {
		val, err := s(exec)
		if err != nil {
			return nil, err
		}

		return val.Cast(t.GetTypeCast())
	})
}

// exprFunc is a function that returns a value.
type exprFunc func(exec *executionContext) (Value, error)

func (i *interpreterPlanner) VisitExpressionLiteral(p0 *parse.ExpressionLiteral) any {
	return cast(p0, func(exec *executionContext) (Value, error) {
		return NewValue(p0.Value)
	})
}

func (i *interpreterPlanner) VisitExpressionFunctionCall(p0 *parse.ExpressionFunctionCall) any {
	args := make([]exprFunc, len(p0.Args))
	for j, arg := range p0.Args {
		args[j] = arg.Accept(i).(exprFunc)
	}

	return cast(p0, func(exec *executionContext) (Value, error) {
		ns, err := exec.getNamespace(p0.Namespace)
		if err != nil {
			return nil, err
		}

		execute, ok := ns.availableFunctions[p0.Name]
		if !ok {
			return nil, fmt.Errorf(`unknown function "%s" in namespace "%s"`, p0.Name, p0.Namespace)
		}

		vals := make([]Value, len(args))
		for j, arg := range args {
			val, err := arg(exec)
			if err != nil {
				return nil, err
			}

			vals[j] = val
		}

		var val Value
		iters := 0
		err = execute.Func(exec, vals, func(received *row) error {
			iters++
			if len(received.Values) != 1 {
				return fmt.Errorf(`expected function "%s" to return 1 value, but it returned %d`, p0.Name, len(received.Values))
			}

			val = received.Values[0]

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
	return cast(p0, func(exec *executionContext) (Value, error) {
		val, found := exec.getVariable(p0.Name)
		if !found {
			return nil, fmt.Errorf("%w: %s", ErrVariableNotFound, p0.Name)
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

	return cast(p0, func(exec *executionContext) (Value, error) {
		arrVal, err := arrFn(exec)
		if err != nil {
			return nil, err
		}

		arr, ok := arrVal.(ArrayValue)
		if !ok {
			return nil, fmt.Errorf("expected array, got %T", arrVal)
		}

		checkArrIdx := func(v Value) error {
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

			if err := checkArrIdx(index); err != nil {
				return nil, err
			}

			return arr.Index(int32(index.RawValue().(int64)))
		}

		// 1-indexed
		start := int32(1)
		end := arr.Len()
		if fromFn != nil {
			fromVal, err := fromFn(exec)
			if err != nil {
				return nil, err
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

			if err := checkArrIdx(toVal); err != nil {
				return nil, err
			}

			end = int32(toVal.RawValue().(int64))
		}

		if start > end {
			// in Postgres, if the start is greater than the end, it returns an empty array.
			return NewZeroValue(arr.Type())
		}
		// in Postgres, if the start is less than 1, it is treated as 1.
		if start < 1 {
			start = 1
		}
		// in Postgres, if the end is greater than the length of the array, it is treated as the length of the array.
		if end > arr.Len() {
			end = arr.Len()
		}

		zv, err := NewZeroValue(arr.Type())
		if err != nil {
			return nil, err
		}

		arrZv, ok := zv.(ArrayValue)
		if !ok {
			// should never happen
			return nil, fmt.Errorf("expected array, got %T", zv)
		}

		j := int32(1)
		for i := start; i <= end; i++ {
			idx, err := arr.Index(i)
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

	return cast(p0, func(exec *executionContext) (Value, error) {
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

	return cast(p0, func(exec *executionContext) (Value, error) {
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

		if leftVal.Null() {
			return newNull(types.BoolType), nil
		}

		if rightVal.Null() {
			return newNull(types.BoolType), nil
		}

		if leftVal.Type() != types.BoolType || rightVal.Type() != types.BoolType {
			return nil, fmt.Errorf("expected bools, got %s and %s", leftVal.Type(), rightVal.Type())
		}

		if and {
			return newBool(leftVal.RawValue().(bool) && rightVal.RawValue().(bool)), nil
		}

		return newBool(leftVal.RawValue().(bool) || rightVal.RawValue().(bool)), nil
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
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		if !exec.interpreter.accessController.HasPrivilege(exec.txCtx.Caller, nil, RolesPrivilege) {
			return fmt.Errorf("%w: %s", ErrDoesNotHavePriv, RolesPrivilege)
		}

		switch {
		case len(p0.Privileges) > 0 && p0.ToRole != "":
			fn := exec.interpreter.accessController.GrantPrivileges
			if !p0.IsGrant {
				fn = exec.interpreter.accessController.RevokePrivileges
			}
			return fn(exec.txCtx.Ctx, exec.db, p0.ToRole, p0.Privileges, p0.Namespace)
		case p0.GrantRole != "" && p0.ToUser != "":
			fn := exec.interpreter.accessController.AssignRole
			if !p0.IsGrant {
				fn = exec.interpreter.accessController.UnassignRole
			}
			return fn(exec.txCtx.Ctx, exec.db, p0.GrantRole, p0.ToUser)
		default:
			// failure to hit these cases should have been caught by the parser, where better error
			// messages can be generated. This is a catch-all for any other invalid cases.
			return fmt.Errorf("invalid grant/revoke statement")
		}
	})
}

func (i *interpreterPlanner) VisitCreateRoleStatement(p0 *parse.CreateRoleStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		if !exec.interpreter.accessController.HasPrivilege(exec.txCtx.Caller, nil, RolesPrivilege) {
			return fmt.Errorf("%w: %s", ErrDoesNotHavePriv, RolesPrivilege)
		}

		if p0.IfNotExists {
			if exec.interpreter.accessController.RoleExists(p0.Role) {
				return nil
			}
		}

		return exec.interpreter.accessController.CreateRole(exec.txCtx.Ctx, exec.db, p0.Role)
	})
}

func (i *interpreterPlanner) VisitDropRoleStatement(p0 *parse.DropRoleStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		if !exec.interpreter.accessController.HasPrivilege(exec.txCtx.Caller, nil, RolesPrivilege) {
			return fmt.Errorf("%w: %s", ErrDoesNotHavePriv, RolesPrivilege)
		}

		if p0.IfExists {
			if !exec.interpreter.accessController.RoleExists(p0.Role) {
				return nil
			}
		}

		return exec.interpreter.accessController.DeleteRole(exec.txCtx.Ctx, exec.db, p0.Role)
	})
}

func (i *interpreterPlanner) VisitTransferOwnershipStatement(p0 *parse.TransferOwnershipStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		if !exec.interpreter.accessController.IsOwner(exec.txCtx.Caller) {
			return fmt.Errorf("%w: %s", ErrDoesNotHavePriv, "caller must be owner")
		}

		return exec.interpreter.accessController.SetOwnership(exec.txCtx.Ctx, exec.db, p0.To)
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

		if err := exec.checkPrivilege(privilege); err != nil {
			return err
		}

		// if the query is trying to mutate state but the exec ctx cant then we should error
		if mutatesState && !exec.canMutateState {
			return fmt.Errorf("%w: SQL statement mutates state, but the execution context is read-only: %s", ErrStatementMutatesState, raw)
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
		return err
	}

	return execute(exec.txCtx.Ctx, exec.db, sql)
}

func (i *interpreterPlanner) VisitAlterTableStatement(p0 *parse.AlterTableStatement) any {
	return stmtFunc(func(exec *executionContext, fn resultFunc) error {
		reset, err := handleNamespaced(exec, p0)
		if err != nil {
			return err
		}
		defer reset()

		// ensure that the caller has the necessary privileges
		if err := exec.checkPrivilege(AlterPrivilege); err != nil {
			return err
		}

		// ensure the table exists
		_, found := exec.getTable("", p0.Table)
		if !found {
			return fmt.Errorf("table %s does not exist", p0.Table)
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

		// ensure that the caller has the necessary privileges
		if err := exec.checkPrivilege(CreatePrivilege); err != nil {
			return err
		}

		// ensure the table does not already exist
		_, found := exec.getTable("", p0.Name)
		if found {
			if p0.IfNotExists {
				return nil
			}

			return fmt.Errorf(`table "%s" already exists`, p0.Name)
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

		// ensure that the caller has the necessary privileges
		if err := exec.checkPrivilege(DropPrivilege); err != nil {
			return err
		}

		for _, table := range p0.Tables {
			// ensure the table exists
			_, found := exec.getTable("", table)
			if !found {
				if p0.IfExists {
					continue
				}

				return fmt.Errorf(`table "%s" does not exist`, table)
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

		// ensure that the caller has the necessary privileges
		if err := exec.checkPrivilege(CreatePrivilege); err != nil {
			return err
		}

		// ensure the table exists
		tbl, found := exec.getTable("", p0.On)
		if !found {
			return fmt.Errorf(`table "%s" does not exist`, p0.On)
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

		config := make(map[string]Value, len(p0.Config))

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

		extNamespace, err := initializeExtension(exec.txCtx.Ctx, exec.interpreter.service, exec.db, initializer, config)
		if err != nil {
			return err
		}

		err = extNamespace.onDeploy(exec)
		if err != nil {
			return err
		}

		err = registerExtensionInitialization(exec.txCtx.Ctx, exec.db, p0.Alias, p0.ExtName, config)
		if err != nil {
			return err
		}

		exec.interpreter.namespaces[p0.Alias] = extNamespace
		exec.interpreter.accessController.registerNamespace(p0.Alias)

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

		err = unregisterExtensionInitialization(exec.txCtx.Ctx, exec.db, p0.Alias)
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

		if err := exec.checkPrivilege(CreatePrivilege); err != nil {
			return err
		}

		namespace := exec.interpreter.namespaces[exec.scope.namespace]

		// we check in the available functions map because there is a chance that the user is overwriting an existing function.
		if _, exists := namespace.availableFunctions[p0.Name]; exists {
			if p0.IfNotExists {
				return nil
			} else if p0.OrReplace {
				delete(namespace.availableFunctions, p0.Name)
			} else {
				return fmt.Errorf(`action/function "%s" already exists`, p0.Name)
			}
		}

		act := Action{}
		if err := act.FromAST(p0); err != nil {
			return err
		}

		err = storeAction(exec.txCtx.Ctx, exec.db, exec.scope.namespace, &act)
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
			return fmt.Errorf(`%w: cannot drop executable "%s" of type %s`, ErrCannotDrop, p0.Name, executable.Type)
		}

		delete(namespace.availableFunctions, p0.Name)

		// there is a case where an action overwrites a function. We should restore the function if it exists.
		if funcDef, ok := parse.Functions[p0.Name]; ok {
			if scalarFunc, ok := funcDef.(*parse.ScalarFunctionDefinition); ok {
				namespace.availableFunctions[p0.Name] = funcDefToExecutable(p0.Name, scalarFunc)
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

			return fmt.Errorf(`%w: "%s"`, ErrNamespaceExists, p0.Namespace)
		}

		if _, err := createNamespace(exec.txCtx.Ctx, exec.db, p0.Namespace, namespaceTypeUser); err != nil {
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

			return fmt.Errorf(`%w: namespace "%s" does not exist`, ErrNamespaceNotFound, p0.Namespace)
		}

		if ns.namespaceType == namespaceTypeSystem {
			return fmt.Errorf(`cannot drop built-in namespace "%s"`, p0.Namespace)
		}
		if ns.namespaceType == namespaceTypeExtension {
			return fmt.Errorf(`cannot drop extension namespace using DROP "%s"`, p0.Namespace)
		}

		if err := dropNamespace(exec.txCtx.Ctx, exec.db, p0.Namespace); err != nil {
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
