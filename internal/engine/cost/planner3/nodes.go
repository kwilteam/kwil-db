package planner3

import (
	"bytes"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/parse"
)

type LogicalPlan interface {
	Children() []LogicalPlan
	Schema() *Schema
}

type Noop struct{}

func (n *Noop) Children() []LogicalPlan {
	return []LogicalPlan{}
}

func (n *Noop) Schema() *Schema {
	return &Schema{}
}

// TableScan represents a scan of a physical table.
type TableScan struct {
	TableName   string
	TableSchema *Schema
}

func (t *TableScan) Children() []LogicalPlan {
	return []LogicalPlan{}
}

func (t *TableScan) Schema() *Schema {
	return t.TableSchema
}

// FunctionScan represents a scan of a function.
// It can call either a local procedure or foreign procedure
// that returns a table.
type FunctionScan struct {
	FunctionName string
	// Args are the base arguments to the procedure.
	Args []LogicalExpr
	// ContextualArgs are the arguments that are passed in if
	// the procedure is a foreign procedure.
	ContextualArgs []LogicalExpr
	// IsForeign is true if the function is a foreign procedure.
	IsForeign bool
	// FunctionSchema is the schema of the function.
	FunctionSchema *Schema
}

func (f *FunctionScan) Children() []LogicalPlan {
	return []LogicalPlan{}
}

func (f *FunctionScan) Schema() *Schema {
	return f.FunctionSchema
}

type Scan struct {
	Child LogicalPlan
	// Alias will always be set.
	// If the scan is a table scan and no alias was specified,
	// the alias will be the table name.
	// All other scan types (functions and subqueries) require an alias.
	Alias string
}

func (s *Scan) Children() []LogicalPlan {
	return []LogicalPlan{s.Child}
}

func (s *Scan) Schema() *Schema {
	return s.Child.Schema()
}

type Project struct {
	Expressions []LogicalExpr
	Child       LogicalPlan
}

func (p *Project) Children() []LogicalPlan {
	return []LogicalPlan{p.Child}
}

func (p *Project) Schema() *Schema {
	panic("TODO: implement")
	// columns := make([]Column, len(p.Expressions))
	// for i, expr := range p.Expressions {
	// 	columns[i] = Column{
	// 		Name:     expr.Name(),
	// 		DataType: expr.DataType(),
	// 	}
	// }
	// return &Schema{Columns: columns}
}

type Filter struct {
	Condition LogicalExpr
	Child     LogicalPlan
}

func (f *Filter) Children() []LogicalPlan {
	return []LogicalPlan{f.Child}
}

func (f *Filter) Schema() *Schema {
	return f.Child.Schema()
}

type Join struct {
	Left      LogicalPlan
	Right     LogicalPlan
	JoinType  JoinType
	Condition LogicalExpr
}

func (j *Join) Children() []LogicalPlan {
	return []LogicalPlan{j.Left, j.Right}
}

func (j *Join) Schema() *Schema {
	leftSchema := j.Left.Schema()
	rightSchema := j.Right.Schema()
	columns := append(leftSchema.Columns, rightSchema.Columns...)
	return &Schema{Columns: columns}
}

type Sort struct {
	SortExpressions []*SortExpression
	Child           LogicalPlan
}

type SortExpression struct {
	Expr      LogicalExpr
	Ascending bool
	NullsLast bool
}

func (s *Sort) Children() []LogicalPlan {
	return []LogicalPlan{s.Child}
}

func (s *Sort) Schema() *Schema {
	return s.Child.Schema()
}

type Limit struct {
	Child  LogicalPlan
	Limit  LogicalExpr
	Offset LogicalExpr
}

func (l *Limit) Children() []LogicalPlan {
	return []LogicalPlan{l.Child}
}

func (l *Limit) Schema() *Schema {
	return l.Child.Schema()
}

type Distinct struct {
	Child LogicalPlan
}

func (d *Distinct) Children() []LogicalPlan {
	return []LogicalPlan{d.Child}
}

func (d *Distinct) Schema() *Schema {
	return d.Child.Schema()
}

type SetOperation struct {
	Left   LogicalPlan
	Right  LogicalPlan
	OpType SetOperationType
}

// SetOperation
func (s *SetOperation) Children() []LogicalPlan {
	return []LogicalPlan{s.Left, s.Right}
}

func (s *SetOperation) Schema() *Schema {
	// Assuming set operations require compatible schemas
	return s.Left.Schema()
}

type Aggregate struct {
	// GroupingExpressions are the expressions used
	// in the GROUP BY clause.
	GroupingExpressions []LogicalExpr
	// AggregateExpressions are the expressions used
	// in the SELECT clause (e.g. SUM(x), COUNT(y)).
	AggregateExpressions []LogicalExpr
	// Child is the input to the aggregation
	// (e.g. a Project node).
	Child LogicalPlan
}

func (a *Aggregate) Children() []LogicalPlan {
	return []LogicalPlan{a.Child}
}

func (a *Aggregate) Schema() *Schema {
	panic("TODO: implement")

	// columns := make([]Column, len(a.GroupingExpressions)+len(a.AggregateExpressions))

	// for i, expr := range a.GroupingExpressions {
	// 	columns[i] = Column{
	// 		Name:     expr.Name(),
	// 		DataType: expr.DataType(),
	// 	}
	// }

	// offset := len(a.GroupingExpressions)
	// for i, aggExpr := range a.AggregateExpressions {
	// 	columns[offset+i] = Column{
	// 		Name:     aggExpr.Name(),
	// 		DataType: aggExpr.DataType(),
	// 	}
	// }

	// return &Schema{Columns: columns}
}

type Having struct {
	Condition LogicalExpr
	Child     LogicalPlan
}

func (h *Having) Children() []LogicalPlan {
	return []LogicalPlan{h.Child}
}

func (h *Having) Schema() *Schema {
	return h.Child.Schema()
}

type JoinType int

const (
	InnerJoin JoinType = iota
	LeftOuterJoin
	RightOuterJoin
	FullOuterJoin
)

type SetOperationType int

const (
	Union SetOperationType = iota
	UnionAll
	Intersect
	Except
)

/*
	###########################
	#                         #
	#   	Expressions		  #
	#                         #
	###########################
*/

type LogicalExpr interface {
	// Name returns the name of the expression.
	// This can be empty, and is generally only set for ColumnRef
	// or aliased expressions.
	Name() string
	// IsAggregate returns true if the expression is an aggregate.
	IsAggregate() bool // TODO: I think we can remove this
	// Project returns the columns that are used by the expression, as well as
	// any aggregation expressions that are used.
	// If the projection results in an ambiguous column name or an unknown column,
	// an error is returned. The returned columns will be fully qualified.
	Project(*Schema) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error)
	// Equal compares two expressions for equality.
	Equal(LogicalExpr) bool
	// String returns a string representation of the expression.
	String() string
}

// baseExpr is a helper struct that implements the default behavior for an Expression.
type baseExpr struct{}

func (b *baseExpr) Name() string { return "" }

func (b *baseExpr) IsAggregate() bool { return false }

func (b *baseExpr) Project(s *Schema) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return nil, nil, nil
}

func projectMany(s *Schema, exprs ...LogicalExpr) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	var columns []*ProjectedColumn
	var exprsToProject []LogicalExpr
	for _, expr := range exprs {
		cols, aggExprs, err := expr.Project(s)
		if err != nil {
			return nil, nil, err
		}
		columns = append(columns, cols...)
		exprsToProject = append(exprsToProject, aggExprs...)
	}
	return columns, exprsToProject, nil
}

// Literal value
type Literal struct {
	baseExpr
	Value interface{}
	Type  *types.DataType
}

func (l *Literal) Equal(other LogicalExpr) bool {
	otherLit, ok := other.(*Literal)
	if !ok {
		return false
	}

	if !l.Type.EqualsStrict(otherLit.Type) {
		return false
	}

	// can be of type int64, string, bool, nil,
	// *decimal.Decimal, *types.UUID, or *types.Uint256
	switch v := l.Value.(type) {
	case int64:
		otherV, ok := otherLit.Value.(int64)
		if !ok {
			return false
		}

		return v == otherV
	case string:
		otherV, ok := otherLit.Value.(string)
		if !ok {
			return false
		}

		return v == otherV
	case bool:
		otherV, ok := otherLit.Value.(bool)
		if !ok {
			return false
		}

		return v == otherV
	case nil:
		return otherLit.Value == nil
	case *types.UUID:
		otherV, ok := otherLit.Value.(*types.UUID)
		if !ok {
			return false
		}

		return bytes.Equal(v[:], otherV[:])
	case *types.Uint256:
		otherV, ok := otherLit.Value.(*types.Uint256)
		if !ok {
			return false
		}

		return v == otherV
	case *decimal.Decimal:
		otherV, ok := otherLit.Value.(*decimal.Decimal)
		if !ok {
			return false
		}

		// we already know they are of the same type, so we can
		// just make them strings and compare them
		return v.String() == otherV.String()
	default:
		panic(fmt.Sprintf("unhandled literal type %T", l.Value))
	}
}

func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.Value)
}

// Variable reference
type Variable struct {
	baseExpr
	// name is something like $id, @caller, etc.
	VarName string
}

func (v *Variable) Equal(other LogicalExpr) bool {
	otherVar, ok := other.(*Variable)
	if !ok {
		return false
	}

	return v.VarName == otherVar.VarName
}

func (v *Variable) String() string {
	return v.VarName
}

// Column reference
type ColumnRef struct {
	baseExpr
	Parent     string // Parent relation name, can be empty
	ColumnName string
}

func (c *ColumnRef) Name() string {
	return c.ColumnName
}

func (c *ColumnRef) Project(s *Schema) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	column, err := s.Search(c.Parent, c.ColumnName)
	if err != nil {
		return nil, nil, err
	}
	return []*ProjectedColumn{column}, nil, nil
}

func (c *ColumnRef) Equal(other LogicalExpr) bool {
	otherCol, ok := other.(*ColumnRef)
	if !ok {
		return false
	}

	// if we passed access to the schema here, we could qualify the columns,
	// but for now we'll just compare the names
	return c.Parent == otherCol.Parent && c.ColumnName == otherCol.ColumnName
}

func (c *ColumnRef) String() string {
	if c.Parent != "" {
		return fmt.Sprintf("%s.%s", c.Parent, c.ColumnName)
	}
	return c.ColumnName
}

// Function call
type FunctionCall struct {
	FunctionName string
	Args         []LogicalExpr
	Star         bool
	Distinct     bool
}

func (f *FunctionCall) Name() string {
	return f.FunctionName
}

func (f *FunctionCall) IsAggregate() bool {
	fn, ok := parse.Functions[f.FunctionName]
	if !ok {
		panic(fmt.Errorf("function %s not found", f.FunctionName))
	}

	return fn.IsAggregate
}

func (f *FunctionCall) Equal(other LogicalExpr) bool {
	otherFn, ok := other.(*FunctionCall)
	if !ok {
		return false
	}

	if f.FunctionName != otherFn.FunctionName {
		return false
	}

	if len(f.Args) != len(otherFn.Args) {
		return false
	}

	for i := range f.Args {
		if !f.Args[i].Equal(otherFn.Args[i]) {
			return false
		}
	}

	return true
}

func (f *FunctionCall) String() string {
	var buf bytes.Buffer
	buf.WriteString(f.FunctionName)
	buf.WriteString("(")
	for i, arg := range f.Args {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(arg.String())
	}
	buf.WriteString(")")
	return buf.String()
}

func (f *FunctionCall) Project(s *Schema) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	args, aggs, err := projectMany(s, f.Args...)
	if err != nil {
		return nil, nil, err
	}

	// if the function is an aggregate, mark the arguments as aggregated.
	// otherwise, return the arguments as-is.
	if !f.IsAggregate() {
		return args, aggs, nil
	}

	for _, arg := range args {
		arg.Aggregated = true
	}

	// append this function to the other known aggregates
	return args, append(aggs, f), nil
}

type ArithmeticOp struct {
	baseExpr
	Left  LogicalExpr
	Right LogicalExpr
	Op    ArithmeticOperator
}

type ArithmeticOperator int

const (
	Add ArithmeticOperator = iota
	Subtract
	Multiply
	Divide
	Modulo
)

func (a *ArithmeticOp) IsAggregate() bool {
	return a.Left.IsAggregate() || a.Right.IsAggregate()
}

func (a *ArithmeticOp) Project(s *Schema) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return projectMany(s, a.Left, a.Right)
}

func (a *ArithmeticOp) Equal(other LogicalExpr) bool {
	otherArith, ok := other.(*ArithmeticOp)
	if !ok {
		return false
	}

	return a.Op == otherArith.Op && a.Left.Equal(otherArith.Left) && a.Right.Equal(otherArith.Right)
}

func (a *ArithmeticOp) String() string {
	var op string
	switch a.Op {
	case Add:
		op = "+"
	case Subtract:
		op = "-"
	case Multiply:
		op = "*"
	case Divide:
		op = "/"
	case Modulo:
		op = "%"
	}

	return fmt.Sprintf("(%s %s %s)", a.Left.String(), op, a.Right.String())
}

type ComparisonOp struct {
	baseExpr
	Left  LogicalExpr
	Right LogicalExpr
	Op    ComparisonOperator
}

type ComparisonOperator int

const (
	Equal ComparisonOperator = iota
	NotEqual
	LessThan
	LessThanOrEqual
	GreaterThan
	GreaterThanOrEqual
)

func (c *ComparisonOp) IsAggregate() bool {
	return c.Left.IsAggregate() || c.Right.IsAggregate()
}

func (c *ComparisonOp) Project(s *Schema) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return projectMany(s, c.Left, c.Right)
}

func (c *ComparisonOp) Equal(other LogicalExpr) bool {
	otherComp, ok := other.(*ComparisonOp)
	if !ok {
		return false
	}

	return c.Op == otherComp.Op && c.Left.Equal(otherComp.Left) && c.Right.Equal(otherComp.Right)
}

func (c *ComparisonOp) String() string {
	var op string
	switch c.Op {
	case Equal:
		op = "="
	case NotEqual:
		op = "!="
	case LessThan:
		op = "<"
	case LessThanOrEqual:
		op = "<="
	case GreaterThan:
		op = ">"
	case GreaterThanOrEqual:
		op = ">="
	}

	return fmt.Sprintf("(%s %s %s)", c.Left.String(), op, c.Right.String())
}

type LogicalOp struct {
	baseExpr
	Left  LogicalExpr
	Right LogicalExpr
	Op    LogicalOperator
}

type LogicalOperator int

const (
	And LogicalOperator = iota
	Or
)

func (l *LogicalOp) IsAggregate() bool {
	return l.Left.IsAggregate() || l.Right.IsAggregate()
}

func (l *LogicalOp) Project(s *Schema) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return projectMany(s, l.Left, l.Right)
}

func (l *LogicalOp) Equal(other LogicalExpr) bool {
	otherLog, ok := other.(*LogicalOp)
	if !ok {
		return false
	}

	return l.Op == otherLog.Op && l.Left.Equal(otherLog.Left) && l.Right.Equal(otherLog.Right)
}

func (l *LogicalOp) String() string {
	var op string
	switch l.Op {
	case And:
		op = "AND"
	case Or:
		op = "OR"
	}

	return fmt.Sprintf("(%s %s %s)", l.Left.String(), op, l.Right.String())
}

type UnaryOp struct {
	baseExpr
	Expr LogicalExpr
	Op   UnaryOperator
}

type UnaryOperator int

const (
	Negate UnaryOperator = iota
	Not
	Positive
)

func (u *UnaryOp) IsAggregate() bool {
	return u.Expr.IsAggregate()
}

func (u *UnaryOp) Project(s *Schema) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return u.Expr.Project(s)
}

func (u *UnaryOp) Equal(other LogicalExpr) bool {
	otherUnary, ok := other.(*UnaryOp)
	if !ok {
		return false
	}

	return u.Op == otherUnary.Op && u.Expr.Equal(otherUnary.Expr)
}

func (u *UnaryOp) String() string {
	var op string
	switch u.Op {
	case Negate:
		op = "-"
	case Not:
		op = "NOT"
	case Positive:
		op = "+"
	}

	return fmt.Sprintf("%s%s", op, u.Expr.String())
}

type TypeCast struct {
	baseExpr
	Expr LogicalExpr
	Type *types.DataType
}

func (t *TypeCast) IsAggregate() bool {
	return t.Expr.IsAggregate()
}

func (t *TypeCast) Name() string {
	return t.Expr.Name()
}

func (t *TypeCast) Project(s *Schema) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return t.Expr.Project(s)
}

func (t *TypeCast) Equal(other LogicalExpr) bool {
	otherCast, ok := other.(*TypeCast)
	if !ok {
		return false
	}

	return t.Type.EqualsStrict(otherCast.Type) && t.Expr.Equal(otherCast.Expr)
}

func (t *TypeCast) String() string {
	return fmt.Sprintf("(%s::%s)", t.Expr.String(), t.Type.Name)
}

type Alias struct {
	baseExpr
	Expr  LogicalExpr
	Alias string
}

func (a *Alias) Name() string {
	return a.Alias
}

func (a *Alias) IsAggregate() bool {
	return a.Expr.IsAggregate()
}

func (a *Alias) Project(s *Schema) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return a.Expr.Project(s)
}

func (a *Alias) Equal(other LogicalExpr) bool {
	otherAlias, ok := other.(*Alias)
	if !ok {
		return false
	}

	return a.Alias == otherAlias.Alias && a.Expr.Equal(otherAlias.Expr)
}

func (a *Alias) String() string {
	return fmt.Sprintf("%s AS %s", a.Expr.String(), a.Alias)
}
