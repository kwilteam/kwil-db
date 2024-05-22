package parse

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
)

var (
	Functions = map[string]*FunctionDefinition{
		"abs": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.IntType) {
					return nil, wrapErrArgumentType(types.IntType, args[0])
				}

				return types.IntType, nil
			},
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "", errStar("abs")
				}
				if distinct {
					return "", errDistinct("abs")
				}

				return fmt.Sprintf("abs(%s)", inputs[0]), nil
			},
		},
		"error": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.NullType, nil
			},
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "", errStar("error")
				}
				if distinct {
					return "", errDistinct("error")
				}

				return fmt.Sprintf("error(%s)", inputs[0]), nil
			},
		},
		"length": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.IntType, nil
			},
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "", fmt.Errorf("cannot use * with length")
				}
				if distinct {
					return "", fmt.Errorf("cannot use distinct with length")
				}

				return fmt.Sprintf("length(%s)", inputs[0]), nil
			},
		},
		"lower": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.TextType, nil
			},
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "", errStar("lower")
				}
				if distinct {
					return "", errDistinct("lower")
				}

				return fmt.Sprintf("lower(%s)", inputs[0]), nil
			},
		},
		"upper": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.TextType, nil
			},
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "", errStar("upper")
				}
				if distinct {
					return "", errDistinct("upper")
				}

				return fmt.Sprintf("upper(%s)", inputs[0]), nil
			},
		},
		"format": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("invalid number of arguments: expected at least 1, got %d", len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.TextType, nil
			},
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "", errStar("format")
				}
				if distinct {
					return "", errDistinct("format")
				}

				return fmt.Sprintf("format(%s)", strings.Join(inputs, ", ")), nil
			},
		},
		"uuid_generate_v5": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// first argument must be a uuid, second argument must be text
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				if !args[0].EqualsStrict(types.UUIDType) {
					return nil, wrapErrArgumentType(types.UUIDType, args[0])
				}

				if !args[1].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[1])
				}

				return types.UUIDType, nil
			},
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "", errStar("uuid_generate_v5")
				}
				if distinct {
					return "", errDistinct("uuid_generate_v5")
				}

				return fmt.Sprintf("uuid_generate_v5(%s)", strings.Join(inputs, ", ")), nil
			},
		},
		// array functions
		"array_append": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				if !args[0].IsArray {
					return nil, fmt.Errorf("%w: expected first argument to be an array, got %s", ErrType, args[0].String())
				}

				if args[1].IsArray {
					return nil, fmt.Errorf("%w: expected second argument to be a scalar, got %s", ErrType, args[1].String())
				}

				if !strings.EqualFold(args[0].Name, args[1].Name) {
					return nil, fmt.Errorf("%w: append type must be equal to scalar array type: array type: %s append type: %s", ErrType, args[0].Name, args[1].Name)
				}

				return args[0], nil
			},
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "", errStar("array_append")
				}
				if distinct {
					return "", errDistinct("array_append")
				}

				return fmt.Sprintf("array_append(%s)", strings.Join(inputs, ", ")), nil
			},
		},
		"array_prepend": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				if args[0].IsArray {
					return nil, fmt.Errorf("%w: expected first argument to be a scalar, got %s", ErrType, args[0].String())
				}

				if !args[1].IsArray {
					return nil, fmt.Errorf("%w: expected second argument to be an array, got %s", ErrType, args[1].String())
				}

				if !strings.EqualFold(args[0].Name, args[1].Name) {
					return nil, fmt.Errorf("%w: prepend type must be equal to scalar array type: array type: %s prepend type: %s", ErrType, args[1].Name, args[0].Name)
				}

				return args[1], nil
			},
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "", errStar("array_prepend")
				}
				if distinct {
					return "", errDistinct("array_prepend")
				}

				return fmt.Sprintf("array_prepend(%s)", strings.Join(inputs, ", ")), nil
			},
		},
		"array_cat": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				if !args[0].IsArray {
					return nil, fmt.Errorf("%w: expected first argument to be an array, got %s", ErrType, args[0].String())
				}

				if !args[1].IsArray {
					return nil, fmt.Errorf("%w: expected second argument to be an array, got %s", ErrType, args[1].String())
				}

				if !strings.EqualFold(args[0].Name, args[1].Name) {
					return nil, fmt.Errorf("%w: expected both arrays to be of the same scalar type, got %s and %s", ErrType, args[0].Name, args[1].Name)
				}

				return args[0], nil
			},
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "", errStar("array_cat")
				}
				if distinct {
					return "", errDistinct("array_cat")
				}

				return fmt.Sprintf("array_cat(%s)", strings.Join(inputs, ", ")), nil
			},
		},
		"array_length": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].IsArray {
					return nil, fmt.Errorf("expected argument to be an array, got %s", args[0].String())
				}

				return types.IntType, nil
			},
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "", errStar("array_length")
				}
				if distinct {
					return "", errDistinct("array_length")
				}

				return fmt.Sprintf("array_length(%s, 1)", inputs[0]), nil
			},
		},
		// Aggregate functions
		"count": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) > 1 {
					return nil, fmt.Errorf("invalid number of arguments: expected at most 1, got %d", len(args))
				}

				return types.IntType, nil
			},
			IsAggregate: true,
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "count(*)", nil
				}
				if distinct {
					return fmt.Sprintf("count(DISTINCT %s)", inputs[0]), nil
				}

				return fmt.Sprintf("count(%s)", inputs[0]), nil
			},
			StarArgReturn: types.IntType,
		},
		"sum": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.IntType) {
					return nil, wrapErrArgumentType(types.IntType, args[0])
				}

				return types.IntType, nil
			},
			IsAggregate: true,
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "", errStar("sum")
				}
				if distinct {
					return "sum(DISTINCT %s)", nil
				}

				return fmt.Sprintf("sum(%s)", inputs[0]), nil
			},
			StarArgReturn: types.IntType,
		},
	}
)

func errDistinct(funcName string) error {
	return fmt.Errorf(`%w: cannot use DISTINCT with function "%s"`, ErrFunctionSignature, funcName)
}

func errStar(funcName string) error {
	return fmt.Errorf(`%w: cannot use * with function "%s"`, ErrFunctionSignature, funcName)
}

// FunctionDefinition defines a function that can be used in the database.
type FunctionDefinition struct {
	// ValidateArgs is a function that checks the arguments passed to the function.
	// It can check the argument type and amount of arguments.
	// It returns the expected return type based on the arguments.
	ValidateArgs func(args []*types.DataType) (*types.DataType, error)
	// StarArgReturn is the type the function returns if * is passed as the sole
	// argument. If it is nil, the function does not support *.
	StarArgReturn *types.DataType
	// IsAggregate is true if the function is an aggregate function.
	IsAggregate bool
	// PGFormat is a function that formats the inputs to the function in Postgres format.
	// For example, the function `sum` would format the inputs as `sum($1)`.
	// It will be given the same amount of inputs as ValidateArgs() was given.
	// ValidateArgs will always be called first.
	PGFormat FormatFunc
}

// FormatFunc is a function that formats a string of inputs for a SQL function.
type FormatFunc func(inputs []string, distinct bool, star bool) (string, error)

func wrapErrArgumentNumber(expected, got int) error {
	return fmt.Errorf("%w: expected %d, got %d", ErrFunctionSignature, expected, got)
}

func wrapErrArgumentType(expected, got *types.DataType) error {
	return fmt.Errorf("%w: expected %s, got %s", ErrType, expected.String(), got.String())
}
