package query_planner

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	lp "github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

type LogicalPlanner interface {
	ToExpr(expr tree.Expression, input lp.LogicalPlan) lp.LogicalExpr
	ToPlan(node tree.Ast) lp.LogicalPlan
}

type queryPlanner struct{}

func NewPlanner() *queryPlanner {
	return &queryPlanner{}
}

func (q *queryPlanner) ToExpr(expr tree.Expression, input lp.LogicalPlan) lp.LogicalExpr {
	switch e := expr.(type) {
	case *tree.ExpressionLiteral:
		if strings.HasPrefix(e.Value, "'") {
			return &lp.LiteralStringExpr{Value: e.Value}
		} else {
			// convert to int
			i, err := strconv.Atoi(e.Value)
			if err != nil {
				panic(fmt.Sprintf("unexpected literal value %s", e.Value))
			}
			return &lp.LiteralIntExpr{Value: i}
		}
	case *tree.ExpressionColumn:
		return &lp.ColumnExpr{
			Table: e.Table,
			Name:  e.Column,
		}
	//case *tree.ExpressionFunction:
	case *tree.ExpressionUnary:
		switch e.Operator {
		//case tree.UnaryOperatorMinus:
		//case tree.UnaryOperatorPlus:
		case tree.UnaryOperatorNot:
			return lp.Not(q.ToExpr(e.Operand, input))
		default:
			panic("unknown unary operator")
		}
	case *tree.ExpressionArithmetic:
		l := q.ToExpr(e.Left, input)
		r := q.ToExpr(e.Right, input)
		switch e.Operator {
		case tree.ArithmeticOperatorAdd:
			return lp.Add(l, r)
		case tree.ArithmeticOperatorSubtract:
			return lp.Sub(l, r)
		case tree.ArithmeticOperatorMultiply:
			return lp.Mul(l, r)
		case tree.ArithmeticOperatorDivide:
			return lp.Div(l, r)
		//case tree.ArithmeticOperatorModulus:
		default:
			panic("unknown arithmetic operator")
		}
	case *tree.ExpressionBinaryComparison:
		l := q.ToExpr(e.Left, input)
		r := q.ToExpr(e.Right, input)
		switch e.Operator {
		case tree.ComparisonOperatorEqual:
			return lp.Eq(l, r)
		case tree.ComparisonOperatorNotEqual:
			return lp.Neq(l, r)
		case tree.ComparisonOperatorGreaterThan:
			return lp.Gt(l, r)
		case tree.ComparisonOperatorLessThan:
			return lp.Lt(l, r)
		case tree.ComparisonOperatorGreaterThanOrEqual:
			return lp.Gte(l, r)
		case tree.ComparisonOperatorLessThanOrEqual:
			return lp.Lte(l, r)
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

func (q *queryPlanner) ToPlan(node tree.Ast) lp.LogicalPlan {
	return q.planStatement(node)
}

func (q *queryPlanner) planStatement(node tree.Ast) lp.LogicalPlan {
	return q.planStatementWithContext(node, NewPlannerContext())
}

func (q *queryPlanner) planStatementWithContext(node tree.Ast, ctx *PlannerContext) lp.LogicalPlan {
	switch n := node.(type) {
	case *tree.Select:
		return q.planSelect(n, ctx)
		//case *tree.Insert:
		//case *tree.Update:
		//case *tree.Delete:
	}
	return nil
}

// planSelect plans a select statement.
// NOTE: we don't support nested select with CTE.
func (q *queryPlanner) planSelect(node *tree.Select, ctx *PlannerContext) lp.LogicalPlan {
	if len(node.CTE) > 0 {
		q.buildCTEs(node.CTE, ctx)
	}

	return q.buildSelect(node.SelectStmt, ctx)
}

func (q *queryPlanner) buildSelect(node *tree.SelectStmt, ctx *PlannerContext) lp.LogicalPlan {
	var plan lp.LogicalPlan
	if len(node.SelectCores) > 1 { // set operation
		left := q.buildSelectPlan(node.SelectCores[0], ctx)
		for _, rSelect := range node.SelectCores[1:] {
			// TODO: change AST tree to represent as left and right?
			setOp := rSelect.Compound.Operator
			right := q.buildSelectPlan(rSelect, ctx)
			switch setOp {
			case tree.CompoundOperatorTypeUnion:
				plan = lp.Builder.From(left).Union(right).Distinct().Build()
			case tree.CompoundOperatorTypeUnionAll:
				plan = lp.Builder.From(left).Union(right).Build()
			case tree.CompoundOperatorTypeIntersect:
				plan = lp.Builder.From(left).Intersect(right).Build()
			case tree.CompoundOperatorTypeExcept:
				plan = lp.Builder.From(left).Except(right).Build()
			default:
				panic(fmt.Sprintf("unknown set operation %s", setOp))
			}
			left = plan
		}
	} else { // plain select
		plan = q.buildSelectPlan(node.SelectCores[0], ctx)
	}

	plan = q.buildOrderBy(plan, node.OrderBy, ctx)
	plan = q.buildLimit(plan, node.Limit)
	return plan
}

func (q *queryPlanner) buildOrderBy(plan lp.LogicalPlan, node *tree.OrderBy, ctx *PlannerContext) lp.LogicalPlan {
	if node == nil {
		return plan
	}

	// handle (select) distinct?

	sortExprs := q.orderByToExprs(node, nil, ctx)

	return lp.Builder.From(plan).Sort(sortExprs...).Build()
}

// buildProjection converts tree.OrderBy to []logical_plan.LogicalExpr
func (q *queryPlanner) orderByToExprs(node *tree.OrderBy, schema *datasource.Schema, ctx *PlannerContext) []lp.LogicalExpr {
	if node == nil {
		return []lp.LogicalExpr{}
	}

	exprs := make([]lp.LogicalExpr, len(node.OrderingTerms), 0)

	for _, order := range node.OrderingTerms {
		asc := order.OrderType.String() == "ASC"
		nullsFirst := order.NullOrdering.String() == "NULLS FIRST"
		exprs = append(exprs, lp.SortExpr(
			q.ToExpr(order.Expression, nil),
			asc, nullsFirst))
	}

	return exprs
}

func (q *queryPlanner) buildLimit(plan lp.LogicalPlan, node *tree.Limit) lp.LogicalPlan {
	if node == nil {
		return plan
	}

	var offset, limit int

	if node.Offset != nil {
		switch t := node.Offset.(type) {
		case *tree.ExpressionLiteral:
			offsetExpr := q.ToExpr(t, plan)
			e, ok := offsetExpr.(*lp.LiteralIntExpr)
			if !ok {
				panic(fmt.Sprintf("unexpected offset value %s", t.Value))
			}

			offset = e.Value

			if offset < 0 {
				panic(fmt.Sprintf("invalid offset value %d", offset))
			}
		default:
			panic(fmt.Sprintf("unexpected offset type %T", t))
		}
	}

	if node.Expression == nil {
		panic("limit expression is not provided")
	}

	switch t := node.Expression.(type) {
	case *tree.ExpressionLiteral:
		limitExpr := q.ToExpr(t, plan)
		e, ok := limitExpr.(*lp.LiteralIntExpr)
		if !ok {
			panic(fmt.Sprintf("unexpected limit value %s", t.Value))
		}

		limit = e.Value
	default:
		panic(fmt.Sprintf("unexpected limit type %T", t))
	}

	return lp.Builder.From(plan).Limit(offset, limit).Build()
}

func (q *queryPlanner) buildSelectPlan(node *tree.SelectCore, ctx *PlannerContext) lp.LogicalPlan {
	var plan lp.LogicalPlan

	plan = q.buildFrom(node.From, ctx)

	plan = q.buildFilter(plan, node.Where) // where

	// expand * in select list

	if node.GroupBy != nil {
		plan = b.buildAggregate(plan, node.GroupBy, node.Columns) // group by
		plan = b.buildFilter(plan, node.GroupBy.Having)           // having
	}

	// if orderBy , project for order

	plan = b.buildDistinct(plan, node.SelectType, node.Columns) // distinct

	plan = b.buildProjection(plan, orderBy, node.Columns) // project

	// done in VisitSelectStmt and VisitTableOrSubQuerySelect
	//plan = b.buildSort()  // order by
	//plan = b.buildLimit() // limit

	return plan
}

func (q *queryPlanner) buildFrom(node *tree.FromClause, ctx *PlannerContext) lp.LogicalPlan {
	if node == nil {
		return lp.Builder.NoRelation().Build()
	}

	return q.buildJoins(node.JoinClause, ctx)
}

func (q *queryPlanner) buildJoins(joins *tree.JoinClause, ctx *PlannerContext) lp.LogicalPlan {
	left := q.buildDataSource(joins.TableOrSubquery, ctx)

	if len(joins.Joins) > 0 {
		var tmpPlan lp.LogicalPlan
		for _, join := range joins.Joins {
			if tmpPlan == nil {
				left = tmpPlan
			}
			right := q.buildDataSource(join.Table, ctx)
			tmpPlan = lp.Builder.JoinOn(
				join.JoinOperator.ToSQL(), right, q.ToExpr(join.Constraint, nil)).Build()
		}
		return tmpPlan
	} else {
		return left
	}
}

func (q *queryPlanner) relationFromTableOrSubquery(t tree.TableOrSubquery, ctx *PlannerContext) lp.LogicalPlan {
	switch tt := t.(type) {
	case *tree.TableOrSubqueryTable:
		return q.buildDataSource(tt, ctx)
	case *tree.TableOrSubquerySelect:
		return q.buildSelect(tt.Select, ctx)
	case *tree.TableOrSubqueryJoin:
		return q.buildJoin(tt, ctx)
	case *tree.TableOrSubqueryList:
		return q.buildTableOrSubqueryList(tt, ctx)
	default:
		panic(fmt.Sprintf("unknown table or subquery type %T", tt))
	}
	//return nil
}

func (q *queryPlanner) buildCTEs(ctes []*tree.CTE, ctx *PlannerContext) lp.LogicalPlan {
	for _, cte := range ctes {
		q.buildCTE(cte, ctx)
	}
	return nil
}

func (q *queryPlanner) buildCTE(cte *tree.CTE, ctx *PlannerContext) lp.LogicalPlan {
	return nil
}

func (q *queryPlanner) buildDataSource(node tree.Ast, ctx *PlannerContext) lp.LogicalPlan {
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
func extractColumnsFromFilterExpr(expr lp.LogicalExpr, seen map[string]bool) {
	switch e := expr.(type) {
	case *lp.LiteralStringExpr:
	case *lp.LiteralIntExpr:
	case *lp.AliasExpr:
		extractColumnsFromFilterExpr(e.Expr, seen)
	case lp.UnaryExpr:
		extractColumnsFromFilterExpr(e.E(), seen)
	case lp.AggregateExpr:
		extractColumnsFromFilterExpr(e.E(), seen)
	case lp.BinaryExpr:
		extractColumnsFromFilterExpr(e.L(), seen)
		extractColumnsFromFilterExpr(e.R(), seen)
	case *lp.ColumnExpr:
		seen[e.Name] = true
	//case *.ColumnIdxExpr:
	//	seen[input.Schema().Fields[e.Idx].Name] = true
	default:
		panic(fmt.Sprintf("unknown expression type %T", e))
	}
}
