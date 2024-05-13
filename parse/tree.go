package parse

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	"github.com/holiman/uint256"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
)

// this file contains the ASTs for SQL, procedures, and actions.

// Node is a node in the AST.
type Node interface {
	Accept(Visitor) any
	GetNode() *parseTypes.Node
	Set(r antlr.ParserRuleContext)
	SetToken(t antlr.Token)
}

type typecastable struct {
	TypeCast *types.DataType
}

func (t *typecastable) Cast(t2 *types.DataType) {
	t.TypeCast = t2
}

type baseExpression struct{}

func (baseExpression) expr() {}

type baseSQLNode struct{}

func (baseSQLNode) sqlNode() {}

type baseProcedureNode struct{}

func (baseProcedureNode) procedureNode() {}

// Expression is an interface for all expressions.
type Expression interface {
	Node
	Cast(*types.DataType)
	expr()
}

// SQLNode is an interface for all nodes that can
// be used in SQL.
type SQLNode interface {
	Node
	sqlNode()
}

// ProcedureNode is an interface for all nodes that can
// be used in procedures.
type ProcedureNode interface {
	Node
	procedureNode()
}

// TODO: I'm not sure if we need SQLExpression and ProcedureExpression,
// or if we can simply treat them all the same.

// ProcedureExpression is an interface for all expressions that can
// be used in procedures. It is a subset of all expressions.
type ProcedureExpression interface {
	Expression
	ProcedureNode
}

// ExpressionLiteral is a literal expression.
type ExpressionLiteral struct {
	parseTypes.Node
	baseExpression
	typecastable
	baseSQLNode
	baseProcedureNode
	Type *types.DataType
	// Value is the value of the literal.
	// It must be of type string, int64, bool, *uint256.Int, *decimal.Decimal,
	// or nil
	Value any
}

func (e *ExpressionLiteral) Accept(v Visitor) any {
	return v.VisitExpressionLiteral(e)
}

// String returns the string representation of the literal.
func (e *ExpressionLiteral) String() string {
	switch v := e.Value.(type) {
	case string:
		return "'" + v + "'"
	case int64:
		return fmt.Sprint(v)
	case bool:
		return fmt.Sprint(v)
	case *uint256.Int:
		return v.String()
	case *decimal.Decimal:
		return v.String()
	case nil:
		return "null"
	default:
		panic("invalid literal value type: " + fmt.Sprintf("%T", v))
	}
}

type ExpressionCall interface {
	Expression
	isCall()
}

// ExpressionFunctionCall is a function call expression.
type ExpressionFunctionCall struct {
	parseTypes.Node
	typecastable
	baseExpression
	baseSQLNode
	baseProcedureNode
	// Name is the name of the function.
	Name string
	// Args are the arguments to the function call.
	// They are passed using ()
	Args []Expression
	// Distinct is true if the function call is a DISTINCT function call.
	Distinct bool
	// Star is true if the function call is a * function call.
	// If it is set, then Args must be empty.
	Star bool
}

func (e *ExpressionFunctionCall) Accept(v Visitor) any {
	return v.VisitExpressionFunctionCall(e)
}

func (e *ExpressionFunctionCall) isCall() {}

// ExpressionForeignCall is a call to an external procedure.
type ExpressionForeignCall struct {
	parseTypes.Node
	typecastable
	baseExpression
	baseSQLNode
	baseProcedureNode
	// Name is the name of the function.
	Name string
	// ContextualArgs are arguments that are contextual to the function call.
	// They are passed using []
	ContextualArgs []Expression
	// Args are the arguments to the function call.
	// They are passed using ()
	Args []Expression
}

func (e *ExpressionForeignCall) Accept(v Visitor) any {
	return v.VisitExpressionForeignCall(e)
}

func (e *ExpressionForeignCall) isCall()

// ExpressionVariable is a variable.
// This can either be $ or @ variables.
type ExpressionVariable struct {
	parseTypes.Node
	typecastable
	baseExpression
	baseSQLNode
	baseProcedureNode
	// Name is the naem of the variable,
	// without the $ or @.
	Name string
	// Prefix is the $ or @ prefix.
	Prefix VariablePrefix
}

func (e *ExpressionVariable) Accept(v Visitor) any {
	return v.VisitExpressionVariable(e)
}

func (e *ExpressionVariable) String() string {
	return string(e.Prefix) + e.Name
}

type VariablePrefix string

const (
	VariablePrefixDollar VariablePrefix = "$"
	VariablePrefixAt     VariablePrefix = "@"
)

// ExpressionArrayAccess accesses an array value.
type ExpressionArrayAccess struct {
	parseTypes.Node
	typecastable
	baseExpression
	baseSQLNode
	baseProcedureNode
	// Array is the array that is being accessed.
	Array Expression
	// Index is the index that is being accessed.
	Index Expression
}

func (e *ExpressionArrayAccess) Accept(v Visitor) any {
	return v.VisitExpressionArrayAccess(e)
}

// ExpressionMakeArray makes a new array.
type ExpressionMakeArray struct {
	parseTypes.Node
	baseExpression
	typecastable
	baseSQLNode
	baseProcedureNode
	Values []Expression
}

func (e *ExpressionMakeArray) Accept(v Visitor) any {
	return v.VisitExpressionMakeArray(e)
}

// ExpressionFieldAccess accesses a field in a record.
type ExpressionFieldAccess struct {
	parseTypes.Node
	baseExpression
	typecastable
	baseSQLNode
	baseProcedureNode
	// Record is the record that is being accessed.
	Record Expression
	// Field is the field that is being accessed.
	Field string
}

func (e *ExpressionFieldAccess) Accept(v Visitor) any {
	return v.VisitExpressionFieldAccess(e)
}

// ExpressionParenthesized is a parenthesized expression.
type ExpressionParenthesized struct {
	parseTypes.Node
	baseExpression
	typecastable
	baseSQLNode
	baseProcedureNode
	// Inner is the inner expression.
	Inner Expression
}

func (e *ExpressionParenthesized) Accept(v Visitor) any {
	return v.VisitExpressionParenthesized(e)
}

// ExpressionComparison is a comparison expression.
type ExpressionComparison struct {
	parseTypes.Node
	baseExpression
	baseSQLNode
	baseProcedureNode
	// Left is the left side of the comparison.
	Left Expression
	// Right is the right side of the comparison.
	Right Expression
	// Operator is the operator of the comparison.
	Operator ComparisonOperator
}

func (e *ExpressionComparison) Accept(v Visitor) any {
	return v.VisitExpressionComparison(e)
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

// ExpressionLogical is a logical expression.
type ExpressionLogical struct {
	parseTypes.Node
	baseExpression
	baseSQLNode
	// Left is the left side of the logical expression.
	Left Expression
	// Right is the right side of the logical expression.
	Right Expression
	// Operator is the operator of the logical expression.
	Operator LogicalOperator
}

func (e *ExpressionLogical) Accept(v Visitor) any {
	return v.VisitExpressionLogical(e)
}

type LogicalOperator string

const (
	LogicalOperatorAnd LogicalOperator = "and"
	LogicalOperatorOr  LogicalOperator = "or"
)

// ExpressionArithmetic is an arithmetic expression.
type ExpressionArithmetic struct {
	parseTypes.Node
	baseExpression
	baseSQLNode
	baseProcedureNode
	// Left is the left side of the arithmetic expression.
	Left Expression
	// Right is the right side of the arithmetic expression.
	Right Expression
	// Operator is the operator of the arithmetic expression.
	Operator ArithmeticOperator
}

func (e *ExpressionArithmetic) Accept(v Visitor) any {
	return v.VisitExpressionArithmetic(e)
}

type ArithmeticOperator string

const (
	ArithmeticOperatorAdd      ArithmeticOperator = "+"
	ArithmeticOperatorSubtract ArithmeticOperator = "-"
	ArithmeticOperatorMultiply ArithmeticOperator = "*"
	ArithmeticOperatorDivide   ArithmeticOperator = "/"
	ArithmeticOperatorModulo   ArithmeticOperator = "%"
	ArithmeticOperatorConcat   ArithmeticOperator = "||"
)

type ExpressionUnary struct {
	parseTypes.Node
	baseExpression
	baseSQLNode
	baseProcedureNode
	// Expression is the expression that is being operated on.
	Expression Expression
	// Operator is the operator of the unary expression.
	Operator UnaryOperator
}

func (e *ExpressionUnary) Accept(v Visitor) any {
	return v.VisitExpressionUnary(e)
}

type UnaryOperator string

const (
	UnaryOperatorNot UnaryOperator = "not"
)

// ExpressionColumn is a column in a table.
type ExpressionColumn struct {
	parseTypes.Node
	baseExpression
	typecastable
	baseSQLNode
	// Table is the table that the column is in.
	// It can be empty if the table has not been specified.
	Table string
	// Column is the name of the column.
	Column string
}

func (e *ExpressionColumn) Accept(v Visitor) any {
	return v.VisitExpressionColumn(e)
}

// ExpressionList is a list of expressions.
type ExpressionList struct {
	parseTypes.Node
	baseExpression
	baseSQLNode
	Expressions []Expression
}

func (e *ExpressionList) Accept(v Visitor) any {
	return v.VisitExpressionList(e)
}

// ExpressionCollate is an expression with a collation.
type ExpressionCollate struct {
	parseTypes.Node
	baseExpression
	baseSQLNode
	// Expression is the expression that is being collated.
	Expression Expression
	// Collation is the collation that is being used.
	Collation string
}

func (e *ExpressionCollate) Accept(v Visitor) any {
	return v.VisitExpressionCollate(e)
}

// ExpressionStringComparison is a string comparison expression.
type ExpressionStringComparison struct {
	parseTypes.Node
	baseExpression
	baseSQLNode
	// Left is the left side of the comparison.
	Left Expression
	// Right is the right side of the comparison.
	Right Expression
	// Operator is the operator of the comparison.
	Operator StringComparisonOperator
}

func (e *ExpressionStringComparison) Accept(v Visitor) any {
	return v.VisitExpressionStringComparison(e)
}

type StringComparisonOperator string

const (
	StringComparisonOperatorLike    StringComparisonOperator = "LIKE"
	StringComparisonOperatorNotLike StringComparisonOperator = "NOT LIKE"
)

// ExpressionIs is an IS expression.
type ExpressionIs struct {
	parseTypes.Node
	baseExpression
	baseSQLNode
	// Left is the left side of the IS expression.
	Left Expression
	// Right is the right side of the IS expression.
	Right Expression
	// Not is true if the IS expression is a NOT IS expression.
	Not bool
	// Distinct is true if the IS expression is a DISTINCT IS expression.
	Distinct bool
}

func (e *ExpressionIs) Accept(v Visitor) any {
	return v.VisitExpressionIs(e)
}

// ExpressionBetween is a BETWEEN expression.
type ExpressionBetween struct {
	parseTypes.Node
	baseExpression
	baseSQLNode
	// Expression is the expression that is being compared.
	Expression Expression
	// Lower is the left side of the BETWEEN expression.
	Lower Expression
	// Upper is the right side of the BETWEEN expression.
	Upper Expression
	// Not is true if the BETWEEN expression is a NOT BETWEEN expression.
	Not bool
}

func (e *ExpressionBetween) Accept(v Visitor) any {
	return v.VisitExpressionBetween(e)
}

type ExpressionIn struct {
	parseTypes.Node
	baseExpression
	baseSQLNode
	// Expression is the expression that is being compared.
	Expression Expression
	// List is the list of expressions that the expression is being compared to.
	// Either List or Subquery is set, but not both.
	List []Expression
	// Subquery is the subquery that the expression is being compared to.
	// Either List or Subquery is set, but not both.
	Subquery *SelectStatement
	// Not is true if the IN expression is a NOT IN expression.
	Not bool
}

func (e *ExpressionIn) Accept(v Visitor) any {
	return v.VisitExpressionIn(e)
}

// ExpressionSubquery is a subquery expression.
type ExpressionSubquery struct {
	parseTypes.Node
	baseExpression
	typecastable
	baseSQLNode
	Not      bool
	Exists   bool
	Subquery *SelectStatement
}

func (e *ExpressionSubquery) Accept(v Visitor) any {
	return v.VisitExpressionSubquery(e)
}

// ExpressionCase is a CASE expression.
type ExpressionCase struct {
	parseTypes.Node
	baseExpression
	baseSQLNode
	Case     Expression
	WhenThen [][2]Expression
	Else     Expression
}

func (e *ExpressionCase) Accept(v Visitor) any {
	return v.VisitExpressionCase(e)
}

// CommonTableExpression is a common table expression.
type CommonTableExpression struct {
	parseTypes.Node
	// Name is the name of the CTE.
	Name string
	// Columns are the columns of the CTE.
	Columns []string
	// Query is the query of the CTE.
	Query *SelectStatement
}

func (c *CommonTableExpression) Accept(v Visitor) any {
	return v.VisitCommonTableExpression(c)
}

// SQLStatement is a SQL statement.
type SQLStatement struct {
	parseTypes.Node
	CTEs []*CommonTableExpression
	// SQL can be an insert, update, delete, or select statement.
	SQL SQLCore
}

func (s *SQLStatement) Accept(v Visitor) any {
	return v.VisitSQLStatement(s)
}

// SQLCore is a top-level statement.
// It can be INSERT, UPDATE, DELETE, SELECT.
type SQLCore interface {
	SQLNode
	StmtType() SQLStatementType
}

type SQLStatementType string

const (
	SQLStatementTypeInsert SQLStatementType = "insert"
	SQLStatementTypeUpdate SQLStatementType = "update"
	SQLStatementTypeDelete SQLStatementType = "delete"
	SQLStatementTypeSelect SQLStatementType = "select"
)

// SelectStatement is a SELECT statement.
type SelectStatement struct {
	parseTypes.Node
	baseSQLNode
	SelectCores       []*SelectCore
	CompoundOperators []CompoundOperator
	Ordering          []*OrderingTerm
	Limit             Expression
	Offset            Expression
}

func (s *SelectStatement) Accept(v Visitor) any {
	return v.VisitSelectStatement(s)
}

func (SelectStatement) StmtType() SQLStatementType {
	return SQLStatementTypeSelect
}

type CompoundOperator string

const (
	CompoundOperatorUnion     CompoundOperator = "UNION"
	CompoundOperatorUnionAll  CompoundOperator = "UNION ALL"
	CompoundOperatorIntersect CompoundOperator = "INTERSECT"
	CompoundOperatorExcept    CompoundOperator = "EXCEPT"
)

// OrderingTerm is a term in an order by clause
type OrderingTerm struct {
	parseTypes.Node
	baseSQLNode
	Expression Expression
	Order      OrderType
	Nulls      NullOrder
}

type OrderType string

const (
	OrderTypeAsc  OrderType = "ASC"
	OrderTypeDesc OrderType = "DESC"
)

type NullOrder string

const (
	NullOrderFirst NullOrder = "FIRST"
	NullOrderLast  NullOrder = "LAST"
)

type SelectCore struct {
	parseTypes.Node
	baseSQLNode
	// Distinct is true if the SELECT statement is a DISTINCT SELECT statement.
	Distinct bool
	Columns  []ResultColumn
	From     Relation     // can be nil
	Joins    []*Join      // can be nil
	Where    Expression   // can be nil
	GroupBy  []Expression // can be nil
	Having   Expression   // can be nil
}

func (s *SelectCore) Accept(v Visitor) any {
	return v.VisitSelectCore(s)
}

type ResultColumn interface {
	Node
	ResultColumnType() ResultColumnType
}

type ResultColumnType string

const (
	ResultColumnTypeExpression ResultColumnType = "expression"
	ResultColumnTypeWildcard   ResultColumnType = "wildcare"
)

type ResultColumnExpression struct {
	parseTypes.Node
	baseSQLNode

	Expression Expression
	Alias      string // can be empty
}

func (r *ResultColumnExpression) Accept(v Visitor) any {
	return v.VisitResultColumnExpression(r)
}

func (r *ResultColumnExpression) ResultColumnType() ResultColumnType {
	return ResultColumnTypeExpression
}

type ResultColumnWildcard struct {
	parseTypes.Node
	baseSQLNode
	Table string // can be empty
}

func (r *ResultColumnWildcard) Accept(v Visitor) any {
	return v.VisitResultColumnWildcard(r)
}

func (r *ResultColumnWildcard) ResultColumnType() ResultColumnType {
	return ResultColumnTypeWildcard
}

type Relation interface {
	SQLNode
	relation()
}

type RelationTable struct {
	parseTypes.Node
	baseSQLNode
	Table string
	Alias string // can be empty
}

func (r *RelationTable) Accept(v Visitor) any {
	return v.VisitRelationTable(r)
}

func (RelationTable) relation() {}

type RelationSubquery struct {
	parseTypes.Node
	baseSQLNode
	Subquery *SelectStatement
	Alias    string // can be empty
}

func (r *RelationSubquery) Accept(v Visitor) any {
	return v.VisitRelationSubquery(r)
}

func (RelationSubquery) relation() {}

type RelationFunctionCall struct {
	parseTypes.Node
	baseSQLNode
	FunctionCall ExpressionCall
	Alias        string // can be empty
}

func (r *RelationFunctionCall) Accept(v Visitor) any {
	return v.VisitRelationFunctionCall(r)
}

func (RelationFunctionCall) relation() {}

// Join is a join in a SELECT statement.
type Join struct {
	parseTypes.Node
	baseSQLNode
	Type     JoinType
	Relation Relation
	On       Expression
}

func (j *Join) Accept(v Visitor) any {
	return v.VisitJoin(j)
}

type JoinType string

const (
	JoinTypeInner JoinType = "INNER"
	JoinTypeLeft  JoinType = "LEFT"
	JoinTypeRight JoinType = "RIGHT"
	JoinTypeFull  JoinType = "FULL"
)

type UpdateStatement struct {
	parseTypes.Node
	baseSQLNode
	Table     string
	Alias     string // can be empty
	SetClause []*UpdateSetClause
	From      Relation         // can be nil
	Joins     []*Join          // can be nil
	Where     Expression       // can be nil
	Returning *ReturningClause // can be nil
}

func (u *UpdateStatement) Accept(v Visitor) any {
	return v.VisitUpdateStatement(u)
}

func (u *UpdateStatement) StmtType() SQLStatementType {
	return SQLStatementTypeUpdate
}

type UpdateSetClause struct {
	parseTypes.Node
	baseSQLNode
	// Either Column or ColumnList is set, but not both.
	Column     string
	ColumnList []string
	Value      Expression
}

func (u *UpdateSetClause) Accept(v Visitor) any {
	return v.VisitUpdateSetClause(u)
}

type ReturningClause struct {
	parseTypes.Node
	baseSQLNode
	Columns []ResultColumn
}

func (r *ReturningClause) Accept(v Visitor) any {
	return v.VisitReturningClause(r)
}

type DeleteStatement struct {
	parseTypes.Node
	baseSQLNode

	Table     string
	Alias     string           // can be empty
	From      Relation         // can be nil
	Joins     []*Join          // can be nil
	Where     Expression       // can be nil
	Returning *ReturningClause // can be nil
}

func (d *DeleteStatement) StmtType() SQLStatementType {
	return SQLStatementTypeDelete
}

func (d *DeleteStatement) Accept(v Visitor) any {
	return v.VisitDeleteStatement(d)
}

type InsertStatement struct {
	parseTypes.Node
	baseSQLNode
	Table     string
	Alias     string   // can be empty
	Columns   []string // can be empty
	Values    []Expression
	Upsert    *UpsertClause    // can be nil
	Returning *ReturningClause // can be nil
}

func (i *InsertStatement) Accept(v Visitor) any {
	return v.VisitInsertStatement(i)
}

func (i *InsertStatement) StmtType() SQLStatementType {
	return SQLStatementTypeInsert
}

type UpsertClause struct {
	parseTypes.Node
	baseSQLNode
	ConflictColumns []string           // can be empty
	ConflictWhere   Expression         // can be nil
	DoUpdate        []*UpdateSetClause // if nil, then do nothing
	UpdateWhere     Expression         // can be nil
}

func (u *UpsertClause) Accept(v Visitor) any {
	return v.VisitUpsertClause(u)
}

// action ast:

type ActionStmt interface {
	ActionStmt() ActionStatementTypes
}

type ActionStatementTypes string

const (
	ActionStatementTypeExtensionCall ActionStatementTypes = "extension_call"
	ActionStatementTypeActionCall    ActionStatementTypes = "action_call"
	ActionStatementTypeSQL           ActionStatementTypes = "sql"
)

type ActionStmtSQL struct {
	parseTypes.Node
	SQL *SQLStatement
}

func (a *ActionStmtSQL) Accept(v Visitor) any {
	return v.VisitSQLStatement(a.SQL)
}

func (a *ActionStmtSQL) ActionStmt() ActionStatementTypes {
	return ActionStatementTypeSQL
}

type ActionStmtExtensionCall struct {
	parseTypes.Node
	Receivers []string
	Extension string
	Method    string
	Args      []ProcedureExpression
}

func (a *ActionStmtExtensionCall) Accept(v Visitor) any {
	return v.VisitExtensionCallStmt(a)
}

func (a *ActionStmtExtensionCall) ActionStmt() ActionStatementTypes {
	return ActionStatementTypeExtensionCall
}

type ActionStmtActionCall struct {
	parseTypes.Node
	Action string
	Args   []ProcedureExpression
}

func (a *ActionStmtActionCall) Accept(v Visitor) any {
	return v.VisitActionCallStmt(a)
}

func (a *ActionStmtActionCall) ActionStmt() ActionStatementTypes {
	return ActionStatementTypeActionCall
}

// procedure ast:

type ProcedureStmt interface {
	ProcedureNode
	procedureStmt()
}

type baseProcedureStmt struct {
	parseTypes.Node
}

func (baseProcedureStmt) procedureStmt() {}

// ProcedureStmtDeclaration is a variable declaration in a procedure.
type ProcedureStmtDeclaration struct {
	baseProcedureStmt
	// Variable is the variable that is being declared.
	Variable string
	Type     *types.DataType
}

func (p *ProcedureStmtDeclaration) Accept(v Visitor) any {
	return v.VisitProcedureStmtDeclaration(p)
}

// ProcedureStmtAssignment is a variable assignment in a procedure.
// It should only be called on variables that have already been declared.
type ProcedureStmtAssignment struct {
	baseProcedureStmt
	// Variable is the variable that is being assigned.
	Variable string
	// Value is the value that is being assigned.
	Value ProcedureExpression
}

func (p *ProcedureStmtAssignment) Accept(v Visitor) any {
	return v.VisitProcedureStmtAssignment(p)
}

// ProcedureStmtDeclareAndAssign declares and assigns a variable in a procedure.
type ProcedureStmtDeclareAndAssign struct {
	baseProcedureStmt
	// Variable is the variable that is being declared and assigned.
	Variable string
	// Type is the type of the variable.
	Type *types.DataType
	// Value is the value that is being assigned.
	Value ProcedureExpression
}

func (p *ProcedureStmtDeclareAndAssign) Accept(v Visitor) any {
	return v.VisitProcedureStmtDeclareAndAssign(p)
}

// ProcedureStmtCall is a call to another procedure or built-in function.
type ProcedureStmtCall struct {
	baseProcedureStmt
	// Receivers are the variables being assigned. If nil, then the
	// receiver can be ignored.
	Receivers []*string
	Call      ExpressionCall
}

func (p *ProcedureStmtCall) Accept(v Visitor) any {
	return v.VisitProcedureStmtCall(p)
}

type ProcedureStmtForLoop struct {
	baseProcedureStmt
	// Receiver is the variable that is assigned on each iteration.
	Receiver string
	// LoopTerm is what the loop is looping through.
	LoopTerm LoopTerm
	// Body is the body of the loop.
	Body []ProcedureStmt
}

func (p *ProcedureStmtForLoop) Accept(v Visitor) any {
	return v.VisitProcedureStmtForLoop(p)
}

// LoopTerm what the loop is looping through.
type LoopTerm interface {
	Node
	loopTerm()
}

type baseLoopTerm struct {
	parseTypes.Node
	baseProcedureNode
}

func (baseLoopTerm) loopTerm() {}

type LoopTermRange struct {
	baseLoopTerm
	// Start is the start of the range.
	Start ProcedureExpression
	// End is the end of the range.
	End ProcedureExpression
}

func (e *LoopTermRange) Accept(v Visitor) interface{} {
	return v.VisitLoopTermRange(e)
}

type LoopTermCall struct {
	baseLoopTerm
	// Call is the procedure call to loop through.
	// It must return either an array or a table.
	Call ExpressionCall
}

func (e *LoopTermCall) Accept(v Visitor) interface{} {
	return v.VisitLoopTermCall(e)
}

type LoopTermSQL struct {
	baseLoopTerm
	// Statement is the Statement statement to execute.
	Statement *SQLStatement
}

func (e *LoopTermSQL) Accept(v Visitor) interface{} {
	return v.VisitLoopTermSQL(e)
}

type LoopTermVariable struct {
	baseLoopTerm
	// Variable is the variable to loop through.
	// It must be an array.
	Variable string
}

func (e *LoopTermVariable) Accept(v Visitor) interface{} {
	return v.VisitLoopTermVariable(e)
}

type ProcedureStmtIf struct {
	baseProcedureStmt
	// IfThens are the if statements.
	// They are evaluated in order, as
	// IF ... THEN ... ELSEIF ... THEN ...
	IfThens []*IfThen
	// Else is the else statement.
	// It is evaluated if no other if statement
	// is true.
	Else []ProcedureStmt
}

func (p *ProcedureStmtIf) Accept(v Visitor) any {
	return v.VisitProcedureStmtIf(p)
}

type IfThen struct {
	parseTypes.Node
	baseProcedureNode
	If   Expression
	Then []ProcedureStmt
}

func (i *IfThen) Accept(v Visitor) any {
	return v.VisitIfThen(i)
}

type ProcedureStmtSQL struct {
	baseProcedureStmt
	SQL *SQLStatement
}

func (p *ProcedureStmtSQL) Accept(v Visitor) any {
	return v.VisitProcedureStmtSQL(p)
}

type ProcedureStmtBreak struct {
	baseProcedureStmt
}

func (p *ProcedureStmtBreak) Accept(v Visitor) any {
	return v.VisitProcedureStmtBreak(p)
}

type ProcedureStmtReturn struct {
	baseProcedureStmt
	// Values are the values to return.
	// Either values is set or SQL is set, but not both.
	Values []ProcedureExpression
	// SQL is the SQL statement to return.
	// Either values is set or SQL is set, but not both.
	SQL *SQLStatement
}

func (p *ProcedureStmtReturn) Accept(v Visitor) any {
	return v.VisitProcedureStmtReturn(p)
}

type ProcedureStmtReturnNext struct {
	baseProcedureStmt
	// Values are the values to return.
	Values []ProcedureExpression
}

func (p *ProcedureStmtReturnNext) Accept(v Visitor) any {
	return v.VisitProcedureStmtReturnNext(p)
}

// Visitor is an interface for visiting nodes in the parse tree.
type Visitor interface {
	VisitExpressionLiteral(*ExpressionLiteral) any
	VisitExpressionFunctionCall(*ExpressionFunctionCall) any
	VisitExpressionForeignCall(*ExpressionForeignCall) any
	VisitExpressionVariable(*ExpressionVariable) any
	VisitExpressionArrayAccess(*ExpressionArrayAccess) any
	VisitExpressionMakeArray(*ExpressionMakeArray) any
	VisitExpressionFieldAccess(*ExpressionFieldAccess) any
	VisitExpressionParenthesized(*ExpressionParenthesized) any
	VisitExpressionComparison(*ExpressionComparison) any
	VisitExpressionLogical(*ExpressionLogical) any
	VisitExpressionArithmetic(*ExpressionArithmetic) any
	VisitExpressionUnary(*ExpressionUnary) any
	VisitExpressionColumn(*ExpressionColumn) any
	VisitExpressionList(*ExpressionList) any
	VisitExpressionCollate(*ExpressionCollate) any
	VisitExpressionStringComparison(*ExpressionStringComparison) any
	VisitExpressionIs(*ExpressionIs) any
	VisitExpressionIn(*ExpressionIn) any
	VisitExpressionBetween(*ExpressionBetween) any
	VisitExpressionSubquery(*ExpressionSubquery) any
	VisitExpressionCase(*ExpressionCase) any
	VisitCommonTableExpression(*CommonTableExpression) any
	VisitSQLStatement(*SQLStatement) any
	VisitSelectStatement(*SelectStatement) any
	VisitSelectCore(*SelectCore) any
	VisitResultColumnExpression(*ResultColumnExpression) any
	VisitResultColumnWildcard(*ResultColumnWildcard) any
	VisitRelationTable(*RelationTable) any
	VisitRelationSubquery(*RelationSubquery) any
	VisitRelationFunctionCall(*RelationFunctionCall) any
	VisitJoin(*Join) any
	VisitUpdateStatement(*UpdateStatement) any
	VisitUpdateSetClause(*UpdateSetClause) any
	VisitReturningClause(*ReturningClause) any
	VisitDeleteStatement(*DeleteStatement) any
	VisitInsertStatement(*InsertStatement) any
	VisitUpsertClause(*UpsertClause) any
	VisitActionStmtSQL(*ActionStmtSQL) any
	VisitExtensionCallStmt(*ActionStmtExtensionCall) any
	VisitActionCallStmt(*ActionStmtActionCall) any
	VisitProcedureStmtDeclaration(*ProcedureStmtDeclaration) any
	VisitProcedureStmtAssignment(*ProcedureStmtAssignment) any
	VisitProcedureStmtDeclareAndAssign(*ProcedureStmtDeclareAndAssign) any
	VisitProcedureStmtCall(*ProcedureStmtCall) any
	VisitProcedureStmtForLoop(*ProcedureStmtForLoop) any
	VisitLoopTermRange(*LoopTermRange) any
	VisitLoopTermCall(*LoopTermCall) any
	VisitLoopTermSQL(*LoopTermSQL) any
	VisitLoopTermVariable(*LoopTermVariable) any
	VisitProcedureStmtIf(*ProcedureStmtIf) any
	VisitIfThen(*IfThen) any
	VisitProcedureStmtSQL(*ProcedureStmtSQL) any
	VisitProcedureStmtBreak(*ProcedureStmtBreak) any
	VisitProcedureStmtReturn(*ProcedureStmtReturn) any
	VisitProcedureStmtReturnNext(*ProcedureStmtReturnNext) any
}
