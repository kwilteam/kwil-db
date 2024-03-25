package logical_plan

import (
	"fmt"

	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

// NoRelation represents a no from operator.
// It corresponds to select without any from clause in SQL.
type NoRelation struct{}

func (n *NoRelation) String() string {
	return "NoRelation"
}

func (n *NoRelation) Schema() *dt.Schema {
	return dt.NewSchema()
}

func (n *NoRelation) Inputs() []LogicalPlan {
	return []LogicalPlan{}
}

func (n *NoRelation) Exprs() []LogicalExpr {
	return []LogicalExpr{}
}

func NoSource() LogicalPlan {
	return &NoRelation{}
}

// ScanOp represents a table scan operator, which produces rows from a table.
// It corresponds to `FROM` clause in SQL.
type ScanOp struct {
	table      *dt.TableRef
	dataSource ds.SchemaSource

	// used for projection push down optimization
	projection []string // TODO: use index?
	// schema after projection(i.e. only keep the projected columns in the schema)
	projectedSchema *dt.Schema
	// used for selection push down optimization
	selection LogicalExprList
}

func (s *ScanOp) Table() *dt.TableRef {
	return s.table
}

func (s *ScanOp) DataSource() ds.SchemaSource {
	return s.dataSource
}

func (s *ScanOp) Projection() []string {
	return s.projection
}

func (s *ScanOp) Selection() []LogicalExpr {
	if len(s.selection) == 0 {
		return []LogicalExpr{}
	}
	return s.selection
}

func (s *ScanOp) String() string {
	if len(s.selection) > 0 {
		return fmt.Sprintf("Scan: %s; selection=%s; projection=%s", s.table, s.selection, s.projection)
	}
	return fmt.Sprintf("Scan: %s; projection=%s", s.table, s.projection)
}

func (s *ScanOp) Schema() *dt.Schema {
	//return s.dataSource.Schema().Select(s.projection...)
	return s.projectedSchema
}

func (s *ScanOp) Inputs() []LogicalPlan {
	return []LogicalPlan{}
}

func (s *ScanOp) Exprs() []LogicalExpr {
	return []LogicalExpr{}
}

// Scan creates a table scan logical plan.
func Scan(table *dt.TableRef, ds ds.SchemaSource,
	selection []LogicalExpr, projection ...string) LogicalPlan {
	projectedSchema := ds.Schema().Select(projection...)
	qualifiedSchema := dt.NewSchemaQualified(table, projectedSchema.Fields...)
	return &ScanOp{table: table, dataSource: ds, projection: projection,
		selection: selection, projectedSchema: qualifiedSchema}
}

// ProjectionOp represents a projection operator, which produces new columns
// from the input by evaluating given expressions.
// It corresponds to `SELECT (expr...)` clause in SQL.
type ProjectionOp struct {
	input LogicalPlan
	exprs LogicalExprList
}

func (p *ProjectionOp) String() string {
	return fmt.Sprintf("Projection: %s", p.exprs)
}

func (p *ProjectionOp) Schema() *dt.Schema {
	fs := make([]dt.Field, len(p.exprs))
	schema := p.input.Schema()
	for i, expr := range p.exprs {
		fs[i] = expr.Resolve(schema)
	}
	return dt.NewSchema(fs...)
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

// FilterOp represents a filter operator, which filters out rows
// from the input that the expr evaluates to false.
// It corresponds to `WHERE expr` clause in SQL.
type FilterOp struct {
	input LogicalPlan
	expr  LogicalExpr
}

func (s *FilterOp) String() string {
	return fmt.Sprintf("Filter: %s", s.expr)
}

func (s *FilterOp) Schema() *dt.Schema {
	return s.input.Schema()
}

func (s *FilterOp) Inputs() []LogicalPlan {
	return []LogicalPlan{s.input}
}

func (s *FilterOp) Exprs() []LogicalExpr {
	return []LogicalExpr{s.expr}
}

// Filter creates a selection logical plan.
func Filter(plan LogicalPlan, expr LogicalExpr) LogicalPlan {
	return &FilterOp{
		input: plan,
		expr:  expr,
	}
}

// AggregateOp represents an aggregation operator, which groups rows by
// groupBy columns and evaluates aggregate expressions.
// It corresponds to `GROUP BY` clause in SQL.
type AggregateOp struct {
	input     LogicalPlan
	groupBy   []LogicalExpr
	aggregate []LogicalExpr
	// aggregated exprs will be added to the schema as fields
	schema *dt.Schema
}

func (a *AggregateOp) GroupBy() []LogicalExpr {
	return a.groupBy
}

func (a *AggregateOp) Aggregate() []LogicalExpr {
	return a.aggregate
}

func (a *AggregateOp) String() string {
	return fmt.Sprintf("Aggregate: %s, %s", a.groupBy, a.aggregate)
}

// Schema returns groupBy fields and aggregate fields
func (a *AggregateOp) Schema() *dt.Schema {
	groupByLen := len(a.groupBy)
	fs := make([]dt.Field, len(a.aggregate)+groupByLen)

	for i, expr := range a.groupBy {
		fs[i] = expr.Resolve(a.input.Schema())
	}

	for i, expr := range a.aggregate {
		fs[i+groupByLen] = expr.Resolve(a.input.Schema())
	}

	return dt.NewSchema(fs...)
}

func (a *AggregateOp) Inputs() []LogicalPlan {
	return []LogicalPlan{a.input}
}

func (a *AggregateOp) Exprs() []LogicalExpr {
	// NOTE: should copy
	lenGroup := len(a.groupBy)
	es := make([]LogicalExpr, lenGroup+len(a.aggregate))
	for i, e := range a.groupBy {
		es[i] = e
	}
	for i, e := range a.aggregate {
		es[i+lenGroup] = e
	}
	return es
}

// Aggregate creates an aggregation logical plan.
func Aggregate(plan LogicalPlan, groupBy []LogicalExpr,
	aggrExpr []LogicalExpr) LogicalPlan {

	// TODO: create new schema for aggregation
	//fields := exprListToFields(groupBy)
	//
	//if len(schema.Fields) != len(groupBy)+len(aggrExpr) {
	//	panic("invalid schema for aggregation")
	//}

	return &AggregateOp{
		input:     plan,
		groupBy:   groupBy,
		aggregate: aggrExpr,
		schema:    nil,
	}
}

// LimitOp represents a limit operator, which limits the number of rows
// from the input.
// It corresponds to `LIMIT` clause in SQL.
type LimitOp struct {
	input LogicalPlan
	fetch int
	skip  int
}

func (l *LimitOp) Limit() int {
	return l.fetch
}

func (l *LimitOp) Offset() int {
	return l.skip
}

func (l *LimitOp) String() string {
	return fmt.Sprintf("Limit: skip=%d, fetch=%d", l.skip, l.fetch)
}

func (l *LimitOp) Schema() *dt.Schema {
	return l.input.Schema()
}

func (l *LimitOp) Inputs() []LogicalPlan {
	return []LogicalPlan{l.input}
}

func (a *LimitOp) Exprs() []LogicalExpr {
	return []LogicalExpr{}
}

// Limit creates a limit logical plan.
func Limit(plan LogicalPlan, skip int, fetch int) LogicalPlan {
	return &LimitOp{
		input: plan,
		fetch: fetch,
		skip:  skip,
	}
}

// SortOp represents a sort operator, which sorts the rows from the input by
// the given column and order.
// It corresponds to `ORDER BY` clause in SQL.
type SortOp struct {
	input LogicalPlan
	by    []LogicalExpr
	//asc   bool
}

//func (s *SortOp) IsAsc() bool {
//	return s.asc
//}

func (s *SortOp) String() string {
	return fmt.Sprintf("Sort: %s", s.by)
}

func (s *SortOp) Schema() *dt.Schema {
	return s.input.Schema()
}

func (s *SortOp) Inputs() []LogicalPlan {
	return []LogicalPlan{s.input}
}

func (s *SortOp) Exprs() []LogicalExpr {
	return s.by
}

// Sort creates a sort logical plan.
// func Sort(plan LogicalPlan, by []LogicalExpr, asc bool) LogicalPlan {
func Sort(plan LogicalPlan, by []LogicalExpr) LogicalPlan {
	return &SortOp{
		input: plan,
		by:    by,
		//asc:   asc,
	}
}

type JoinType int

const (
	InnerJoin JoinType = iota
	LeftJoin
	RightJoin
	FullJoin
)

func (j JoinType) String() string {
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

// JoinOp represents a join operator, which joins two inputs(combine columns).
// It corresponds to `JOIN` clause in SQL.
type JoinOp struct {
	left   LogicalPlan
	right  LogicalPlan
	opType JoinType
	On     LogicalExpr
}

func (j *JoinOp) String() string {
	return fmt.Sprintf("%s: %s", j.opType, j.On)
}

func (j *JoinOp) OpType() JoinType {
	return j.opType
}

// Schema returns the combination of left and right schema
func (j *JoinOp) Schema() *dt.Schema {
	leftFields := j.left.Schema().Fields
	rightFields := j.right.Schema().Fields
	fields := make([]dt.Field, len(leftFields)+len(rightFields))
	copy(fields, leftFields)
	copy(fields[len(leftFields):], rightFields)
	return dt.NewSchema(fields...)
}

func (j *JoinOp) Inputs() []LogicalPlan {
	return []LogicalPlan{j.left, j.right}
}

func (j *JoinOp) Exprs() []LogicalExpr {
	return []LogicalExpr{j.On}
}

// Join creates a join logical plan.
func Join(left LogicalPlan, right LogicalPlan, kind JoinType,
	on LogicalExpr) LogicalPlan {
	return &JoinOp{
		left:   left,
		right:  right,
		opType: kind,
		On:     on,
	}
}

// BagOpType represents the opType of bag operator.
// deduplicate is handled by DistinctOp.
type BagOpType int

const (
	BagUnion BagOpType = iota
	BagIntersect
	BagExcept
)

func (b BagOpType) String() string {
	switch b {
	case BagUnion:
		return "UNION"
	case BagIntersect:
		return "INTERSECT"
	case BagExcept:
		return "EXCEPT"
	default:
		panic("unknown bag op type")
	}
}

// BagOp represents a union(all) or intersect operator, which combines two
// inputs(combine rows).
// It corresponds to `UNION`, `UNION ALL` and `INTERSECT` clause in SQL.
// NOTE: here we use 'bag' instead of 'set', since it allows duplicate rows for
// BagUnionAll.
type BagOp struct {
	left   LogicalPlan
	right  LogicalPlan
	opType BagOpType
}

func (u *BagOp) OpType() BagOpType {
	return u.opType
}

func (u *BagOp) String() string {
	return fmt.Sprintf("%s: %s, %s", u.opType, u.left, u.right)
}

// Schema returns the schema of the left plan, since they should be the same.
func (u *BagOp) Schema() *dt.Schema {
	return u.left.Schema()
}

func (u *BagOp) Inputs() []LogicalPlan {
	return []LogicalPlan{u.left, u.right}
}

func (u *BagOp) Exprs() []LogicalExpr {
	return []LogicalExpr{}
}

// Union creates a union or union all logical plan.
func Union(left LogicalPlan, right LogicalPlan) LogicalPlan {
	return &BagOp{
		left:   left,
		right:  right,
		opType: BagUnion,
	}
}

// Intersect creates an intersect logical plan.
func Intersect(left LogicalPlan, right LogicalPlan) LogicalPlan {
	return &BagOp{
		left:   left,
		right:  right,
		opType: BagIntersect,
	}
}

// Except creates an except logical plan.
func Except(left LogicalPlan, right LogicalPlan) LogicalPlan {
	return &BagOp{
		left:   left,
		right:  right,
		opType: BagExcept,
	}
}

// SubqueryOp represents a subquery operator.
// It corresponds to a subquery in SQL.
type SubqueryOp struct {
	input LogicalPlan
	alias string
}

func (s *SubqueryOp) Alias() string {
	return s.alias
}

func (s *SubqueryOp) String() string {
	return fmt.Sprintf("Subquery: %s", s.alias)
}

func (s *SubqueryOp) Schema() *dt.Schema {
	return s.input.Schema()
}

func (s *SubqueryOp) Inputs() []LogicalPlan {
	return []LogicalPlan{s.input}
}

func (s *SubqueryOp) Exprs() []LogicalExpr {
	return s.input.Exprs()
}

// Subquery creates a subquery logical plan.
func Subquery(plan LogicalPlan, alias string) LogicalPlan {
	return &SubqueryOp{
		input: plan,
		alias: alias,
	}
}

type DistinctOp struct {
	input LogicalPlan
}

func (d *DistinctOp) String() string {
	return "Distinct"
}

func (d *DistinctOp) Schema() *dt.Schema {
	return d.input.Schema()
}

func (d *DistinctOp) Inputs() []LogicalPlan {
	return []LogicalPlan{d.input}
}

func (d *DistinctOp) Exprs() []LogicalExpr {
	return []LogicalExpr{}
}

// DistinctAll creates a distinct logical plan, on all selections.
func DistinctAll(plan LogicalPlan) *DistinctOp {
	return &DistinctOp{input: plan}
}
