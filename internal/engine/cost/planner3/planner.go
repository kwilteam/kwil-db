package planner3

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
)

// the planner file converts the parse AST into a logical query plan.

type plannerVisitor struct {
	parse.UnimplementedSqlVisitor

	// schema is the underlying schema that the AST was parsed against.
	schema *types.Schema

	// ctes is a slice of CTEs and their logical plans, in the order they were defined.
	ctes []struct {
		alias       string
		columnNames []string
		plan        LogicalPlan
	}
}

// the following maps map constants from parse to their logical
// equivalents. I thought about just using the parse constants,
// but decided to have clean separation so that other parts of the
// planner don't have to rely on the parse package.

var comparisonOps = map[parse.ComparisonOperator]ComparisonOperator{
	parse.ComparisonOperatorEqual:              Equal,
	parse.ComparisonOperatorNotEqual:           NotEqual,
	parse.ComparisonOperatorLessThan:           LessThan,
	parse.ComparisonOperatorLessThanOrEqual:    LessThanOrEqual,
	parse.ComparisonOperatorGreaterThan:        GreaterThan,
	parse.ComparisonOperatorGreaterThanOrEqual: GreaterThanOrEqual,
}

var logicalOps = map[parse.LogicalOperator]LogicalOperator{
	parse.LogicalOperatorAnd: And,
	parse.LogicalOperatorOr:  Or,
}

var arithmeticOps = map[parse.ArithmeticOperator]ArithmeticOperator{
	parse.ArithmeticOperatorAdd:      Add,
	parse.ArithmeticOperatorSubtract: Subtract,
	parse.ArithmeticOperatorMultiply: Multiply,
	parse.ArithmeticOperatorDivide:   Divide,
	parse.ArithmeticOperatorModulo:   Modulo,
}

var unaryOps = map[parse.UnaryOperator]UnaryOperator{
	parse.UnaryOperatorNeg: Negate,
	parse.UnaryOperatorNot: Not,
	parse.UnaryOperatorPos: Positive,
}

var joinTypes = map[parse.JoinType]JoinType{
	parse.JoinTypeInner: InnerJoin,
	parse.JoinTypeLeft:  LeftOuterJoin,
	parse.JoinTypeRight: RightOuterJoin,
	parse.JoinTypeFull:  FullOuterJoin,
}

var compoundTypes = map[parse.CompoundOperator]SetOperationType{
	parse.CompoundOperatorUnion:     Union,
	parse.CompoundOperatorUnionAll:  UnionAll,
	parse.CompoundOperatorIntersect: Intersect,
	parse.CompoundOperatorExcept:    Except,
}

var orderAsc = map[parse.OrderType]bool{
	parse.OrderTypeAsc:  true,
	parse.OrderTypeDesc: false,
	"":                  true, // default to ascending
}

var orderNullsLast = map[parse.NullOrder]bool{
	parse.NullOrderFirst: false,
	parse.NullOrderLast:  true,
	"":                   true, // default to nulls last
}

// get retrieves a value from a map, and panics if the key is not found.
// it is used to catch internal errors if we add new nodes to the AST
// without updating the planner.
func get[A comparable, B any](m map[A]B, a A) B {
	// overall, not worried about the marginal overhead of this, since plans will
	// probably be cached.
	if v, ok := m[a]; ok {
		return v
	}
	panic(fmt.Sprintf("key %v not found in map %v", a, m))
}

/*
	all of the following methods should return a LogicalExpr
*/

// cast wraps the expression with a typecast if one is used in the node.
func cast(expr LogicalExpr, node interface{ GetTypeCast() *types.DataType }) LogicalExpr {
	if node.GetTypeCast() != nil {
		return &TypeCast{
			Expr: expr,
			Type: node.GetTypeCast(),
		}
	}

	return expr
}

// exprs converts a slice of parse expressions to a slice of logical expressions.
func (p *plannerVisitor) exprs(nodes []parse.Expression) []LogicalExpr {
	exprs := make([]LogicalExpr, len(nodes))
	for i, node := range nodes {
		exprs[i] = node.Accept(p).(LogicalExpr)
	}
	return exprs
}

func (p *plannerVisitor) VisitExpressionLiteral(node *parse.ExpressionLiteral) any {
	return cast(&Literal{
		Value: node.Value,
		Type:  node.Type,
	}, node)
}

func (p *plannerVisitor) VisitExpressionFunctionCall(node *parse.ExpressionFunctionCall) any {
	return cast(&FunctionCall{
		FunctionName: node.Name,
		Args:         p.exprs(node.Args),
		Star:         node.Star, // TODO: do we need star here? Would we rather convert it for the sake of the planner?
		Distinct:     node.Distinct,
	}, node)
}

func (p *plannerVisitor) VisitExpressionForeignCall(node *parse.ExpressionForeignCall) any {
	panic("TODO: Implement")
}

func (p *plannerVisitor) VisitExpressionVariable(node *parse.ExpressionVariable) any {
	return cast(&Variable{
		VarName: node.Name,
	}, node)
}

func (p *plannerVisitor) VisitExpressionArrayAccess(node *parse.ExpressionArrayAccess) any {
	return cast(&ArrayAccess{
		Array: node.Array.Accept(p).(LogicalExpr),
		Index: node.Index.Accept(p).(LogicalExpr),
	}, node)
}

func (p *plannerVisitor) VisitExpressionMakeArray(node *parse.ExpressionMakeArray) any {
	return cast(&ArrayConstructor{
		Elements: p.exprs(node.Values),
	}, node)
}

func (p *plannerVisitor) VisitExpressionFieldAccess(node *parse.ExpressionFieldAccess) any {
	return cast(&FieldAccess{
		Object: node.Record.Accept(p).(LogicalExpr),
		Field:  node.Field,
	}, node)
}

func (p *plannerVisitor) VisitExpressionParenthesized(node *parse.ExpressionParenthesized) any {
	return cast(node.Inner.Accept(p).(LogicalExpr), node)
}

func (p *plannerVisitor) VisitExpressionComparison(node *parse.ExpressionComparison) any {
	return &ComparisonOp{
		Left:  node.Left.Accept(p).(LogicalExpr),
		Right: node.Right.Accept(p).(LogicalExpr),
		Op:    get(comparisonOps, node.Operator),
	}
}

func (p *plannerVisitor) VisitExpressionLogical(node *parse.ExpressionLogical) any {
	return &LogicalOp{
		Left:  node.Left.Accept(p).(LogicalExpr),
		Right: node.Right.Accept(p).(LogicalExpr),
		Op:    get(logicalOps, node.Operator),
	}
}

func (p *plannerVisitor) VisitExpressionArithmetic(node *parse.ExpressionArithmetic) any {
	return &ArithmeticOp{
		Left:  node.Left.Accept(p).(LogicalExpr),
		Right: node.Right.Accept(p).(LogicalExpr),
		Op:    get(arithmeticOps, node.Operator),
	}
}

func (p *plannerVisitor) VisitExpressionUnary(node *parse.ExpressionUnary) any {
	return &UnaryOp{
		Expr: node.Expression.Accept(p).(LogicalExpr),
		Op:   get(unaryOps, node.Operator),
	}
}

func (p *plannerVisitor) VisitExpressionColumn(node *parse.ExpressionColumn) any {
	return cast(&ColumnRef{
		Parent:     node.Table,
		ColumnName: node.Column,
	}, node)
}

func (p *plannerVisitor) VisitExpressionCollate(node *parse.ExpressionCollate) any {
	panic("TODO: Implement")
}

func (p *plannerVisitor) VisitExpressionStringComparison(node *parse.ExpressionStringComparison) any {
	panic("TODO: Implement")
}

func (p *plannerVisitor) VisitExpressionIs(node *parse.ExpressionIs) any {
	var op ComparisonOperator
	switch {
	case node.Not && node.Distinct:
		op = IsNotDistinctFrom
	case node.Not && !node.Distinct:
		op = IsNot
	case !node.Not && node.Distinct:
		op = IsDistinctFrom
	case !node.Not && !node.Distinct:
		op = Is
	default:
		panic("internal bug: unexpected combination of NOT and DISTINCT")
	}

	return &ComparisonOp{
		Left:  node.Left.Accept(p).(LogicalExpr),
		Right: node.Right.Accept(p).(LogicalExpr),
		Op:    op,
	}
}

func (p *plannerVisitor) VisitExpressionIn(node *parse.ExpressionIn) any {
	panic("TODO: Implement")
}

func (p *plannerVisitor) VisitExpressionBetween(node *parse.ExpressionBetween) any {
	leftOp, rightOp := GreaterThanOrEqual, LessThanOrEqual
	if node.Not {
		leftOp, rightOp = LessThan, GreaterThan
	}

	// we will simply convert this to a logical AND expression of two comparisons
	return &LogicalOp{
		Left: &ComparisonOp{
			Left:  node.Expression.Accept(p).(LogicalExpr),
			Right: node.Lower.Accept(p).(LogicalExpr),
			Op:    leftOp,
		},
		Right: &ComparisonOp{
			Left:  node.Expression.Accept(p).(LogicalExpr),
			Right: node.Upper.Accept(p).(LogicalExpr),
			Op:    rightOp,
		},
		Op: And,
	}
}

func (p *plannerVisitor) VisitExpressionSubquery(node *parse.ExpressionSubquery) any {
	panic("TODO: Implement")
}

func (p *plannerVisitor) VisitExpressionCase(node *parse.ExpressionCase) any {
	panic("TODO: Implement")
}

/*
	all of the following methods should return a LogicalPlan
*/

func (p *plannerVisitor) VisitCommonTableExpression(node *parse.CommonTableExpression) any {
	plan := node.Query.Accept(p).(LogicalPlan)

	rel := plan.Relation(p.schemaCtx()) // we need to check that the columns are valid
	if len(node.Columns) != len(rel.Columns) {
		panic(fmt.Sprintf(`cte "%s" has %d columns, but %d were specified`, node.Name, len(rel.Columns), len(node.Columns)))
	}

	p.ctes = append(p.ctes, struct {
		alias       string
		columnNames []string
		plan        LogicalPlan
	}{
		alias:       node.Name,
		columnNames: node.Columns,
		plan:        plan,
	})

	// I am unsure if we need to return this plan, since all that matters
	// is that it is added to the ctes slice.
	return plan
}

func (p *plannerVisitor) VisitSQLStatement(node *parse.SQLStatement) any {
	panic("TODO: Implement")
}

// The order of building is:
// 1. All select cores.
// 2. Set operations combining the select cores.
// 3. Order by
// 4. Limit and offset
func (p *plannerVisitor) VisitSelectStatement(node *parse.SelectStatement) any {
	if len(node.SelectCores) == 0 {
		panic("no select cores")
	}

	plan := node.SelectCores[0].Accept(p).(LogicalPlan)
	for i, core := range node.SelectCores[1:] {
		plan = &SetOperation{
			Left:   plan,
			Right:  core.Accept(p).(LogicalPlan),
			OpType: get(compoundTypes, node.CompoundOperators[i]),
		}
	}

	if len(node.Ordering) > 0 {
		sort := Sort{
			Child: plan,
		}

		for _, order := range node.Ordering {
			sort.SortExpressions = append(sort.SortExpressions, &SortExpression{
				Expr:      order.Expression.Accept(p).(LogicalExpr),
				Ascending: get(orderAsc, order.Order),
				NullsLast: get(orderNullsLast, order.Nulls),
			})
		}

		plan = &sort
	}

	if node.Limit != nil {
		lim := &Limit{
			Child: plan,
			Limit: node.Limit.Accept(p).(LogicalExpr),
		}

		if node.Offset != nil {
			lim.Offset = node.Offset.Accept(p).(LogicalExpr)
		}

		plan = lim
	}

	return plan
}

// schemaCtx returns a schema context based on the current schema and the cte relations.
// All passed relations will be joined into the schema context.
func (p *plannerVisitor) schemaCtx(relations ...*Relation) *SchemaContext {
	rel := &Relation{}
	for _, rel := range relations {
		rel.Columns = append(rel.Columns, rel.Columns...)
	}

	// we need to calculate the cte relations
	ctx := &SchemaContext{
		Schema:        p.schema,
		CTEs:          make(map[string]*Relation),
		OuterRelation: rel,
	}

	for _, cte := range p.ctes {
		rel := cte.plan.Relation(ctx)

		if len(cte.columnNames) != len(rel.Columns) {
			// this should get caught during construction of the cte
			panic(fmt.Sprintf(`cte "%s" has %d columns, but %d were specified`, cte.alias, len(rel.Columns), len(cte.columnNames)))
		}

		// we need to rename the columns to match the cte column names
		for i, col := range rel.Columns {
			col.Parent = cte.alias
			col.Name = cte.columnNames[i]
		}

		ctx.CTEs[cte.alias] = rel
	}

	return ctx
}

// The order of building is:
// 1. from (combining any joins into single source plan)
// 2. where
// 3. group by(can use reference from select)
// 4. having(can use reference from select)
// 5. select
// 6. distinct
func (p *plannerVisitor) VisitSelectCore(node *parse.SelectCore) any {
	// if there is no from, then we will simply return a projection
	// of the return values on a noop plan.
	if node.From == nil {
		var exprs []LogicalExpr
		for _, resultCol := range node.Columns {
			switch resultCol := resultCol.(type) {
			default:
				panic(fmt.Sprintf("unexpected result column type %T", resultCol))
			case *parse.ResultColumnExpression:
				expr2 := resultCol.Expression.Accept(p).(LogicalExpr)
				if resultCol.Alias != "" {
					expr2 = &AliasExpr{
						Expr:  expr2,
						Alias: resultCol.Alias,
					}
				}

				exprs = append(exprs, expr2)
			case *parse.ResultColumnWildcard:
				// if there is no from, we cannot expand the wildcard
				panic(`wildcard "*" cannot be used without a FROM clause`)
			}
		}

		return &Project{
			Expressions: exprs,
			Child:       &Noop{},
		}
	}

	// otherwise, we will build the plan from the from clause
	plan := node.From.Accept(p).(LogicalPlan)

	for _, join := range node.Joins {
		plan = &Join{
			Left:      plan,
			Right:     join.Relation.Accept(p).(LogicalPlan),
			Condition: join.On.Accept(p).(LogicalExpr),
			JoinType:  get(joinTypes, join.Type),
		}
	}

	if node.Where != nil {
		plan = &Filter{
			Condition: node.Where.Accept(p).(LogicalExpr),
			Child:     plan,
		}
	}

	// despite this being out of order, we will analyze the returned columns,
	// since they are needed for building aggregation
	var results []LogicalExpr
	for _, resultCol := range node.Columns {
		switch resultCol := resultCol.(type) {
		default:
			panic(fmt.Sprintf("unexpected result column type %T", resultCol))
		case *parse.ResultColumnExpression:
			// if expression, we need to ensure it is an aggregate
			expr := resultCol.Expression.Accept(p).(LogicalExpr)
			if resultCol.Alias != "" {
				expr = &AliasExpr{
					Expr:  expr,
					Alias: resultCol.Alias,
				}
			}

			results = append(results, expr)
		case *parse.ResultColumnWildcard:
			var cols []*Column
			// expand the wildcard
			if resultCol.Table != "" {
				cols = plan.Relation(p.schemaCtx()).ColumnsByParent(resultCol.Table)
				if len(cols) == 0 {
					panic(fmt.Sprintf(`table "%s" not found`, resultCol.Table))
				}
			} else {
				cols = plan.Relation(p.schemaCtx()).Columns
			}

			for _, col := range cols {
				results = append(results, &ColumnRef{
					Parent:     col.Parent,
					ColumnName: col.Name,
				})
			}
		}
	}

	if node.GroupBy != nil {
		agg := &Aggregate{
			GroupingExpressions: p.exprs(node.GroupBy),
			Child:               plan,
		}

		// get the current relation prior to analyzing the aggregation
		currentRel := plan.Relation(p.schemaCtx())

		// now we need to check that each unaggregated result is in the grouping expressions
		aggregatedExprs, err := checkAggregation(p.schemaCtx(currentRel), agg.GroupingExpressions, results)
		if err != nil {
			panic(err)
		}

		// set the found aggregated expressions to the aggregation plan.
		agg.AggregateExpressions = aggregatedExprs
		plan = agg

		if node.Having != nil {
			// get the relation of the query after the aggregation
			// to analyze the having clause
			currentRel := plan.Relation(p.schemaCtx())

			// we must also check for aggregation in the grouping expressions.
			// The aggregated expressions will be added to the aggregation plan.
			aggs, err := checkAggregation(p.schemaCtx(currentRel), agg.GroupingExpressions, []LogicalExpr{node.Having.Accept(p).(LogicalExpr)})
			if err != nil {
				panic(err)
			}

			// merge the found aggregated expressions with the existing ones
			agg.AggregateExpressions = mergeEquals(agg.AggregateExpressions, aggs)

			plan = &Filter{
				Condition: node.Having.Accept(p).(LogicalExpr),
				Child:     plan,
			}
		}
	} else {
		currentRel := plan.Relation(p.schemaCtx())

		// we can still have aggregation without grouping, e.g.
		// SELECT COUNT(*) FROM table;
		// in this case, we need to check that all columns are aggregated.
		aggregatedExprs, err := checkAggregation(p.schemaCtx(currentRel), nil, results)
		if err != nil {
			panic(err)
		}

		plan = &Aggregate{
			GroupingExpressions:  nil,
			AggregateExpressions: aggregatedExprs,
			Child:                plan,
		}
	}

	// we need to project the results
	plan = &Project{
		Expressions: results,
		Child:       plan,
	}

	if node.Distinct {
		plan = &Distinct{
			Child: plan,
		}
	}

	return plan
}

// checkAggregation checks that all terms (and their projected columns) are in the
// groupByTerms, or aggregated in an aggregate function. It returns all aggregated terms
func checkAggregation(ctx *SchemaContext, groupByTerms, termsToCheck []LogicalExpr) ([]LogicalExpr, error) {
	// we construct a map for better lookup on the terms included in
	// the grouping expressions

	groupedBy := make(map[[2]string]struct{})
	for _, term := range groupByTerms {
		projected, _, err := term.UsedColumns(ctx)
		if err != nil {
			return nil, err
		}

		for _, p := range projected {
			// we do not care about duplicates, since we are just checking for inclusion
			groupedBy[[2]string{p.Parent, p.Name}] = struct{}{}
		}
	}

	var allAggregatedTerms []LogicalExpr
	for _, term := range termsToCheck {
		projected, aggs, err := term.UsedColumns(ctx)
		if err != nil {
			return nil, err
		}

		allAggregatedTerms = append(allAggregatedTerms, aggs...)

		for _, p := range projected {
			if !p.Aggregated {
				// if the term is not in an aggregate function, it must be in the group by clause
				_, ok := groupedBy[[2]string{p.Parent, p.Name}]
				if !ok {
					return nil, fmt.Errorf(`unaggregated column "%s" must be included in group by clause`, p.Name)
				}
			}
			// if used in an aggregate, then we do not need to care, since both grouped and ungrouped columns are allowed
		}
	}

	return allAggregatedTerms, nil
}

// mergeEquals merges two slices of expressions, and returns a slice of all unique expressions.
// We could use the LogicalExpr.Equals method, but that would be O(n^2) in the worst case,
// and could very easily be attacked by submitted a query with a lot of slightly different expressions.
func mergeEquals(a, b []LogicalExpr) []LogicalExpr {
	// we will use a map to ensure uniqueness
	exprs := make(map[LogicalExpr]struct{})
	for _, expr := range a {
		exprs[expr] = struct{}{}
	}
	for _, expr := range b {
		exprs[expr] = struct{}{}
	}

	// now we convert the map back to a slice
	var result []LogicalExpr
	for expr := range exprs {
		result = append(result, expr)
	}
	return result
}

func (p *plannerVisitor) VisitRelationTable(node *parse.RelationTable) any {
	alias := node.Table
	if node.Alias != "" {
		alias = node.Alias
	}

	return &ScanAlias{
		Child: &TableScan{
			TableName: node.Table,
		},
		Alias: alias,
	}
}

func (p *plannerVisitor) VisitRelationSubquery(node *parse.RelationSubquery) any {
	if node.Alias == "" {
		panic("subquery must have an alias")
	}

	return &ScanAlias{
		Child: node.Subquery.Accept(p).(LogicalPlan),
		Alias: node.Alias,
	}
}

func (p *plannerVisitor) VisitRelationFunctionCall(node *parse.RelationFunctionCall) any {
	if node.Alias == "" {
		panic("joins against function calls must have an alias")
	}

	// the function call must either be a procedure, or foreign procedure, that returns
	// a table.

	var procReturns *types.ProcedureReturn
	var isForeign bool
	proc, found := p.schema.FindProcedure(node.FunctionCall.FunctionName())
	if found {
		procReturns = proc.Returns
	} else {
		foreignProc, found := p.schema.FindForeignProcedure(node.FunctionCall.FunctionName())
		if !found {
			panic(fmt.Sprintf(`no procedure or foreign procedure "%s" found`, node.FunctionCall.FunctionName()))
		}

		procReturns = foreignProc.Returns
		isForeign = true
	}

	if procReturns == nil {
		panic(fmt.Sprintf(`procedure "%s" does not return a table`, node.FunctionCall.FunctionName()))
	}
	if !proc.Returns.IsTable {
		panic(fmt.Sprintf(`procedure "%s" does not return a table`, node.FunctionCall.FunctionName()))
	}

	var args []LogicalExpr
	var contextualArgs []LogicalExpr
	switch t := node.FunctionCall.(type) {
	default:
		panic(fmt.Sprintf("unexpected function call type %T", t))
	case *parse.ExpressionFunctionCall:
		args = p.exprs(t.Args)
	case *parse.ExpressionForeignCall:
		args = p.exprs(t.Args)
		contextualArgs = p.exprs(t.ContextualArgs)
	}

	return &ScanAlias{
		Child: &ProcedureScan{
			ProcedureName:  node.FunctionCall.FunctionName(),
			Args:           args,
			ContextualArgs: contextualArgs,
			IsForeign:      isForeign,
		},
		Alias: node.Alias,
	}
}

func (p *plannerVisitor) VisitUpdateStatement(node *parse.UpdateStatement) any {
	panic("TODO: Implement")
}

func (p *plannerVisitor) VisitUpdateSetClause(node *parse.UpdateSetClause) any {
	panic("TODO: Implement")
}

func (p *plannerVisitor) VisitDeleteStatement(node *parse.DeleteStatement) any {
	panic("TODO: Implement")
}

func (p *plannerVisitor) VisitInsertStatement(node *parse.InsertStatement) any {
	panic("TODO: Implement")
}

func (p *plannerVisitor) VisitUpsertClause(node *parse.UpsertClause) any {
	panic("TODO: Implement")
}

/*
	to make sure that we do not have any unimplemented visitor methods (which would cause unexpected bugs),
	below we include all used ones, and panic if they are called.
*/

func (p *plannerVisitor) VisitResultColumnExpression(node *parse.ResultColumnExpression) any {
	panic("internal bug: VisitResultColumnExpression should not be called directly while building relational algebra")
}

func (p *plannerVisitor) VisitResultColumnWildcard(node *parse.ResultColumnWildcard) any {
	panic("internal bug: VisitResultColumnWildcard should not be called directly while building relational algebra")
}

func (p *plannerVisitor) VisitOrderingTerm(node *parse.OrderingTerm) any {
	panic("internal bug: VisitOrderingTerm should not be called directly while building relational algebra")
}

func (p *plannerVisitor) VisitJoin(node *parse.Join) any {
	// for safety, since it is easier to iterate over the joins in the select core
	panic("internal bug: VisitJoin should not be called directly while building relational algebra")
}
