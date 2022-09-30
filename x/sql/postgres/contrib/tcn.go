// Code generated by sqlc-pg-gen. DO NOT EDIT.

package contrib

import (
	"kwil/x/sql/ast"
	"kwil/x/sql/catalog"
)

func TcnFuncs0() []*catalog.Function {
	return []*catalog.Function{
		{
			Name:       "triggered_change_notification",
			Args:       []*catalog.Argument{},
			ReturnType: &ast.TypeName{Name: "trigger"},
		},
	}
}

func TcnFuncs() []*catalog.Function {
	funcs := []*catalog.Function{}
	funcs = append(funcs, TcnFuncs0()...)
	return funcs
}

func Tcn() *catalog.Schema {
	s := &catalog.Schema{Name: "pg_catalog"}
	s.Funcs = TcnFuncs()
	return s
}
