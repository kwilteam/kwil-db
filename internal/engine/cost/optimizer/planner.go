package optimizer

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"github.com/kwilteam/kwil-db/internal/engine/cost/optimizer/virtual_plan"
)

// type plannerCtx struct {}

type VirtualPlanner interface {
	ToPlan(logicalPlan logical_plan.LogicalPlan) virtual_plan.VirtualPlan
	ToExpr(expr logical_plan.LogicalExpr, input logical_plan.LogicalPlan) virtual_plan.VirtualExpr
}

// defaultVirtualPlanner creates a virtual plan from a logical plan.
type defaultVirtualPlanner struct{}

func NewPlanner() *defaultVirtualPlanner {
	return &defaultVirtualPlanner{}
}

func (q *defaultVirtualPlanner) ToPlan(logicalPlan logical_plan.LogicalPlan) virtual_plan.VirtualPlan {
	switch p := logicalPlan.(type) {
	case *logical_plan.ScanOp:
		dataSrc := p.DataSource().(datasource.FullDataSource)
		// what about any p.Filter()?
		return virtual_plan.VSeqScan(dataSrc, p.Projection()...)

	case *logical_plan.ProjectionOp:
		input := q.ToPlan(p.Inputs()[0])
		selectExprs := make([]virtual_plan.VirtualExpr, 0, len(p.Exprs()))
		for _, expr := range p.Exprs() {
			selectExprs = append(selectExprs, q.ToExpr(expr, p.Inputs()[0]))
		}
		projectedFields := make([]datatypes.Field, 0, len(selectExprs))
		for _, expr := range p.Exprs() {
			projectedFields = append(projectedFields, expr.Resolve(p.Inputs()[0].Schema()))
		}
		projectedSchema := datatypes.NewSchema(projectedFields...)

		return virtual_plan.VProjection(input, projectedSchema, selectExprs...)

	case *logical_plan.FilterOp:
		input := q.ToPlan(p.Inputs()[0])
		// NOTE: we break the predicates into individual filters
		// TODO: p.Exprs()[0] is not correct,
		// maybe change VSelection to accept multiple filters
		filterExpr := q.ToExpr(p.Exprs()[0], p.Inputs()[0])

		return virtual_plan.VSelection(input, filterExpr)

	case *logical_plan.SortOp:
		input := q.ToPlan(p.Inputs()[0])
		sortExprs := make([]virtual_plan.VirtualExpr, 0, len(p.Exprs()))
		for _, expr := range p.Exprs() {
			sortExprs = append(sortExprs, q.ToExpr(expr, p.Inputs()[0]))
		}

		return virtual_plan.VSortSTUB(input, sortExprs...)

	default:
		panic(fmt.Sprintf("ToPlan: unknown logical plan type %T", p))
	}
}

func (q *defaultVirtualPlanner) ToExpr(expr logical_plan.LogicalExpr,
	input logical_plan.LogicalPlan) virtual_plan.VirtualExpr {

	switch e := expr.(type) {
	case *logical_plan.LiteralNumericExpr:
		return &virtual_plan.VLiteralNumericExpr{Value: e.Value}
	case *logical_plan.LiteralTextExpr:
		return &virtual_plan.VLiteralStringExpr{Value: e.Value}
	case *logical_plan.AliasExpr:
		return q.ToExpr(e.Expr, input)
	case *logical_plan.ColumnExpr:
		//fmt.Println("ColumnExpr", e.Name, input.Schema().Fields)
		for i, field := range input.Schema().Fields {
			if field.Name == e.Name {
				return virtual_plan.VColumn(i)
			}
		}
		panic(fmt.Sprintf("field %s not found", e.Name)) // need projection with sort / order by?
	case *logical_plan.ColumnIdxExpr:
		return virtual_plan.VColumn(e.Idx)
	case logical_plan.BinaryExpr:
		left := q.ToExpr(e.L(), input)
		right := q.ToExpr(e.R(), input)
		switch e.Op() {
		case "AND":
			return virtual_plan.VAnd(left, right)
		case "OR":
			return virtual_plan.VOr(left, right)
		case "=":
			return virtual_plan.VEq(left, right)
		case "!=":
			return virtual_plan.VNeq(left, right)
		case ">":
			return virtual_plan.VGt(left, right)
		case "<":
			return virtual_plan.VLt(left, right)
		case ">=":
			return virtual_plan.VGte(left, right)
		case "<=":
			return virtual_plan.VLte(left, right)
		case "+":
			return virtual_plan.VAdd(left, right)
		case "-":
			return virtual_plan.VSub(left, right)
		case "*":
			return virtual_plan.VMul(left, right)
		case "/":
			return virtual_plan.VDiv(left, right)
		default:
			panic(fmt.Sprintf("unknown logical operator %s", e.Op()))
		}
	case *logical_plan.SortExpression:
		return virtual_plan.VSortExpr(q.ToExpr(e.Expr, input))
	default:
		panic(fmt.Sprintf("unknown logical expression type %T", e))
	}
}
