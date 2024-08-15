package planner

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
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

// Traversable is an interface for nodes that can be traversed.
type Traversable interface {
	// Children returns the children of the node.
	// These are all LogicalPlans and LogicalExprs that are
	// referenced by the node.
	Children() []LogicalNode
	// Plans returns all the logical plans that are referenced
	// by the node (or the nearest node that contains them).
	Plans() []LogicalPlan
	// Accept is used to traverse the node.
	Accept(Visitor) any
	// Equal is used to compare two nodes.
	Equal(other Traversable) bool
}

// ScanSource is a source of data that a Scan can be performed on.
// This is either a physical table, a procedure call that returns a table,
// or a subquery.
type ScanSource interface {
	Traversable
	FormatScan() string
	Relation() *Relation
}

// TableScanSource represents a scan of a physical table or a CTE.
type TableScanSource struct {
	TableName string
	Type      TableSourceType

	// rel is the relation that the table scan source represents.
	// It is set during the evaluation phase.
	rel *Relation
}

func (s *TableScanSource) Accept(v Visitor) any {
	return v.VisitTableScanSource(s)
}
func (t *TableScanSource) Children() []LogicalNode {
	return nil
}

func (t *TableScanSource) FormatScan() string {
	return t.TableName + " [" + t.Type.String() + "]"
}

func (t *TableScanSource) Plans() []LogicalPlan {
	return nil
}

func (t *TableScanSource) Relation() *Relation {
	// we will copy it since this is meant to be re-useable,
	// but some callers may modify it
	return t.rel.Copy()
}

func (t *TableScanSource) Equal(other Traversable) bool {
	o, ok := other.(*TableScanSource)
	if !ok {
		return false
	}

	return t.TableName == o.TableName && t.Type == o.Type
}

type TableSourceType int

const (
	TableSourcePhysical TableSourceType = iota // physical table (default)
	TableSourceCTE                             // common table expression
	// if/when we support views, we will add a view type here
)

func (t TableSourceType) String() string {
	switch t {
	case TableSourcePhysical:
		return "physical"
	case TableSourceCTE:
		return "cte"
	default:
		panic(fmt.Sprintf("unknown table source type %d", t))
	}
}

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
	// rel is the relation that the procedure scan source represents.
	// It is set during the evaluation phase.
	rel *Relation
}

func (s *ProcedureScanSource) Accept(v Visitor) any {
	return v.VisitProcedureScanSource(s)
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

func (f *ProcedureScanSource) FormatScan() string {
	str := strings.Builder{}
	str.WriteString("[foreign=")
	str.WriteString(strconv.FormatBool(f.IsForeign))
	str.WriteString("] ")
	if f.IsForeign {
		str.WriteString("[dbid=")
		str.WriteString(f.ContextualArgs[0].String())
		str.WriteString("] ")
		str.WriteString("[proc=")
		str.WriteString(f.ContextualArgs[1].String())
		str.WriteString("] ")
	}
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

func (f *ProcedureScanSource) Plans() []LogicalPlan {
	var plans []LogicalPlan

	for _, arg := range f.Args {
		plans = append(plans, arg.Plans()...)
	}

	for _, arg := range f.ContextualArgs {
		plans = append(plans, arg.Plans()...)
	}

	return plans
}

func (f *ProcedureScanSource) Relation() *Relation {
	return f.rel.Copy()
}

func (f *ProcedureScanSource) Equal(other Traversable) bool {
	o, ok := other.(*ProcedureScanSource)
	if !ok {
		return false
	}

	if f.ProcedureName != o.ProcedureName || f.IsForeign != o.IsForeign {
		return false
	}

	if len(f.Args) != len(o.Args) || len(f.ContextualArgs) != len(o.ContextualArgs) {
		return false
	}

	for i, arg := range f.Args {
		if !arg.Equal(o.Args[i]) {
			return false
		}
	}

	for i, arg := range f.ContextualArgs {
		if !arg.Equal(o.ContextualArgs[i]) {
			return false
		}
	}

	return true
}

// Subquery holds information for a subquery.
// It is a TableSource that can be used in a Scan, but
// also can be used within expressions. If ReturnsRelation
// is true, it is a TableSource. If false, it is a scalar
// subquery (used in expressions).
type Subquery struct {
	// ReturnsRelation is true if the subquery returns an entire relation.
	// If false, the subquery returns a single value.
	ReturnsRelation bool
	// Plan is the logical plan for the subquery.
	Plan *Subplan

	// Everything below this is set during evaluation.

	// Correlated is the list of columns that are correlated
	// to the outer query. If empty, the subquery is uncorrelated.
	// TODO: we need to revisit this, because expressions in result sets can be correlated
	Correlated []*ColumnRef
}

func (s *Subquery) Accept(v Visitor) any {
	return v.VisitSubquery(s)
}
func (s *Subquery) Children() []LogicalNode {
	c := []LogicalNode{s.Plan}
	for _, col := range s.Correlated {
		c = append(c, col)
	}

	return c
}

func (s *Subquery) FormatScan() string {
	return ""
}

func (s *Subquery) Plans() []LogicalPlan {
	var plans []LogicalPlan
	plans = append(plans, s.Plan)
	for _, col := range s.Correlated {
		plans = append(plans, col.Plans()...)
	}

	return plans
}

func (s *Subquery) Relation() *Relation {
	return s.Plan.Relation()
}

func (s *Subquery) Equal(other Traversable) bool {
	o, ok := other.(*Subquery)
	if !ok {
		return false
	}

	// we want to check all of the fields without an
	// Equal method first, to try to exit early

	if s.ReturnsRelation != o.ReturnsRelation {
		return false
	}

	if len(s.Correlated) != len(o.Correlated) {
		return false
	}

	if !s.Plan.Equal(o.Plan) {
		return false
	}

	for i, col := range s.Correlated {
		if !col.Equal(o.Correlated[i]) {
			return false
		}
	}

	return true
}

type LogicalNode interface {
	fmt.Stringer
	Traversable
}

type LogicalPlan interface {
	LogicalNode
	Relation() *Relation
	plan()
}

type baseLogicalPlan struct{}

func (b *baseLogicalPlan) plan() {}

type EmptyScan struct {
	baseLogicalPlan
}

func (s *EmptyScan) Accept(v Visitor) any {
	return v.VisitEmptyScan(s)
}
func (n *EmptyScan) Children() []LogicalNode {
	return nil
}

func (n *EmptyScan) Plans() []LogicalPlan {
	return nil
}

func (n *EmptyScan) String() string {
	return "Empty Scan"
}

func (n *EmptyScan) Relation() *Relation {
	return &Relation{}
}

func (n *EmptyScan) Equal(other Traversable) bool {
	_, ok := other.(*EmptyScan)
	return ok
}

type Scan struct {
	baseLogicalPlan
	Source ScanSource
	// RelationName will always be set.
	// If the scan is a table scan and no alias was specified,
	// the RelationName will be the table name.
	// All other scan types (functions and subqueries) require an alias.
	RelationName string
}

func (s *Scan) Accept(v Visitor) any {
	return v.VisitScan(s)
}
func (s *Scan) Children() []LogicalNode {
	return s.Source.Children()
}

func (s *Scan) Plans() []LogicalPlan {
	return s.Source.Plans()
}

func (s *Scan) String() string {
	end := fmt.Sprintf(` [alias="%s"]:`, s.RelationName)
	switch t := s.Source.(type) {
	case *TableScanSource:
		// if relation name == table name, remove the alias
		if t.TableName == s.RelationName {
			end = ":"
		}

		return fmt.Sprintf(`Scan Table%s %s`, end, s.Source.FormatScan())
	case *ProcedureScanSource:
		return fmt.Sprintf(`Scan Procedure%s %s`, end, s.Source.FormatScan())
	case *Subquery:
		str := fmt.Sprintf(`Scan Subquery%s [subplan_id=%s]`, end, t.Plan.ID)
		if len(t.Correlated) > 0 {
			str += " (correlated: "
			for i, col := range t.Correlated {
				if i > 0 {
					str += ", "
				}
				str += col.String()
			}
			str += ")"
		} else {
			str += " (uncorrelated)"
		}
		return str
	default:
		panic(fmt.Sprintf("unknown scan source type %T", s.Source))
	}
}

func (s *Scan) Relation() *Relation {
	rel := s.Source.Relation()

	for _, col := range rel.Fields {
		col.Parent = s.RelationName
	}

	return rel
}

func (s *Scan) Equal(other Traversable) bool {
	o, ok := other.(*Scan)
	if !ok {
		return false
	}

	if s.RelationName != o.RelationName {
		return false
	}

	return s.Source.Equal(o.Source)
}

type Project struct {
	baseLogicalPlan

	// ! Expressions aren't set until the evaluation phase,
	// by the expandFuncs.

	// Expressions are the expressions that are projected.
	Expressions []LogicalExpr
	Child       LogicalPlan
	// expandFuncs are functions that adds to the list of expressions.
	// It is set while visiting the parse AST, and should be called during
	// the evaluation phase. It is used to expand wildcards like "SELECT *",
	// which can only be done during evaluation when we know the full relation.
	expandFuncs []expandFunc
}

func (s *Project) Accept(v Visitor) any {
	return v.VisitProject(s)
}
func (p *Project) Children() []LogicalNode {
	var c []LogicalNode
	for _, expr := range p.Expressions {
		c = append(c, expr)
	}
	c = append(c, p.Child)

	return c
}

func (p *Project) Plans() []LogicalPlan {
	c := []LogicalPlan{p.Child}
	for _, expr := range p.Expressions {
		c = append(c, expr.Plans()...)
	}

	return c
}

func (p *Project) String() string {
	str := strings.Builder{}
	str.WriteString("Projection: ")

	for i, expr := range p.Expressions {
		if i > 0 {
			str.WriteString("; ")
		}
		str.WriteString(expr.String())
	}

	return str.String()
}

func (p *Project) Relation() *Relation {
	var fields []*Field

	for _, expr := range p.Expressions {
		fields = append(fields, expr.Field())
	}

	return &Relation{Fields: fields}
}

func (p *Project) Equal(other Traversable) bool {
	o, ok := other.(*Project)
	if !ok {
		return false
	}

	if len(p.Expressions) != len(o.Expressions) {
		return false
	}

	for i, expr := range p.Expressions {
		if !expr.Equal(o.Expressions[i]) {
			return false
		}
	}

	return p.Child.Equal(o.Child)
}

type Filter struct {
	baseLogicalPlan
	Condition LogicalExpr
	Child     LogicalPlan
}

func (s *Filter) Accept(v Visitor) any {
	return v.VisitFilter(s)
}
func (f *Filter) Children() []LogicalNode {
	return []LogicalNode{f.Child, f.Condition}
}

func (f *Filter) String() string {
	return fmt.Sprintf("Filter: %s", f.Condition.String())
}

func (f *Filter) Plans() []LogicalPlan {
	return append([]LogicalPlan{f.Child}, f.Condition.Plans()...)
}

func (f *Filter) Relation() *Relation {
	return f.Child.Relation()
}

func (f *Filter) Equal(other Traversable) bool {
	o, ok := other.(*Filter)
	if !ok {
		return false
	}

	return f.Condition.Equal(o.Condition) && f.Child.Equal(o.Child)
}

type Join struct {
	baseLogicalPlan
	Left      LogicalPlan
	Right     LogicalPlan
	JoinType  JoinType
	Condition LogicalExpr
}

func (s *Join) Accept(v Visitor) any {
	return v.VisitJoin(s)
}
func (j *Join) Children() []LogicalNode {
	return []LogicalNode{j.Left, j.Right, j.Condition}
}

func (j *Join) String() string {
	str := strings.Builder{}
	str.WriteString("Join [")
	str.WriteString(j.JoinType.String())
	str.WriteString("]: ")

	str.WriteString(j.Condition.String())
	return str.String()
}

func (j *Join) Plans() []LogicalPlan {
	return append([]LogicalPlan{j.Left, j.Right}, j.Condition.Plans()...)
}

func (j *Join) Relation() *Relation {
	left := j.Left.Relation()
	right := j.Right.Relation()

	return &Relation{
		Fields: append(left.Fields, right.Fields...),
	}
}

func (j *Join) Equal(other Traversable) bool {
	o, ok := other.(*Join)
	if !ok {
		return false
	}

	if j.JoinType != o.JoinType {
		return false
	}

	if !j.Condition.Equal(o.Condition) {
		return false
	}

	return j.Left.Equal(o.Left) && j.Right.Equal(o.Right)
}

type Sort struct {
	baseLogicalPlan
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
	str.WriteString("Sort:")
	for i, sortExpr := range s.SortExpressions {
		if i > 0 {
			str.WriteString(";")
		}

		str.WriteString(" [")
		str.WriteString(sortExpr.Expr.String())
		str.WriteString("] ")
		if sortExpr.Ascending {
			str.WriteString("asc ")
		} else {
			str.WriteString("desc ")
		}
		if sortExpr.NullsLast {
			str.WriteString("nulls last")
		} else {
			str.WriteString("nulls first")
		}
	}
	return str.String()
}

func (s *Sort) Accept(v Visitor) any {
	return v.VisitSort(s)
}
func (s *Sort) Children() []LogicalNode {
	var c []LogicalNode
	for _, sortExpr := range s.SortExpressions {
		c = append(c, sortExpr.Expr)
	}
	c = append(c, s.Child)

	return c
}

func (s *Sort) Plans() []LogicalPlan {
	c := []LogicalPlan{s.Child}
	for _, sortExpr := range s.SortExpressions {
		c = append(c, sortExpr.Expr.Plans()...)
	}

	return c
}

func (s *Sort) Relation() *Relation {
	return s.Child.Relation()
}

func (s *Sort) Equal(other Traversable) bool {
	o, ok := other.(*Sort)
	if !ok {
		return false
	}

	if len(s.SortExpressions) != len(o.SortExpressions) {
		return false
	}

	for i, sortExpr := range s.SortExpressions {
		if sortExpr.Ascending != o.SortExpressions[i].Ascending {
			return false
		}

		if sortExpr.NullsLast != o.SortExpressions[i].NullsLast {
			return false
		}

		if !sortExpr.Expr.Equal(o.SortExpressions[i].Expr) {
			return false
		}
	}

	return s.Child.Equal(o.Child)
}

type Limit struct {
	baseLogicalPlan
	Child  LogicalPlan
	Limit  LogicalExpr
	Offset LogicalExpr
}

func (s *Limit) Accept(v Visitor) any {
	return v.VisitLimit(s)
}
func (l *Limit) Children() []LogicalNode {
	return []LogicalNode{l.Child, l.Limit, l.Offset}
}

func (l *Limit) String() string {
	str := strings.Builder{}
	str.WriteString("Limit")
	if l.Offset != nil {
		str.WriteString(" [offset=")
		str.WriteString(l.Offset.String())
		str.WriteString("]")
	}
	str.WriteString(": ")
	str.WriteString(l.Limit.String())

	return str.String()
}

func (l *Limit) Plans() []LogicalPlan {
	c := []LogicalPlan{l.Child}
	if l.Limit != nil {
		c = append(c, l.Limit.Plans()...)
	}
	if l.Offset != nil {
		c = append(c, l.Offset.Plans()...)
	}
	return c
}

func (l *Limit) Relation() *Relation {
	return l.Child.Relation()
}

func (l *Limit) Equal(other Traversable) bool {
	o, ok := other.(*Limit)
	if !ok {
		return false
	}

	if l.Limit != nil && !l.Limit.Equal(o.Limit) {
		return false
	}

	if l.Offset != nil && !l.Offset.Equal(o.Offset) {
		return false
	}

	return l.Child.Equal(o.Child)
}

type Distinct struct {
	baseLogicalPlan
	Child LogicalPlan
}

func (s *Distinct) Accept(v Visitor) any {
	return v.VisitDistinct(s)
}
func (d *Distinct) Children() []LogicalNode {
	return []LogicalNode{d.Child}
}

func (d *Distinct) String() string {
	return "Distinct"
}

func (d *Distinct) Plans() []LogicalPlan {
	return []LogicalPlan{d.Child}
}

func (d *Distinct) Relation() *Relation {
	return d.Child.Relation()
}

func (d *Distinct) Equal(other Traversable) bool {
	_, ok := other.(*Distinct)
	if !ok {
		return false
	}

	return d.Child.Equal(other)
}

type SetOperation struct {
	baseLogicalPlan
	Left   LogicalPlan
	Right  LogicalPlan
	OpType SetOperationType
}

// SetOperation
func (s *SetOperation) Accept(v Visitor) any {
	return v.VisitSetOperation(s)
}
func (s *SetOperation) Children() []LogicalNode {
	return []LogicalNode{s.Left, s.Right}
}

func (s *SetOperation) Plans() []LogicalPlan {
	return []LogicalPlan{s.Left, s.Right}
}

func (s *SetOperation) String() string {
	str := strings.Builder{}
	str.WriteString("Set: ")
	str.WriteString(s.OpType.String())
	return str.String()
}

func (s *SetOperation) Relation() *Relation {
	left := s.Left.Relation()
	right := s.Right.Relation()

	for _, field := range left.Fields {
		field.Parent = ""
	}
	for _, field := range right.Fields {
		field.Parent = ""
	}

	return &Relation{
		Fields: append(left.Fields, right.Fields...),
	}
}

func (s *SetOperation) Equal(other Traversable) bool {
	o, ok := other.(*SetOperation)
	if !ok {
		return false
	}

	return s.OpType == o.OpType && s.Left.Equal(o.Left) && s.Right.Equal(o.Right)
}

type Aggregate struct {
	baseLogicalPlan
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

func (s *Aggregate) Accept(v Visitor) any {
	return v.VisitAggregate(s)
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

func (a *Aggregate) Plans() []LogicalPlan {
	c := []LogicalPlan{a.Child}
	for _, expr := range a.GroupingExpressions {
		c = append(c, expr.Plans()...)
	}

	for _, expr := range a.AggregateExpressions {
		c = append(c, expr.Plans()...)
	}

	return c
}

func (a *Aggregate) String() string {
	str := strings.Builder{}
	str.WriteString("Aggregate")

	for _, expr := range a.GroupingExpressions {
		str.WriteString(" [")
		str.WriteString(expr.String())
		str.WriteString("]")
	}
	if len(a.AggregateExpressions) == 0 {
		return str.String()
	}

	str.WriteString(": ")

	for i, expr := range a.AggregateExpressions {
		if i > 0 {
			str.WriteString("; ")
		}
		str.WriteString(expr.String())
	}

	return str.String()
}

func (a *Aggregate) Relation() *Relation {
	// we return the grouping expressions and the aggregate expressions
	var fields []*Field
	for _, expr := range a.GroupingExpressions {
		fields = append(fields, expr.Field())
	}

	for _, expr := range a.AggregateExpressions {
		fields = append(fields, expr.Field())
	}

	return &Relation{Fields: fields}
}

func (a *Aggregate) Equal(other Traversable) bool {
	o, ok := other.(*Aggregate)
	if !ok {
		return false
	}

	if len(a.GroupingExpressions) != len(o.GroupingExpressions) {
		return false
	}

	if len(a.AggregateExpressions) != len(o.AggregateExpressions) {
		return false
	}

	for i, expr := range a.GroupingExpressions {
		if !expr.Equal(o.GroupingExpressions[i]) {
			return false
		}
	}

	for i, expr := range a.AggregateExpressions {
		if !expr.Equal(o.AggregateExpressions[i]) {
			return false
		}
	}

	return a.Child.Equal(o.Child)
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
		return "inner"
	case LeftOuterJoin:
		return "left"
	case RightOuterJoin:
		return "right"
	case FullOuterJoin:
		return "outer"
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
		return "union"
	case UnionAll:
		return "union all"
	case Intersect:
		return "intersect"
	case Except:
		return "except"
	default:
		panic(fmt.Sprintf("unknown set operation type %d", s))
	}
}

type Subplan struct {
	baseLogicalPlan
	Plan LogicalPlan
	ID   string
	Type SubplanType

	// extraInfo is used to print additional information about the subplan.
	// It shouldn't be used for planning purposes, and should only be used
	// for debugging.
	extraInfo string
}

func (s *Subplan) Accept(v Visitor) any {
	return v.VisitSubplan(s)
}
func (s *Subplan) Children() []LogicalNode {
	return []LogicalNode{s.Plan}
}

func (s *Subplan) Plans() []LogicalPlan {
	return []LogicalPlan{s.Plan}
}

func (s *Subplan) String() string {
	str := strings.Builder{}
	str.WriteString("Subplan [")
	str.WriteString(s.Type.String())
	str.WriteString("] [id=")
	str.WriteString(s.ID)
	str.WriteString("]")
	str.WriteString(s.extraInfo)

	return str.String()
}

func (s *Subplan) Relation() *Relation {
	return s.Plan.Relation()
}

func (s *Subplan) Equal(other Traversable) bool {
	o, ok := other.(*Subplan)
	if !ok {
		return false
	}

	return s.ID == o.ID && s.Type == o.Type && s.Plan.Equal(o.Plan)
}

type SubplanType int

const (
	SubplanTypeSubquery SubplanType = iota
	SubplanTypeCTE
)

func (s SubplanType) String() string {
	switch s {
	case SubplanTypeSubquery:
		return "subquery"
	case SubplanTypeCTE:
		return "cte"
	default:
		panic(fmt.Sprintf("unknown subplan type %d", s))
	}
}

// CartesianProduct is a logical plan node that represents a cartesian product
// between two relations. This is used in joins and set operations.
// Kwil doesn't actually allow cartesian products to be executed, but it is a necessary
// intermediate step for planning complex updates and deletes.
type CartesianProduct struct {
	baseLogicalPlan
	Left  LogicalPlan
	Right LogicalPlan
}

func (c *CartesianProduct) String() string {
	return "Cartesian Product"
}

func (s *CartesianProduct) Accept(v Visitor) any {
	return v.VisitCartesianProduct(s)
}
func (c *CartesianProduct) Children() []LogicalNode {
	return []LogicalNode{c.Left, c.Right}
}

func (c *CartesianProduct) Plans() []LogicalPlan {
	return []LogicalPlan{c.Left, c.Right}
}

func (c *CartesianProduct) Relation() *Relation {
	// we return the relation of the left side of the cartesian product
	return &Relation{
		Fields: append(c.Left.Relation().Fields, c.Right.Relation().Fields...),
	}
}

func (c *CartesianProduct) Equal(other Traversable) bool {
	o, ok := other.(*CartesianProduct)
	if !ok {
		return false
	}

	return c.Left.Equal(o.Left) && c.Right.Equal(o.Right)
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
	Field() *Field
}

// Literal value
type Literal struct {
	Value interface{}
	Type  *types.DataType
}

func (l *Literal) String() string {
	switch c := l.Value.(type) {
	case string:
		return "'" + c + "'"
	case []byte:
		return "0x" + hex.EncodeToString(c)
	default:
		return fmt.Sprintf("%v", l.Value)
	}
}

func (s *Literal) Accept(v Visitor) any {
	return v.VisitLiteral(s)
}
func (l *Literal) Children() []LogicalNode {
	return nil
}

func (l *Literal) Plans() []LogicalPlan {
	return nil
}

func (l *Literal) Field() *Field {
	return anonField(l.Type.Copy())
}

func (l *Literal) Equal(other Traversable) bool {
	o, ok := other.(*Literal)
	if !ok {
		return false
	}

	if !l.Type.EqualsStrict(o.Type) {
		return false
	}

	typeOf := reflect.TypeOf(l.Value)
	typeOfOther := reflect.TypeOf(o.Value)

	if typeOf.Kind() != typeOfOther.Kind() {
		return false
	}

	if typeOf.Kind() == reflect.Ptr {
		return reflect.ValueOf(l.Value).Pointer() == reflect.ValueOf(o.Value).Pointer()
	}

	return l.Value == o.Value
}

// Variable reference
type Variable struct {
	// name is something like $id, @caller, etc.
	VarName string
	// dataType is the data type, which is detected
	// during the evaluation phase.
	// it is either a *types.DataType or map[string]*types.DataType
	dataType any
}

func (v *Variable) String() string {
	return v.VarName
}

func (s *Variable) Accept(v Visitor) any {
	return v.VisitVariable(s)
}
func (v *Variable) Children() []LogicalNode {
	return nil
}

func (v *Variable) Plans() []LogicalPlan {
	return nil
}

func (v *Variable) Field() *Field {
	return &Field{
		val: v.dataType,
	}
}

func (v *Variable) Equal(other Traversable) bool {
	o, ok := other.(*Variable)
	if !ok {
		return false
	}

	return v.VarName == o.VarName
}

// Column reference
type ColumnRef struct {
	// Parent relation name, can be empty.
	// If not specified by user, it will be qualified
	// during the planning phase.
	Parent     string
	ColumnName string
	// dataType is the data type, which is detected
	// during the evaluation phase.
	dataType *types.DataType
}

func (c *ColumnRef) String() string {
	if c.Parent != "" {
		return fmt.Sprintf(`%s.%s`, c.Parent, c.ColumnName)
	}
	return c.ColumnName
}

func (s *ColumnRef) Accept(v Visitor) any {
	return v.VisitColumnRef(s)
}
func (c *ColumnRef) Children() []LogicalNode {
	return nil
}

func (c *ColumnRef) Plans() []LogicalPlan {
	return nil
}

func (c *ColumnRef) Field() *Field {
	return &Field{
		Parent: c.Parent,
		Name:   c.ColumnName,
		val:    c.dataType,
	}
}

func (c *ColumnRef) Equal(other Traversable) bool {
	o, ok := other.(*ColumnRef)
	if !ok {
		return false
	}

	return c.Parent == o.Parent && c.ColumnName == o.ColumnName
}

type AggregateFunctionCall struct {
	FunctionName string
	Args         []LogicalExpr
	Star         bool
	Distinct     bool
	// returnType is the data type of the return value.
	// It is set during the evaluation phase.
	returnType *types.DataType
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
				buf.WriteString("distinct ")
			}
			buf.WriteString(arg.String())
		}
	}
	buf.WriteString(")")
	return buf.String()
}

func (s *AggregateFunctionCall) Accept(v Visitor) any {
	return v.VisitAggregateFunctionCall(s)
}
func (a *AggregateFunctionCall) Children() []LogicalNode {
	var c []LogicalNode
	for _, arg := range a.Args {
		c = append(c, arg)
	}
	return c
}

func (a *AggregateFunctionCall) Plans() []LogicalPlan {
	var c []LogicalPlan
	for _, arg := range a.Args {
		c = append(c, arg.Plans()...)
	}
	return c
}

func (a *AggregateFunctionCall) Field() *Field {
	return anonField(a.returnType.Copy())
}

func (a *AggregateFunctionCall) Equal(other Traversable) bool {
	o, ok := other.(*AggregateFunctionCall)
	if !ok {
		return false
	}

	if a.FunctionName != o.FunctionName {
		return false
	}

	if a.Star != o.Star {
		return false
	}

	if a.Distinct != o.Distinct {
		return false
	}

	if len(a.Args) != len(o.Args) {
		return false
	}

	for i, arg := range a.Args {
		if !arg.Equal(o.Args[i]) {
			return false
		}
	}

	return true
}

// Function call
type ScalarFunctionCall struct {
	FunctionName string
	Args         []LogicalExpr
	// returnType is the data type of the return value.
	// It is set during the evaluation phase.
	returnType *types.DataType
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

func (s *ScalarFunctionCall) Accept(v Visitor) any {
	return v.VisitScalarFunctionCall(s)
}
func (f *ScalarFunctionCall) Children() []LogicalNode {
	var c []LogicalNode
	for _, arg := range f.Args {
		c = append(c, arg)
	}
	return c
}

func (f *ScalarFunctionCall) Plans() []LogicalPlan {
	var c []LogicalPlan
	for _, arg := range f.Args {
		c = append(c, arg.Plans()...)
	}
	return c
}

func (f *ScalarFunctionCall) Field() *Field {
	return anonField(f.returnType.Copy())
}

func (f *ScalarFunctionCall) Equal(other Traversable) bool {
	o, ok := other.(*ScalarFunctionCall)
	if !ok {
		return false
	}

	if f.FunctionName != o.FunctionName {
		return false
	}

	if len(f.Args) != len(o.Args) {
		return false
	}

	for i, arg := range f.Args {
		if !arg.Equal(o.Args[i]) {
			return false
		}
	}

	return true
}

// ProcedureCall is a call to a procedure.
// This can be a call to either a procedure in the same schema, or
// to a foreign procedure.
type ProcedureCall struct {
	ProcedureName string
	Foreign       bool
	Args          []LogicalExpr
	ContextArgs   []LogicalExpr
	// returnType is the data type of the return value.
	// It is set during the evaluation phase.
	returnType *types.DataType
}

func (p *ProcedureCall) String() string {
	var buf bytes.Buffer
	buf.WriteString(p.ProcedureName)
	if p.Foreign {
		buf.WriteString("[")
		buf.WriteString(p.ContextArgs[0].String())
		buf.WriteString(", ")
		buf.WriteString(p.ContextArgs[1].String())
		buf.WriteString("]")
	}
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

func (s *ProcedureCall) Accept(v Visitor) any {
	return v.VisitProcedureCall(s)
}
func (p *ProcedureCall) Children() []LogicalNode {
	var c []LogicalNode
	for _, arg := range p.Args {
		c = append(c, arg)
	}
	return c
}

func (p *ProcedureCall) Plans() []LogicalPlan {
	var c []LogicalPlan
	for _, arg := range p.Args {
		c = append(c, arg.Plans()...)
	}
	return c
}

func (p *ProcedureCall) Field() *Field {
	return anonField(p.returnType.Copy())
}

func (p *ProcedureCall) Equal(other Traversable) bool {
	o, ok := other.(*ProcedureCall)
	if !ok {
		return false
	}

	if p.ProcedureName != o.ProcedureName {
		return false
	}

	if p.Foreign != o.Foreign {
		return false
	}

	if len(p.Args) != len(o.Args) {
		return false
	}

	for i, arg := range p.Args {
		if !arg.Equal(o.Args[i]) {
			return false
		}
	}

	return true
}

type ArithmeticOp struct {
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

	return fmt.Sprintf("%s %s %s", a.Left.String(), op, a.Right.String())
}

func (s *ArithmeticOp) Accept(v Visitor) any {
	return v.VisitArithmeticOp(s)
}
func (a *ArithmeticOp) Children() []LogicalNode {
	return []LogicalNode{a.Left, a.Right}
}

func (a *ArithmeticOp) Plans() []LogicalPlan {
	return append(a.Left.Plans(), a.Right.Plans()...)
}

func (a *ArithmeticOp) Field() *Field {
	scalar, err := a.Left.Field().Scalar()
	if err != nil {
		panic(err)
	}

	return anonField(scalar.Copy())
}

func (a *ArithmeticOp) Equal(other Traversable) bool {
	o, ok := other.(*ArithmeticOp)
	if !ok {
		return false
	}

	return a.Op == o.Op && a.Left.Equal(o.Left) && a.Right.Equal(o.Right)
}

type ComparisonOp struct {
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
	Is             // rarely sargable
	IsDistinctFrom // not sargable
	Like           // not sargable
	ILike          // not sargable
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
		return "IS"
	case IsDistinctFrom:
		return "IS DISTINCT FROM"
	case Like:
		return "LIKE"
	case ILike:
		return "ILIKE"
	default:
		panic(fmt.Sprintf("unknown comparison operator %d", c))
	}
}

func (s *ComparisonOp) Accept(v Visitor) any {
	return v.VisitComparisonOp(s)
}
func (c *ComparisonOp) Children() []LogicalNode {
	return []LogicalNode{c.Left, c.Right}
}

func (c *ComparisonOp) Plans() []LogicalPlan {
	return append(c.Left.Plans(), c.Right.Plans()...)
}

func (c *ComparisonOp) String() string {
	return fmt.Sprintf("%s %s %s", c.Left.String(), c.Op.String(), c.Right.String())
}

func (c *ComparisonOp) Field() *Field {
	return anonField(types.BoolType.Copy())
}

func (c *ComparisonOp) Equal(other Traversable) bool {
	o, ok := other.(*ComparisonOp)
	if !ok {
		return false
	}

	return c.Op == o.Op && c.Left.Equal(o.Left) && c.Right.Equal(o.Right)
}

type LogicalOp struct {
	Left  LogicalExpr
	Right LogicalExpr
	Op    LogicalOperator
}

type LogicalOperator uint8

const (
	And LogicalOperator = iota
	Or
)

func (l *LogicalOp) String() string {
	var op string
	switch l.Op {
	case And:
		op = "AND"
	case Or:
		op = "OR"
	}

	return fmt.Sprintf("%s %s %s", l.Left.String(), op, l.Right.String())
}

func (s *LogicalOp) Accept(v Visitor) any {
	return v.VisitLogicalOp(s)
}
func (l *LogicalOp) Children() []LogicalNode {
	return []LogicalNode{l.Left, l.Right}
}

func (l *LogicalOp) Plans() []LogicalPlan {
	return append(l.Left.Plans(), l.Right.Plans()...)
}

func (l *LogicalOp) Field() *Field {
	return anonField(types.BoolType.Copy())
}

func (l *LogicalOp) Equal(other Traversable) bool {
	o, ok := other.(*LogicalOp)
	if !ok {
		return false
	}

	return l.Op == o.Op && l.Left.Equal(o.Left) && l.Right.Equal(o.Right)
}

type UnaryOp struct {
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

func (s *UnaryOp) Accept(v Visitor) any {
	return v.VisitUnaryOp(s)
}
func (u *UnaryOp) Children() []LogicalNode {
	return []LogicalNode{u.Expr}
}

func (u *UnaryOp) Plans() []LogicalPlan {
	return u.Expr.Plans()
}

func (u *UnaryOp) Field() *Field {
	return u.Expr.Field()
}

func (u *UnaryOp) Equal(other Traversable) bool {
	o, ok := other.(*UnaryOp)
	if !ok {
		return false
	}

	return u.Op == o.Op && u.Expr.Equal(o.Expr)
}

type TypeCast struct {
	Expr LogicalExpr
	Type *types.DataType
}

func (t *TypeCast) String() string {
	return fmt.Sprintf("%s::%s", t.Expr.String(), t.Type.Name)
}

func (s *TypeCast) Accept(v Visitor) any {
	return v.VisitTypeCast(s)
}
func (t *TypeCast) Children() []LogicalNode {
	return []LogicalNode{t.Expr}
}

func (t *TypeCast) Plans() []LogicalPlan {
	return t.Expr.Plans()
}

func (t *TypeCast) Field() *Field {
	return anonField(t.Type.Copy())
}

func (t *TypeCast) Equal(other Traversable) bool {
	o, ok := other.(*TypeCast)
	if !ok {
		return false
	}

	return t.Type.EqualsStrict(o.Type) && t.Expr.Equal(o.Expr)
}

type AliasExpr struct {
	Expr  LogicalExpr
	Alias string
}

func (a *AliasExpr) String() string {
	return fmt.Sprintf("%s AS %s", a.Expr.String(), a.Alias)
}

func (s *AliasExpr) Accept(v Visitor) any {
	return v.VisitAliasExpr(s)
}
func (a *AliasExpr) Children() []LogicalNode {
	return []LogicalNode{a.Expr}
}

func (a *AliasExpr) Plans() []LogicalPlan {
	return a.Expr.Plans()
}

func (a *AliasExpr) Field() *Field {
	return a.Expr.Field()
}

func (a *AliasExpr) Equal(other Traversable) bool {
	o, ok := other.(*AliasExpr)
	if !ok {
		return false
	}

	return a.Alias == o.Alias && a.Expr.Equal(o.Expr)
}

type ArrayAccess struct {
	Array LogicalExpr
	Index LogicalExpr
}

func (a *ArrayAccess) String() string {
	return fmt.Sprintf("%s[%s]", a.Array.String(), a.Index.String())
}

func (s *ArrayAccess) Accept(v Visitor) any {
	return v.VisitArrayAccess(s)
}
func (a *ArrayAccess) Children() []LogicalNode {
	return []LogicalNode{a.Array, a.Index}
}

func (a *ArrayAccess) Plans() []LogicalPlan {
	return append(a.Array.Plans(), a.Index.Plans()...)
}

func (a *ArrayAccess) Field() *Field {
	scalar, err := a.Array.Field().Scalar()
	if err != nil {
		panic(err)
	}

	scalar.IsArray = false
	return anonField(scalar.Copy())
}

func (a *ArrayAccess) Equal(other Traversable) bool {
	o, ok := other.(*ArrayAccess)
	if !ok {
		return false
	}

	return a.Array.Equal(o.Array) && a.Index.Equal(o.Index)
}

type ArrayConstructor struct {
	Elements []LogicalExpr
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

func (s *ArrayConstructor) Accept(v Visitor) any {
	return v.VisitArrayConstructor(s)
}
func (a *ArrayConstructor) Children() []LogicalNode {
	var c []LogicalNode
	for _, elem := range a.Elements {
		c = append(c, elem)
	}
	return c
}

func (a *ArrayConstructor) Plans() []LogicalPlan {
	var c []LogicalPlan
	for _, elem := range a.Elements {
		c = append(c, elem.Plans()...)
	}
	return c
}

func (a *ArrayConstructor) Field() *Field {
	if len(a.Elements) == 0 {
		// should get caught several times before this point
		panic("empty array constructor")
	}

	scalar, err := a.Elements[0].Field().Scalar()
	if err != nil {
		panic(err)
	}

	scalar.IsArray = true
	return anonField(scalar.Copy())
}

func (a *ArrayConstructor) Equal(other Traversable) bool {
	o, ok := other.(*ArrayConstructor)
	if !ok {
		return false
	}

	if len(a.Elements) != len(o.Elements) {
		return false
	}

	for i, elem := range a.Elements {
		if !elem.Equal(o.Elements[i]) {
			return false
		}
	}

	return true
}

type FieldAccess struct {
	Object LogicalExpr
	Key    string
}

func (f *FieldAccess) String() string {
	return fmt.Sprintf("%s.%s", f.Object.String(), f.Key)
}

func (s *FieldAccess) Accept(v Visitor) any {
	return v.VisitFieldAccess(s)
}
func (f *FieldAccess) Children() []LogicalNode {
	return []LogicalNode{f.Object}
}

func (f *FieldAccess) Plans() []LogicalPlan {
	return f.Object.Plans()
}

func (f *FieldAccess) Field() *Field {
	scalar, err := f.Object.Field().Object()
	if err != nil {
		panic(err)
	}

	val, ok := scalar[f.Key]
	if !ok {
		panic(fmt.Sprintf("field %s not found in object", f.Key))
	}

	return anonField(val.Copy())
}

func (f *FieldAccess) Equal(other Traversable) bool {
	o, ok := other.(*FieldAccess)
	if !ok {
		return false
	}

	return f.Key == o.Key && f.Object.Equal(o.Object)
}

type SubqueryExpr struct {
	Query *Subquery
	// If Exists is true, we are checking if the subquery returns any rows.
	// Otherwise, the subquery will return a single value.
	// If the query is a NOT EXISTS, a unary negation will wrap this expression.
	Exists bool
}

var _ LogicalExpr = (*SubqueryExpr)(nil)

func (s *SubqueryExpr) String() string {
	str := strings.Builder{}
	str.WriteString("[subquery (")

	if s.Exists {
		str.WriteString("exists")
	} else {
		str.WriteString("scalar")
	}

	str.WriteString(") (subplan_id=")
	str.WriteString(s.Query.Plan.ID)
	str.WriteString(") ")

	if len(s.Query.Correlated) == 0 {
		str.WriteString("(uncorrelated)")
	} else {
		str.WriteString("(correlated: ")

		for i, field := range s.Query.Correlated {
			if i > 0 {
				str.WriteString(", ")
			}
			if field.Parent != "" {
				str.WriteString(field.Parent)
				str.WriteString(".")
			}

			str.WriteString(field.ColumnName)
		}
		str.WriteString(")")
	}

	str.WriteString("]")

	return str.String()
}

func (s *SubqueryExpr) Accept(v Visitor) any {
	return v.VisitSubqueryExpr(s)
}
func (s *SubqueryExpr) Children() []LogicalNode {
	return []LogicalNode{s.Query.Plan}
}

func (s *SubqueryExpr) Plans() []LogicalPlan {
	return s.Query.Plans()
}

func (s *SubqueryExpr) Field() *Field {
	if s.Exists {
		return anonField(types.BoolType.Copy())
	}

	return s.Query.Plan.Relation().Fields[0]
}

func (s *SubqueryExpr) Equal(other Traversable) bool {
	o, ok := other.(*SubqueryExpr)
	if !ok {
		return false
	}

	return s.Query.Plan.ID == o.Query.Plan.ID && s.Exists == o.Exists
}

type Collate struct {
	Expr      LogicalExpr
	Collation CollationType
}

func (c *Collate) String() string {
	return fmt.Sprintf("%s COLLATE %s", c.Expr.String(), c.Collation.String())
}

func (s *Collate) Accept(v Visitor) any {
	return v.VisitCollate(s)
}
func (c *Collate) Children() []LogicalNode {
	return []LogicalNode{c.Expr}
}

func (c *Collate) Plans() []LogicalPlan {
	return c.Expr.Plans()
}

func (c *Collate) Field() *Field {
	return c.Expr.Field()
}

func (c *Collate) Equal(other Traversable) bool {
	o, ok := other.(*Collate)
	if !ok {
		return false
	}

	return c.Collation == o.Collation && c.Expr.Equal(o.Expr)
}

type CollationType uint8

const (
	// NoCaseCollation is a collation that is case-insensitive.
	NoCaseCollation CollationType = iota
)

func (c CollationType) String() string {
	switch c {
	case NoCaseCollation:
		return "nocase"
	default:
		panic(fmt.Sprintf("unknown collation type %d", c))
	}
}

type IsIn struct {
	// Left is the expression that is being compared.
	Left LogicalExpr

	// IsIn can have either a list of expressions or a subquery.
	// Either Expressions or Subquery will be set, but not both.

	Expressions []LogicalExpr
	Subquery    *SubqueryExpr
}

func (i *IsIn) String() string {
	str := strings.Builder{}
	str.WriteString(i.Left.String())

	str.WriteString(" IN (")
	if i.Expressions != nil {
		for j, expr := range i.Expressions {
			if j > 0 {
				str.WriteString(", ")
			}
			str.WriteString(expr.String())
		}
	} else {
		str.WriteString(i.Subquery.String())
	}
	str.WriteString(")")

	return str.String()
}

func (s *IsIn) Accept(v Visitor) any {
	return v.VisitIsIn(s)
}
func (i *IsIn) Children() []LogicalNode {
	var c []LogicalNode
	c = append(c, i.Left)
	if i.Expressions != nil {
		for _, expr := range i.Expressions {
			c = append(c, expr)
		}
	} else {
		c = append(c, i.Subquery)
	}
	return c
}

func (i *IsIn) Plans() []LogicalPlan {
	c := i.Left.Plans()
	if i.Expressions != nil {
		for _, expr := range i.Expressions {
			c = append(c, expr.Plans()...)
		}
	} else {
		c = append(c, i.Subquery.Plans()...)
	}
	return c
}

func (i *IsIn) Field() *Field {
	return anonField(types.BoolType.Copy())
}

func (i *IsIn) Equal(other Traversable) bool {
	o, ok := other.(*IsIn)
	if !ok {
		return false
	}

	if !i.Left.Equal(o.Left) {
		return false
	}

	if len(i.Expressions) != len(o.Expressions) {
		return false
	}

	for j, expr := range i.Expressions {
		if !expr.Equal(o.Expressions[j]) {
			return false
		}
	}

	if i.Subquery != nil {
		if o.Subquery == nil {
			return false
		}
		return i.Subquery.Equal(o.Subquery)
	}

	return o.Subquery == nil
}

type Case struct {
	// Value is the value that is being compared.
	// Can be nil if there is no value to compare.
	Value LogicalExpr
	// WhenClauses are the list of when/then pairs.
	// The first element of each pair is the condition,
	// which must match the data type of Value. If Value
	// is nil, the condition must be a boolean.
	WhenClauses [][2]LogicalExpr
	// Else is the else clause. Can be nil.
	Else LogicalExpr
}

func (c *Case) String() string {
	str := strings.Builder{}
	str.WriteString("CASE")
	if c.Value != nil {
		str.WriteString(" [")
		str.WriteString(c.Value.String())
		str.WriteString("]")
	}
	for _, when := range c.WhenClauses {
		str.WriteString(" WHEN [")
		str.WriteString(when[0].String())
		str.WriteString("] THEN [")
		str.WriteString(when[1].String())
		str.WriteString("]")
	}
	if c.Else != nil {
		str.WriteString(" ELSE [")
		str.WriteString(c.Else.String())
		str.WriteString("]")
	}
	str.WriteString(" END")
	return str.String()
}

func (s *Case) Accept(v Visitor) any {
	return v.VisitCase(s)
}
func (c *Case) Children() []LogicalNode {
	var ch []LogicalNode
	if c.Value != nil {
		ch = append(ch, c.Value)
	}
	for _, when := range c.WhenClauses {
		ch = append(ch, when[0], when[1])
	}
	if c.Else != nil {
		ch = append(ch, c.Else)
	}
	return ch
}

func (c *Case) Plans() []LogicalPlan {
	var ch []LogicalPlan
	if c.Value != nil {
		ch = append(ch, c.Value.Plans()...)
	}
	for _, when := range c.WhenClauses {
		ch = append(ch, when[0].Plans()...)
		ch = append(ch, when[1].Plans()...)
	}
	if c.Else != nil {
		ch = append(ch, c.Else.Plans()...)
	}
	return ch
}

func (c *Case) Field() *Field {
	if c.Else != nil {
		return c.Else.Field()
	}

	if len(c.WhenClauses) == 0 {
		// should get caught before this point
		panic("case statement must have at least one when clause")
	}

	return c.WhenClauses[0][1].Field()
}

func (c *Case) Equal(other Traversable) bool {
	o, ok := other.(*Case)
	if !ok {
		return false
	}

	if c.Value != nil && o.Value != nil {
		if !c.Value.Equal(o.Value) {
			return false
		}
	} else if c.Value != nil || o.Value != nil {
		return false
	}

	if len(c.WhenClauses) != len(o.WhenClauses) {
		return false
	}

	for i, when := range c.WhenClauses {
		if !when[0].Equal(o.WhenClauses[i][0]) {
			return false
		}
		if !when[1].Equal(o.WhenClauses[i][1]) {
			return false
		}
	}

	if c.Else != nil && o.Else != nil {
		return c.Else.Equal(o.Else)
	}

	return c.Else == nil && o.Else == nil
}

// ExprRef is a reference to an expression in the query.
// It is used when a part of the query is referenced in another.
// For example, SELECT SUM(a) FROM table1 GROUP BY a HAVING SUM(a) > 10,
// both SUM(a) expressions are the same, and will reference the same ExprRef
// (which would actually occur in the Aggregate node).
type ExprRef struct {
	// Identified is the expression that is being referenced.
	// It is a pointer to the actual expression.
	Identified *IdentifiedExpr
}

func (e *ExprRef) String() string {
	return fmt.Sprintf(`{%s}`, e.Identified.ref())
}

func (e *ExprRef) Field() *Field {
	return e.Identified.Field()
}

func (s *ExprRef) Accept(v Visitor) any {
	return v.VisitExprRef(s)
}
func (e *ExprRef) Children() []LogicalNode {
	return []LogicalNode{e.Identified}
}

func (e *ExprRef) Plans() []LogicalPlan {
	return e.Identified.Plans()
}

func (e *ExprRef) Equal(other Traversable) bool {
	o, ok := other.(*ExprRef)
	if !ok {
		return false
	}

	return e.Identified.Equal(o.Identified)
}

// IdentifiedExpr is an expression that can be referenced.
type IdentifiedExpr struct {
	// ID is the unique identifier for the expression.
	ID string
	// Expr is the expression that is being identified.
	Expr LogicalExpr
}

func (i *IdentifiedExpr) String() string {
	return fmt.Sprintf(`{%s = %s}`, i.ref(), i.Expr.String())
}

func (i *IdentifiedExpr) Field() *Field {
	return i.Expr.Field()
}

func (s *IdentifiedExpr) Accept(v Visitor) any {
	return v.VisitIdentifiedExpr(s)
}
func (i *IdentifiedExpr) Children() []LogicalNode {
	return []LogicalNode{i.Expr}
}

func (i *IdentifiedExpr) Plans() []LogicalPlan {
	return i.Expr.Plans()
}

func (i *IdentifiedExpr) Equal(other Traversable) bool {
	o, ok := other.(*IdentifiedExpr)
	if !ok {
		return false
	}

	return i.ID == o.ID && i.Expr.Equal(o.Expr)
}

// ref prints the reference name for the identified expression.
func (i *IdentifiedExpr) ref() string {
	return fmt.Sprintf(`#ref(%s)`, i.ID)
}

// traverse traverses a logical plan in preorder.
// It will call the callback function for each node in the plan.
// If the callback function returns false, the traversal will not
// continue to the children of the node.
func traverse(node Traversable, callback func(node Traversable) bool) {
	if !callback(node) {
		return
	}
	for _, child := range node.Children() {
		traverse(child, callback)
	}
}

func Format(plan LogicalNode) string {
	str := strings.Builder{}
	inner, topLevel := innerFormat(plan, 0, []bool{})
	str.WriteString(inner)

	printSubplans(&str, topLevel)

	return str.String()
}

// printSubplans is a recursive function that prints the subplans
func printSubplans(str *strings.Builder, subplans []*Subplan) {
	for _, sub := range subplans {
		str.WriteString(sub.String())
		str.WriteString("\n")
		strs, subs := innerFormat(sub.Plan, 1, []bool{false})
		str.WriteString(strs)
		printSubplans(str, subs)
	}
}

// innerFormat is a function that allows us to give more complex
// formatting logic.
// It returns subplans that should be added to the top level.
func innerFormat(plan LogicalNode, count int, printLong []bool) (string, []*Subplan) {
	if sub, ok := plan.(*Subplan); ok {
		return "", []*Subplan{sub}
	}

	var msg strings.Builder
	for i := 0; i < count; i++ {
		if i == count-1 && len(printLong) > i && !printLong[i] {
			msg.WriteString("└─")
		} else if i == count-1 && len(printLong) > i && printLong[i] {
			msg.WriteString("├─")
		} else if len(printLong) > i && printLong[i] {
			msg.WriteString("│ ")
		} else {
			msg.WriteString("  ")
		}
	}
	msg.WriteString(plan.String())
	msg.WriteString("\n")
	var topLevel []*Subplan
	plans := plan.Plans()
	for i, child := range plans {
		showLong := true
		// if it is the last plan, or if the next plan is a subplan,
		// we should not show the long line
		if i == len(plans)-1 {
			showLong = false
		} else if _, ok := plans[i+1].(*Subplan); ok {
			showLong = false
		}

		str, children := innerFormat(child, count+1, append(printLong, showLong))
		msg.WriteString(str)
		topLevel = append(topLevel, children...)
	}
	return msg.String(), topLevel
}

/*
	###########################
	#                         #
	#      Top Level Ops      #
	#                         #
	###########################
*/

// TopLevel is a logical plan that is at the top level of a query.
type TopLevel interface {
	LogicalPlan
	topLevel()
}

type baseTopLevel struct {
	baseLogicalPlan
}

func (b *baseTopLevel) topLevel() {}

// Return is a node that plans a return operation. It specifies columns to
// return from a query. This is similar to a projection, however it cannot
// be optimized using pushdowns, since it is at the top level (and therefore
// user requested).
type Return struct {
	baseTopLevel
	// Fields are the fields to return.
	Fields []*Field
	// Child is the input to the return.
	Child LogicalPlan
}

// expandFuncs take a relation and return a list of expressions that should be added to the projection.
type expandFunc func(*Relation) []LogicalExpr

func (r *Return) String() string {
	str := strings.Builder{}
	str.WriteString("Return: ")

	for i, expr := range r.Fields {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(expr.String())
	}

	return str.String()
}

func (s *Return) Accept(v Visitor) any {
	return v.VisitReturn(s)
}
func (r *Return) Children() []LogicalNode {
	return []LogicalNode{r.Child}
}

func (r *Return) Plans() []LogicalPlan {
	return []LogicalPlan{r.Child}
}

// Relation returns the relation of the child.
func (r *Return) Relation() *Relation {
	return r.Child.Relation()
}

func (r *Return) Equal(t Traversable) bool {
	o, ok := t.(*Return)
	if !ok {
		return false
	}

	if len(r.Fields) != len(o.Fields) {
		return false
	}

	for i, field := range r.Fields {
		if field != o.Fields[i] {
			return false
		}
	}

	return r.Child.Equal(o.Child)
}

/*
	For modifying relations, we will use the following process:
	1. Materialize the source relation (exactly same as any SELECT).
	This is all relations specified in FROM and JOIN in UPDATE and DELETE.
	2. Produce a cartesian product of the target relation and the source relation.
	We will use the ModifiableProduct node to represent this, since we want to keep
	this logically separate from joins so that we can know which side of the join
	can be modified.
	3. Apply the filter to the cartesian product / ModifiableProduct.
*/

// Update is a node that plans an update operation.
type Update struct {
	baseTopLevel
	// Child is the input to the update.
	Child LogicalPlan
	// Table is the target table name.
	// It will always be the table name and not an alias.
	Table string
	// Assignments are the assignments to update.
	Assignments []*Assignment
}

func (u *Update) String() string {
	str := strings.Builder{}
	str.WriteString("Update [")
	str.WriteString(u.Table)
	str.WriteString("]: ")

	for i, assign := range u.Assignments {
		if i > 0 {
			str.WriteString("; ")
		}
		str.WriteString(assign.Column)
		str.WriteString(" = ")
		str.WriteString(assign.Value.String())
	}

	return str.String()
}

func (s *Update) Accept(v Visitor) any {
	return v.VisitUpdate(s)
}
func (u *Update) Children() []LogicalNode {
	var c []LogicalNode
	for _, assign := range u.Assignments {
		c = append(c, assign.Value)
	}
	c = append(c, u.Child)
	return c
}

func (u *Update) Plans() []LogicalPlan {
	var c []LogicalPlan
	for _, assign := range u.Assignments {
		c = append(c, assign.Value.Plans()...)
	}
	c = append(c, u.Child)
	return c
}

func (u *Update) Relation() *Relation {
	return &Relation{}
}

func (u *Update) Equal(t Traversable) bool {
	o, ok := t.(*Update)
	if !ok {
		return false
	}

	if u.Table != o.Table {
		return false
	}

	if len(u.Assignments) != len(o.Assignments) {
		return false
	}

	for i, assign := range u.Assignments {
		if assign.Column != o.Assignments[i].Column {
			return false
		}

		if !assign.Value.Equal(o.Assignments[i].Value) {
			return false
		}
	}

	return u.Child.Equal(o.Child)
}

// Delete is a node that plans a delete operation.
type Delete struct {
	baseTopLevel
	// Child is the input to the delete.
	Child LogicalPlan
	// Table is the target table name.
	// It will always be the table name and not an alias.
	Table string
}

func (d *Delete) String() string {
	return fmt.Sprintf("Delete [%s]", d.Table)
}

func (s *Delete) Accept(v Visitor) any {
	return v.VisitDelete(s)
}
func (d *Delete) Children() []LogicalNode {
	return []LogicalNode{d.Child}
}

func (d *Delete) Plans() []LogicalPlan {
	return []LogicalPlan{d.Child}
}

func (d *Delete) Relation() *Relation {
	return &Relation{}
}

func (d *Delete) Equal(t Traversable) bool {
	o, ok := t.(*Delete)
	if !ok {
		return false
	}

	if d.Table != o.Table {
		return false
	}

	return d.Child.Equal(o.Child)
}

// TODO: I dont love this insert. Everything else feels very relational, but this
// feels like it is too much about the ast still.
// The general complexity of both the AST visitor and the evaluation phase
// makes me feel uneasy.
// I really think it is the UPDATE conflict target that feels wrong to me, but
// I am not sure what to do.
// Will revisit tomorrow.

// Insert is a node that plans an insert operation.
type Insert struct {
	baseTopLevel
	// Table is the physical table to insert into.
	Table string
	// Alias is the alias of the table.
	// It can only be referenced in the conflict resolution.
	// It can be empty.
	Alias string
	// Values are the values to insert.
	// The length of each second dimensional slice in Values must be equal to all others.
	Values [][]LogicalExpr
	// ConflictResolution is the conflict resolution to use if there is a conflict.
	ConflictResolution ConflictResolution
}

func (i *Insert) String() string {
	str := strings.Builder{}
	str.WriteString("Insert [")
	str.WriteString(i.Table)
	str.WriteString("]")
	if i.Alias != "" {
		str.WriteString(" [alias=")
		str.WriteString(i.Alias)
		str.WriteString("]")
	}
	str.WriteString(": ")

	for i, val := range i.Values {
		if i > 0 {
			str.WriteString("; ")
		}
		str.WriteString("(")
		for j, v := range val {
			if j > 0 {
				str.WriteString(", ")
			}
			str.WriteString(v.String())
		}
		str.WriteString(")")
	}

	return str.String()
}

func (s *Insert) Accept(v Visitor) any {
	return v.VisitInsert(s)
}
func (i *Insert) Children() []LogicalNode {
	var c []LogicalNode
	for _, val := range i.Values {
		for _, v := range val {
			c = append(c, v)
		}
	}

	if i.ConflictResolution != nil {
		if con, ok := i.ConflictResolution.(*ConflictUpdate); ok {
			if con.ConflictFilter != nil {
				c = append(c, con.ConflictFilter)
			}

			for _, assign := range con.Assignments {
				c = append(c, assign.Value)
			}
		}
	}

	return c
}

func (i *Insert) Plans() []LogicalPlan {
	var c []LogicalPlan
	for _, val := range i.Values {
		for _, v := range val {
			c = append(c, v.Plans()...)
		}
	}

	if i.ConflictResolution != nil {
		if con, ok := i.ConflictResolution.(*ConflictUpdate); ok {
			if con.ConflictFilter != nil {
				c = append(c, con.ConflictFilter.Plans()...)
			}

			for _, assign := range con.Assignments {
				c = append(c, assign.Value.Plans()...)
			}
		}
	}

	return c
}

func (i *Insert) Relation() *Relation {
	return &Relation{}
}

func (i *Insert) Equal(t Traversable) bool {
	o, ok := t.(*Insert)
	if !ok {
		return false
	}

	if i.Table != o.Table || i.Alias != o.Alias {
		return false
	}

	if len(i.Values) != len(o.Values) {
		return false
	}

	for j, val := range i.Values {
		if len(val) != len(o.Values[j]) {
			return false
		}

		for k, v := range val {
			if !v.Equal(o.Values[j][k]) {
				return false
			}
		}
	}

	if i.ConflictResolution == nil && o.ConflictResolution == nil {
		return true
	}

	if i.ConflictResolution == nil || o.ConflictResolution == nil {
		return false
	}

	return i.ConflictResolution.Equal(o.ConflictResolution)
}

// Assignment is a struct that represents an assignment in an update statement.
type Assignment struct {
	// Column is the column to update.
	Column string
	// Value is the value to update the column to.
	Value LogicalExpr
}

type ConflictResolution interface {
	Equal(ConflictResolution) bool
	conflictResolution()
}

// ConflictDoNothing is a struct that represents the resolution of a conflict
// using DO NOTHING.
type ConflictDoNothing struct {
	// ArbiterIndex is the index to use to determine if there is a conflict.
	// If/when Kwil supports partial indexes, we will turn this into a list
	// of indexes. Can be nil when DO NOTHING is used.
	ArbiterIndex Index
}

func (c *ConflictDoNothing) conflictResolution() {}

func (c *ConflictDoNothing) Equal(other ConflictResolution) bool {
	o, ok := other.(*ConflictDoNothing)
	if !ok {
		return false
	}

	return c.ArbiterIndex.Equal(o.ArbiterIndex)
}

// ConflictUpdate is a struct that represents the resolution of a conflict
// using DO UPDATE SET.
type ConflictUpdate struct {
	// ArbiterIndex is the index to use to determine if there is a conflict.
	// If/when Kwil supports partial indexes, we will turn this into a list
	// of indexes. See: https://github.com/cockroachdb/cockroach/issues/53170
	// Cannot be nil when DO UPDATE is used.
	ArbiterIndex Index
	// Assignments are the expressions to update if there is a conflict.
	// Cannot be nil.
	Assignments []*Assignment
	// ConflictFilter is a predicate that allows us to selectively
	// update or raise an error if there is a conflict.
	// Can be nil.
	ConflictFilter LogicalExpr
}

func (c *ConflictUpdate) conflictResolution() {}

func (c *ConflictUpdate) Equal(other ConflictResolution) bool {
	o, ok := other.(*ConflictUpdate)
	if !ok {
		return false
	}

	if !c.ArbiterIndex.Equal(o.ArbiterIndex) {
		return false
	}

	if (c.ConflictFilter == nil) != (o.ConflictFilter == nil) {
		return false
	}

	if c.ConflictFilter != nil && o.ConflictFilter != nil {
		if !c.ConflictFilter.Equal(o.ConflictFilter) {
			return false
		}
	}

	if len(c.Assignments) != len(o.Assignments) {
		return false
	}

	for i, assign := range c.Assignments {
		if !assign.Value.Equal(o.Assignments[i].Value) {
			return false
		}
	}

	return true
}

// Index is an interface that represents an index.
// Since Kwil's internal catalog does not individually name
// all indexes (e.g. UNIQUE and PRIMARY column constraints),
// we use this interface to represent an index.
type Index interface {
	index()
	Equal(Index) bool
}

type IndexColumnConstraint struct {
	// Table is the physical table that the index is on.
	Table string
	// Column is the column that the constraint is on.
	Column string
	// ConstraintType is the type of constraint that the index is.
	ConstraintType IndexConstraintType
}

func (i *IndexColumnConstraint) index() {}

func (i *IndexColumnConstraint) Equal(other Index) bool {
	o, ok := other.(*IndexColumnConstraint)
	if !ok {
		return false
	}

	return i.Table == o.Table && i.Column == o.Column && i.ConstraintType == o.ConstraintType
}

type IndexConstraintType uint8

const (
	UniqueConstraintIndex IndexConstraintType = iota
	PrimaryKeyConstraintIndex
)

// IndexNamed is any index that is specified explicitly
// and has a referenceable name.
type IndexNamed struct {
	// Name is the name of the index.
	Name string
}

func (i *IndexNamed) index() {}

func (i *IndexNamed) Equal(other Index) bool {
	o, ok := other.(*IndexNamed)
	if !ok {
		return false
	}

	return i.Name == o.Name
}

type Visitor interface {
	VisitTableScanSource(*TableScanSource) any
	VisitProcedureScanSource(*ProcedureScanSource) any
	VisitSubquery(*Subquery) any
	VisitEmptyScan(*EmptyScan) any
	VisitScan(*Scan) any
	VisitProject(*Project) any
	VisitFilter(*Filter) any
	VisitJoin(*Join) any
	VisitSort(*Sort) any
	VisitLimit(*Limit) any
	VisitDistinct(*Distinct) any
	VisitSetOperation(*SetOperation) any
	VisitAggregate(*Aggregate) any
	VisitSubplan(*Subplan) any
	VisitLiteral(*Literal) any
	VisitVariable(*Variable) any
	VisitColumnRef(*ColumnRef) any
	VisitAggregateFunctionCall(*AggregateFunctionCall) any
	VisitScalarFunctionCall(*ScalarFunctionCall) any
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
	VisitSubqueryExpr(*SubqueryExpr) any
	VisitCollate(*Collate) any
	VisitIsIn(*IsIn) any
	VisitCase(*Case) any
	VisitExprRef(*ExprRef) any
	VisitIdentifiedExpr(*IdentifiedExpr) any
	VisitReturn(*Return) any
	VisitCartesianProduct(*CartesianProduct) any
	VisitUpdate(*Update) any
	VisitDelete(*Delete) any
	VisitInsert(*Insert) any
}
