package utils

import "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"

type Column struct {
	Table  string
	Column string
}

// SearchResultColumns returns a list of columns returned by the expression.
func SearchResultColumns(expr tree.Expression) []*Column {
	switch e := expr.(type) {
	case *tree.ExpressionLiteral:
		return nil
	case *tree.ExpressionBindParameter:
		return nil
	case *tree.ExpressionColumn:
		return []*Column{&Column{Table: e.Table, Column: e.Column}}
	case *tree.ExpressionUnary:
		return SearchResultColumns(e.Operand)
	case *tree.ExpressionBinaryComparison:
		return nil // a binary expression will not return a result column
	case *tree.ExpressionFunction:
		found := make([]*Column, 0)
		for _, arg := range e.Inputs {
			found = append(found, SearchResultColumns(arg)...)
		}

		return found
	case *tree.ExpressionList:
		found := make([]*Column, 0)
		for _, arg := range e.Expressions {
			found = append(found, SearchResultColumns(arg)...)
		}

		return found
	case *tree.ExpressionCollate:
		return SearchResultColumns(e.Expression)
	case *tree.ExpressionStringCompare:
		return append(SearchResultColumns(e.Left), SearchResultColumns(e.Right)...)
	case *tree.ExpressionIsNull:
		return SearchResultColumns(e.Expression)
	case *tree.ExpressionDistinct:
		return append(SearchResultColumns(e.Left), SearchResultColumns(e.Right)...)
	case *tree.ExpressionBetween:
		return SearchResultColumns(e.Expression)
	case *tree.ExpressionSelect:
	case *tree.ExpressionCase:
	case *tree.ExpressionArithmetic:
	}

	return nil
}
