package logical_plan

import (
	"fmt"
	"github.com/kwilteam/kwil-db/internal/engine/cost/internal/datasource"
	"strings"
)

type ScanOp struct {
	table      string
	dataSource datasource.DataSource
	//projection []string
}

func (s *ScanOp) String() string {
	return fmt.Sprintf("Scan: %s", s.table)
}

func (s *ScanOp) Schema() *datasource.Schema {
	return s.dataSource.Schema()
}

func (s *ScanOp) Inputs() []LogicalPlan {
	return []LogicalPlan{}
}

func (s *ScanOp) Exprs() []LogicalExpr {
	return []LogicalExpr{}
}

// Scan creates a table scan logical plan.
func Scan(table string, ds datasource.DataSource) LogicalPlan {
	return &ScanOp{table: table, dataSource: ds}
}

type ProjectionOp struct {
	input LogicalPlan
	exprs []LogicalExpr
}

func (p *ProjectionOp) String() string {
	fields := make([]string, len(p.exprs))
	for i, expr := range p.exprs {
		fields[i] = expr.String()
	}
	return fmt.Sprintf("Projection: %s", strings.Join(fields, ", p"))
}

func (p *ProjectionOp) Schema() *datasource.Schema {
	fs := make([]datasource.Field, len(p.exprs))
	for i, expr := range p.exprs {
		fs[i] = expr.Resolve(p.input)
	}
	return datasource.NewSchema(fs...)
}

func (p *ProjectionOp) Inputs() []LogicalPlan {
	return []LogicalPlan{p.input}
}

func (p *ProjectionOp) Exprs() []LogicalExpr {
	return p.exprs
}

// Projection creates a projection logical plan.
func Projection(plan LogicalPlan, exprs ...LogicalExpr) LogicalPlan {
	return &ProjectionOp{
		input: plan,
		exprs: exprs,
	}
}

type SelectionOp struct {
	input LogicalPlan
	exprs []LogicalExpr // here we break to individual filter
}

func (s *SelectionOp) String() string {
	return fmt.Sprintf("Selection: %s", s.exprs)
}

func (s *SelectionOp) Schema() *datasource.Schema {
	return s.input.Schema()
}

func (s *SelectionOp) Inputs() []LogicalPlan {
	return []LogicalPlan{s.input}
}

func (s *SelectionOp) Exprs() []LogicalExpr {
	return s.exprs
}

// Selection creates a selection logical plan.
func Selection(plan LogicalPlan, exprs ...LogicalExpr) LogicalPlan {
	return &SelectionOp{
		input: plan,
		exprs: exprs,
	}
}

type AggregateOp struct {
	input         LogicalPlan
	groupBy       []LogicalExpr
	aggregateExpr []AggregateExpr
}

func (a *AggregateOp) String() string {
	return fmt.Sprintf("Aggregate: %s, %s", a.groupBy, a.aggregateExpr)
}

// Schema returns groupBy fields and aggregate fields
func (a *AggregateOp) Schema() *datasource.Schema {
	groupByLen := len(a.groupBy)
	fs := make([]datasource.Field, len(a.aggregateExpr)+groupByLen)

	for i, expr := range a.groupBy {
		fs[i] = expr.Resolve(a.input)
	}

	for i, expr := range a.aggregateExpr {
		fs[i+groupByLen] = expr.Resolve(a.input)
	}

	return datasource.NewSchema(fs...)
}

func (a *AggregateOp) Inputs() []LogicalPlan {
	return []LogicalPlan{a.input}
}

func (a *AggregateOp) Exprs() []LogicalExpr {
	// NOTE: should copy
	lenGroup := len(a.groupBy)
	es := make([]LogicalExpr, lenGroup+len(a.aggregateExpr))
	for i, e := range a.groupBy {
		es[i] = e
	}
	for i, e := range a.aggregateExpr {
		es[i+lenGroup] = e
	}
	return es
}

// Aggregate creates an aggregation logical plan.
func Aggregate(plan LogicalPlan, groupBy []LogicalExpr,
	aggregateExpr []AggregateExpr) LogicalPlan {
	return &AggregateOp{
		input:         plan,
		groupBy:       groupBy,
		aggregateExpr: aggregateExpr,
	}
}

type LimitOp struct {
	input  LogicalPlan
	limit  int
	offset int
}

func (l *LimitOp) String() string {
	return fmt.Sprintf("Limit: %d, offset %d", l.limit, l.offset)
}

func (l *LimitOp) Schema() *datasource.Schema {
	return l.input.Schema()
}

func (l *LimitOp) Inputs() []LogicalPlan {
	return []LogicalPlan{l.input}
}

func (a *LimitOp) Exprs() []LogicalExpr {
	return []LogicalExpr{}
}

// Limit creates a limit logical plan.
func Limit(plan LogicalPlan, _limit int, offset int) LogicalPlan {
	return &LimitOp{
		input:  plan,
		limit:  _limit,
		offset: offset,
	}
}

type SortOp struct {
	input LogicalPlan
	by    []LogicalExpr
	asc   bool
}

func (s *SortOp) String() string {
	return fmt.Sprintf("Sort: %s", s.by)
}

func (s *SortOp) Schema() *datasource.Schema {
	return s.input.Schema()
}

func (s *SortOp) Inputs() []LogicalPlan {
	return []LogicalPlan{s.input}
}

func (s *SortOp) Exprs() []LogicalExpr {
	return []LogicalExpr{}
}

// Sort creates a sort logical plan.
func Sort(plan LogicalPlan, by []LogicalExpr, asc bool) LogicalPlan {
	return &SortOp{
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

type JoinOp struct {
	left  LogicalPlan
	right LogicalPlan
	kind  JoinKind
	on    LogicalExpr
}

func (j *JoinOp) String() string {
	return fmt.Sprintf("%s: %s", j.kind, j.on)
}

func (j *JoinOp) Schema() *datasource.Schema {
	leftFields := j.left.Schema().Fields
	rightFields := j.right.Schema().Fields
	fields := make([]datasource.Field, len(leftFields)+len(rightFields))
	copy(fields, leftFields)
	copy(fields[len(leftFields):], rightFields)
	return datasource.NewSchema(fields...)
}

func (j *JoinOp) Inputs() []LogicalPlan {
	return []LogicalPlan{j.left, j.right}
}

func (j *JoinOp) Exprs() []LogicalExpr {
	return []LogicalExpr{j.on}
}

// Join creates a join logical plan.
func Join(left LogicalPlan, right LogicalPlan, kind JoinKind,
	on LogicalExpr) LogicalPlan {
	return &JoinOp{
		left:  left,
		right: right,
		kind:  kind,
		on:    on,
	}
}
