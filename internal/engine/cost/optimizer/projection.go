package optimizer

import (
	"fmt"
	"slices"

	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
)

type ProjectionRule struct {
}

func (r *ProjectionRule) Optimize(plan logical_plan.LogicalPlan) logical_plan.LogicalPlan {
	// TODO: seen map key shuold be a combination of dbName and tableName
	return r.pushDown(plan, make(map[string]bool))
}

// pushDown tries to push down projections to source plan asap after reading data
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
	case *logical_plan.ProjectionOp:
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[0], seen)
		newInput := r.pushDown(p.Inputs()[0], seen)
		return logical_plan.Projection(newInput, p.Exprs()...)
	case *logical_plan.SelectionOp:
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[0], seen)
		newInput := r.pushDown(p.Inputs()[0], seen)
		return logical_plan.Selection(newInput, p.Exprs()[0])
	case *logical_plan.JoinOp:
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[0], seen)
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[1], seen)
		newLeft := r.pushDown(p.Inputs()[0], seen)
		newRight := r.pushDown(p.Inputs()[1], seen)
		return logical_plan.Join(newLeft, newRight, p.OpType(), p.On)
	case *logical_plan.SortOp:
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[0], seen)
		newInput := r.pushDown(p.Inputs()[0], seen)
		return logical_plan.Sort(newInput, p.Exprs(), p.IsAsc())
	case *logical_plan.AggregateOp:
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[0], seen)
		newInput := r.pushDown(p.Inputs()[0], seen)
		return logical_plan.Aggregate(newInput, p.GroupBy(), p.Aggregate())
	case *logical_plan.LimitOp:
		newInput := r.pushDown(p.Inputs()[0], seen)
		return logical_plan.Limit(newInput, p.Offset(), p.Limit())
	case *logical_plan.ScanOp:
		// Apply the 'seen' columns to the scan operator.
		// This will remove any columns that are not needed.
		columns := make([]string, 0, len(seen))
		for k := range seen {
			columns = append(columns, k)
		}

		slices.Sort(columns)
		// NOTE: what about *?
		// NOTE: what about dbName, tableName?
		return logical_plan.Scan(p.Table(), p.DataSource(), nil, columns...)
	default:
		panic("unknown logical operator")
	}
}

// ExtractColumnsFromExprs extracts the columns from the expressions.
// It calls extractColumns for each expression.
func extractColumnsFromExprs(exprs []logical_plan.LogicalExpr,
	input logical_plan.LogicalPlan, seen map[string]bool) {
	for _, expr := range exprs {
		extractColumns(expr, input, seen)
	}
}

// extractColumns extracts the columns from the expression.
// It keeps track of the columns that have been seen in the 'seen' map.
func extractColumns(expr logical_plan.LogicalExpr,
	input logical_plan.LogicalPlan, seen map[string]bool) {
	switch e := expr.(type) {
	case *logical_plan.LiteralStringExpr:
	case *logical_plan.LiteralIntExpr:
	case *logical_plan.AliasExpr:
		extractColumns(e.Expr, input, seen)
	case logical_plan.UnaryExpr:
		extractColumns(e.E(), input, seen)
	case logical_plan.AggregateExpr:
		extractColumns(e.E(), input, seen)
	case logical_plan.BinaryExpr:
		extractColumns(e.L(), input, seen)
		extractColumns(e.R(), input, seen)
	case *logical_plan.ColumnExpr:
		seen[e.Name] = true
	case *logical_plan.ColumnIdxExpr:
		seen[input.Schema().Fields[e.Idx].Name] = true
	default:
		panic(fmt.Sprintf("unknown expression type %T", e))
	}
}
