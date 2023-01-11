package execution

import "kwil/x/execution"

type Executable struct {
	Name      string `json:"name" yaml:"name"`
	Statement string `json:"statement" yaml:"statement"`

	UserInputs []*UserInput `json:"user_inputs" yaml:"user_inputs"`
	Args       []*Arg       `json:"args" yaml:"args"`
}

type Arg struct {
	// position of the arg in the query
	Position uint8 `json:"position" yaml:"position"`

	// if the arg is static, it will not be filled by the user
	Static bool `json:"static" yaml:"static"`

	// the name of the input
	Name string `json:"name" yaml:"name"`

	// the value of the arg if it is static
	Value any `json:"value" yaml:"value"`

	// the type of the arg
	Type execution.DataType `json:"type" yaml:"type"`

	// the modifier of the arg
	Modifier execution.ModifierType `json:"modifier" yaml:"modifier"`
}

// This is what the user has to send back in order to execute
type UserInput struct {
	// Position is the position of the input relative to the rest of the inputs
	// for example, if there are 3 2 params and 1 default, and 1 where and 1 default where, the user-inputted WHERE will be position 3
	Name string `json:"name" yaml:"name"` // Name is the name of the input

	// Value is the value of the input
	Value any `json:"value" yaml:"value"` // Value is the value of the input
}
