package planner

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
)

func Plan(statement *parse.SQLStatement, schema *types.Schema, vars map[string]*types.DataType, objects map[string]map[string]*types.DataType) (LogicalPlan, error) {
	if vars == nil {
		vars = make(map[string]*types.DataType)
	}
	if objects == nil {
		objects = make(map[string]map[string]*types.DataType)
	}

	ctx := &planContext{
		Schema:    schema,
		CTEs:      make(map[string]*Relation),
		Variables: vars,
		Objects:   objects,
	}

	visitor := &plannerVisitor{
		planCtx: ctx,
		schema:  schema,
	}

	return statement.Accept(visitor).(LogicalPlan), nil
}

// the planner file converts the parse AST into a logical query plan.

type plannerVisitor struct {
	parse.UnimplementedSqlVisitor
	planCtx *planContext

	// schema is the underlying schema that the AST was parsed against.
	schema *types.Schema
}

// planContext holds information that is needed during the planning process.
type planContext struct {
	// Schema is the underlying database schema that the query should
	// be evaluated against.
	Schema *types.Schema
	// CTEs are the common table expressions in the query.
	// This field should be updated as the query planner
	// processes the query.
	CTEs map[string]*Relation
	// Variables are the variables in the query.
	Variables map[string]*types.DataType
	// Objects are the objects in the query.
	// Kwil supports one-dimensional objects, so this would be
	// accessible via objname.fieldname.
	Objects map[string]map[string]*types.DataType
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
	var args []LogicalExpr
	for _, arg := range node.Args {
		args = append(args, arg.Accept(p).(LogicalExpr))
	}

	// can be either a procedure call or a built-in function
	funcDef, ok := parse.Functions[node.Name]
	if !ok {
		if node.Star {
			panic("star (*) not allowed in procedure calls")
		}
		if node.Distinct {
			panic("DISTINCT not allowed in procedure calls")
		}

		// must be a procedure call
		_, found := p.schema.FindProcedure(node.Name)
		if !found {
			panic(fmt.Sprintf(`no function or procedure "%s" found`, node.Name))
		}

		return cast(&ProcedureCall{
			ProcedureName: node.Name,
			Args:          args,
		}, node)
	}

	// now we need to apply rules depending on if it is aggregate or not
	if funcDef.IsAggregate {
		return cast(&AggregateFunctionCall{
			FunctionName: node.Name,
			Args:         args,
			Star:         node.Star,
			Distinct:     node.Distinct,
		}, node)
	}

	if node.Star {
		panic("star (*) not allowed in non-aggregate function calls")
	}
	if node.Distinct {
		panic("DISTINCT not allowed in non-aggregate function calls")
	}

	return cast(&ScalarFunctionCall{
		FunctionName: node.Name,
		Args:         p.exprs(node.Args),
	}, node)
}

func (p *plannerVisitor) VisitExpressionForeignCall(node *parse.ExpressionForeignCall) any {
	_, found := p.schema.FindForeignProcedure(node.Name)
	if !found {
		panic(fmt.Sprintf(`no foreign procedure "%s" found`, node.Name))
	}

	if len(node.ContextualArgs) != 2 {
		panic("foreign calls must have 2 contextual arguments")
	}

	return cast(&ProcedureCall{
		ProcedureName: node.Name,
		Foreign:       true,
		Args:          p.exprs(node.Args),
		ContextArgs:   p.exprs(node.ContextualArgs),
	}, node)
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
	subqType := ScalarSubquery
	if node.Exists {
		subqType = ExistsSubquery
		if node.Not {
			subqType = NotExistsSubquery
		}
	}

	stmt := node.Subquery.Accept(p).(LogicalPlan)
	return cast(&Subquery{
		SubqueryType: subqType,
		Query:        stmt,
	}, node)
}

func (p *plannerVisitor) VisitExpressionCase(node *parse.ExpressionCase) any {
	panic("TODO: Implement")
}

/*
	all of the following methods should return a LogicalPlan
*/

func (p *plannerVisitor) VisitCommonTableExpression(node *parse.CommonTableExpression) any {
	// still have a bit to do here.
	panic("CTE not yet supported")
	plan := node.Query.Accept(p).(LogicalPlan)

	rel, err := newEvalCtx(p.planCtx).evalRelation(plan)
	if err != nil {
		panic(err)
	}

	// we need to check that the columns are valid
	if len(node.Columns) != len(rel.Fields) {
		panic(fmt.Sprintf(`cte "%s" has %d columns, but %d were specified`, node.Name, len(rel.Fields), len(node.Columns)))
	}

	p.planCtx.CTEs[node.Name] = rel

	// I am unsure if we need to return this plan, since all that matters
	// is that it is added to the ctes slice.
	return plan
}

func (p *plannerVisitor) VisitSQLStatement(node *parse.SQLStatement) any {
	for _, cte := range node.CTEs {
		cte.Accept(p)
	}

	return node.SQL.Accept(p)
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

	// TODO: if there is more than 1 select core, then the passed context needs to be updated
	// to only have the relation of the RETURN VALUE of the first select core.
	// For example, "SELECT id from users order by name" is valid, but
	// "SELECT id from users UNION SELECT id from users2 order by name" is not.
	// ? Should this instead go into the relational algebra that represents compound operations? - yes

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

// The order of building is:
// 1. from (combining any joins into single source plan)
// 2. where
// 3. group by(can use reference from select)
// 4. having(can use reference from select)
// 5. select (project)
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
			Child:       &EmptyScan{},
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

	// we analyze the returned columns to see if there are any aggregates
	// we will revisit them later for the full analysis after GROUP BY
	var aggs []*AggregateFunctionCall
	for _, resultCol := range node.Columns {
		if resultCol, ok := resultCol.(*parse.ResultColumnExpression); ok {
			logicalExpr := resultCol.Expression.Accept(p).(LogicalExpr)
			found := getAggregateTerms(logicalExpr)
			aggs = append(aggs, found...)
		}
	}

	if node.GroupBy != nil {
		agg := &Aggregate{
			GroupingExpressions:  p.exprs(node.GroupBy),
			AggregateExpressions: aggs,
			Child:                plan,
		}

		plan = agg

		if node.Having != nil {
			havingExpr := node.Having.Accept(p).(LogicalExpr)
			havingAggs := getAggregateTerms(havingExpr)
			agg.AggregateExpressions = mergeAggregates(agg.AggregateExpressions, havingAggs)

			plan = &Filter{
				Condition: havingExpr,
				Child:     plan,
			}
		}
	} else if len(aggs) > 0 { // otherwise, still need to see if we have any aggregates without grouping
		plan = &Aggregate{
			GroupingExpressions:  nil,
			AggregateExpressions: aggs,
			Child:                plan,
		}
	}

	// now, we re-analyze results and expand wildcards
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
			// expand the wildcard
			rel, err := newEvalCtx(p.planCtx).evalRelation(plan)
			if err != nil {
				panic(err)
			}

			var fields []*Field
			if resultCol.Table != "" {
				fields = rel.ColumnsByParent(resultCol.Table)
			} else {
				fields = rel.Fields
			}

			for _, col := range fields {
				results = append(results, &ColumnRef{
					Parent:     col.Parent,
					ColumnName: col.Name,
				})
			}
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

func (p *plannerVisitor) VisitRelationTable(node *parse.RelationTable) any {
	alias := node.Table
	if node.Alias != "" {
		alias = node.Alias
	}

	return &Scan{
		Source: &TableScanSource{
			TableName: node.Table,
		},
		RelationName: alias,
	}
}

func (p *plannerVisitor) VisitRelationSubquery(node *parse.RelationSubquery) any {
	if node.Alias == "" {
		panic("subquery must have an alias")
	}

	subq := node.Subquery.Accept(p).(LogicalPlan)

	return &Scan{
		Source:       &SubqueryScanSource{Subquery: subq},
		RelationName: node.Alias,
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

	return &Scan{
		Source: &ProcedureScanSource{
			ProcedureName:  node.FunctionCall.FunctionName(),
			Args:           args,
			ContextualArgs: contextualArgs,
			IsForeign:      isForeign,
		},
		RelationName: node.Alias,
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
