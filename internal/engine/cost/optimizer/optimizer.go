package optimizer

import "github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"

type OptimizeRule interface {
	Optimize(plan logical_plan.LogicalPlan) logical_plan.LogicalPlan
}
