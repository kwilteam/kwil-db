package order

import (
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

func NewOrderWalker(tables []*types.Table) tree.AstWalker {
	// copy tables, since we will be modifying the tables slice to register CTEs
	tbls := make([]*types.Table, len(tables))
	copy(tbls, tables)

	return &orderAnalyzer{
		AstWalker:    tree.NewBaseWalker(),
		schemaTables: tbls,
	}
}

type orderAnalyzer struct {
	tree.AstWalker
	context *orderContext
	// schemaTables is a list of all tables in the schema
	schemaTables []*types.Table
}

func (o *orderAnalyzer) newScope() {
	oldCtx := o.context
	o.context = newOrderContext(oldCtx)
}

// oldScope pops the current scope and returns to the parent scope
// if there is no parent scope, it simply sets the current scope to nil
func (o *orderAnalyzer) oldScope() {
	if o.context == nil {
		return
	}

	o.context = o.context.Parent
}
