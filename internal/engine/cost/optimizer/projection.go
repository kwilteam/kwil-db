package optimizer

import (
	"slices"

	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"golang.org/x/exp/maps"
)

// ProjectionRule transforms by pushing projections down to input plans.
type ProjectionRule struct {
}

func (r *ProjectionRule) Transform(plan logical_plan.LogicalPlan) logical_plan.LogicalPlan {
	// TODO: seen map key should be a combination of dbName and tableName
	return r.pushDown(plan, make(map[string]bool))
}

// pushDown tries to push down projections to source plan (Inputs) asap after reading data
// from the disk. It works by keeping track of the columns that have been seen,
// and applying those columns to the ScanOp.
// TODO: use index instead of name for the 'seen' map.
func (r *ProjectionRule) pushDown(plan logical_plan.LogicalPlan,
	seen map[string]bool) logical_plan.LogicalPlan {
	// At each step, a new copied plan will be created, not mutating the
	// original plan.
	// Also build up the 'seen' map.
	// NOTE: this can be rewritten to use a visitor pattern, the visitor can
	// keep track of the 'seen' map.
	switch p := plan.(type) {
	case *logical_plan.ScanOp: // terminal plan type, no more pushDown recursion
		// Apply the 'seen' columns to the scan operator. e.g. from projection, filter, etc.
		// This will remove any columns that are not needed.
		columns := maps.Keys(seen)
		slices.Sort(columns)

		// NOTE: what about *?
		// NOTE: what about dbName, tableName?
		return logical_plan.ScanPlan(p.Table(), p.DataSource(), nil, columns...)

	case *logical_plan.ProjectionOp:
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[0], seen)
		newInput := r.pushDown(p.Inputs()[0], seen)
		return logical_plan.Projection(newInput, p.Exprs()...)

	case *logical_plan.FilterOp:
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[0], seen)
		newInput := r.pushDown(p.Inputs()[0], seen)
		return logical_plan.Filter(newInput, p.Exprs()[0])

	case *logical_plan.JoinOp:
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[0], seen)
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[1], seen)
		newLeft := r.pushDown(p.Inputs()[0], seen)
		newRight := r.pushDown(p.Inputs()[1], seen)
		return logical_plan.Join(newLeft, newRight, p.OpType(), p.On)

	case *logical_plan.SortOp:
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[0], seen)
		newInput := r.pushDown(p.Inputs()[0], seen)
		return logical_plan.Sort(newInput, p.Exprs())

	case *logical_plan.AggregateOp:
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[0], seen)
		newInput := r.pushDown(p.Inputs()[0], seen)
		return logical_plan.Aggregate(newInput, p.GroupBy(), p.Aggregate())

	case *logical_plan.LimitOp:
		newInput := r.pushDown(p.Inputs()[0], seen)
		return logical_plan.Limit(newInput, p.Offset(), p.Limit())

	default:
		panic("unknown logical operator")
	}
}

// ExtractColumnsFromExprs extracts the columns from the expressions.
// It calls ExtractColumns for each expression.
func extractColumnsFromExprs(exprs []logical_plan.LogicalExpr,
	input logical_plan.LogicalPlan, seen map[string]bool) {
	for _, expr := range exprs {
		logical_plan.ExtractColumns(expr, input.Schema(), seen)
	}
}
