package logical_plan

type LogicalPlanBuilder struct {
	plan LogicalPlan
}

func NewLogicalPlanBuilder() *LogicalPlanBuilder {
	return &LogicalPlanBuilder{}
}

// NoRelation creates a new LogicalPlanBuilder with no relation(from).
func (b *LogicalPlanBuilder) NoRelation() *LogicalPlanBuilder {
	//b.plan = ConstantRelation
	return b
}

func (b *LogicalPlanBuilder) From(plan LogicalPlan) *LogicalPlanBuilder {
	b.plan = plan
	return b
}

func (b *LogicalPlanBuilder) JoinOn(_type string, right LogicalPlan, on LogicalExpr) *LogicalPlanBuilder {
	return b
}

func (b *LogicalPlanBuilder) Select(exprs ...LogicalExpr) *LogicalPlanBuilder {
	return b
}

func (b *LogicalPlanBuilder) Build() LogicalPlan {
	return b.plan
}
