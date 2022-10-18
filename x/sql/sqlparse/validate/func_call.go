package validate

import (
	"errors"
	"fmt"

	"kwil/x/sql/catalog"
	"kwil/x/sql/sqlerr"
	"kwil/x/sql/sqlparse/ast"
	"kwil/x/sql/sqlparse/astutils"
)

type funcCallVisitor struct {
	catalog *catalog.Catalog
	strict  bool
	err     error
}

func (v *funcCallVisitor) Visit(node ast.Node) astutils.Visitor {
	if v.err != nil {
		return nil
	}

	call, ok := node.(*ast.FuncCall)
	if !ok {
		return v
	}

	fn := call.Func
	if fn == nil {
		return v
	}

	// Custom validation for kwil.arg
	// TODO: Replace this once type-checking is implemented
	if fn.Schema == "kwil" {
		if !(fn.Name == "arg" || fn.Name == "narg") {
			v.err = sqlerr.FunctionNotFound("kwil." + fn.Name)
			return nil
		}

		if len(call.Args.Items) != 1 {
			v.err = &sqlerr.Error{
				Message:  fmt.Sprintf("expected 1 parameter to kwil.arg; got %d", len(call.Args.Items)),
				Location: call.Pos(),
			}
			return nil
		}
		switch n := call.Args.Items[0].(type) {
		case *ast.A_Const:
		case *ast.ColumnRef:
		default:
			v.err = &sqlerr.Error{
				Message:  fmt.Sprintf("expected parameter to kwil.arg to be string or reference; got %T", n),
				Location: call.Pos(),
			}
			return nil
		}

		// If we have kwil.arg or kwil.narg, there is no need to resolve the function call.
		// It won't resolve anyway, sinc it is not a real function.
		return nil
	}

	fun, err := v.catalog.ResolveFuncCall(call)
	if fun != nil {
		return v
	}
	if errors.Is(err, sqlerr.ErrNotFound) && !v.strict {
		return v
	}
	v.err = err
	return nil
}

func FuncCall(c *catalog.Catalog, n ast.Node, strict bool) error {
	visitor := funcCallVisitor{catalog: c, strict: strict}
	astutils.Walk(&visitor, n)
	return visitor.err
}
