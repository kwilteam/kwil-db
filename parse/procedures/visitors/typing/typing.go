// package typing checks the type safety of a procedure.
// It will return data types for expressions, map[string]DataType for loop terms,
// and nothing for statements.
package typing

import (
	"errors"
	"fmt"
	"runtime"

	coreTypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/metadata"
	"github.com/kwilteam/kwil-db/parse/procedures/parser"
	"github.com/kwilteam/kwil-db/parse/sql/sqlanalyzer/typing"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"github.com/kwilteam/kwil-db/parse/types"
	"github.com/kwilteam/kwil-db/parse/util"
)

func EnsureTyping(stmts []parser.Statement, procedure *coreTypes.Procedure, schema *coreTypes.Schema, cleanedInputs []*coreTypes.NamedType,
	sessionVars map[string]*coreTypes.DataType, errorListeners types.NativeErrorListener) (anonReceiverTypes []*coreTypes.DataType, err error) {

	declarations := make(map[string]*coreTypes.DataType)
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
		anonymousDeclarations:  make(map[string]map[string]*coreTypes.DataType),
		anonymousReceiverTypes: make([]*coreTypes.DataType, 0),
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

			// add stack trace since this is a bug:
			buf := make([]byte, 1<<16)
			stackSize := runtime.Stack(buf, false)
			err = fmt.Errorf("%w\n%s", err, buf[:stackSize])
		}
	}()

	for _, stmt := range stmts {
		stmt.Accept(t)
	}

	return t.anonymousReceiverTypes, nil
}

type typingVisitor struct {
	// currentSchema holds the schema of the current procedure
	currentSchema *coreTypes.Schema
	// declarations holds information about all variable declarations in the procedure
	declarations map[string]*coreTypes.DataType
	// currentProcedure holds the information about the current procedure
	currentProcedure *coreTypes.Procedure
	// anonymousDeclarations holds information about all anonymous variable declarations in the procedure
	// currently, the only anonymous declarations are loop targets (e.g. FOR i IN SELECT * FROM users).
	// anonymousDeclarations are essentially anonymous compound types, so the map maps the name
	// to the fields to their coreTypes.
	anonymousDeclarations map[string]map[string]*coreTypes.DataType

	// loopTarget is the anonymous declaration of the current loop target.
	// Its type can be found in anonymousDeclarations.
	// If we are not in a loop, this will be an empty string.
	loopTarget string

	// anonymousReceiverTypes holds the types of the anonymous receivers
	anonymousReceiverTypes []*coreTypes.DataType

	// errs is the current error listener
	errs types.NativeErrorListener
}

var _ parser.Visitor = &typingVisitor{}

func (t *typingVisitor) VisitExpressionArithmetic(p0 *parser.ExpressionArithmetic) any {
	left := p0.Left.Accept(t).(*coreTypes.DataType)
	right := p0.Right.Accept(t).(*coreTypes.DataType)

	if !left.Equals(right) {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, leftRightErr(types.ErrArithmeticType, left, right))
	}

	if !isNumeric(left) {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, types.ErrNotNumericType)
	}

	return coreTypes.IntType
}

func (t *typingVisitor) VisitExpressionArrayAccess(p0 *parser.ExpressionArrayAccess) any {
	t.asserIntType(p0.Index)

	r, ok := p0.Target.Accept(t).(*coreTypes.DataType)
	if !ok {
		panic("BUG: expected data type")
	}
	if !r.IsArray {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, errors.New("expected array"))
	}

	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return &coreTypes.DataType{
		Name: r.Name,
	}
}

func (t *typingVisitor) VisitExpressionBlobLiteral(p0 *parser.ExpressionBlobLiteral) any {
	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return coreTypes.BlobType
}

func (t *typingVisitor) VisitExpressionBooleanLiteral(p0 *parser.ExpressionBooleanLiteral) any {
	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return coreTypes.BoolType
}

func (t *typingVisitor) VisitExpressionCall(p0 *parser.ExpressionCall) any {
	funcDef, ok := metadata.Functions[p0.Name]
	if ok {
		argsTypes := make([]*coreTypes.DataType, len(p0.Arguments))
		for i, arg := range p0.Arguments {
			argsTypes[i] = arg.Accept(t).(*coreTypes.DataType)
		}

		returnType, err := funcDef.ValidateArgs(argsTypes)
		if err != nil {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, err)
			return coreTypes.UnknownType
		}

		// return type cast as the type if it exists
		if p0.TypeCast != nil {
			return p0.TypeCast
		}

		return returnType
	}

	params, returns, err := util.FindProcOrForeign(t.currentSchema, p0.Name)
	if err != nil {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, err)
		return coreTypes.UnknownType
	}

	// check the args
	if len(p0.Arguments) != len(params) {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, fmt.Errorf(`expected %d arguments, got %d`, len(params), len(p0.Arguments)))
		return coreTypes.UnknownType
	}

	for i, arg := range p0.Arguments {
		argType := arg.Accept(t).(*coreTypes.DataType)
		if !argType.Equals(params[i]) {
			t.errs.NodeErr(arg.GetNode(), types.ParseErrorTypeType, fmt.Errorf(`argument %d has wrong type`, i))
		}
	}

	// it must not return a table, and only return exactly one value
	if returns == nil {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, fmt.Errorf(`procedure "%s" does not return anything`, p0.Name))
		return coreTypes.UnknownType
	}

	if returns.IsTable {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, fmt.Errorf(`%w: procedure "%s" returns a table`, types.ErrAssignment, p0.Name))
		return coreTypes.UnknownType
	}

	if len(returns.Fields) != 1 {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, fmt.Errorf(`procedure "%s" does not return exactly 1 value`, p0.Name))
		return coreTypes.UnknownType
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
	var proc *coreTypes.ForeignProcedure
	found := false

	for _, proc = range t.currentSchema.ForeignProcedures {
		if proc.Name == p0.Name {
			found = true
			break
		}
	}
	if !found {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, fmt.Errorf(`%w: "%s"`, types.ErrUnknownForeignProcedure, p0.Name))
		return coreTypes.UnknownType
	}

	// we need to verify that there are exactly two contextual args, and that they are
	// both strings
	if len(p0.ContextArgs) != 2 {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, errors.New("expected exactly two contextual arguments"))
		return coreTypes.UnknownType
	}
	for _, arg := range p0.ContextArgs {
		r, ok := arg.Accept(t).(*coreTypes.DataType)
		if !ok {
			panic("BUG: expected data type")
		}

		if !r.Equals(coreTypes.TextType) {
			t.errs.NodeErr(arg.GetNode(), types.ParseErrorTypeType, errors.New("expected text type"))
		}
	}

	// check the args
	if len(p0.Arguments) != len(proc.Parameters) {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, fmt.Errorf(`expected %d arguments, got %d`, len(proc.Parameters), len(p0.Arguments)))
		return coreTypes.UnknownType
	}

	for i, arg := range p0.Arguments {
		argType := arg.Accept(t).(*coreTypes.DataType)
		if !argType.Equals(proc.Parameters[i]) {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, fmt.Errorf(`argument %d has wrong type`, i))
		}
	}

	if proc.Returns == nil {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, fmt.Errorf(`foreign procedure "%s" does not return anything`, p0.Name))
		return coreTypes.UnknownType
	}

	// proc must return exactly one value that is not a table
	if proc.Returns.IsTable {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, fmt.Errorf(`%w: foreign procedure "%s" returns a table`, types.ErrAssignment, p0.Name))
		return coreTypes.UnknownType
	}
	if len(proc.Returns.Fields) != 1 {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, fmt.Errorf(`foreign procedure "%s" does not return exactly 1 value`, p0.Name))
		return coreTypes.UnknownType
	}

	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return proc.Returns.Fields[0].Type
}

func (t *typingVisitor) VisitExpressionComparison(p0 *parser.ExpressionComparison) any {
	left := p0.Left.Accept(t).(*coreTypes.DataType)
	right := p0.Right.Accept(t).(*coreTypes.DataType)

	// left and right must either be the same type,
	// or one of them must be null
	if !left.Equals(right) {
		if !left.Equals(coreTypes.NullType) && !right.Equals(coreTypes.NullType) {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, types.ErrComparisonType)
		}
	}

	return coreTypes.BoolType
}

func (t *typingVisitor) VisitExpressionFieldAccess(p0 *parser.ExpressionFieldAccess) any {
	anonType, ok := p0.Target.Accept(t).(map[string]*coreTypes.DataType)
	if !ok {
		// this is a bug, so we panic
		panic("expected anonymous type")
	}

	dt, ok := anonType[p0.Field]
	if !ok {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, types.ErrUnknownField)
		return coreTypes.UnknownType
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

	return coreTypes.IntType
}

func (t *typingVisitor) VisitExpressionMakeArray(p0 *parser.ExpressionMakeArray) any {
	var arrayType *coreTypes.DataType
	for _, e := range p0.Values {
		dataType, ok := e.Accept(t).(*coreTypes.DataType)
		if !ok {
			// this is a bug, so we panic
			panic(fmt.Sprintf("expected data type in array, got %T", e.Accept(t)))
		}

		if dataType.IsArray {
			t.errs.NodeErr(e.GetNode(), types.ParseErrorTypeType, errors.New("array cannot contain arrays"))
			return coreTypes.UnknownType
		}

		if arrayType == nil {
			arrayType = dataType
			continue
		}

		if !arrayType.Equals(dataType) {
			t.errs.NodeErr(e.GetNode(), types.ParseErrorTypeType,
				fmt.Errorf("%w: %s != %s", types.ErrArrayType, arrayType, dataType))
		}
	}

	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return &coreTypes.DataType{
		Name:    arrayType.Name,
		IsArray: true,
	}
}

func (t *typingVisitor) VisitExpressionNullLiteral(p0 *parser.ExpressionNullLiteral) any {
	// return type cast as the type if it exists
	if p0.TypeCast != nil {
		return p0.TypeCast
	}

	return coreTypes.NullType
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

	return coreTypes.TextType
}

func (t *typingVisitor) VisitExpressionVariable(p0 *parser.ExpressionVariable) any {
	dt, ok := t.declarations[p0.Name]
	if !ok {
		anonType, ok := t.anonymousDeclarations[p0.Name]
		if !ok {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType,
				fmt.Errorf(`%w: "%s"`, types.ErrUndeclaredVariable, util.UnformatParameterName(p0.Name)))
			return coreTypes.UnknownType
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
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, errors.New("loops on procedures must return a table"))
	}

	vals := make(map[string]*coreTypes.DataType)
	for _, col := range r.Fields {
		vals[col.Name] = col.Type
	}

	return vals
}

func (t *typingVisitor) VisitLoopTargetRange(p0 *parser.LoopTargetRange) any {
	t.asserIntType(p0.Start)
	t.asserIntType(p0.End)
	return coreTypes.IntType
}

func (t *typingVisitor) VisitLoopTargetSQL(p0 *parser.LoopTargetSQL) any {
	rel := t.analyzeSQL(p0.StatementLocation, p0.Statement)

	r := make(map[string]*coreTypes.DataType)
	err := rel.Loop(func(s string, a *typing.Attribute) error {
		r[s] = a.Type
		return nil
	})
	if err != nil {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, err)
	}
	return r
}

func (t *typingVisitor) VisitLoopTargetVariable(p0 *parser.LoopTargetVariable) any {
	// must check that the variable is an array
	r, ok := p0.Variable.Accept(t).(*coreTypes.DataType)
	if !ok {
		// this is a bug, so we panic
		panic("expected data type")
	}

	if !r.IsArray {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType, errors.New("expected array"))
		return coreTypes.UnknownType
	}

	return &coreTypes.DataType{
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
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType,
			fmt.Errorf(`%w: expected: %d received: %d`, types.ErrReturnCount, len(returns.Fields), len(p0.Variables)))
		return nil
	}

	if returns.IsTable {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType,
			fmt.Errorf("%w: procedure returns a table, which cannot be assigned to variables", types.ErrAssignment))
	}

	for i, v := range p0.Variables {
		if v == nil {
			// skip if nil, since it is an anonymous receiver
			t.anonymousReceiverTypes = append(t.anonymousReceiverTypes, returns.Fields[i].Type)
			continue
		}
		varType, ok := t.declarations[*v]
		if !ok {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType,
				fmt.Errorf(`%w: "%s"`, types.ErrUndeclaredVariable, util.UnformatParameterName(*v)))
			continue
		}

		if !varType.Equals(returns.Fields[i].Type) {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType,
				fmt.Errorf(`%w: cannot assign return type %s to variable "%s" of type %s`, types.ErrAssignment, returns.Fields[i].Type.String(), util.UnformatParameterName(*v), varType.String()))
		}
	}

	return nil
}

// analyzeProcedureCall is used to visit a procedure call and get info on the return type.
// This is kept separate from the visitor, since the visit visits procedure/function calls
// as expressions. When used as expressions, procedures must return exactly 1 value
// (e.g. if a < other_val() {...}). This is used to return more detailed information.
func (t *typingVisitor) analyzeProcedureCall(p0 parser.ICallExpression) *coreTypes.ProcedureReturn {
	switch call := p0.(type) {
	case *parser.ExpressionCall:
		// check if it is a function
		funcDef, ok := metadata.Functions[call.Name]
		if ok {
			argsTypes := make([]*coreTypes.DataType, len(call.Arguments))
			for i, arg := range call.Arguments {
				argsTypes[i] = arg.Accept(t).(*coreTypes.DataType)
			}

			returnType, err := funcDef.ValidateArgs(argsTypes)
			if err != nil {
				t.errs.NodeErr(call.GetNode(), types.ParseErrorTypeType, err)
				return &coreTypes.ProcedureReturn{}
			}

			return &coreTypes.ProcedureReturn{
				Fields: []*coreTypes.NamedType{
					{
						Type: returnType,
					},
				},
			}
		}

		params, returns, err := util.FindProcOrForeign(t.currentSchema, call.Name)
		if err != nil {
			t.errs.NodeErr(call.GetNode(), types.ParseErrorTypeSemantic, err)
			return &coreTypes.ProcedureReturn{}
		}

		// check the args
		if len(call.Arguments) != len(params) {
			t.errs.NodeErr(call.GetNode(), types.ParseErrorTypeType, fmt.Errorf(`expected %d arguments, got %d`, len(params), len(call.Arguments)))
			return &coreTypes.ProcedureReturn{}
		}

		for i, arg := range call.Arguments {
			argType := arg.Accept(t).(*coreTypes.DataType)
			if !argType.Equals(params[i]) {
				t.errs.NodeErr(arg.GetNode(), types.ParseErrorTypeType, fmt.Errorf(`argument %d has wrong type`, i))
			}
		}

		if returns == nil {
			return &coreTypes.ProcedureReturn{} // avoid nil pointer
		}

		return returns
	case *parser.ExpressionForeignCall:
		// foreign call must be defined in the schema.
		// We will reverse-clean to get the original name,
		// and search for it in the foreign procedures.
		var proc *coreTypes.ForeignProcedure
		found := false

		for _, proc = range t.currentSchema.ForeignProcedures {
			if proc.Name == call.Name {
				found = true
				break
			}
		}
		if !found {
			t.errs.NodeErr(call.GetNode(), types.ParseErrorTypeSemantic, fmt.Errorf(`%w: "%s"`, types.ErrUnknownForeignProcedure, call.Name))
			return &coreTypes.ProcedureReturn{}
		}

		// we need to verify that there are exactly two contextual args, and that they are
		// both strings
		if len(call.ContextArgs) != 2 {
			t.errs.NodeErr(call.GetNode(), types.ParseErrorTypeType, errors.New("expected exactly two contextual arguments"))
			return &coreTypes.ProcedureReturn{}
		}
		for _, arg := range call.ContextArgs {
			r, ok := arg.Accept(t).(*coreTypes.DataType)
			if !ok {
				panic("BUG: expected data type")
			}

			if !r.Equals(coreTypes.TextType) {
				t.errs.NodeErr(arg.GetNode(), types.ParseErrorTypeType, errors.New("expected text type"))
			}
		}

		// check the args
		if len(call.Arguments) != len(proc.Parameters) {
			t.errs.NodeErr(call.GetNode(), types.ParseErrorTypeType, fmt.Errorf(`expected %d arguments, got %d`, len(proc.Parameters), len(call.Arguments)))
			return &coreTypes.ProcedureReturn{}
		}

		for i, arg := range call.Arguments {
			argType := arg.Accept(t).(*coreTypes.DataType)
			if !argType.Equals(proc.Parameters[i]) {
				t.errs.NodeErr(call.GetNode(), types.ParseErrorTypeType, fmt.Errorf(`argument %d has wrong type`, i))
			}
		}

		if proc.Returns == nil {
			return &coreTypes.ProcedureReturn{} // avoid nil pointer
		}

		return proc.Returns
	default:
		panic("unknown call type")
	}
}

func (t *typingVisitor) VisitStatementBreak(p0 *parser.StatementBreak) any {
	if t.loopTarget == "" {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, types.ErrBreakUsedOutsideOfLoop)
	}
	return nil
}

func (t *typingVisitor) VisitStatementForLoop(p0 *parser.StatementForLoop) any {
	// set the loop target so children know we are in a loop
	t.loopTarget = p0.Variable

	switch target := p0.Target.(type) {
	case *parser.LoopTargetVariable:
		r := target.Accept(t).(*coreTypes.DataType)

		// we will not declare these as anonymous, since this is simply
		// a field in an array of a known type
		_, ok := t.declarations[p0.Variable]
		if ok {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, types.ErrVariableAlreadyDeclared)
			return nil
		}

		t.declarations[p0.Variable] = r

		// check that we are iterating over an array
		dataType, ok := t.declarations[target.Variable.Name]
		if !ok {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic,
				fmt.Errorf(`%w: "%s"`, types.ErrUndeclaredVariable, util.UnformatParameterName(target.Variable.Name)))
			return nil
		}

		if !dataType.IsArray {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType,
				fmt.Errorf(`%w: cannot loop over type %s`, types.ErrInvalidIterable, dataType.String()))
			return nil
		}
	case *parser.LoopTargetCall:
		// this can be either a table or a custom data type
		r := target.Accept(t).(map[string]*coreTypes.DataType)
		_, ok := t.anonymousDeclarations[p0.Variable]
		if ok {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, types.ErrVariableAlreadyDeclared)
		}

		t.anonymousDeclarations[p0.Variable] = r
	case *parser.LoopTargetRange:
		r := target.Accept(t).(*coreTypes.DataType) // int

		// we will not declare these as anonymous, since it is a simple int
		_, ok := t.declarations[p0.Variable]
		if ok {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, types.ErrVariableAlreadyDeclared)
		}

		t.declarations[p0.Variable] = r
	case *parser.LoopTargetSQL:
		r := target.Accept(t).(map[string]*coreTypes.DataType)

		_, ok := t.anonymousDeclarations[p0.Variable]
		if ok {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, types.ErrVariableAlreadyDeclared)
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
func (t *typingVisitor) analyzeSQL(location *types.Node, stmt tree.AstNode) *typing.Relation {
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
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, errors.New("procedure does not return anything"))
		return nil
	}

	if t.currentProcedure.Returns.IsTable {
		if p0.SQL == nil {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, errors.New("procedure returning table must have a return a SQL statement"))
			return nil
		}

		r := t.analyzeSQL(p0.SQLLocation, p0.SQL)

		for _, col := range t.currentProcedure.Returns.Fields {
			attr, ok := r.Attribute(col.Name)
			if !ok {
				t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic,
					fmt.Errorf(`missing column: procedure expects column "%s" in return table`, col.Name))
				continue
			}

			if !col.Type.Equals(attr.Type) {
				t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType,
					fmt.Errorf(`column "%s" returns wrong type`, col.Name))
			}
		}
	} else {
		if p0.Values == nil {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, errors.New("procedure expects return values"))
			return nil
		}

		if len(p0.Values) != len(t.currentProcedure.Returns.Fields) {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType,
				fmt.Errorf(`%w: expected: %d received: %d`, types.ErrReturnCount, len(t.currentProcedure.Returns.Fields), len(p0.Values)))
			return nil
		}

		for i, v := range p0.Values {
			r, ok := v.Accept(t).(*coreTypes.DataType)
			if !ok {
				panic("BUG: expected data type")
			}

			if !t.currentProcedure.Returns.Fields[i].Type.Equals(r) {
				t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType,
					fmt.Errorf("%w: expected: %s received: %s", types.ErrAssignment, t.currentProcedure.Returns.Fields[i].Type.String(), r.String()))
			}
		}
	}

	return nil
}

func (t *typingVisitor) VisitStatementReturnNext(p0 *parser.StatementReturnNext) any {
	if t.loopTarget == "" {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, types.ErrReturnNextUsedOutsideOfLoop)
		return nil
	}

	if t.currentProcedure.Returns == nil {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, errors.New("procedure does not return anything"))
		return nil
	}

	if !t.currentProcedure.Returns.IsTable {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, types.ErrReturnNextUsedInNonTableProc)
		return nil
	}

	if len(p0.Returns) != len(t.currentProcedure.Returns.Fields) {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, types.ErrReturnNextInvalidCount)
	}

	for i, col := range t.currentProcedure.Returns.Fields {
		r, ok := p0.Returns[i].Accept(t).(*coreTypes.DataType)
		if !ok {
			panic("BUG: expected data type")
		}

		if !col.Type.Equals(r) {
			t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType,
				fmt.Errorf("%w: expected: %s received: %s", types.ErrAssignment, col.Type.String(), r.String()))
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
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic,
			fmt.Errorf(`%w: "%s"`, types.ErrUntypedVariable, util.UnformatParameterName(p0.Name)))
		return nil
	}

	r, ok := p0.Value.Accept(t).(*coreTypes.DataType)
	if !ok {
		panic("BUG: expected data type")
	}

	if !typ.Equals(r) {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType,
			fmt.Errorf("%w: expected: %s received: %s", types.ErrAssignment, typ, r))
	}

	return nil
}

func (t *typingVisitor) VisitStatementVariableAssignmentWithDeclaration(p0 *parser.StatementVariableAssignmentWithDeclaration) any {
	retType, ok := p0.Value.Accept(t).(*coreTypes.DataType)
	if !ok {
		panic("BUG: expected data type")
	}

	if !p0.Type.Equals(retType) {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeType,
			fmt.Errorf("%w: expected: %s received: %s", types.ErrAssignment, p0.Type, retType))
	}

	_, ok = t.declarations[p0.Name]
	if ok {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, types.ErrVariableAlreadyDeclared)
		return nil
	}

	t.declarations[p0.Name] = p0.Type

	return nil
}

func (t *typingVisitor) VisitStatementVariableDeclaration(p0 *parser.StatementVariableDeclaration) any {
	_, ok := t.declarations[p0.Name]
	if ok {
		t.errs.NodeErr(p0.GetNode(), types.ParseErrorTypeSemantic, types.ErrVariableAlreadyDeclared)
		return nil
	}

	t.declarations[p0.Name] = p0.Type
	return nil
}

// asserIntType asserts that the given expression is an integer type.
// It will panic if it is not.
func (t *typingVisitor) asserIntType(dt parser.Expression) {
	res, ok := dt.Accept(t).(*coreTypes.DataType)
	if !ok {
		t.errs.NodeErr(dt.GetNode(), types.ParseErrorTypeType, errors.New("expected integer type"))
		return
	}
	if res == nil {
		t.errs.NodeErr(dt.GetNode(), types.ParseErrorTypeType, errors.New("expected integer type"))
		return
	}

	if !res.Equals(coreTypes.IntType) {
		t.errs.NodeErr(dt.GetNode(), types.ParseErrorTypeType, errors.New("expected integer type"))
		return
	}
}

// assertBoolType asserts that the given expression is a boolean type.
// It will panic if it is not.
func (t *typingVisitor) assertBoolType(dt parser.Expression) {
	res, ok := dt.Accept(t).(*coreTypes.DataType)
	if !ok {
		t.errs.NodeErr(dt.GetNode(), types.ParseErrorTypeType, errors.New("expected boolean type"))
		return
	}
	if res == nil {
		t.errs.NodeErr(dt.GetNode(), types.ParseErrorTypeType, errors.New("expected boolean type"))
		return
	}

	if !res.Equals(coreTypes.BoolType) {
		t.errs.NodeErr(dt.GetNode(), types.ParseErrorTypeType, errors.New("expected boolean type"))
		return
	}
}

// leftRightErr returns an error with the message that the left and right. It is
// used for incompatible coreTypes.
func leftRightErr(err error, left, right *coreTypes.DataType) error {
	return fmt.Errorf("%w: %s not comparable to %s", err, left.String(), right.String())
}

// isNumeric returns true if the given data type is numeric.
// right now, this is only int, but we are about to add numeric and uint256
// TODO: delete the above comment once we add numeric and uint256
func isNumeric(dt *coreTypes.DataType) bool {
	return dt.Equals(coreTypes.IntType)
}
