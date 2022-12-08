package schema

import (
	"encoding/json"
	"fmt"
)

func (d *Database) ListQueries() []string {
	return d.Queries.ListAll()
}

type ExecutableQuery struct {
	Statement string

	// The inputs by order
	Args map[int]arg

	// maps column name to arg position
	UserInputs []*requiredInput
}

type requiredInput struct {
	Column string
	Type   string
}

type arg struct {
	Column   string
	Type     string
	Default  string
	Fillable bool
}

// maps column name to value
type UserInputs map[string]string

/*
PrepareInputs will take a map of column names to values and return the array of inputs and an error.

It will loop through all Args and see if the arg is fillable.  If it is, it will get that value from the userInputs map.
If the value is not in the user inputs, it will check to see if there is a default value.  If there is, it will use that, otherwise it will return an error.

If the arg is not fillable, it will use the default value if there is one, otherwise it will return an error.
*/
func (q *ExecutableQuery) PrepareInputs(usrInpts *UserInputs) ([]string, error) {
	var inputs []string
	i := 1
	// looping through this way to ensure that the inputs are in the correct order
	for {
		arg, ok := q.Args[i]
		if !ok {
			break
		}

		if arg.Fillable {
			val, ok := (*usrInpts)[arg.Column]
			if !ok {
				if arg.Default == "" {
					return nil, fmt.Errorf("missing required input: %s", arg.Column)
				}
				val = arg.Default
			}
			inputs = append(inputs, val)
		} else {
			inputs = append(inputs, arg.Default)
		}

		i++
	}

	return inputs, nil

}

func (e *ExecutableQuery) Bytes() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ExecutableQuery) Unmarshal(b []byte) error {
	return json.Unmarshal(b, e)
}
