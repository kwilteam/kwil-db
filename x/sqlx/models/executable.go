package models

import (
	"encoding/json"
	"fmt"
	types "kwil/x/sqlx"
)

type ExecutableQuery struct {
	Name      string
	Statement string
	Table     string
	Type      string

	UserInputs []*UserInput
	Args       []*Arg
}

type Arg struct {
	// position of the arg in the query
	Position int

	// if the arg is static, it will not be filled by the user
	Static bool

	// the position of the arg in the user input
	InputPosition int

	// the value of the arg if it is static
	Value any

	// the type of the arg
	Type types.DataType

	// the modifier of the arg
	Modifier types.ModifierType
}

// maps column name to value
type UserInput struct {
	// Position is the position of the input relative to the rest of the inputs
	// for example, if there are 3 2 params and 1 default, and 1 where and 1 default where, the user-inputted WHERE will be position 3
	Position int `json:"position" yaml:"position"`

	// InputType is the type of input this is
	Type types.DataType `json:"type" yaml:"type"`

	// Value is the value of the input
	Value string `json:"value" yaml:"value"`
}

func (e *ExecutableQuery) Bytes() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ExecutableQuery) Unmarshal(b []byte) error {
	return json.Unmarshal(b, e)
}

// an interface for building Args from params and where clauses
type arger interface {
	getColumn() string
	getModifier() string
	getStatic() bool
	getValue() any
}

// buildArg will build an arg from a param or where clause
func buildArg(tbl *Table, position int, param arger) (*Arg, error) {
	col := tbl.GetColumn(param.getColumn())
	if col == nil {
		return nil, fmt.Errorf(`column "%s" does not exist`, param.getColumn())
	}

	// get kwil type from column type
	kwilType, err := types.Conversion.StringToKwilType(col.Type)
	if err != nil {
		return nil, fmt.Errorf(`invalid column type "%s" for column "%s": %w`, col.Type, param.getColumn(), err)
	}

	mod, err := types.Conversion.ConvertModifier(param.getModifier())
	if err != nil {
		return nil, fmt.Errorf(`invalid modifier for where clause on column "%s": %w`, param.getColumn(), err)
	}

	return &Arg{
		Position:      position,
		Static:        param.getStatic(),
		Value:         param.getValue(),
		Type:          kwilType,
		InputPosition: -1,
		Modifier:      mod,
	}, nil
}

// buildInput will build a user input from an arg
func (a *Arg) buildInput(position int) *UserInput {
	a.InputPosition = position
	return &UserInput{
		Position: position,
		Type:     a.Type,
		Value:    fmt.Sprintf("%v", a.Value),
	}
}
