package optimizer

import "github.com/kwilteam/kwil-db/parse/planner/logical"

// pushdownPredicates pushes down filters to the lowest possible level.
// It returns a rewritten plan with the filters pushed down. The old
// plan should not be used, as its contents may have been modified.
func pushdownPredicates(n logical.LogicalNode) (logical.LogicalNode, error) {
	logical.Rewrite(n, &logical.RewriteConfig{
		PlanCallback: func(lp logical.LogicalPlan) (logical.LogicalPlan, bool, error) {
			switch lp := lp.(type) {
			case *logical.Filter:
			}
		},
	})
}
