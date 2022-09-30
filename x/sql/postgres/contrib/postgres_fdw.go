// Code generated by sqlc-pg-gen. DO NOT EDIT.

package contrib

import (
	"kwil/x/sql/ast"
	"kwil/x/sql/catalog"
)

func PostgresFdwFuncs0() []*catalog.Function {
	return []*catalog.Function{
		{
			Name: "postgres_fdw_disconnect",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "text"},
				},
			},
			ReturnType: &ast.TypeName{Name: "boolean"},
		},
		{
			Name:       "postgres_fdw_disconnect_all",
			Args:       []*catalog.Argument{},
			ReturnType: &ast.TypeName{Name: "boolean"},
		},
		{
			Name:       "postgres_fdw_get_connections",
			Args:       []*catalog.Argument{},
			ReturnType: &ast.TypeName{Name: "record"},
		},
		{
			Name:       "postgres_fdw_handler",
			Args:       []*catalog.Argument{},
			ReturnType: &ast.TypeName{Name: "fdw_handler"},
		},
		{
			Name: "postgres_fdw_validator",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "text[]"},
				},
				{
					Type: &ast.TypeName{Name: "oid"},
				},
			},
			ReturnType: &ast.TypeName{Name: "void"},
		},
	}
}

func PostgresFdwFuncs() []*catalog.Function {
	funcs := []*catalog.Function{}
	funcs = append(funcs, PostgresFdwFuncs0()...)
	return funcs
}

func PostgresFdw() *catalog.Schema {
	s := &catalog.Schema{Name: "pg_catalog"}
	s.Funcs = PostgresFdwFuncs()
	return s
}
