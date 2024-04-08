package parser

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/parse/sql/tree"
)

type Accepter interface {
	Accept(Visitor) interface{}
}

type Statement interface {
	Accepter
	statement()
}

type baseStatement struct{}

func (baseStatement) statement() {}
func (e *StatementVariableDeclaration) Accept(v Visitor) interface{} {
	return v.VisitStatementVariableDeclaration(e)
}

type StatementVariableDeclaration struct {
	*baseStatement
	// Name is the name of the variable.
	// It is case-insensitive.
	// It does include the $.
	Name string

	// Type is the type of the variable.
	// If it is a custom type, it is the name of the custom type.
	// If it is a built-in type, it is the name of the type.
	Type *types.DataType
}

func (e *StatementVariableAssignment) Accept(v Visitor) interface{} {
	return v.VisitStatementVariableAssignment(e)
}

type StatementVariableAssignment struct {
	baseStatement
	// Name is the name of the variable.
	// It is case-insensitive.
	// It does include the $.
	Name string
	// Value is the value to assign to the variable.
	Value Expression
}

func (e *StatementVariableAssignmentWithDeclaration) Accept(v Visitor) interface{} {
	return v.VisitStatementVariableAssignmentWithDeclaration(e)
}

type StatementVariableAssignmentWithDeclaration struct {
	baseStatement
	// Name is the name of the variable.
	// It is case-insensitive.
	// It does include the $.
	Name string

	// Type is the type of the variable.
	// If it is a custom type, it is the name of the custom type.
	// If it is a built-in type, it is the name of the type.
	Type *types.DataType

	// Value is the value to assign to the variable.
	Value Expression
}

func (e *StatementProcedureCall) Accept(v Visitor) interface{} {
	return v.VisitStatementProcedureCall(e)
}

type StatementProcedureCall struct {
	baseStatement
	Variables []string
	Call      *ExpressionCall
}

func (e *StatementForLoop) Accept(v Visitor) interface{} {
	return v.VisitStatementForLoop(e)
}

type StatementForLoop struct {
	baseStatement
	// Variable is the variable to assign the value to.
	Variable string
	// Target is the target of the loop.
	Target LoopTarget

	// Body is the body of the loop.
	Body []Statement
}

// LoopTarget is the target of the loop.
type LoopTarget interface {
	Accepter
	loopTarget()
}

type baseLoopTarget struct{}

func (baseLoopTarget) loopTarget() {}

func (e *LoopTargetRange) Accept(v Visitor) interface{} {
	return v.VisitLoopTargetRange(e)
}

type LoopTargetRange struct {
	baseLoopTarget
	// Start is the start of the range.
	Start Expression
	// End is the end of the range.
	End Expression
}

func (e *LoopTargetCall) Accept(v Visitor) interface{} {
	return v.VisitLoopTargetCall(e)
}

type LoopTargetCall struct {
	baseLoopTarget
	// Call is the procedure call to loop through.
	// It must return either an array or a table.
	Call *ExpressionCall
}

func (e *LoopTargetVariable) Accept(v Visitor) interface{} {
	return v.VisitLoopTargetVariable(e)
}

type LoopTargetSQL struct {
	baseLoopTarget
	// Statement is the Statement statement to execute.
	Statement tree.AstNode
}

func (e *LoopTargetSQL) Accept(v Visitor) interface{} {
	return v.VisitLoopTargetSQL(e)
}

type LoopTargetVariable struct {
	baseLoopTarget
	// Variable is the variable to loop through.
	// It must be an array.
	Variable *ExpressionVariable
}

func (e *StatementIf) Accept(v Visitor) interface{} {
	return v.VisitStatementIf(e)
}

type StatementIf struct {
	baseStatement
	// IfThens are the if statements.
	// They are evaluated in order, as
	// IF ... THEN ... ELSEIF ... THEN ...
	IfThens []*IfThen

	// Else is the else statement.
	// It is evaluated if none of the ifs are true.
	// It is optional.
	Else []Statement
}

type IfThen struct {
	If   Expression
	Then []Statement
}

func (e *StatementSQL) Accept(v Visitor) interface{} {
	return v.VisitStatementSQL(e)
}

type StatementSQL struct {
	baseStatement
	// Statement is the SQL statement to execute.
	Statement tree.AstNode
}

func (e *StatementReturn) Accept(v Visitor) interface{} {
	return v.VisitStatementReturn(e)
}

type StatementReturn struct {
	baseStatement
	// Values is the value to return.
	// It can be nil.
	Values []Expression

	// SQL is the SQL statement to execute.
	// If this is not nil, Value must be nil.
	SQL tree.AstNode
}

func (e *StatementReturnNext) Accept(v Visitor) interface{} {
	return v.VisitStatementReturnNext(e)
}

type StatementReturnNext struct {
	baseStatement
	// Variable is the variable to return.
	Variable string
}

func (e *StatementBreak) Accept(v Visitor) interface{} {
	return v.VisitStatementBreak(e)
}

// StatementBreak is a statement that breaks out of a loop.
type StatementBreak struct {
	baseStatement
}

type Expression interface {
	Accepter
	expression()
}

type baseExpression struct{}

func (baseExpression) expression() {}

func (e *ExpressionTextLiteral) Accept(v Visitor) interface{} {
	return v.VisitExpressionTextLiteral(e)
}

type ExpressionTextLiteral struct {
	baseExpression
	Value string
}

func (e *ExpressionBooleanLiteral) Accept(v Visitor) interface{} {
	return v.VisitExpressionBooleanLiteral(e)
}

type ExpressionBooleanLiteral struct {
	baseExpression
	Value bool
}

func (e *ExpressionIntLiteral) Accept(v Visitor) interface{} {
	return v.VisitExpressionIntLiteral(e)
}

type ExpressionIntLiteral struct {
	baseExpression
	Value int64
}

func (e *ExpressionNullLiteral) Accept(v Visitor) interface{} {
	return v.VisitExpressionNullLiteral(e)
}

type ExpressionNullLiteral struct {
	baseExpression
}

func (e *ExpressionBlobLiteral) Accept(v Visitor) interface{} {
	return v.VisitExpressionBlobLiteral(e)
}

type ExpressionBlobLiteral struct {
	baseExpression
	Value []byte
}

func (e *ExpressionMakeArray) Accept(v Visitor) interface{} {
	return v.VisitExpressionMakeArray(e)
}

type ExpressionMakeArray struct {
	baseExpression
	Values []Expression
}

func (e *ExpressionCall) Accept(v Visitor) interface{} {
	return v.VisitExpressionCall(e)
}

type ExpressionCall struct {
	baseExpression
	// Name is the name of the procedure.
	// It should always be lower case.
	Name string
	// Arguments are the arguments to the procedure.
	Arguments []Expression // can be nil
}

func (e *ExpressionVariable) Accept(v Visitor) interface{} {
	return v.VisitExpressionVariable(e)
}

type ExpressionVariable struct {
	baseExpression
	// Name is the name of the variable.
	// It is case-insensitive.
	// It does include the $.
	// It should include all fields, separated by dots.
	Name string
}

type VariablePrefix uint8

const (
	VariablePrefixDollar VariablePrefix = iota
	VariablePrefixAt
)

func (e *ExpressionArrayAccess) Accept(v Visitor) interface{} {
	return v.VisitExpressionArrayAccess(e)
}

type ExpressionArrayAccess struct {
	baseExpression
	// Target is the array to access the index from.
	Target Expression
	// Index is the index to access.
	Index Expression
}

func (e *ExpressionFieldAccess) Accept(v Visitor) interface{} {
	return v.VisitExpressionFieldAccess(e)
}

type ExpressionFieldAccess struct {
	baseExpression
	// Target is the object to access the field from.
	Target Expression
	// Field is the field to access.
	Field string
}

func (e *ExpressionParenthesized) Accept(v Visitor) interface{} {
	return v.VisitExpressionParenthesized(e)
}

type ExpressionParenthesized struct {
	baseExpression
	// Expression is the expression inside the parentheses.
	Expression Expression
}

func (e *ExpressionComparison) Accept(v Visitor) interface{} {
	return v.VisitExpressionComparison(e)
}

type ExpressionComparison struct {
	baseExpression
	Left     Expression
	Operator ComparisonOperator
	Right    Expression
}

func (e *ExpressionArithmetic) Accept(v Visitor) interface{} {
	return v.VisitExpressionArithmetic(e)
}

type ExpressionArithmetic struct {
	baseExpression
	Left     Expression
	Operator ArithmeticOperator
	Right    Expression
}

type ArithmeticOperator string

const (
	ArithmeticOperatorAdd ArithmeticOperator = "+"
	ArithmeticOperatorSub ArithmeticOperator = "-"
	ArithmeticOperatorMul ArithmeticOperator = "*"
	ArithmeticOperatorDiv ArithmeticOperator = "/"
	ArithmeticOperatorMod ArithmeticOperator = "%"
)

func (a *ArithmeticOperator) Validate() error {
	switch *a {
	case ArithmeticOperatorAdd, ArithmeticOperatorSub, ArithmeticOperatorMul, ArithmeticOperatorDiv, ArithmeticOperatorMod:
		return nil
	default:
		return fmt.Errorf("invalid arithmetic operator: %s", *a)
	}
}

type ComparisonOperator string

const (
	ComparisonOperatorEqual              ComparisonOperator = "="
	ComparisonOperatorNotEqual           ComparisonOperator = "!="
	ComparisonOperatorGreaterThan        ComparisonOperator = ">"
	ComparisonOperatorLessThan           ComparisonOperator = "<"
	ComparisonOperatorGreaterThanOrEqual ComparisonOperator = ">="
	ComparisonOperatorLessThanOrEqual    ComparisonOperator = "<="
)

func (c *ComparisonOperator) Validate() error {
	switch *c {
	case ComparisonOperatorEqual, ComparisonOperatorNotEqual, ComparisonOperatorGreaterThan, ComparisonOperatorLessThan, ComparisonOperatorGreaterThanOrEqual, ComparisonOperatorLessThanOrEqual:
		return nil
	default:
		return fmt.Errorf("invalid comparison operator: %s", *c)
	}
}

type Visitor interface {
	VisitStatementVariableDeclaration(*StatementVariableDeclaration) interface{}
	VisitStatementVariableAssignment(*StatementVariableAssignment) interface{}
	VisitStatementVariableAssignmentWithDeclaration(*StatementVariableAssignmentWithDeclaration) interface{}
	VisitStatementProcedureCall(*StatementProcedureCall) interface{}
	VisitStatementForLoop(*StatementForLoop) interface{}
	VisitStatementIf(*StatementIf) interface{}
	VisitStatementSQL(*StatementSQL) interface{}
	VisitStatementReturn(*StatementReturn) interface{}
	VisitStatementReturnNext(*StatementReturnNext) interface{}
	VisitStatementBreak(*StatementBreak) interface{}
	VisitExpressionTextLiteral(*ExpressionTextLiteral) interface{}
	VisitExpressionBooleanLiteral(*ExpressionBooleanLiteral) interface{}
	VisitExpressionIntLiteral(*ExpressionIntLiteral) interface{}
	VisitExpressionNullLiteral(*ExpressionNullLiteral) interface{}
	VisitExpressionBlobLiteral(*ExpressionBlobLiteral) interface{}
	VisitExpressionMakeArray(*ExpressionMakeArray) interface{}
	VisitExpressionCall(*ExpressionCall) interface{}
	VisitExpressionVariable(*ExpressionVariable) interface{}
	VisitExpressionArrayAccess(*ExpressionArrayAccess) interface{}
	VisitExpressionFieldAccess(*ExpressionFieldAccess) interface{}
	VisitExpressionParenthesized(*ExpressionParenthesized) interface{}
	VisitExpressionComparison(*ExpressionComparison) interface{}
	VisitExpressionArithmetic(*ExpressionArithmetic) interface{}
	VisitLoopTargetRange(*LoopTargetRange) interface{}
	VisitLoopTargetCall(*LoopTargetCall) interface{}
	VisitLoopTargetVariable(*LoopTargetVariable) interface{}
	VisitLoopTargetSQL(*LoopTargetSQL) interface{}
}

// BaseVisitor is a base implementation of Visitor.
type BaseVisitor struct{}

var _ Visitor = &BaseVisitor{}

func (v *BaseVisitor) VisitStatementVariableDeclaration(s *StatementVariableDeclaration) interface{} {
	return nil
}
func (v *BaseVisitor) VisitStatementVariableAssignment(*StatementVariableAssignment) interface{} {
	return nil
}
func (v *BaseVisitor) VisitStatementVariableAssignmentWithDeclaration(*StatementVariableAssignmentWithDeclaration) interface{} {
	return nil
}
func (v *BaseVisitor) VisitStatementProcedureCall(*StatementProcedureCall) interface{} {
	return nil
}
func (v *BaseVisitor) VisitStatementForLoop(*StatementForLoop) interface{} { return nil }
func (v *BaseVisitor) VisitStatementIf(*StatementIf) interface{}           { return nil }
func (v *BaseVisitor) VisitStatementSQL(*StatementSQL) interface{}         { return nil }
func (v *BaseVisitor) VisitStatementReturn(*StatementReturn) interface{}   { return nil }
func (v *BaseVisitor) VisitStatementReturnNext(*StatementReturnNext) interface{} {
	return nil
}
func (v *BaseVisitor) VisitStatementBreak(*StatementBreak) interface{} {
	return nil
}
func (v *BaseVisitor) VisitExpressionTextLiteral(*ExpressionTextLiteral) interface{} { return nil }
func (v *BaseVisitor) VisitExpressionBooleanLiteral(*ExpressionBooleanLiteral) interface{} {
	return nil
}
func (v *BaseVisitor) VisitExpressionIntLiteral(*ExpressionIntLiteral) interface{}       { return nil }
func (v *BaseVisitor) VisitExpressionNullLiteral(*ExpressionNullLiteral) interface{}     { return nil }
func (v *BaseVisitor) VisitExpressionBlobLiteral(*ExpressionBlobLiteral) interface{}     { return nil }
func (v *BaseVisitor) VisitExpressionMakeArray(*ExpressionMakeArray) interface{}         { return nil }
func (v *BaseVisitor) VisitExpressionCall(*ExpressionCall) interface{}                   { return nil }
func (v *BaseVisitor) VisitExpressionVariable(*ExpressionVariable) interface{}           { return nil }
func (v *BaseVisitor) VisitExpressionArrayAccess(*ExpressionArrayAccess) interface{}     { return nil }
func (v *BaseVisitor) VisitExpressionFieldAccess(*ExpressionFieldAccess) interface{}     { return nil }
func (v *BaseVisitor) VisitExpressionParenthesized(*ExpressionParenthesized) interface{} { return nil }
func (v *BaseVisitor) VisitExpressionComparison(*ExpressionComparison) interface{}       { return nil }
func (v *BaseVisitor) VisitExpressionArithmetic(*ExpressionArithmetic) interface{}       { return nil }
func (b *BaseVisitor) VisitLoopTargetCall(p0 *LoopTargetCall) interface{}                { return nil }
func (b *BaseVisitor) VisitLoopTargetRange(p0 *LoopTargetRange) interface{}              { return nil }
func (b *BaseVisitor) VisitLoopTargetVariable(p0 *LoopTargetVariable) interface{}        { return nil }
func (b *BaseVisitor) VisitLoopTargetSQL(p0 *LoopTargetSQL) interface{}                  { return nil }
