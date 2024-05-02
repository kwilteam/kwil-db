package parameters

import "github.com/kwilteam/kwil-db/parse/sql/tree"

// RenameVariables calls the callback function for each bind parameter in the statement.
// It does not verify syntax / semantics, so it does not need an error listener.
func RenameVariables(ast tree.AstWalker, fn func(b string) string) error {
	walker := &tree.ImplementedListener{
		FuncEnterExpressionBindParameter: func(p0 *tree.ExpressionBindParameter) error {
			p0.Parameter = fn(p0.Parameter)
			return nil
		},
	}

	return ast.Walk(walker)
}
