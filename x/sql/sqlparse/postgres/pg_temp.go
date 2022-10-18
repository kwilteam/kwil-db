package postgres

import (
	"kwil/x/sql/catalog"
)

func pgTemp() *catalog.Schema {
	return &catalog.Schema{Name: "pg_temp"}
}

func typeName(name string) *catalog.QualName {
	return &catalog.QualName{Name: name}
}

func argN(name string, n int) *catalog.Function {
	var args []*catalog.Argument
	for i := 0; i < n; i++ {
		args = append(args, &catalog.Argument{
			Type: &catalog.QualName{Name: "any"},
		})
	}
	return &catalog.Function{
		Name:       name,
		Args:       args,
		ReturnType: &catalog.QualName{Name: "any"},
	}
}
