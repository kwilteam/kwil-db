package aggregate

import "github.com/kwilteam/kwil-db/parse/sql/tree"

type context struct {
	isAggregateStatement bool

	groupedColumns         map[tree.ExpressionColumn]struct{}
	havingClause           tree.Expression
	returningClauses       []*analyzedReturningClause
	returningClauseContext *returningClauseContext
	containsSelectAll      bool

	Parent *context
}

func newContext(parent *context) *context {
	return &context{
		groupedColumns:   make(map[tree.ExpressionColumn]struct{}),
		returningClauses: make([]*analyzedReturningClause, 0),

		Parent: parent,
	}
}

type analyzedReturningClause struct {
	containsAggregateFunc bool
	bareColumns           []*tree.ExpressionColumn
	containsSelectAll     bool
}

type returningClauseContext struct {
	containsAggregateFunc        bool
	bareColumns                  []*tree.ExpressionColumn
	currentlyInsideAggregateFunc bool
	containsSelectAll            bool

	Parent *returningClauseContext
}

func newReturningClauseContext(parent *returningClauseContext) *returningClauseContext {
	return &returningClauseContext{
		Parent: parent,
	}
}

func (g *context) newReturningClauseContext() {
	oldCtx := g.returningClauseContext
	g.returningClauseContext = newReturningClauseContext(oldCtx)
}

func (g *context) oldReturningClauseContext() {
	if g.returningClauseContext == nil {
		return
	}

	g.returningClauses = append(g.returningClauses, &analyzedReturningClause{
		containsAggregateFunc: g.returningClauseContext.containsAggregateFunc,
		bareColumns:           g.returningClauseContext.bareColumns,
		containsSelectAll:     g.returningClauseContext.containsSelectAll,
	})
	g.returningClauseContext = g.returningClauseContext.Parent
}
