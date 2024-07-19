package query_planner

import (
	"fmt"
	"slices"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	lp "github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	pt "github.com/kwilteam/kwil-db/internal/engine/cost/plantree"

	"github.com/kwilteam/kwil-db/core/types"
	tree "github.com/kwilteam/kwil-db/parse"
)

type LogicalPlanner interface {
	ToExpr(expr tree.Expression, schema *dt.Schema) lp.LogicalExpr
	ToPlan(node tree.SQLStatement) lp.LogicalPlan
}

type Catalog interface {
	GetDataSource(tableRef *dt.TableRef) (datasource.DataSource, error)
}

// type defaultCatalogProvider struct {
// 	dbidAliases map[string]string // alias -> dbid
// 	srcs        map[string]datasource.DataSource
// }

type queryPlanner struct {
	catalog Catalog
}

func NewPlanner(catalog Catalog) *queryPlanner {
	return &queryPlanner{
		catalog: catalog,
	}
}

// ToExpr converts a tree.Expression to a logical expression.
// TODO: use iterator or stack to traverse the tree, instead of recursive, to avoid stack overflow.
func (q *queryPlanner) ToExpr(expr tree.Expression, schema *dt.Schema) lp.LogicalExpr {
	switch e := expr.(type) {
	case *tree.ExpressionLiteral:
		switch e.Type {
		case types.IntType:
			// NOTE: I think planner will need to acknowledge the types that engine supports
			v, _ := e.Value.(int64)
			return lp.LiteralNumeric(v)
		case types.TextType:
			v, _ := e.Value.(string)
			return lp.LiteralText(v)
		case types.BlobType:
			v, _ := e.Value.([]byte)
			return lp.LiteralBlob(v)
		case types.BoolType:
			v, _ := e.Value.(bool)
			return lp.LiteralBool(v)
		case types.NullType:
			return lp.LiteralNull()
		}
	case *tree.ExpressionColumn:
		// TODO: handle relation (use the Table field, and the input schema)
		// lp.Column(&datatypes.TableRef{Namespace: "", Table: e.Table}, e.Column) // split Table on "."?
		return lp.ColumnUnqualified(e.Column)
	//case *tree.ExpressionFunction:
	case *tree.ExpressionUnary:
		switch e.Operator {
		//case tree.UnaryOperatorNeg:
		//case tree.UnaryOperatorPos:
		case tree.UnaryOperatorNot:
			return lp.Not(q.ToExpr(e.Expression, schema))
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
	case *tree.ExpressionComparison:
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

		default:
			panic("unknown comparison operator")
		}
	case *tree.ExpressionLogical:
		l := q.ToExpr(e.Left, schema)
		r := q.ToExpr(e.Right, schema)
		switch e.Operator {
		case tree.LogicalOperatorAnd:
			return lp.And(l, r)
		case tree.LogicalOperatorOr:
			return lp.Or(l, r)
		default:
			panic("unknown logical operator")
		}
	case *tree.ExpressionFunctionCall:
		var inputs []lp.LogicalExpr
		for _, arg := range e.Args {
			inputs = append(inputs, q.ToExpr(arg, schema))
		}

		// use catalog? since there will be user-defined/kwil-defined functions
		fn, ok := tree.Functions[e.FunctionName()]
		if !ok {
			panic(fmt.Sprintf("function %s not found", e.FunctionName()))
		}

		if fn.IsAggregate {
			return lp.AggregateFunc(e, inputs, e.Distinct, nil)
		} else {
			return lp.ScalarFunc(e, inputs...)
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
		panic("ToExpr: unknown expression type")
	}

	panic("unreachable")
}

func (q *queryPlanner) ToPlan(node *tree.SQLStatement) lp.LogicalPlan {
	return q.planStatementWithContext(node, NewPlannerContext())
}

func (q *queryPlanner) planStatementWithContext(node *tree.SQLStatement, ctx *PlannerContext) lp.LogicalPlan {
	if len(node.CTEs) > 0 {
		q.buildCTEs(node.CTEs, ctx) // this does nothing?
		// It should remember temp tables in schema for dependent plans, and... ?
		// consider each an adjacent plan run sequentially and independently?
	}

	switch n := node.SQL.(type) {
	case *tree.SelectStatement:
		return q.buildSelectStmt(n, ctx)

	case *tree.InsertStatement:
		// INSERT INTO table VALUES ...

		// Values [][]Expression => [][]LogicalExpr

		// plan for UpsertClause too, as it might be costly

		// cost is length of values and cost of each logical expression within
		// maybe consider stats of the target table for cost of the inserts

	case *tree.UpdateStatement:

	case *tree.DeleteStatement:
		// delete can have joins
		// cost increased by conditions/where (but with a constant negative for freed space?)
	}

	return nil
}

func (q *queryPlanner) buildInsertStmt(node *tree.InsertStatement, ctx *PlannerContext) lp.LogicalPlan {
	// type InsertStatement struct {
	// 	Table   string
	// 	Alias   string   // can be empty
	// 	Columns []string // can be empty
	// 	Values  [][]Expression
	// 	Upsert  *UpsertClause // can be nil
	// }

	// What expressions are even valid inside Values other than literals?

	return nil
}

func (q *queryPlanner) buildUpdateStmt(node *tree.UpdateStatement, ctx *PlannerContext) lp.LogicalPlan {

	// type UpdateStatement struct {
	// 	Table     string
	// 	Alias     string // can be empty
	// 	SetClause []*UpdateSetClause
	// 	From      Table      // can be nil
	// 	Joins     []*Join    // can be nil
	// 	Where     Expression // can be nil
	// }

	// What expressions are even valid inside Values other than literals?

	return nil
}

// buildSelectStmt build a logical plan from a select statement.
// NOTE: we don't support nested select with CTE.
func (q *queryPlanner) buildSelectStmt(node *tree.SelectStatement, ctx *PlannerContext) lp.LogicalPlan {
	//if len(node.CTEs) > 0 {
	//	q.buildCTEs(node.CTEs, ctx)
	//}

	//return q.buildSelect(node.Stmt, ctx)

	plan := q.buildSelectCore(node.SelectCores[0], ctx)

	// merge multiple selects combined with UNION etc.
	for i, rSelect := range node.SelectCores[1:] { // set operation
		// TODO: change AST tree to represent as left and right?
		//setOp := rSelect.Compound.Operator
		setOP := node.CompoundOperators[i]
		right := q.buildSelectCore(rSelect, ctx)
		switch setOP {
		case tree.CompoundOperatorUnion:
			plan = lp.Builder.FromPlan(plan).Union(right).Distinct().Build() // lp.DistinctAll(lp.Union(plan, right))
		case tree.CompoundOperatorUnionAll:
			plan = lp.Builder.FromPlan(plan).Union(right).Build()
		case tree.CompoundOperatorIntersect:
			plan = lp.Builder.FromPlan(plan).Intersect(right).Build()
		case tree.CompoundOperatorExcept:
			plan = lp.Builder.FromPlan(plan).Except(right).Build()
		default:
			panic(fmt.Sprintf("unknown set operation %s", setOP))
		}
	}

	// NOTE: we don't support use index of an output column as sort_expression
	// only support column name or alias
	// actually, it's allowed in parser
	// TODO: support this? @brennan: thought?
	plan = q.buildOrderBy(plan, node.Ordering, nil, ctx)

	// TODO: change/unwrap tree.OrderBy,use []*tree.OrderingTerm directly ?
	plan = q.buildLimit(plan, node.Limit, node.Offset)
	return plan
}

func (q *queryPlanner) buildOrderBy(plan lp.LogicalPlan, nodes []*tree.OrderingTerm, schema *dt.Schema, ctx *PlannerContext) lp.LogicalPlan {
	if nodes == nil {
		return plan
	}

	// handle (select) distinct?

	//sortExprs := q.orderByToExprs(nodes, nil, ctx)

	sortExprs := make([]lp.LogicalExpr, 0, len(nodes))

	for _, order := range nodes {
		asc := order.Order != tree.OrderTypeDesc
		nullsFirst := order.Nulls == tree.NullOrderFirst
		sortExprs = append(sortExprs, lp.SortExpr(
			q.ToExpr(order.Expression, schema),
			asc, nullsFirst))
	}

	return lp.Builder.FromPlan(plan).Sort(sortExprs...).Build()
}

func (q *queryPlanner) buildLimit(plan lp.LogicalPlan, limit tree.Expression, offset tree.Expression) lp.LogicalPlan {
	if limit == nil {
		return plan
	}

	// TODO: change tree.Limit, use skip and fetch?

	var skip, fetch int64

	if offset != nil {
		switch t := offset.(type) {
		case *tree.ExpressionLiteral:
			offsetExpr := q.ToExpr(t, plan.Schema())
			e, ok := offsetExpr.(*lp.LiteralNumericExpr)
			if !ok {
				panic(fmt.Sprintf("unexpected offset expr %T", offsetExpr))
			}

			skip = e.Value

			if skip < 0 {
				panic(fmt.Sprintf("invalid offset value %v", skip))
			}
		default:
			panic(fmt.Sprintf("unexpected skip type %T", t))
		}
	}

	switch t := limit.(type) {
	case *tree.ExpressionLiteral:
		limitExpr := q.ToExpr(t, plan.Schema())
		e, ok := limitExpr.(*lp.LiteralNumericExpr)
		if !ok {
			panic(fmt.Sprintf("unexpected limit expr %T", limitExpr))
		}

		fetch = e.Value
	default:
		panic(fmt.Sprintf("unexpected limit type %T", t))
	}

	return lp.Builder.FromPlan(plan).Limit(skip, fetch).Build()
}

// buildSelectCore builds a logical plan for a simple select statement.
// The order of building is:
// 1. from (combining any joins into single source plan)
// 2. where
// 3. group by(can use reference from select)
// 4. having(can use reference from select)
// 5. select
// 6. distinct
// 7. order by, done in buildSelect
// 8. limit, done in buildSelect
func (q *queryPlanner) buildSelectCore(node *tree.SelectCore, ctx *PlannerContext) lp.LogicalPlan {
	var plan lp.LogicalPlan

	// from clause
	plan = q.buildFrom(node.From, node.Joins, ctx)

	_, noFrom := plan.(*lp.NoRelationOp) // i.e. node.From == nil

	// where clause
	// after this step, we got a schema(maybe combined from different tables) to work with
	sourcePlan := q.buildFilter(plan, node.Where, ctx)

	// try to qualify column exprs, also expand `*`
	// i.e. []tree.ResultColumn => lp.LogicalExpr
	projectExprs := q.prepareProjectionExprs(sourcePlan, node.Columns, noFrom, ctx)

	// aggregation
	var err error
	plan, projectExprs, err = q.buildAggregation(node.GroupBy, node.Having,
		sourcePlan, projectExprs, ctx)
	if err != nil {
		panic(err)
	}

	// projection after aggregate
	plan = lp.Projection(plan, projectExprs...) // q.buildProjection(plan, projectExprs)

	// distinct
	if node.Distinct {
		plan = lp.DistinctAll(plan) // lp.Builder.FromPlan(plan).Distinct().Build()
	}

	return plan
}

// buildFrom builds a logical plan for a from clause.
// NOTE: our AST uses `From` and `Joins` together represent a relation.
func (q *queryPlanner) buildFrom(node tree.Table, joins []*tree.Join, ctx *PlannerContext) lp.LogicalPlan {
	if node == nil {
		return lp.Builder.NoRelation().Build()
	}

	left := q.buildRelation(node, ctx)

	for _, join := range joins {
		right := q.buildRelation(join.Relation, ctx)

		joinSchema := left.Schema().Join(right.Schema())
		onExpr := q.ToExpr(join.On, joinSchema)

		joinType := lp.JoinTypeFromParseType(join.Type)
		left = lp.Builder.FromPlan(left).JoinOn(joinType, right, onExpr).Build()
	}

	return left
}

func (q *queryPlanner) buildRelation(relation tree.Table, ctx *PlannerContext) lp.LogicalPlan {
	var left lp.LogicalPlan

	switch t := relation.(type) {
	case *tree.RelationTable:
		left = q.buildTableSource(t, ctx)
	case *tree.RelationSubquery:
		left = q.buildSelectStmt(t.Subquery, ctx)
	default:
		panic(fmt.Sprintf("unknown relation type %T", t))
	}

	return left
}

func (q *queryPlanner) buildCTEs(ctes []*tree.CommonTableExpression, ctx *PlannerContext) lp.LogicalPlan {
	for _, cte := range ctes {
		q.buildCTE(cte, ctx)
	}
	return nil
}

func (q *queryPlanner) buildCTE(cte *tree.CommonTableExpression, ctx *PlannerContext) lp.LogicalPlan {
	return nil // plan the cte.SelectStatement? store in planner for GetCTE?
}

func (q *queryPlanner) buildTableSource(node *tree.RelationTable, ctx *PlannerContext) lp.LogicalPlan {
	//tableRef := dt.TableRefQualified(node.Schema, node.Name)   // TODO: handle schema
	tableRef := dt.TableRefQualified("", node.Table)

	// TODO: handle cte
	//relName := tableRef.String()
	//cte := ctx.GetCTE(relName)
	// return ctePlan

	schemaProvider, err := q.catalog.GetDataSource(tableRef)
	if err != nil {
		panic(err)
	}

	// lp.ScanPlan(tableRef, schemaProvider, nil)
	return lp.Builder.Scan(tableRef, schemaProvider).Build()
}

func (q *queryPlanner) buildFilter(plan lp.LogicalPlan, node tree.Expression, ctx *PlannerContext) lp.LogicalPlan {
	if node == nil {
		return plan
	}

	// TODO: handle parent schema <- isn't this the plan.Schema() parts below?

	expr := q.ToExpr(node, plan.Schema())
	expr = qualifyExpr(expr, plan.Schema()) // why doesn't ToExpr do this?
	return lp.Filter(plan, expr)
	// return lp.Builder.FromPlan(plan).Filter(expr).Build()
}

func (q *queryPlanner) buildProjection(plan lp.LogicalPlan, exprs []lp.LogicalExpr) lp.LogicalPlan {
	return lp.Builder.FromPlan(plan).Project(exprs...).Build()
}

func (q *queryPlanner) buildHaving(node tree.Expression, schema *dt.Schema,
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
func (q *queryPlanner) buildAggregation(groupBy []tree.Expression,
	having tree.Expression,
	sourcePlan lp.LogicalPlan, projectExprs []lp.LogicalExpr,
	ctx *PlannerContext) (lp.LogicalPlan, []lp.LogicalExpr, error) {
	if groupBy == nil {
		return sourcePlan, projectExprs, nil
	}

	aliasMap := extractAliases(projectExprs)
	// having/group_by may refer to the alias in select
	projectedPlan := q.buildProjection(sourcePlan, projectExprs)
	combinedSchema := sourcePlan.Schema().Clone().Merge(projectedPlan.Schema())

	havingExpr := q.buildHaving(having, combinedSchema, aliasMap, ctx)

	aggrExprs := slices.Clone(projectExprs) // shallow copy
	if havingExpr != nil {
		aggrExprs = append(aggrExprs, havingExpr)
	}
	aggrExprs = extractAggrExprs(aggrExprs)

	var groupByExprs []lp.LogicalExpr
	for _, gbExpr := range groupBy {
		groupByExpr := q.ToExpr(gbExpr, combinedSchema)

		// avoid conflict
		aliasMapClone := cloneAliases(aliasMap)
		for _, f := range sourcePlan.Schema().Fields {
			delete(aliasMapClone, f.Name)
		}

		groupByExpr = resolveAlias(groupByExpr, aliasMapClone)
		if err := ensureSchemaSatisfiesExprs(combinedSchema, []lp.LogicalExpr{groupByExpr}); err != nil {
			// panic(err)
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

	// rewrite projection to refer to columns that are output of aggregate plan
	groupByExprs = lp.NormalizeExprs(groupByExprs, input)
	aggrExprs = lp.NormalizeExprs(aggrExprs, input)
	aggPlan := lp.Aggregate(input, groupByExprs, aggrExprs)

	var plan lp.LogicalPlan = aggPlan

	// rewrite projection to refer to columns that are output of aggregate plan.
	// like replace a function call with a column
	aggrProjectionExprs := slices.Clone(groupByExprs)
	aggrProjectionExprs = append(aggrProjectionExprs, aggrExprs...)
	// resolve the columns in projection to qualified columns
	resolvedAggrProjectionExprs := make([]lp.LogicalExpr, len(aggrProjectionExprs))
	for i, expr := range aggrProjectionExprs {
		e := pt.TransformPostOrder(expr, func(n pt.TreeNode) pt.TreeNode {
			if c, ok := n.(*lp.ColumnExpr); ok {
				field := plan.Schema().Field(c.Relation, c.Name) // c.Resolve(plan.Schema())
				return lp.Column(field.Rel, field.Name)
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

		plan = lp.Builder.FromPlan(plan).Project(havingExpr).Build() // Project?
	}

	return plan, projectedExprsAfterAggr
}

func (q *queryPlanner) prepareProjectionExprs(plan lp.LogicalPlan, node []tree.ResultColumn,
	noFrom bool, ctx *PlannerContext) []lp.LogicalExpr {
	var exprs []lp.LogicalExpr
	for _, col := range node {
		exprs = append(exprs, q.projectColumnToExpr(col, plan, noFrom, ctx)...)
	}
	return exprs
}

// projectColumnToExpr returns a slice for the case where tree.ResultColumn is a
// wildcard (*), which is expanded. Roughly speaking, this method converts the
// tree.ResultColumn to a LogicalExpr (or a slice) while qualifying column
// expressions with the source plan's schema.
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

	case *tree.ResultColumnWildcard: // expand *
		if noFrom {
			panic("cannot use * in select list without FROM clause")
		}

		if t.Table == "" { // *
			return expandStar(localSchema)
		} else { // table.*
			return expandQualifiedStar(localSchema, t.Table)
		}

	default:
		panic(fmt.Sprintf("unknown result column type %T", t))
	}
}
