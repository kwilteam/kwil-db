package dto

import "kwil/x/execution"

type Executable struct {
	Name      string
	Statement string

	UserInputs []*UserInput
	Args       []*Arg
}

type Arg struct {
	// position of the arg in the query
	Position int

	// if the arg is static, it will not be filled by the user
	Static bool

	// the name of the input
	Name string

	// the value of the arg if it is static
	Value any

	// the type of the arg
	Type execution.DataType

	// the modifier of the arg
	Modifier execution.ModifierType
}

// This is what the user has to send back in order to execute
type UserInput struct {
	// Position is the position of the input relative to the rest of the inputs
	// for example, if there are 3 2 params and 1 default, and 1 where and 1 default where, the user-inputted WHERE will be position 3
	Name string `json:"name" yaml:"name"`

	// InputType is the type of input this is
	Type execution.DataType `json:"type" yaml:"type"`

	// Value is the value of the input
	Value string `json:"value" yaml:"value"`
}
