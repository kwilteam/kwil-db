package optimizer

import (
	"slices"

	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
)

type PredicatePushDownRule struct{}

func (r *PredicatePushDownRule) Transform(plan logical_plan.LogicalPlan) logical_plan.LogicalPlan {
	return r.pushDown(plan)
}

// pushDown pushes down the predicate to applicable point asap, hence close to
// scan operation. It's applicable if the predicate can be evaluated at one
// LogicalPlan node.
// In row based storage, it's more important than projection push down
// optimization, because it can reduce the number of rows to be processed.
// In column based storage, it's less important than projection push down
// optimization, because it can reduce the number of columns to be processed.
func (r *PredicatePushDownRule) pushDown(plan logical_plan.LogicalPlan) logical_plan.LogicalPlan {
	// when the predicate can be evaluated?
	// - if both arms of the predicate are literals
	// - if one arm is a column and the other is a literal, and the column is
	//   from the same table as the scan

	switch p := plan.(type) {
	case *logical_plan.FilterOp:
		input := p.Inputs()[0]

		switch input := input.(type) {
		case *logical_plan.ScanOp:
			// all predicates can be pushed down if the input is a scan
			preds := logical_plan.SplitConjunction(p.Exprs()[0])
			newScan := logical_plan.ScanPlan(input.Table(), input.DataSource(), preds, input.Projection()...)
			// NOTE: should keep the original selection? e.g. return a new selection, not a scan
			// we return a selection in case some of the datasource doesn't
			// support predicate push down in scan

			return logical_plan.Filter(newScan, logical_plan.Conjunction(preds...))

		case *logical_plan.FilterOp:
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
			newSelection := logical_plan.Filter(input.Inputs()[0], newPred)

			return r.pushDown(newSelection)

		//case *logical_plan.JoinOp:
		//case *logical_plan.BagOp:
		//case *logical_plan.AggregateOp:
		//case *logical_plan.ProjectionOp:

		default:
			// NOTE: this is just placeholder
			return plan
		}

	case *logical_plan.JoinOp:
		// TODO: implement
		panic("not implemented")

	default:
		panic("logical plan type not supported")
	}
}
