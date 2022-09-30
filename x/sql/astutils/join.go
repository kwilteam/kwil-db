package astutils

import (
	"strings"

	"kwil/x/sql/ast"
)

func Join(list *ast.List, sep string) string {
	items := []string{}
	for _, item := range list.Items {
		if n, ok := item.(*ast.String); ok {
			items = append(items, n.Str)
		}
	}
	return strings.Join(items, sep)
}
