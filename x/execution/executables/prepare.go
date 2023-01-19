package executables

import (
	"fmt"
	"kwil/x/execution"
	datatypes "kwil/x/types/data_types"
	execTypes "kwil/x/types/execution"
)

// prepare prepares execution of a query.  It does not check access control rights.
func (d *executableInterface) Prepare(query string, caller string, inputs []*execTypes.UserInput) (string, []any, error) {
	// get the executable
	executable := d.Executables[query]
	if executable == nil {
		return "", nil, fmt.Errorf(`query "%s" does not exist`, query)
	}

	// convert user inputs to a map for easier lookup
	inputMap := make(map[string]*execTypes.UserInput)
	for _, input := range inputs {
		inputMap[input.Name] = input
	}

	// now loop through and fill in the returns
	returns := make([]any, len(executable.Args))
	for _, arg := range executable.Args {
		// if it is static, just set the value
		if arg.Static {
			defVal, err := determineDefault(arg, caller)
			if err != nil {
				return "", nil, fmt.Errorf(`invalid default for arg "%s": %w`, arg.Name, err)
			}

			returns[arg.Position] = defVal
			continue
		}

		// if not static, the arg must contain a corresponding user input
		input, ok := inputMap[arg.Name]
		if !ok {
			return "", nil, fmt.Errorf(`missing user input for arg "%s"`, arg.Name)
		}

		// try to convert value to the arg type
		value, err := datatypes.Utils.ConvertAny(input.Value, arg.Type)
		if err != nil {
			return "", nil, fmt.Errorf(`invalid user input for arg "%s": %w`, arg.Name, err)
		}

		returns[arg.Position] = value
	}

	return executable.Statement, returns, nil
}

// determineDefault will determine the default value for an arg.
// for example, if there is a caller modifier, it will return the caller.
func determineDefault(arg *execTypes.Arg, caller string) (any, error) {
	switch arg.Modifier {
	case execution.NO_MODIFIER:
		return arg.Value, nil
	case execution.CALLER:
		return caller, nil
	default:
		return nil, fmt.Errorf(`invalid modifier "%s"`, arg.Modifier.String())
	}
}
