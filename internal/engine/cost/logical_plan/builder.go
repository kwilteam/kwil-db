package logical_plan

var Builder = newLogicalPlanBuilder()

// logicalPlanBuilder is a helper to build logical plans.
// Method From and NoRelation return a new logicalPlanBuilder, other methods
// modify the current builder.
type logicalPlanBuilder struct {
	plan LogicalPlan
}

func newLogicalPlanBuilder() *logicalPlanBuilder {
	return &logicalPlanBuilder{}
}

// NoRelation creates a new logicalPlanBuilder with no relation(from).
func (b *logicalPlanBuilder) NoRelation() *logicalPlanBuilder {
	return &logicalPlanBuilder{plan: NoSource()}
}

// From creates a new logicalPlanBuilder with a logical plan.
func (b *logicalPlanBuilder) From(plan LogicalPlan) *logicalPlanBuilder {
	return &logicalPlanBuilder{plan: plan}
}

func (b *logicalPlanBuilder) JoinOn(_type string, right LogicalPlan, on LogicalExpr) *logicalPlanBuilder {
	return b
}

// Select applies a projection to the logical plan.
func (b *logicalPlanBuilder) Select(exprs ...LogicalExpr) *logicalPlanBuilder {
	b.plan = Projection(b.plan, exprs...)
	return b
}

// Limit applies LIMIT clause to the logical plan.
func (b *logicalPlanBuilder) Limit(offset, limit int) *logicalPlanBuilder {
	b.plan = Limit(b.plan, offset, limit)
	return b
}

// Sort applies ORDER BY clause to the logical plan.
func (b *logicalPlanBuilder) Sort(exprs ...LogicalExpr) *logicalPlanBuilder {
	// TODO: remove pushed down columns
	b.plan = Sort(b.plan, exprs)
	b.plan = Projection(b.plan, exprs...)
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
