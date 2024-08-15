package planner2

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils/order"
	"github.com/kwilteam/kwil-db/parse"
)

// Plan creates a logical plan from a SQL statement.
func Plan(statement *parse.SQLStatement, schema *types.Schema, vars map[string]*types.DataType,
	objects map[string]map[string]*types.DataType) (analyzed *AnalyzedPlan, err error) {
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

	scope := &scopeContext{
		plan:          ctx,
		OuterRelation: &Relation{},
	}

	plan, err := scope.sqlStmt(statement)
	if err != nil {
		return nil, err
	}

	return &AnalyzedPlan{
		Plan: plan,
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

// scopeContext contains information about the current scope of the query.
type scopeContext struct {
	// plan is the larger plan context that applies to the entire query.
	plan *planContext
	// OuterRelation is the relation of all outer queries that can be
	// referenced from a subquery.
	OuterRelation *Relation
	// Correlations are the fields that are corellated to an outer query.
	Correlations []*Field
}

// sqlStmt builds a logical plan for a top-level SQL statement.
func (s *scopeContext) sqlStmt(node *parse.SQLStatement) (TopLevelPlan, error) {
	for _, cte := range node.CTEs {
		if err := s.cte(cte); err != nil {
			return nil, err
		}
	}

	switch node := node.SQL.(type) {
	default:
		panic(fmt.Sprintf("unexpected SQL statement type %T", node))
	case *parse.SelectStatement:
		plan, res, err := s.selectStmt(node)
		if err != nil {
			return nil, err
		}

		for _, field := range res.Fields {
			if field.Name == "" {
				field.Name = "?column?"
			}
		}

		return &Return{
			Child:  plan,
			Fields: res.Fields,
		}, nil
	case *parse.UpdateStatement:
		return s.update(node)
	case *parse.DeleteStatement:
		return s.delete(node)
	case *parse.InsertStatement:
		return s.insert(node)
	}
}

// cte builds a common table expression.
func (s *scopeContext) cte(node *parse.CommonTableExpression) error {
	plan, rel, err := s.selectStmt(node.Query)
	if err != nil {
		return err
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

	s.plan.CTEs[node.Name] = rel
	s.plan.CTEPlans = append(s.plan.CTEPlans, &Subplan{
		Plan:      plan,
		ID:        node.Name,
		Type:      SubplanTypeCTE,
		extraInfo: extraInfo,
	})

	return nil
}

// select builds a logical plan for a select statement.
func (s *scopeContext) selectStmt(node *parse.SelectStatement) (plan LogicalPlan, rel *Relation, err error) {
	if len(node.SelectCores) == 0 {
		panic("no select cores")
	}

	var projectFunc func(LogicalPlan) (LogicalPlan, *Relation)
	plan, rel, projectFunc, err = s.selectCore(node.SelectCores[0])
	if err != nil {
		return nil, nil, err
	}

	// if there is one select core, we want to project after sorting and limiting.
	// if there are multiple select cores, we want to project before the set operation.
	// see the documentation for selectCoreResult for more info as to why
	// we perform this if statement.
	if len(node.SelectCores) == 1 {
		defer func() {
			plan, rel = projectFunc(plan)
		}()
	} else {
		// otherwise, apply immediately so that we can apply the set operation(s)
		plan, rel = projectFunc(plan)

		for i, core := range node.SelectCores[1:] {
			rightPlan, _, projectFunc, err := s.selectCore(core)
			if err != nil {
				return nil, nil, err
			}

			rightPlan, _ = projectFunc(rightPlan)
			plan = &SetOperation{
				Left:   plan,
				Right:  rightPlan,
				OpType: get(compoundTypes, node.CompoundOperators[i]),
			}
		}
	}

	// apply order by, limit, and offset
	if len(node.Ordering) > 0 {
		sort := &Sort{
			Child: plan,
		}

		for _, order := range node.Ordering {
			sortExpr, _, err := s.expr(order.Expression, rel)
			if err != nil {
				return nil, nil, err
			}

			sort.SortExpressions = append(sort.SortExpressions, &SortExpression{
				Expr:      sortExpr,
				Ascending: get(orderAsc, order.Order),
				NullsLast: get(orderNullsLast, order.Nulls),
			})
		}

		plan = sort
	}

	if node.Limit != nil {
		limitExpr, _, err := s.expr(node.Limit, rel)
		if err != nil {
			return nil, nil, err
		}

		lim := &Limit{
			Child: plan,
			Limit: limitExpr,
		}

		if node.Offset != nil {
			offsetExpr, _, err := s.expr(node.Offset, rel)
			if err != nil {
				return nil, nil, err
			}

			lim.Offset = offsetExpr
		}

		plan = lim
	}

	return plan, rel, nil
}

// selectCore builds a logical plan for a select core.
// The order of building is:
// 1. from (combining any joins into single source plan)
// 2. where
// 3. group by(can use reference from select)
// 4. having(can use reference from select)
// 5. select (project)
// 6. distinct
// It returns a logical plan and relation that are PRIOR to any projection,
// a function that will apply a projection and return the resulting relation,
// and an error if one occurred.
// It returns these because we need to handle conditionally
// adding projection. If a query has a SET (a.k.a. compound) operation, we want to project before performing
// the set. If a query has one select, then we want to project after sorting and limiting.
// To give a concrete example of this, imagine a table users (id int, name text) with the queries:
// 1.
// "SELECT name FROM users ORDER BY id" - this is valid in Postgres, and since we can access "id", projection
// should be done after sorting.
// 2.
// "SELECT name FROM users UNION 'hello' ORDER BY id" - this is invalid in Postgres, since "id" is not in the
// result set. We need to project before the UNION.
func (s *scopeContext) selectCore(node *parse.SelectCore) (LogicalPlan, *Relation, func(LogicalPlan) (LogicalPlan, *Relation), error) {
	// if there is no from, we just project the columns and return
	if node.From == nil {
		var exprs []LogicalExpr
		rel := &Relation{}
		for _, resultCol := range node.Columns {
			switch resultCol := resultCol.(type) {
			default:
				panic(fmt.Sprintf("unexpected result column type %T", resultCol))
			case *parse.ResultColumnExpression:
				expr, field, err := s.expr(resultCol.Expression, rel)
				if err != nil {
					return nil, nil, nil, err
				}

				if resultCol.Alias != "" {
					expr = &AliasExpr{
						Expr:  expr,
						Alias: resultCol.Alias,
					}
					field.Parent = ""
					field.Name = resultCol.Alias
				}

				exprs = append(exprs, expr)
				rel.Fields = append(rel.Fields, field)
			case *parse.ResultColumnWildcard:
				// if there is no from, we cannot expand the wildcard
				panic(`wildcard "*" cannot be used without a FROM clause`)
			}
		}

		return &EmptyScan{}, rel, func(lp LogicalPlan) (LogicalPlan, *Relation) {
			var p LogicalPlan = &Project{
				Child:       lp,
				Expressions: exprs,
			}

			if node.Distinct {
				p = &Distinct{
					Child: p,
				}
			}

			return p, rel
		}, nil
	}

	// otherwise, we need to build the from and join clauses
	scan, rel, err := s.table(node.From)
	if err != nil {
		return nil, nil, nil, err
	}
	var plan LogicalPlan = scan

	for _, join := range node.Joins {
		plan, rel, err = s.join(plan, rel, join)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	if node.Where != nil {
		whereExpr, _, err := s.expr(node.Where, rel)
		if err != nil {
			return nil, nil, nil, err
		}

		plan = &Filter{
			Child:     plan,
			Condition: whereExpr,
		}
	}

	// at this point, we have the full relation for the select core, and can expand the columns
	results, err := s.expandResultCols(rel, node.Columns)
	if err != nil {
		return nil, nil, nil, err
	}

	containsAgg := false
	for _, result := range results {
		containsAgg = hasAggregate(result.Expr)
	}

	// if there is no group by or aggregate, we can apply any distinct and return
	if len(node.GroupBy) == 0 && !containsAgg {
		return plan, rel, func(lp LogicalPlan) (LogicalPlan, *Relation) {
			var resExprs []LogicalExpr
			var resFields []*Field
			for _, result := range results {
				resExprs = append(resExprs, result.Expr)
				resFields = append(resFields, result.Field)
			}

			var p LogicalPlan = &Project{
				Child:       lp,
				Expressions: resExprs,
			}

			if node.Distinct {
				p = &Distinct{
					Child: p,
				}
			}

			return p, &Relation{
				Fields: resFields,
			}
		}, nil
	}

	// otherwise, we need to build the group by and having clauses.
	// This means that for all result columns, we need to rewrite any
	// column references or aggregate usage as columnrefs to the aggregate
	// functions matching term.
	aggTerms := make(map[string]*exprFieldPair[*IdentifiedExpr]) // any aggregate function used in the result or having
	groupingTerms := make(map[string]*IdentifiedExpr)            // any grouping term used in the GROUP BY
	aggregateRel := &Relation{}                                  // the relation resulting from the aggregation

	aggPlan := &Aggregate{ // defined separately so we can reference it in the below clauses
		Child: plan,
	}
	plan = aggPlan

	for _, groupTerm := range node.GroupBy {
		groupExpr, field, err := s.expr(groupTerm, rel)
		if err != nil {
			return nil, nil, nil, err
		}

		// TODO: disallow subqueries in group by
		if hasAggregate(groupExpr) {
			return nil, nil, nil, fmt.Errorf(`aggregate functions are not allowed in GROUP BY`)
		}

		_, ok := groupingTerms[groupExpr.String()]
		if ok {
			continue
		}

		// we should identify it so it can be referenced
		identified := &IdentifiedExpr{
			Expr: groupExpr,
			ID:   s.plan.uniqueRefIdentifier(),
		}
		aggPlan.GroupingExpressions = append(aggPlan.GroupingExpressions, identified)

		field.ReferenceID = identified.ID
		aggregateRel.Fields = append(aggregateRel.Fields, field)

		groupingTerms[groupExpr.String()] = identified
	}

	if node.Having != nil {
		// hmmmmm this doesnt work because the having rel needs to be the aggregation rel,
		// but we need to use this to build the aggregation rel :(
		// 2: on second thought, maybe not. We will have to do some tree matching and rewriting,
		// but it should be possible.
		havingExpr, _, err := s.expr(node.Having, rel)
		if err != nil {
			return nil, nil, nil, err
		}

		// rewrite the having expression to use the aggregate functions
		havingExpr, err = s.rewriteAccordingToAggregate(havingExpr, groupingTerms, aggTerms)
		if err != nil {
			return nil, nil, nil, err
		}

		plan = &Filter{
			Child:     plan,
			Condition: havingExpr,
		}
	}

	// now we need to rewrite the select list to use the aggregate functions
	for i, resultCol := range results {
		results[i].Expr, err = s.rewriteAccordingToAggregate(resultCol.Expr, groupingTerms, aggTerms)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	// finally, all of the aggregated columns need to be added to the Aggregate node
	for _, agg := range order.OrderMap(aggTerms) {
		aggPlan.AggregateExpressions = append(aggPlan.AggregateExpressions, agg.Value.Expr)
		aggregateRel.Fields = append(aggregateRel.Fields, agg.Value.Field)
	}

	return plan, aggregateRel, func(lp LogicalPlan) (LogicalPlan, *Relation) {
		var resultColExprs []LogicalExpr
		var resultFields []*Field
		for _, resultCol := range results {
			resultColExprs = append(resultColExprs, resultCol.Expr)
			resultFields = append(resultFields, resultCol.Field)
		}

		var p LogicalPlan = &Project{
			Child:       lp,
			Expressions: resultColExprs,
		}

		if node.Distinct {
			p = &Distinct{
				Child: p,
			}
		}

		return p, &Relation{
			Fields: resultFields,
		}
	}, nil
}

// hasAggregate returns true if the expression contains an aggregate function.
func hasAggregate(expr LogicalNode) bool {
	var hasAggregate bool
	traverse(expr, func(node Traversable) bool {
		if _, ok := node.(*AggregateFunctionCall); ok {
			hasAggregate = true
			return false
		}

		return true
	})

	return hasAggregate
}

// exprFieldPair is a helper struct that pairs an expression with a field.
// It uses a generic because there are some times where we want to guarantee
// that the expression is an IdentifiedExpr, and other times where we don't
// care about the concrete type.
type exprFieldPair[T LogicalExpr] struct {
	Expr  T
	Field *Field
}

// rewriteAccordingToAggregate rewrites an expression according to the rules of aggregation.
// This is used to rewrite both the select list and having clause to validate that all columns
// are either captured in aggregates or have an exactly matching expression in the group by.
func (s *scopeContext) rewriteAccordingToAggregate(expr LogicalExpr, groupingTerms map[string]*IdentifiedExpr, aggTerms map[string]*exprFieldPair[*IdentifiedExpr]) (LogicalExpr, error) {
	node, err := Rewrite(expr, &RewriteConfig{
		ExprCallback: func(le LogicalExpr) (LogicalExpr, bool, error) {
			// if it matches any group by term, we need to rewrite it
			// and stop traversing any children
			identified, ok := groupingTerms[le.String()]
			if ok {
				return &ExprRef{
					Identified: identified,
				}, false, nil
			}

			switch le := le.(type) {
			case *ColumnRef:
				// if it is a column and in the current relation, it is an error, since
				// it was not contained in an aggregate function or group by.
				return nil, false, fmt.Errorf(`%w: column "%s" must appear in the GROUP BY clause or be used in an aggregate function`, ErrIllegalAggregate, le.String())
			case *AggregateFunctionCall:
				// TODO: do we need to check for the aggregate being called on a correlated column?
				// if it matches any aggregate function, we need to rewrite it
				// to that reference. Otherwise, register it as a new aggregate
				identified, ok := aggTerms[le.String()]
				if ok {
					return &ExprRef{
						Identified: identified.Expr,
					}, false, nil
				}

				newIdentified := &IdentifiedExpr{
					Expr: le,
					ID:   s.plan.uniqueRefIdentifier(),
				}

				aggTerms[le.String()] = &exprFieldPair[*IdentifiedExpr]{
					Expr: newIdentified,
					Field: &Field{
						Name:        le.FunctionName,
						val:         le.returnType.Copy(),
						ReferenceID: newIdentified.ID,
					},
				}

				return &ExprRef{
					Identified: newIdentified,
				}, false, nil
			default:
				return le, true, nil
			}
		},
	})
	if err != nil {
		return nil, err
	}

	return node.(LogicalExpr), nil
}

// expandResultCols takes a relation and result columns, and converts them to expressions
// in the order provided. This is used to expand a wildcard in a select statement.
func (s *scopeContext) expandResultCols(rel *Relation, cols []parse.ResultColumn) ([]*exprFieldPair[LogicalExpr], error) {
	var resultCols []LogicalExpr
	var resultFields []*Field
	for _, col := range cols {
		switch col := col.(type) {
		default:
			panic(fmt.Sprintf("unexpected result column type %T", col))
		case *parse.ResultColumnExpression:
			expr, field, err := s.expr(col.Expression, rel)
			if err != nil {
				return nil, err
			}

			if col.Alias != "" {
				expr = &AliasExpr{
					Expr:  expr,
					Alias: col.Alias,
				}
				// since it is aliased, we now ignore the parent
				field.Parent = ""
				field.Name = col.Alias
			}

			resultFields = append(resultFields, field)
			resultCols = append(resultCols, expr)
		case *parse.ResultColumnWildcard:
			var newFields []*Field
			if col.Table != "" {
				newFields = rel.ColumnsByParent(col.Table)
			} else {
				newFields = rel.Fields
			}

			for _, field := range newFields {
				resultCols = append(resultCols, &ColumnRef{
					Parent:     field.Parent,
					ColumnName: field.Name,
				})
				resultFields = append(resultFields, field)
			}
		}
	}

	var pairs []*exprFieldPair[LogicalExpr]
	for i, expr := range resultCols {
		pairs = append(pairs, &exprFieldPair[LogicalExpr]{
			Expr:  expr,
			Field: resultFields[i],
		})
	}

	return pairs, nil
}

// expr visits an expression node.
func (s *scopeContext) expr(node parse.Expression, currentRel *Relation) (LogicalExpr, *Field, error) {
	// cast is a helper function for type casting results based on the current node
	cast := func(expr LogicalExpr, field *Field) (LogicalExpr, *Field, error) {
		castable, ok := node.(interface{ GetTypeCast() *types.DataType })
		if !ok {
			return expr, field, nil
		}

		if castable.GetTypeCast() != nil {
			field2 := field.Copy()
			field2.val = castable.GetTypeCast()

			return &TypeCast{
				Expr: expr,
				Type: castable.GetTypeCast(),
			}, field2, nil
		}

		return expr, field, nil
	}

	switch node := node.(type) {
	default:
		panic(fmt.Sprintf("unexpected expression type %T", node))
	case *parse.ExpressionLiteral:
		return cast(&Literal{
			Value: node.Value,
			Type:  node.Type,
		}, anonField(node.Type))
	case *parse.ExpressionFunctionCall:
		args, fields, err := s.manyExprs(node.Args, currentRel)
		if err != nil {
			return nil, nil, err
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
			proc, found := s.plan.Schema.FindProcedure(node.Name)
			if !found {
				panic(fmt.Sprintf(`no function or procedure "%s" found`, node.Name))
			}

			returns, err := procedureReturnExpr(proc.Returns)
			if err != nil {
				return nil, nil, err
			}

			return cast(&ProcedureCall{
				ProcedureName: node.Name,
				Args:          args,
				returnType:    returns,
			}, &Field{
				Name: node.Name,
				val:  returns,
			})
		}

		// it is a built-in function

		types, err := dataTypes(fields)
		if err != nil {
			return nil, nil, err
		}

		returnVal, err := funcDef.ValidateArgs(types)
		if err != nil {
			return nil, nil, err
		}

		returnField := &Field{
			Name: node.Name,
			val:  returnVal,
		}

		// now we need to apply rules depending on if it is aggregate or not
		if funcDef.IsAggregate {
			// we apply cast outside the reference because we want to keep the reference
			// specific to the aggregate function call.
			return cast(&AggregateFunctionCall{
				FunctionName: node.Name,
				Args:         args,
				Star:         node.Star,
				Distinct:     node.Distinct,
				returnType:   returnVal,
			}, returnField)
		}

		if node.Star {
			panic("star (*) not allowed in non-aggregate function calls")
		}
		if node.Distinct {
			panic("DISTINCT not allowed in non-aggregate function calls")
		}

		return cast(&ScalarFunctionCall{
			FunctionName: node.Name,
			Args:         args,
			returnType:   returnVal,
		}, returnField)
	case *parse.ExpressionForeignCall:
		proc, found := s.plan.Schema.FindForeignProcedure(node.Name)
		if !found {
			panic(fmt.Sprintf(`unknown foreign procedure "%s"`, node.Name))
		}

		if len(node.ContextualArgs) != 2 {
			panic("foreign calls must have 2 contextual arguments")
		}

		returns, err := procedureReturnExpr(proc.Returns)
		if err != nil {
			return nil, nil, err
		}

		args, _, err := s.manyExprs(node.Args, currentRel)
		if err != nil {
			return nil, nil, err
		}

		contextArgs, _, err := s.manyExprs(node.ContextualArgs, currentRel)
		if err != nil {
			return nil, nil, err
		}

		return cast(&ProcedureCall{
			ProcedureName: node.Name,
			Foreign:       true,
			Args:          args,
			ContextArgs:   contextArgs,
		}, &Field{
			Name: node.Name,
			val:  returns,
		})
	case *parse.ExpressionVariable:
		var val any // can be a data type or object
		dt, ok := s.plan.Variables[node.String()]
		if !ok {
			// might be an object
			obj, ok := s.plan.Objects[node.String()]
			if !ok {
				return nil, nil, fmt.Errorf(`unknown variable "%s"`, node.String())
			}

			val = obj
		} else {
			val = dt
		}

		return cast(&Variable{
			VarName: node.String(),
		}, &Field{val: val})
	case *parse.ExpressionArrayAccess:
		array, field, err := s.expr(node.Array, currentRel)
		if err != nil {
			return nil, nil, err
		}

		index, _, err := s.expr(node.Index, currentRel)
		if err != nil {
			return nil, nil, err
		}

		field2 := field.Copy()
		scalar, err := field2.Scalar()
		if err != nil {
			return nil, nil, err
		}
		scalar.IsArray = false // since we are accessing an array, it is no longer an array

		return cast(&ArrayAccess{
			Array: array,
			Index: index,
		}, field2)
	case *parse.ExpressionMakeArray:
		if len(node.Values) == 0 {
			return nil, nil, fmt.Errorf("array constructor must have at least one element")
		}

		exprs, fields, err := s.manyExprs(node.Values, currentRel)
		if err != nil {
			return nil, nil, err
		}

		firstVal, err := fields[0].Scalar()
		if err != nil {
			return nil, nil, err
		}

		for _, field := range fields[1:] {
			scalar, err := field.Scalar()
			if err != nil {
				return nil, nil, err
			}

			if !firstVal.Equals(scalar) {
				return nil, nil, fmt.Errorf("array constructor must have elements of the same type")
			}
		}

		firstValCopy := firstVal.Copy()
		firstValCopy.IsArray = true

		return cast(&ArrayConstructor{
			Elements: exprs,
		}, &Field{
			val: firstValCopy,
		})
	case *parse.ExpressionFieldAccess:
		obj, field, err := s.expr(node.Record, currentRel)
		if err != nil {
			return nil, nil, err
		}

		objType, err := field.Object()
		if err != nil {
			return nil, nil, err
		}

		fieldType, ok := objType[node.Field]
		if !ok {
			return nil, nil, fmt.Errorf(`object "%s" does not have field "%s"`, field.Name, node.Field)
		}

		return cast(&FieldAccess{
			Object: obj,
			Key:    node.Field,
		}, &Field{
			val: fieldType,
		})
	case *parse.ExpressionParenthesized:
		expr, field, err := s.expr(node.Inner, currentRel)
		if err != nil {
			return nil, nil, err
		}

		return cast(expr, field)
	case *parse.ExpressionComparison:
		left, _, err := s.expr(node.Left, currentRel)
		if err != nil {
			return nil, nil, err
		}

		right, _, err := s.expr(node.Right, currentRel)
		if err != nil {
			return nil, nil, err
		}

		return &ComparisonOp{
			Left:  left,
			Right: right,
			Op:    get(comparisonOps, node.Operator),
		}, anonField(types.BoolType.Copy()), nil
	case *parse.ExpressionLogical:
		left, _, err := s.expr(node.Left, currentRel)
		if err != nil {
			return nil, nil, err
		}

		right, _, err := s.expr(node.Right, currentRel)
		if err != nil {
			return nil, nil, err
		}

		return &LogicalOp{
			Left:  left,
			Right: right,
			Op:    get(logicalOps, node.Operator),
		}, anonField(types.BoolType.Copy()), nil
	case *parse.ExpressionArithmetic:
		left, leftField, err := s.expr(node.Left, currentRel)
		if err != nil {
			return nil, nil, err
		}

		right, _, err := s.expr(node.Right, currentRel)
		if err != nil {
			return nil, nil, err
		}

		return &ArithmeticOp{
			Left:  left,
			Right: right,
			Op:    get(arithmeticOps, node.Operator),
		}, &Field{val: leftField.val}, nil
	case *parse.ExpressionUnary:
		expr, field, err := s.expr(node.Expression, currentRel)
		if err != nil {
			return nil, nil, err
		}

		// surprisingly, Postgres won't return a columns name
		// if it is wrapped in a unary operator
		return &UnaryOp{
			Expr: expr,
			Op:   get(unaryOps, node.Operator),
		}, &Field{val: field.val}, nil
	case *parse.ExpressionColumn:
		field, err := currentRel.Search(node.Table, node.Column)
		if errors.Is(err, ErrColumnNotFound) {
			// might be in the outer relation, correlated
			field, err = s.OuterRelation.Search(node.Table, node.Column)
			if err != nil {
				return nil, nil, err
			}

			// add to correlations
			s.Correlations = append(s.Correlations, field)

			return cast(&ColumnRef{
				Parent:     field.Parent,
				ColumnName: field.Name,
			}, field)
		}
		if err != nil {
			return nil, nil, err
		}

		return cast(&ColumnRef{
			Parent:     field.Parent,
			ColumnName: field.Name,
		}, field)
	case *parse.ExpressionCollate:
		expr, field, err := s.expr(node.Expression, currentRel)
		if err != nil {
			return nil, nil, err
		}

		c := &Collate{
			Expr: expr,
		}

		switch strings.ToLower(node.Collation) {
		case "nocase":
			c.Collation = NoCaseCollation
		default:
			return nil, nil, fmt.Errorf(`unknown collation "%s"`, node.Collation)
		}

		// return the whole field since collations don't overwrite the return value's name
		return c, field, nil
	case *parse.ExpressionStringComparison:
		left, _, err := s.expr(node.Left, currentRel)
		if err != nil {
			return nil, nil, err
		}

		right, _, err := s.expr(node.Right, currentRel)
		if err != nil {
			return nil, nil, err
		}

		var expr LogicalExpr = &ComparisonOp{
			Left:  left,
			Right: right,
			Op:    get(stringComparisonOps, node.Operator),
		}

		if node.Not {
			expr = &UnaryOp{
				Expr: expr,
				Op:   Not,
			}
		}

		return expr, anonField(types.BoolType.Copy()), nil
	case *parse.ExpressionIs:
		op := Is
		if node.Distinct {
			op = IsDistinctFrom
		}

		left, _, err := s.expr(node.Left, currentRel)
		if err != nil {
			return nil, nil, err
		}

		right, _, err := s.expr(node.Right, currentRel)
		if err != nil {
			return nil, nil, err
		}

		var expr LogicalExpr = &ComparisonOp{
			Left:  left,
			Right: right,
			Op:    op,
		}

		if node.Not {
			expr = &UnaryOp{
				Expr: expr,
				Op:   Not,
			}
		}

		return expr, anonField(types.BoolType.Copy()), nil
	case *parse.ExpressionIn:
		left, _, err := s.expr(node.Expression, currentRel)
		if err != nil {
			return nil, nil, err
		}

		in := &IsIn{
			Left: left,
		}

		if node.Subquery != nil {
			subq, rel, err := s.planSubquery(node.Subquery, currentRel)
			if err != nil {
				return nil, nil, err
			}

			if len(rel.Fields) != 1 {
				return nil, nil, fmt.Errorf("subquery must return exactly one column")
			}

			in.Subquery = &SubqueryExpr{
				Query: subq,
			}
		} else {
			right, _, err := s.manyExprs(node.List, currentRel)
			if err != nil {
				return nil, nil, err
			}

			in.Expressions = right
		}

		var expr LogicalExpr = in

		if node.Not {
			expr = &UnaryOp{
				Expr: expr,
				Op:   Not,
			}
		}

		return expr, anonField(types.BoolType.Copy()), nil
	case *parse.ExpressionBetween:
		leftOp, rightOp := GreaterThanOrEqual, LessThanOrEqual
		if node.Not {
			leftOp, rightOp = LessThan, GreaterThan
		}

		left, _, err := s.expr(node.Expression, currentRel)
		if err != nil {
			return nil, nil, err
		}

		lower, _, err := s.expr(node.Lower, currentRel)
		if err != nil {
			return nil, nil, err
		}

		upper, _, err := s.expr(node.Upper, currentRel)
		if err != nil {
			return nil, nil, err
		}

		return &LogicalOp{
			Left: &ComparisonOp{
				Left:  left,
				Right: lower,
				Op:    leftOp,
			},
			Right: &ComparisonOp{
				Left:  left,
				Right: upper,
				Op:    rightOp,
			},
			Op: And,
		}, anonField(types.BoolType.Copy()), nil
	case *parse.ExpressionCase:
		c := &Case{}

		// all whens must be bool unless an expression is used before CASE
		expectedWhenType := types.BoolType.Copy()
		if node.Case != nil {
			caseExpr, field, err := s.expr(node.Case, currentRel)
			if err != nil {
				return nil, nil, err
			}

			c.Value = caseExpr
			expectedWhenType, err = field.Scalar()
			if err != nil {
				return nil, nil, err
			}
		}

		var returnType *types.DataType
		for _, whenThen := range node.WhenThen {
			whenExpr, whenField, err := s.expr(whenThen[0], currentRel)
			if err != nil {
				return nil, nil, err
			}

			thenExpr, thenField, err := s.expr(whenThen[1], currentRel)
			if err != nil {
				return nil, nil, err
			}

			thenType, err := thenField.Scalar()
			if err != nil {
				return nil, nil, err
			}
			if returnType == nil {
				returnType = thenType
			} else {
				if !returnType.Equals(thenType) {
					return nil, nil, fmt.Errorf(`all THEN expressions must be of the same type %s, received %s`, returnType, thenType)
				}
			}

			whenScalar, err := whenField.Scalar()
			if err != nil {
				return nil, nil, err
			}

			if !expectedWhenType.Equals(whenScalar) {
				return nil, nil, fmt.Errorf(`WHEN expression must be of type %s, received %s`, expectedWhenType, whenScalar)
			}

			c.WhenClauses = append(c.WhenClauses, [2]LogicalExpr{whenExpr, thenExpr})
		}

		if node.Else != nil {
			elseExpr, elseField, err := s.expr(node.Else, currentRel)
			if err != nil {
				return nil, nil, err
			}

			elseType, err := elseField.Scalar()
			if err != nil {
				return nil, nil, err
			}

			if !returnType.Equals(elseType) {
				return nil, nil, fmt.Errorf(`ELSE expression must be of the same type of THEN expressions %s, received %s`, returnType, elseExpr)
			}

			c.Else = elseExpr
		}

		return c, anonField(returnType), nil
	case *parse.ExpressionSubquery:
		subq, rel, err := s.planSubquery(node.Subquery, currentRel)
		if err != nil {
			return nil, nil, err
		}

		subqExpr := &SubqueryExpr{
			Query: subq,
		}
		if node.Exists {
			subqExpr.Exists = true

			var plan LogicalExpr = subqExpr
			if node.Not {
				plan = &UnaryOp{
					Expr: plan,
					Op:   Not,
				}
			}

			return plan, anonField(types.BoolType.Copy()), nil
		} else {
			if len(rel.Fields) != 1 {
				return nil, nil, fmt.Errorf("scalar subquery must return exactly one column")
			}
		}

		return subqExpr, rel.Fields[0], nil
	}
}

// planSubquery plans a subquery.
// It takes the relation of the calling query to allow for correlated subqueries.
func (s *scopeContext) planSubquery(node *parse.SelectStatement, currentRel *Relation) (*Subquery, *Relation, error) {
	// for a subquery, we will add the current relation to the outer relation,
	// to allow for correlated subqueries
	oldOuter := s.OuterRelation
	oldCorrelations := s.Correlations

	s.OuterRelation = &Relation{
		Fields: append(s.OuterRelation.Fields, currentRel.Fields...),
	}
	// we don't need access to the old correlations since we will simply
	// recognize them as correlated again if they are used in the subquery
	s.Correlations = []*Field{}

	defer func() {
		s.OuterRelation = oldOuter
		s.Correlations = oldCorrelations
	}()

	query, rel, err := s.selectStmt(node)
	if err != nil {
		return nil, nil, err
	}

	// for all new correlations, we need to check if they are present on
	// the oldOuter relation. If so, then we simply add them as correlated
	// to the subplan. If not, then we also need to pass them back to the
	// oldCorrelations so that they can be used in the outer query (in the case
	// of a multi-level correlated subquery)
	oldMap := make(map[[2]string]struct{})
	for _, cor := range oldCorrelations {
		oldMap[[2]string{cor.Parent, cor.Name}] = struct{}{}
	}
	for _, cor := range s.Correlations {
		_, err = currentRel.Search(cor.Parent, cor.Name)
		// if the column is not found in the current relation, then we need to
		// pass it back to the oldCorrelations
		if errors.Is(err, ErrColumnNotFound) {
			// if not known to the outer correlation, then add it
			_, ok := oldMap[[2]string{cor.Parent, cor.Name}]
			if !ok {
				oldCorrelations = append(oldCorrelations, cor)
				continue
			}
		} else if err != nil {
			// some other error occurred
			return nil, nil, err
		}
		// if no error, it is correlated to this query, do nothing
	}

	plan := &Subplan{
		Plan: query,
		ID:   strconv.Itoa(s.plan.SubqueryCount),
		Type: SubplanTypeSubquery,
	}

	s.plan.SubqueryCount++

	// the returned relation should not have any parent tables
	for _, col := range rel.Fields {
		col.Parent = ""
	}

	return &Subquery{
		Plan:       plan,
		Correlated: s.Correlations,
	}, rel, nil
}

// manyExprs is a helper function that applies the expr function to many expressions.
func (s *scopeContext) manyExprs(nodes []parse.Expression, currentRel *Relation) ([]LogicalExpr, []*Field, error) {
	var exprs []LogicalExpr
	var fields []*Field
	for _, node := range nodes {
		expr, field, err := s.expr(node, currentRel)
		if err != nil {
			return nil, nil, err
		}

		exprs = append(exprs, expr)
		fields = append(fields, field)
	}

	return exprs, fields, nil
}

// procedureReturnExpr gets the returned data type from a procedure return.
func procedureReturnExpr(node *types.ProcedureReturn) (*types.DataType, error) {
	if node == nil {
		return nil, fmt.Errorf("procedure does not return a value")
	}

	if node.IsTable {
		return nil, fmt.Errorf("procedure returns a table, not a scalar value")
	}

	if len(node.Fields) != 1 {
		return nil, fmt.Errorf("procedures in expressions must return exactly one value, received %d", len(node.Fields))
	}

	return node.Fields[0].Type.Copy(), nil
}

// anonField creates an anonymous field with the given data type.
func anonField(dt *types.DataType) *Field {
	return &Field{
		val: dt,
	}
}

// table takes a parse.Table interface and returns the plan and relation
// for the table.
func (s *scopeContext) table(node parse.Table) (*Scan, *Relation, error) {
	switch node := node.(type) {
	default:
		panic(fmt.Sprintf("unexpected parse table type %T", node))
	case *parse.RelationTable:
		// either a CTE or a physical table
		alias := node.Table
		if node.Alias != "" {
			alias = node.Alias
		}

		var scanTblType TableSourceType
		var rel *Relation
		if physicalTbl, ok := s.plan.Schema.FindTable(node.Table); ok {
			scanTblType = TableSourcePhysical
			rel = relationFromTable(physicalTbl)
		} else if cte, ok := s.plan.CTEs[node.Table]; ok {
			scanTblType = TableSourceCTE
			rel = cte
		} else {
			return nil, nil, fmt.Errorf(`unknown table "%s"`, node.Table)
		}

		for _, col := range rel.Fields {
			col.Parent = alias
		}

		return &Scan{
			Source: &TableScanSource{
				TableName: node.Table,
				Type:      scanTblType,
				rel:       rel.Copy(),
			},
			RelationName: alias,
		}, rel, nil
	case *parse.RelationSubquery:
		if node.Alias == "" {
			return nil, nil, fmt.Errorf("join against subquery must have an alias")
		}

		// we pass an empty relation because the subquery can't
		// refer to the current relation, but they can be correlated against some
		// outer relation.
		// for example, "select * from users u inner join (select * from posts where posts.id = u.id) as p on u.id=p.id;"
		// is invalid, but
		// "select * from users where id = (select posts.id from posts inner join (select * from posts where id = users.id) as s on s.id=posts.id);"
		// is valid
		subq, rel, err := s.planSubquery(node.Subquery, &Relation{})
		if err != nil {
			return nil, nil, err
		}

		for _, col := range rel.Fields {
			col.Parent = node.Alias
		}

		return &Scan{
			Source:       subq,
			RelationName: node.Alias,
		}, rel, nil
	case *parse.RelationFunctionCall:
		if node.Alias == "" {
			return nil, nil, fmt.Errorf("join against procedure calls must have an alias")
		}

		// the function call must either be a procedure or foreign procedure that returns
		// a table.

		var args []LogicalExpr
		var contextArgs []LogicalExpr
		var procReturns *types.ProcedureReturn
		var isForeign bool
		if proc, ok := s.plan.Schema.FindProcedure(node.FunctionCall.FunctionName()); ok {
			procReturns = proc.Returns

			procCall, ok := node.FunctionCall.(*parse.ExpressionFunctionCall)
			if !ok {
				// I don't think this is possible, but just in case
				return nil, nil, fmt.Errorf(`unexpected procedure type "%T"`, node.FunctionCall)
			}

			var fields []*Field
			var err error
			// we pass an empty relation because the subquery can't
			// refer to the current relation, but they can be correlated against some
			// outer relation.
			args, fields, err = s.manyExprs(procCall.Args, &Relation{})
			if err != nil {
				return nil, nil, err
			}

			if len(fields) != len(proc.Parameters) {
				return nil, nil, fmt.Errorf(`procedure "%s" expects %d arguments, received %d`, node.FunctionCall.FunctionName(), len(proc.Parameters), len(fields))
			}

			for i, field := range fields {
				scalar, err := field.Scalar()
				if err != nil {
					return nil, nil, err
				}

				if !scalar.Equals(proc.Parameters[i].Type) {
					return nil, nil, fmt.Errorf(`procedure "%s" expects argument %d to be of type %s, received %s`, node.FunctionCall.FunctionName(), i+1, proc.Parameters[i].Type, field)
				}
			}

		} else if proc, ok := s.plan.Schema.FindForeignProcedure(node.FunctionCall.FunctionName()); ok {
			procReturns = proc.Returns
			isForeign = true

			procCall, ok := node.FunctionCall.(*parse.ExpressionForeignCall)
			if !ok {
				// this is possible if the user doesn't pass contextual arguments,
				// (the parser will parse it as a regular function call instead of a foreign call)
				return nil, nil, fmt.Errorf(`procedure "%s" is a foreign procedure and must have contextual arguments passed with []`, node.FunctionCall.FunctionName())
			}

			var fields []*Field
			var err error
			// we pass an empty relation because the subquery can't
			// refer to the current relation, but they can be correlated against some
			// outer relation.
			args, fields, err = s.manyExprs(procCall.Args, &Relation{})
			if err != nil {
				return nil, nil, err
			}

			if len(fields) != len(proc.Parameters) {
				return nil, nil, fmt.Errorf(`foreign procedure "%s" expects %d arguments, received %d`, node.FunctionCall.FunctionName(), len(proc.Parameters), len(fields))
			}

			for i, field := range fields {
				scalar, err := field.Scalar()
				if err != nil {
					return nil, nil, err
				}

				if !scalar.Equals(proc.Parameters[i]) {
					return nil, nil, fmt.Errorf(`foreign procedure "%s" expects argument %d to be of type %s, received %s`, node.FunctionCall.FunctionName(), i+1, proc.Parameters[i], field)
				}
			}

			// must have 2 contextual arguments
			if len(procCall.ContextualArgs) != 2 {
				return nil, nil, fmt.Errorf(`foreign procedure "%s" must have 2 contextual arguments`, node.FunctionCall.FunctionName())
			}

			contextArgs, fields, err = s.manyExprs(procCall.ContextualArgs, &Relation{})
			if err != nil {
				return nil, nil, err
			}

			if len(fields) != 2 {
				return nil, nil, fmt.Errorf(`foreign procedure "%s" expects 2 contextual arguments, received %d`, node.FunctionCall.FunctionName(), len(fields))
			}

			for i, field := range fields {
				scalar, err := field.Scalar()
				if err != nil {
					return nil, nil, err
				}

				if !scalar.Equals(types.TextType) {
					return nil, nil, fmt.Errorf(`foreign procedure "%s" expects contextual argument %d to be of type %s, received %s`, node.FunctionCall.FunctionName(), i+1, types.TextType, field)
				}
			}
		} else {
			return nil, nil, fmt.Errorf(`unknown procedure "%s"`, node.FunctionCall.FunctionName())
		}

		if procReturns == nil {
			return nil, nil, fmt.Errorf(`procedure "%s" does not return a table`, node.FunctionCall.FunctionName())
		}
		if !procReturns.IsTable {
			return nil, nil, fmt.Errorf(`procedure "%s" does not return a table`, node.FunctionCall.FunctionName())
		}

		rel := &Relation{}
		for _, field := range procReturns.Fields {
			rel.Fields = append(rel.Fields, &Field{
				Parent: node.Alias,
				Name:   field.Name,
				val:    field.Type.Copy(),
			})
		}

		return &Scan{
			Source: &ProcedureScanSource{
				ProcedureName:  node.FunctionCall.FunctionName(),
				Args:           args,
				ContextualArgs: contextArgs,
				IsForeign:      isForeign,
				rel:            rel.Copy(),
			},
			RelationName: node.Alias,
		}, rel, nil
	}
}

// join wraps the given plan in a join node.
func (s *scopeContext) join(child LogicalPlan, childRel *Relation, join *parse.Join) (LogicalPlan, *Relation, error) {
	tbl, tblRel, err := s.table(join.Relation)
	if err != nil {
		return nil, nil, err
	}

	newRel := joinRels(childRel, tblRel)

	onExpr, _, err := s.expr(join.On, newRel)
	if err != nil {
		return nil, nil, err
	}

	plan := &Join{
		Left:      child,
		Right:     tbl,
		Condition: onExpr,
		JoinType:  get(joinTypes, join.Type),
	}

	return plan, newRel, nil
}

// update builds a plan for an update
func (s *scopeContext) update(node *parse.UpdateStatement) (*Update, error) {
	plan, targetRel, cartesianRel, err := s.cartesian(node.Table, node.Alias, node.From, node.Joins, node.Where)
	if err != nil {
		return nil, err
	}

	assigns, err := s.assignments(node.SetClause, targetRel, cartesianRel)
	if err != nil {
		return nil, err
	}

	return &Update{
		Child:       plan,
		Assignments: assigns,
		Table:       node.Table,
	}, nil
}

// delete builds a plan for a delete
func (s *scopeContext) delete(node *parse.DeleteStatement) (*Delete, error) {
	plan, _, _, err := s.cartesian(node.Table, node.Alias, node.From, node.Joins, node.Where)
	if err != nil {
		return nil, err
	}

	return &Delete{
		Child: plan,
		Table: node.Table,
	}, nil
}

// insert builds a plan for an insert
func (s *scopeContext) insert(node *parse.InsertStatement) (*Insert, error) {
	ins := &Insert{
		Table:        node.Table,
		ReferencedAs: node.Alias,
	}

	tbl, found := s.plan.Schema.FindTable(node.Table)
	if !found {
		return nil, fmt.Errorf(`%w: "%s"`, ErrUnknownTable, node.Table)
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
	var expectedColTypes []*types.DataType
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
		expectedColTypes = make([]*types.DataType, len(tbl.Columns))
		for i, col := range tbl.Columns {
			expectedColTypes[i] = col.Type.Copy()
		}
	}

	rel := relationFromTable(tbl)

	// check the value types and lengths
	for _, vals := range node.Values {
		if len(vals) != expectedColLen {
			return nil, fmt.Errorf(`insert has %d columns but %d values were supplied`, expectedColLen, len(vals))
		}

		var row []LogicalExpr

		for i, val := range vals {
			expr, field, err := s.expr(val, rel)
			if err != nil {
				return nil, err
			}

			scalar, err := field.Scalar()
			if err != nil {
				return nil, err
			}

			if !scalar.Equals(expectedColTypes[i]) {
				return nil, fmt.Errorf(`insert value %d must be of type %s, received %s`, i+1, expectedColTypes[i], field.val)
			}

			row = append(row, expr)
		}

		ins.Values = append(ins.Values, orderAndFillNulls(row))
	}

	// finally, we need to check if there is an ON CONFLICT clause,
	// and if so, we need to process it.
	if node.Upsert != nil {
		conflict, err := s.buildUpsert(node.Upsert, tbl)
		if err != nil {
			panic(err)
		}

		ins.ConflictResolution = conflict
	}

	return ins, nil
}

func (s *scopeContext) buildUpsert(node *parse.UpsertClause, table *types.Table) (ConflictResolution, error) {
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
		return nil, errors.New("engine does not yet support index predicates on upsert. Try using a WHERE constraint after the SET clause")
	}
	if arbiterIndex == nil {
		return nil, fmt.Errorf("conflict column must be specified for DO UPDATE")
	}

	res := &ConflictUpdate{
		ArbiterIndex: arbiterIndex,
	}

	rel := relationFromTable(table)

	var err error
	res.Assignments, err = s.assignments(node.DoUpdate, rel, rel)
	if err != nil {
		return nil, err
	}

	if node.UpdateWhere != nil {
		conflictFilter, field, err := s.expr(node.UpdateWhere, rel)
		if err != nil {
			return nil, err
		}

		scalar, err := field.Scalar()
		if err != nil {
			return nil, err
		}

		if !scalar.Equals(types.BoolType) {
			return nil, fmt.Errorf("conflict filter must be of type bool, received %s", field)
		}

		res.ConflictFilter = conflictFilter
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

// cartesian builds a cartesian product for several relations. It is meant to be used
// explicitly for update and delete, where we start by planning a cartesian join between the
// target table and the FROM + JOIN tables, and later optimize the filter.
// It returns the plan for the join, the relation that is being targeted, the relation that is the cartesian join
// between the target and the FROM + JOIN tables, and an error if one occurred.
func (s *scopeContext) cartesian(targetTable, alias string, from parse.Table, joins []*parse.Join,
	filter parse.Expression) (plan LogicalPlan, targetRel *Relation, cartesianRel *Relation, err error) {

	tbl, ok := s.plan.Schema.FindTable(targetTable)
	if !ok {
		return nil, nil, nil, fmt.Errorf(`unknown table "%s"`, targetTable)
	}
	if alias == "" {
		alias = targetTable
	}

	targetRel = relationFromTable(tbl)
	// copy that can be overwritten
	rel := targetRel.Copy()

	// plan the target table
	var targetPlan LogicalPlan = &Scan{
		Source: &TableScanSource{
			TableName: targetTable,
			Type:      TableSourcePhysical,
		},
		RelationName: alias,
	}

	// if there is no FROM clause, we can simply apply the filter and return
	if from == nil {
		if filter != nil {
			expr, _, err := s.expr(filter, rel)
			if err != nil {
				return nil, nil, nil, err
			}

			return &Filter{
				Child:     targetPlan,
				Condition: expr,
			}, targetRel, rel, nil
		}

		return targetPlan, rel, nil, nil
	}

	// update and delete statements with a FROM require a WHERE clause,
	// otherwise it is impossible to optimize the query to not be a cartesian product
	if filter == nil {
		return nil, nil, nil, ErrUpdateOrDeleteWithoutWhere
	}

	// plan the FROM clause

	var sourceRel LogicalPlan
	var fromRel *Relation
	sourceRel, fromRel, err = s.table(from)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, join := range joins {
		sourceRel, fromRel, err = s.join(sourceRel, fromRel, join)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	// build a cartesian product and apply the filter
	targetPlan = &CartesianProduct{
		Left:  targetPlan,
		Right: sourceRel,
	}

	rel = joinRels(fromRel, rel)

	expr, _, err := s.expr(filter, rel)
	if err != nil {
		return nil, nil, nil, err
	}

	return &Filter{
		Child:     targetPlan,
		Condition: expr,
	}, targetRel, rel, nil
}

// assignments builds the assignments for update and conflict clauses.
// It takes a list of update set clauses, the target relation (where columns are being assigned),
// and a relation that can be referenced in the assigning expressions.
func (s *scopeContext) assignments(assignments []*parse.UpdateSetClause, targetRel *Relation, referenceRel *Relation) ([]*Assignment, error) {
	assigns := make([]*Assignment, len(assignments))
	for i, assign := range assignments {
		field, err := targetRel.Search("", assign.Column)
		if err != nil {
			return nil, err
		}

		expr, assignType, err := s.expr(assign.Value, referenceRel)
		if err != nil {
			return nil, err
		}

		scalarField, err := assignType.Scalar()
		if err != nil {
			return nil, err
		}

		scalarAssign, err := field.Scalar()
		if err != nil {
			return nil, err
		}

		if !scalarField.Equals(scalarAssign) {
			return nil, fmt.Errorf(`cannot assign type %s to column "%s", expected type %s`, scalarField, field.Name, scalarAssign)
		}

		assigns[i] = &Assignment{
			Column: field.Name,
			Value:  expr,
		}
	}

	return assigns, nil
}

const (
	alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base     = len(alphabet)
)

// uniqueRefIdentifier generates a unique reference identifier for an expression.
// This is used to avoid conflicts when referencing expressions in the query.
// It uses letters instead of numbers to avoid confusion with subplan references.
// It uses a base 26 system, where A = 0, B = 1, ..., Z = 25, AA = 26, AB = 27, etc.
func (p *planContext) uniqueRefIdentifier() string {
	if p.ReferenceCount == 0 {
		p.ReferenceCount++
		return string(alphabet[0])
	}

	n := p.ReferenceCount

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

	p.ReferenceCount++

	return string(runes)
}
