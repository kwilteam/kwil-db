package attributes

import (
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

// TableFromAttributes generates a table structure from a list of relation attributes.
// It will do it's best to interpret the proper name for the attributes.
// If a column is given (either as a table.column or just column), it will apply the new table name.
// If any other expression is given (math on top of a column, aggregate, etc), it will enforce that an alias
// is given in the ResultColumnExpression of the relation attribute.  The types of ambiguous naming supported by SQLite for CTEs is
// not clear from their docs, so this is to be safe.
// It takes a boolean to determine if a primary key should be added to the table.
// If true, the primary key is simply a composite key of all of the columns in the table.
// If it will return two columns of the same name, it will add a suffix of ":1", ":2", etc.
func TableFromAttributes(tableName string, attrs []*RelationAttribute, withPrimaryKey bool) (*types.Table, error) {
	cols := []*types.Column{}
	nameCounts := map[string]int{}

	for _, attr := range attrs {
		var colToAdd *types.Column

		// if it's a column, then we can just use that
		exprColumn, ok := attr.ResultExpression.Expression.(*tree.ExpressionColumn)
		if ok {
			colName := exprColumn.Column
			if attr.ResultExpression.Alias != "" {
				colName = attr.ResultExpression.Alias
			}

			colToAdd = &types.Column{
				Name: colName,
				Type: attr.Type,
			}
		} else {
			// else we need to make sure it has an alias
			if attr.ResultExpression.Alias == "" {
				return nil, fmt.Errorf("%w: result columns that contain complex statements must have an alias", ErrInvalidReturnExpression)
			}

			colToAdd = &types.Column{
				Name: attr.ResultExpression.Alias,
				Type: attr.Type,
			}
		}

		timesAppeared, ok := nameCounts[colToAdd.Name]
		if ok {
			nameCounts[colToAdd.Name] = timesAppeared + 1
			colToAdd.Name = fmt.Sprintf("%s:%d", colToAdd.Name, timesAppeared)
		} else {
			nameCounts[colToAdd.Name] = 1
		}

		cols = append(cols, colToAdd)
	}

	table := &types.Table{
		Name:    tableName,
		Columns: cols,
	}

	if withPrimaryKey {
		colNames := []string{}
		for _, col := range cols {
			colNames = append(colNames, col.Name)
		}

		table.Indexes = []*types.Index{
			{
				Name:    fmt.Sprintf("%s_pk", tableName),
				Columns: colNames,
				Type:    types.PRIMARY,
			},
		}
	}

	return table, nil
}
