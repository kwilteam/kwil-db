package query_planner

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

type LogicalPlanner interface {
	ToExpr(expr tree.Expression, input logical_plan.LogicalPlan) logical_plan.LogicalExpr
	ToPlan(node tree.Ast) logical_plan.LogicalPlan
}

type queryPlanner struct{}

func NewPlanner() *queryPlanner {
	return &queryPlanner{}
}

func (q *queryPlanner) ToExpr(expr tree.Expression, input logical_plan.LogicalPlan) logical_plan.LogicalExpr {
	switch e := expr.(type) {
	case *tree.ExpressionLiteral:
		return &logical_plan.LiteralStringExpr{Value: e.Value}
	case *tree.ExpressionColumn:
		return &logical_plan.ColumnExpr{
			Table: e.Table,
			Name:  e.Column,
		}
	//case *tree.ExpressionFunction:
	case *tree.ExpressionUnary:
		switch e.Operator {
		//case tree.UnaryOperatorMinus:
		//case tree.UnaryOperatorPlus:
		case tree.UnaryOperatorNot:
			return logical_plan.Not(q.ToExpr(e.Operand, input))
		default:
			panic("unknown unary operator")
		}
	case *tree.ExpressionArithmetic:
		l := q.ToExpr(e.Left, input)
		r := q.ToExpr(e.Right, input)
		switch e.Operator {
		case tree.ArithmeticOperatorAdd:
			return logical_plan.Add(l, r)
		case tree.ArithmeticOperatorSubtract:
			return logical_plan.Sub(l, r)
		case tree.ArithmeticOperatorMultiply:
			return logical_plan.Mul(l, r)
		case tree.ArithmeticOperatorDivide:
			return logical_plan.Div(l, r)
		//case tree.ArithmeticOperatorModulus:
		default:
			panic("unknown arithmetic operator")
		}
	case *tree.ExpressionBinaryComparison:
		l := q.ToExpr(e.Left, input)
		r := q.ToExpr(e.Right, input)
		switch e.Operator {
		case tree.ComparisonOperatorEqual:
			return logical_plan.Eq(l, r)
		case tree.ComparisonOperatorNotEqual:
			return logical_plan.Neq(l, r)
		case tree.ComparisonOperatorGreaterThan:
			return logical_plan.Gt(l, r)
		case tree.ComparisonOperatorLessThan:
			return logical_plan.Lt(l, r)
		case tree.ComparisonOperatorGreaterThanOrEqual:
			return logical_plan.Gte(l, r)
		case tree.ComparisonOperatorLessThanOrEqual:
			return logical_plan.Lte(l, r)
		default:
			panic("unknown comparison operator")
		}
	//case *tree.ExpressionStringCompare:
	//	switch e.Operator {
	//	case tree.StringOperatorNotLike:
	//	}
	//case *tree.ExpressionBindParameter:
	//case *tree.ExpressionCollate:
	//case *tree.ExpressionIs:
	//case *tree.ExpressionList:
	//case *tree.ExpressionSelect:
	//case *tree.ExpressionBetween:
	//case *tree.ExpressionCase:
	default:
		panic("unknown expression type")
	}
}

func (q *queryPlanner) ToPlan(node tree.Ast) logical_plan.LogicalPlan {
	return q.planStatement(node)
}

func (q *queryPlanner) planStatement(node tree.Ast) logical_plan.LogicalPlan {
	return q.planStatementWithContext(node, NewPlannerContext())
}

func (q *queryPlanner) planStatementWithContext(node tree.Ast, ctx *PlannerContext) logical_plan.LogicalPlan {
	switch n := node.(type) {
	case *tree.Select:
		return q.planSelect(n, ctx)
		//case *tree.Insert:
		//case *tree.Update:
		//case *tree.Delete:
	}
	return nil
}

func (q *queryPlanner) planSelect(node *tree.Select, ctx *PlannerContext) logical_plan.LogicalPlan {
	if len(node.CTE) > 0 {
		q.buildCTEs(node.CTE, ctx)
	}

	return nil
}

func (q *queryPlanner) buildSelect(node *tree.SelectStmt, ctx *PlannerContext) logical_plan.LogicalPlan {
	var plan logical_plan.LogicalPlan
	//if len(node.SelectCores) > 1 {
	//	// set operation (it's tree.CompoundOperator)
	//	//left := b.visitSelectCore(node.SelectCores[0], node.OrderBy)
	//	//for _, core := range node.SelectCores[1:] {
	//	//	right := b.visitSelectCore(core, node.OrderBy)
	//	//	plan = NewLogicalSet(left, right, core.Compound.Operator)
	//	//	left = plan
	//	//}
	//	left := p.buildSelectPlan(node.SelectCores[0], node.OrderBy)
	//	for _, core := range node.SelectCores[1:] {
	//		right := p.buildSelectPlan(core, node.OrderBy)
	//		plan = NewLogicalSet(left, right, core.Compound.Operator)
	//		left = plan
	//	}
	//} else { // plain select
	//	plan = p.buildSelectPlan(node.SelectCores[0], node.OrderBy)
	//}

	// add order by, limit
	//plan = p.withOrderByLimit(plan, node)
	return plan
}

func (q *queryPlanner) buildSelectPlan(node *tree.SelectCore, ctx *PlannerContext) (plan logical_plan.LogicalPlan) {
	//var plan logical_plan.LogicalPlan
	//
	//plan = q.buildFrom(node.From, ctx)
	//
	//plan = q.buildFilter(plan, node.Where) // where
	//
	//// expand * in select list
	//
	//if node.GroupBy != nil {
	//	plan = b.buildAggregate(plan, node.GroupBy, node.Columns) // group by
	//	plan = b.buildFilter(plan, node.GroupBy.Having)           // having
	//}
	//
	//// if orderBy , project for order
	//
	//plan = b.buildDistinct(plan, node.SelectType, node.Columns) // distinct
	//
	//plan = b.buildProjection(plan, orderBy, node.Columns) // project
	//
	//// done in VisitSelectStmt and VisitTableOrSubQuerySelect
	////plan = b.buildSort()  // order by
	////plan = b.buildLimit() // limit

	return plan
}

func (q *queryPlanner) buildFrom(node *tree.FromClause, ctx *PlannerContext) logical_plan.LogicalPlan {
	//if node == nil {
	//	return logical_plan.NewLogicalPlanBuilder().NoRelation().Build()
	//}
	//
	//joins := node.JoinClause
	//
	//left := q.buildDataSource(joins.TableOrSubquery, ctx)
	//if len(joins.Joins) > 1 {
	//
	//}
	//
	//	rel := q.relationFromTableOrSubquery(node.JoinClause., ctx)
	//	return logical_plan.NewLogicalPlanBuilder().From(plan).Build()

	return nil
}

func (q *queryPlanner) relationFromTableOrSubquery(t tree.TableOrSubquery, ctx *PlannerContext) logical_plan.LogicalPlan {
	//switch tt := t.(type) {
	//case *tree.TableOrSubqueryTable:
	//	return q.buildDataSource(tt, ctx)
	//case *tree.TableOrSubquerySelect:
	//	return q.buildSelectPlan(tt.Select, ctx)
	//case *tree.TableOrSubqueryJoin:
	//	return q.buildJoin(tt, ctx)
	//case *tree.TableOrSubqueryList:
	//	return q.buildTableOrSubqueryList(tt, ctx)
	//default:
	//	panic(fmt.Sprintf("unknown table or subquery type %T", tt))
	//}
	return nil
}

func (q *queryPlanner) buildCTEs(ctes []*tree.CTE, ctx *PlannerContext) logical_plan.LogicalPlan {
	for _, cte := range ctes {
		q.buildCTE(cte, ctx)
	}
	return nil
}

func (q *queryPlanner) buildCTE(cte *tree.CTE, ctx *PlannerContext) logical_plan.LogicalPlan {
	return nil
}

func (q *queryPlanner) buildDataSource(node tree.Ast, ctx *PlannerContext) logical_plan.LogicalPlan {
	switch t := node.(type) {
	case tree.TableOrSubquery:
		switch tt := t.(type) {
		case *tree.TableOrSubqueryTable: // simple table
		case *tree.TableOrSubquerySelect: // subquery
		case *tree.TableOrSubqueryJoin: // join
		case *tree.TableOrSubqueryList: // values
		default:
			panic(fmt.Sprintf("unknown table or subquery type %T", tt))
		}
	// TODO: make SelectStmt a AST node
	//case *tree.SelectStmt: // select in CTE
	default:
		panic(fmt.Sprintf("unknown data source type %T", t))
	}
	return nil
}

// extractColumnsFromFilterExpr extracts the columns are references by the filter expression.
// It keeps track of the columns that have been seen in the 'seen' map.
func extractColumnsFromFilterExpr(expr logical_plan.LogicalExpr, seen map[string]bool) {
	switch e := expr.(type) {
	case *logical_plan.LiteralStringExpr:
	case *logical_plan.LiteralIntExpr:
	case *logical_plan.AliasExpr:
		extractColumnsFromFilterExpr(e.Expr, seen)
	case logical_plan.UnaryExpr:
		extractColumnsFromFilterExpr(e.E(), seen)
	case logical_plan.AggregateExpr:
		extractColumnsFromFilterExpr(e.E(), seen)
	case logical_plan.BinaryExpr:
		extractColumnsFromFilterExpr(e.L(), seen)
		extractColumnsFromFilterExpr(e.R(), seen)
	case *logical_plan.ColumnExpr:
		seen[e.Name] = true
	//case *.ColumnIdxExpr:
	//	seen[input.Schema().Fields[e.Idx].Name] = true
	default:
		panic(fmt.Sprintf("unknown expression type %T", e))
	}
}
