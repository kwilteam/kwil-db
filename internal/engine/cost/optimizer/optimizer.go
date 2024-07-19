package optimizer

import (
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"github.com/kwilteam/kwil-db/internal/engine/cost/optimizer/rules"
	"github.com/kwilteam/kwil-db/internal/engine/cost/optimizer/virtual_plan"
)

type LogicalOptimizeRule interface {
	Transform(plan logical_plan.LogicalPlan) logical_plan.LogicalPlan
}

type Optimizer struct {
	rules []LogicalOptimizeRule
}

// NewOptimizer creates a new optimizer with some default rules, including
// ProjectionRule and PredicatePushDownRule.
func NewOptimizer() *Optimizer {
	return &Optimizer{
		rules: []LogicalOptimizeRule{
			// default rules on logical plan
			&rules.PredicatePushDownRule{},
			&rules.ProjectionRule{},
		},
	}
}

//// Optimize transforms a logical plan into a virtual plan.
//func (o *Optimizer) Optimize(plan logical_plan.LogicalPlan) logical_plan.LogicalPlan {
//	// NOTE: or we should just use rules to select a plan?
//	// memo is definitely better IMO, and we also can maintain the determinism
//
//	for _, rule := range o.rules {
//		// TODO: check if the rule can be applied
//		plan = rule.Transform(plan)
//	}
//
//	// for demo purpose, we use logical plan cost
//
//	return plan
//}

// Optimize transforms a logical plan into a virtual plan.
func (o *Optimizer) Optimize(plan logical_plan.LogicalPlan) virtual_plan.VirtualPlan {
	// NOTE: or we should just use rules to select a plan?
	// memo is definitely better IMO, and we also can maintain the determinism

	for _, rule := range o.rules {
		// TODO: check if the rule can be applied
		plan = rule.Transform(plan)
	}

	// for demo purpose, we use logical plan cost

	// TODO
	// explorer possible virtual plans

	//// Step 1: Initialization
	//Memo memo = new Memo();
	//memo.init(logicalPlanRoot);
	//
	//// Step 2: Exploration
	//applies transformation rules to the group
	//expressions in the memo to generate alternative logical plans. These
	//alternatives are added to the memo as new group expressions.
	//
	//for (Rule rule : explorationRules) {
	//	memo.applyRule(rule);
	//}
	//
	//// Step 3: Implementation
	// applies implementation rules to the group expressions in the memo to
	//generate virtual plans. These virtual plans are also added to the memo
	//as new group expressions.
	//
	//for (Rule rule : implementationRules) {
	//	memo.applyRule(rule);
	//}
	//
	//// Step 4: Costing
	// estimates the cost of each physical plan in the memo. The cost estimation
	//considers various factors such as data distribution, data locality,
	//system resources, and query complexity.
	//
	//for (GroupExpression groupExpression : memo.getAllGroupExpressions()) {
	//	Cost cost = costEstimator.estimateCost(groupExpression);
	//	groupExpression.setCost(cost);
	//}
	//
	//// Step 5: Selection
	// selects the physical plan with the lowest cost as the final plan
	//
	//GroupExpression bestPlan = memo.getLowestCostPlan();

	o.virtualRewrite(plan)

	panic("not implemented")
}

// virtualRewrite does ???
func (o *Optimizer) virtualRewrite(plan logical_plan.LogicalPlan) logical_plan.LogicalPlan {
	panic("not implemented")
}

func (o *Optimizer) AddRule(rule LogicalOptimizeRule) {
	o.rules = append(o.rules, rule)
}

func (o *Optimizer) Cost(plan logical_plan.LogicalPlan) int64 {
	vp := o.Optimize(plan)
	return vp.Cost()
}
