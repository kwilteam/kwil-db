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

type defaultQueryPlanner struct{}

func NewPlanner() *defaultQueryPlanner {
	return &defaultQueryPlanner{}
}

func (q *defaultQueryPlanner) ToExpr(expr tree.Expression, input logical_plan.LogicalPlan) logical_plan.LogicalExpr {
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

func (q *defaultQueryPlanner) ToPlan(node tree.Ast) logical_plan.LogicalPlan {
	return q.planStatement(node)
}

func (q *defaultQueryPlanner) planStatement(node tree.Ast) logical_plan.LogicalPlan {
	return q.planStatementWithContext(node, NewPlannerContext())
}

func (q *defaultQueryPlanner) planStatementWithContext(node tree.Ast, ctx *PlannerContext) logical_plan.LogicalPlan {
	switch n := node.(type) {
	case *tree.Select:
		return q.planSelect(n, ctx)
		//case *tree.Insert:
		//case *tree.Update:
		//case *tree.Delete:
	}
	return nil
}

func (q *defaultQueryPlanner) planSelect(node *tree.Select, ctx *PlannerContext) logical_plan.LogicalPlan {

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
