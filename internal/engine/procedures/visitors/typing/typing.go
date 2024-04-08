// package typing checks the type safety of a procedure.
// It will return data types for expressions, map[string]DataType for loop terms,
// and nothing for statements.
package typing

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine"
	"github.com/kwilteam/kwil-db/internal/engine/procedures/parser"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/typing"
	"github.com/kwilteam/kwil-db/internal/parse/sql/tree"
)

func EnsureTyping(stmts []parser.Statement, procedure *types.Procedure, schema *types.Schema, cleanedInputs []*types.NamedType) (err error) {
	declarations := make(map[string]*types.DataType)
	for _, param := range cleanedInputs {
		declarations[param.Name] = param.Type
	}

	t := &typingVisitor{
		currentSchema:         schema,
		declarations:          declarations,
		currentProcedure:      procedure,
		anonymousDeclarations: make(map[string]map[string]*types.DataType),
	}

	defer func() {
		if r := recover(); r != nil {
			if t.err != nil {
				err = t.err
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	for _, stmt := range stmts {
		stmt.Accept(t)
	}

	return nil
}

type typingVisitor struct {
	// currentSchema holds the schema of the current procedure
	currentSchema *types.Schema
	// declarations holds information about all variable declarations in the procedure
	declarations map[string]*types.DataType
	// currentProcedure holds the information about the current procedure
	currentProcedure *types.Procedure
	// anonymousDeclarations holds information about all anonymous variable declarations in the procedure
	// currently, the only anonymous declarations are loop targets (e.g. FOR i IN SELECT * FROM users).
	// anonymousDeclarations are essentially anonymous compound types, so the map maps the name
	// to the fields to their types.
	anonymousDeclarations map[string]map[string]*types.DataType

	// holds the last error that occurred
	err error
}

var _ parser.Visitor = &typingVisitor{}

func (t *typingVisitor) VisitExpressionArithmetic(p0 *parser.ExpressionArithmetic) interface {
} {
	t.asserIntType(p0.Left)
	t.asserIntType(p0.Right)
	return types.IntType
}

func (t *typingVisitor) VisitExpressionArrayAccess(p0 *parser.ExpressionArrayAccess) interface {
} {
	t.asserIntType(p0.Index)

	r, ok := p0.Target.Accept(t).(*types.DataType)
	if !ok {
		panic("expected custom data type")
	}
	if !r.IsArray {
		panic("expected array")
	}

	return &types.DataType{
		Name: r.Name,
	}
}

func (t *typingVisitor) VisitExpressionBlobLiteral(p0 *parser.ExpressionBlobLiteral) interface {
} {
	return types.BlobType
}

func (t *typingVisitor) VisitExpressionBooleanLiteral(p0 *parser.ExpressionBooleanLiteral) interface {
} {
	return types.BoolType
}

func (t *typingVisitor) VisitExpressionCall(p0 *parser.ExpressionCall) interface {
} {
	funcDef, ok := engine.Functions[p0.Name]
	if ok {
		argsTypes := make([]*types.DataType, len(p0.Arguments))
		for i, arg := range p0.Arguments {
			argsTypes[i] = arg.Accept(t).(*types.DataType)
		}

		returnType, err := funcDef.Args(argsTypes)
		if err != nil {
			panic(err)
		}

		return returnType
	}

	var proc *types.Procedure
	found := false
	for _, p := range t.currentSchema.Procedures {
		if p.Name == p0.Name {
			proc = p
			found = true
			break
		}
	}
	if !found {
		panic("procedure not found")
	}

	return proc.Returns
}

func (t *typingVisitor) VisitExpressionComparison(p0 *parser.ExpressionComparison) interface {
} {
	t.asserIntType(p0.Left)
	t.asserIntType(p0.Right)
	return types.BoolType
}

func (t *typingVisitor) VisitExpressionFieldAccess(p0 *parser.ExpressionFieldAccess) interface {
} {
	anonType, ok := p0.Target.Accept(t).(map[string]*types.DataType)
	if !ok {
		panic("expected anonymous type")
	}

	dt, ok := anonType[p0.Field]
	if !ok {
		panic(fmt.Sprintf("field %s not found", p0.Field))
	}

	return dt
}

func (t *typingVisitor) VisitExpressionIntLiteral(p0 *parser.ExpressionIntLiteral) interface {
} {
	return types.IntType
}

func (t *typingVisitor) VisitExpressionMakeArray(p0 *parser.ExpressionMakeArray) interface {
} {
	var arrayType *types.DataType
	for _, e := range p0.Values {
		dataType, ok := e.Accept(t).(*types.DataType)
		if !ok {
			panic(fmt.Sprintf("expected data type in array, got %T", e.Accept(t)))
		}

		if dataType.IsArray {
			panic("array cannot contain arrays")
		}

		if arrayType == nil {
			arrayType = dataType
			continue
		}

		if !arrayType.Equals(dataType) {
			panic(fmt.Sprintf("array type mismatch: %s != %s", arrayType, dataType))
		}
	}

	return &types.DataType{
		Name:    arrayType.Name,
		IsArray: true,
	}
}

func (t *typingVisitor) VisitExpressionNullLiteral(p0 *parser.ExpressionNullLiteral) interface {
} {
	return &types.NullType
}

func (t *typingVisitor) VisitExpressionParenthesized(p0 *parser.ExpressionParenthesized) interface {
} {
	return p0.Expression.Accept(t)
}

func (t *typingVisitor) VisitExpressionTextLiteral(p0 *parser.ExpressionTextLiteral) interface {
} {
	return types.TextType
}

func (t *typingVisitor) VisitExpressionVariable(p0 *parser.ExpressionVariable) interface {
} {
	dt, ok := t.declarations[p0.Name]
	if !ok {
		anonType, ok := t.anonymousDeclarations[p0.Name]
		if !ok {
			panic(fmt.Sprintf("variable %s not declared", p0.Name))
		}

		return anonType
	}

	return dt
}

func (t *typingVisitor) VisitLoopTargetCall(p0 *parser.LoopTargetCall) interface {
} {
	r := p0.Call.Accept(t)
	tbl, ok := r.(*types.ProcedureReturn)
	if !ok {
		panic("procedure loop target must return a table")
	}
	if tbl.Table == nil {
		panic("procedure loop target must return a table")
	}

	vals := make(map[string]*types.DataType)
	for _, col := range tbl.Table {
		vals[col.Name] = col.Type
	}

	return vals
}

func (t *typingVisitor) VisitLoopTargetRange(p0 *parser.LoopTargetRange) interface {
} {
	t.asserIntType(p0.Start)
	t.asserIntType(p0.End)
	return types.IntType
}

func (t *typingVisitor) VisitLoopTargetSQL(p0 *parser.LoopTargetSQL) interface {
} {
	rel := t.analyzeSQL(p0.Statement)

	r := make(map[string]*types.DataType)
	err := rel.Loop(func(s string, a *engine.Attribute) error {
		r[s] = a.Type
		return nil
	})
	if err != nil {
		panic(err)
	}
	return r
}

func (t *typingVisitor) VisitLoopTargetVariable(p0 *parser.LoopTargetVariable) interface {
} {
	// must check that the variable is an array
	r, ok := p0.Variable.Accept(t).(*types.DataType)
	if !ok {
		panic("expected data type")
	}

	if !r.IsArray {
		panic("expected array")
	}

	return &types.DataType{
		Name: r.Name,
	}
}

func (t *typingVisitor) VisitStatementProcedureCall(p0 *parser.StatementProcedureCall) interface {
} {
	returns := p0.Call.Accept(t).(*types.ProcedureReturn)
	if returns == nil {
		if len(p0.Variables) > 0 {
			panic("procedure has no returns")
		}

		return nil
	}

	if returns.Table != nil {
		panic("cannot assign return table to variable")
	}

	if len(returns.Types) != len(p0.Variables) {
		panic("expected return values")
	}

	for i, v := range p0.Variables {
		varType, ok := t.declarations[v]
		if !ok {
			panic(fmt.Sprintf("variable %s not declared", v))
		}

		if !varType.Equals(returns.Types[i]) {
			panic(fmt.Sprintf("variable %s has wrong type", v))
		}
	}

	p0.Call.Accept(t) // we accept to ensure it visits, but we do not care about the return value
	return nil
}

func (t *typingVisitor) VisitStatementBreak(p0 *parser.StatementBreak) interface {
} {
	return nil
}

func (t *typingVisitor) VisitStatementForLoop(p0 *parser.StatementForLoop) interface {
} {
	switch target := p0.Target.(type) {
	case *parser.LoopTargetVariable:
		r := target.Accept(t).(*types.DataType)

		// we will not declare these as anonymous, since this is simply
		// a field in an array of a known type
		_, ok := t.declarations[p0.Variable]
		if ok {
			panic("variable already declared")
		}

		t.declarations[p0.Variable] = r
	case *parser.LoopTargetCall:
		// this can be either a table or a custom data type
		r := target.Accept(t).(map[string]*types.DataType)
		_, ok := t.anonymousDeclarations[p0.Variable]
		if ok {
			panic("variable already declared")
		}

		t.anonymousDeclarations[p0.Variable] = r
	case *parser.LoopTargetRange:
		r := target.Accept(t).(*types.DataType) // int

		// we will not declare these as anonymous, since it is a simple int
		_, ok := t.declarations[p0.Variable]
		if ok {
			panic("variable already declared")
		}

		t.declarations[p0.Variable] = r
	case *parser.LoopTargetSQL:
		r := target.Accept(t).(map[string]*types.DataType)

		_, ok := t.anonymousDeclarations[p0.Variable]
		if ok {
			panic("variable already declared")
		}

		t.anonymousDeclarations[p0.Variable] = r
	default:
		panic("unknown loop target")
	}

	for _, stmt := range p0.Body {
		stmt.Accept(t)
	}

	return nil
}

// analyzeSQL analyzes the given SQL statement and returns the resulting relation.
func (t *typingVisitor) analyzeSQL(stmt tree.AstNode) *engine.Relation {
	m, err := typing.AnalyzeTypes(stmt, t.currentSchema.Tables, t.declarations)
	if err != nil {
		panic(err)
	}

	return m
}

func (t *typingVisitor) VisitStatementIf(p0 *parser.StatementIf) interface {
} {
	for _, it := range p0.IfThens {
		t.assertBoolType(it.If)

		for _, stmt := range it.Then {
			stmt.Accept(t)
		}
	}

	for _, stmt := range p0.Else {
		stmt.Accept(t)
	}

	return nil
}

func (t *typingVisitor) VisitStatementReturn(p0 *parser.StatementReturn) interface {
} {
	if t.currentProcedure.Returns == nil {
		panic("procedure does not return anything")
	}

	switch {
	case t.currentProcedure.Returns.Table != nil:
		r := t.analyzeSQL(p0.SQL)

		for _, col := range t.currentProcedure.Returns.Table {
			attr, ok := r.Attribute(col.Name)
			if !ok {
				panic(fmt.Sprintf(`column "%s" not found in return table`, col.Name))
			}

			if !col.Type.Equals(attr.Type) {
				panic(fmt.Sprintf(`column "%s" has wrong type`, col.Name))
			}
		}
	case t.currentProcedure.Returns.Types != nil:
		if p0.Values == nil {
			panic("expected return value")
		}

		if len(p0.Values) != len(t.currentProcedure.Returns.Types) {
			panic(fmt.Sprintf("expected %d return values, got %d", len(t.currentProcedure.Returns.Types), len(p0.Values)))
		}

		for i, v := range p0.Values {
			r, ok := v.Accept(t).(*types.DataType)
			if !ok {
				panic("expected custom data type")
			}

			if !t.currentProcedure.Returns.Types[i].Equals(r) {
				panic(fmt.Sprintf("return type does not match procedure return type: %s != %s", t.currentProcedure.Returns.Types[i], r))
			}
		}
	}

	return nil
}

func (t *typingVisitor) VisitStatementReturnNext(p0 *parser.StatementReturnNext) interface {
} {
	// we can only call return next on records,
	// which should be declared as anonymous
	r, ok := t.anonymousDeclarations[p0.Variable]
	if !ok {
		panic("variable not declared")
	}

	if t.currentProcedure.Returns == nil {
		panic("procedure does not return anything")
	}

	if t.currentProcedure.Returns.Table == nil {
		panic("procedure does not return a table")
	}

	for _, col := range t.currentProcedure.Returns.Table {
		dataType, ok := r[col.Name]
		if !ok {
			panic(fmt.Sprintf(`column "%s" not found in return table`, col.Name))
		}

		if !col.Type.Equals(dataType) {
			panic(fmt.Sprintf(`column "%s" has wrong type`, col.Name))
		}
	}

	return nil
}

func (t *typingVisitor) VisitStatementSQL(p0 *parser.StatementSQL) interface {
} {
	// we do not care about the return value
	return nil
}

func (t *typingVisitor) VisitStatementVariableAssignment(p0 *parser.StatementVariableAssignment) interface {
} {
	typ, ok := t.declarations[p0.Name]
	if !ok {
		panic(fmt.Sprintf("variable %s not declared", p0.Name))
	}

	r, ok := p0.Value.Accept(t).(*types.DataType)
	if !ok {
		panic("expected custom data type")
	}

	if !typ.Equals(r) {
		panic(fmt.Sprintf("assignment type does not match variable type: %s != %s", typ, r))
	}

	return nil
}

func (t *typingVisitor) VisitStatementVariableAssignmentWithDeclaration(p0 *parser.StatementVariableAssignmentWithDeclaration) interface {
} {
	retType, ok := p0.Value.Accept(t).(*types.DataType)
	if !ok {
		panic("expected custom data type")
	}

	if !p0.Type.Equals(retType) {
		panic(fmt.Sprintf("assignment type does not match variable type: %s != %s", p0.Type, retType))
	}

	_, ok = t.declarations[p0.Name]
	if ok {
		t.err = fmt.Errorf(`%w: "%s" already declared`, ErrVariableAlreadyDeclared, p0.Name)
		panic("variable already declared")
	}

	t.declarations[p0.Name] = p0.Type

	return nil
}

func (t *typingVisitor) VisitStatementVariableDeclaration(p0 *parser.StatementVariableDeclaration) interface {
} {
	_, ok := t.declarations[p0.Name]
	if ok {
		t.err = fmt.Errorf(`%w: "%s" already declared`, ErrVariableAlreadyDeclared, p0.Name)
		panic("variable already declared")
	}

	t.declarations[p0.Name] = p0.Type
	return nil
}

// asserIntType asserts that the given expression is an integer type.
// It will panic if it is not.
func (t *typingVisitor) asserIntType(dt parser.Expression) {
	res, ok := dt.Accept(t).(*types.DataType)
	if !ok {
		panic("expected integer type")
	}
	if res == nil {
		panic("expected integer type")
	}

	if !res.Equals(types.IntType) {
		panic("expected integer type")
	}
}

// assertBoolType asserts that the given expression is a boolean type.
// It will panic if it is not.
func (t *typingVisitor) assertBoolType(dt parser.Expression) {
	res, ok := dt.Accept(t).(*types.DataType)
	if !ok {
		panic("expected boolean type")
	}
	if res == nil {
		panic("expected boolean type")
	}

	if !res.Equals(types.BoolType) {
		panic("expected boolean type")
	}
}
