package planner

import (
	"bytes"
	"fmt"
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

func Format(plan LogicalNode, indent int) string {
	var msg strings.Builder
	for i := 0; i < indent; i++ {
		msg.WriteString(" ")
	}
	msg.WriteString(plan.String())
	msg.WriteString("\n")
	for _, child := range plan.Children() {
		if _, ok := child.(LogicalPlan); ok {
			msg.WriteString(Format(child, indent+2))
		}
	}
	return msg.String()
}

// ScanSource is a source of data that a Scan can be performed on.
// This is either a physical table, a procedure call that returns a table,
// or a subquery.
type ScanSource interface {
	Children() []LogicalNode
	FormatScan() string
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

func (t *TableScanSource) FormatScan() string {
	return t.TableName
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

func (f *ProcedureScanSource) FormatScan() string {
	str := strings.Builder{}
	if f.IsForeign {
		str.WriteString("[foreign] ")
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

func (s *SubqueryScanSource) FormatScan() string {
	return ""
}

type LogicalNode interface {
	fmt.Stringer
	Children() []LogicalNode
}

type LogicalPlan interface {
	LogicalNode
	plan()
}

type baseLogicalPlan struct{}

func (b *baseLogicalPlan) plan() {}

type EmptyScan struct {
	baseLogicalPlan
}

func (n *EmptyScan) Children() []LogicalNode {
	return []LogicalNode{}
}

func (f *EmptyScan) Accept(v Visitor) any {
	return v.VisitNoop(f)
}

func (n *EmptyScan) String() string {
	return "Empty Scan"
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

func (s *Scan) Children() []LogicalNode {
	return s.Source.Children()
}

func (f *Scan) Accept(v Visitor) any {
	return v.VisitScanAlias(f)
}

func (s *Scan) String() string {
	switch s.Source.(type) {
	case *TableScanSource:
		return fmt.Sprintf("Scan Table [alias=%s]: %s", s.RelationName, s.Source.FormatScan())
	case *ProcedureScanSource:
		return fmt.Sprintf("Scan Procedure [alias=%s]: %s", s.RelationName, s.Source.FormatScan())
	case *SubqueryScanSource:
		return fmt.Sprintf("Scan Subquery [alias=%s]: %s", s.RelationName, s.Source.FormatScan())
	default:
		panic(fmt.Sprintf("unknown scan source type %T", s.Source))
	}
}

type Project struct {
	baseLogicalPlan
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

func (p *Project) String() string {
	str := strings.Builder{}
	str.WriteString("Projection: ")

	for i, expr := range p.Expressions {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(expr.String())
	}

	return str.String()
}

type Filter struct {
	baseLogicalPlan
	Condition LogicalExpr
	Child     LogicalPlan
}

func (f *Filter) Children() []LogicalNode {
	return []LogicalNode{f.Child, f.Condition}
}

func (f *Filter) Accept(v Visitor) any {
	return v.VisitFilter(f)
}

func (f *Filter) String() string {
	return fmt.Sprintf("Filter: %s", f.Condition.String())
}

type Join struct {
	baseLogicalPlan
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

func (j *Join) String() string {
	str := strings.Builder{}
	str.WriteString(j.JoinType.String())
	str.WriteString(" Join: ")
	str.WriteString(j.Condition.String())
	return str.String()

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

type Limit struct {
	baseLogicalPlan
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
	baseLogicalPlan
	Child LogicalPlan
}

func (d *Distinct) Children() []LogicalNode {
	return []LogicalNode{d.Child}
}

func (f *Distinct) Accept(v Visitor) any {
	return v.VisitDistinct(f)
}

func (d *Distinct) String() string {
	return "DISTINCT"
}

type SetOperation struct {
	baseLogicalPlan
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
	baseLogicalPlan
	// GroupingExpressions are the expressions used
	// in the GROUP BY clause.
	GroupingExpressions []LogicalExpr
	// AggregateExpressions are the expressions used
	// in the SELECT clause (e.g. SUM(x), COUNT(y)).
	AggregateExpressions []*AggregateFunctionCall
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
		return "Inner"
	case LeftOuterJoin:
		return "Left Outer"
	case RightOuterJoin:
		return "Right Outer"
	case FullOuterJoin:
		return "Full Outer"
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
	expr()
}

type baseLogicalExpr struct{}

func (b *baseLogicalExpr) expr() {}

// Literal value
type Literal struct {
	baseLogicalExpr
	Value interface{}
	Type  *types.DataType
}

func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.Value)
}

func (l *Literal) Children() []LogicalNode {
	return []LogicalNode{}
}

// Variable reference
type Variable struct {
	baseLogicalExpr
	// name is something like $id, @caller, etc.
	VarName string
}

func (v *Variable) String() string {
	return v.VarName
}

func (v *Variable) Children() []LogicalNode {
	return []LogicalNode{}
}

// Column reference
type ColumnRef struct {
	baseLogicalExpr
	Parent     string // Parent relation name, can be empty
	ColumnName string
}

func (c *ColumnRef) String() string {
	if c.Parent != "" {
		return fmt.Sprintf("%s.%s", c.Parent, c.ColumnName)
	}
	return c.ColumnName
}

func (c *ColumnRef) Children() []LogicalNode {
	return []LogicalNode{}
}

type AggregateFunctionCall struct {
	baseLogicalExpr
	FunctionName string
	Args         []LogicalExpr
	Star         bool
	Distinct     bool
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

func (a *AggregateFunctionCall) Children() []LogicalNode {
	var c []LogicalNode
	for _, arg := range a.Args {
		c = append(c, arg)
	}
	return c
}

// Function call
type ScalarFunctionCall struct {
	baseLogicalExpr
	FunctionName string
	Args         []LogicalExpr
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
	baseLogicalExpr
	ProcedureName string
	Foreign       bool
	Args          []LogicalExpr
	ContextArgs   []LogicalExpr
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

func (p *ProcedureCall) Children() []LogicalNode {
	var c []LogicalNode
	for _, arg := range p.Args {
		c = append(c, arg)
	}
	return c
}

type ArithmeticOp struct {
	baseLogicalExpr
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

	return fmt.Sprintf("(%s %s %s)", a.Left.String(), op, a.Right.String())
}

func (a *ArithmeticOp) Children() []LogicalNode {
	return []LogicalNode{a.Left, a.Right}
}

type ComparisonOp struct {
	baseLogicalExpr
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

func (c *ComparisonOp) Children() []LogicalNode {
	return []LogicalNode{c.Left, c.Right}
}

func (c *ComparisonOp) String() string {
	return fmt.Sprintf("%s %s %s", c.Left.String(), c.Op.String(), c.Right.String())
}

type LogicalOp struct {
	baseLogicalExpr
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

	return fmt.Sprintf("(%s %s %s)", l.Left.String(), op, l.Right.String())
}

func (l *LogicalOp) Children() []LogicalNode {
	return []LogicalNode{l.Left, l.Right}
}

type UnaryOp struct {
	baseLogicalExpr
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

func (u *UnaryOp) Children() []LogicalNode {
	return []LogicalNode{u.Expr}
}

type TypeCast struct {
	baseLogicalExpr
	Expr LogicalExpr
	Type *types.DataType
}

func (t *TypeCast) String() string {
	return fmt.Sprintf("(%s::%s)", t.Expr.String(), t.Type.Name)
}

func (t *TypeCast) Children() []LogicalNode {
	return []LogicalNode{t.Expr}
}

type AliasExpr struct {
	baseLogicalExpr
	Expr  LogicalExpr
	Alias string
}

func (a *AliasExpr) String() string {
	return fmt.Sprintf("%s AS %s", a.Expr.String(), a.Alias)
}

func (a *AliasExpr) Children() []LogicalNode {
	return []LogicalNode{a.Expr}
}

type ArrayAccess struct {
	baseLogicalExpr
	Array LogicalExpr
	Index LogicalExpr
}

func (a *ArrayAccess) String() string {
	return fmt.Sprintf("%s[%s]", a.Array.String(), a.Index.String())
}

func (a *ArrayAccess) Children() []LogicalNode {
	return []LogicalNode{a.Array, a.Index}
}

type ArrayConstructor struct {
	baseLogicalExpr
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

func (a *ArrayConstructor) Children() []LogicalNode {
	var c []LogicalNode
	for _, elem := range a.Elements {
		c = append(c, elem)
	}
	return c
}

type FieldAccess struct {
	baseLogicalExpr
	Object LogicalExpr
	Field  string
}

func (f *FieldAccess) String() string {
	return fmt.Sprintf("%s.%s", f.Object.String(), f.Field)
}

func (f *FieldAccess) Children() []LogicalNode {
	return []LogicalNode{f.Object}
}

type Subquery struct {
	baseLogicalExpr
	SubqueryType SubqueryType
	Query        LogicalPlan
	// Correlated indicates whether the subquery is correlated.
	// It is set during planning
	Correlated bool
}

var _ LogicalExpr = (*Subquery)(nil)

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
	VisitNoop(*EmptyScan) any
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
