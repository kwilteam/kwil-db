package parse

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
)

// this file contains the ASTs for SQL, procedures, and actions.

// Node is a node in the AST.
type Node interface {
	Positionable
	Accept(Visitor) any
}

type GetPositioner interface {
	GetPosition() *Position
	Clear()
}

type Positionable interface {
	GetPositioner
	Set(r antlr.ParserRuleContext)
	SetToken(t antlr.Token)
}

type Typecastable struct {
	TypeCast *types.DataType
}

func (t *Typecastable) Cast(t2 *types.DataType) {
	t.TypeCast = t2
}

func (t *Typecastable) GetTypeCast() *types.DataType {
	return t.TypeCast
}

// Expression is an interface for all expressions.
type Expression interface {
	Node
}

// Assignable is an interface for all expressions that can be assigned to.
type Assignable interface {
	Expression
	assignable()
}

// ExpressionLiteral is a literal expression.
type ExpressionLiteral struct {
	Position
	Typecastable
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
	s, err := literalToString(e.Value)
	if err != nil {
		panic(err.Error() + ": " + fmt.Sprintf("%T", e.Value))
	}
	return s
}

// literalToString formats a literal value to be used in a SQL / DDL statement.
func literalToString(value any) (string, error) {
	str := strings.Builder{}
	switch v := value.(type) {
	case string: // for text type
		str.WriteString("'" + v + "'")
	case int64, int, int32: // for int type
		str.WriteString(fmt.Sprint(v))
	case *types.Uint256:
		str.WriteString(v.String())
	case *decimal.Decimal:
		str.WriteString(v.String())
	case bool: // for bool type
		if v {
			str.WriteString("true")
		}
		str.WriteString("false")
	case []byte:
		str.WriteString("0x" + hex.EncodeToString(v))
	case nil:
		// do nothing
	default:
		return "", fmt.Errorf("unsupported literal type: %T", v)
	}

	return str.String(), nil
}

// ExpressionFunctionCall is a function call expression.
type ExpressionFunctionCall struct {
	Position
	Typecastable
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

// ExpressionWindowFunctionCall is a window function call expression.
type ExpressionWindowFunctionCall struct {
	Position
	FunctionCall *ExpressionFunctionCall
	// Filter is the filter clause.
	// If nil, then there is no filter clause.
	Filter Expression
	// Window is the window function that is being called.
	Window Window
}

func (e *ExpressionWindowFunctionCall) Accept(v Visitor) any {
	return v.VisitExpressionWindowFunctionCall(e)
}

// Window is an interface for all window functions.
// It can either reference an exact window (e.g. OVER (partition by ... order by ...))
// or it can reference a window function name (e.g. OVER my_window).
type Window interface {
	Node
	window()
}

type WindowImpl struct {
	Position
	// PartitionBy is the partition by clause.
	PartitionBy []Expression
	// OrderBy is the order by clause.
	OrderBy []*OrderingTerm
	// In the future, when/if we support frame clauses, we can add it here.
}

func (w *WindowImpl) Accept(v Visitor) any {
	return v.VisitWindowImpl(w)
}

func (w *WindowImpl) window() {}

type WindowReference struct {
	Position
	// Name is the name of the window.
	Name string
}

func (w *WindowReference) Accept(v Visitor) any {
	return v.VisitWindowReference(w)
}

func (w *WindowReference) window() {}

// ExpressionVariable is a variable.
// This can either be $ or @ variables.
type ExpressionVariable struct {
	Position
	Typecastable
	// Name is the naem of the variable,
	// without the $ or @.
	Name string
	// Prefix is the $ or @ prefix.
	Prefix VariablePrefix
}

func (e *ExpressionVariable) Accept(v Visitor) any {
	return v.VisitExpressionVariable(e)
}

// String returns the string representation, as it was passed
// in Kuneiform.
func (e *ExpressionVariable) String() string {
	return e.Name
}

func (e *ExpressionVariable) assignable() {}

type VariablePrefix string

const (
	VariablePrefixDollar VariablePrefix = "$"
	VariablePrefixAt     VariablePrefix = "@"
)

// ExpressionArrayAccess accesses an array value.
type ExpressionArrayAccess struct {
	Position
	Typecastable
	// Array is the array that is being accessed.
	Array Expression
	// Index is the index that is being accessed.
	// Either Index or FromTo is set, but not both.
	Index Expression
	// FromTo is the range that is being accessed.
	// Either Index or FromTo is set, but not both.
	// If FromTo is set, then it is a range access.
	// If both values are set, then it is arr[FROM:TO].
	// If only From is set, then it is arr[FROM:].
	// If only To is set, then it is arr[:TO].
	// If neither are set and index is not set, then it is arr[:].
	FromTo [2]Expression
}

func (e *ExpressionArrayAccess) Accept(v Visitor) any {
	return v.VisitExpressionArrayAccess(e)
}

func (e *ExpressionArrayAccess) assignable() {}

// ExpressionMakeArray makes a new array.
type ExpressionMakeArray struct {
	Position
	Typecastable
	Values []Expression
}

func (e *ExpressionMakeArray) Accept(v Visitor) any {
	return v.VisitExpressionMakeArray(e)
}

// ExpressionFieldAccess accesses a field in a record.
type ExpressionFieldAccess struct {
	Position
	Typecastable
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
	Position
	Typecastable
	// Inner is the inner expression.
	Inner Expression
}

func (e *ExpressionParenthesized) Accept(v Visitor) any {
	return v.VisitExpressionParenthesized(e)
}

// ExpressionComparison is a comparison expression.
type ExpressionComparison struct {
	Position
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
	Position
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
	Position
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
	Position
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
	// Not can be either NOT or !
	UnaryOperatorNot UnaryOperator = "not"
	UnaryOperatorNeg UnaryOperator = "-"
	UnaryOperatorPos UnaryOperator = "+"
)

// ExpressionColumn is a column in a table.
type ExpressionColumn struct {
	Position
	Typecastable
	// Table is the table that the column is in.
	Table string // can be empty
	// Column is the name of the column.
	Column string
}

func (e *ExpressionColumn) String() string {
	if e.Table == "" {
		return e.Column
	}
	return e.Table + "." + e.Column
}

func (e *ExpressionColumn) Accept(v Visitor) any {
	return v.VisitExpressionColumn(e)
}

// ExpressionCollate is an expression with a collation.
type ExpressionCollate struct {
	Position
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
	Position
	// Left is the left side of the comparison.
	Left Expression
	// Right is the right side of the comparison.
	Right Expression
	Not   bool
	// Operator is the operator of the comparison.
	Operator StringComparisonOperator
}

func (e *ExpressionStringComparison) Accept(v Visitor) any {
	return v.VisitExpressionStringComparison(e)
}

type StringComparisonOperator string

const (
	StringComparisonOperatorLike  StringComparisonOperator = "LIKE"
	StringComparisonOperatorILike StringComparisonOperator = "ILIKE"
)

// ExpressionIs is an IS expression.
type ExpressionIs struct {
	Position
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
	Position
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
	Position
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
	Position
	Typecastable
	Not      bool
	Exists   bool
	Subquery *SelectStatement
}

func (e *ExpressionSubquery) Accept(v Visitor) any {
	return v.VisitExpressionSubquery(e)
}

// ExpressionCase is a CASE expression.
type ExpressionCase struct {
	Position
	Case     Expression
	WhenThen [][2]Expression
	Else     Expression
}

func (e *ExpressionCase) Accept(v Visitor) any {
	return v.VisitExpressionCase(e)
}

// CommonTableExpression is a common table expression.
type CommonTableExpression struct {
	Position
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

// SQLStmt is top-level statement, can be any SQL statement.
type SQLStmt interface {
	Node
	StmtType() SQLStatementType
}

// SQLStatement is a DML statement with common table expression.
type SQLStatement struct {
	Position
	CTEs []*CommonTableExpression
	// Recursive is true if the RECURSIVE keyword is present.
	Recursive bool
	// SQL can be an insert, update, delete, or select statement.
	SQL SQLCore
	// raw is the raw SQL string.
	raw *string
}

func (s *SQLStatement) Accept(v Visitor) any {
	return v.VisitSQLStatement(s)
}

func (s *SQLStatement) StmtType() SQLStatementType {
	return s.SQL.StmtType()
}

func (s *SQLStatement) Raw() (string, error) {
	if s.raw == nil {
		return "", fmt.Errorf("raw SQL is not set")
	}

	return *s.raw, nil
}

// SQLCore is a DML statement.
// It can be INSERT, UPDATE, DELETE, SELECT.
type SQLCore interface {
	Node
	StmtType() SQLStatementType
}

type SQLStatementType string

const (
	SQLStatementTypeInsert      SQLStatementType = "insert"
	SQLStatementTypeUpdate      SQLStatementType = "update"
	SQLStatementTypeDelete      SQLStatementType = "delete"
	SQLStatementTypeSelect      SQLStatementType = "select"
	SQLStatementTypeCreateTable SQLStatementType = "create_table"
	SQLStatementTypeAlterTable  SQLStatementType = "alter_table"
	SQLStatementTypeDropTable   SQLStatementType = "drop_table"
	SQLStatementTypeCreateIndex SQLStatementType = "create_index"
	SQLStatementTypeDropIndex   SQLStatementType = "drop_index"
)

// CreateTableStatement is a CREATE TABLE statement.
type CreateTableStatement struct {
	Position

	IfNotExists bool
	Name        string
	Columns     []*Column
	// Constraints contains the non-inline constraints
	Constraints []*OutOfLineConstraint
}

func (c *CreateTableStatement) Accept(v Visitor) any {
	return v.VisitCreateTableStatement(c)
}

func (c *CreateTableStatement) StmtType() SQLStatementType {
	return SQLStatementTypeCreateTable
}

// Column represents a table column.
type Column struct {
	Position

	Name        string
	Type        *types.DataType
	Constraints []InlineConstraint
}

func (c *Column) Accept(v Visitor) any {
	return v.VisitColumn(c)
}

// OutOfLineConstraint is a constraint that is not inline with the column.
// e.g. CREATE TABLE t (a INT, CONSTRAINT c CHECK (a > 0))
type OutOfLineConstraint struct {
	Position
	Name       string // can be empty if the name should be auto-generated
	Constraint OutOfLineConstraintClause
}

// InlineConstraint is a constraint that is inline with the column.
type InlineConstraint interface {
	Positionable
	inlineConstraint()
}

// OutOfLineConstraintClause is a constraint that is not inline with the column.
type OutOfLineConstraintClause interface {
	Positionable
	outOfLineConstraintClause()
	// LocalColumns returns the local columns that the constraint is applied to.
	LocalColumns() []string
}

type PrimaryKeyInlineConstraint struct {
	Position
}

func (c *PrimaryKeyInlineConstraint) inlineConstraint() {}

type PrimaryKeyOutOfLineConstraint struct {
	Position
	Columns []string
}

func (c *PrimaryKeyOutOfLineConstraint) outOfLineConstraintClause() {}

func (c *PrimaryKeyOutOfLineConstraint) LocalColumns() []string { return c.Columns }

type UniqueInlineConstraint struct {
	Position
}

func (c *UniqueInlineConstraint) inlineConstraint() {}

type UniqueOutOfLineConstraint struct {
	Position
	Columns []string
}

func (c *UniqueOutOfLineConstraint) outOfLineConstraintClause() {}

func (c *UniqueOutOfLineConstraint) LocalColumns() []string { return c.Columns }

type DefaultConstraint struct {
	Position
	Value *ExpressionLiteral
}

func (c *DefaultConstraint) inlineConstraint() {}

type NotNullConstraint struct {
	Position
}

func (c *NotNullConstraint) inlineConstraint() {}

type CheckConstraint struct {
	Position
	Expression Expression
}

func (c *CheckConstraint) inlineConstraint() {}

func (c *CheckConstraint) outOfLineConstraintClause() {}

func (c *CheckConstraint) LocalColumns() []string { return nil }

type ForeignKeyReferences struct {
	Position

	RefTable   string
	RefColumns []string
	Actions    []*ForeignKeyAction
}

func (c *ForeignKeyReferences) inlineConstraint() {}

type ForeignKeyOutOfLineConstraint struct {
	Position
	Columns    []string
	References *ForeignKeyReferences
}

func (c *ForeignKeyOutOfLineConstraint) outOfLineConstraintClause() {}

func (c *ForeignKeyOutOfLineConstraint) LocalColumns() []string { return c.Columns }

type IndexType string

const (
	// IndexTypeBTree is the default index, created by using `INDEX`.
	IndexTypeBTree IndexType = "btree"
	// IndexTypeUnique is a unique BTree index, created by using `UNIQUE INDEX`.
	IndexTypeUnique IndexType = "unique"
)

// TableIndex represents table index declaration, both inline and non-inline.
type TableIndex struct {
	Position

	Name    string
	Columns []string
	Type    IndexType
}

func (i *TableIndex) String() string {
	if len(i.Columns) == 0 {
		if i.Type == IndexTypeUnique {
			return "UNIQUE"
		}
		panic("inline index can only be UNIQUE")
	}

	str := strings.Builder{}

	switch i.Type {
	case IndexTypeBTree:
		str.WriteString("INDEX ")
	case IndexTypeUnique:
		str.WriteString("UNIQUE INDEX ")
	default:
		// should not happen
		panic("unknown index type")
	}

	if i.Name != "" {
		str.WriteString(i.Name + " ")
	}

	str.WriteString("(" + strings.Join(i.Columns, ", ") + ")")

	return str.String()
}

func (i *TableIndex) Accept(v Visitor) any {
	return v.VisitTableIndex(i)
}

// ForeignKey is a foreign key in a table.
type ForeignKey struct {
	// ChildKeys are the columns that are referencing another.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "a" is the child key
	ChildKeys []string `json:"child_keys"`

	// ParentKeys are the columns that are being referred to.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "b" is the parent key
	ParentKeys []string `json:"parent_keys"`

	// ParentTable is the table that holds the parent columns.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "tbl2" is the parent table
	ParentTable string `json:"parent_table"`

	// Action refers to what the foreign key should do when the parent is altered.
	// This is NOT the same as a database action.
	// For example, ON DELETE CASCADE is a foreign key action
	Actions []*ForeignKeyAction `json:"actions"`
}

// ForeignKeyActionOn specifies when a foreign key action should occur.
// It can be either "UPDATE" or "DELETE".
type ForeignKeyActionOn string

// ForeignKeyActionOn types
const (
	// ON_UPDATE is used to specify an action should occur when a parent key is updated
	ON_UPDATE ForeignKeyActionOn = "UPDATE"
	// ON_DELETE is used to specify an action should occur when a parent key is deleted
	ON_DELETE ForeignKeyActionOn = "DELETE"
)

// ForeignKeyActionDo specifies what should be done when a foreign key action is triggered.
type ForeignKeyActionDo string

// ForeignKeyActionDo types
const (
	// DO_NO_ACTION does nothing when a parent key is altered
	DO_NO_ACTION ForeignKeyActionDo = "NO ACTION"

	// DO_RESTRICT prevents the parent key from being altered
	DO_RESTRICT ForeignKeyActionDo = "RESTRICT"

	// DO_SET_NULL sets the child key(s) to NULL
	DO_SET_NULL ForeignKeyActionDo = "SET NULL"

	// DO_SET_DEFAULT sets the child key(s) to their default values
	DO_SET_DEFAULT ForeignKeyActionDo = "SET DEFAULT"

	// DO_CASCADE updates the child key(s) or deletes the records (depending on the action type)
	DO_CASCADE ForeignKeyActionDo = "CASCADE"
)

// ForeignKeyAction is used to specify what should occur
// if a parent key is updated or deleted
type ForeignKeyAction struct {
	// On can be either "UPDATE" or "DELETE"
	On ForeignKeyActionOn `json:"on"`

	// Do specifies what a foreign key action should do
	Do ForeignKeyActionDo `json:"do"`
}

type DropBehavior string

const (
	DropBehaviorDefault  DropBehavior = ""
	DropBehaviorCascade  DropBehavior = "CASCADE"
	DropBehaviorRestrict DropBehavior = "RESTRICT"
)

type DropTableStatement struct {
	Position

	Tables   []string
	IfExists bool
	Behavior DropBehavior
}

func (s *DropTableStatement) Accept(v Visitor) any {
	return v.VisitDropTableStatement(s)
}

func (s *DropTableStatement) StmtType() SQLStatementType {
	return SQLStatementTypeDropTable
}

type AlterTableAction interface {
	Node

	alterTableAction()
	ToSQL() string
}

// AlterTableStatement is a ALTER TABLE statement.
type AlterTableStatement struct {
	Position

	Table  string
	Action AlterTableAction
}

func (a *AlterTableStatement) Accept(v Visitor) any {
	return v.VisitAlterTableStatement(a)
}

func (a *AlterTableStatement) StmtType() SQLStatementType {
	return SQLStatementTypeAlterTable
}

// ConstraintType is a constraint in a table.
type ConstraintType interface {
	String() string
	constraint()
}

// SingleColumnConstraintType is a constraint type that can only ever
// be applied to a single column. These are NOT NULL and DEFAULT.
type SingleColumnConstraintType string

func (t SingleColumnConstraintType) String() string {
	return string(t)
}

func (t SingleColumnConstraintType) constraint() {}

const (
	ConstraintTypeNotNull SingleColumnConstraintType = "NOT NULL"
	ConstraintTypeDefault SingleColumnConstraintType = "DEFAULT"
)

// MultiColumnConstraintType is a constraint type that can be applied
// to multiple columns. These are PRIMARY KEY, FOREIGN KEY, UNIQUE, and CHECK.

type MultiColumnConstraintType string

func (t MultiColumnConstraintType) String() string {
	return string(t)
}

func (t MultiColumnConstraintType) constraint() {}

const (
	ConstraintTypeUnique     MultiColumnConstraintType = "UNIQUE"
	ConstraintTypeCheck      MultiColumnConstraintType = "CHECK"
	ConstraintTypeForeignKey MultiColumnConstraintType = "FOREIGN KEY"
	ConstraintTypePrimaryKey MultiColumnConstraintType = "PRIMARY KEY"
)

type SetColumnConstraint struct {
	Position
	// Column is the column that is being altered.
	Column string
	// Type is the type of constraint that is being set.
	Type SingleColumnConstraintType
	// Value is the value of the constraint.
	// It is only set if the type is DEFAULT.
	Value *ExpressionLiteral
}

func (a *SetColumnConstraint) Accept(v Visitor) any {
	return v.VisitSetColumnConstraint(a)
}

func (a *SetColumnConstraint) alterTableAction() {}

func (a *SetColumnConstraint) ToSQL() string {
	str := strings.Builder{}
	str.WriteString("ALTER COLUMN ")
	str.WriteString(a.Column)
	str.WriteString(" SET ")
	switch a.Type {
	case ConstraintTypeNotNull:
		str.WriteString("NOT NULL")
	case ConstraintTypeDefault:
		str.WriteString("DEFAULT ")
		str.WriteString(a.Value.String())
	default:
		panic("unknown constraint type")
	}

	return str.String()
}

type DropColumnConstraint struct {
	Position

	Column string
	Type   SingleColumnConstraintType
}

func (a *DropColumnConstraint) Accept(v Visitor) any {
	return v.VisitDropColumnConstraint(a)
}

func (a *DropColumnConstraint) alterTableAction() {}

func (a *DropColumnConstraint) ToSQL() string {
	str := strings.Builder{}
	str.WriteString("ALTER COLUMN ")
	str.WriteString(a.Column)
	str.WriteString(" DROP ")

	if a.Type != "" {
		switch a.Type {
		case ConstraintTypeNotNull:
			str.WriteString("NOT NULL")
		case ConstraintTypeDefault:
			str.WriteString("DEFAULT")
		default:
			panic("unknown constraint type")
		}
	}

	return str.String()
}

type AddColumn struct {
	Position

	Name string
	Type *types.DataType
}

func (a *AddColumn) Accept(v Visitor) any {
	return v.VisitAddColumn(a)
}

func (a *AddColumn) alterTableAction() {}

func (a *AddColumn) ToSQL() string {
	return "ADD COLUMN " + a.Name + " " + a.Type.String()
}

type DropColumn struct {
	Position

	Name string
}

func (a *DropColumn) Accept(v Visitor) any {
	return v.VisitDropColumn(a)
}

func (a *DropColumn) alterTableAction() {}

func (a *DropColumn) ToSQL() string {
	return "DROP COLUMN " + a.Name
}

type RenameColumn struct {
	Position

	OldName string
	NewName string
}

func (a *RenameColumn) Accept(v Visitor) any {
	return v.VisitRenameColumn(a)
}

func (a *RenameColumn) alterTableAction() {}

func (a *RenameColumn) ToSQL() string {
	return "RENAME COLUMN " + a.OldName + " TO " + a.NewName
}

type RenameTable struct {
	Position

	Name string
}

func (a *RenameTable) Accept(v Visitor) any {
	return v.VisitRenameTable(a)
}

func (a *RenameTable) alterTableAction() {}

func (a *RenameTable) ToSQL() string {
	return "RENAME TO " + a.Name
}

// AddTableConstraint is a constraint that is being added to a table.
// It is used to specify multi-column constraints.
type AddTableConstraint struct {
	Position

	Constraint *OutOfLineConstraint
}

func (a *AddTableConstraint) Accept(v Visitor) any {
	return v.VisitAddTableConstraint(a)
}

func (a *AddTableConstraint) alterTableAction() {}

func (a *AddTableConstraint) ToSQL() string {
	return ""
}

type DropTableConstraint struct {
	Position

	Name string
}

func (a *DropTableConstraint) Accept(v Visitor) any {
	return v.VisitDropTableConstraint(a)
}

func (a *DropTableConstraint) alterTableAction() {}

func (a *DropTableConstraint) ToSQL() string {
	return "DROP CONSTRAINT " + a.Name
}

type CreateIndexStatement struct {
	Position

	IfNotExists bool
	Name        string
	On          string
	Columns     []string
	Type        IndexType
}

func (s *CreateIndexStatement) Accept(v Visitor) any {
	return v.VisitCreateIndexStatement(s)
}

func (s *CreateIndexStatement) StmtType() SQLStatementType {
	return SQLStatementTypeCreateIndex
}

type DropIndexStatement struct {
	Position

	Name       string
	CheckExist bool
}

func (s *DropIndexStatement) Accept(v Visitor) any {
	return v.VisitDropIndexStatement(s)
}

func (s *DropIndexStatement) StmtType() SQLStatementType {
	return SQLStatementTypeDropIndex
}

type GrantOrRevokeStatement struct {
	Position
	// IsGrant is true if the statement is a GRANT statement.
	// If it is false, then it is a REVOKE statement.
	IsGrant bool
	// Privileges are the privileges that are being granted.
	// Either Privileges or Role must be set, but not both.
	Privileges []string
	// Namespace is the namespace that the privileges are being granted on.
	// It can be nil if they are global.
	Namespace *string
	// OnNam
	// Role is the role being granted
	// Either Privileges or Role must be set, but not both.
	GrantRole string
	// ToRole is the role being granted to.
	// Either ToUser or ToRole must be set, but not both.
	ToRole string
	// ToUser is the user being granted to.
	// Either ToUser or ToRole must be set, but not both.
	ToUser string
}

func (g *GrantOrRevokeStatement) Accept(v Visitor) any {
	return v.VisitGrantOrRevokeStatement(g)
}

type CreateRoleStatement struct {
	Position
	// IfNotExists is true if the IF NOT EXISTS clause is present.
	IfNotExists bool
	// Role is the role that is being created or dropped.
	Role string
}

func (c *CreateRoleStatement) Accept(v Visitor) any {
	return v.VisitCreateRoleStatement(c)
}

type DropRoleStatement struct {
	Position
	// IfExists is true if the IF EXISTS clause is present.
	IfExists bool
	// Role is the role that is being created or dropped.
	Role string
}

func (d *DropRoleStatement) Accept(v Visitor) any {
	return v.VisitDropRoleStatement(d)
}

type TransferOwnershipStatement struct {
	Position
	// To is the user that the ownership is being transferred to.
	To string
}

func (t *TransferOwnershipStatement) Accept(v Visitor) any {
	return v.VisitTransferOwnershipStatement(t)
}

// SelectStatement is a SELECT statement.
type SelectStatement struct {
	Position
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
	Position
	Expression Expression
	Order      OrderType
	Nulls      NullOrder
}

func (o *OrderingTerm) Accept(v Visitor) any {
	return v.VisitOrderingTerm(o)
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
	Position
	// Distinct is true if the SELECT statement is a DISTINCT SELECT statement.
	Distinct bool
	Columns  []ResultColumn
	From     Table        // can be nil
	Joins    []*Join      // can be nil
	Where    Expression   // can be nil
	GroupBy  []Expression // can be nil
	Having   Expression   // can be nil
	Windows  []*struct {
		Name   string
		Window *WindowImpl
	} // can be nil
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
	Position

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
	Position
	Table string // can be empty
}

func (r *ResultColumnWildcard) Accept(v Visitor) any {
	return v.VisitResultColumnWildcard(r)
}

func (r *ResultColumnWildcard) ResultColumnType() ResultColumnType {
	return ResultColumnTypeWildcard
}

type Table interface {
	Node
	table()
}

type RelationTable struct {
	Position
	Table string
	Alias string // can be empty
}

func (r *RelationTable) Accept(v Visitor) any {
	return v.VisitRelationTable(r)
}

func (RelationTable) table() {}

type RelationSubquery struct {
	Position
	Subquery *SelectStatement
	// Alias cannot be empty, as our syntax
	// forces it for subqueries.
	Alias string
}

func (r *RelationSubquery) Accept(v Visitor) any {
	return v.VisitRelationSubquery(r)
}

func (RelationSubquery) table() {}

// Join is a join in a SELECT statement.
type Join struct {
	Position
	Type     JoinType
	Relation Table
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
	Position
	Table     string
	Alias     string // can be empty
	SetClause []*UpdateSetClause
	From      Table      // can be nil
	Joins     []*Join    // can be nil
	Where     Expression // can be nil
}

func (u *UpdateStatement) Accept(v Visitor) any {
	return v.VisitUpdateStatement(u)
}

func (u *UpdateStatement) StmtType() SQLStatementType {
	return SQLStatementTypeUpdate
}

type UpdateSetClause struct {
	Position
	Column string
	Value  Expression
}

func (u *UpdateSetClause) Accept(v Visitor) any {
	return v.VisitUpdateSetClause(u)
}

type DeleteStatement struct {
	Position

	Table string
	Alias string     // can be empty
	From  Table      // can be nil
	Joins []*Join    // can be nil
	Where Expression // can be nil
}

func (d *DeleteStatement) StmtType() SQLStatementType {
	return SQLStatementTypeDelete
}

func (d *DeleteStatement) Accept(v Visitor) any {
	return v.VisitDeleteStatement(d)
}

type InsertStatement struct {
	Position
	Table   string
	Alias   string   // can be empty
	Columns []string // can be empty
	// Either Values or Select is set, but not both.
	Values     [][]Expression   // can be empty
	Select     *SelectStatement // can be nil
	OnConflict *OnConflict      // can be nil
}

func (i *InsertStatement) Accept(v Visitor) any {
	return v.VisitInsertStatement(i)
}

func (i *InsertStatement) StmtType() SQLStatementType {
	return SQLStatementTypeInsert
}

type OnConflict struct {
	Position
	ConflictColumns []string           // can be empty
	ConflictWhere   Expression         // can be nil
	DoUpdate        []*UpdateSetClause // if nil, then do nothing
	UpdateWhere     Expression         // can be nil
}

func (u *OnConflict) Accept(v Visitor) any {
	return v.VisitUpsertClause(u)
}

// action ast:

type ActionStmt interface {
	Node
	ActionStmt() ActionStatementTypes
}

type ActionStatementTypes string

const (
	ActionStatementTypeExtensionCall ActionStatementTypes = "extension_call"
	ActionStatementTypeActionCall    ActionStatementTypes = "action_call"
	ActionStatementTypeSQL           ActionStatementTypes = "sql"
)

type ActionStmtSQL struct {
	Position
	SQL *SQLStatement
}

func (a *ActionStmtSQL) Accept(v Visitor) any {
	return v.VisitActionStmtSQL(a)
}

func (a *ActionStmtSQL) ActionStmt() ActionStatementTypes {
	return ActionStatementTypeSQL
}

type ActionStmtExtensionCall struct {
	Position
	Receivers []string
	Extension string
	Method    string
	Args      []Expression
}

func (a *ActionStmtExtensionCall) Accept(v Visitor) any {
	return v.VisitActionStmtExtensionCall(a)
}

func (a *ActionStmtExtensionCall) ActionStmt() ActionStatementTypes {
	return ActionStatementTypeExtensionCall
}

type ActionStmtActionCall struct {
	Position
	Action string
	Args   []Expression
}

func (a *ActionStmtActionCall) Accept(v Visitor) any {
	return v.VisitActionStmtActionCall(a)
}

func (a *ActionStmtActionCall) ActionStmt() ActionStatementTypes {
	return ActionStatementTypeActionCall
}

// procedure ast:

// ProcedureStmt is a statement in a procedure.
// it is the top-level interface for all procedure statements.
type ProcedureStmt interface {
	Node
	procedureStmt()
}

type baseProcedureStmt struct {
	Position
}

func (baseProcedureStmt) procedureStmt() {}

// ProcedureStmtDeclaration is a variable declaration in a procedure.
type ProcedureStmtDeclaration struct {
	baseProcedureStmt
	// Variable is the variable that is being declared.
	Variable *ExpressionVariable
	Type     *types.DataType
}

func (p *ProcedureStmtDeclaration) Accept(v Visitor) any {
	return v.VisitProcedureStmtDeclaration(p)
}

// ProcedureStmtAssign is a variable assignment in a procedure.
// It should only be called on variables that have already been declared.
type ProcedureStmtAssign struct {
	baseProcedureStmt
	// Variable is the variable that is being assigned.
	Variable Assignable
	// Type is the type of the variable.
	// It can be nil if the variable is not being assigned,
	// or if the type should be inferred.
	Type *types.DataType
	// Value is the value that is being assigned.
	Value Expression
}

func (p *ProcedureStmtAssign) Accept(v Visitor) any {
	return v.VisitProcedureStmtAssignment(p)
}

// ProcedureStmtCall is a call to another procedure or built-in function.
type ProcedureStmtCall struct {
	baseProcedureStmt
	// Receivers are the variables being assigned. If nil, then the
	// receiver can be ignored.
	Receivers []*ExpressionVariable
	Call      *ExpressionFunctionCall
}

func (p *ProcedureStmtCall) Accept(v Visitor) any {
	return v.VisitProcedureStmtCall(p)
}

type ProcedureStmtForLoop struct {
	baseProcedureStmt
	// Receiver is the variable that is assigned on each iteration.
	Receiver *ExpressionVariable
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
	Position
}

func (baseLoopTerm) loopTerm() {}

type LoopTermRange struct {
	baseLoopTerm
	// Start is the start of the range.
	Start Expression
	// End is the end of the range.
	End Expression
}

func (e *LoopTermRange) Accept(v Visitor) interface{} {
	return v.VisitLoopTermRange(e)
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
	Variable *ExpressionVariable
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
	Position
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
	Values []Expression
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
	Values []Expression
}

func (p *ProcedureStmtReturnNext) Accept(v Visitor) any {
	return v.VisitProcedureStmtReturnNext(p)
}

/*
	There are three types of visitors, all which compose on each other:
	- Visitor: top-level visitor capable of visiting actions, procedures, and SQL.
	- ProcedureVisitor: a visitor capable of only visiting procedures and SQL. It must include
	SQL because procedures themselves rely on SQL/
	- SQLVisitor: a visitor capable of only visiting SQL.
*/

// Visitor is an interface for visiting nodes in the parse tree.
type Visitor interface {
	ProcedureVisitor
	DDLVisitor
	VisitActionStmtSQL(*ActionStmtSQL) any
	VisitActionStmtExtensionCall(*ActionStmtExtensionCall) any
	VisitActionStmtActionCall(*ActionStmtActionCall) any
}

// DDLVisitor includes visit methods only needed to analyze DDL statements.
type DDLVisitor interface {
	// DDL
	VisitCreateTableStatement(*CreateTableStatement) any
	VisitAlterTableStatement(*AlterTableStatement) any
	VisitDropTableStatement(*DropTableStatement) any
	VisitCreateIndexStatement(*CreateIndexStatement) any
	VisitDropIndexStatement(*DropIndexStatement) any
	VisitGrantOrRevokeStatement(*GrantOrRevokeStatement) any
	VisitSetColumnConstraint(*SetColumnConstraint) any
	VisitDropColumnConstraint(*DropColumnConstraint) any
	VisitAddColumn(*AddColumn) any
	VisitDropColumn(*DropColumn) any
	VisitRenameColumn(*RenameColumn) any
	VisitRenameTable(*RenameTable) any
	VisitAddTableConstraint(*AddTableConstraint) any
	VisitDropTableConstraint(*DropTableConstraint) any
	VisitTableIndex(*TableIndex) any
	VisitColumn(*Column) any
	VisitCreateRoleStatement(*CreateRoleStatement) any
	VisitDropRoleStatement(*DropRoleStatement) any
	VisitTransferOwnershipStatement(*TransferOwnershipStatement) any
}

// ProcedureVisitor includes visit methods only needed to analyze procedures.
// It does not need visit methods for structs that are for the schema or actions
type ProcedureVisitor interface {
	SQLVisitor
	VisitProcedureStmtDeclaration(*ProcedureStmtDeclaration) any
	VisitProcedureStmtAssignment(*ProcedureStmtAssign) any
	VisitProcedureStmtCall(*ProcedureStmtCall) any
	VisitProcedureStmtForLoop(*ProcedureStmtForLoop) any
	VisitLoopTermRange(*LoopTermRange) any
	VisitLoopTermSQL(*LoopTermSQL) any
	VisitLoopTermVariable(*LoopTermVariable) any
	VisitProcedureStmtIf(*ProcedureStmtIf) any
	VisitIfThen(*IfThen) any
	VisitProcedureStmtSQL(*ProcedureStmtSQL) any
	VisitProcedureStmtBreak(*ProcedureStmtBreak) any
	VisitProcedureStmtReturn(*ProcedureStmtReturn) any
	VisitProcedureStmtReturnNext(*ProcedureStmtReturnNext) any
}

// SQLVisitor is a visitor that only has methods for SQL nodes.
type SQLVisitor interface {
	VisitExpressionLiteral(*ExpressionLiteral) any
	VisitExpressionFunctionCall(*ExpressionFunctionCall) any
	VisitExpressionWindowFunctionCall(*ExpressionWindowFunctionCall) any
	VisitWindowImpl(*WindowImpl) any
	VisitWindowReference(*WindowReference) any
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
	VisitJoin(*Join) any
	VisitUpdateStatement(*UpdateStatement) any
	VisitUpdateSetClause(*UpdateSetClause) any
	VisitDeleteStatement(*DeleteStatement) any
	VisitInsertStatement(*InsertStatement) any
	VisitUpsertClause(*OnConflict) any
	VisitOrderingTerm(*OrderingTerm) any
}

// UnimplementedSqlVisitor is meant to be used when an implementing visitor only intends
// to implement the SQLVisitor interface. It will implement the full visitor interface,
// but will panic if any of the methods are called. It does not implement the SQLVisitor
// interface, so it alone cannot be used as a visitor.
type UnimplementedSqlVisitor struct {
	UnimplementedProcedureVisitor
}

func (s *UnimplementedSqlVisitor) VisitActionStmtSQL(p0 *ActionStmtSQL) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

func (s *UnimplementedSqlVisitor) VisitActionStmtExtensionCall(p0 *ActionStmtExtensionCall) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

func (s *UnimplementedSqlVisitor) VisitActionStmtActionCall(p0 *ActionStmtActionCall) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

// UnimplementedProcedureVisitor is meant to be used when an implementing visitor only intends
// to implement the ProcedureVisitor interface. It will implement the full visitor interface,
// but will panic if any of the methods are called. It does not implement the SQLVisitor or
// ProcedureVisitor interfaces, so it alone cannot be used as a visitor.
type UnimplementedProcedureVisitor struct{}

func (s *UnimplementedProcedureVisitor) VisitProcedureStmtDeclaration(p0 *ProcedureStmtDeclaration) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

func (s *UnimplementedProcedureVisitor) VisitProcedureStmtAssignment(p0 *ProcedureStmtAssign) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

func (s *UnimplementedProcedureVisitor) VisitProcedureStmtCall(p0 *ProcedureStmtCall) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

func (s *UnimplementedProcedureVisitor) VisitProcedureStmtForLoop(p0 *ProcedureStmtForLoop) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

func (s *UnimplementedProcedureVisitor) VisitLoopTermRange(p0 *LoopTermRange) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

func (s *UnimplementedProcedureVisitor) VisitLoopTermSQL(p0 *LoopTermSQL) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

func (s *UnimplementedProcedureVisitor) VisitLoopTermVariable(p0 *LoopTermVariable) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

func (s *UnimplementedProcedureVisitor) VisitProcedureStmtIf(p0 *ProcedureStmtIf) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

func (s *UnimplementedProcedureVisitor) VisitIfThen(p0 *IfThen) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

func (s *UnimplementedProcedureVisitor) VisitProcedureStmtSQL(p0 *ProcedureStmtSQL) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

func (s *UnimplementedProcedureVisitor) VisitProcedureStmtBreak(p0 *ProcedureStmtBreak) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

func (s *UnimplementedProcedureVisitor) VisitProcedureStmtReturn(p0 *ProcedureStmtReturn) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

func (s *UnimplementedProcedureVisitor) VisitProcedureStmtReturnNext(p0 *ProcedureStmtReturnNext) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", s))
}

type UnimplementedDDLVisitor struct{}

func (u *UnimplementedDDLVisitor) VisitCreateTableStatement(p0 *CreateTableStatement) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitAlterTableStatement(p0 *AlterTableStatement) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitDropTableStatement(p0 *DropTableStatement) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitCreateIndexStatement(p0 *CreateIndexStatement) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitDropIndexStatement(p0 *DropIndexStatement) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitGrantOrRevokeStatement(p0 *GrantOrRevokeStatement) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitSetColumnConstraint(p0 *SetColumnConstraint) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitDropColumnConstraint(p0 *DropColumnConstraint) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitAddColumn(p0 *AddColumn) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitDropColumn(p0 *DropColumn) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitRenameColumn(p0 *RenameColumn) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitRenameTable(p0 *RenameTable) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitAddTableConstraint(p0 *AddTableConstraint) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitDropTableConstraint(p0 *DropTableConstraint) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitTableIndex(p0 *TableIndex) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitColumn(p0 *Column) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitCreateRoleStatement(p0 *CreateRoleStatement) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitDropRoleStatement(p0 *DropRoleStatement) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}

func (u *UnimplementedDDLVisitor) VisitTransferOwnershipStatement(p0 *TransferOwnershipStatement) any {
	panic(fmt.Sprintf("api misuse: cannot visit %T in constrained visitor", u))
}
