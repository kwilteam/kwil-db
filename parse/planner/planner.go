package planner

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils/order"
	"github.com/kwilteam/kwil-db/parse"
)

func Plan(statement *parse.SQLStatement, schema *types.Schema, vars map[string]*types.DataType, objects map[string]map[string]*types.DataType) (analyzed *AnalyzedPlan, err error) {
	defer func() {
		if r := recover(); r != nil {
			err2, ok := r.(error)
			if !ok {
				err2 = fmt.Errorf("%v", r)
			}
			err = err2
		}
	}()

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

	lp := statement.Accept(visitor).(LogicalPlan)

	return &AnalyzedPlan{
		Plan: lp,
		CTEs: ctx.CTEPlans,
	}, nil
}

// AnalyzedPlan is the full result of a logical plan that has been analyzed.
type AnalyzedPlan struct {
	// Plan is the plan of the query.
	Plan LogicalPlan
	// CTEs are plans for the common table expressions in the query.
	// They are in the order that they were defined.
	CTEs []*Subplan
}

// Format formats the plan into a human-readable string.
func (a *AnalyzedPlan) Format() string {
	res := Format(a.Plan)

	str := strings.Builder{}
	str.WriteString(res)

	// we will copy and reverse the cte list for printing, so that
	// any CTE that references another CTE will be above it.
	// This matches the printing of subqueries.
	cte2 := slices.Clone(a.CTEs)
	slices.Reverse(cte2)

	printSubplans(&str, cte2)

	return str.String()
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
	// CTEPlans are the logical plans for the common table expressions
	// in the query. This field should be updated as the query planner
	// processes the query.
	CTEPlans []*Subplan
	// Variables are the variables in the query.
	Variables map[string]*types.DataType
	// Objects are the objects in the query.
	// Kwil supports one-dimensional objects, so this would be
	// accessible via objname.fieldname.
	Objects map[string]map[string]*types.DataType
	// SubqueryCount is the number of subqueries in the query.
	// This field should be updated as the query planner
	// processes the query.
	SubqueryCount int
	// ReferenceCount is the number of references in the query.
	// This field should be updated as the query planner
	// processes the query.
	ReferenceCount int
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

var stringComparisonOps = map[parse.StringComparisonOperator]ComparisonOperator{
	parse.StringComparisonOperatorLike:  Like,
	parse.StringComparisonOperatorILike: ILike,
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
		// If an aggregate, we wrap it in an ExprRef, so that later, we can
		// replace it with a reference to the aggregate in the aggregate node.

		// return cast(&AggregateFunctionCall{
		// 	FunctionName: node.Name,
		// 	Args:         args,
		// 	Star:         node.Star,
		// 	Distinct:     node.Distinct,
		// }, node)

		// we apply cast outside the reference because we want to keep the reference
		// specific to the aggregate function call.
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
		VarName: node.String(),
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
		Key:    node.Field,
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
	c := &Collate{
		Expr: node.Expression.Accept(p).(LogicalExpr),
	}

	switch strings.ToLower(node.Collation) {
	case "nocase":
		c.Collation = NoCaseCollation
	default:
		panic(fmt.Sprintf(`unknown collation "%s"`, node.Collation))
	}

	return c
}

func (p *plannerVisitor) VisitExpressionStringComparison(node *parse.ExpressionStringComparison) any {
	var expr LogicalExpr = &ComparisonOp{
		Left:  node.Left.Accept(p).(LogicalExpr),
		Right: node.Right.Accept(p).(LogicalExpr),
		Op:    get(stringComparisonOps, node.Operator),
	}

	if node.Not {
		expr = &UnaryOp{
			Expr: expr,
			Op:   Not,
		}
	}

	return expr
}

func (p *plannerVisitor) VisitExpressionIs(node *parse.ExpressionIs) any {
	op := Is
	if node.Distinct {
		op = IsDistinctFrom
	}

	var plan LogicalExpr = &ComparisonOp{
		Left:  node.Left.Accept(p).(LogicalExpr),
		Right: node.Right.Accept(p).(LogicalExpr),
		Op:    op,
	}

	if node.Not {
		plan = &UnaryOp{
			Expr: plan,
			Op:   Not,
		}
	}

	return plan
}

func (p *plannerVisitor) VisitExpressionIn(node *parse.ExpressionIn) any {
	in := &IsIn{
		Left: node.Expression.Accept(p).(LogicalExpr),
	}

	if node.Subquery != nil {
		in.Subquery = node.Subquery.Accept(p).(*SubqueryExpr)
	} else {
		in.Expressions = p.exprs(node.List)
	}

	var expr LogicalExpr = in

	if node.Not {
		expr = &UnaryOp{
			Expr: expr,
			Op:   Not,
		}
	}

	return expr
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

	var sub LogicalExpr = &SubqueryExpr{
		Query: &Subquery{
			Plan: &Subplan{
				Plan: node.Subquery.Accept(p).(LogicalPlan),
				ID:   strconv.Itoa(p.planCtx.SubqueryCount),
				Type: SubplanTypeSubquery,
			},
		},
		Exists: node.Exists,
	}
	p.planCtx.SubqueryCount++

	if node.Exists && node.Not {
		sub = &UnaryOp{
			Expr: sub,
			Op:   Not,
		}
	}

	return sub
}

func (p *plannerVisitor) VisitExpressionCase(node *parse.ExpressionCase) any {
	c := &Case{}

	if node.Case != nil {
		c.Value = node.Case.Accept(p).(LogicalExpr)
	}

	for _, when := range node.WhenThen {
		c.WhenClauses = append(c.WhenClauses, [2]LogicalExpr{
			when[0].Accept(p).(LogicalExpr),
			when[1].Accept(p).(LogicalExpr),
		})
	}

	if node.Else != nil {
		c.Else = node.Else.Accept(p).(LogicalExpr)
	}

	return c
}

/*
	all of the following methods should return a LogicalPlan
*/

func (p *plannerVisitor) VisitCommonTableExpression(node *parse.CommonTableExpression) any {
	// still have a bit to do here.
	plan := node.Query.Accept(p).(LogicalPlan)

	rel, err := newEvalCtx(p.planCtx).evalRelation(plan)
	if err != nil {
		panic(err)
	}

	var extraInfo string // debug info

	// if there are columns specific, we need to check that the columns are valid
	// and rename the relation fields
	if len(node.Columns) > 0 {
		if len(node.Columns) != len(rel.Fields) {
			panic(fmt.Sprintf(`cte "%s" has %d columns, but %d were specified`, node.Name, len(rel.Fields), len(node.Columns)))
		}

		for i, col := range node.Columns {
			extraInfo += fmt.Sprintf(" [%s.%s -> %s]", rel.Fields[i].Parent, rel.Fields[i].Name, col)

			rel.Fields[i].Parent = node.Name
			rel.Fields[i].Name = col
		}
	} else {
		// otherwise, we need to rename the relation parents
		// to the CTE's name
		for _, field := range rel.Fields {
			extraInfo += fmt.Sprintf(" [%s.%s -> %s]", field.Parent, field.Name, field.Name)
			field.Parent = node.Name
		}
	}

	p.planCtx.CTEs[node.Name] = rel
	p.planCtx.CTEPlans = append(p.planCtx.CTEPlans, &Subplan{
		Plan:      plan,
		ID:        node.Name,
		Type:      SubplanTypeCTE,
		extraInfo: extraInfo,
	})

	// I am unsure if we need to return this plan, since all that matters
	// is that it is added to the ctes slice.
	return plan
}

func (p *plannerVisitor) VisitSQLStatement(node *parse.SQLStatement) any {
	for _, cte := range node.CTEs {
		cte.Accept(p)
	}

	stmt := node.SQL.Accept(p).(LogicalPlan)

	rel, err := newEvalCtx(p.planCtx).evalRelation(stmt)
	if err != nil {
		panic(err)
	}
	rel = rel.Copy()

	// if it a sql select, we should add a return operation.
	// We can't add this within VisitSelectStatement because
	// we don't want to add it to subqueries.
	if _, ok := node.SQL.(*parse.SelectStatement); ok {
		var fields []*Field
		for _, col := range rel.Fields {
			if col.Name == "" {
				col.Name = "?column?"
			}

			fields = append(fields, col)
		}

		return &Return{
			Fields: fields,
			Child:  stmt,
		}
	}

	return stmt
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

	var plan LogicalPlan

	selectCore := node.SelectCores[0].Accept(p).(*selectCoreResult)

	// finish is a function that is called at the end of the function.
	// Normally, we would handle this with a defer, but since the visitor has
	// to return "any", this is not an option. By default, it does nothing,
	// but in the logic directly below, we might set it to apply a projection
	finish := func(pln LogicalPlan) LogicalPlan {
		return pln
	}

	// see the documentation for selectCoreResult for an explanation as to why
	// we perform this if statement.
	if len(node.SelectCores) == 1 {
		plan = selectCore.plan
		finish = selectCore.projectFunc
	} else {
		// otherwise, apply immediately
		plan = selectCore.projectFunc(selectCore.plan)
		for i, core := range node.SelectCores[1:] {
			right := core.Accept(p).(*selectCoreResult)

			plan = &SetOperation{
				Left:   plan,
				Right:  right.projectFunc(right.plan),
				OpType: get(compoundTypes, node.CompoundOperators[i]),
			}
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

	res := finish(plan)
	return res
}

// The order of building is:
// 1. from (combining any joins into single source plan)
// 2. where
// 3. group by(can use reference from select)
// 4. having(can use reference from select)
// 5. select (project)
// 6. distinct
// ! This method is insanely complex and needs to be refactored.
func (p *plannerVisitor) VisitSelectCore(node *parse.SelectCore) any {
	/*
		If a user does "SELECT sum(id), age/2 from users group by age/2", what needs to happen is:
			- project ref(a), ref(b)
			- aggregate [sum(id) as a] group by [age/2 as b]
			- scan users

		In order to do this, we need to be able to:
			a. match and rewrite all aggregate functions in having and return to be ExprRef
			b. match any arbitrary tree in the grouping to any term in the having and return clause, and then rewrite them as ExprRef
			c. recognize exprefs by name in both the aggregate and other clause to ensure they are accessible within that scope (maybe)?

	*/
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

		isDistinct := node.Distinct

		return &selectCoreResult{
			plan: &EmptyScan{},
			projectFunc: func(newPlan LogicalPlan) LogicalPlan {
				var p LogicalPlan = &Project{
					Child:       newPlan,
					expandFuncs: []expandFunc{func() []LogicalExpr { return exprs }},
				}

				if isDistinct {
					p = &Distinct{
						Child: p,
					}
				}

				return p
			},
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

	lastJoin := plan // will be referenced later to expand aggregate functions

	if node.Where != nil {
		plan = &Filter{
			Condition: node.Where.Accept(p).(LogicalExpr),
			Child:     plan,
		}
	}

	// aggTerms maps the signature of an aggregate function to its identified expression.
	// It is used to rewrite the having and return clauses to reference aggregate expressions.
	aggTerms := make(map[string]*IdentifiedExpr)

	// groupingTerms is a list of all terms in the group by clause.
	// It is used to cut and reference terms from the having clause
	// and result columns
	groupingTerms := make(map[string]*IdentifiedExpr)

	// havingExpr is the expression that will be used in the having clause.
	// If there is no having clause, this will be nil.
	var havingExpr LogicalExpr

	// wrapAggPlan is a function that wraps the plan in an aggregate, and optionally
	// a filter (if there is a having clause).
	wrapAggPlan := func(plan LogicalPlan) LogicalPlan {
		if len(aggTerms)+len(groupingTerms) == 0 {
			return plan
		}

		agg := &Aggregate{
			Child: plan,
		}

		// for determinism
		for _, aggTerm := range order.OrderMap(aggTerms) {
			agg.AggregateExpressions = append(agg.AggregateExpressions, aggTerm.Value)
		}

		for _, groupTerm := range order.OrderMap(groupingTerms) {
			agg.GroupingExpressions = append(agg.GroupingExpressions, groupTerm.Value)
		}

		if havingExpr != nil {
			return &Filter{
				Condition: havingExpr,
				Child:     agg,
			}
		}

		return agg
	}

	if node.GroupBy != nil {
		for _, expr := range node.GroupBy {
			// if there is a duplicate, we will ignore it
			logicalExpr := expr.Accept(p).(LogicalExpr)
			_, ok := groupingTerms[logicalExpr.String()]
			if ok {
				continue
			}

			identified := &IdentifiedExpr{
				Expr: logicalExpr,
				ID:   numberToLetters(p.planCtx.ReferenceCount),
			}
			p.planCtx.ReferenceCount++

			groupingTerms[logicalExpr.String()] = identified
		}

		if node.Having != nil {
			havingExpr = node.Having.Accept(p).(LogicalExpr)

			// rewrite all aggregate functions to be a reference to the identified expression.
			// If we find any, we add them to the mapping of aggTerms
			havingNode, err := Rewrite(havingExpr, &RewriteConfig{
				ExprCallback: func(le LogicalExpr) (LogicalExpr, bool, error) {
					switch n := le.(type) {
					case *AggregateFunctionCall:
						existingIdent, ok := aggTerms[n.String()]
						if ok {
							return &ExprRef{
								Identified: existingIdent,
							}, false, nil
						}

						ident := &IdentifiedExpr{
							Expr: n,
							ID:   numberToLetters(p.planCtx.ReferenceCount),
						}
						p.planCtx.ReferenceCount++
						aggTerms[n.String()] = ident
						return &ExprRef{
							Identified: ident,
						}, false, nil
					}
					return le, true, nil
				},
			})
			if err != nil {
				panic(err)
			}

			havingExpr = havingNode.(LogicalExpr)
		}
	}

	// now, we visit the result columns. We will attempt to rewrite any aggregate functions
	// or expressions that match

	// expandFuncs delay expanding wildcards until we can evaluate the relation
	// they are projecting on. We have to do this at an entirely different step
	// once the entire tree is constructed (in evalRelation).
	var expandFuncs []expandFunc

	// TODO: this doesn't work because in order to expand the wildcard, we need to know
	// the relation that is joined. The relation joined should be PRIOR to any aggregate,
	// but with expandFuncs, this occurs after the aggregate. Take for example:
	// "SELECT * FROM users GROUP BY name". In this query, we need to expand all columns
	// from the users table, but due to the way expandFuncs work, we are expanding only
	// the aggregate result. IMO it seems like we should do something to get rid of expandFuncs,
	// however I am not sure what, since we don't know what we are expanding until we evaluate
	// relations. It seems like we need to run the expand funcs directly after all of the joins
	// are performed (but the result must take place after the aggregate).
	for _, resultCol := range node.Columns {
		switch resultCol := resultCol.(type) {
		default:
			panic(fmt.Sprintf("unexpected result column type %T", resultCol))
		case *parse.ResultColumnExpression:
			logicalExpr := resultCol.Expression.Accept(p).(LogicalExpr)

			// attempt to rewrite, if necessary
			logicalNode, err := Rewrite(logicalExpr, &RewriteConfig{
				ExprCallback: func(le LogicalExpr) (LogicalExpr, bool, error) {
					switch n := le.(type) {
					case *AggregateFunctionCall:
						existingIdent, ok := aggTerms[n.String()]
						if ok {
							return &ExprRef{
								Identified: existingIdent,
							}, false, nil
						}

						ident := &IdentifiedExpr{
							Expr: n,
							ID:   numberToLetters(p.planCtx.ReferenceCount),
						}
						p.planCtx.ReferenceCount++

						aggTerms[n.String()] = ident
						return &ExprRef{
							Identified: ident,
						}, false, nil
					case *ExprRef, *IdentifiedExpr:
						// if it has already been replaced, do not replace it again
						return le, false, nil
					default:
						groupingTerm, ok := groupingTerms[n.String()]
						if ok {
							return &ExprRef{
								Identified: groupingTerm,
							}, false, nil
						}

						return le, true, nil
					}
				},
			})
			if err != nil {
				panic(err)
			}
			logicalExpr = logicalNode.(LogicalExpr)

			if resultCol.Alias != "" {
				logicalExpr = &AliasExpr{
					Expr:  logicalExpr,
					Alias: resultCol.Alias,
				}
			}

			newFunc := func() []LogicalExpr {
				return []LogicalExpr{logicalExpr}
			}

			expandFuncs = append(expandFuncs, newFunc)
		case *parse.ResultColumnWildcard:
			// avoid loop variable capture
			tbl := resultCol.Table

			// expand the wildcard
			// wrap any other expandFunc in a new function that will
			// expand the current wildcard
			newExpand := func() []LogicalExpr {
				rel := lastJoin.Relation()

				var newFields []*Field
				if tbl != "" {
					newFields = rel.ColumnsByParent(tbl)
				} else {
					newFields = rel.Fields
				}

				var exprs []LogicalExpr
				for _, field := range newFields {
					var colRef LogicalExpr = &ColumnRef{
						// we don't immediately set the parent, because we need to
						// check if this same field is in the grouping terms.
						// If it is, then the term in the grouping terms will be used
						// (which is already qualified).
						ColumnName: field.Name,
					}

					// if it is in the grouping terms, we need to replace it with a reference.
					groupingTerm, ok := groupingTerms[colRef.String()]
					if ok {
						colRef = &ExprRef{
							Identified: groupingTerm,
						}
					} else {
						// if not, then we can qualify.
						colRef.(*ColumnRef).Parent = field.Parent
					}

					exprs = append(exprs, colRef)
				}

				return exprs
			}

			expandFuncs = append(expandFuncs, newExpand)
		}
	}

	// see the selectCoreResult documentation below for an explanation as to
	// why this is returned instead of just the plan.
	plan = wrapAggPlan(plan)

	return &selectCoreResult{
		plan: plan,
		projectFunc: func(newPlan LogicalPlan) LogicalPlan {
			var plan2 LogicalPlan = &Project{
				Child:       newPlan,
				expandFuncs: expandFuncs,
			}
			if node.Distinct {
				plan2 = &Distinct{
					Child: plan2,
				}
			}

			return plan2
		},
	}
}

// selectCoreResult is a helper struct that is only returned from VisitSelectCore.
// It is returned from VisitSelectCore because we need to handle conditionally
// adding projection. If a query has a SET (a.k.a. compound) operation, we want to project before performing
// the set. If a query has one select, then we want to project after sorting and limiting.
// To give a concrete example of this, imagine a table users (id int, name text) with the queries:
// 1.
// "SELECT name FROM users ORDER BY id" - this is valid in Postgres, and since we can access "id", projection
// should be done after sorting.
// 2.
// "SELECT name FROM users UNION 'hello' ORDER BY id" - this is invalid in Postgres, since "id" is not in the
// result set. We need to project before the UNION.
//
// This struct allows us to conditionally handle this logic in the calling VisitSelectStatement method.
type selectCoreResult struct {
	// plan is the plan that is returned from the select core prior
	// to applying any projection.
	plan LogicalPlan
	// projectFunc is a function that will apply a projection to the plan.
	// it is not directly applied within the select core because if there
	// are multiple select cores, we need to apply the projection before
	// the set operation, but if there is 1, we want to apply the projection
	// after the sort and limit.
	projectFunc func(LogicalPlan) LogicalPlan
}

func (p *plannerVisitor) VisitRelationTable(node *parse.RelationTable) any {
	alias := node.Table
	if node.Alias != "" {
		alias = node.Alias
	}

	var scanTblType TableSourceType
	// determine the type:
	if _, ok := p.schema.FindTable(node.Table); ok {
		scanTblType = TableSourcePhysical
	} else if _, ok = p.planCtx.CTEs[node.Table]; ok {
		scanTblType = TableSourceCTE
	} else {
		panic(fmt.Sprintf(`no table or cte "%s" found`, node.Table))
	}

	return &Scan{
		Source: &TableScanSource{
			TableName: node.Table,
			Type:      scanTblType,
		},
		RelationName: alias,
	}
}

func (p *plannerVisitor) VisitRelationSubquery(node *parse.RelationSubquery) any {
	if node.Alias == "" {
		panic("subquery must have an alias")
	}

	subq := node.Subquery.Accept(p).(LogicalPlan)

	s := &Scan{
		Source: &Subquery{
			ReturnsRelation: true,
			Plan: &Subplan{
				Plan: subq,
				ID:   strconv.Itoa(p.planCtx.SubqueryCount),
				Type: SubplanTypeSubquery,
			},
			// Correlated will be set later
		},
		RelationName: node.Alias,
	}

	p.planCtx.SubqueryCount++
	return s
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
	if !procReturns.IsTable {
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

// ALIAS: aliases can be used everywhere in update statements except in the SET clause
func (p *plannerVisitor) VisitUpdateStatement(node *parse.UpdateStatement) any {
	plan, err := p.buildCartesian(node.Table, node.Alias, node.From, node.Joins, node.Where)
	if err != nil {
		panic(err)
	}

	assignments := make([]*Assignment, len(node.SetClause))
	for i, set := range node.SetClause {
		assignments[i] = &Assignment{
			Column: set.Column,
			Value:  set.Value.Accept(p).(LogicalExpr),
		}
	}

	return &Update{
		Child:       plan,
		Assignments: assignments,
		Table:       node.Table,
	}
}

func (p *plannerVisitor) VisitDeleteStatement(node *parse.DeleteStatement) any {
	plan, err := p.buildCartesian(node.Table, node.Alias, node.From, node.Joins, node.Where)
	if err != nil {
		panic(err)
	}

	return &Delete{
		Child: plan,
		Table: node.Table,
	}
}

// buildCartesian builds a cartesian product for several relations. It is meant to be used
// explicitly for update and delete, where we start by planning a cartesian join between the
// target table and the FROM + JOIN tables, and later optimize the filter.
func (p *plannerVisitor) buildCartesian(targetTable, alias string, from parse.Table, joins []*parse.Join, filter parse.Expression) (LogicalPlan, error) {
	_, ok := p.schema.FindTable(targetTable)
	if !ok {
		return nil, fmt.Errorf(`unknown table "%s"`, targetTable)
	}
	if alias == "" {
		alias = targetTable
	}

	var targetRel LogicalPlan = &Scan{
		Source: &TableScanSource{
			TableName: targetTable,
			Type:      TableSourcePhysical,
		},
		RelationName: alias,
	}

	// if there is no FROM clause, we will simply return the target relation
	if from == nil {
		if filter != nil {
			filterExpr := filter.Accept(p).(LogicalExpr)
			return &Filter{
				Condition: filterExpr,
				Child:     targetRel,
			}, nil
		}

		return targetRel, nil
	}

	// we will build the source rel, and then apply the cartesian join
	var sourceRel LogicalPlan = from.Accept(p).(LogicalPlan)

	for _, join := range joins {
		sourceRel = &Join{
			Left:      sourceRel,
			Right:     join.Relation.Accept(p).(LogicalPlan),
			Condition: join.On.Accept(p).(LogicalExpr),
			JoinType:  get(joinTypes, join.Type),
		}
	}

	// cartesian product with filter
	targetRel = &CartesianProduct{
		Left:  targetRel,
		Right: sourceRel,
	}

	if filter == nil {
		return nil, fmt.Errorf("a WHERE clause must be provided for update and delete statements that use FROM")
	}

	return &Filter{
		Condition: filter.Accept(p).(LogicalExpr),
		Child:     targetRel,
	}, nil
}

// ALIAS: the alias can only be used in the "ON CONFLICT DO UPDATE SET ... WHERE [here]" clause
func (p *plannerVisitor) VisitInsertStatement(node *parse.InsertStatement) any {
	ins := &Insert{
		Table: node.Table,
		Alias: node.Alias,
	}

	tbl, found := p.schema.FindTable(node.Table)
	if !found {
		panic(fmt.Sprintf(`unknown table "%s"`, node.Table))
	}

	// orderAndFillNulls is a helper function that orders logical expressions
	// according to their position in the table, and fills in nulls for any
	// columns that were not specified in the insert. It starts as being empty,
	// since it only needs logic if the user specifies columns.
	orderAndFillNulls := func(exprs []LogicalExpr) []LogicalExpr {
		return exprs
	}

	// if Columns are specified, then the second dimension of the Values
	// must exactly match the number of columns. Otherwise, the second
	// dimension of Values must exactly match the number of columns in the table.
	var expectedColLen int
	var expectedColTypes []*types.DataType // TODO: delete, we check this in eval
	if len(node.Columns) > 0 {
		expectedColLen = len(node.Columns)

		// check if the columns are valid
		var err error
		expectedColTypes, err = checkNullableColumns(tbl, node.Columns)
		if err != nil {
			panic(err)
		}

		// we need to set the orderAndFillNulls function
		// We will do this by creating a map of the position
		// of a specified column's position to its column index in the table.

		// first, we will create a map of the table's columns
		tableColPos := make(map[string]int, len(tbl.Columns))
		for i, col := range tbl.Columns {
			tableColPos[col.Name] = i
		}

		colPos := make(map[int]int, len(node.Columns))
		for i, col := range node.Columns {
			colPos[i] = tableColPos[col]
		}

		orderAndFillNulls = func(exprs []LogicalExpr) []LogicalExpr {
			newExprs := make([]LogicalExpr, len(tbl.Columns))

			for i, expr := range exprs {
				newExprs[colPos[i]] = expr
			}

			for i := range tbl.Columns {
				if newExprs[i] != nil {
					continue
				}

				newExprs[i] = &Literal{
					Value: nil,
					Type:  types.NullType.Copy(),
				}
			}

			return newExprs
		}

	} else {
		expectedColLen = len(tbl.Columns)

		for _, col := range tbl.Columns {
			expectedColTypes = append(expectedColTypes, col.Type.Copy())
		}
	}

	for _, vals := range node.Values {
		if len(vals) != expectedColLen {
			panic(fmt.Sprintf("expected %d insert values, got %d", expectedColLen, len(vals)))
		}

		var row []LogicalExpr

		for i, val := range vals {
			individualVal := val.Accept(p).(LogicalExpr)

			// get the expected type
			field, err := newEvalCtx(p.planCtx).evalExpression(individualVal, &Relation{})
			if err != nil {
				panic(err)
			}

			scalar, err := field.Scalar()
			if err != nil {
				panic(err)
			}

			if !expectedColTypes[i].Equals(scalar) {
				panic(fmt.Sprintf("expected type %s for insert position %d, got %s", expectedColTypes[i], i+1, scalar))
			}

			row = append(row, individualVal)
		}

		ins.Values = append(ins.Values, orderAndFillNulls(row))
	}

	// finally, we need to check if there is an ON CONFLICT clause,
	// and if so, we need to process it.
	if node.Upsert != nil {
		conflict, err := p.buildUpsert(node.Upsert, tbl)
		if err != nil {
			panic(err)
		}

		ins.ConflictResolution = conflict
	}

	return ins
}

func (p *plannerVisitor) buildUpsert(node *parse.UpsertClause, table *types.Table) (ConflictResolution, error) {
	// all DO UPDATE upserts need to have an arbiter index.
	// DO NOTHING can optionally have one, but it is not required.
	var arbiterIndex Index
	switch len(node.ConflictColumns) {
	// must be a unique index or pk that exactly matches the columns
	case 0:
		// do nothing
	case 1:
		// check the column for a unique or pk contraint, as well as all indexes
		col, ok := table.FindColumn(node.ConflictColumns[0])
		if !ok {
			return nil, fmt.Errorf(`conflict column "%s" not found in table`, node.ConflictColumns[0])
		}

		if col.HasAttribute(types.PRIMARY_KEY) {
			arbiterIndex = &IndexColumnConstraint{
				Table:          table.Name,
				Column:         col.Name,
				ConstraintType: PrimaryKeyConstraintIndex,
			}
		} else if col.HasAttribute(types.UNIQUE) {
			arbiterIndex = &IndexColumnConstraint{
				Table:          table.Name,
				Column:         col.Name,
				ConstraintType: UniqueConstraintIndex,
			}
		} else {
			// check all indexes for unique indexes that match the column
			for _, idx := range table.Indexes {
				if (idx.Type == types.UNIQUE_BTREE || idx.Type == types.PRIMARY) && len(idx.Columns) == 1 && idx.Columns[0] == col.Name {
					arbiterIndex = &IndexNamed{
						Name: idx.Name,
					}
				}
			}
		}

		if arbiterIndex == nil {
			return nil, fmt.Errorf(`conflict column "%s" must be have a unique index or primary key`, node.ConflictColumns[0])
		}
	default:
		// check all indexes for a unique or pk index that matches the columns
		for _, idx := range table.Indexes {
			if idx.Type != types.UNIQUE_BTREE && idx.Type != types.PRIMARY {
				continue
			}

			if len(idx.Columns) != len(node.ConflictColumns) {
				continue
			}

			inIdxCols := make(map[string]struct{}, len(idx.Columns))
			for _, col := range idx.Columns {
				inIdxCols[col] = struct{}{}
			}

			hasAllCols := true
			for _, col := range node.ConflictColumns {
				_, ok := inIdxCols[col]
				if !ok {
					hasAllCols = false
					break
				}
			}

			if hasAllCols {
				arbiterIndex = &IndexNamed{
					Name: idx.Name,
				}
				break
			}
		}

		if arbiterIndex == nil {
			return nil, fmt.Errorf(`conflict columns must have a unique index or primary key`)
		}
	}

	if len(node.DoUpdate) == 0 {
		return &ConflictDoNothing{
			ArbiterIndex: arbiterIndex,
		}, nil
	}
	if node.ConflictWhere != nil {
		// This would be "ON CONFLICT(id) [WHERE ...] DO UPDATE SET ..."
		// This is the `index_predicate`, specified here:
		// https://www.postgresql.org/docs/current/sql-insert.html
		// IDK why our syntax supports this, there is literally not a way
		// somebody could make use of this within Kwil right now.
		panic("engine does not yet support index predicates on upsert. Try using a WHERE constraint after the SET clause.")
	}
	if arbiterIndex == nil {
		return nil, fmt.Errorf("conflict column must be specified for DO UPDATE")
	}

	res := &ConflictUpdate{
		ArbiterIndex: arbiterIndex,
	}

	for _, set := range node.DoUpdate {
		res.Assignments = append(res.Assignments, &Assignment{
			Column: set.Column,
			Value:  set.Value.Accept(p).(LogicalExpr),
		})
	}

	if node.UpdateWhere != nil {
		res.ConflictFilter = node.UpdateWhere.Accept(p).(LogicalExpr)
	}

	return res, nil
}

// checkNullableColumns takes a table and a set of columns, and checks if
// any column in the table not in the set is nullable. If so, it returns
// an error. It also checks if all columns in the set are in the table.
// If not, it returns an error. If all checks pass, it returns a slice
// of data types that the insert order must match.
func checkNullableColumns(tbl *types.Table, cols []string) ([]*types.DataType, error) {
	specifiedColSet := make(map[string]struct{}, len(cols))
	for _, col := range cols {
		specifiedColSet[col] = struct{}{}
	}

	pks, err := tbl.GetPrimaryKey()
	if err != nil {
		return nil, err
	}
	pkSet := make(map[string]struct{}, len(pks))
	for _, pk := range pks {
		pkSet[pk] = struct{}{}
	}

	// we will build a set of columns to decrease the time complexity
	// for checking if a column is in the set.
	tblColSet := make(map[string]*types.DataType, len(tbl.Columns))
	for _, col := range tbl.Columns {
		tblColSet[col.Name] = col.Type.Copy()

		_, ok := specifiedColSet[col.Name]
		if ok {
			continue
		}

		// the column is not in the set, so we need to check if it is nullable
		if col.HasAttribute(types.NOT_NULL) || col.HasAttribute(types.PRIMARY_KEY) {
			return nil, fmt.Errorf(`column "%s" cannot be null, and was not specified as an insert column`, col.Name)
		}

		// it is also possible that a primary index contains the column
		_, ok = pkSet[col.Name]
		if !ok {
			return nil, fmt.Errorf(`column "%s" cannot be null, and was not specified as an insert column`, col.Name)
		}
		// otherwise, we are good
	}

	dataTypeArr := make([]*types.DataType, len(cols))
	// now we need to check if all columns in the set are in the table
	for _, col := range cols {
		colType, ok := tblColSet[col]
		if !ok {
			return nil, fmt.Errorf(`column "%s" not found in table`, col)
		}

		dataTypeArr = append(dataTypeArr, colType)
	}

	return dataTypeArr, nil
}

func (p *plannerVisitor) VisitUpsertClause(node *parse.UpsertClause) any {
	panic("internal bug: do not use this directly. Use the buildUpsert method.")
}

/*
	to make sure that we do not have any unimplemented visitor methods (which would cause unexpected bugs),
	below we include them, and panic if they are called.
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

func (p *plannerVisitor) VisitUpdateSetClause(node *parse.UpdateSetClause) any {
	panic("internal bug: VisitUpdateSetClause should not be called directly while building relational algebra")
}

const (
	alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base     = len(alphabet)
)

// numberToLetters allows us to assign a letter-based identifier for
// expression references. We do this to avoid confusion with subplan
// references, which are numbers.
func numberToLetters(n int) string {
	if n == 0 {
		return string(alphabet[0])
	}

	var sb strings.Builder
	for n > 0 {
		remainder := n % base
		sb.WriteByte(alphabet[remainder])
		n = n / base
	}

	// Reverse the string because the current order is backward
	result := sb.String()
	runes := []rune(result)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}
