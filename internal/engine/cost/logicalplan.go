package cost

import (
	"fmt"
	"strings"
)

type scan struct {
	table      string
	dataSource DataSource
	//projection []string
}

func (s *scan) String() string {
	return fmt.Sprintf("Scan: %s", s.table)
}

func (s *scan) Schema() *schema {
	return s.dataSource.Schema()
}

func (s *scan) Inputs() []LogicalPlan {
	return nil
}

// Scan creates a table scan.
func Scan(table string, ds DataSource) LogicalPlan {
	return &scan{table: table, dataSource: ds}
}

type projection struct {
	input LogicalPlan
	exprs []LogicalExpr
}

func (p *projection) String() string {
	fields := make([]string, len(p.exprs))
	for i, expr := range p.exprs {
		fields[i] = expr.String()
	}
	return fmt.Sprintf("Projection: %s", strings.Join(fields, ", p"))
}

func (p *projection) Schema() *schema {
	fs := make([]Field, len(p.exprs))
	for i, expr := range p.exprs {
		fs[i] = expr.Resolve(p.input)
	}
	return Schema(fs...)
}

func (p *projection) Inputs() []LogicalPlan {
	return []LogicalPlan{p.input}
}

// Projection creates a projection.
func Projection(plan LogicalPlan, exprs ...LogicalExpr) LogicalPlan {
	return &projection{
		input: plan,
		exprs: exprs,
	}
}

type selection struct {
	input LogicalPlan
	expr  []LogicalExpr // here we break to individual filter
}

func (s *selection) String() string {
	return fmt.Sprintf("Selection: %s", s.expr)
}

func (s *selection) Schema() *schema {
	return s.input.Schema()
}

func (s *selection) Inputs() []LogicalPlan {
	return []LogicalPlan{s.input}
}

// Selection creates a selection.
func Selection(plan LogicalPlan, expr ...LogicalExpr) LogicalPlan {
	return &selection{
		input: plan,
		expr:  expr,
	}
}

type aggregate struct {
	input         LogicalPlan
	groupBy       []LogicalExpr
	aggregateExpr []AggregateExpr
}

func (a *aggregate) String() string {
	return fmt.Sprintf("Aggregate: %s, %s", a.groupBy, a.aggregateExpr)
}

// Schema returns groupBy fields and aggregate fields
func (a *aggregate) Schema() *schema {
	groupByLen := len(a.groupBy)
	fs := make([]Field, len(a.aggregateExpr)+groupByLen)

	for i, expr := range a.groupBy {
		fs[i] = expr.Resolve(a.input)
	}

	for i, expr := range a.aggregateExpr {
		fs[i+groupByLen] = expr.Resolve(a.input)
	}

	return Schema(fs...)
}

func (a *aggregate) Inputs() []LogicalPlan {
	return []LogicalPlan{a.input}
}

// Aggregate creates an aggregation.
func Aggregate(plan LogicalPlan, groupBy []LogicalExpr, aggregateExpr []AggregateExpr) LogicalPlan {
	return &aggregate{
		input:         plan,
		groupBy:       groupBy,
		aggregateExpr: aggregateExpr,
	}
}

type limit struct {
	input  LogicalPlan
	limit  int
	offset int
}

func (l *limit) String() string {
	return fmt.Sprintf("Limit: %d, offset %d", l.limit, l.offset)
}

func (l *limit) Schema() *schema {
	return l.input.Schema()
}

func (l *limit) Inputs() []LogicalPlan {
	return []LogicalPlan{l.input}
}

// Limit creates a limit.
func Limit(plan LogicalPlan, _limit int, offset int) LogicalPlan {
	return &limit{
		input:  plan,
		limit:  _limit,
		offset: offset,
	}
}

type sort struct {
	input LogicalPlan
	by    []LogicalExpr
	asc   bool
}

func (s *sort) String() string {
	return fmt.Sprintf("Sort: %s", s.by)
}

func (s *sort) Schema() *schema {
	return s.input.Schema()
}

func (s *sort) Inputs() []LogicalPlan {
	return []LogicalPlan{s.input}
}

// Sort creates a sort.
func Sort(plan LogicalPlan, by []LogicalExpr, asc bool) LogicalPlan {
	return &sort{
		input: plan,
		by:    by,
		asc:   asc,
	}
}

type JoinKind int

const (
	InnerJoin JoinKind = iota
	LeftJoin
	RightJoin
	FullJoin
)

func (j JoinKind) String() string {
	switch j {
	case InnerJoin:
		return "InnerJoin"
	case LeftJoin:
		return "LeftJoin"
	case RightJoin:
		return "RightJoin"
	case FullJoin:
		return "FullJoin"
	default:
		return "Unknown"
	}
}

type join struct {
	left  LogicalPlan
	right LogicalPlan
	kind  JoinKind
	on    LogicalExpr
}

func (j *join) String() string {
	return fmt.Sprintf("%s: %s", j.kind, j.on)
}

func (j *join) Schema() *schema {
	leftFields := j.left.Schema().Fields
	rightFields := j.right.Schema().Fields
	fields := make([]Field, len(leftFields)+len(rightFields))
	copy(fields, leftFields)
	copy(fields[len(leftFields):], rightFields)
	return Schema(fields...)
}

func (j *join) Inputs() []LogicalPlan {
	return []LogicalPlan{j.left, j.right}
}

// Join creates a join.
func Join(left LogicalPlan, right LogicalPlan, kind JoinKind, on LogicalExpr) LogicalPlan {
	return &join{
		left:  left,
		right: right,
		kind:  kind,
		on:    on,
	}
}
