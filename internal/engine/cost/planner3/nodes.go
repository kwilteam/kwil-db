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

	The Relation() method of LogicalPlan returns the structure of the relation that the plan represents.
	This is NOT equivalent to the set of reference-able columns in the plan. Reference-able columns are
	tracked in the passed context, where the called method will modify the passed ctx. For example, if we were
	evaluating the query "SELECT name from users where id = 1", the returned relation would be a single
	column "name", but the reference-able columns would be "name" and "id" (and any other columns in the "users" table).
*/

func Format(plan LogicalNode, indent int) string {
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

type LogicalNode interface {
	fmt.Stringer
	Accepter
	Children() []LogicalNode
}

type LogicalPlan interface {
	LogicalNode
	Relation(*PlanContext) *Relation
}

type Noop struct{}

func (n *Noop) Children() []LogicalNode {
	return []LogicalNode{}
}

func (f *Noop) Accept(v Visitor) any {
	return v.VisitNoop(f)
}

func (n *Noop) Relation(ctx *PlanContext) *Relation {
	return &Relation{}
}

func (n *Noop) String() string {
	return "NOOP"
}

// ScanSource is a source of data that a Scan can be performed on.
// This is either a physical table, a procedure call that returns a table,
// or a subquery. Scan sources themselves are logical plans, however their
// implementations should NOT alter the schema context.
type ScanSource interface {
	LogicalPlan
	scanSource()
}

// TableScanSource represents a scan of a physical table or a CTE.
type TableScanSource struct {
	TableName string
}

func (t *TableScanSource) Children() []LogicalNode {
	return []LogicalNode{}
}

func (f *TableScanSource) Accept(v Visitor) any {
	return v.VisitTableScanSource(f)
}

func (t *TableScanSource) Relation(ctx *PlanContext) *Relation {
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

func (t *TableScanSource) String() string {
	return fmt.Sprintf("SCAN TABLE %s", t.TableName)
}

func (t *TableScanSource) scanSource() {}

// ProcedureScanSource represents a scan of a function.
// It can call either a local procedure or foreign procedure
// that returns a table.
type ProcedureScanSource struct {
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

func (f *ProcedureScanSource) Children() []LogicalNode {
	var c []LogicalNode
	for _, arg := range f.Args {
		c = append(c, arg)
	}

	for _, arg := range f.ContextualArgs {
		c = append(c, arg)
	}

	return c
}

func (f *ProcedureScanSource) Accept(v Visitor) any {
	return v.VisitProcedureScanSource(f)
}

func (f *ProcedureScanSource) Relation(ctx *PlanContext) *Relation {
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

	var cols []*ReferenceableColumn
	for _, field := range procReturns.Fields {
		cols = append(cols, &ReferenceableColumn{
			// the Parent will get set by the ScanAlias
			Name:     field.Name,
			DataType: field.Type.Copy(),
		})
	}

	return &Relation{Columns: cols}
}

func (f *ProcedureScanSource) String() string {
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

func (f *ProcedureScanSource) scanSource() {}

// SubqueryScanSource represents a scan of a subquery.
// This is used, for example, in the query "SELECT * FROM (SELECT * FROM users) AS subquery".
type SubqueryScanSource struct {
	Subquery LogicalPlan
}

func (s *SubqueryScanSource) Children() []LogicalNode {
	return []LogicalNode{s.Subquery}
}

func (f *SubqueryScanSource) Accept(v Visitor) any {
	return v.VisitSubqueryScanSource(f)
}

func (s *SubqueryScanSource) Relation(ctx *PlanContext) *Relation {
	// we create a new shallow-copied context, and also create new
	// current and outer relations. The passed current relation is
	// added to the outer relation, and the current relation is emptied.
	// This doesn't affect the schema context passed in.
	newCtx := *ctx
	newCtx.CurrentRelation = &Relation{}
	newCtx.OuterRelation = &Relation{
		Columns: append(ctx.CurrentRelation.Columns, ctx.OuterRelation.Columns...),
	}

	rel := s.Subquery.Relation(&newCtx)

	return rel
}

func (s *SubqueryScanSource) String() string {
	return "SCAN SUBQUERY"
}

func (s *SubqueryScanSource) scanSource() {}

type Scan struct {
	Child ScanSource
	// RelationName will always be set.
	// If the scan is a table scan and no alias was specified,
	// the RelationName will be the table name.
	// All other scan types (functions and subqueries) require an alias.
	RelationName string
}

func (s *Scan) Children() []LogicalNode {
	return []LogicalNode{s.Child}
}

func (f *Scan) Accept(v Visitor) any {
	return v.VisitScanAlias(f)
}

func (s *Scan) Relation(ctx *PlanContext) *Relation {
	rel := s.Child.Relation(ctx)

	// apply the name to all scanned columns
	for _, col := range rel.Columns {
		col.Parent = s.RelationName
	}

	ctx.Join(rel)

	return rel
}

func (s *Scan) String() string {
	return fmt.Sprintf("ALIAS %s", s.RelationName)
}

type Project struct {
	Expressions []LogicalExpr
	Child       LogicalPlan
}

func (p *Project) Children() []LogicalNode {
	var c []LogicalNode
	for _, expr := range p.Expressions {
		c = append(c, expr)
	}
	c = append(c, p.Child)

	return c
}

func (f *Project) Accept(v Visitor) any {
	return v.VisitProject(f)
}

func (p *Project) Relation(ctx *PlanContext) *Relation {
	// we can ignore the childs relation, since the returned relation
	// from this method will be the columns projected on the ctx.CurrenRelation
	p.Child.Relation(ctx)

	columns := make([]*ReferenceableColumn, len(p.Expressions))
	for i, expr := range p.Expressions {
		dt, err := expr.Analyze(ctx).Scalar()
		if err != nil {
			panic(err)
		}

		// TODO: this might end up causing issues because
		// the column is not fully qualified
		columns[i] = &ReferenceableColumn{
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

func (f *Filter) Children() []LogicalNode {
	return []LogicalNode{f.Child, f.Condition}
}

func (f *Filter) Accept(v Visitor) any {
	return v.VisitFilter(f)
}

func (f *Filter) Relation(ctx *PlanContext) *Relation {
	// we don't care about the result, just that it is a scalar, and is a boolean
	dt, err := f.Condition.Analyze(ctx).Scalar()
	if err != nil {
		panic(err)
	}

	if !dt.Equals(types.BoolType) {
		panic(fmt.Errorf("filter condition evaluate to a boolean, got %s", dt.String()))
	}

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

func (j *Join) Children() []LogicalNode {
	return []LogicalNode{j.Left, j.Right, j.Condition}
}

func (f *Join) Accept(v Visitor) any {
	return v.VisitJoin(f)
}

// With a Join, the schema context is modified to include the columns from both relations.
// These can then be referenced by all callers of the join, as well as in the join condition.
func (j *Join) Relation(ctx *PlanContext) *Relation {
	// we don't need to worry about modifying the passed context, since the
	// the tables being joined within Left and Right will already have been
	// joined in the schema context.
	leftRel := j.Left.Relation(ctx)
	rightRel := j.Right.Relation(ctx)
	columns := append(leftRel.Columns, rightRel.Columns...)

	// we need to check that the join condition is a boolean
	dt, err := j.Condition.Analyze(ctx).Scalar()
	if err != nil {
		panic(err)
	}

	if !dt.Equals(types.BoolType) {
		panic(fmt.Errorf("join condition evaluate to a boolean, got %s", dt.String()))
	}

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

func (s *Sort) Children() []LogicalNode {
	var c []LogicalNode
	for _, sortExpr := range s.SortExpressions {
		c = append(c, sortExpr.Expr)
	}
	c = append(c, s.Child)

	return c
}

func (f *Sort) Accept(v Visitor) any {
	return v.VisitSort(f)
}

func (s *Sort) Relation(ctx *PlanContext) *Relation {
	// we need to visit the child first before sorting, so that
	// the expressions in the sort can be validated against the current schema context.
	rel := s.Child.Relation(ctx)

	for _, sortExpr := range s.SortExpressions {
		_, err := sortExpr.Expr.Analyze(ctx).Scalar()
		if err != nil {
			panic(err)
		}
	}

	return rel
}

type Limit struct {
	Child  LogicalPlan
	Limit  LogicalExpr
	Offset LogicalExpr
}

func (l *Limit) Children() []LogicalNode {
	return []LogicalNode{l.Child, l.Limit, l.Offset}
}

func (f *Limit) Accept(v Visitor) any {
	return v.VisitLimit(f)
}

func (l *Limit) Relation(ctx *PlanContext) *Relation {
	// we need to visit the child first before limiting, so that
	// the expressions in the limit can be validated against the current schema context.
	rel := l.Child.Relation(ctx)

	// the limit and offset must evaluate to integers
	dt, err := l.Limit.Analyze(ctx).Scalar()
	if err != nil {
		panic(err)
	}

	if !dt.Equals(types.IntType) {
		panic(fmt.Errorf("limit must evaluate to an integer, got %s", dt.String()))
	}

	if l.Offset != nil {
		dt, err := l.Offset.Analyze(ctx).Scalar()
		if err != nil {
			panic(err)
		}

		if !dt.Equals(types.IntType) {
			panic(fmt.Errorf("offset must evaluate to an integer, got %s", dt.String()))
		}
	}

	return rel
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

func (d *Distinct) Children() []LogicalNode {
	return []LogicalNode{d.Child}
}

func (f *Distinct) Accept(v Visitor) any {
	return v.VisitDistinct(f)
}

func (d *Distinct) Relation(ctx *PlanContext) *Relation {
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
func (s *SetOperation) Children() []LogicalNode {
	return []LogicalNode{s.Left, s.Right}
}

func (f *SetOperation) Accept(v Visitor) any {
	return v.VisitSetOperation(f)
}

func (s *SetOperation) Relation(ctx *PlanContext) *Relation {
	// a set operation is pretty unique, and modifies the schema context.
	// Any query that relies on a set operation can only reference elements
	// that are present in the returned left-most relation. For example, say
	// we have a table "users" with 3 columns: id, name, age. If we have
	// "SELECT name, age FROM users UNION SELECT other_text, other_int FROM table2",
	// the calling relation can only reference columns "name" and "age". They cannot be
	// referenced as part of the "users" relation, and no other columns from "users"
	// or "table2" can be referenced.

	leftRel := s.Left.Relation(ctx)
	rightRel := s.Right.Relation(ctx)

	if len(leftRel.Columns) != len(rightRel.Columns) {
		panic("compound operations must have the same number of columns")
	}

	for i, col := range rightRel.Columns {
		if !leftRel.Columns[i].DataType.Equals(col.DataType) {
			panic(fmt.Errorf("cannot use compound query: mismatched data types %s and %s at index %d", leftRel.Columns[i].DataType.String(), col.DataType.String(), i+1))
		}
	}

	// modify the schema context to only include the left relation, unqualified
	for _, col := range leftRel.Columns {
		col.Parent = ""
	}
	ctx.CurrentRelation = leftRel

	return leftRel
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

func (a *Aggregate) Children() []LogicalNode {
	var c []LogicalNode
	for _, expr := range a.GroupingExpressions {
		c = append(c, expr)
	}
	for _, expr := range a.AggregateExpressions {
		c = append(c, expr)
	}
	c = append(c, a.Child)

	return c
}

func (f *Aggregate) Accept(v Visitor) any {
	return v.VisitAggregate(f)
}

func (a *Aggregate) Relation(ctx *PlanContext) *Relation {
	// aggregate is another tricky one, because while it doesn't change the columns
	// that can be referenced, it does change how they can be referenced. If an aggregate
	// expression is present (which is implied if we are in this method), then all columns
	// referenced above the aggregate must be part of the GROUP BY clause OR be within
	// an aggregate function.

	a.Child.Relation(ctx)

	for _, expr := range a.GroupingExpressions {
		_, err := expr.Analyze(ctx).Scalar()
		if err != nil {
			panic(err)
		}
	}

	aggChecker, err := newAggregateChecker(a.GroupingExpressions)
	if err != nil {
		panic(err)
	}

	// In an aggregate, only expressions that are part of the GROUP BY clause
	// or are within an aggregate function can be referenced.

	columns := make([]*ReferenceableColumn, len(a.AggregateExpressions))
	for i, aggExpr := range a.AggregateExpressions {
		dt, err := aggExpr.Analyze(ctx).Scalar()
		if err != nil {
			panic(err)
		}

		columns[i] = &ReferenceableColumn{
			Name:     aggExpr.Name(),
			DataType: dt,
		}
	}

	err = aggChecker.checkMany(a.AggregateExpressions)
	if err != nil {
		panic(err)
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
	LogicalNode
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
	UsedColumns(*PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error)
	// DataType returns the data type of the expression.
	Analyze(*PlanContext) *ReturnedType
}

// baseExpr is a helper struct that implements the default behavior for an Expression.
type baseExpr struct{}

func (b *baseExpr) Name() string { return "" }

func (b *baseExpr) IsAggregate() bool { return false }

func (b *baseExpr) UsedColumns(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return nil, nil, nil
}

// projectMany is a helper function that projects multiple expressions and combines the results.
func projectMany(ctx *PlanContext, exprs ...LogicalExpr) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
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

func (l *Literal) Analyze(ctx *PlanContext) *ReturnedType {
	return &ReturnedType{val: l.Type}
}

func (l *Literal) Children() []LogicalNode {
	return []LogicalNode{}
}

// Variable reference
type Variable struct {
	baseExpr
	// name is something like $id, @caller, etc.
	VarName string
	// DataType is the data type of the variable.
	Type *types.DataType // TODO: make sure the planner adds this
}

func (v *Variable) String() string {
	return v.VarName
}

func (n *Variable) Accept(v Visitor) any {
	return v.VisitVariable(n)
}

func (v *Variable) Analyze(ctx *PlanContext) *ReturnedType {
	varType, ok := ctx.Variables[v.VarName]
	if !ok {
		// could also be an object
		obj, ok := ctx.Objects[v.VarName]
		if !ok {
			return &ReturnedType{err: fmt.Errorf(`unknown variable "%s"`, v.VarName)}
		}

		return &ReturnedType{val: obj}
	}

	return &ReturnedType{val: varType}
}

func (v *Variable) Children() []LogicalNode {
	return []LogicalNode{}
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

func (c *ColumnRef) UsedColumns(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	column, err := ctx.OuterRelation.Search(c.Parent, c.ColumnName)
	if err != nil {
		return nil, nil, err
	}
	return []*ProjectedColumn{{
		Parent:   c.Parent,
		Name:     c.ColumnName,
		DataType: column.DataType,
	}}, nil, nil
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

func (c *ColumnRef) Analyze(ctx *PlanContext) *ReturnedType {
	col, err := ctx.OuterRelation.Search(c.Parent, c.ColumnName)
	if err != nil {
		return &ReturnedType{err: err}
	}

	// qualify the column reference
	if c.Parent == "" {
		c.Parent = col.Parent
	}

	return &ReturnedType{val: col.DataType}
}

func (c *ColumnRef) Children() []LogicalNode {
	return []LogicalNode{}
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

func (a *AggregateFunctionCall) Children() []LogicalNode {
	var c []LogicalNode
	for _, arg := range a.Args {
		c = append(c, arg)
	}
	return c
}

func (a *AggregateFunctionCall) UsedColumns(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	panic("not implemented")
}

func (a *AggregateFunctionCall) Analyze(ctx *PlanContext) *ReturnedType {
	fn, ok := parse.Functions[a.FunctionName]
	if !ok {
		return &ReturnedType{err: fmt.Errorf(`unknown function "%s"`, a.FunctionName)}
	}

	argTypes := make([]*types.DataType, len(a.Args))
	for i, arg := range a.Args {
		var err error
		argTypes[i], err = arg.Analyze(ctx).Scalar()
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

// Function call
type ScalarFunctionCall struct {
	FunctionName string
	Args         []LogicalExpr
}

func (f *ScalarFunctionCall) Name() string {
	return f.FunctionName
}

func (f *ScalarFunctionCall) IsAggregate() bool {
	fn, ok := parse.Functions[f.FunctionName]
	if !ok {
		panic(fmt.Errorf("function %s not found", f.FunctionName))
	}

	return fn.IsAggregate
}

func (f *ScalarFunctionCall) String() string {
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

func (f *ScalarFunctionCall) UsedColumns(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
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

func (n *ScalarFunctionCall) Accept(v Visitor) any {
	return v.VisitFunctionCall(n)
}

func (f *ScalarFunctionCall) Analyze(ctx *PlanContext) *ReturnedType {
	fn, ok := parse.Functions[f.FunctionName]
	if ok {
		argTypes := make([]*types.DataType, len(f.Args))
		for i, arg := range f.Args {
			var err error
			argTypes[i], err = arg.Analyze(ctx).Scalar()
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

func (f *ScalarFunctionCall) Children() []LogicalNode {
	var c []LogicalNode
	for _, arg := range f.Args {
		c = append(c, arg)
	}
	return c
}

// ProcedureCall is a call to a procedure.
// This can be a call to either a procedure in the same schema, or
// to a foreign procedure.
type ProcedureCall struct {
	ProcedureName string
	Foreign       bool
	Args          []LogicalExpr
	ContextArgs   []LogicalExpr
}

func (p *ProcedureCall) Name() string {
	return p.ProcedureName
}

func (p *ProcedureCall) IsAggregate() bool {
	return false
}

func (p *ProcedureCall) String() string {
	var buf bytes.Buffer
	if p.Foreign {
		buf.WriteString("FOREIGN ")
	}
	buf.WriteString("PROCEDURE ")
	buf.WriteString(p.ProcedureName)
	buf.WriteString("(")
	for i, arg := range p.Args {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(arg.String())
	}
	buf.WriteString(")")
	return buf.String()
}

func (p *ProcedureCall) UsedColumns(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	panic("not implemented")
}

func (n *ProcedureCall) Accept(v Visitor) any {
	return v.VisitProcedureCall(n)
}

func (p *ProcedureCall) Analyze(ctx *PlanContext) *ReturnedType {
	// must be either a procedure or a foreign procedure
	var neededArgs []*types.DataType
	var returns *types.ProcedureReturn
	if p.Foreign {
		foreignProc, ok := ctx.Schema.FindForeignProcedure(p.ProcedureName)
		if !ok {
			return &ReturnedType{err: fmt.Errorf(`foreign procedure "%s" not found`, p.ProcedureName)}
		}
		neededArgs = foreignProc.Parameters
		returns = foreignProc.Returns

		// if it is foreign, there must be 2 contextual variables, both evaluating to strings
		if len(p.ContextArgs) != 2 {
			return &ReturnedType{err: fmt.Errorf(`foreign procedure "%s" expects 2 contextual arguments, got %d`, p.ProcedureName, len(p.ContextArgs))}
		}

		for i, arg := range p.ContextArgs {
			argType, err := arg.Analyze(ctx).Scalar()
			if err != nil {
				return &ReturnedType{err: err}
			}

			if !argType.Equals(types.TextType) {
				return &ReturnedType{err: fmt.Errorf(`contextual argument %d to foreign procedure "%s" expects type %s, got %s`, i+1, p.ProcedureName, types.TextType.String(), argType.String())}
			}
		}
	} else {
		proc, ok := ctx.Schema.FindProcedure(p.ProcedureName)
		if !ok {
			return &ReturnedType{err: fmt.Errorf(`procedure "%s" not found`, p.ProcedureName)}
		}
		for _, param := range proc.Parameters {
			neededArgs = append(neededArgs, param.Type)
		}
		returns = proc.Returns
	}

	if returns == nil {
		return &ReturnedType{err: fmt.Errorf(`procedure "%s" does not return a value`, p.ProcedureName)}
	}
	if returns.IsTable {
		return &ReturnedType{err: fmt.Errorf(`procedure used in an expression "%s" cannot return a table`, p.ProcedureName)}
	}
	if len(returns.Fields) != 1 {
		return &ReturnedType{err: fmt.Errorf(`procedure expression needs to return exactly one value, got %d`, len(returns.Fields))}
	}

	if len(p.Args) != len(neededArgs) {
		return &ReturnedType{err: fmt.Errorf(`procedure "%s" expects %d arguments, got %d`, p.ProcedureName, len(neededArgs), len(p.Args))}
	}

	for i, arg := range p.Args {
		argType, err := arg.Analyze(ctx).Scalar()
		if err != nil {
			return &ReturnedType{err: err}
		}

		if !argType.Equals(neededArgs[i]) {
			return &ReturnedType{err: fmt.Errorf(`argument %d to procedure "%s" expects type %s, got %s`, i+1, p.ProcedureName, neededArgs[i].String(), argType.String())}
		}
	}

	return &ReturnedType{val: returns.Fields[0].Type.Copy()}
}

func (p *ProcedureCall) Children() []LogicalNode {
	var c []LogicalNode
	for _, arg := range p.Args {
		c = append(c, arg)
	}
	return c
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

func (a *ArithmeticOp) Project(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return projectMany(ctx, a.Left, a.Right)
}

func (n *ArithmeticOp) Accept(v Visitor) any {
	return v.VisitArithmeticOp(n)
}

func (a *ArithmeticOp) Analyze(ctx *PlanContext) *ReturnedType {
	left, err := a.Left.Analyze(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	right, err := a.Right.Analyze(ctx).Scalar()
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

func (a *ArithmeticOp) Children() []LogicalNode {
	return []LogicalNode{a.Left, a.Right}
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

func (c *ComparisonOp) UsedColumns(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return projectMany(ctx, c.Left, c.Right)
}

func (n *ComparisonOp) Accept(v Visitor) any {
	return v.VisitComparisonOp(n)
}

func (c *ComparisonOp) Analyze(ctx *PlanContext) *ReturnedType {
	left, err := c.Left.Analyze(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	right, err := c.Right.Analyze(ctx).Scalar()
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

func (c *ComparisonOp) Children() []LogicalNode {
	return []LogicalNode{c.Left, c.Right}
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

func (l *LogicalOp) Project(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return projectMany(ctx, l.Left, l.Right)
}

func (n *LogicalOp) Accept(v Visitor) any {
	return v.VisitLogicalOp(n)
}

func (l *LogicalOp) Analyze(ctx *PlanContext) *ReturnedType {
	left, err := l.Left.Analyze(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	right, err := l.Right.Analyze(ctx).Scalar()
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

func (l *LogicalOp) Children() []LogicalNode {
	return []LogicalNode{l.Left, l.Right}
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

func (u *UnaryOp) Project(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return u.Expr.UsedColumns(ctx)
}

func (n *UnaryOp) Accept(v Visitor) any {
	return v.VisitUnaryOp(n)
}

func (u *UnaryOp) Analyze(ctx *PlanContext) *ReturnedType {
	dt, err := u.Expr.Analyze(ctx).Scalar()
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

func (u *UnaryOp) Children() []LogicalNode {
	return []LogicalNode{u.Expr}
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

func (t *TypeCast) UsedColumns(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return t.Expr.UsedColumns(ctx)
}

func (n *TypeCast) Accept(v Visitor) any {
	return v.VisitTypeCast(n)
}

func (t *TypeCast) Analyze(ctx *PlanContext) *ReturnedType {
	// to enforce validation of the child, we call it but ignore the result
	_, err := t.Expr.Analyze(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	return &ReturnedType{val: t.Type}
}

func (t *TypeCast) String() string {
	return fmt.Sprintf("(%s::%s)", t.Expr.String(), t.Type.Name)
}

func (t *TypeCast) Children() []LogicalNode {
	return []LogicalNode{t.Expr}
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

func (a *AliasExpr) UsedColumns(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return a.Expr.UsedColumns(ctx)
}

func (n *AliasExpr) Accept(v Visitor) any {
	return v.VisitAliasExpr(n)
}

func (a *AliasExpr) Analyze(ctx *PlanContext) *ReturnedType {
	return a.Expr.Analyze(ctx)
}

func (a *AliasExpr) String() string {
	return fmt.Sprintf("%s AS %s", a.Expr.String(), a.Alias)
}

func (a *AliasExpr) Children() []LogicalNode {
	return []LogicalNode{a.Expr}
}

type ArrayAccess struct {
	baseExpr
	Array LogicalExpr
	Index LogicalExpr
}

func (a *ArrayAccess) IsAggregate() bool {
	return a.Array.IsAggregate() || a.Index.IsAggregate()
}

func (a *ArrayAccess) UsedColumns(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return projectMany(ctx, a.Array, a.Index)
}

func (n *ArrayAccess) Accept(v Visitor) any {
	return v.VisitArrayAccess(n)
}

func (a *ArrayAccess) Analyze(ctx *PlanContext) *ReturnedType {
	arrayType, err := a.Array.Analyze(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	if !arrayType.IsArray {
		return &ReturnedType{err: fmt.Errorf("non-array type %s used in array access", arrayType.String())}
	}

	idxType, err := a.Index.Analyze(ctx).Scalar()
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

func (a *ArrayAccess) Children() []LogicalNode {
	return []LogicalNode{a.Array, a.Index}
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

func (a *ArrayConstructor) UsedColumns(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return projectMany(ctx, a.Elements...)
}

func (n *ArrayConstructor) Accept(v Visitor) any {
	return v.VisitArrayConstructor(n)
}

func (a *ArrayConstructor) Analyze(ctx *PlanContext) *ReturnedType {
	if len(a.Elements) == 0 {
		return &ReturnedType{err: fmt.Errorf("empty array constructor")}
	}

	elemType, err := a.Elements[0].Analyze(ctx).Scalar()
	if err != nil {
		return &ReturnedType{err: err}
	}

	for _, elem := range a.Elements[1:] {
		elem2Type, err := elem.Analyze(ctx).Scalar()
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

func (a *ArrayConstructor) Children() []LogicalNode {
	var c []LogicalNode
	for _, elem := range a.Elements {
		c = append(c, elem)
	}
	return c
}

type FieldAccess struct {
	baseExpr
	Object LogicalExpr
	Field  string
}

func (f *FieldAccess) IsAggregate() bool {
	return f.Object.IsAggregate()
}

func (f *FieldAccess) UsedColumns(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
	return f.Object.UsedColumns(ctx)
}

func (n *FieldAccess) Accept(v Visitor) any {
	return v.VisitFieldAccess(n)
}

func (f *FieldAccess) Analyze(ctx *PlanContext) *ReturnedType {
	objType, err := f.Object.Analyze(ctx).Object()
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

func (f *FieldAccess) Children() []LogicalNode {
	return []LogicalNode{f.Object}
}

type Subquery struct {
	SubqueryType SubqueryType
	Query        LogicalPlan
}

var _ LogicalExpr = (*Subquery)(nil)

func (n *Subquery) Accept(v Visitor) any {
	return v.VisitSubquery(n)
}

func (s *Subquery) Analyze(ctx *PlanContext) *ReturnedType {
	ctx2 := ctx.Copy() // copy to not pollute the outer context
	// TODO: this wont work, since the subquery wont think we are in an aggregate,
	// but a subquery that is a result column that is also not an aggregate will
	// need to check for invalid aggregate correlated column references
	childRel := s.Query.Relation(ctx2)

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

func (s *Subquery) UsedColumns(ctx *PlanContext) (projectedColumns []*ProjectedColumn, aggregationExprs []LogicalExpr, err error) {
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

func (s *Subquery) Children() []LogicalNode {
	return []LogicalNode{s.Query}
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
	VisitTableScanSource(*TableScanSource) any
	VisitProcedureScanSource(*ProcedureScanSource) any
	VisitSubqueryScanSource(*SubqueryScanSource) any
	VisitScanAlias(*Scan) any
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
	VisitFunctionCall(*ScalarFunctionCall) any
	VisitProcedureCall(*ProcedureCall) any
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

// flatten flattens a logical plan into a slice of nodes.
func flatten(node LogicalNode) []LogicalNode {
	nodes := []LogicalNode{node}
	for _, child := range node.Children() {
		nodes = append(nodes, flatten(child)...)
	}
	return nodes
}

// traverse traverses a logical plan in preorder.
// It will call the callback function for each node in the plan.
// If the callback function returns false, the traversal will not
// continue to the children of the node.
func traverse(node LogicalNode, callback func(node LogicalNode) bool) {
	if !callback(node) {
		return
	}
	for _, child := range node.Children() {
		traverse(child, callback)
	}
}
