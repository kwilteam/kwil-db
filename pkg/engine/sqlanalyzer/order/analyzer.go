package order

import (
	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
)

type orderAnalyzer struct {
	*tree.BaseVisitor
	context *orderContext
}

func (o *orderAnalyzer) VisitOrderBy(node *tree.OrderBy) error {
	required, err := o.context.requiredOrderingTerms()
	if err != nil {
		return err
	}

	for _, term := range required {
		generatedTerm, err := o.context.generateOrder(term)
		if err != nil {
			return err
		}

		node.OrderingTerms = append(node.OrderingTerms, generatedTerm)
	}

	return nil
}

type orderableTerm struct {
	Table  string
	Column string
}
