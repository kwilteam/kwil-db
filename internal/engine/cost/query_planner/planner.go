package query_planner

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	lp "github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	pt "github.com/kwilteam/kwil-db/internal/engine/cost/plantree"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

type LogicalPlanner interface {
	ToExpr(expr tree.Expression, schema *ds.Schema) lp.LogicalExpr
	ToPlan(node tree.Ast) lp.LogicalPlan
}

type queryPlanner struct{}

func NewPlanner() *queryPlanner {
	return &queryPlanner{}
}

// ToExpr converts a tree.Expression to a logical expression.
// TODO: use iterator or stack to traverse the tree, instead of recursive, to avoid stack overflow.
func (q *queryPlanner) ToExpr(expr tree.Expression, schema *ds.Schema) lp.LogicalExpr {
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
		// TODO: handle relation
		return lp.ColumnUnqualified(e.Column)
	//case *tree.ExpressionFunction:
	case *tree.ExpressionUnary:
		switch e.Operator {
		//case tree.UnaryOperatorMinus:
		//case tree.UnaryOperatorPlus:
		case tree.UnaryOperatorNot:
			return lp.Not(q.ToExpr(e.Operand, schema))
		default:
			panic("unknown unary operator")
		}
	case *tree.ExpressionArithmetic:
		l := q.ToExpr(e.Left, schema)
		r := q.ToExpr(e.Right, schema)
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
		l := q.ToExpr(e.Left, schema)
		r := q.ToExpr(e.Right, schema)
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
				panic(fmt.Sprintf("unknown set operation %s", setOp.ToSQL()))
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
func (q *queryPlanner) orderByToExprs(node *tree.OrderBy, schema *ds.Schema, ctx *PlannerContext) []lp.LogicalExpr {
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
			offsetExpr := q.ToExpr(t, plan.Schema())
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
		limitExpr := q.ToExpr(t, plan.Schema())
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

// buildSelectPlan builds a logical plan for a select statement.
// The order of building is:
// 1. from
// 2. where
// 3. group by(can use reference from select)
// 4. having(can use reference from select)
// 5. select
// 6. distinct
// 7. order by
// 8. limit
func (q *queryPlanner) buildSelectPlan(node *tree.SelectCore, ctx *PlannerContext) lp.LogicalPlan {
	var plan lp.LogicalPlan

	// from clause
	plan = q.buildFrom(node.From, ctx)

	noFrom := false
	if _, ok := plan.(*lp.NoFrom); ok {
		noFrom = true
	}

	// where clause
	// after this step, we got a schema(maybe combined from different tables) to work with
	sourcePlan := q.buildFilter(plan, node.Where, ctx)

	// try qualify expr, also expand `*`
	projectExprs := q.prepareProjectionExprs(sourcePlan, node.Columns, noFrom, ctx)

	// for having/group_by exprs
	aliasMap := extractAliases(projectExprs)

	projectedPlan := q.buildProjection(sourcePlan, projectExprs)

	combinedSchema := sourcePlan.Schema().Clone().Merge(projectedPlan.Schema())

	/////////////
	// THIS IS WHERE I LEFT!!!!!!!!
	var havingExpr lp.LogicalExpr
	if node.GroupBy != nil {
		havingExpr = q.buildHaving(node.GroupBy.Having, combinedSchema, aliasMap, ctx)
	}

	aggrExprs := slices.Clone(projectExprs) // shallow copy
	if havingExpr != nil {
		aggrExprs = append(aggrExprs, havingExpr)
	}
	aggrExprs = extractAggrExprs(aggrExprs)

	var groupByExprs []lp.LogicalExpr
	if node.GroupBy != nil {
		for _, gbExpr := range node.GroupBy.Expressions {
			groupByExpr := q.ToExpr(gbExpr, combinedSchema)

			// avoid conflict
			aliasMapClone := cloneAliases(aliasMap)
			for _, f := range sourcePlan.Schema().Fields {
				delete(aliasMapClone, f.Name)
			}

			groupByExpr = resolveAlias(groupByExpr, aliasMapClone)
			if err := ensureSchemaSatifiesExprs(combinedSchema, []lp.LogicalExpr{groupByExpr}); err != nil {
				panic(err)
			}

			groupByExprs = append(groupByExprs, groupByExpr)
		}
	}

	var planAfterAggr lp.LogicalPlan
	var projectedExpsAfterAggr []lp.LogicalExpr

	if len(groupByExprs) > 0 || len(aggrExprs) > 0 {
		planAfterAggr, projectedExpsAfterAggr = q.buildAggregate(
			sourcePlan, projectExprs, havingExpr, groupByExprs, aggrExprs)
	} else {
		if havingExpr != nil {
			panic("having expression without group by")
		}
	}

	////////////

	// another projection
	plan = q.buildProjection(planAfterAggr, projectedExpsAfterAggr)

	// distinct
	if node.SelectType == tree.SelectTypeDistinct {
		plan = lp.Builder.From(plan).Distinct().Build()
	}

	//////////

	//if node.GroupBy != nil {
	//	plan = b.buildAggregate(plan, node.GroupBy, node.Columns) // group by
	//	plan = b.buildFilter(plan, node.GroupBy.Having)           // having
	//}
	//
	//// if orderBy , project for order
	//
	//plan = b.buildDistinct(plan, node.SelectType, node.Columns) // distinct

	//// TODO: handle group by,distinct, order by, limit
	//newPlan := projectedPlan
	//var projectExprAfterAggr []lp.LogicalExpr
	//plan = q.buildProjection(newPlan, projectExprAfterAggr) // final project

	// done in VisitSelectStmt and VisitTableOrSubQuerySelect
	//plan = b.buildSort()  // order by
	//plan = b.buildLimit() // limit

	return plan
}

func (q *queryPlanner) buildFrom(node *tree.FromClause, ctx *PlannerContext) lp.LogicalPlan {
	if node == nil {
		return lp.Builder.NoRelation().Build()
	}

	return q.buildRelation(node.Relation, ctx)
}

func (q *queryPlanner) buildRelation(relation tree.Relation, ctx *PlannerContext) lp.LogicalPlan {
	var left lp.LogicalPlan

	switch t := relation.(type) {
	case *tree.RelationTable:
		left = q.buildTableSource(t, ctx)
	case *tree.RelationSubquery:
		left = q.buildSelect(t.Select, ctx)
	case *tree.RelationJoin:
		left = q.buildRelation(t.Relation, ctx)
		for _, join := range t.Joins {
			right := q.buildRelation(join.Table, ctx)
			joinSchema := left.Schema().Join(right.Schema())
			expr := q.ToExpr(join.Constraint, joinSchema)
			left = lp.Builder.JoinOn(
				join.JoinOperator.ToSQL(), left, expr).Build()
		}
	default:
		panic(fmt.Sprintf("unknown relation type %T", t))
	}

	return left
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

func (q *queryPlanner) buildTableSource(node *tree.RelationTable, ctx *PlannerContext) lp.LogicalPlan {
	//switch t := node.(type) {
	//case tree.RelationTable:
	//	switch tt := t.(type) {
	//	case *tree.TableOrSubqueryTable: // simple table
	//	case *tree.TableOrSubquerySelect: // subquery
	//	case *tree.TableOrSubqueryJoin: // join
	//	case *tree.TableOrSubqueryList: // values
	//	default:
	//		panic(fmt.Sprintf("unknown table or subquery type %T", tt))
	//	}
	//// TODO: make SelectStmt a AST node
	////case *tree.SelectStmt: // select in CTE
	//default:
	//	panic(fmt.Sprintf("unknown data source type %T", t))
	//}
	//return nil

	//tableRef, err := relationNameToTableRef(node.Name)
	//if err != nil {
	//	panic(err)
	//}
	//
	//return lp.Builder.From(node.Relation).Build()
	return nil
}

func (q *queryPlanner) buildFilter(plan lp.LogicalPlan, node tree.Expression, ctx *PlannerContext) lp.LogicalPlan {
	if node == nil {
		return plan
	}

	// TODO: handle parent schema

	expr := q.ToExpr(node, plan.Schema())
	//seen := make(map[*lp.ColumnExpr]bool)
	//extractColumnsFromFilterExpr(expr, seen)
	//expr = qualifyExpr(expr, seen, plan.Schema())
	expr = qualifyExpr(expr, plan.Schema())
	return lp.Builder.From(plan).Select(expr).Build()
}

func (q *queryPlanner) buildProjection(plan lp.LogicalPlan, exprs []lp.LogicalExpr) lp.LogicalPlan {
	return lp.Builder.From(plan).Select(exprs...).Build()
}

func (q *queryPlanner) buildHaving(node tree.Expression, schema *ds.Schema,
	aliasMap map[string]lp.LogicalExpr, ctx *PlannerContext) lp.LogicalExpr {
	if node == nil {
		return nil
	}

	expr := q.ToExpr(node, schema)
	expr = resolveAlias(expr, aliasMap)
	expr = qualifyExpr(expr, schema)
	return expr
}

// buildAggregate builds a logical plan for an aggregate.
// A typical aggregate plan has group by, having, and aggregate expressions.
func (q *queryPlanner) buildAggregate(input lp.LogicalPlan,
	projectedExprs []lp.LogicalExpr, havingExpr lp.LogicalExpr,
	groupByExprs, aggrExprs []lp.LogicalExpr) (lp.LogicalPlan, []lp.LogicalExpr) {
	plan := lp.Builder.From(input).Aggregate(groupByExprs, aggrExprs).Build()
	if p, ok := plan.(*lp.AggregateOp); ok {
		// rewrite projection to refer to columns that are output of aggregate plan.
		plan = p
		groupByExprs = p.GroupBy()
	} else {
		panic(fmt.Sprintf("unexpected plan type %T", plan))
	}

	// rewrite projection to refer to columns that are output of aggregate plan.
	//
	aggrProjectionExprs := slices.Clone(groupByExprs)
	aggrProjectionExprs = append(aggrProjectionExprs, aggrExprs...)
	// resolve the columns in projection
	resolvedAggrProjectionExprs := make([]lp.LogicalExpr, len(aggrProjectionExprs))
	for i, expr := range aggrProjectionExprs {
		e := expr.TransformUp(func(n pt.TreeNode) pt.TreeNode {
			if c, ok := n.(*lp.ColumnExpr); ok {
				field := c.Resolve(plan.Schema())
				return lp.ColumnFromDefToExpr(field.QualifiedColumn())
			}
			return n
		})

		resolvedAggrProjectionExprs[i] = e.(lp.LogicalExpr)
	}
	// replace any expressions that are not a column with a column
	// like `1+2` or `group by a+b`(a,b are alias)
	var columnsAfterAggr []lp.LogicalExpr
	for _, expr := range resolvedAggrProjectionExprs {
		columnsAfterAggr = append(columnsAfterAggr, exprAsColumn(expr, plan))
	}
	//
	// rewrite projection
	var projectedExprsAfterAggr []lp.LogicalExpr
	for _, expr := range projectedExprs {
		projectedExprsAfterAggr = append(projectedExprsAfterAggr,
			rebaseExprs(expr, resolvedAggrProjectionExprs, plan))
	}
	// make sure projection exprs can be resolved from columns

	if err := checkExprsProjectFromColumns(projectedExprsAfterAggr,
		columnsAfterAggr); err != nil {
		panic(fmt.Sprintf("build aggregation: %s", err))
	}

	if havingExpr != nil {
		havingExpr = rebaseExprs(havingExpr, resolvedAggrProjectionExprs, plan)
		if err := checkExprsProjectFromColumns(
			[]lp.LogicalExpr{havingExpr}, columnsAfterAggr); err != nil {
			panic(fmt.Sprintf("build aggregation: %s", err))
		}

		plan = lp.Builder.From(plan).Select(havingExpr).Build()
	}

	return plan, projectedExprsAfterAggr
}

func (q *queryPlanner) prepareProjectionExprs(plan lp.LogicalPlan, node []tree.ResultColumn, noFrom bool, ctx *PlannerContext) []lp.LogicalExpr {
	var exprs []lp.LogicalExpr
	for _, col := range node {
		exprs = append(exprs, q.projectColumnToExpr(col, plan, noFrom, ctx)...)
	}
	return exprs
}

func (q *queryPlanner) projectColumnToExpr(col tree.ResultColumn, plan lp.LogicalPlan, noFrom bool, ctx *PlannerContext) []lp.LogicalExpr {
	switch t := col.(type) {
	case *tree.ResultColumnExpression: // single column
		expr := q.ToExpr(t.Expression, plan.Schema())
		column := qualifyExpr(expr, nil, plan.Schema())
		if t.Alias != "" { // only add alias if it's not the same as column name
			if c, ok := column.(*lp.ColumnExpr); ok {
				if c.Name != t.Alias {
					column = lp.Alias(column, t.Alias)
				}
			}
		}
		return []lp.LogicalExpr{column}
	case *tree.ResultColumnStar: // expand *
		if noFrom {
			panic("cannot use * in select list without FROM clause")
		}

		return expandStar(plan.Schema())
	case *tree.ResultColumnTable: // expand table.*
		return expandQualifiedStar(plan.Schema(), t.TableName)
	default:
		panic(fmt.Sprintf("unknown result column type %T", t))
	}
}
