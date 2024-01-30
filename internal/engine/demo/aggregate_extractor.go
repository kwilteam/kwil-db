package demo

import "github.com/kwilteam/kwil-db/parse/sql/tree"

type aggregateExtractor struct {
	*tree.BaseAstVisitor

	// aggrList is the list of aggregate functions in the query.
	aggrList []*tree.AggregateFunc
}

func newAggregateExtractor() *aggregateExtractor {
	return &aggregateExtractor{
		BaseAstVisitor: &tree.BaseAstVisitor{},
	}
}

func (e *aggregateExtractor) extractAggregates(cols []tree.ResultColumn) []*tree.AggregateFunc {

	return e.aggrList
}

// VisitResultColumnExpression visits the result column expression.
func (e *aggregateExtractor) VisitResultColumnExpression(node *tree.ResultColumnExpression) any {
	if ef, ok := node.Expression.(*tree.ExpressionFunction); ok {
		if f, ok := ef.Function.(*tree.AggregateFunc); ok {
			e.aggrList = append(e.aggrList, f)
		}
	}
	return nil
}
