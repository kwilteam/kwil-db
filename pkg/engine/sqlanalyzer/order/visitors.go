package order

import (
	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

func NewOrderWalker(tables []*types.Table) tree.Walker {
	// copy tables, since we will be modifying the tables slice to register CTEs
	tbls := make([]*types.Table, len(tables))
	copy(tbls, tables)

	return &orderAnalyzer{
		Walker:       tree.NewBaseWalker(),
		schemaTables: tbls,
	}
}

type orderAnalyzer struct {
	tree.Walker
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
