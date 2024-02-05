package aggregate

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/utils"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func (g *groupByAnalyzer) EnterResultColumnExpression(r *tree.ResultColumnExpression) error {
	g.context.newReturningClauseContext()
	return nil
}

func (g *groupByAnalyzer) EnterResultColumnStar(r *tree.ResultColumnStar) error {
	g.context.newReturningClauseContext()
	g.context.returningClauseContext.containsSelectAll = true
	return nil
}

func (g *groupByAnalyzer) EnterResultColumnTable(r *tree.ResultColumnTable) error {
	g.context.newReturningClauseContext()
	g.context.returningClauseContext.containsSelectAll = true
	return nil
}

func (g *groupByAnalyzer) ExitResultColumnExpression(r *tree.ResultColumnExpression) error {
	g.context.oldReturningClauseContext()
	return nil
}

func (g *groupByAnalyzer) ExitResultColumnStar(r *tree.ResultColumnStar) error {
	g.context.oldReturningClauseContext()
	return nil
}

func (g *groupByAnalyzer) ExitResultColumnTable(r *tree.ResultColumnTable) error {
	g.context.oldReturningClauseContext()
	return nil
}

func (g *groupByAnalyzer) EnterExpressionFunction(a *tree.ExpressionFunction) error {
	_, isAggFunc := a.Function.(*tree.AggregateFunc)
	if !isAggFunc {
		return nil
	}

	if !g.returningClauseContextDefined() {
		return nil
	}

	g.context.returningClauseContext.containsAggregateFunc = true
	g.context.returningClauseContext.currentlyInsideAggregateFunc = true

	for i, arg := range a.Inputs {
		if i == 0 {
			continue
		}

		containedColumns := utils.SearchResultColumns(arg)
		if len(containedColumns) > 0 {
			return fmt.Errorf("%w: %s", ErrAggregateFuncHasInvalidPosArg, a.Function.Name())
		}
	}

	return nil
}

func (g *groupByAnalyzer) EnterSelectStmt(s *tree.SelectStmt) error {
	if !g.returningClauseContextDefined() {
		return nil
	}

	if g.context.returningClauseContext.currentlyInsideAggregateFunc {
		return fmt.Errorf("%w", ErrAggregateFuncContainsSubquery)
	}
	return nil
}

func (g *groupByAnalyzer) EnterExpressionColumn(e *tree.ExpressionColumn) error {
	if !g.returningClauseContextDefined() {
		return nil
	}

	if !g.context.returningClauseContext.currentlyInsideAggregateFunc {
		g.context.returningClauseContext.bareColumns = append(g.context.returningClauseContext.bareColumns, e)
	}

	return nil
}

func (g *groupByAnalyzer) returningClauseContextDefined() bool {
	if g.context == nil {
		return false
	}
	if g.context.returningClauseContext == nil {
		return false
	}

	return true
}

/*
// containsValidAggregateFunc recursively checks if the expression contains an aggregate function
// that also contains a column in its first argument, and does not contains columns in its others.
// if it contains a valid aggregate function, it returns true, otherwise false.
// if it encounters an aggregate function that contains a column in an argument that is not its first,
// it returns an error.
func containsValidAggregateFunction(expr tree.Expression) (*tree.ExpressionColumn, error) {
	switch e := expr.(type) {
	case *tree.ExpressionLiteral:
		return nil, nil
	case *tree.ExpressionBindParameter:
		return nil, nil
	case *tree.ExpressionColumn:
		return nil, nil
	case *tree.ExpressionUnary:
		return containsValidAggregateFunction(e.Operand)
	case *tree.ExpressionBinaryComparison:
		contains, err := containsValidAggregateFunction(e.Left)
		if err != nil {
			return false, err
		}

		if contains {
			return true, nil
		}

		return containsValidAggregateFunction(e.Right)
	case *tree.ExpressionFunction:
		_, isAggFunc := e.Function.(*tree.AggregateFunc)
		if !isAggFunc {
			for _, input := range e.Inputs {
				contains, err := containsValidAggregateFunction(input)
				if err != nil {
					return false, err
				}

				if contains {
					return true, nil
				}
			}

			return nil, nil
		}

		if len(e.Inputs) == 0 {
			return nil, nil
		}
		for i, arg := range e.Inputs {
			if i == 0 {
				continue
			}

			if containsColumn(arg) {
				return false, fmt.Errorf("aggregate function %s cannot contain a column in an argument that is not its first", e.Function.name())
			}
		}

		return containsColumn(e.Inputs[0]), nil
	case *tree.ExpressionList:
		for _, arg := range e.Expressions {
			contains, err := containsValidAggregateFunction(arg)
			if err != nil {
				return false, err
			}

			if contains {
				return true, nil
			}
		}

		return nil, nil
	case *tree.ExpressionCollate:
		return containsValidAggregateFunction(e.Expression)
	case *tree.ExpressionStringCompare:
		contains, err := containsValidAggregateFunction(e.Left)
		if err != nil {
			return false, err
		}

		if contains {
			return true, nil
		}

		contains, err = containsValidAggregateFunction(e.Right)
		if err != nil {
			return false, err
		}

		if contains {
			return true, nil
		}

		return containsValidAggregateFunction(e.Escape)
	case *tree.ExpressionIsNull:
		return containsValidAggregateFunction(e.Expression)
	case *tree.ExpressionDistinct:
		return containsValidAggregateFunction(e.Left)
	case *tree.ExpressionBetween:
		contains, err := containsValidAggregateFunction(e.Expression)
		if err != nil {
			return false, err
		}

		if contains {
			return true, nil
		}

		contains, err = containsValidAggregateFunction(e.Left)
		if err != nil {
			return false, err
		}

		if contains {
			return true, nil
		}

		return containsValidAggregateFunction(e.Right)
	case *tree.ExpressionSelect:
		newWalker := NewGroupByWalker()
		err := e.Select.Accept(newWalker)
		return false, err
	case *tree.ExpressionCase:
		contains, err := containsValidAggregateFunction(e.CaseExpression)
		if err != nil {
			return false, err
		}

		if contains {
			return true, nil
		}

		for _, pair := range e.WhenThenPairs {
			if len(pair) != 2 {
				return false, fmt.Errorf("expected pair to have 2 elements, got %d", len(pair))
			}
			containsWhen, err := containsValidAggregateFunction(pair[0])
			if err != nil {
				return false, err
			}

			if containsWhen {
				return true, nil
			}

			containsThen, err := containsValidAggregateFunction(pair[1])
			if err != nil {
				return false, err
			}

			if containsThen {
				return true, nil
			}
		}

		return containsValidAggregateFunction(e.ElseExpression)
	case *tree.ExpressionArithmetic:
		contains, err := containsValidAggregateFunction(e.Left)
		if err != nil {
			return false, err
		}

		if contains {
			return true, nil
		}

		return containsValidAggregateFunction(e.Right)
	default:
		return false, fmt.Errorf("unknown expression type %T", expr)
	}
}

func containsColumn(expr tree.Expression) bool {
	return false
}
*/
