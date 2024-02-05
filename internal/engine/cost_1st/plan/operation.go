package plan

import (
	"bytes"

	"github.com/kwilteam/kwil-db/internal/engine/cost/operator"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

//// RelationOperation is the interface for all relational operations.
//// It can be visited by a RelationOperationVisitor.
//// It represents a physical relational operation.
//type RelationOperation interface {
//	operator.Visitor
//
//	Inputs() []RelationOperation
//}

//type logicalSchema

// Operation represents a relational operation node.
// It represents a (logical???) relational operation.
type Operation struct {
	op operator.Operator

	//cost float64

	inputs []Operation
}

func (o *Operation) Inputs() []Operation {
	return o.inputs
}

func (o *Operation) Explain() string {
	return o.explain("", "")
}

func (o *Operation) explain(titlePrefix string, bodyPrefix string) string {
	var msg bytes.Buffer
	msg.WriteString(titlePrefix)

	ov := operator.NewExplainVisitor()
	msg.WriteString(o.op.Accept(ov).(string))
	msg.WriteString("\n")

	for _, child := range o.inputs {
		msg.WriteString(child.explain(
			bodyPrefix+"->  ",
			bodyPrefix+"      "))
	}
	return msg.String()

}

func (o *Operation) cost() int {
	cost := 0
	for _, child := range o.inputs {
		cost += child.cost()
	}
	return 0
}

func NewOperation(op operator.Operator, inputs []Operation) *Operation {
	return &Operation{
		op:     op,
		inputs: inputs,
	}
}

// ExpressionResolver resolves an expression to a column.
type ExpressionResolver struct {
	scope *scope

	mapping map[string]string
}

type scope struct {
	parent *scope

	mapping map[string]string //
}

func (s *scope) descend() *scope {
	return &scope{
		parent: s,
	}
}

type operationNode struct {
}

type OperationContext struct {
	scope *scope

	ctes map[string]*tree.CTE

	allColumns []*OutputColumn
}

type OperationBuilder struct {
	ctx    *OperationContext
	op     operator.Operator
	inputs []*OperationBuilder

	// temporary schema
	schema *schema
}

type OutputColumn struct {
	OriginalTblName string
	OriginalColName string
	TblName         string
	ColName         string
	DB              string

	used bool
}

type Outputs []*OutputColumn

func NewOperationBuilder(ctx *OperationContext, op operator.Operator,
	inputs ...*OperationBuilder) *OperationBuilder {
	return &OperationBuilder{
		op:     op,
		inputs: inputs,
		ctx:    ctx,
	}
}

func (b *OperationBuilder) AddChild(c *OperationBuilder) {
	b.inputs = append(b.inputs, c)
}

func (b *OperationBuilder) WithNewRoot(root operator.Operator) *OperationBuilder {
	return &OperationBuilder{
		op:     root,
		inputs: []*OperationBuilder{b},
		ctx:    b.ctx,
	}
}

// Build builds the operation tree recursively.
func (b *OperationBuilder) Build() *Operation {
	if len(b.inputs) > 0 {
		inputs := make([]Operation, len(b.inputs))
		for i, input := range b.inputs {
			inputs[i] = *input.Build()
		}
		return NewOperation(b.op, inputs)
	} else {
		return NewOperation(b.op, nil)
	}
}
