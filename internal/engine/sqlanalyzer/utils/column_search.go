package utils

import (
	"fmt"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// TODO: test this
// SearchResultColumns returns a list of columns used the expression.
// It does not return columns used in subqueries.
func SearchResultColumns(expr tree.Expression) []*tree.ExpressionColumn {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *tree.ExpressionLiteral:
		return nil
	case *tree.ExpressionBindParameter:
		return nil
	case *tree.ExpressionColumn:
		return []*tree.ExpressionColumn{e}
	case *tree.ExpressionUnary:
		return SearchResultColumns(e.Operand)
	case *tree.ExpressionBinaryComparison:
		return nil // a binary expression will not return a result column
	case *tree.ExpressionFunction:
		found := make([]*tree.ExpressionColumn, 0)
		for _, arg := range e.Inputs {
			found = append(found, SearchResultColumns(arg)...)
		}

		return found
	case *tree.ExpressionList:
		found := make([]*tree.ExpressionColumn, 0)
		for _, arg := range e.Expressions {
			found = append(found, SearchResultColumns(arg)...)
		}

		return found
	case *tree.ExpressionCollate:
		return SearchResultColumns(e.Expression)
	case *tree.ExpressionStringCompare:
		return append(SearchResultColumns(e.Left), SearchResultColumns(e.Right)...)
	case *tree.ExpressionIs:
		return append(SearchResultColumns(e.Left), SearchResultColumns(e.Right)...)
	case *tree.ExpressionBetween:
		return SearchResultColumns(e.Expression)
	case *tree.ExpressionSelect:
		return nil
	case *tree.ExpressionCase:
		found := SearchResultColumns(e.CaseExpression)
		for _, pair := range e.WhenThenPairs {
			for _, expr := range pair {
				found = append(found, SearchResultColumns(expr)...)
			}
		}

		return append(found, SearchResultColumns(e.ElseExpression)...)
	case *tree.ExpressionArithmetic:
		return append(SearchResultColumns(e.Left), SearchResultColumns(e.Right)...)
	}

	fmt.Println("UNEXPECTED BUG: unhandled expression type in AstListener column search", expr)
	return nil
}
