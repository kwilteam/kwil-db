package parameters

import "github.com/kwilteam/kwil-db/internal/parse/sql/tree"

// RenameVariables calls the callback function for each bind parameter in the statement.
func RenameVariables(ast tree.AstWalker, fn func(b string) string) error {
	walker := &tree.ImplementedListener{
		FuncEnterExpressionBindParameter: func(p0 *tree.ExpressionBindParameter) error {
			p0.Parameter = fn(p0.Parameter)
			return nil
		},
	}

	return ast.Walk(walker)
}
