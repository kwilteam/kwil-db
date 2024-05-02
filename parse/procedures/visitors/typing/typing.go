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
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
	"github.com/kwilteam/kwil-db/parse/util"
)

func EnsureTyping(stmts []parser.Statement, procedure *types.Procedure, schema *types.Schema, cleanedInputs []*types.NamedType,
	sessionVars map[string]*types.DataType, errorListeners parseTypes.NativeErrorListener) (anonReceiverTypes []*types.DataType, err error) {

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
		errs:                   errorListeners,
	}

	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("panic: %v", r)
			} else {
				err = fmt.Errorf("panic: %w", err)
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

	// errs is the current error listener
	errs parseTypes.NativeErrorListener
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
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, "expected array")
	}

	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return &types.DataType{
		Name: r.Name,
	}
}

func (t *typingVisitor) VisitExpressionBlobLiteral(p0 *parser.ExpressionBlobLiteral) any {
	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return types.BlobType
}

func (t *typingVisitor) VisitExpressionBooleanLiteral(p0 *parser.ExpressionBooleanLiteral) any {
	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

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
			t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, err.Error())
			return types.UnknownType
		}

		// return type cast as the type if it exists
		if p0.TypeCast != nil {
			return p0.TypeCast
		}

		return returnType
	}

	params, returns, err := util.FindProcOrForeign(t.currentSchema, p0.Name)
	if err != nil {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
		return types.UnknownType
	}

	// check the args
	if len(p0.Arguments) != len(params) {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`expected %d arguments, got %d`, len(params), len(p0.Arguments)))
		return types.UnknownType
	}

	for i, arg := range p0.Arguments {
		argType := arg.Accept(t).(*types.DataType)
		if !argType.Equals(params[i]) {
			t.errs.NodeErr(arg.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`argument %d has wrong type`, i))
		}
	}

	// it must not return a table, and only return exactly one value
	if returns == nil {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`procedure "%s" does not return anything`, p0.Name))
		return types.UnknownType
	}

	if returns.IsTable {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`procedure "%s" returns a table`, p0.Name))
		return types.UnknownType
	}

	if len(returns.Fields) != 1 {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`procedure "%s" does not return exactly 1 value`, p0.Name))
		return types.UnknownType
	}

	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
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
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf(`%s: "%s"`, metadata.ErrUnknownForeignProcedure.Error(), p0.Name))
		return types.UnknownType
	}

	// we need to verify that there are exactly two contextual args, and that they are
	// both strings
	if len(p0.ContextArgs) != 2 {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, "expected exactly two contextual arguments")
		return types.UnknownType
	}
	for _, arg := range p0.ContextArgs {
		r, ok := arg.Accept(t).(*types.DataType)
		if !ok {
			panic("BUG: expected data type")
		}

		if !r.Equals(types.TextType) {
			t.errs.NodeErr(arg.GetNode(), parseTypes.ParseErrorTypeType, "expected text type")
		}
	}

	// check the args
	if len(p0.Arguments) != len(proc.Parameters) {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`expected %d arguments, got %d`, len(proc.Parameters), len(p0.Arguments)))
		return types.UnknownType
	}

	for i, arg := range p0.Arguments {
		argType := arg.Accept(t).(*types.DataType)
		if !argType.Equals(proc.Parameters[i]) {
			t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`argument %d has wrong type`, i))
		}
	}

	if proc.Returns == nil {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`foreign procedure "%s" does not return anything`, p0.Name))
		return types.UnknownType
	}

	// proc must return exactly one value that is not a table
	if proc.Returns.IsTable {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`foreign procedure "%s" returns a table`, p0.Name))
		return types.UnknownType
	}
	if len(proc.Returns.Fields) != 1 {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`foreign procedure "%s" does not return exactly 1 value`, p0.Name))
		return types.UnknownType
	}

	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
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
			t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, metadata.ErrComparisonTypesDoNotMatch.Error())
		}
	}

	return types.BoolType
}

func (t *typingVisitor) VisitExpressionFieldAccess(p0 *parser.ExpressionFieldAccess) any {
	anonType, ok := p0.Target.Accept(t).(map[string]*types.DataType)
	if !ok {
		// this is a bug, so we panic
		panic("expected anonymous type")
	}

	dt, ok := anonType[p0.Field]
	if !ok {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, metadata.ErrUnknownField.Error())
		return types.UnknownType
	}

	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return dt
}

func (t *typingVisitor) VisitExpressionIntLiteral(p0 *parser.ExpressionIntLiteral) any {
	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return types.IntType
}

func (t *typingVisitor) VisitExpressionMakeArray(p0 *parser.ExpressionMakeArray) any {
	var arrayType *types.DataType
	for _, e := range p0.Values {
		dataType, ok := e.Accept(t).(*types.DataType)
		if !ok {
			// this is a bug, so we panic
			panic(fmt.Sprintf("expected data type in array, got %T", e.Accept(t)))
		}

		if dataType.IsArray {
			t.errs.NodeErr(e.GetNode(), parseTypes.ParseErrorTypeType, "array cannot contain arrays")
			return types.UnknownType
		}

		if arrayType == nil {
			arrayType = dataType
			continue
		}

		if !arrayType.Equals(dataType) {
			t.errs.NodeErr(e.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("%s: %s != %s", metadata.ErrArrayElementTypesDoNotMatch.Error(), arrayType, dataType))
		}
	}

	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return &types.DataType{
		Name:    arrayType.Name,
		IsArray: true,
	}
}

func (t *typingVisitor) VisitExpressionNullLiteral(p0 *parser.ExpressionNullLiteral) any {
	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return types.NullType
}

func (t *typingVisitor) VisitExpressionParenthesized(p0 *parser.ExpressionParenthesized) any {
	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return p0.Expression.Accept(t)
}

func (t *typingVisitor) VisitExpressionTextLiteral(p0 *parser.ExpressionTextLiteral) any {
	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return types.TextType
}

func (t *typingVisitor) VisitExpressionVariable(p0 *parser.ExpressionVariable) any {
	dt, ok := t.declarations[p0.Name]
	if !ok {
		anonType, ok := t.anonymousDeclarations[p0.Name]
		if !ok {
			t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`%s: "%s"`, metadata.ErrUndeclaredVariable.Error(), util.UnformatParameterName(p0.Name)))
			return types.UnknownType
		}

		// return type cast as the type if it exists
		if p0.TypeCast != nil {
			return p0.TypeCast
		}

		return anonType
	}

	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return dt
}

func (t *typingVisitor) VisitLoopTargetCall(p0 *parser.LoopTargetCall) any {
	r := t.analyzeProcedureCall(p0.Call)

	if !r.IsTable {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, "loops on procedures must return a table")
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
	rel := t.analyzeSQL(p0.StatementLocation, p0.Statement)

	r := make(map[string]*types.DataType)
	err := rel.Loop(func(s string, a *typing.Attribute) error {
		r[s] = a.Type
		return nil
	})
	if err != nil {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
	}
	return r
}

func (t *typingVisitor) VisitLoopTargetVariable(p0 *parser.LoopTargetVariable) any {
	// must check that the variable is an array
	r, ok := p0.Variable.Accept(t).(*types.DataType)
	if !ok {
		// this is a bug, so we panic
		panic("expected data type")
	}

	if !r.IsArray {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, "expected array")
		return types.UnknownType
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
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("expected %d return values, got %d", len(returns.Fields), len(p0.Variables)))
		return nil
	}

	if returns.IsTable {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, "procedure returns a table, which cannot be assigned to variables")
	}

	for i, v := range p0.Variables {
		if v == nil {
			// skip if nil, since it is an anonymous receiver
			t.anonymousReceiverTypes = append(t.anonymousReceiverTypes, returns.Fields[i].Type)
			continue
		}
		varType, ok := t.declarations[*v]
		if !ok {
			t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`%s: "%s"`, metadata.ErrUndeclaredVariable.Error(), util.UnformatParameterName(*v)))
			continue
		}

		if !varType.Equals(returns.Fields[i].Type) {
			t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`variable %s has wrong type`, util.UnformatParameterName(*v)))
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
				t.errs.NodeErr(call.GetNode(), parseTypes.ParseErrorTypeType, err.Error())
				return &types.ProcedureReturn{}
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
			t.errs.NodeErr(call.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
			return &types.ProcedureReturn{}
		}

		// check the args
		if len(call.Arguments) != len(params) {
			t.errs.NodeErr(call.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`expected %d arguments, got %d`, len(params), len(call.Arguments)))
			return &types.ProcedureReturn{}
		}

		for i, arg := range call.Arguments {
			argType := arg.Accept(t).(*types.DataType)
			if !argType.Equals(params[i]) {
				t.errs.NodeErr(arg.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`argument %d has wrong type`, i))
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
			t.errs.NodeErr(call.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf(`%s: "%s"`, metadata.ErrUnknownForeignProcedure.Error(), call.Name))
			return &types.ProcedureReturn{}
		}

		// we need to verify that there are exactly two contextual args, and that they are
		// both strings
		if len(call.ContextArgs) != 2 {
			t.errs.NodeErr(call.GetNode(), parseTypes.ParseErrorTypeType, "expected exactly two contextual arguments")
			return &types.ProcedureReturn{}
		}
		for _, arg := range call.ContextArgs {
			r, ok := arg.Accept(t).(*types.DataType)
			if !ok {
				panic("BUG: expected data type")
			}

			if !r.Equals(types.TextType) {
				t.errs.NodeErr(arg.GetNode(), parseTypes.ParseErrorTypeType, "expected text type")
			}
		}

		// check the args
		if len(call.Arguments) != len(proc.Parameters) {
			t.errs.NodeErr(call.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`expected %d arguments, got %d`, len(proc.Parameters), len(call.Arguments)))
			return &types.ProcedureReturn{}
		}

		for i, arg := range call.Arguments {
			argType := arg.Accept(t).(*types.DataType)
			if !argType.Equals(proc.Parameters[i]) {
				t.errs.NodeErr(call.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`argument %d has wrong type`, i))
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
	if t.loopTarget == "" {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, metadata.ErrBreakUsedOutsideOfLoop.Error())
	}
	return nil
}

func (t *typingVisitor) VisitStatementForLoop(p0 *parser.StatementForLoop) any {
	// set the loop target so children know we are in a loop
	t.loopTarget = p0.Variable

	switch target := p0.Target.(type) {
	case *parser.LoopTargetVariable:
		r := target.Accept(t).(*types.DataType)

		// we will not declare these as anonymous, since this is simply
		// a field in an array of a known type
		_, ok := t.declarations[p0.Variable]
		if ok {
			t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, metadata.ErrVariableAlreadyDeclared.Error())
			return nil
		}

		t.declarations[p0.Variable] = r
	case *parser.LoopTargetCall:
		// this can be either a table or a custom data type
		r := target.Accept(t).(map[string]*types.DataType)
		_, ok := t.anonymousDeclarations[p0.Variable]
		if ok {
			t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, metadata.ErrVariableAlreadyDeclared.Error())
		}

		t.anonymousDeclarations[p0.Variable] = r
	case *parser.LoopTargetRange:
		r := target.Accept(t).(*types.DataType) // int

		// we will not declare these as anonymous, since it is a simple int
		_, ok := t.declarations[p0.Variable]
		if ok {
			t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, metadata.ErrVariableAlreadyDeclared.Error())
		}

		t.declarations[p0.Variable] = r
	case *parser.LoopTargetSQL:
		r := target.Accept(t).(map[string]*types.DataType)

		_, ok := t.anonymousDeclarations[p0.Variable]
		if ok {
			t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, metadata.ErrVariableAlreadyDeclared.Error())
		}

		t.anonymousDeclarations[p0.Variable] = r
	default:
		panic("unknown loop target")
	}

	for _, stmt := range p0.Body {
		stmt.Accept(t)
	}

	// unset the loop target
	t.loopTarget = ""

	return nil
}

// analyzeSQL analyzes the given SQL statement and returns the resulting relation.
func (t *typingVisitor) analyzeSQL(location *parseTypes.Node, stmt tree.AstNode) *typing.Relation {
	// we create a new error listener taking into account our current position
	errLis := t.errs.Child("sql-types", location.StartLine, location.StartCol)
	m, err := typing.AnalyzeTypes(stmt, t.currentSchema.Tables, &typing.AnalyzeOptions{
		BindParams:       t.declarations,
		Qualify:          true,
		VerifyProcedures: true,
		Schema:           t.currentSchema,
		ErrorListener:    errLis,
	})
	if err != nil {
		panic(err)
	}

	t.errs.Add(errLis.Errors()...)

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
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, "procedure does not return anything")
		return nil
	}

	if t.currentProcedure.Returns.IsTable {
		if p0.SQL == nil {
			t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, "procedure returning table must have a return a SQL statement")
			return nil
		}

		r := t.analyzeSQL(p0.SQLLocation, p0.SQL)

		for _, col := range t.currentProcedure.Returns.Fields {
			attr, ok := r.Attribute(col.Name)
			if !ok {
				t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf(`missing column: procedure expects column "%s" in return table`, col.Name))
				continue
			}

			if !col.Type.Equals(attr.Type) {
				t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf(`column "%s" returns wrong type`, col.Name))
			}
		}
	} else {
		if p0.Values == nil {
			t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, "procedure expects return values")
			return nil
		}

		if len(p0.Values) != len(t.currentProcedure.Returns.Fields) {
			t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("expected %d return values, got %d", len(t.currentProcedure.Returns.Fields), len(p0.Values)))
			return nil
		}

		for i, v := range p0.Values {
			r, ok := v.Accept(t).(*types.DataType)
			if !ok {
				panic("BUG: expected data type")
			}

			if !t.currentProcedure.Returns.Fields[i].Type.Equals(r) {
				t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("%s: expected: %s received: %s", metadata.ErrIncorrectReturnType.Error(), t.currentProcedure.Returns.Fields[i], r))
			}
		}
	}

	return nil
}

func (t *typingVisitor) VisitStatementReturnNext(p0 *parser.StatementReturnNext) any {
	if t.loopTarget == "" {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, metadata.ErrReturnNextUsedOutsideOfLoop.Error())
		return nil
	}

	if t.currentProcedure.Returns == nil {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, "procedure does not return anything")
		return nil
	}

	if !t.currentProcedure.Returns.IsTable {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, metadata.ErrReturnNextUsedInNonTableProc.Error())
		return nil
	}

	if len(p0.Returns) != len(t.currentProcedure.Returns.Fields) {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, metadata.ErrReturnNextInvalidCount.Error())
	}

	for i, col := range t.currentProcedure.Returns.Fields {
		r, ok := p0.Returns[i].Accept(t).(*types.DataType)
		if !ok {
			panic("BUG: expected data type")
		}

		if !col.Type.Equals(r) {
			t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("%s: expected: %s received: %s", metadata.ErrIncorrectReturnType.Error(), col.Type, r))
		}
	}

	return nil
}

func (t *typingVisitor) VisitStatementSQL(p0 *parser.StatementSQL) any {
	t.analyzeSQL(p0.StatementLocation, p0.Statement)
	// we do not care about the return value
	return nil
}

func (t *typingVisitor) VisitStatementVariableAssignment(p0 *parser.StatementVariableAssignment) any {
	typ, ok := t.declarations[p0.Name]
	if !ok {
		// I don't think this can happen
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf(`%s: "%s"`, metadata.ErrUndeclaredVariable.Error(), util.UnformatParameterName(p0.Name)))
		return nil
	}

	r, ok := p0.Value.Accept(t).(*types.DataType)
	if !ok {
		panic("BUG: expected data type")
	}

	if !typ.Equals(r) {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("%s: expected: %s received: %s", metadata.ErrAssignmentTypeMismatch.Error(), typ, r))
	}

	return nil
}

func (t *typingVisitor) VisitStatementVariableAssignmentWithDeclaration(p0 *parser.StatementVariableAssignmentWithDeclaration) any {
	retType, ok := p0.Value.Accept(t).(*types.DataType)
	if !ok {
		panic("BUG: expected data type")
	}

	if !p0.Type.Equals(retType) {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("%s: expected: %s received: %s", metadata.ErrAssignmentTypeMismatch.Error(), p0.Type, retType))
	}

	_, ok = t.declarations[p0.Name]
	if ok {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, metadata.ErrVariableAlreadyDeclared.Error())
		return nil
	}

	t.declarations[p0.Name] = p0.Type

	return nil
}

func (t *typingVisitor) VisitStatementVariableDeclaration(p0 *parser.StatementVariableDeclaration) any {
	_, ok := t.declarations[p0.Name]
	if ok {
		t.errs.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, metadata.ErrVariableAlreadyDeclared.Error())
		return nil
	}

	t.declarations[p0.Name] = p0.Type
	return nil
}

// asserIntType asserts that the given expression is an integer type.
// It will panic if it is not.
func (t *typingVisitor) asserIntType(dt parser.Expression) {
	res, ok := dt.Accept(t).(*types.DataType)
	if !ok {
		t.errs.NodeErr(dt.GetNode(), parseTypes.ParseErrorTypeType, "expected integer type")
		return
	}
	if res == nil {
		t.errs.NodeErr(dt.GetNode(), parseTypes.ParseErrorTypeType, "expected integer type")
		return
	}

	if !res.Equals(types.IntType) {
		t.errs.NodeErr(dt.GetNode(), parseTypes.ParseErrorTypeType, "expected integer type")
		return
	}
}

// assertBoolType asserts that the given expression is a boolean type.
// It will panic if it is not.
func (t *typingVisitor) assertBoolType(dt parser.Expression) {
	res, ok := dt.Accept(t).(*types.DataType)
	if !ok {
		t.errs.NodeErr(dt.GetNode(), parseTypes.ParseErrorTypeType, "expected boolean type")
		return
	}
	if res == nil {
		t.errs.NodeErr(dt.GetNode(), parseTypes.ParseErrorTypeType, "expected boolean type")
		return
	}

	if !res.Equals(types.BoolType) {
		t.errs.NodeErr(dt.GetNode(), parseTypes.ParseErrorTypeType, "expected boolean type")
		return
	}
}
