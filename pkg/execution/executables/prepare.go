package executables

import (
	"fmt"
	execution2 "kwil/pkg/execution"
	"kwil/pkg/types/data_types/any_type"
	execTypes "kwil/pkg/types/execution"
	"strings"
)

func (d *executableInterface) Prepare(query, caller string, inputs []*execTypes.UserInput[anytype.KwilAny]) (string, []any, error) {
	// get the executable
	executable, ok := d.Executables[query]
	if !ok {
		return "", nil, fmt.Errorf(`query "%s" does not exist`, query)
	}

	// convert user inputs to a map for easier lookup
	inputMap := make(map[string]*execTypes.UserInput[anytype.KwilAny])
	for _, input := range inputs {
		inputMap[input.Name] = input
	}

	rows := make([]*row, len(executable.Parameters))
	for i, param := range executable.Parameters {
		var val any
		if param.Static {
			val = param.Value
			if param.Modifier == execution2.CALLER {
				val = strings.ToLower(caller)
			}
		} else {
			input, ok := inputMap[param.Name]
			if !ok {
				return "", nil, fmt.Errorf(`required parameter "%s" was not provided`, param.Name)
			}
			val = input.Value.Value()
		}
		rows[i] = &row{
			column: param.Column,
			value:  val,
		}
	}

	// now the wheres
	wheres := make([]*where, len(executable.Where))
	for i, whereClause := range executable.Where {
		var val any
		// this is the same as the process above.  shit needs to get refactored
		if whereClause.Static {
			val = whereClause.Value
			if whereClause.Modifier == execution2.CALLER {
				val = strings.ToLower(caller)
			}
		} else {
			input, ok := inputMap[whereClause.Name]
			if !ok {
				return "", nil, fmt.Errorf(`required parameter "%s" was not provided`, whereClause.Name)
			}
			val = input.Value.Value()
		}

		wheres[i] = &where{
			column:   whereClause.Column,
			value:    val,
			operator: whereClause.Operator,
		}
	}

	switch executable.Type {
	default:
		return "", nil, fmt.Errorf(`invalid executable type "%d"`, executable.Type)
	case execution2.INSERT:
		return d.prepareInsert(executable, rows)
	case execution2.UPDATE:
		return d.prepareUpdate(executable, rows, wheres)
	case execution2.DELETE:
		return d.prepareDelete(executable, wheres)
	}
}

/*
// prepare prepares execution of a query.  It does not check access control rights.
func (d *executableInterface) Prepare(query string, caller string, inputs []*execTypes.UserInput[anytype.KwilAny]) (string, []any, error) {
	// get the executable
	executable := d.Executables[query]
	if executable == nil {
		return "", nil, fmt.Errorf(`query "%s" does not exist`, query)
	}

	// convert user inputs to a map for easier lookup
	inputMap := make(map[string]*execTypes.UserInput[anytype.KwilAny])
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

		returns[arg.Position] = input.Value.Value()
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
*/
