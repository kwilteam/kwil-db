package attributes

import "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"

// columnWalker walks a tree.Column.
// it will apply the table name to the column if it is not already present.
type columnWalker struct {
	tree.BaseWalker
	tableName string
}

func (c *columnWalker) EnterExpressionColumn(expr *tree.ExpressionColumn) error {
	if expr.Table == "" {
		expr.Table = c.tableName
	}
	return nil
}

// addTableIfNotPresent adds the table name to the column if it is not already present.
func addTableIfNotPresent(tableName string, expr tree.Accepter) error {
	w := &columnWalker{
		tableName: tableName,
	}
	return expr.Accept(w)
}
