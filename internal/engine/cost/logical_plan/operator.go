package logical_plan

import (
	"fmt"

	pt "github.com/kwilteam/kwil-db/internal/engine/cost/plantree"
	"github.com/kwilteam/kwil-db/parse"

	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

// NoRelationOp represents an omitted FROM operator.
// It corresponds to select without any FROM clause in SQL.
type NoRelationOp struct {
	*pt.BaseTreeNode
}

func (o *NoRelationOp) String() string {
	return "NoRelationOp"
}

func (o *NoRelationOp) Schema() *dt.Schema {
	return dt.NewSchema()
}

func (o *NoRelationOp) Inputs() []LogicalPlan {
	return []LogicalPlan{}
}

func (o *NoRelationOp) Exprs() []LogicalExpr {
	return []LogicalExpr{}
}

func NoSource() LogicalPlan {
	return &NoRelationOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
	}
}

// ScanOp represents a table scan operator, which produces rows from a table.
// It corresponds to `FROM` clause in SQL.
type ScanOp struct {
	*pt.BaseTreeNode

	table      *dt.TableRef
	dataSource ds.DataSource

	// used for projection push down optimization
	projection []string // TODO: use index?
	// schema after projection(i.e. only keep the projected columns in the schema)
	projectedSchema *dt.Schema
	// used for filter/predicate push down optimization
	filter []LogicalExpr
}

func (o *ScanOp) Table() *dt.TableRef {
	return o.table
}

func (o *ScanOp) DataSource() ds.DataSource {
	return o.dataSource
}

func (o *ScanOp) Projection() []string {
	return o.projection
}

func (o *ScanOp) Filter() []LogicalExpr {
	if len(o.filter) == 0 {
		return []LogicalExpr{}
	}
	return o.filter
}

func (o *ScanOp) String() string {
	output := fmt.Sprintf("Scan: %s", o.table)
	if len(o.filter) > 0 {
		output += fmt.Sprintf("; filter=[%s]", PpList(o.filter))
	}
	if len(o.projection) > 0 {
		output += fmt.Sprintf("; projection=[%s]", PpList(o.projection))
	}
	return output
}

func (o *ScanOp) Schema() *dt.Schema {
	//return o.dataSource.Schema().Project(o.projection...)
	return o.projectedSchema
}

func (o *ScanOp) Inputs() []LogicalPlan {
	return []LogicalPlan{}
}

func (o *ScanOp) Exprs() []LogicalExpr {
	return []LogicalExpr{}
}

// ScanPlan creates a table scan logical plan. This is the plan that will bring
// in an actual DataSource.
func ScanPlan(table *dt.TableRef, ds ds.DataSource,
	filter []LogicalExpr, projection ...string) *ScanOp {
	projectedSchema := ds.Schema().Project(projection...)
	qualifiedSchema := dt.NewSchemaQualified(table, projectedSchema.Fields...)
	return &ScanOp{
		BaseTreeNode:    pt.NewBaseTreeNode(),
		table:           table,
		dataSource:      ds,
		projection:      projection,
		filter:          filter,
		projectedSchema: qualifiedSchema,
	}
}

// ProjectionOp represents a projection operator, which produces new columns
// from the input by evaluating given expressions.
// It corresponds to `SELECT (expr...)` clause in SQL.
type ProjectionOp struct {
	*pt.BaseTreeNode

	input LogicalPlan // e.g. a ScanOp or a FilterOp
	exprs []LogicalExpr
}

func (o *ProjectionOp) String() string {
	return fmt.Sprintf("Projection: %s", PpList(o.exprs))
}

// Schema for a ProjectionOp is the expressions resolved using the input plan
// schema.
func (o *ProjectionOp) Schema() *dt.Schema {
	fs := make([]dt.Field, len(o.exprs))
	schema := o.input.Schema()
	for i, expr := range o.exprs {
		fs[i] = expr.Resolve(schema)
	}
	return dt.NewSchema(fs...)
}

func (o *ProjectionOp) Inputs() []LogicalPlan {
	return []LogicalPlan{o.input}
}

func (o *ProjectionOp) Exprs() []LogicalExpr {
	return o.exprs
}

// Projection creates a projection logical plan.
func Projection(plan LogicalPlan, exprs ...LogicalExpr) LogicalPlan {
	return &ProjectionOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		input:        plan,
		exprs:        exprs,
	}
}

// FilterOp represents a filter operator, which filters out rows
// from the input that the expr evaluates to false.
// It corresponds to `WHERE expr` clause in SQL.
type FilterOp struct {
	*pt.BaseTreeNode

	input LogicalPlan
	expr  LogicalExpr // like Lt
}

func (o *FilterOp) String() string {
	return fmt.Sprintf("Filter: %s", o.expr)
}

func (o *FilterOp) Schema() *dt.Schema {
	return o.input.Schema()
}

func (o *FilterOp) Inputs() []LogicalPlan {
	return []LogicalPlan{o.input}
}

func (o *FilterOp) Exprs() []LogicalExpr {
	return []LogicalExpr{o.expr}
}

// Filter creates a selection logical plan.
func Filter(plan LogicalPlan, expr LogicalExpr) LogicalPlan {
	return &FilterOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		input:        plan,
		expr:         expr,
	}
}

// AggregateOp represents an aggregation operator, which groups rows by
// groupBy columns and evaluates aggregate expressions.
// It corresponds to `GROUP BY` clause in SQL.
type AggregateOp struct {
	*pt.BaseTreeNode

	input     LogicalPlan
	groupBy   []LogicalExpr
	aggregate []LogicalExpr
	// aggregated exprs will be added to the schema as fields
	schema *dt.Schema
}

func (o *AggregateOp) GroupBy() []LogicalExpr {
	return o.groupBy
}

func (o *AggregateOp) Aggregate() []LogicalExpr {
	return o.aggregate
}

func (o *AggregateOp) String() string {
	output := "Aggregate: "
	if len(o.groupBy) > 0 {
		output += fmt.Sprintf("groupBy=[%s]", PpList(o.groupBy))
	}
	if len(o.aggregate) > 0 {
		output += fmt.Sprintf("; aggr=[%s]", PpList(o.aggregate))

	}

	return output
}

// Schema returns groupBy fields and aggregate fields
func (o *AggregateOp) Schema() *dt.Schema {
	groupByLen := len(o.groupBy)
	fs := make([]dt.Field, len(o.aggregate)+groupByLen)

	for i, expr := range o.groupBy {
		fs[i] = expr.Resolve(o.input.Schema())
	}

	for i, expr := range o.aggregate {
		fs[i+groupByLen] = expr.Resolve(o.input.Schema())
	}

	return dt.NewSchema(fs...)
}

func (o *AggregateOp) Inputs() []LogicalPlan {
	return []LogicalPlan{o.input}
}

func (o *AggregateOp) Exprs() []LogicalExpr {
	// NOTE: should copy
	lenGroup := len(o.groupBy)
	es := make([]LogicalExpr, lenGroup+len(o.aggregate))
	copy(es, o.groupBy)
	copy(es[lenGroup:], o.aggregate) // for i, e := range o.aggregate { es[i+lenGroup] = e }
	return es
}

// Aggregate creates an aggregation logical plan.
func Aggregate(plan LogicalPlan, groupBy, aggrExpr []LogicalExpr) *AggregateOp {

	// TODO: create new schema for aggregation
	//fields := exprListToFields(groupBy)
	//
	//if len(schema.Fields) != len(groupBy)+len(aggrExpr) {
	//	panic("invalid schema for aggregation")
	//}

	return &AggregateOp{
		BaseTreeNode: pt.NewBaseTreeNode(),

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
	*pt.BaseTreeNode

	input LogicalPlan
	fetch int64
	skip  int64
}

func (o *LimitOp) Limit() int64 {
	return o.fetch
}

func (o *LimitOp) Offset() int64 {
	return o.skip
}

func (o *LimitOp) String() string {
	return fmt.Sprintf("Limit: skip=%d, fetch=%d", o.skip, o.fetch)
}

func (o *LimitOp) Schema() *dt.Schema {
	return o.input.Schema()
}

func (o *LimitOp) Inputs() []LogicalPlan {
	return []LogicalPlan{o.input}
}

func (o *LimitOp) Exprs() []LogicalExpr {
	return []LogicalExpr{}
}

// Limit creates a limit logical plan.
func Limit(plan LogicalPlan, skip int64, fetch int64) LogicalPlan {
	return &LimitOp{
		BaseTreeNode: pt.NewBaseTreeNode(),

		input: plan,
		fetch: fetch,
		skip:  skip,
	}
}

// SortOp represents a sort operator, which sorts the rows from the input by
// the given column and order.
// It corresponds to `ORDER BY` clause in SQL.
type SortOp struct {
	*pt.BaseTreeNode

	input LogicalPlan
	by    []LogicalExpr
}

func (o *SortOp) String() string {
	return fmt.Sprintf("Sort: %s", PpList(o.by))
}

func (o *SortOp) Schema() *dt.Schema {
	return o.input.Schema()
}

func (o *SortOp) Inputs() []LogicalPlan {
	return []LogicalPlan{o.input}
}

func (o *SortOp) Exprs() []LogicalExpr {
	return o.by
}

// Sort creates a sort logical plan.
func Sort(plan LogicalPlan, by []LogicalExpr) LogicalPlan {
	return &SortOp{
		BaseTreeNode: pt.NewBaseTreeNode(),

		input: plan,
		by:    by,
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
		return "UnknownJoin"
	}
}

func JoinTypeFromParseType(pType parse.JoinType) JoinType {
	switch pType {
	case parse.JoinTypeFull:
		return FullJoin
	case parse.JoinTypeInner:
		return InnerJoin
	case parse.JoinTypeLeft:
		return LeftJoin
	case parse.JoinTypeRight:
		return RightJoin
	default:
		panic(fmt.Sprintf("unknown join type %s", string(pType)))
	}
}

func JoinPlan(jType JoinType, left, right LogicalPlan, on LogicalExpr) *JoinOp {
	return &JoinOp{
		left:   left,
		right:  right,
		opType: jType,
		On:     on,
	}
}

// JoinOp represents a join operator, which joins two inputs(combine columns).
// It corresponds to `JOIN` clause in SQL.
type JoinOp struct {
	*pt.BaseTreeNode

	left   LogicalPlan
	right  LogicalPlan
	opType JoinType
	On     LogicalExpr
}

var _ LogicalPlan = (*JoinOp)(nil)

func (o *JoinOp) String() string {
	return fmt.Sprintf("%s: %s", o.opType, o.On)
}

func (o *JoinOp) OpType() JoinType {
	return o.opType
}

// Schema returns the combination of left and right schema
func (o *JoinOp) Schema() *dt.Schema {
	leftFields := o.left.Schema().Fields
	rightFields := o.right.Schema().Fields
	fields := make([]dt.Field, len(leftFields)+len(rightFields))
	copy(fields, leftFields)
	copy(fields[len(leftFields):], rightFields)
	return dt.NewSchema(fields...)
}

func (o *JoinOp) Inputs() []LogicalPlan {
	return []LogicalPlan{o.left, o.right}
}

func (o *JoinOp) Exprs() []LogicalExpr {
	return []LogicalExpr{o.On}
}

// Join creates a join logical plan.
func Join(left LogicalPlan, right LogicalPlan, kind JoinType,
	on LogicalExpr) LogicalPlan {
	return &JoinOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		left:         left,
		right:        right,
		opType:       kind,
		On:           on,
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
	*pt.BaseTreeNode

	left   LogicalPlan
	right  LogicalPlan
	opType BagOpType
}

func (o *BagOp) OpType() BagOpType {
	return o.opType
}

func (o *BagOp) String() string {
	return fmt.Sprintf("%s: %s, %s", o.opType, o.left, o.right)
}

// Schema returns the schema of the left plan, since they should be the same.
func (o *BagOp) Schema() *dt.Schema {
	return o.left.Schema()
}

func (o *BagOp) Inputs() []LogicalPlan {
	return []LogicalPlan{o.left, o.right}
}

func (o *BagOp) Exprs() []LogicalExpr {
	return []LogicalExpr{}
}

// Union creates a union or union all logical plan.
func Union(left LogicalPlan, right LogicalPlan) LogicalPlan {
	return &BagOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		left:         left,
		right:        right,
		opType:       BagUnion,
	}
}

// Intersect creates an intersect logical plan.
func Intersect(left LogicalPlan, right LogicalPlan) LogicalPlan {
	return &BagOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		left:         left,
		right:        right,
		opType:       BagIntersect,
	}
}

// Except creates an except logical plan.
func Except(left LogicalPlan, right LogicalPlan) LogicalPlan {
	return &BagOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		left:         left,
		right:        right,
		opType:       BagExcept,
	}
}

// SubqueryOp represents a subquery operator.
// It corresponds to a subquery in SQL.
type SubqueryOp struct {
	*pt.BaseTreeNode

	input LogicalPlan
	alias string
}

func (o *SubqueryOp) Alias() string {
	return o.alias
}

func (o *SubqueryOp) String() string {
	return fmt.Sprintf("Subquery: %s", o.alias)
}

func (o *SubqueryOp) Schema() *dt.Schema {
	return o.input.Schema()
}

func (o *SubqueryOp) Inputs() []LogicalPlan {
	return []LogicalPlan{o.input}
}

func (o *SubqueryOp) Exprs() []LogicalExpr {
	return o.input.Exprs()
}

// Subquery creates a subquery logical plan.
func Subquery(plan LogicalPlan, alias string) LogicalPlan {
	return &SubqueryOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		input:        plan,
		alias:        alias,
	}
}

type DistinctOp struct {
	*pt.BaseTreeNode

	input LogicalPlan
}

var _ LogicalPlan = (*DistinctOp)(nil)

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
	return &DistinctOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		input:        plan,
	}
}

// // pt.TreeNode implementations

// Children() implementations, it's basically the same as Inputs(), only
// that it returns a slice of pt.TreeNode.

func (o *NoRelationOp) Children() []pt.TreeNode {
	return []pt.TreeNode{}
}

func (o *ScanOp) Children() []pt.TreeNode {
	return []pt.TreeNode{}
}

func (o *ProjectionOp) Children() []pt.TreeNode {
	return []pt.TreeNode{o.input}
}

func (o *FilterOp) Children() []pt.TreeNode {
	return []pt.TreeNode{o.input}
}

func (o *AggregateOp) Children() []pt.TreeNode {
	return []pt.TreeNode{o.input}
}

func (o *LimitOp) Children() []pt.TreeNode {
	return []pt.TreeNode{o.input}
}

func (o *SortOp) Children() []pt.TreeNode {
	return []pt.TreeNode{o.input}
}

func (o *JoinOp) Children() []pt.TreeNode {
	return []pt.TreeNode{o.left, o.right}
}

func (o *BagOp) Children() []pt.TreeNode {
	return []pt.TreeNode{o.left, o.right}
}

func (o *SubqueryOp) Children() []pt.TreeNode {
	return []pt.TreeNode{o.input}
}

func (o *DistinctOp) Children() []pt.TreeNode {
	return []pt.TreeNode{o.input}
}

// TransformChildren() implementations

func (o *NoRelationOp) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return o
}

func (o *ScanOp) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return o
}

func (o *ProjectionOp) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &ProjectionOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		input:        fn(o.input).(LogicalPlan),
		exprs:        o.exprs,
	}
}

func (o *FilterOp) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &FilterOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		input:        fn(o.input).(LogicalPlan),
		expr:         o.expr,
	}
}

func (o *AggregateOp) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &AggregateOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		input:        fn(o.input).(LogicalPlan),
		groupBy:      o.groupBy,
		aggregate:    o.aggregate,
		schema:       o.schema,
	}
}

func (o *LimitOp) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &LimitOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		input:        fn(o.input).(LogicalPlan),
		fetch:        o.fetch,
		skip:         o.skip,
	}
}

func (o *SortOp) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &SortOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		input:        fn(o.input).(LogicalPlan),
		by:           o.by,
	}
}

func (o *JoinOp) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &JoinOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		left:         fn(o.left).(LogicalPlan),
		right:        fn(o.right).(LogicalPlan),
		opType:       o.opType,
		On:           o.On,
	}
}

func (o *BagOp) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &BagOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		left:         fn(o.left).(LogicalPlan),
		right:        fn(o.right).(LogicalPlan),
		opType:       o.opType,
	}
}

func (o *SubqueryOp) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &SubqueryOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		input:        fn(o.input).(LogicalPlan),
		alias:        o.alias,
	}
}

func (o *DistinctOp) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &DistinctOp{
		BaseTreeNode: pt.NewBaseTreeNode(),
		input:        fn(o.input).(LogicalPlan),
	}
}

// PlanNode() implementations

func (o *NoRelationOp) PlanNode() {}
func (o *ScanOp) PlanNode()       {}
func (o *ProjectionOp) PlanNode() {}
func (o *FilterOp) PlanNode()     {}
func (o *AggregateOp) PlanNode()  {}
func (o *LimitOp) PlanNode()      {}
func (o *SortOp) PlanNode()       {}
func (o *JoinOp) PlanNode()       {}
func (o *BagOp) PlanNode()        {}
func (o *SubqueryOp) PlanNode()   {}
func (o *DistinctOp) PlanNode()   {}
