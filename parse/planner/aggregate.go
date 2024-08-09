package planner

import (
	"fmt"
	"reflect"
	"sort"
)

/*
	This file handles enforcement of aggregation rules in the logical plan.
	For queries that have group by clauses, we need to ensure that all referenced
	columns are either in the group by clause or are aggregated. Furthermore, if an
	expression such as "a + 1" is used in the group by clause, then that same expression
	must be used in the select clause (simply referencing "a" is not enough).

	This file has an aggregateChecker struct which takes a list of expressions from the
	GROUP BY clause, flattens them, and indexes them according to their used columns.
	These can then be re-used to check that expressions built on top of the aggregate
	are valid.

	It checks for validity by flattening the incoming expressions, and checking that
	for all unaggregated columns, there is a matching expression in the GROUP BY clause.
	To account for referencing the same column many times within one expression, it cuts
	away the matching expression from the flattened expression that is being validated,
	and continues until all columns are accounted for.
*/

// aggregateChecker is a helper struct that performs validation on aggregate functions.
// It stores a flattened list of expressions, which can be used to check
// the validity of expressions relying on aggregated relations.
type aggregateChecker struct {
	// includedColumns contains the column (relation and name) and its associated
	// expressions. The same column can be present with multiple expressions, so we
	// store it in a 2d list (e.g. we can have GROUP BY id+1, id+2)
	// Each 2d slice of logical expressions will be sorted longest to shortest.
	includedColumns map[[2]string][][]LogicalExpr
}

// newAggregateChecker creates a new aggregateChecker from a list of expressions.
// It will flatten the expressions, so that they can be used to check
// the validity of other expressions. It will also validate that no aggregate functions
// or subqueries are present in the expressions.
func newAggregateChecker(exprs []LogicalExpr) (*aggregateChecker, error) {
	var err error

	in := map[[2]string][][]LogicalExpr{}
	for _, expr := range exprs {
		var cols [][2]string // track all columns used in the expression
		var plan []LogicalExpr
		traverse(expr, func(node LogicalNode) bool {
			switch n := node.(type) {
			case *AggregateFunctionCall:
				err = fmt.Errorf("aggregate functions are not allowed in GROUP BY clause")
				return false
			case *SubqueryExpr:
				err = fmt.Errorf("subqueries are not allowed in GROUP BY clause")
				return false
			case *ColumnRef:
				cols = append(cols, [2]string{n.Parent, n.ColumnName})
				return true
			default:
				expr, ok := n.(LogicalExpr)
				if !ok {
					// this should never happen, since we can only reach non-LogicalExpr
					// via Subquery, which is handled above
					err = fmt.Errorf("unexpected node type %T in GROUP BY clause. this is an internal bug", n)
					return false
				}

				plan = append(plan, expr)
				return true
			}
		})
		if err != nil {
			return nil, err
		}

		for _, col := range cols {
			planList, ok := in[col]
			if !ok {
				planList = [][]LogicalExpr{}
			}
			planList = append(planList, plan)
			in[col] = planList
		}
	}

	// sort the plans from longest to shortest
	for _, plans := range in {
		sort.Slice(plans, func(i, j int) bool {
			return len(plans[i]) > len(plans[j])
		})
	}

	return &aggregateChecker{includedColumns: in}, nil
}

// checkMany takes a list of expressions, and checks that they are valid
// given the expressions in the GROUP BY clause.
func (a *aggregateChecker) checkMany(exprs []LogicalExpr) error {
	for _, expr := range exprs {
		err := a.check(expr)
		if err != nil {
			return err
		}
	}

	return nil
}

// check takes a logical expression, and checks that for any columns it uses that
// are not captured by aggregates, that it has a subexpression that exactly matches
// one of the expressions in the GROUP BY clause.
func (a *aggregateChecker) check(e LogicalExpr) error {
	// how we actually achieve this is that we will walk the given expression,
	// and if a column is found that is not aggregated, we will
	// check that there is a flattened plan for that column. We check for
	// plan equality by cutting the matching plan from the list of plans.
	// e.g. if we have [a,b,c,d,e], and [b,c] would be a matching plan,
	// while [b,d] would not be.

	// we track foundCols in a slice because if the same column is referenced twice,
	// we need to cut for it twice
	var foundCols [][2]string
	var traversed []LogicalExpr
	traverse(e, func(node LogicalNode) bool {
		switch n := node.(type) {
		case *AggregateFunctionCall:
			// if it is an aggregate, we don't care what it does
			// and can skip it
			return false
		case *ColumnRef:
			// if there is a column, we will need to check that we have a matching
			// plan for it.
			foundCols = append(foundCols, [2]string{n.Parent, n.ColumnName})
			return true
		default:
			expr, ok := n.(LogicalExpr)
			if !ok {
				// if it is not a logical expression, we can just return true.
				// we want to traverse all subqueries, so we do not return false,
				// but since there wont be any LogicalPlan nodes in our "included"
				// list, we dont need to worry about them.
				return true
			}

			traversed = append(traversed, expr)
			return true
		}
	})
	// fast path: if there are no column references, we can return early
	if len(foundCols) == 0 {
		return nil
	}

	// another fast path, since this is a commen error case and effects parse time.
	// we check that all columns have a reference in the included list.
	for _, col := range foundCols {
		_, ok := a.includedColumns[col]
		if !ok {
			colName := col[1]
			if col[0] != "" {
				colName = col[0] + "." + colName
			}
			// I wish we could get a better error message here containing the full
			// column that still needs to be referenced, but even Postgres cant return
			// that, since it is essentially impossible to determine the full missing
			// expression.
			return fmt.Errorf("column %s must be included in GROUP BY clause", colName)
		}
	}

	// now, for each found column, we need to check that it has a matching group by plan.
	// If it doesn't, we will return an error. If it does, we will cut the matching plan,
	// so that we can check that columns used several times all have matching plans.
	for _, col := range foundCols {
		// already checked that it exists, so we can safely index
		plans := a.includedColumns[col]

		// attempt to cut each plan from the list
		found := false
		for _, plan := range plans {
			found = cutFrom(&traversed, plan, equalExpr)
			if found {
				break
			}
		}
		if !found {
			colName := col[1]
			if col[0] != "" {
				colName = col[0] + "." + colName
			}
			return fmt.Errorf("column %s must be included in GROUP BY clause", colName)
		}
	}

	return nil
}

// cutFrom attempts to cut b from a. If it is successful, it will return
// true, and modify a to remove the matching elements. If it is not successful,
// it will return false, and a will be unmodified.
func cutFrom[T any](a *[]T, b []T, equal func(T, T) bool) bool {
	if len(b) == 0 {
		return true
	}

	aRef := *a

	if len(b) > len(aRef) {
		return false
	}

	for i := 0; i <= len(aRef)-len(b); i++ {
		match := true
		for j := 0; j < len(b); j++ {
			if !equal(aRef[i+j], b[j]) {
				match = false
				break
			}
		}
		if match {
			// Found a match, remove the matching elements
			copy(aRef[i:], aRef[i+len(b):])
			a2 := aRef[:len(aRef)-len(b)]
			*a = a2
			return true
		}
	}
	return false
}

// equalExpr checks if two expressions are equal.
func equalExpr(a, b LogicalExpr) bool {
	return reflect.DeepEqual(a, b)
	// TODO: once we include position info in the expressions,
	// deep equality will not be enough, initial work is done below

	// switch a := a.(type) {
	// case *ColumnRef:
	// 	b, ok := b.(*ColumnRef)
	// 	if !ok {
	// 		return false
	// 	}

	// 	return a.Parent == b.Parent && a.ColumnName == b.ColumnName
	// case *Literal:
	// 	b, ok := b.(*Literal)
	// 	if !ok {
	// 		return false
	// 	}

	// 	if !a.Type.Equals(b.Type) {
	// 		return false
	// 	}

	// 	return reflect.DeepEqual(a.Value, b.Value)
	// case *Variable:
	// 	b, ok := b.(*Variable)
	// 	if !ok {
	// 		return false
	// 	}

	// 	return a.VarName == b.VarName
	// case *FunctionCall:
	// 	b, ok := b.(*FunctionCall)
	// 	if !ok {
	// 		return false
	// 	}

	// 	if a.FunctionName != b.FunctionName {
	// 		return false
	// 	}

	// 	if
	// }
}

// getAggregateTerms takes an expression and gets all used aggregate terms.
// It will not get aggregate terms from subqueries, or aggregates within aggregates.
// TODO: we need a more sophisticated way of detecting aggregation usage for
// correlated subqueries
func getAggregateTerms(e LogicalExpr) []*AggregateFunctionCall {
	var aggs []*AggregateFunctionCall
	traverse(e, func(node LogicalNode) bool {
		switch n := node.(type) {
		case *AggregateFunctionCall:
			aggs = append(aggs, n)
			return false
		case *SubqueryExpr:
			return false
		case LogicalPlan:
			// if it is a plan like a scan / project, exit
			return false
		default:
			return true
		}
	})

	return aggs
}

// mergeAggregates merges two arrays of exprs
// It does this by flattening the expressions, and then
// checking that each node is equal to another node in the other list.
// To avoid making this O(n^2), we will convert each element to a string,
// and check that their string representations are equal. Only if this
// is the case will we then compare the actual nodes.
func mergeAggregates(a []*AggregateFunctionCall, b []*AggregateFunctionCall) []*AggregateFunctionCall {
	flattenWithoutSubqueries := func(a *AggregateFunctionCall) []LogicalExpr {
		var flat []LogicalExpr
		traverse(a, func(n LogicalNode) bool {
			switch n := n.(type) {
			case *SubqueryExpr:
				return false
			default:
				expr, ok := n.(LogicalExpr)
				if !ok {
					return false
				}

				flat = append(flat, expr)
				return true
			}
		})
		return flat
	}

	aMap := make(map[string][]LogicalExpr)
	for _, node := range a {
		aMap[node.String()] = flattenWithoutSubqueries(node)
	}

	unique := a
	for _, node := range b {
		key := node.String()
		plan, ok := aMap[key]
		if !ok {
			// if not found, it is a new node
			unique = append(unique, node)
			continue
		}

		// if found, we need to check that the plans are equal
		bFlat := flattenWithoutSubqueries(node)

		if len(plan) != len(bFlat) {
			unique = append(unique, node)
			continue
		}

		equal := true
		for i := range plan {
			if !equalExpr(plan[i], bFlat[i]) {
				equal = false
				break
			}
		}

		if !equal {
			unique = append(unique, node)
		}
	}

	return unique
}
