// package typing checks the type safety of a procedure.
// It will return data types for expressions, map[string]DataType for loop terms,
// and nothing for statements.
package typing

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/metadata"
	"github.com/kwilteam/kwil-db/parse/procedures/parser"
	"github.com/kwilteam/kwil-db/parse/sql/sqlanalyzer/typing"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"github.com/kwilteam/kwil-db/parse/util"
)

func EnsureTyping(stmts []parser.Statement, procedure *types.Procedure, schema *types.Schema, cleanedInputs []*types.NamedType,
	sessionVars map[string]*types.DataType) (anonReceiverTypes []*types.DataType, err error) {

	declarations := make(map[string]*types.DataType)
	for _, param := range cleanedInputs {
		declarations[param.Name] = param.Type
	}

	for v, typ := range sessionVars {
		_, ok := declarations[v]
		if ok {
			// this should never happen, since session variables have a unique
			// prefix
			return nil, fmt.Errorf("session variable %s collision", v)
		}

		declarations[v] = typ
	}

	t := &typingVisitor{
		currentSchema:          schema,
		declarations:           declarations,
		currentProcedure:       procedure,
		anonymousDeclarations:  make(map[string]map[string]*types.DataType),
		anonymousReceiverTypes: make([]*types.DataType, 0),
	}

	defer func() {
		if r := recover(); r != nil {
			if t.err != nil {
				err = t.err
			} else {
				var ok bool
				err, ok = r.(error)
				if !ok {
					err = fmt.Errorf("panic: %v", r)
				}
			}
		}
	}()

	for _, stmt := range stmts {
		stmt.Accept(t)
	}

	return t.anonymousReceiverTypes, nil
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

	// loopTarget is the anonymous declaration of the current loop target.
	// Its type can be found in anonymousDeclarations.
	// If we are not in a loop, this will be an empty string.
	loopTarget string

	// anonymousReceiverTypes holds the types of the anonymous receivers
	anonymousReceiverTypes []*types.DataType

	// holds the last error that occurred
	err error
}

var _ parser.Visitor = &typingVisitor{}

func (t *typingVisitor) VisitExpressionArithmetic(p0 *parser.ExpressionArithmetic) any {
	t.asserIntType(p0.Left)
	t.asserIntType(p0.Right)
	return types.IntType
}

func (t *typingVisitor) VisitExpressionArrayAccess(p0 *parser.ExpressionArrayAccess) any {
	t.asserIntType(p0.Index)

	r, ok := p0.Target.Accept(t).(*types.DataType)
	if !ok {
		panic("BUG: expected data type")
	}
	if !r.IsArray {
		panic("expected array")
	}

	return &types.DataType{
		Name: r.Name,
	}
}

func (t *typingVisitor) VisitExpressionBlobLiteral(p0 *parser.ExpressionBlobLiteral) any {
	return types.BlobType
}

func (t *typingVisitor) VisitExpressionBooleanLiteral(p0 *parser.ExpressionBooleanLiteral) any {
	return types.BoolType
}

func (t *typingVisitor) VisitExpressionCall(p0 *parser.ExpressionCall) any {
	funcDef, ok := metadata.Functions[p0.Name]
	if ok {
		argsTypes := make([]*types.DataType, len(p0.Arguments))
		for i, arg := range p0.Arguments {
			argsTypes[i] = arg.Accept(t).(*types.DataType)
		}

		returnType, err := funcDef.ValidateArgs(argsTypes)
		if err != nil {
			panic(err)
		}

		return returnType
	}

	params, returns, err := util.FindProcOrForeign(t.currentSchema, p0.Name)
	if err != nil {
		panic(err)
	}

	// check the args
	if len(p0.Arguments) != len(params) {
		panic(fmt.Sprintf(`expected %d arguments, got %d`, len(params), len(p0.Arguments)))
	}

	for i, arg := range p0.Arguments {
		argType := arg.Accept(t).(*types.DataType)
		if !argType.Equals(params[i]) {
			panic(fmt.Sprintf(`argument %d has wrong type`, i))
		}
	}

	// it must not return a table, and only return exactly one value
	if returns == nil {
		panic(fmt.Sprintf(`procedure "%s" does not return anything, so it cannot be called as an expression`, p0.Name))
	}

	if returns.IsTable {
		panic(fmt.Sprintf(`procedure "%s" returns a table, which cannot be called as an expression`, p0.Name))
	}

	if len(returns.Fields) != 1 {
		panic(fmt.Sprintf(`procedure must "%s" return exactly one value to be called as an expression`, p0.Name))
	}

	return returns.Fields[0].Type
}

func (t *typingVisitor) VisitExpressionForeignCall(p0 *parser.ExpressionForeignCall) any {
	// foreign call must be defined in the schema.
	// We will reverse-clean to get the original name,
	// and search for it in the foreign procedures.
	var proc *types.ForeignProcedure
	found := false

	for _, proc = range t.currentSchema.ForeignProcedures {
		if proc.Name == p0.Name {
			found = true
			break
		}
	}
	if !found {
		panic(fmt.Errorf(`%w: "%s"`, metadata.ErrUnknownForeignProcedure, p0.Name))
	}

	// we need to verify that there are exactly two contextual args, and that they are
	// both strings
	if len(p0.ContextArgs) != 2 {
		panic("expected exactly two contextual arguments")
	}
	for _, arg := range p0.ContextArgs {
		r, ok := arg.Accept(t).(*types.DataType)
		if !ok {
			panic("BUG: expected data type")
		}

		if !r.Equals(types.TextType) {
			panic("expected text type")
		}
	}

	// check the args
	if len(p0.Arguments) != len(proc.Parameters) {
		panic(fmt.Sprintf(`expected %d arguments, got %d`, len(proc.Parameters), len(p0.Arguments)))
	}

	for i, arg := range p0.Arguments {
		argType := arg.Accept(t).(*types.DataType)
		if !argType.Equals(proc.Parameters[i]) {
			panic(fmt.Sprintf(`argument %d has wrong type`, i))
		}
	}

	if proc.Returns == nil {
		panic("procedure does not return anything")
	}

	// proc must return exactly one value that is not a table
	if proc.Returns.IsTable {
		panic("foreign procedure returns a table, which cannot be called as an expression")
	}
	if len(proc.Returns.Fields) != 1 {
		panic("foreign procedure must return exactly one value to be called as an expression")
	}
	return proc.Returns.Fields[0].Type
}

func (t *typingVisitor) VisitExpressionComparison(p0 *parser.ExpressionComparison) any {
	left := p0.Left.Accept(t).(*types.DataType)
	right := p0.Right.Accept(t).(*types.DataType)

	// left and right must either be the same type,
	// or one of them must be null
	if !left.Equals(right) {
		if !left.Equals(types.NullType) && !right.Equals(types.NullType) {
			panic("comparison types do not match")
		}
	}

	return types.BoolType
}

func (t *typingVisitor) VisitExpressionFieldAccess(p0 *parser.ExpressionFieldAccess) any {
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

func (t *typingVisitor) VisitExpressionIntLiteral(p0 *parser.ExpressionIntLiteral) any {
	return types.IntType
}

func (t *typingVisitor) VisitExpressionMakeArray(p0 *parser.ExpressionMakeArray) any {
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

func (t *typingVisitor) VisitExpressionNullLiteral(p0 *parser.ExpressionNullLiteral) any {
	return types.NullType
}

func (t *typingVisitor) VisitExpressionParenthesized(p0 *parser.ExpressionParenthesized) any {
	return p0.Expression.Accept(t)
}

func (t *typingVisitor) VisitExpressionTextLiteral(p0 *parser.ExpressionTextLiteral) any {
	return types.TextType
}

func (t *typingVisitor) VisitExpressionVariable(p0 *parser.ExpressionVariable) any {
	dt, ok := t.declarations[p0.Name]
	if !ok {
		anonType, ok := t.anonymousDeclarations[p0.Name]
		if !ok {
			panic(fmt.Errorf(`%w: "%s"`, metadata.ErrUndeclaredVariable, util.UnformatParameterName(p0.Name)))
		}

		return anonType
	}

	return dt
}

func (t *typingVisitor) VisitLoopTargetCall(p0 *parser.LoopTargetCall) any {
	r := t.analyzeProcedureCall(p0.Call)

	if !r.IsTable {
		panic("loops on procedures must return a table")
	}

	vals := make(map[string]*types.DataType)
	for _, col := range r.Fields {
		vals[col.Name] = col.Type
	}

	return vals
}

func (t *typingVisitor) VisitLoopTargetRange(p0 *parser.LoopTargetRange) any {
	t.asserIntType(p0.Start)
	t.asserIntType(p0.End)
	return types.IntType
}

func (t *typingVisitor) VisitLoopTargetSQL(p0 *parser.LoopTargetSQL) any {
	rel := t.analyzeSQL(p0.Statement)

	r := make(map[string]*types.DataType)
	err := rel.Loop(func(s string, a *typing.Attribute) error {
		r[s] = a.Type
		return nil
	})
	if err != nil {
		panic(err)
	}
	return r
}

func (t *typingVisitor) VisitLoopTargetVariable(p0 *parser.LoopTargetVariable) any {
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

func (t *typingVisitor) VisitStatementProcedureCall(p0 *parser.StatementProcedureCall) any {
	returns := t.analyzeProcedureCall(p0.Call)

	// a procedure that returns data can choose
	// not to use the return values
	if len(p0.Variables) == 0 {
		return nil
	}

	if len(returns.Fields) != len(p0.Variables) {
		panic(fmt.Sprintf("expected %d return values, got %d", len(returns.Fields), len(p0.Variables)))
	}

	if returns.IsTable {
		panic("procedure returns a table, which cannot be assigned to variables")
	}

	for i, v := range p0.Variables {
		if v == nil {
			// skip if nil, since it is an anonymous receiver
			t.anonymousReceiverTypes = append(t.anonymousReceiverTypes, returns.Fields[i].Type)
			continue
		}
		varType, ok := t.declarations[*v]
		if !ok {
			panic(fmt.Errorf(`%w: "%s"`, metadata.ErrUndeclaredVariable, util.UnformatParameterName(*v)))
		}

		if !varType.Equals(returns.Fields[i].Type) {
			panic(fmt.Sprintf("variable %s has wrong type", util.UnformatParameterName(*v)))
		}
	}

	return nil
}

// analyzeProcedureCall is used to visit a procedure call and get info on the return type.
// This is kept separate from the visitor, since the visit visits procedure/function calls
// as expressions. When used as expressions, procedures must return exactly 1 value
// (e.g. if a < other_val() {...}). This is used to return more detailed information.
func (t *typingVisitor) analyzeProcedureCall(p0 parser.ICallExpression) *types.ProcedureReturn {
	switch call := p0.(type) {
	case *parser.ExpressionCall:
		// check if it is a function
		funcDef, ok := metadata.Functions[call.Name]
		if ok {
			argsTypes := make([]*types.DataType, len(call.Arguments))
			for i, arg := range call.Arguments {
				argsTypes[i] = arg.Accept(t).(*types.DataType)
			}

			returnType, err := funcDef.ValidateArgs(argsTypes)
			if err != nil {
				panic(err)
			}

			return &types.ProcedureReturn{
				Fields: []*types.NamedType{
					{
						Type: returnType,
					},
				},
			}
		}

		params, returns, err := util.FindProcOrForeign(t.currentSchema, call.Name)
		if err != nil {
			panic(err)
		}

		// check the args
		if len(call.Arguments) != len(params) {
			panic(fmt.Sprintf(`expected %d arguments, got %d`, len(params), len(call.Arguments)))
		}

		for i, arg := range call.Arguments {
			argType := arg.Accept(t).(*types.DataType)
			if !argType.Equals(params[i]) {
				panic(fmt.Sprintf(`argument %d has wrong type`, i))
			}
		}

		if returns == nil {
			return &types.ProcedureReturn{} // avoid nil pointer
		}

		return returns
	case *parser.ExpressionForeignCall:
		// foreign call must be defined in the schema.
		// We will reverse-clean to get the original name,
		// and search for it in the foreign procedures.
		var proc *types.ForeignProcedure
		found := false

		for _, proc = range t.currentSchema.ForeignProcedures {
			if proc.Name == call.Name {
				found = true
				break
			}
		}
		if !found {
			panic(fmt.Errorf(`%w: "%s"`, metadata.ErrUnknownForeignProcedure, call.Name))
		}

		// we need to verify that there are exactly two contextual args, and that they are
		// both strings
		if len(call.ContextArgs) != 2 {
			panic("expected exactly two contextual arguments")
		}
		for _, arg := range call.ContextArgs {
			r, ok := arg.Accept(t).(*types.DataType)
			if !ok {
				panic("BUG: expected data type")
			}

			if !r.Equals(types.TextType) {
				panic("expected text type")
			}
		}

		// check the args
		if len(call.Arguments) != len(proc.Parameters) {
			panic(fmt.Sprintf(`expected %d arguments, got %d`, len(proc.Parameters), len(call.Arguments)))
		}

		for i, arg := range call.Arguments {
			argType := arg.Accept(t).(*types.DataType)
			if !argType.Equals(proc.Parameters[i]) {
				panic(fmt.Sprintf(`argument %d has wrong type`, i))
			}
		}

		if proc.Returns == nil {
			return &types.ProcedureReturn{} // avoid nil pointer
		}

		return proc.Returns
	default:
		panic("unknown call type")
	}
}

func (t *typingVisitor) VisitStatementBreak(p0 *parser.StatementBreak) any {
	return nil
}

func (t *typingVisitor) VisitStatementForLoop(p0 *parser.StatementForLoop) any {
	t.loopTarget = p0.Variable

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

	t.loopTarget = ""

	return nil
}

// analyzeSQL analyzes the given SQL statement and returns the resulting relation.
func (t *typingVisitor) analyzeSQL(stmt tree.AstNode) *typing.Relation {
	// TODO: we have a problem here where the sql analyzer cannot analyze
	// the @ vars, since it is using the current_setting() command.
	m, err := typing.AnalyzeTypes(stmt, t.currentSchema.Tables, &typing.AnalyzeOptions{
		BindParams:       t.declarations,
		Qualify:          true,
		VerifyProcedures: true,
		Schema:           t.currentSchema,
	})
	if err != nil {
		panic(err)
	}

	return m
}

func (t *typingVisitor) VisitStatementIf(p0 *parser.StatementIf) any {
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

func (t *typingVisitor) VisitStatementReturn(p0 *parser.StatementReturn) any {
	if t.currentProcedure.Returns == nil {
		panic("procedure does not return anything")
	}

	if t.currentProcedure.Returns.IsTable {
		if p0.SQL == nil {
			panic("procedure returning table must have a return a SQL statement")
		}
		r := t.analyzeSQL(p0.SQL)

		for _, col := range t.currentProcedure.Returns.Fields {
			attr, ok := r.Attribute(col.Name)
			if !ok {
				panic(fmt.Sprintf(`column "%s" not found in return table`, col.Name))
			}

			if !col.Type.Equals(attr.Type) {
				panic(fmt.Sprintf(`column "%s" has wrong type`, col.Name))
			}
		}
	} else {
		if p0.Values == nil {
			panic(fmt.Sprintf("procedure %s expects return values", t.currentProcedure.Name))
		}

		if len(p0.Values) != len(t.currentProcedure.Returns.Fields) {
			panic(fmt.Sprintf("expected %d return values, got %d", len(t.currentProcedure.Returns.Fields), len(p0.Values)))
		}

		for i, v := range p0.Values {
			r, ok := v.Accept(t).(*types.DataType)
			if !ok {
				panic("BUG: expected data type")
			}

			if !t.currentProcedure.Returns.Fields[i].Type.Equals(r) {
				panic(fmt.Sprintf("return type does not match procedure return type: %s != %s", t.currentProcedure.Returns.Fields[i], r))
			}
		}
	}

	return nil
}

func (t *typingVisitor) VisitStatementReturnNext(p0 *parser.StatementReturnNext) any {
	if t.loopTarget == "" {
		panic("RETURN NEXT can only be used in a loop")
	}

	if t.currentProcedure.Returns == nil {
		panic("procedure does not return anything")
	}

	if !t.currentProcedure.Returns.IsTable {
		panic("RETURN NEXT can only be used in procedures that return a table")
	}

	if len(p0.Returns) != len(t.currentProcedure.Returns.Fields) {
		panic("RETURN NEXT must return the same number of fields as the procedure return")
	}

	for i, col := range t.currentProcedure.Returns.Fields {
		r, ok := p0.Returns[i].Accept(t).(*types.DataType)
		if !ok {
			panic("BUG: expected data type")
		}

		if !col.Type.Equals(r) {
			panic(fmt.Sprintf("return type does not match procedure return type: %s != %s", col.Type, r))
		}
	}

	return nil
}

func (t *typingVisitor) VisitStatementSQL(p0 *parser.StatementSQL) any {
	t.analyzeSQL(p0.Statement)
	// we do not care about the return value
	return nil
}

func (t *typingVisitor) VisitStatementVariableAssignment(p0 *parser.StatementVariableAssignment) any {
	typ, ok := t.declarations[p0.Name]
	if !ok {
		panic(fmt.Errorf(`%w: "%s"`, metadata.ErrUntypedVariable, util.UnformatParameterName(p0.Name)))
	}

	r, ok := p0.Value.Accept(t).(*types.DataType)
	if !ok {
		panic("BUG: expected data type")
	}

	if !typ.Equals(r) {
		panic(fmt.Sprintf("assignment type does not match variable type: %s != %s", typ, r))
	}

	return nil
}

func (t *typingVisitor) VisitStatementVariableAssignmentWithDeclaration(p0 *parser.StatementVariableAssignmentWithDeclaration) any {
	retType, ok := p0.Value.Accept(t).(*types.DataType)
	if !ok {
		panic("BUG: expected data type")
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

func (t *typingVisitor) VisitStatementVariableDeclaration(p0 *parser.StatementVariableDeclaration) any {
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
