package virtual_plan

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
)

type plannerCtx struct {
}

type VirtualPlanner interface {
	ToPlan(logicalPlan logical_plan.LogicalPlan) VirtualPlan
	ToExpr(expr logical_plan.LogicalExpr, input logical_plan.LogicalPlan) VirtualExpr
}

// defaultVirtualPlanner creates a virtual plan from a logical plan.
type defaultVirtualPlanner struct{}

func NewPlanner() *defaultVirtualPlanner {
	return &defaultVirtualPlanner{}
}

func (q *defaultVirtualPlanner) ToPlan(logicalPlan logical_plan.LogicalPlan) VirtualPlan {
	switch p := logicalPlan.(type) {
	case *logical_plan.ScanOp:
		return VScan(p.DataSource(), p.Projection()...)
	case *logical_plan.ProjectionOp:
		input := q.ToPlan(p.Inputs()[0])
		selectExprs := make([]VirtualExpr, 0, len(p.Exprs()))
		for _, expr := range p.Exprs() {
			selectExprs = append(selectExprs, q.ToExpr(expr, p.Inputs()[0]))
		}
		projectedFields := make([]datasource.Field, 0, len(selectExprs))
		for _, expr := range p.Exprs() {
			projectedFields = append(projectedFields, expr.Resolve(p.Inputs()[0]))
		}
		projectedSchema := datasource.NewSchema(projectedFields...)
		return VProjection(input, projectedSchema, selectExprs...)
	case *logical_plan.SelectionOp:
		input := q.ToPlan(p.Inputs()[0])
		// NOTE: we break the predicates into individual filters
		// TODO: p.Exprs()[0] is not correct,
		// maybe change VSelection to accept multiple filters
		filterExpr := q.ToExpr(p.Exprs()[0], p.Inputs()[0])
		return VSelection(input, filterExpr)
	default:
		panic(fmt.Sprintf("unknown logical plan type %T", p))
	}
}

func (q *defaultVirtualPlanner) ToExpr(expr logical_plan.LogicalExpr,
	input logical_plan.LogicalPlan) VirtualExpr {
	switch e := expr.(type) {
	case *logical_plan.LiteralIntExpr:
		return &VLiteralIntExpr{e.Value}
	case *logical_plan.LiteralStringExpr:
		return &VLiteralStringExpr{e.Value}
	case *logical_plan.AliasExpr:
		return q.ToExpr(e.Expr, input)
	case *logical_plan.ColumnExpr:
		//fmt.Println("ColumnExpr", e.Name, input.Schema().Fields)
		for i, field := range input.Schema().Fields {
			if field.Name == e.Name {
				return VColumn(i)
			}
		}
		panic(fmt.Sprintf("field %s not found", e.Name))
	case *logical_plan.ColumnIdxExpr:
		return VColumn(e.Idx)
	case logical_plan.BinaryExpr:
		left := q.ToExpr(e.L(), input)
		right := q.ToExpr(e.R(), input)
		switch e.Op() {
		case "AND":
			return VAnd(left, right)
		case "OR":
			return VOr(left, right)
		case "=":
			return VEq(left, right)
		case "!=":
			return VNeq(left, right)
		case ">":
			return VGt(left, right)
		case "<":
			return VLt(left, right)
		case ">=":
			return VGte(left, right)
		case "<=":
			return VLte(left, right)
		case "+":
			return VAdd(left, right)
		case "-":
			return VSub(left, right)
		case "*":
			return VMul(left, right)
		case "/":
			return VDiv(left, right)
		default:
			panic(fmt.Sprintf("unknown logical operator %s", e.Op()))
		}

	default:
		panic(fmt.Sprintf("unknown logical expression type %T", e))
	}
}
