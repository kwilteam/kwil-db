package aggregate

import "github.com/kwilteam/kwil-db/parse/sql/tree"

type groupByAnalyzer struct {
	tree.AstWalker
	context *context
}

func NewGroupByWalker() tree.AstWalker {
	return &groupByAnalyzer{
		AstWalker: tree.NewBaseWalker(),
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
