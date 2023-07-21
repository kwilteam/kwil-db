package order

import (
	"errors"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

func (o *orderAnalyzer) ExitOrderBy(node *tree.OrderBy) error {
	if o.context.IsCompound {
		// sort all result columns
		node.OrderingTerms = append(node.OrderingTerms, o.context.getReturnedColumnOrderingTerms()...)
		return nil
	}

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

func (o *orderAnalyzer) EnterSelectStmt(node *tree.SelectStmt) error {
	o.newScope()

	if len(node.SelectCores) > 1 {
		o.context.IsCompound = true
	}

	o.context.ResultColumns = node.SelectCores[0].Columns

	return nil
}

func (o *orderAnalyzer) ExitSelectStmt(node *tree.SelectStmt) error {
	o.oldScope()
	return nil
}

func (o *orderAnalyzer) ExitSelectCore(node *tree.SelectCore) error {
	o.context.currentSelectPosition++

	return nil
}

func (o *orderAnalyzer) EnterTableOrSubqueryTable(node *tree.TableOrSubqueryTable) error {
	if o.context.currentSelectPosition != 0 {
		return nil
	}
	tbl, err := findTable(o.schemaTables, node.Name)
	if err != nil {
		return err
	}

	identifier := node.Name
	if node.Alias != "" {
		identifier = node.Alias
	}

	o.context.PrimaryUsedTables = append(o.context.PrimaryUsedTables, &types.Table{
		Name:        identifier,
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
