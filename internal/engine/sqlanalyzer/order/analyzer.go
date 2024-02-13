package order

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/attributes"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/utils"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// EnterExpressionFunction adds aggregate function columns to the list of aggregated columns.
func (o *orderAnalyzer) EnterExpressionFunction(node *tree.ExpressionFunction) error {
	switch node.Function.(type) {
	default:
		return nil
	case *tree.AggregateFunc:
		for _, arg := range node.Inputs {
			cols := utils.SearchResultColumns(arg)
			switch len(cols) {
			case 0:
				continue
			case 1:
				o.context.AggregatedColumns[orderableTerm{
					Table:  cols[0].Table,
					Column: cols[0].Column,
				}] = struct{}{}
			default:
				return fmt.Errorf("aggregate function with multiple columns not supported")
			}
		}
	}

	return nil
}

// EnterGroupBy will remove the columns from the list of aggregated columns.
func (o *orderAnalyzer) EnterGroupBy(node *tree.GroupBy) error {
	for _, expr := range node.Expressions {
		cols := utils.SearchResultColumns(expr)
		for _, col := range cols {
			delete(o.context.AggregatedColumns, orderableTerm{
				Table:  col.Table,
				Column: col.Column,
			})
		}
	}

	return nil

}

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

// EnterCommonTableExpression adds the table to the list of used tables.
// This allows it to be used later for the ordering terms.
func (o *orderAnalyzer) EnterCTE(node *tree.CTE) error {
	if len(node.Select.SelectCores) == 0 {
		return nil
	}

	cteAttributes, err := attributes.GetSelectCoreRelationAttributes(node.Select.SelectCores[0], o.schemaTables)
	if err != nil {
		return err
	}

	cteTable, err := attributes.TableFromAttributes(node.Table, cteAttributes, true)
	if err != nil {
		return err
	}

	o.schemaTables = append(o.schemaTables, cteTable)

	return nil
}

// we need to add common table expressions to the list of the schemas tables, as well as the list of used tables
// this means we need to detect the structure of the common table expression
func findTable(tables []*types.Table, name string) (*types.Table, error) {
	for _, t := range tables {
		if t.Name == name {
			return t, nil
		}
	}

	return nil, fmt.Errorf(`table "%s" not found`, name)
}
