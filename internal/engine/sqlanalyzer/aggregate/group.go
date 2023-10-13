package aggregate

import "github.com/kwilteam/kwil-db/parse/sql/tree"

type groupByAnalyzer struct {
	tree.Walker
	context *context
}

func NewGroupByWalker() tree.Walker {
	return &groupByAnalyzer{
		Walker: tree.NewBaseWalker(),
	}
}

func (g *groupByAnalyzer) newContext() {
	oldCtx := g.context
	g.context = newContext(oldCtx)
}

func (g *groupByAnalyzer) oldContext() {
	if g.context == nil {
		return
	}

	g.context = g.context.Parent
}
