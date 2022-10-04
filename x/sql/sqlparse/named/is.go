package named

import (
	"kwil/x/sql/sqlparse/ast"
	"kwil/x/sql/sqlparse/astutils"
)

// IsParamFunc fulfills the astutils.Search
func IsParamFunc(node ast.Node) bool {
	call, ok := node.(*ast.FuncCall)
	if !ok {
		return false
	}

	if call.Func == nil {
		return false
	}

	isValid := call.Func.Schema == "kwil" && (call.Func.Name == "arg" || call.Func.Name == "narg")
	return isValid
}

func IsParamSign(node ast.Node) bool {
	expr, ok := node.(*ast.A_Expr)
	return ok && astutils.Join(expr.Name, ".") == "@"
}
