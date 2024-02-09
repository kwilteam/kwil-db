package optimizer

import (
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"slices"
)

type PredicateRule struct{}

func (r *PredicateRule) Optimize(plan logical_plan.LogicalPlan) logical_plan.LogicalPlan {
	return r.pushDown(plan)
}

// pushDown pushes down the predicate to applicable point asap, hence close to
// scan operation. It's applicable if the predicate can be evaluated at one
// LogicalPlan node.
// In general it's more important than projection push down.
func (r *PredicateRule) pushDown(plan logical_plan.LogicalPlan) logical_plan.LogicalPlan {
	// when the predicate can be evaluated?
	// - if both arms of the predicate are literals
	// - if one arm is a column and the other is a literal, and the column is
	//   from the same table as the scan

	switch p := plan.(type) {
	case *logical_plan.SelectionOp:
		input := p.Inputs()[0]
		switch input := input.(type) {
		case *logical_plan.ScanOp:
			// all predicates can be pushed down if the input is a scan
			preds := logical_plan.SplitConjunction(p.Exprs()[0])
			newScan := logical_plan.Scan(input.Table(), input.DataSource(), preds, input.Projection()...)
			// NOTE: should keep the original selection? e.g. return a new selection, not a scan
			return logical_plan.Selection(newScan, logical_plan.Conjunction(preds...))
		//case *logical_plan.JoinOp:
		//case *logical_plan.BagOp:
		//case *logical_plan.AggregateOp:
		//case *logical_plan.ProjectionOp:
		case *logical_plan.SelectionOp:
			// nested selection, then merge the predicates
			outerPreds := logical_plan.SplitConjunction(p.Exprs()[0])
			innerPreds := logical_plan.SplitConjunction(input.Exprs()[0])
			newPreds := outerPreds
			for _, innerPred := range innerPreds {
				if !slices.Contains(outerPreds, innerPred) {
					newPreds = append(newPreds, innerPred)
				}
			}
			newPred := logical_plan.Conjunction(newPreds...)
			newSelection := logical_plan.Selection(input.Inputs()[0], newPred)
			return r.pushDown(newSelection)
		default:
			// NOTE: this is just placeholder
			return plan
		}
	case *logical_plan.JoinOp:
		// TODO: implement
		return plan
	default:
		return plan
	}
}
