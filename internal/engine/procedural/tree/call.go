package tree

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/procedural/types"
)

// ProcedureCall is a call to another procedure.
type ProcedureCall struct {
	// Schema is the schema of the procedure.
	// If it is a built-in procedure, this is empty.
	Schema string
	// Name is the name of the procedure.
	// It should always be lower case.
	Name string
	// Arguments are the arguments to the procedure.
	Arguments []Expression // can be nil
}

func (p *ProcedureCall) MarshalPG(info *SystemInfo) (string, error) {
	str := strings.Builder{}
	if p.Schema != "" {
		str.WriteString(p.Schema)
		str.WriteRune('.')
	}
	str.WriteString(p.Name)
	str.WriteRune('(')
	for i, arg := range p.Arguments {
		if i != 0 && i < len(p.Arguments) {
			str.WriteString(", ")
		}
		argStr, err := arg.MarshalPG(info)
		if err != nil {
			return "", err
		}
		str.WriteString(argStr)
	}
	str.WriteRune(')')

	return str.String(), nil
}

func (ProcedureCall) expression() {}
func (p *ProcedureCall) assigns(info *SystemInfo) (ReturnType, types.DataType, error) {
	schema, ok := info.Schemas[p.Schema]
	if !ok {
		return 0, nil, fmt.Errorf("schema %s not found while detecting assignment variable", p.Schema)
	}

	proc, ok := schema.Procedures[p.Name]
	if !ok {
		return 0, nil, fmt.Errorf("procedure %s not found while detecting assignment variable", p.Name)
	}

	if proc.Returns == nil {
		return ReturnTypeNone, nil, nil
	}

	dataType, ok := proc.Returns.(types.DataType)
	if ok {
		return ReturnTypeValue, dataType, nil
	}

	return ReturnTypeTable, nil, nil
}

func (p *ProcedureCall) loopterm(info *SystemInfo) (string, error) {
	return p.MarshalPG(info)
}

// BareProcedureCall is a call to a procedure that does not return a value.
type BareProcedureCall struct {
	P *ProcedureCall
}

func (b *BareProcedureCall) MarshalPG(info *SystemInfo) (string, error) {
	str, err := b.P.MarshalPG(info)
	if err != nil {
		return "", err
	}

	return str + ";", nil
}

func (BareProcedureCall) clause() {}
