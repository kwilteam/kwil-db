package planner2

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils/order"
	"github.com/kwilteam/kwil-db/parse"
)

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
	CTEPlans []*Subplan // TODO: idk if we need this anymore
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

// select builds a logical plan for a select statement.
func (s *scopeContext) selectStmt(node *parse.SelectStatement) (plan LogicalPlan, rel *Relation, err error) {
	if len(node.SelectCores) == 0 {
		panic("no select cores")
	}

	selectCoreRes, err := s.selectCore(node.SelectCores[0])
	if err != nil {
		return nil, nil, err
	}

	// if there is one select core, we want to project after sorting and limiting.
	// if there are multiple select cores, we want to project before the set operation.
	// see the documentation for selectCoreResult for more info as to why
	// we perform this if statement.
	if len(node.SelectCores) == 1 {
		plan = selectCoreRes.plan
		defer func() {
			plan, rel = selectCoreRes.projectFunc(plan)
		}()
	} else {
		// otherwise, apply immediately so that we can apply the set operation(s)
		plan, rel = selectCoreRes.projectFunc(selectCoreRes.plan)

		for i, core := range node.SelectCores[1:] {
			right, err := s.selectCore(core)
			if err != nil {
				return nil, nil, err
			}

			rightPlan, _ := right.projectFunc(right.plan)
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
func (s *scopeContext) selectCore(node *parse.SelectCore) (*selectCoreResult, error) {
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
					return nil, err
				}

				if resultCol.Alias != "" {
					expr = &AliasExpr{
						Expr:  expr,
						Alias: resultCol.Alias,
					}
				}

				exprs = append(exprs, expr)
				rel.Fields = append(rel.Fields, field)
			case *parse.ResultColumnWildcard:
				// if there is no from, we cannot expand the wildcard
				panic(`wildcard "*" cannot be used without a FROM clause`)
			}
		}

		return &selectCoreResult{
			plan: &EmptyScan{},
			projectFunc: func(lp LogicalPlan) (LogicalPlan, *Relation) {
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
			},
		}, nil
	}

	// otherwise, we need to build the from and join clauses
	scan, rel, err := s.table(node.From)
	if err != nil {
		return nil, err
	}
	var plan LogicalPlan = scan

	for _, join := range node.Joins {
		plan, rel, err = s.join(plan, rel, join)
		if err != nil {
			return nil, err
		}
	}

	if node.Where != nil {
		whereExpr, _, err := s.expr(node.Where, rel)
		if err != nil {
			return nil, err
		}

		plan = &Filter{
			Child:     plan,
			Condition: whereExpr,
		}
	}

	// at this point, we have the full relation for the select core, and can expand the columns
	resultColExprs, resultFields, err := s.expandResultCols(rel, node.Columns)
	if err != nil {
		return nil, err
	}

	containsAgg := false
	for _, resultCol := range resultColExprs {
		containsAgg, err = hasAggregate(resultCol)
		if err != nil {
			return nil, err
		}
	}

	// if there is no group by or aggregate, we can apply any distinct and return
	if len(node.GroupBy) == 0 && !containsAgg {
		return &selectCoreResult{
			plan: plan,
			projectFunc: func(lp LogicalPlan) (LogicalPlan, *Relation) {
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
			},
		}, nil
	}

	// otherwise, we need to build the group by and having clauses.
	// This means that for all result columns, we need to rewrite any
	// column references or aggregate usage as columnrefs to the aggregate
	// functions matching term.
	aggTerms := make(map[string]*identFieldPair)      // any aggregate function used in the result or having
	groupingTerms := make(map[string]*IdentifiedExpr) // any grouping term used in the GROUP BY
	aggregateRel := &Relation{}                       // the relation resulting from the aggregation

	aggPlan := &Aggregate{ // defined separately so we can reference it in the below clauses
		Child: plan,
	}
	plan = aggPlan

	for _, groupTerm := range node.GroupBy {
		groupExpr, field, err := s.expr(groupTerm, rel)
		if err != nil {
			return nil, err
		}

		containsAgg, err := hasAggregate(groupExpr)
		if err != nil {
			return nil, err
		}
		if containsAgg {
			return nil, fmt.Errorf(`aggregate functions are not allowed in GROUP BY`)
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
		aggPlan.GroupingExpressions = append(aggPlan.GroupingExpressions, groupExpr)

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
			return nil, err
		}

		// rewrite the having expression to use the aggregate functions
		havingExpr, err = s.rewriteAccordingToAggregate(havingExpr, groupingTerms, aggTerms)

		plan = &Filter{
			Child:     plan,
			Condition: havingExpr,
		}
	}

	// now we need to rewrite the select list to use the aggregate functions
	for i, resultCol := range resultColExprs {
		resultColExprs[i], err = s.rewriteAccordingToAggregate(resultCol, groupingTerms, aggTerms)
		if err != nil {
			return nil, err
		}
	}

	// finally, all of the aggregated columns need to be added to the Aggregate node
	for _, agg := range order.OrderMap(aggTerms) {
		aggPlan.AggregateExpressions = append(aggPlan.AggregateExpressions, agg.Value.Identified)
		aggregateRel.Fields = append(aggregateRel.Fields, agg.Value.Field)
	}

	return &selectCoreResult{
		plan: plan,
		projectFunc: func(lp LogicalPlan) (LogicalPlan, *Relation) {
			var p LogicalPlan = &Project{
				Child:       lp,
				Expressions: resultColExprs,
			}

			if node.Distinct {
				p = &Distinct{
					Child: p,
				}
			}

			return p, aggregateRel
		},
	}, nil
}

// hasAggregate returns true if the expression contains an aggregate function.
func hasAggregate(expr LogicalNode) (bool, error) {
	var hasAggregate bool
	err := visitAllNodes(expr, func(le Traversable) {
		if _, ok := le.(*AggregateFunctionCall); ok {
			hasAggregate = true
		}
	})
	if err != nil {
		return false, err
	}

	return hasAggregate, nil
}

// identFieldPair is a helper struct that pairs an identified expression with a field.
type identFieldPair struct {
	Identified *IdentifiedExpr
	Field      *Field
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
	projectFunc func(LogicalPlan) (LogicalPlan, *Relation)
}

// rewriteAccordingToAggregate rewrites an expression according to the rules of aggregation.
// This is used to rewrite both the select list and having clause to validate that all columns
// are either captured in aggregates or have an exactly matching expression in the group by.
func (s *scopeContext) rewriteAccordingToAggregate(expr LogicalExpr, groupingTerms map[string]*IdentifiedExpr, aggTerms map[string]*identFieldPair) (LogicalExpr, error) {
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
				return nil, false, fmt.Errorf(`column "%s" must appear in the GROUP BY clause or be used in an aggregate function`, le.String())
			case *AggregateFunctionCall:
				// TODO: do we need to check for the aggregate being called on a correlated column?
				// if it matches any aggregate function, we need to rewrite it
				// to that reference. Otherwise, register it as a new aggregate
				identified, ok := aggTerms[le.String()]
				if ok {
					return &ExprRef{
						Identified: identified.Identified,
					}, false, nil
				}

				newIdentified := &IdentifiedExpr{
					Expr: le,
					ID:   s.plan.uniqueRefIdentifier(),
				}

				aggTerms[le.String()] = &identFieldPair{
					Identified: newIdentified,
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
func (s *scopeContext) expandResultCols(rel *Relation, cols []parse.ResultColumn) ([]LogicalExpr, []*Field, error) {
	var resultCols []LogicalExpr
	var resultFields []*Field
	for _, col := range cols {
		switch col := col.(type) {
		default:
			panic(fmt.Sprintf("unexpected result column type %T", col))
		case *parse.ResultColumnExpression:
			expr, field, err := s.expr(col.Expression, rel)
			if err != nil {
				return nil, nil, err
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

			rel.Fields = append(rel.Fields, field)
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

	return resultCols, resultFields, nil
}

// expr visits an expression node.
func (s *scopeContext) expr(node parse.Expression, currentRel *Relation) (LogicalExpr, *Field, error) {
	switch node := node.(type) {
	default:
		panic(fmt.Sprintf("unexpected expression type %T", node))
	case *parse.ExpressionLiteral:
		return cast(&Literal{
			Value: node.Value,
			Type:  node.Type,
		}, node), anonField(node.Type), nil
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
				}, node), &Field{
					Name: node.Name,
					val:  returns,
				}, nil
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
			}, node), returnField, nil
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
		}, node), returnField, nil
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
			}, node), &Field{
				Name: node.Name,
				val:  returns,
			}, nil
	case *parse.ExpressionVariable:
		var val any // can be a data type or object
		dt, ok := s.plan.Variables[node.Name]
		if !ok {
			// might be an object
			obj, ok := s.plan.Objects[node.Name]
			if !ok {
				return nil, nil, fmt.Errorf(`unknown variable "%s"`, node.Name)
			}

			val = obj
		} else {
			val = dt
		}

		return cast(&Variable{
			VarName: node.Name,
		}, node), &Field{val: val}, nil
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
		}, node), field2, nil
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
			}, node), &Field{
				val: firstValCopy,
			}, nil
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
			}, node), &Field{
				val: fieldType,
			}, nil
	case *parse.ExpressionParenthesized:
		expr, field, err := s.expr(node.Inner, currentRel)
		if err != nil {
			return nil, nil, err
		}

		return cast(expr, node), field, nil
	case *parse.ExpressionComparison:
		left, leftField, err := s.expr(node.Left, currentRel)
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
		}, &Field{val: leftField.val}, nil
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
			}, node), field, nil
		}
		if err != nil {
			return nil, nil, err
		}

		return cast(&ColumnRef{
			Parent:     field.Parent,
			ColumnName: field.Name,
		}, node), field, nil
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
					return nil, nil, fmt.Errorf(`all THEN expressions must be of the same type %s, received %s`, returnType, thenExpr)
				}
			}

			whenScalar, err := whenField.Scalar()
			if err != nil {
				return nil, nil, err
			}

			if !expectedWhenType.Equals(whenScalar) {
				return nil, nil, fmt.Errorf(`WHEN expression must be of type %s, received %s`, expectedWhenType, whenExpr)
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
				Name: field.Name,
				val:  field.Type.Copy(),
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

	var sb strings.Builder
	for p.ReferenceCount > 0 {
		remainder := p.ReferenceCount % base
		sb.WriteByte(alphabet[remainder])
		p.ReferenceCount = p.ReferenceCount / base
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
