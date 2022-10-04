package engine

import (
	"strings"

	"kwil/x/sql/catalog"
	"kwil/x/sql/sqlparse/ast"
	"kwil/x/sql/sqlparse/astutils"
)

func isArray(n *ast.TypeName) bool {
	if n == nil {
		return false
	}
	return len(n.ArrayBounds.Items) > 0
}

func toColumn(n *ast.TypeName) *Column {
	if n == nil {
		panic("can't build column for nil type name")
	}
	typ, err := ParseTypeName(n)
	if err != nil {
		panic("toColumn: " + err.Error())
	}
	return &Column{
		Type:     &catalog.QualName{Catalog: typ.Catalog, Schema: typ.Schema, Name: typ.Name},
		DataType: strings.TrimPrefix(astutils.Join(n.Names, "."), "."),
		NotNull:  true, // XXX: How do we know if this should be null?
		IsArray:  isArray(n),
	}
}
