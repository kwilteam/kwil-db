package logical_plan

import (
	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

var Builder = newLogicalPlanBuilder()

func NewBuilder(plan LogicalPlan) *logicalPlanBuilder {
	return Builder.FromPlan(plan)
}

// logicalPlanBuilder is a helper to build logical plans.
// Method `FromPlan`, `Scan`, and `NoRelationOp` return a new logicalPlanBuilder, other methods
// modify the current builder.
type logicalPlanBuilder struct {
	plan LogicalPlan
}

func newLogicalPlanBuilder() *logicalPlanBuilder {
	return &logicalPlanBuilder{}
}

// NoRelationOp creates a new logicalPlanBuilder with no relation(from).
func (b *logicalPlanBuilder) NoRelation() *logicalPlanBuilder {
	return &logicalPlanBuilder{plan: NoSource()}
}

// FromPlan creates a new logicalPlanBuilder with a logical plan.
func (b *logicalPlanBuilder) FromPlan(plan LogicalPlan) *logicalPlanBuilder {
	return &logicalPlanBuilder{plan: plan}
}

// Scan is shorthand for Builder.FromPlan(ScanPlan(...))
func (b *logicalPlanBuilder) Scan(relation *datatypes.TableRef,
	source datasource.DataSource, projection ...string) *logicalPlanBuilder {

	scanPlan := ScanPlan(relation, source, []LogicalExpr{}, projection...)
	return b.FromPlan(scanPlan)
}

func (b *logicalPlanBuilder) JoinOn(jType JoinType, right LogicalPlan, on LogicalExpr) *logicalPlanBuilder {
	b.plan = JoinPlan(jType, b.plan, right, on)
	return b
}

// Project applies a projection to the logical plan.
// This applies to HAVING, which is like WHERE for a GROUP BY???
func (b *logicalPlanBuilder) Project(exprs ...LogicalExpr) *logicalPlanBuilder {
	b.plan = Projection(b.plan, exprs...)
	return b
}

// Filter applies a selection to the logical plan.
func (b *logicalPlanBuilder) Filter(expr LogicalExpr) *logicalPlanBuilder {
	b.plan = Filter(b.plan, expr)
	return b
}

// Limit applies LIMIT clause to the logical plan.
func (b *logicalPlanBuilder) Limit(skip, fetch int64) *logicalPlanBuilder {
	b.plan = Limit(b.plan, skip, fetch)
	return b
}

// Sort applies ORDER BY clause to the logical plan.
func (b *logicalPlanBuilder) Sort(exprs ...LogicalExpr) *logicalPlanBuilder {
	// TODO: remove pushed down columns
	b.plan = Sort(b.plan, exprs)
	return b
}

func (b *logicalPlanBuilder) Union(right LogicalPlan) *logicalPlanBuilder {
	b.plan = Union(b.plan, right)
	return b
}

func (b *logicalPlanBuilder) Distinct() *logicalPlanBuilder {
	b.plan = DistinctAll(b.plan)
	return b
}

func (b *logicalPlanBuilder) Intersect(right LogicalPlan) *logicalPlanBuilder {
	b.plan = Intersect(b.plan, right)
	return b
}

func (b *logicalPlanBuilder) Except(right LogicalPlan) *logicalPlanBuilder {
	b.plan = Except(b.plan, right)
	return b
}

func (b *logicalPlanBuilder) Aggregate(keys []LogicalExpr, aggregates []LogicalExpr) *logicalPlanBuilder {
	keys = NormalizeExprs(keys, b.plan)
	aggregates = NormalizeExprs(aggregates, b.plan)
	b.plan = Aggregate(b.plan, keys, aggregates)
	return b
}

func (b *logicalPlanBuilder) Build() LogicalPlan {
	return b.plan
}
