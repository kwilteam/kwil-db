package parameters

import (
	"fmt"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// ParametersVisitor visits the AST and replaces all bind parameters with numbered parameters.
type ParametersVisitor struct {
	numberedParams []string
	renamedParams  map[string]string // maps $bindParam to $1
	tree.Walker
}

func NewParametersVisitor() *ParametersVisitor {
	return &ParametersVisitor{
		numberedParams: []string{},
		renamedParams:  map[string]string{},
		Walker:         tree.NewBaseWalker(),
	}
}

// NumberedParameters returns the passed named identifiers in the order that they have become numbered.
// For example, if a query was SELECT * FROM tbl WHERE a = $a AND b = $b, the query would be rewritten
// as SELECT * FROM tbl WHERE a = $1 AND b = $2, and this function would return []string{"$a", "$b"}.
func (p *ParametersVisitor) NumberedParameters() []string {
	return p.numberedParams
}

func (p *ParametersVisitor) EnterExpressionBindParameter(b *tree.ExpressionBindParameter) error {
	// check if the parameter has already been numbered
	// if not, then we will number it
	if param, ok := p.renamedParams[b.Parameter]; ok {
		b.Parameter = param
		return nil
	}

	// the parameter has not been numbered yet
	num := len(p.numberedParams) + 1
	p.numberedParams = append(p.numberedParams, b.Parameter)

	numberedName := "$" + fmt.Sprint(num)

	// rename the parameter
	p.renamedParams[b.Parameter] = numberedName

	b.Parameter = numberedName

	return nil
}
