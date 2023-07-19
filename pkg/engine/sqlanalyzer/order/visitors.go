package order

import (
	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

func NewOrderVisitors() *OrderVisitors {
	ctx := &orderContext{
		JoinedTables: make([]*types.Table, 0),
	}

	return &OrderVisitors{
		analyzer: &orderAnalyzer{
			BaseVisitor: tree.NewBaseVisitor(),
			context:     ctx,
		},
		contextualizer: &orderContextualizer{
			BaseVisitor: tree.NewBaseVisitor(),
			context:     ctx,
		},
	}
}

type OrderVisitors struct {
	analyzer       tree.Visitor
	contextualizer tree.Visitor
}

func (o *OrderVisitors) Analyzer() tree.Visitor {
	return o.analyzer
}

func (o *OrderVisitors) Contextualizer() tree.Visitor {
	return o.contextualizer
}
