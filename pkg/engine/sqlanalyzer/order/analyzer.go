package order

import (
	"errors"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

// ExitOrderBy adds the required ordering terms to the statement.
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

// EnterSelectStmt creates a new scope.
// if the statement does not have an order by clause, one is created.
// it checks if the statement is a compound statement, and if so, sets the flag.
func (o *orderAnalyzer) EnterSelectStmt(node *tree.SelectStmt) error {
	o.newScope()

	// a bug was found where nil OrderBy would cause no ordering terms to be added
	// this needs to be cleaned up later if there are no ordering terms
	if node.OrderBy == nil {
		node.OrderBy = &tree.OrderBy{
			OrderingTerms: []*tree.OrderingTerm{},
		}
	}

	if len(node.SelectCores) > 1 {
		o.context.IsCompound = true
	}

	o.context.ResultColumns = node.SelectCores[0].Columns

	return nil
}

// ExitSelectStmt pops the current scope.
func (o *orderAnalyzer) ExitSelectStmt(node *tree.SelectStmt) error {
	// we created a provisional order by clause in case one does not exist
	// we clean it up here if it is empty
	if node.OrderBy != nil && len(node.OrderBy.OrderingTerms) == 0 {
		node.OrderBy = nil
	}

	o.oldScope()
	return nil
}

// ExitSelectCore increments the current select position.
// This is can be used to determine compound select position.
func (o *orderAnalyzer) ExitSelectCore(node *tree.SelectCore) error {
	o.context.currentSelectPosition++

	return nil
}

// EnterTableOrSubqueryTable adds the table to the list of used tables.
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
