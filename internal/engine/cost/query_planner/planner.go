package query_planner

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/cost/catalog"
	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	lp "github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	pt "github.com/kwilteam/kwil-db/internal/engine/cost/plantree"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

type LogicalPlanner interface {
	ToExpr(expr tree.Expression, schema *ds.Schema) lp.LogicalExpr
	ToPlan(node tree.Statement) lp.LogicalPlan
}

type queryPlanner struct {
	catalog catalog.Catalog
}

func NewPlanner(catalog catalog.Catalog) *queryPlanner {
	return &queryPlanner{
		catalog: catalog,
	}
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
		// TODO: make this a separate LogicalExpression?
		case tree.LogicalOperatorAnd:
			return lp.And(l, r)
		case tree.LogicalOperatorOr:
			return lp.Or(l, r)
		default:
			panic("unknown comparison operator")
		}
	case *tree.ExpressionFunction:
		var inputs []lp.LogicalExpr
		for _, arg := range e.Inputs {
			inputs = append(inputs, q.ToExpr(arg, schema))
		}

		// use catalog? since there will be user-defined/kwil-defined functions

		switch t := e.Function.(type) {
		case *tree.ScalarFunction:
			return lp.ScalarFunc(t, inputs...)
		case *tree.AggregateFunc:
			return lp.AggregateFunc(t, inputs, t.Distinct, nil)
		default:
			panic(fmt.Sprintf("unknown function type %T", t))
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

func (q *queryPlanner) ToPlan(node tree.Statement) lp.LogicalPlan {
	return q.planStatement(node)
}

func (q *queryPlanner) planStatement(node tree.Statement) lp.LogicalPlan {
	return q.planStatementWithContext(node, NewPlannerContext())
}

func (q *queryPlanner) planStatementWithContext(node tree.Statement, ctx *PlannerContext) lp.LogicalPlan {
	switch n := node.(type) {
	case *tree.SelectStmt:
		return q.planSelect(n, ctx)
		//case *tree.Insert:
		//case *tree.Update:
		//case *tree.Delete:
	}
	return nil
}

// planSelect plans a select statement.
// NOTE: we don't support nested select with CTE.
func (q *queryPlanner) planSelect(node *tree.SelectStmt, ctx *PlannerContext) lp.LogicalPlan {
	if len(node.CTE) > 0 {
		q.buildCTEs(node.CTE, ctx)
	}

	return q.buildSelect(node.Stmt, ctx)
}

func (q *queryPlanner) buildSelect(node *tree.SelectStmtNoCte, ctx *PlannerContext) lp.LogicalPlan {
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

	// NOTE: we don't support use index of an output column as sort_expression
	// only support column name or alias
	// actually, it's allowed in parser
	// TODO: support this? @brennan: thought?
	plan = q.buildOrderBy(plan, node.OrderBy, ctx)

	// TODO: change/unwrap tree.OrderBy,use []*tree.OrderingTerm directly ?
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

	exprs := make([]lp.LogicalExpr, 0, len(node.OrderingTerms))

	for _, order := range node.OrderingTerms {
		asc := order.OrderType != tree.OrderTypeDesc
		var nullsFirst bool
		// From PostgreSQL documentation:
		// By default, null values sort as if larger than any non-null value;
		// that is, NULLS FIRST is the default for DESC order, and NULLS LAST otherwise.
		if order.NullOrdering == tree.NullOrderingTypeNone {
			if order.OrderType == tree.OrderTypeDesc {
				nullsFirst = true
			}
		} else {
			nullsFirst = order.NullOrdering == tree.NullOrderingTypeFirst
		}
		exprs = append(exprs, lp.SortExpr(
			q.ToExpr(order.Expression, schema),
			asc, nullsFirst))
	}

	return exprs
}

func (q *queryPlanner) buildLimit(plan lp.LogicalPlan, node *tree.Limit) lp.LogicalPlan {
	if node == nil {
		return plan
	}

	// TODO: change tree.Limit, use skip and fetch?

	var skip, fetch int

	if node.Offset != nil {
		switch t := node.Offset.(type) {
		case *tree.ExpressionLiteral:
			offsetExpr := q.ToExpr(t, plan.Schema())
			e, ok := offsetExpr.(*lp.LiteralIntExpr)
			if !ok {
				panic(fmt.Sprintf("unexpected offset value %s", t.Value))
			}

			skip = e.Value

			if skip < 0 {
				panic(fmt.Sprintf("invalid offset value %d", skip))
			}
		default:
			panic(fmt.Sprintf("unexpected skip type %T", t))
		}
	}

	if node.Expression == nil {
		panic("fetch expression is not provided")
	}

	switch t := node.Expression.(type) {
	case *tree.ExpressionLiteral:
		limitExpr := q.ToExpr(t, plan.Schema())
		e, ok := limitExpr.(*lp.LiteralIntExpr)
		if !ok {
			panic(fmt.Sprintf("unexpected limit value %s", t.Value))
		}

		fetch = e.Value
	default:
		panic(fmt.Sprintf("unexpected limit type %T", t))
	}

	return lp.Builder.From(plan).Limit(skip, fetch).Build()
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
	if _, ok := plan.(*lp.NoRelation); ok {
		noFrom = true
	}

	// where clause
	// after this step, we got a schema(maybe combined from different tables) to work with
	sourcePlan := q.buildFilter(plan, node.Where, ctx)

	// try to qualify expr, also expand `*`
	projectExprs := q.prepareProjectionExprs(sourcePlan, node.Columns, noFrom, ctx)

	// aggregation
	planAfterAggr, projectedExpsAfterAggr, err :=
		q.buildAggregation(node.GroupBy, sourcePlan, projectExprs, ctx)
	if err != nil {
		panic(err)
	}

	// projection after aggregate
	plan = q.buildProjection(planAfterAggr, projectedExpsAfterAggr)

	// distinct
	if node.SelectType == tree.SelectTypeDistinct {
		plan = lp.Builder.From(plan).Distinct().Build()
	}

	return plan
}

func (q *queryPlanner) buildFrom(node tree.Relation, ctx *PlannerContext) lp.LogicalPlan {
	if node == nil {
		return lp.Builder.NoRelation().Build()
	}

	return q.buildRelation(node, ctx)
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
	//tableRef := ds.TableRefQualified(node.Schema, node.Name)   // TODO: handle schema
	tableRef := ds.TableRefQualified("", node.Name)

	// TODO: handle cte
	//relName := tableRef.String()
	//cte := ctx.GetCTE(relName)
	// return ctePlan

	schemaProvider, err := q.catalog.GetSchemaSource(tableRef)
	if err != nil {
		panic(err)
	}

	return lp.Builder.Scan(tableRef, schemaProvider).Build()
}

func (q *queryPlanner) buildFilter(plan lp.LogicalPlan, node tree.Expression, ctx *PlannerContext) lp.LogicalPlan {
	if node == nil {
		return plan
	}

	// TODO: handle parent schema

	expr := q.ToExpr(node, plan.Schema())
	expr = qualifyExpr(expr, plan.Schema())
	return lp.Builder.From(plan).Filter(expr).Build()
}

func (q *queryPlanner) buildProjection(plan lp.LogicalPlan, exprs []lp.LogicalExpr) lp.LogicalPlan {
	return lp.Builder.From(plan).Project(exprs...).Build()
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

// buildAggregation builds a logical plan for an aggregate, returns the new plan
// and projected exprs.
func (q *queryPlanner) buildAggregation(groupBy *tree.GroupBy,
	sourcePlan lp.LogicalPlan, projectExprs []lp.LogicalExpr,
	ctx *PlannerContext) (lp.LogicalPlan, []lp.LogicalExpr, error) {
	if groupBy == nil {
		return sourcePlan, projectExprs, nil
	}

	aliasMap := extractAliases(projectExprs)
	// having/group_by may refer to the alias in select
	projectedPlan := q.buildProjection(sourcePlan, projectExprs)
	combinedSchema := sourcePlan.Schema().Clone().Merge(projectedPlan.Schema())

	havingExpr := q.buildHaving(groupBy.Having, combinedSchema, aliasMap, ctx)

	aggrExprs := slices.Clone(projectExprs) // shallow copy
	if havingExpr != nil {
		aggrExprs = append(aggrExprs, havingExpr)
	}
	aggrExprs = extractAggrExprs(aggrExprs)

	var groupByExprs []lp.LogicalExpr
	for _, gbExpr := range groupBy.Expressions {
		groupByExpr := q.ToExpr(gbExpr, combinedSchema)

		// avoid conflict
		aliasMapClone := cloneAliases(aliasMap)
		for _, f := range sourcePlan.Schema().Fields {
			delete(aliasMapClone, f.Name)
		}

		groupByExpr = resolveAlias(groupByExpr, aliasMapClone)
		if err := ensureSchemaSatifiesExprs(combinedSchema, []lp.LogicalExpr{groupByExpr}); err != nil {
			panic(err)
			return nil, nil, fmt.Errorf("build aggregation: %w", err)
		}

		groupByExprs = append(groupByExprs, groupByExpr)
	}

	if len(groupByExprs) > 0 || len(aggrExprs) > 0 {
		planAfterAggr, projectedExpsAfterAggr := q.buildAggregate(
			sourcePlan, projectExprs, havingExpr, groupByExprs, aggrExprs)
		return planAfterAggr, projectedExpsAfterAggr, nil
	} else {
		if havingExpr != nil {
			return nil, nil, fmt.Errorf("build aggregation: having expression without group by")
		}
		return sourcePlan, projectExprs, nil
	}
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
	// like replace a function call with a column
	aggrProjectionExprs := slices.Clone(groupByExprs)
	aggrProjectionExprs = append(aggrProjectionExprs, aggrExprs...)
	// resolve the columns in projection to qualified columns
	resolvedAggrProjectionExprs := make([]lp.LogicalExpr, len(aggrProjectionExprs))
	for i, expr := range aggrProjectionExprs {
		e := pt.TransformPostOrder(expr, func(n pt.TreeNode) pt.TreeNode {
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
	fmt.Println("==================columnsAfterAggr: ", columnsAfterAggr)

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

		plan = lp.Builder.From(plan).Project(havingExpr).Build()
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

func (q *queryPlanner) projectColumnToExpr(col tree.ResultColumn,
	plan lp.LogicalPlan, noFrom bool, ctx *PlannerContext) []lp.LogicalExpr {
	localSchema := plan.Schema()
	switch t := col.(type) {
	case *tree.ResultColumnExpression: // single column
		expr := q.ToExpr(t.Expression, localSchema)
		column := qualifyExpr(expr, localSchema)
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

		return expandStar(localSchema)
	case *tree.ResultColumnTable: // expand table.*
		return expandQualifiedStar(localSchema, t.TableName)
	default:
		panic(fmt.Sprintf("unknown result column type %T", t))
	}
}
