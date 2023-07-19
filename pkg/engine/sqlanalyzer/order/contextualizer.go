package order

import (
	"errors"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

type orderContextualizer struct {
	*tree.BaseVisitor
	context      *orderContext
	schemaTables []*types.Table
}

func (o *orderContextualizer) VisitTableOrSubqueryTable(node *tree.TableOrSubqueryTable) error {
	tbl, err := findTable(o.schemaTables, node.Name)
	if err != nil {
		return err
	}

	o.context.JoinedTables = append(o.context.JoinedTables, &types.Table{
		Name:        node.Alias,
		Columns:     tbl.Columns,
		Indexes:     tbl.Indexes,
		ForeignKeys: tbl.ForeignKeys,
	})

	return nil
}

func findTable(tables []*types.Table, name string) (*types.Table, error) {
	for _, t := range tables {
		if t.Name == name {
			return t, nil
		}
	}

	return nil, errors.New("table not found")
}
