package parameters

import (
	"fmt"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// // NamedParametersVisitor visits the AST and changes all $var binds to @var_arg
// // binds so that a uniform (with @caller etc.) numbered parameter rewriting can
// // be done downstream.
// type NamedParametersVisitor struct {
// 	tree.Walker
// 	Binds map[string]string // base => new param e.g. "id": "@id_arg"
// }

// func NewNamedParametersVisitor() *NamedParametersVisitor {
// 	return &NamedParametersVisitor{
// 		Walker: tree.NewBaseWalker(),
// 		Binds:  make(map[string]string),
// 	}
// }

// func (p *NamedParametersVisitor) EnterExpressionBindParameter(b *tree.ExpressionBindParameter) error {
// 	// $id => @id_arg
// 	if base, cut := strings.CutPrefix(b.Parameter, "$"); cut {
// 		b.Parameter = "@" + base + "_arg"
// 		p.Binds[base] = b.Parameter
// 		return nil
// 	}

// 	if base, cut := strings.CutPrefix(b.Parameter, "@"); cut {
// 		p.Binds[base] = b.Parameter
// 		return nil
// 	}

// 	// shouldn't happen since parser says binds are prefixed by [$@]

// 	p.Binds[b.Parameter] = b.Parameter

// 	return nil
// }

// ParametersVisitor visits the AST and replaces all bind parameters with numbered parameters.
type ParametersVisitor struct {
	// OrderedParameters are the passed named identifiers in the order that they have become numbered.
	// For example, if a query was SELECT * FROM tbl WHERE a = $a AND b = $b, the query would be rewritten
	// as SELECT * FROM tbl WHERE a = $1 AND b = $2, and OrderedParameters would be []string{"$a", "$b"}.
	OrderedParameters []string
	renamedParams     map[string]string // maps $bindParam to $1
	tree.Walker
}

func NewParametersVisitor() *ParametersVisitor {
	return &ParametersVisitor{
		OrderedParameters: []string{},
		renamedParams:     map[string]string{},
		Walker:            tree.NewBaseWalker(),
	}
}

func (p *ParametersVisitor) EnterExpressionBindParameter(b *tree.ExpressionBindParameter) error {
	// check if the parameter has already been numbered
	// if not, then we will number it
	if param, ok := p.renamedParams[b.Parameter]; ok {
		b.Parameter = param
		return nil
	}

	// the parameter has not been numbered yet
	num := len(p.OrderedParameters) + 1
	p.OrderedParameters = append(p.OrderedParameters, b.Parameter)

	numberedName := "$" + fmt.Sprint(num)

	// rename the parameter
	p.renamedParams[b.Parameter] = numberedName

	b.Parameter = numberedName

	return nil
}
