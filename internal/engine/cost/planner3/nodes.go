package planner3

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
)

/*
	This file contains the nodes necessary to construct logical plans. These are made up of
	- LogicalPlan: a node in the logical plan tree, which returns a relation
	- LogicalExpr: an expression that can be used in a logical plan, which returns a data type

	The nodes are purely representative of a transformed query, and natively do not have any
	understanding of the underlying data or schema. When context of the underlying data is needed
	(e.g. to determine the data type of a column reference), a SchemaContext is passed in.
*/

func Format(plan LogicalPlan, indent int) string {
	var msg strings.Builder
	for i := 0; i < indent; i++ {
		msg.WriteString(" ")
	}
	msg.WriteString(plan.String())
	msg.WriteString("\n")
	for _, child := range plan.Children() {
		msg.WriteString(Format(child, indent+2))
	}
	return msg.String()
}

type LogicalPlan interface {
	Accepter
	fmt.Stringer
	// TODO: I dont know if we need Children().
	Children() []LogicalPlan
	Relation(*SchemaContext) *Relation
}

type Noop struct{}

func (n *Noop) Children() []LogicalPlan {
	return []LogicalPlan{}
}

func (f *Noop) Accept(v Visitor) any {
	return v.VisitNoop(f)
}

func (n *Noop) Relation(ctx *SchemaContext) *Relation {
	return &Relation{}
}

func (n *Noop) String() string {
	return "NOOP"
}

// TableScan represents a scan of a physical table or a CTE.
type TableScan struct {
	TableName string
}

func (t *TableScan) Children() []LogicalPlan {
	return []LogicalPlan{}
}

func (f *TableScan) Accept(v Visitor) any {
	return v.VisitTableScan(f)
}

func (t *TableScan) Relation(ctx *SchemaContext) *Relation {
	tbl, found := ctx.Schema.FindTable(t.TableName)
	if found {
		// if found, we convert the table to a relation
		return relationFromTable(tbl)
	}

	// otherwise, check CTE
	cteRel, ok := ctx.CTEs[t.TableName]
	if !ok {
		panic(fmt.Errorf(`table "%s" not found`, t.TableName))
	}

	return cteRel
}

func (t *TableScan) String() string {
	return fmt.Sprintf("SCAN TABLE %s", t.TableName)
}

// ProcedureScan represents a scan of a function.
// It can call either a local procedure or foreign procedure
// that returns a table.
type ProcedureScan struct {
	// ProcedureName is the name of the procedure being targeted.
	ProcedureName string
	// Args are the base arguments to the procedure.
	Args []LogicalExpr
	// ContextualArgs are the arguments that are passed in if
	// the procedure is a foreign procedure.
	ContextualArgs []LogicalExpr
	// IsForeign is true if the function is a foreign procedure.
	IsForeign bool
}

func (f *ProcedureScan) Children() []LogicalPlan {
	return []LogicalPlan{}
}

func (f *ProcedureScan) Accept(v Visitor) any {
	return v.VisitProcedureScan(f)
}

func (f *ProcedureScan) Relation(ctx *SchemaContext) *Relation {
	var procReturns *types.ProcedureReturn

	if f.IsForeign {
		fp, ok := ctx.Schema.FindForeignProcedure(f.ProcedureName)
		if !ok {
			panic(fmt.Sprintf(`foreign procedure "%s" not found`, f.ProcedureName))
		}
		procReturns = fp.Returns
	} else {
		proc, ok := ctx.Schema.FindProcedure(f.ProcedureName)
		if !ok {
			panic(fmt.Sprintf(`procedure "%s" not found`, f.ProcedureName))
		}
		procReturns = proc.Returns
	}

	// these should get caught during construction of the logical plan
	if procReturns == nil {
		panic(fmt.Sprintf(`procedure "%s" does not return a table`, f.ProcedureName))
	}
	if !procReturns.IsTable {
		panic(fmt.Sprintf(`procedure "%s" does not return a table`, f.ProcedureName))
	}

	var cols []*Column
	for _, field := range procReturns.Fields {
		cols = append(cols, &Column{
			// the Parent will get set by the ScanAlias
			Name:     field.Name,
			DataType: field.Type.Copy(),
			Nullable: true,
			// not unique or indexed
		})
	}

	return &Relation{Columns: cols}
}

func (f *ProcedureScan) String() string {
	str := strings.Builder{}
	str.WriteString("SCAN ")
	if f.IsForeign {
		str.WriteString("FOREIGN ")
	}
	str.WriteString("PROCEDURE ")
	str.WriteString(f.ProcedureName)
	str.WriteString("(")
	for i, arg := range f.Args {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(arg.String())
	}
	str.WriteString(")")
	return str.String()
}

type ScanAlias struct {
	Child LogicalPlan
	// Alias will always be set.
	// If the scan is a table scan and no alias was specified,
	// the alias will be the table name.
	// All other scan types (functions and subqueries) require an alias.
	Alias string
}

func (s *ScanAlias) Children() []LogicalPlan {
	return []LogicalPlan{s.Child}
}

func (f *ScanAlias) Accept(v Visitor) any {
	return v.VisitScanAlias(f)
}

func (s *ScanAlias) Relation(ctx *SchemaContext) *Relation {
	rel := s.Child.Relation(ctx)

	// apply the alias to all scanned columns
	for _, col := range rel.Columns {
		col.Parent = s.Alias
	}

	return rel
}

func (s *ScanAlias) String() string {
	return fmt.Sprintf("ALIAS %s", s.Alias)
}

type Project struct {
	Expressions []LogicalExpr
	Child       LogicalPlan
}

func (p *Project) Children() []LogicalPlan {
	return []LogicalPlan{p.Child}
}

func (f *Project) Accept(v Visitor) any {
	return v.VisitProject(f)
}

func (p *Project) Relation(ctx *SchemaContext) *Relation {
	rel := p.Child.Relation(ctx)

	ctx2 := ctx.Join(rel)

	columns := make([]*Column, len(p.Expressions))
	for i, expr := range p.Expressions {
		dt, err := expr.DataType(ctx2).Scalar()
		if err != nil {
			panic(err)
		}

		columns[i] = &Column{
			Name:     expr.Name(),
			DataType: dt,
		}
	}
	return &Relation{Columns: columns}
}

func (p *Project) String() string {
	str := strings.Builder{}
	str.WriteString("PROJECT ")

	for i, expr := range p.Expressions {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(expr.String())
	}

	return str.String()
}

type Filter struct {
	Condition LogicalExpr
	Child     LogicalPlan
}

func (f *Filter) Children() []LogicalPlan {
	return []LogicalPlan{f.Child}
}

func (f *Filter) Accept(v Visitor) any {
	return v.VisitFilter(f)
}

func (f *Filter) Relation(ctx *SchemaContext) *Relation {
	return f.Child.Relation(ctx)
}

func (f *Filter) String() string {
	return fmt.Sprintf("FILTER %s", f.Condition.String())
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

func (f *Join) Accept(v Visitor) any {
	return v.VisitJoin(f)
}

func (j *Join) Relation(ctx *SchemaContext) *Relation {
	leftRel := j.Left.Relation(ctx)
	rightRel := j.Right.Relation(ctx)
	columns := append(leftRel.Columns, rightRel.Columns...)
	return &Relation{Columns: columns}
}

func (j *Join) String() string {
	str := strings.Builder{}
	str.WriteString(j.JoinType.String())
	str.WriteString(" JOIN: left: ")
	str.WriteString(j.Left.String())
	str.WriteString(", right: ")
	str.WriteString(j.Right.String())
	str.WriteString(", on: ")
	str.WriteString(j.Condition.String())
	return str.String()

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

func (s *Sort) String() string {
	str := strings.Builder{}
	str.WriteString("SORT BY ")
	for i, sortExpr := range s.SortExpressions {
		if i > 0 {
			str.WriteString("; ")
		}
		str.WriteString(sortExpr.Expr.String())

		str.WriteString("order=")
		if !sortExpr.Ascending {
			str.WriteString("desc ")
		} else {
			str.WriteString("asc ")
		}

		str.WriteString("nulls=")
		if sortExpr.NullsLast {
			str.WriteString("last")
		} else {
			str.WriteString("first")
		}
	}
	return str.String()
}

func (s *Sort) Children() []LogicalPlan {
	return []LogicalPlan{s.Child}
}

func (f *Sort) Accept(v Visitor) any {
	return v.VisitSort(f)
}

func (s *Sort) Relation(ctx *SchemaContext) *Relation {
	return s.Child.Relation(ctx)
}

type Limit struct {
	Child  LogicalPlan
	Limit  LogicalExpr
	Offset LogicalExpr
}

func (l *Limit) Children() []LogicalPlan {
	return []LogicalPlan{l.Child}
}

func (f *Limit) Accept(v Visitor) any {
	return v.VisitLimit(f)
}

func (l *Limit) Relation(ctx *SchemaContext) *Relation {
	return l.Child.Relation(ctx)
}

func (l *Limit) String() string {
	str := strings.Builder{}
	str.WriteString("LIMIT [")
	str.WriteString(l.Limit.String())
	str.WriteString("]")
	if l.Offset != nil {
		str.WriteString("; offset=[")
		str.WriteString(l.Offset.String())
		str.WriteString("]")
	}
	return str.String()
}

type Distinct struct {
	Child LogicalPlan
}

func (d *Distinct) Children() []LogicalPlan {
	return []LogicalPlan{d.Child}
}

func (f *Distinct) Accept(v Visitor) any {
	return v.VisitDistinct(f)
}

func (d *Distinct) Relation(ctx *SchemaContext) *Relation {
	return d.Child.Relation(ctx)
}

func (d *Distinct) String() string {
	return "DISTINCT"
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

func (f *SetOperation) Accept(v Visitor) any {
	return v.VisitSetOperation(f)
}

func (s *SetOperation) Relation(ctx *SchemaContext) *Relation {
	// Assuming set operations require compatible schemas
	return s.Left.Relation(ctx)
}

func (s *SetOperation) String() string {
	str := strings.Builder{}
	str.WriteString("SET: op=")
	str.WriteString(s.OpType.String())
	str.WriteString("; left=[")
	str.WriteString(s.Left.String())
	str.WriteString("]; right=[")
	str.WriteString(s.Right.String())
	str.WriteString("]")
	return str.String()
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

func (f *Aggregate) Accept(v Visitor) any {
	return v.VisitAggregate(f)
}

func (a *Aggregate) Relation(ctx *SchemaContext) *Relation {
	childRel := a.Child.Relation(ctx)

	ctx2 := ctx.Join(childRel)

	columns := make([]*Column, len(a.AggregateExpressions))
	for i, aggExpr := range a.AggregateExpressions {
		dt, err := aggExpr.DataType(ctx2).Scalar()
		if err != nil {
			panic(err)
		}

		columns[i] = &Column{
			Name:     aggExpr.Name(),
			DataType: dt,
		}
	}

	return &Relation{Columns: columns}
}

func (a *Aggregate) String() string {
	str := strings.Builder{}
	str.WriteString("AGGREGATE: group_by=[")
	for i, expr := range a.GroupingExpressions {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(expr.String())
	}
	str.WriteString("]; aggregates=[")
	for i, expr := range a.AggregateExpressions {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(expr.String())
	}
	str.WriteString("]")
	return str.String()
}

type JoinType int

const (
	InnerJoin JoinType = iota
	LeftOuterJoin
	RightOuterJoin
	FullOuterJoin
)

func (j JoinType) String() string {
	switch j {
	case InnerJoin:
		return "INNER"
	case LeftOuterJoin:
		return "LEFT OUTER"
	case RightOuterJoin:
		return "RIGHT OUTER"
	case FullOuterJoin:
		return "FULL OUTER"
	default:
		panic(fmt.Sprintf("unknown join type %d", j))
	}
}

type SetOperationType int

const (
	Union SetOperationType = iota
	UnionAll
	Intersect
	Except
)

func (s SetOperationType) String() string {
	switch s {
	case Union:
		return "UNION"
	case UnionAll:
		return "UNION ALL"
	case Intersect:
		return "INTERSECT"
	case Except:
		return "EXCEPT"
	default:
		panic(fmt.Sprintf("unknown set operation type %d", s))
	}
}

/*
	###########################
	#                         #
	#   	Expressions		  #
	#                         #
	###########################
*/

type LogicalExpr interface {
	Accepter
	fmt.Stringer
	// Name returns the name of the expression.
	// This can be empty, and is generally only set for ColumnRef
	// or aliased expressions.
	Name() string
	// IsAggregate returns true if the expression is an aggregate.
	IsAggregate() bool
	// UsedColumns returns the columns that are used by the expression, as well as
	// any aggregation expressions that are used.
	// If the projection results in an ambiguous column name or an unknown column,
	// an error is returned. The returned columns will be fully qualified.
	UsedColumns(*SchemaContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error)
	// DataType returns the data type of the expression.
	DataType(*SchemaContext) *ReturnedType
}

// baseExpr is a helper struct that implements the default behavior for an Expression.
type baseExpr struct{}

func (b *baseExpr) Name() string { return "" }

func (b *baseExpr) IsAggregate() bool { return false }

func (b *baseExpr) UsedColumns(ctx *SchemaContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return nil, nil, nil
}

// projectMany is a helper function that projects multiple expressions and combines the results.
func projectMany(ctx *SchemaContext, exprs ...LogicalExpr) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	var columns []*ProjectedColumn
	var exprsToProject []LogicalExpr
	for _, expr := range exprs {
		cols, aggExprs, err := expr.UsedColumns(ctx)
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

func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.Value)
}

func (n *Literal) Accept(v Visitor) any {
	return v.VisitLiteral(n)
}

func (l *Literal) DataType(ctx *SchemaContext) *ReturnedType {
	return &ReturnedType{val: l.Type}
}

// Variable reference
type Variable struct {
	baseExpr
	// name is something like $id, @caller, etc.
	VarName string
}

func (v *Variable) String() string {
	return v.VarName
}

func (n *Variable) Accept(v Visitor) any {
	return v.VisitVariable(n)
}

func (v *Variable) DataType(ctx *SchemaContext) *ReturnedType {
	varType, ok := ctx.Variables[v.VarName]
	if !ok {
		return &ReturnedType{err: fmt.Errorf(`variable "%s" not found`, v.VarName)}
	}

	return &ReturnedType{val: varType}
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

func (c *ColumnRef) UsedColumns(ctx *SchemaContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	column, err := ctx.OuterRelation.Search(c.Parent, c.ColumnName)
	if err != nil {
		return nil, nil, err
	}
	return []*ProjectedColumn{column}, nil, nil
}

func (c *ColumnRef) String() string {
	if c.Parent != "" {
		return fmt.Sprintf("%s.%s", c.Parent, c.ColumnName)
	}
	return c.ColumnName
}

func (n *ColumnRef) Accept(v Visitor) any {
	return v.VisitColumnRef(n)
}

func (c *ColumnRef) DataType(ctx *SchemaContext) *ReturnedType {
	col, err := ctx.OuterRelation.Search(c.Parent, c.ColumnName)
	if err != nil {
		return &ReturnedType{err: err}
	}
	return &ReturnedType{val: col.DataType}
}

type AggregateFunctionCall struct {
	FunctionName string
	Args         []LogicalExpr
	Star         bool
	Distinct     bool
}

func (a *AggregateFunctionCall) Name() string {
	return a.FunctionName
}

func (a *AggregateFunctionCall) IsAggregate() bool {
	return true
}

func (a *AggregateFunctionCall) String() string {
	var buf bytes.Buffer
	buf.WriteString(a.FunctionName)
	buf.WriteString("(")
	if a.Star {
		buf.WriteString("*")
	} else {
		for i, arg := range a.Args {
			if i > 0 {
				buf.WriteString(", ")
			} else if a.Distinct {
				buf.WriteString("DISTINCT ")
			}
			buf.WriteString(arg.String())
		}
	}
	buf.WriteString(")")
	return buf.String()
}

func (n *AggregateFunctionCall) Accept(v Visitor) any {
	return v.VisitAggregateFunctionCall(n)
}

// Function call
// TODO: we should split this into: ScalarFunctionCall, AggregateFunctionCall, ProcedureCall (which would include foreign procedures)
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

func (f *FunctionCall) UsedColumns(ctx *SchemaContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	args, aggs, err := projectMany(ctx, f.Args...)
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

func (n *FunctionCall) Accept(v Visitor) any {
	return v.VisitFunctionCall(n)
}

func (f *FunctionCall) DataType(ctx *SchemaContext) *ReturnedType {
	fn, ok := parse.Functions[f.FunctionName]
	if ok {
		argTypes := make([]*types.DataType, len(f.Args))
		for i, arg := range f.Args {
			var err error
			argTypes[i], err = arg.DataType(ctx).Scalar()
			if err != nil {
				return &ReturnedType{err: err}
			}
		}

		retType, err := fn.ValidateArgs(argTypes)
		if err != nil {
			return &ReturnedType{err: err}
		}

		return &ReturnedType{val: retType}
	}

	// if not built-in, then it must be a procedure
	proc, found := ctx.Schema.FindProcedure(f.FunctionName)
	if !found {
		return &ReturnedType{err: fmt.Errorf(`unknown function or procedure "%s"`, f.FunctionName)}
	}

	if proc.Returns == nil {
		return &ReturnedType{err: fmt.Errorf(`procedure "%s" does not return a value`, f.FunctionName)}
	}

	if len(proc.Returns.Fields) != 1 {
		return &ReturnedType{err: fmt.Errorf(`procedure expression needs to return exactly one value, got %d`, len(proc.Returns.Fields))}
	}
	if proc.Returns.IsTable {
		return &ReturnedType{err: fmt.Errorf(`procedure expression cannot return a table`)}
	}

	return &ReturnedType{val: proc.Returns.Fields[0].Type.Copy()}
}

type ArithmeticOp struct {
	baseExpr
	Left  LogicalExpr
	Right LogicalExpr
	Op    ArithmeticOperator
}

type ArithmeticOperator uint8

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

func (a *ArithmeticOp) Project(ctx *SchemaContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return projectMany(ctx, a.Left, a.Right)
}

func (n *ArithmeticOp) Accept(v Visitor) any {
	return v.VisitArithmeticOp(n)
}

func (a *ArithmeticOp) DataType(ctx *SchemaContext) *ReturnedType {
	left, err := a.Left.DataType(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	right, err := a.Right.DataType(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	if !left.Equals(right) {
		return &ReturnedType{err: fmt.Errorf("mismatched types %s and %s", left.String(), right.String())}
	}
	if !left.IsNumeric() {
		return &ReturnedType{err: fmt.Errorf("non-numeric type %s used in arithmetic", left.String())}
	}

	return &ReturnedType{val: left}
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

type ComparisonOperator uint8

const (
	// operators can have 3 types of classifications:
	// - sargable
	// - not sargable
	// - rarely sarable
	// https://en.wikipedia.org/wiki/Sargable
	// for the purposes of our planner, we will treat rarely
	// sargable as not sargable
	Equal              ComparisonOperator = iota // sargable
	NotEqual                                     // rarely sargable
	LessThan                                     // sargable
	LessThanOrEqual                              // sargable
	GreaterThan                                  // sargable
	GreaterThanOrEqual                           // sargable
	// IS and IS NOT are rarely sargable because they can only be used
	// with NULL values or BOOLEAN values
	Is                // rarely sargable
	IsNot             // rarely sargable
	IsDistinctFrom    // not sargable
	IsNotDistinctFrom // not sargable
)

func (c ComparisonOperator) String() string {
	switch c {
	case Equal:
		return "="
	case NotEqual:
		return "!="
	case LessThan:
		return "<"
	case LessThanOrEqual:
		return "<="
	case GreaterThan:
		return ">"
	case GreaterThanOrEqual:
		return ">="
	case Is:
		return " IS "
	case IsNot:
		return " IS NOT "
	case IsDistinctFrom:
		return " IS DISTINCT FROM "
	case IsNotDistinctFrom:
		return " IS NOT DISTINCT FROM "
	default:
		panic(fmt.Sprintf("unknown comparison operator %d", c))
	}
}

func (c *ComparisonOp) IsAggregate() bool {
	return c.Left.IsAggregate() || c.Right.IsAggregate()
}

func (c *ComparisonOp) UsedColumns(ctx *SchemaContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return projectMany(ctx, c.Left, c.Right)
}

func (n *ComparisonOp) Accept(v Visitor) any {
	return v.VisitComparisonOp(n)
}

func (c *ComparisonOp) DataType(ctx *SchemaContext) *ReturnedType {
	left, err := c.Left.DataType(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	right, err := c.Right.DataType(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	if !left.Equals(right) {
		return &ReturnedType{err: fmt.Errorf("mismatched types %s and %s", left.String(), right.String())}
	}

	return &ReturnedType{val: types.BoolType}
}

func (c *ComparisonOp) String() string {

	return fmt.Sprintf("(%s %s %s)", c.Left.String(), c.Op.String(), c.Right.String())
}

type LogicalOp struct {
	baseExpr
	Left  LogicalExpr
	Right LogicalExpr
	Op    LogicalOperator
}

type LogicalOperator uint8

const (
	And LogicalOperator = iota
	Or
)

func (l *LogicalOp) IsAggregate() bool {
	return l.Left.IsAggregate() || l.Right.IsAggregate()
}

func (l *LogicalOp) Project(ctx *SchemaContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return projectMany(ctx, l.Left, l.Right)
}

func (n *LogicalOp) Accept(v Visitor) any {
	return v.VisitLogicalOp(n)
}

func (l *LogicalOp) DataType(ctx *SchemaContext) *ReturnedType {
	left, err := l.Left.DataType(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	right, err := l.Right.DataType(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	if !left.Equals(types.BoolType) {
		return &ReturnedType{err: fmt.Errorf("non-boolean type %s used in logical operation", left.String())}
	}
	if !right.Equals(types.BoolType) {
		return &ReturnedType{err: fmt.Errorf("non-boolean type %s used in logical operation", right.String())}
	}

	return &ReturnedType{val: types.BoolType}
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

type UnaryOperator uint8

const (
	Negate UnaryOperator = iota
	Not
	Positive
)

func (u UnaryOperator) String() string {
	switch u {
	case Negate:
		return "-"
	case Not:
		return "NOT "
	case Positive:
		return "+"
	default:
		panic(fmt.Sprintf("unknown unary operator %d", u))
	}
}

func (u UnaryOp) String() string {
	return fmt.Sprintf("%s%s", u.Op.String(), u.Expr.String())
}

func (u *UnaryOp) IsAggregate() bool {
	return u.Expr.IsAggregate()
}

func (u *UnaryOp) Project(ctx *SchemaContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return u.Expr.UsedColumns(ctx)
}

func (n *UnaryOp) Accept(v Visitor) any {
	return v.VisitUnaryOp(n)
}

func (u *UnaryOp) DataType(ctx *SchemaContext) *ReturnedType {
	dt, err := u.Expr.DataType(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	switch u.Op {
	case Negate, Positive:
		if !dt.IsNumeric() {
			return &ReturnedType{err: fmt.Errorf("non-numeric type %s used in unary operation", dt.String())}
		}
		if u.Op == Negate && dt.Equals(types.Uint256Type) {
			return &ReturnedType{err: fmt.Errorf("cannot negate uint256 type")}
		}
	case Not:
		if !dt.Equals(types.BoolType) {
			return &ReturnedType{err: fmt.Errorf("non-boolean type %s used in unary operation", dt.String())}
		}
	}

	return &ReturnedType{val: dt}
}

type TypeCast struct {
	Expr LogicalExpr
	Type *types.DataType
}

func (t *TypeCast) IsAggregate() bool {
	return t.Expr.IsAggregate()
}

func (t *TypeCast) Name() string {
	return t.Expr.Name()
}

func (t *TypeCast) UsedColumns(ctx *SchemaContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return t.Expr.UsedColumns(ctx)
}

func (n *TypeCast) Accept(v Visitor) any {
	return v.VisitTypeCast(n)
}

func (t *TypeCast) DataType(ctx *SchemaContext) *ReturnedType {
	// to enforce validation of the child, we call it but ignore the result
	_, err := t.Expr.DataType(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	return &ReturnedType{val: t.Type}
}

func (t *TypeCast) String() string {
	return fmt.Sprintf("(%s::%s)", t.Expr.String(), t.Type.Name)
}

type AliasExpr struct {
	Expr  LogicalExpr
	Alias string
}

func (a *AliasExpr) Name() string {
	return a.Alias
}

func (a *AliasExpr) IsAggregate() bool {
	return a.Expr.IsAggregate()
}

func (a *AliasExpr) UsedColumns(ctx *SchemaContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return a.Expr.UsedColumns(ctx)
}

func (n *AliasExpr) Accept(v Visitor) any {
	return v.VisitAliasExpr(n)
}

func (a *AliasExpr) DataType(ctx *SchemaContext) *ReturnedType {
	return a.Expr.DataType(ctx)
}

func (a *AliasExpr) String() string {
	return fmt.Sprintf("%s AS %s", a.Expr.String(), a.Alias)
}

type ArrayAccess struct {
	baseExpr
	Array LogicalExpr
	Index LogicalExpr
}

func (a *ArrayAccess) IsAggregate() bool {
	return a.Array.IsAggregate() || a.Index.IsAggregate()
}

func (a *ArrayAccess) UsedColumns(ctx *SchemaContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return projectMany(ctx, a.Array, a.Index)
}

func (n *ArrayAccess) Accept(v Visitor) any {
	return v.VisitArrayAccess(n)
}

func (a *ArrayAccess) DataType(ctx *SchemaContext) *ReturnedType {
	arrayType, err := a.Array.DataType(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	if !arrayType.IsArray {
		return &ReturnedType{err: fmt.Errorf("non-array type %s used in array access", arrayType.String())}
	}

	idxType, err := a.Index.DataType(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	if !idxType.IsNumeric() {
		return &ReturnedType{err: fmt.Errorf("non-numeric type %s used as array index", idxType.String())}
	}

	retTyp := arrayType.Copy()
	retTyp.IsArray = false
	return &ReturnedType{val: retTyp}
}

func (a *ArrayAccess) String() string {
	return fmt.Sprintf("%s[%s]", a.Array.String(), a.Index.String())
}

type ArrayConstructor struct {
	baseExpr
	Elements []LogicalExpr
}

func (a *ArrayConstructor) IsAggregate() bool {
	for _, elem := range a.Elements {
		if elem.IsAggregate() {
			return true
		}
	}
	return false
}

func (a *ArrayConstructor) UsedColumns(ctx *SchemaContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return projectMany(ctx, a.Elements...)
}

func (n *ArrayConstructor) Accept(v Visitor) any {
	return v.VisitArrayConstructor(n)
}

func (a *ArrayConstructor) DataType(ctx *SchemaContext) *ReturnedType {
	if len(a.Elements) == 0 {
		return &ReturnedType{err: fmt.Errorf("empty array constructor")}
	}

	elemType, err := a.Elements[0].DataType(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	for _, elem := range a.Elements[1:] {
		elem2Type, err := elem.DataType(ctx).Scalar()
		if err != nil {
			return &ReturnedType{err: err}
		}

		if !elem2Type.Equals(elemType) {
			return &ReturnedType{err: fmt.Errorf("mismatched types in array constructor")}
		}
	}

	dt := elemType.Copy()
	dt.IsArray = true
	return &ReturnedType{val: dt}
}

func (a *ArrayConstructor) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	for i, elem := range a.Elements {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(elem.String())
	}
	buf.WriteString("]")
	return buf.String()
}

type FieldAccess struct {
	baseExpr
	Object LogicalExpr
	Field  string
}

func (f *FieldAccess) IsAggregate() bool {
	return f.Object.IsAggregate()
}

func (f *FieldAccess) UsedColumns(ctx *SchemaContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return f.Object.UsedColumns(ctx)
}

func (n *FieldAccess) Accept(v Visitor) any {
	return v.VisitFieldAccess(n)
}

func (f *FieldAccess) DataType(ctx *SchemaContext) *ReturnedType {
	objType, err := f.Object.DataType(ctx).Object()
	if err != nil {
		return &ReturnedType{err: err}
	}

	dt, ok := objType[f.Field]
	if !ok {
		return &ReturnedType{err: fmt.Errorf(`unknown field "%s"`, f.Field)}
	}

	return &ReturnedType{val: dt}
}

func (f *FieldAccess) String() string {
	return fmt.Sprintf("%s.%s", f.Object.String(), f.Field)
}

// ReturnedType is a struct that is returned from the Scalar() method
// of LogicalExpr implementations. It can be used to coerce the return type,
// and to handle error returns. Callers should never access the fields directly.
type ReturnedType struct {
	// val is the data type that is returned by the expression.
	// It is either a single data type or a map of data types.
	val any
	// err is the error that was returned during the evaluation of the expression.
	// It is added here as a convenience so that DataType itself does not have to
	// return an error, requiring the callers to check for errors twice.
	err error
}

// Scalar attempts to coerce the return type to a single data type.
func (r *ReturnedType) Scalar() (*types.DataType, error) {
	if r.err != nil {
		return nil, r.err
	}

	dt, ok := r.val.(*types.DataType)
	if !ok {
		// this can be triggered by a user if they try to directly use an object
		// in an expression
		_, ok = r.val.(map[string]*types.DataType)
		if ok {
			return nil, fmt.Errorf("referenced expression is an object, expected scalar or array. specify a field to access using the . operator")
		}

		// this is an internal bug
		panic(fmt.Sprintf("unexpected return type %T", r.val))
	}
	return dt, nil
}

// Object attempts to coerce the return type to a map of data types.
func (r *ReturnedType) Object() (map[string]*types.DataType, error) {
	if r.err != nil {
		return nil, r.err
	}

	obj, ok := r.val.(map[string]*types.DataType)
	if !ok {
		// this can be triggered by a user if they try to use dot notation
		// on a scalar
		v, ok := r.val.(*types.DataType)
		if ok {
			if v.IsArray {
				return nil, fmt.Errorf("referenced expression is an array, expected object")
			}
			return nil, fmt.Errorf("referenced expression is a scalar, expected object")
		}

		// this is an internal bug
		panic(fmt.Sprintf("unexpected return type %T", r.val))
	}
	return obj, nil
}

type Subquery struct {
	SubqueryType SubqueryType
	Query        LogicalPlan
}

var _ LogicalExpr = (*Subquery)(nil)

func (n *Subquery) Accept(v Visitor) any {
	return v.VisitSubquery(n)
}

func (s *Subquery) DataType(ctx *SchemaContext) *ReturnedType {
	childRel := s.Query.Relation(ctx)

	// subqueries must only return 1 column
	if s.SubqueryType == ScalarSubquery {
		if len(childRel.Columns) != 1 {
			return &ReturnedType{err: fmt.Errorf("scalar subquery must return exactly one column")}
		}

		return &ReturnedType{val: childRel.Columns[0].DataType}

	}

	// EXISTS and NOT EXISTS subqueries always return a boolean.
	// They can have as many underlying columns as they want.
	return &ReturnedType{val: types.BoolType}
}

func (s *Subquery) UsedColumns(ctx *SchemaContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	panic("not implemented")
}

func (s *Subquery) String() string {
	var subqueryType string
	switch s.SubqueryType {
	case ScalarSubquery:
		subqueryType = "SCALAR"
	case ExistsSubquery:
		subqueryType = "EXISTS"
	case NotExistsSubquery:
		subqueryType = "NOT EXISTS"
	}

	return fmt.Sprintf("%s SUBQUERY %s", subqueryType, s.Query.String())
}

func (s *Subquery) IsAggregate() bool {
	return false
}

func (s *Subquery) Name() string {
	return ""
}

type SubqueryType uint8

const (
	ScalarSubquery SubqueryType = iota
	ExistsSubquery
	NotExistsSubquery
)

type Accepter interface {
	Accept(Visitor) any
}

// Visitor is an interface that can be implemented to visit all nodes in a logical plan.
// It allows easy construction of both pre and post-order traversal of the logical plan.
// For preorder traversal, state can be passed down the tree via the struct that implements Visitor.
// For postorder traversal, state can be passed up the tree via the return value of the Visit method.
type Visitor interface {
	VisitNoop(*Noop) any
	VisitTableScan(*TableScan) any
	VisitProcedureScan(*ProcedureScan) any
	VisitScanAlias(*ScanAlias) any
	VisitProject(*Project) any
	VisitFilter(*Filter) any
	VisitJoin(*Join) any
	VisitSort(*Sort) any
	VisitLimit(*Limit) any
	VisitDistinct(*Distinct) any
	VisitSetOperation(*SetOperation) any
	VisitAggregate(*Aggregate) any
	VisitLiteral(*Literal) any
	VisitVariable(*Variable) any
	VisitColumnRef(*ColumnRef) any
	VisitAggregateFunctionCall(*AggregateFunctionCall) any
	VisitFunctionCall(*FunctionCall) any
	VisitArithmeticOp(*ArithmeticOp) any
	VisitComparisonOp(*ComparisonOp) any
	VisitLogicalOp(*LogicalOp) any
	VisitUnaryOp(*UnaryOp) any
	VisitTypeCast(*TypeCast) any
	VisitAliasExpr(*AliasExpr) any
	VisitArrayAccess(*ArrayAccess) any
	VisitArrayConstructor(*ArrayConstructor) any
	VisitFieldAccess(*FieldAccess) any
	VisitSubquery(*Subquery) any
}
