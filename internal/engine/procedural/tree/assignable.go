package tree

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/procedural/types"
)

// Assigner is a type that can be assigned to a variable.
type Assigner interface {
	PGMarshaler
	assigns(info *SystemInfo) (ReturnType, types.DataType, error)
}

// VariableDeclaration is a variable declaration.
type VariableDeclaration struct {
	Variable *Variable // name is the variable name, without the $.
	Type     types.DataType
}

// VariableAssignment is a variable assignment.
// It assigns to an already declared variable.
type VariableAssignment struct {
	Variable *Variable
	Value    Assigner
}

// MarshalPG will return the variable assignment as a string.
func (v *VariableAssignment) MarshalPG(info *SystemInfo) (string, error) {
	variable, ok := info.Context.Variables[v.Variable.Name]
	if !ok {
		return "", fmt.Errorf("variable %s not found", v.Variable.Name)
	}

	returnType, dataType, err := v.Value.assigns(info)
	if err != nil {
		return "", err
	}

	if returnType != ReturnTypeValue {
		return "", fmt.Errorf("cannot assign non-value to variable")
	}

	if !variable.Equals(dataType) {
		return "", fmt.Errorf("cannot assign %s to %s", dataType, variable)
	}

	str := strings.Builder{}
	str.WriteString(v.Variable.Name)
	str.WriteString(" := ")
	value, err := v.Value.MarshalPG(info)
	if err != nil {
		return "", err
	}

	str.WriteString(value)
	str.WriteString(";")

	return str.String(), nil
}

func (VariableAssignment) clause() {}

// VariableDeclarationAssignment is a variable declaration that also assigns a value to the variable.
// Variables can be assigned to constants, other variables, arithmetic expressions, or procedure calls.
type VariableDeclarationAssignment struct {
	Declaration *VariableDeclaration
	Value       Assigner
}

// MarshalPG will return the variable assignment as a string.
func (v *VariableDeclarationAssignment) MarshalPG(info *SystemInfo) (string, error) {
	// TODO: we will need to ensure that the procedure declares the variable
	// before we can assign it. this will have to be done with a walker that walks the
	// tree and identifies all variables, and then declares them
	returnType, dataType, err := v.Value.assigns(info)
	if err != nil {
		return "", err
	}
	if returnType != ReturnTypeValue {
		return "", fmt.Errorf("cannot assign non-value to variable")
	}

	if !v.Declaration.Type.Equals(dataType) {
		return "", fmt.Errorf("cannot assign %s to %s", dataType, v.Declaration.Type)
	}

	str := strings.Builder{}
	str.WriteString(v.Declaration.Variable.Name)
	str.WriteString(" := ")
	value, err := v.Value.MarshalPG(info)
	if err != nil {
		return "", err
	}

	str.WriteString(value)
	str.WriteString(";")

	return str.String(), nil
}

func (VariableDeclarationAssignment) clause() {}

// ReturnType is the type of return value.
type ReturnType uint8

const (
	ReturnTypeNone ReturnType = iota
	ReturnTypeValue
	ReturnTypeTable
)
