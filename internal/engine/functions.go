package engine

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
)

var (
	Functions = map[string]*FunctionDefinition{
		"abs": {
			Args: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].Equals(types.IntType) {
					return nil, wrapErrArgumentType(types.IntType, args[0])
				}

				return types.IntType, nil
			},
			PGName: "abs",
		},
		"error": {
			Args: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].Equals(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.NullType, nil
			},
			PGName: "error",
		},
		"length": {
			Args: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].Equals(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.IntType, nil
			},
			PGName: "length",
		},
		"lower": {
			Args: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].Equals(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.TextType, nil
			},
			PGName: "lower",
		},
		"upper": {
			Args: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].Equals(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.TextType, nil
			},
			PGName: "upper",
		},
		"format": {
			Args: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("invalid number of arguments: expected at least 1, got %d", len(args))
				}

				if !args[0].Equals(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.TextType, nil
			},
			PGName: "format",
		},
		"uuid_generate_v5": {
			Args: func(args []*types.DataType) (*types.DataType, error) {
				// first argument must be a uuid, second argument must be text
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				if !args[0].Equals(types.UUIDType) {
					return nil, wrapErrArgumentType(types.UUIDType, args[0])
				}

				if !args[1].Equals(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[1])
				}

				return types.UUIDType, nil
			},
			PGName: "uuid_generate_v5",
		},
		// array functions
		"array_append": {
			Args: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				if !args[0].IsArray {
					return nil, fmt.Errorf("expected first argument to be an array, got %s", args[0].String())
				}

				if args[1].IsArray {
					return nil, fmt.Errorf("expected second argument to be a scalar, got %s", args[1].String())
				}

				if !strings.EqualFold(args[0].Name, args[1].Name) {
					return nil, fmt.Errorf("expected both arguments to be of the same base type, got %s and %s", args[0].Name, args[1].Name)
				}

				return args[0], nil
			},
			PGName: "array_append",
		},
		"array_cat": {
			Args: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				if !args[0].IsArray {
					return nil, fmt.Errorf("expected first argument to be an array, got %s", args[0].String())
				}

				if !args[1].IsArray {
					return nil, fmt.Errorf("expected second argument to be an array, got %s", args[1].String())
				}

				if !strings.EqualFold(args[0].Name, args[1].Name) {
					return nil, fmt.Errorf("expected both arguments to be of the same base type, got %s and %s", args[0].Name, args[1].Name)
				}

				return args[0], nil
			},
			PGName: "array_cat",
		},
		"array_length": {
			Args: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].IsArray {
					return nil, fmt.Errorf("expected argument to be an array, got %s", args[0].String())
				}

				return types.IntType, nil
			},
			PGName: "array_length",
		},
		// Aggregate functions
		"count": {
			Args: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) > 1 {
					return nil, fmt.Errorf("invalid number of arguments: expected at most 1, got %d", len(args))
				}

				return types.IntType, nil
			},
			IsAggregate: true,
			PGName:      "count",
		},
		"sum": {
			Args: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].Equals(types.IntType) {
					return nil, wrapErrArgumentType(types.IntType, args[0])
				}

				return types.IntType, nil
			},
			IsAggregate: true,
			PGName:      "sum",
		},
	}
)

type FunctionDefinition struct {
	// Args is a function that checks the arguments passed to the function.
	// It can check the argument type and amount of arguments.
	// It returns the expected return type based on the arguments.
	Args func(args []*types.DataType) (*types.DataType, error)
	// IsAggregate is true if the function is an aggregate function.
	IsAggregate bool
	// PGName is the name of the function in Postgres.
	PGName string
}

var (
	// ErrInvalidArgumentNumber is returned when the number of arguments passed to a function is invalid.
	ErrInvalidArgumentNumber = errors.New("invalid number of arguments")
	// ErrInvalidArgumentType is returned when the type of an argument passed to a function is invalid.
	ErrInvalidArgumentType = errors.New("invalid argument type")
)

func wrapErrArgumentNumber(expected, got int) error {
	return fmt.Errorf("%w: expected %d, got %d", ErrInvalidArgumentNumber, expected, got)
}

func wrapErrArgumentType(expected, got *types.DataType) error {
	return fmt.Errorf("%w: expected %s, got %s", ErrInvalidArgumentType, expected.String(), got.String())
}
