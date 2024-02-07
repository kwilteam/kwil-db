package optimizer

import "github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"

type ProjectionRule struct {
}

func (p *ProjectionRule) optimize(plan logical_plan.LogicalPlan) logical_plan.LogicalPlan {
	return plan
}

func (p *ProjectionRule) pushDown(plan logical_plan.LogicalPlan, seen map[string]bool) logical_plan.LogicalPlan {
	switch p := plan.(type) {
	case *logical_plan.ProjectionOp:
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[0], seen)
	case *logical_plan.SelectionOp:
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[0], seen)
	case *logical_plan.JoinOp:
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[0], seen)
		extractColumnsFromExprs(p.Exprs(), p.Inputs()[1], seen)
	case *logical_plan.ScanOp:
	}

	return nil
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
	case *logical_plan.ColumnExpr:
		seen[e.Name] = true
	case *logical_plan.ColumnIdxExpr:
		seen[input.Schema().Fields[e.Idx].Name] = true
	case *logical_plan.AliasExpr:
		extractColumns(e.Expr, input, seen)
	case logical_plan.UnaryExpr:
		extractColumns(e.E(), input, seen)
	case logical_plan.AggregateExpr:
		extractColumns(e.E(), input, seen)
	case logical_plan.BinaryExpr:
		extractColumns(e.L(), input, seen)
		extractColumns(e.R(), input, seen)
	default:
		panic("unknown expression type")
	}
}
