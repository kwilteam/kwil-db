package parameters

import (
	"fmt"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// ParametersWalker walks the AST and replaces all bind parameters with numbered parameters.
type ParametersWalker struct {
	// OrderedParameters are the passed named identifiers in the order that they have become numbered.
	// For example, if a query was SELECT * FROM tbl WHERE a = $a AND b = $b, the query would be rewritten
	// as SELECT * FROM tbl WHERE a = $1 AND b = $2, and OrderedParameters would be []string{"$a", "$b"}.
	OrderedParameters []string
	renamedParams     map[string]string // maps $bindParam to $1
	tree.AstListener
}

func NewParametersWalker() *ParametersWalker {
	return &ParametersWalker{
		OrderedParameters: []string{},
		renamedParams:     map[string]string{},
		AstListener:       tree.NewBaseListener(),
	}
}

func (p *ParametersWalker) EnterExpressionBindParameter(b *tree.ExpressionBindParameter) error {
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
