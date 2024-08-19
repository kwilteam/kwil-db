package optimizer

import (
	"fmt"

	"github.com/kwilteam/kwil-db/parse/planner/logical"
)

// PushdownPredicates pushes down filters to the lowest possible level.
// It rewrites the logical plan to push down filters as far as possible.
// The returned plan should always be used, and the original plan should be discarded,
// as it might be in an inconsistent state.
func PushdownPredicates(n logical.Plan) (logical.Plan, error) {
	// we start by pushing down filters
	plan, err := push(n, nil)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

// push is a recursive function that pushes down filters.
// It passes the filter expression to the next node.
// The expr can be nil if there is no filter to push.
// It returns the expression that could not be pushed down,
// which should be set as the filter.
func push(n logical.Plan, expr logical.Expression) (logical.Plan, error) {
	// pushLeftRight is a helper function that determines whether or not an
	// expression can be pushed down either side of a join.
	// It is defined separately here because the logic is used both in Join
	// and CartesianProduct.
	pushLeftRight := func(left, right logical.Plan, expr logical.Expression) (logical.Plan, logical.Plan, logical.Expression) {
		ands := splitAnds(expr)
		var leftover logical.Expression

		for _, and := range ands {
			cols := findColumns(and)

			var leftCount, rightCount int

			// if all columns are from one side, push down the filter to that side.
			// otherwise, apply the filter to the join condition.
			leftRel := left.Relation()
			for _, field := range leftRel.Fields {
				if _, ok := cols[[2]string{field.Parent, field.Name}]; ok {
					leftCount++
				}
			}

			rightRel := right.Relation()
			for _, field := range rightRel.Fields {
				if _, ok := cols[[2]string{field.Parent, field.Name}]; ok {
					rightCount++
				}
			}

			switch {
			case leftCount == 0 && rightCount == 0:
				// we can't push down the filter
				leftover = makeAnd(leftover, and)
			case leftCount == 0 && rightCount > 0:
				// push down to the right side
				res, err := push(right, and)
				if err != nil {
					return nil, nil, nil
				}

				right = res
			case leftCount > 0 && rightCount == 0:
				// push down to the left side
				res, err := push(left, and)
				if err != nil {
					return nil, nil, nil
				}

				left = res
			case leftCount > 0 && rightCount > 0:
				// apply the filter to the join condition
				leftover = makeAnd(leftover, and)
			default:
				panic("unexpected column count case")
			}
		}

		return left, right, leftover
	}

	switch n := n.(type) {
	case *logical.Filter:
		if expr != nil {
			return nil, fmt.Errorf("unexpected pushdown of filter to filter")
		}

		if _, ok := n.Child.(*logical.Aggregate); ok {
			// we can't push down filters to aggregates
			return n, nil
		}

		fin, err := push(n.Child, n.Condition)
		if err != nil {
			return nil, err
		}

		// since we no longer have a condition, we can just return the child
		return fin, nil
	case *logical.Join:
		left, right, leftover := pushLeftRight(n.Left, n.Right, expr)
		n.Left = left
		n.Right = right
		n.Condition = makeAnd(n.Condition, leftover)

		return n, nil
	case *logical.Scan:
		n.Filter = makeAnd(n.Filter, expr)
		return n, nil
	case *logical.Project:
		res, err := push(n.Child, expr)
		if err != nil {
			return nil, err
		}

		n.Child = res
		return n, nil
	case *logical.CartesianProduct:
		if expr == nil {
			return n, nil
		}

		left, right, leftover := pushLeftRight(n.Left, n.Right, expr)
		n.Left = left
		n.Right = right
		// if leftover is not nil, then we can rewrite as a join
		if leftover == nil {
			return n, nil
		}

		return &logical.Join{
			Left:      left,
			Right:     right,
			Condition: leftover,
		}, nil
	case *logical.Update:
		res, err := push(n.Child, expr)
		if err != nil {
			return nil, err
		}

		n.Child = res
		return n, nil
	case *logical.Delete:
		res, err := push(n.Child, expr)
		if err != nil {
			return nil, err
		}

		n.Child = res
		return n, nil
	case *logical.Aggregate:
		res, err := push(n.Child, nil)
		if err != nil {
			return nil, err
		}

		n.Child = res
		return n, nil
	default:
		if expr != nil {
			return nil, fmt.Errorf("unhandled predicate pushdown for %T", n)
		}

		// by default, we just can't push down the filter, but we should visit the children.
		// If any rewrite is attempted, we should return an error.
		for _, child := range n.Plans() {
			res, err := push(child, nil)
			if err != nil {
				return nil, err
			}
			if res != child {
				return nil, fmt.Errorf("unhandled rewrite: tried to rewrite a %T as a child of a %T", res, n)
			}
		}

		return n, nil
	}
}

// makeAnd combines two expressions with an AND operator.
// If either expression is nil, the other expression is returned.
func makeAnd(left, right logical.Expression) logical.Expression {
	if left == nil {
		return right
	}

	if right == nil {
		return left
	}

	return &logical.LogicalOp{
		Left:  left,
		Op:    logical.And,
		Right: right,
	}
}

// findColumns returns all columns in the expression.
// It does not search for columns in subqueries.
func findColumns(expr logical.Expression) map[[2]string]*logical.ColumnRef {
	cols := make(map[[2]string]*logical.ColumnRef)
	logical.Traverse(expr, func(node logical.Traversable) bool {
		switch node := node.(type) {
		case *logical.ColumnRef:
			cols[[2]string{node.Parent, node.ColumnName}] = node
			return false
		case *logical.SubqueryExpr:
			return false
		}
		return true
	})

	return cols
}

// splitAnds splits a tree of AND expressions into a list of expressions.
func splitAnds(expr logical.Expression) []logical.Expression {
	and, ok := expr.(*logical.LogicalOp)
	if !ok {
		return []logical.Expression{expr}
	}

	if and.Op != logical.And {
		return []logical.Expression{expr}
	}

	return append(splitAnds(and.Left), splitAnds(and.Right)...)
}
