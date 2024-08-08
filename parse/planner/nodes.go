package planner

import (
	"bytes"
	"encoding/hex"
	"fmt"
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
}

// ScanSource is a source of data that a Scan can be performed on.
// This is either a physical table, a procedure call that returns a table,
// or a subquery.
type ScanSource interface {
	Traversable
	FormatScan() string
}

// TableScanSource represents a scan of a physical table or a CTE.
type TableScanSource struct {
	TableName string
	Type      TableSourceType
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

// SubqueryScanSource represents a scan of a subquery.
// This is used, for example, in the query "SELECT * FROM (SELECT * FROM users) AS subquery".
type SubqueryScanSource struct {
	Subquery LogicalPlan
}

func (s *SubqueryScanSource) Children() []LogicalNode {
	return []LogicalNode{s.Subquery}
}

func (s *SubqueryScanSource) FormatScan() string {
	return ""
}

func (s *SubqueryScanSource) Plans() []LogicalPlan {
	return []LogicalPlan{s.Subquery}
}

type LogicalNode interface {
	fmt.Stringer
	Traversable
	// Subplan returns the children of the node that are
	// logical plans.
	Plans() []LogicalPlan
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
	return nil
}

func (n *EmptyScan) Plans() []LogicalPlan {
	return nil
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
	case *SubqueryScanSource:
		return fmt.Sprintf(`Scan Subquery%s`, end)
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

type Filter struct {
	baseLogicalPlan
	Condition LogicalExpr
	Child     LogicalPlan
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

type Limit struct {
	baseLogicalPlan
	Child  LogicalPlan
	Limit  LogicalExpr
	Offset LogicalExpr
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

type Distinct struct {
	baseLogicalPlan
	Child LogicalPlan
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

func (s *SetOperation) Plans() []LogicalPlan {
	return []LogicalPlan{s.Left, s.Right}
}

func (s *SetOperation) String() string {
	str := strings.Builder{}
	str.WriteString("Set: ")
	str.WriteString(s.OpType.String())
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
	str.WriteString(": ")

	for i, expr := range a.AggregateExpressions {
		if i > 0 {
			str.WriteString("; ")
		}
		str.WriteString(expr.String())
	}

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
	switch c := l.Value.(type) {
	case string:
		return "'" + c + "'"
	case []byte:
		return "0x" + hex.EncodeToString(c)
	default:
		return fmt.Sprintf("%v", l.Value)
	}
}

func (l *Literal) Children() []LogicalNode {
	return nil
}

func (l *Literal) Plans() []LogicalPlan {
	return nil
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
	return nil
}

func (v *Variable) Plans() []LogicalPlan {
	return nil
}

// Column reference
type ColumnRef struct {
	baseLogicalExpr
	// Parent relation name, can be empty.
	// If not specified by user, it will be qualified
	// during the planning phase.
	Parent     string
	ColumnName string
}

func (c *ColumnRef) String() string {
	if c.Parent != "" {
		return fmt.Sprintf(`%s.%s`, c.Parent, c.ColumnName)
	}
	return c.ColumnName
}

func (c *ColumnRef) Children() []LogicalNode {
	return nil
}

func (c *ColumnRef) Plans() []LogicalPlan {
	return nil
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
				buf.WriteString("distinct ")
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

func (a *AggregateFunctionCall) Plans() []LogicalPlan {
	var c []LogicalPlan
	for _, arg := range a.Args {
		c = append(c, arg.Plans()...)
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

func (f *ScalarFunctionCall) Plans() []LogicalPlan {
	var c []LogicalPlan
	for _, arg := range f.Args {
		c = append(c, arg.Plans()...)
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

	return fmt.Sprintf("%s %s %s", a.Left.String(), op, a.Right.String())
}

func (a *ArithmeticOp) Children() []LogicalNode {
	return []LogicalNode{a.Left, a.Right}
}

func (a *ArithmeticOp) Plans() []LogicalPlan {
	return append(a.Left.Plans(), a.Right.Plans()...)
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

func (c *ComparisonOp) Children() []LogicalNode {
	return []LogicalNode{c.Left, c.Right}
}

func (c *ComparisonOp) Plans() []LogicalPlan {
	return append(c.Left.Plans(), c.Right.Plans()...)
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

	return fmt.Sprintf("%s %s %s", l.Left.String(), op, l.Right.String())
}

func (l *LogicalOp) Children() []LogicalNode {
	return []LogicalNode{l.Left, l.Right}
}

func (l *LogicalOp) Plans() []LogicalPlan {
	return append(l.Left.Plans(), l.Right.Plans()...)
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

func (u *UnaryOp) Plans() []LogicalPlan {
	return u.Expr.Plans()
}

type TypeCast struct {
	baseLogicalExpr
	Expr LogicalExpr
	Type *types.DataType
}

func (t *TypeCast) String() string {
	return fmt.Sprintf("%s::%s", t.Expr.String(), t.Type.Name)
}

func (t *TypeCast) Children() []LogicalNode {
	return []LogicalNode{t.Expr}
}

func (t *TypeCast) Plans() []LogicalPlan {
	return t.Expr.Plans()
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

func (a *AliasExpr) Plans() []LogicalPlan {
	return a.Expr.Plans()
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

func (a *ArrayAccess) Plans() []LogicalPlan {
	return append(a.Array.Plans(), a.Index.Plans()...)
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

func (a *ArrayConstructor) Plans() []LogicalPlan {
	var c []LogicalPlan
	for _, elem := range a.Elements {
		c = append(c, elem.Plans()...)
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

func (f *FieldAccess) Plans() []LogicalPlan {
	return f.Object.Plans()
}

type Subquery struct {
	baseLogicalExpr
	SubqueryType SubqueryType
	Query        *Subplan
	// ID is the number of the subquery in the query.
	ID string
	// Correlated is the list of columns that are correlated
	// to the outer query. If empty, the subquery is uncorrelated.
	Correlated []*ColumnRef
}

var _ LogicalExpr = (*Subquery)(nil)

func (s *Subquery) String() string {
	str := strings.Builder{}
	str.WriteString("[subquery (")

	str.WriteString(s.SubqueryType.String())
	str.WriteString(") (subplan_id=")
	str.WriteString(s.ID)
	str.WriteString(") ")

	if len(s.Correlated) == 0 {
		str.WriteString("(uncorrelated)")
	} else {
		str.WriteString("(correlated: ")

		for i, field := range s.Correlated {
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

func (s *Subquery) Children() []LogicalNode {
	return []LogicalNode{s.Query}
}

func (s *Subquery) Plans() []LogicalPlan {
	return []LogicalPlan{s.Query}
}

type SubqueryType uint8

const (
	ScalarSubquery SubqueryType = iota
	ExistsSubquery
	NotExistsSubquery
)

func (s SubqueryType) String() string {
	switch s {
	case ScalarSubquery:
		return "scalar"
	case ExistsSubquery:
		return "exists"
	case NotExistsSubquery:
		return "not exists"
	default:
		panic(fmt.Sprintf("unknown subquery type %d", s))
	}
}

type Collate struct {
	baseLogicalExpr
	Expr      LogicalExpr
	Collation CollationType
}

func (c *Collate) String() string {
	return fmt.Sprintf("%s COLLATE %s", c.Expr.String(), c.Collation.String())
}

func (c *Collate) Children() []LogicalNode {
	return []LogicalNode{c.Expr}
}

func (c *Collate) Plans() []LogicalPlan {
	return c.Expr.Plans()
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
	baseLogicalExpr
	// Left is the expression that is being compared.
	Left LogicalExpr

	// IsIn can have either a list of expressions or a subquery.
	// Either Expressions or Subquery will be set, but not both.

	Expressions []LogicalExpr
	Subquery    *Subquery
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

type Case struct {
	baseLogicalExpr
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
