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

				if !args[0].EqualsStrict(types.IntType) && args[0].Name != types.DecimalStr {
					return nil, fmt.Errorf("expected argument to be int or decimal, got %s", args[0].String())
				}

				return args[0], nil
			},
			PGFormat: defaultFormat("abs"),
		},
		"error": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				// technically error returns nothing, but for backwards compatibility with SELECT CASE we return null.
				// It doesn't really matter, since error will cancel execution anyways.
				return types.NullType, nil
			},
			PGFormat: defaultFormat("error"),
		},
		"parse_unix_timestamp": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// two args, both text
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				if !args[1].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[1])
				}

				return decimal16_6, nil
			},
			PGFormat: defaultFormat("parse_unix_timestamp"),
		},
		"format_unix_timestamp": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// first arg must be decimal(16, 6), second arg must be text
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				if !args[0].EqualsStrict(decimal16_6) {
					return nil, wrapErrArgumentType(decimal16_6, args[0])
				}

				if !args[1].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[1])
				}

				return types.TextType, nil
			},
			PGFormat: defaultFormat("format_unix_timestamp"),
		},
		"notice": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				// technically error returns nothing, but for backwards compatibility with SELECT CASE we return null.
				// It doesn't really matter, since error will cancel execution anyways.
				return types.NullType, nil
			},
			PGFormat: func(inputs []string, distinct, star bool) (string, error) {
				if star {
					return "", errStar("notice")
				}
				if distinct {
					return "", errDistinct("notice")
				}

				// TODO: this is implicitly coupled to internal/engine/generate, and should be moved there.
				// we can only move this there once we move all PGFormat, which will also be affected by
				// v0.9 changes, so leaving it here for now.
				return fmt.Sprintf("notice('txid:' || current_setting('ctx.txid') || ' ' || %s)", inputs[0]), nil
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
			PGFormat: defaultFormat("uuid_generate_v5"),
		},
		"encode": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// first must be blob, second must be text
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				if !args[0].EqualsStrict(types.BlobType) {
					return nil, wrapErrArgumentType(types.BlobType, args[0])
				}

				if !args[1].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[1])
				}

				return types.TextType, nil
			},
			PGFormat: defaultFormat("encode"),
		},
		"decode": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// first must be text, second must be text
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				if !args[1].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[1])
				}

				return types.BlobType, nil
			},
			PGFormat: defaultFormat("decode"),
		},
		"digest": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// first must be either text or blob, second must be text
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) && !args[0].EqualsStrict(types.BlobType) {
					return nil, fmt.Errorf("expected first argument to be text or blob, got %s", args[0].String())
				}

				if !args[1].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[1])
				}

				return types.BlobType, nil
			},
			PGFormat: defaultFormat("digest"),
		},
		"generate_dbid": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// first should be text, second should be blob
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				if !args[1].EqualsStrict(types.BlobType) {
					return nil, wrapErrArgumentType(types.BlobType, args[1])
				}

				return types.TextType, nil
			},
			PGFormat: func(inputs []string, distinct, star bool) (string, error) {
				if star {
					return "", errStar("generate_dbid")
				}

				if distinct {
					return "", errDistinct("generate_dbid")
				}

				return fmt.Sprintf(`(select 'x' || encode(sha224(lower(%s)::bytea || %s), 'hex'))`, inputs[0], inputs[1]), nil
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
			PGFormat: defaultFormat("array_append"),
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
			PGFormat: defaultFormat("array_prepend"),
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
			PGFormat: defaultFormat("array_cat"),
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
		"array_remove": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
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
					return nil, fmt.Errorf("remove type must be equal to scalar array type: array type: %s remove type: %s", args[0].Name, args[1].Name)
				}

				return args[0], nil
			},
			PGFormat: defaultFormat("array_remove"),
		},
		// string functions
		// the main SQL string functions defined here: https://www.postgresql.org/docs/16.1/functions-string.html
		"bit_length": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.IntType, nil
			},
			PGFormat: defaultFormat("bit_length"),
		},
		"char_length": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.IntType, nil
			},
			PGFormat: defaultFormat("char_length"),
		},
		"character_length": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.IntType, nil
			},
			PGFormat: defaultFormat("character_length"),
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
			PGFormat: defaultFormat("length"),
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
			PGFormat: defaultFormat("lower"),
		},
		"lpad": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				//can have 2-3 args. 1 and 3 must be text, 2 must be int
				if len(args) < 2 || len(args) > 3 {
					return nil, fmt.Errorf("invalid number of arguments: expected 2 or 3, got %d", len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				if !args[1].EqualsStrict(types.IntType) {
					return nil, wrapErrArgumentType(types.IntType, args[1])
				}

				if len(args) == 3 && !args[2].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[2])
				}

				return types.TextType, nil
			},
			PGFormat: func(inputs []string, distinct, star bool) (string, error) {
				// we need a custom function to type cast big int to int
				if distinct {
					return "", errDistinct("lpad")
				}

				if star {
					return "", errStar("lpad")
				}

				str := strings.Builder{}
				str.WriteString("lpad(")
				str.WriteString(inputs[0])
				str.WriteString(", ")
				str.WriteString(inputs[1])
				str.WriteString("::INT4")
				if len(inputs) == 3 {
					str.WriteString(", ")
					str.WriteString(inputs[2])
				}
				str.WriteString(")")

				return str.String(), nil
			},
		},
		"ltrim": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				//can have 1 or 2 args. both must be text
				if len(args) < 1 || len(args) > 2 {
					return nil, fmt.Errorf("invalid number of arguments: expected 1 or 2, got %d", len(args))
				}

				for _, arg := range args {
					if !arg.EqualsStrict(types.TextType) {
						return nil, wrapErrArgumentType(types.TextType, arg)
					}
				}

				return types.TextType, nil
			},
			PGFormat: defaultFormat("ltrim"),
		},
		"octet_length": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.IntType, nil
			},
			PGFormat: defaultFormat("octet_length"),
		},
		"overlay": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// 3-4 arguments. 1 and 2 must be text, 3 must be int, 4 must be int
				if len(args) < 3 || len(args) > 4 {
					return nil, fmt.Errorf("invalid number of arguments: expected 3 or 4, got %d", len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				if !args[1].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[1])
				}

				if !args[2].EqualsStrict(types.IntType) {
					return nil, wrapErrArgumentType(types.IntType, args[2])
				}

				if len(args) == 4 && !args[3].EqualsStrict(types.IntType) {
					return nil, wrapErrArgumentType(types.IntType, args[3])
				}

				return types.TextType, nil
			},
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if distinct {
					return "", errDistinct("overlay")
				}

				if star {
					return "", errStar("overlay")
				}

				str := strings.Builder{}
				str.WriteString("overlay(")
				str.WriteString(inputs[0])
				str.WriteString(" placing ")
				str.WriteString(inputs[1])
				str.WriteString(" from ")
				str.WriteString(inputs[2])
				str.WriteString("::INT4")
				if len(inputs) == 4 {
					str.WriteString(" for ")
					str.WriteString(inputs[3])
					str.WriteString("::INT4")
				}
				str.WriteString(")")

				return str.String(), nil
			},
		},
		"position": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// 2 arguments. both must be text
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				for _, arg := range args {
					if !arg.EqualsStrict(types.TextType) {
						return nil, wrapErrArgumentType(types.TextType, arg)
					}
				}

				return types.IntType, nil
			},
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if distinct {
					return "", errDistinct("position")
				}

				if star {
					return "", errStar("position")
				}

				return fmt.Sprintf("position(%s in %s)", inputs[0], inputs[1]), nil
			},
		},
		"rpad": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// 2-3 args, 1 and 3 must be text, 2 must be int
				if len(args) < 2 || len(args) > 3 {
					return nil, fmt.Errorf("invalid number of arguments: expected 2 or 3, got %d", len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				if !args[1].EqualsStrict(types.IntType) {
					return nil, wrapErrArgumentType(types.IntType, args[1])
				}

				if len(args) == 3 && !args[2].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[2])
				}

				return types.TextType, nil
			},
			PGFormat: func(inputs []string, distinct, star bool) (string, error) {
				// we need a custom function to type cast big int to int
				if distinct {
					return "", errDistinct("lpad")
				}

				if star {
					return "", errStar("lpad")
				}

				str := strings.Builder{}
				str.WriteString("rpad(")
				str.WriteString(inputs[0])
				str.WriteString(", ")
				str.WriteString(inputs[1])
				str.WriteString("::INT4")
				if len(inputs) == 3 {
					str.WriteString(", ")
					str.WriteString(inputs[2])
				}
				str.WriteString(")")

				return str.String(), nil
			},
		},
		"rtrim": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// 1-2 args, both must be text
				if len(args) < 1 || len(args) > 2 {
					return nil, fmt.Errorf("invalid number of arguments: expected 1 or 2, got %d", len(args))
				}

				for _, arg := range args {
					if !arg.EqualsStrict(types.TextType) {
						return nil, wrapErrArgumentType(types.TextType, arg)
					}
				}

				return types.TextType, nil
			},
			PGFormat: defaultFormat("rtrim"),
		},
		"substring": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// 2-3 args, 1 must be text, 2 and 3 must be int
				// Postgres supports several different usages of substring, however Kwil only supports 1.
				// In Postgres, substring can be used to both impose a string over a range, or to perform
				// regex matching. Kwil only supports the former, as regex matching is not supported.
				// Therefore, the second and third arguments must be integers.
				if len(args) < 2 || len(args) > 3 {
					return nil, fmt.Errorf("invalid number of arguments: expected 2 or 3, got %d", len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				if !args[1].EqualsStrict(types.IntType) {
					return nil, wrapErrArgumentType(types.IntType, args[1])
				}

				if len(args) == 3 && !args[2].EqualsStrict(types.IntType) {
					return nil, wrapErrArgumentType(types.IntType, args[2])
				}

				return types.TextType, nil
			},
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if distinct {
					return "", errDistinct("substring")
				}

				if star {
					return "", errStar("substring")
				}

				str := strings.Builder{}
				str.WriteString("substring(")
				str.WriteString(inputs[0])
				str.WriteString(" from ")
				str.WriteString(inputs[1])
				str.WriteString("::INT4")
				if len(inputs) == 3 {
					str.WriteString(" for ")
					str.WriteString(inputs[2])
					str.WriteString("::INT4")
				}
				str.WriteString(")")

				return str.String(), nil
			},
		},
		"trim": { // kwil only supports trim both
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// 1-2 args, both must be text
				if len(args) < 1 || len(args) > 2 {
					return nil, fmt.Errorf("invalid number of arguments: expected 1 or 2, got %d", len(args))
				}

				for _, arg := range args {
					if !arg.EqualsStrict(types.TextType) {
						return nil, wrapErrArgumentType(types.TextType, arg)
					}
				}

				return types.TextType, nil
			},
			PGFormat: defaultFormat("trim"),
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
			PGFormat: defaultFormat("upper"),
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
			PGFormat: defaultFormat("format"),
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
				// per https://www.postgresql.org/docs/current/datatype-numeric.html#DATATYPE-NUMERIC-TABLE
				// the result of sum will be made a decimal(1000, 0)
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].IsNumeric() {
					return nil, fmt.Errorf("expected argument to be numeric, got %s", args[0].String())
				}

				// we check if it is an unknown type before the switch,
				// as unknown will be true for all EqualsStrict checks
				if args[0] == types.UnknownType {
					return types.UnknownType, nil
				}

				var retType *types.DataType
				switch {
				case args[0].EqualsStrict(types.IntType):
					retType = decimal1000.Copy()
				case args[0].Name == types.DecimalStr:
					retType = args[0].Copy()
					retType.Metadata[0] = 1000 // max precision
				case args[0].EqualsStrict(types.Uint256Type):
					retType = decimal1000.Copy()
				default:
					panic(fmt.Sprintf("unexpected numeric type: %s", retType.String()))
				}

				return retType, nil
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
		"min": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// as per postgres docs, min can take any numeric or string type: https://www.postgresql.org/docs/8.0/functions-aggregate.html
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].IsNumeric() && !args[0].EqualsStrict(types.TextType) {
					return nil, fmt.Errorf("expected argument to be numeric or text, got %s", args[0].String())
				}

				return args[0], nil
			},
			IsAggregate: true,
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "", errStar("min")
				}
				if distinct {
					return "min(DISTINCT %s)", nil
				}

				return fmt.Sprintf("min(%s)", inputs[0]), nil
			},
		},
		"max": {
			ValidateArgs: func(args []*types.DataType) (*types.DataType, error) {
				// as per postgres docs, max can take any numeric or string type: https://www.postgresql.org/docs/8.0/functions-aggregate.html
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].IsNumeric() && !args[0].EqualsStrict(types.TextType) {
					return nil, fmt.Errorf("expected argument to be numeric or text, got %s", args[0].String())
				}

				return args[0], nil
			},
			IsAggregate: true,
			PGFormat: func(inputs []string, distinct bool, star bool) (string, error) {
				if star {
					return "", errStar("max")
				}
				if distinct {
					return "max(DISTINCT %s)", nil
				}

				return fmt.Sprintf("max(%s)", inputs[0]), nil
			},
		},
	}
)

// defaultFormat is the default PGFormat function for functions that do not have a custom one.
func defaultFormat(name string) FormatFunc {
	return func(inputs []string, distinct bool, star bool) (string, error) {
		if star {
			return "", errStar(name)
		}
		if distinct {
			return "", errDistinct(name)
		}

		return fmt.Sprintf("%s(%s)", name, strings.Join(inputs, ", ")), nil
	}
}

var (
	// decimal1000 is a decimal type with a precision of 1000.
	decimal1000 *types.DataType
	// decimal16_6 is a decimal type with a precision of 16 and a scale of 6.
	// it is used to represent UNIX timestamps, allowing microsecond precision.
	// see internal/sql/pg/sql.go/sqlCreateParseUnixTimestampFunc for more info
	decimal16_6 *types.DataType
)

func init() {
	var err error
	decimal1000, err = types.NewDecimalType(1000, 0)
	if err != nil {
		panic(fmt.Sprintf("failed to create decimal type: 1000, 0: %v", err))
	}

	decimal16_6, err = types.NewDecimalType(16, 6)
	if err != nil {
		panic(fmt.Sprintf("failed to create decimal type: 16, 6: %v", err))
	}
}

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
	// TODO: PGFormat is related to the plpgsql generation, and therefore should be moved to
	// internal/engine/generate. There is some implicit coupling here.
}

// FormatFunc is a function that formats a string of inputs for a SQL function.
type FormatFunc func(inputs []string, distinct bool, star bool) (string, error)

func wrapErrArgumentNumber(expected, got int) error {
	return fmt.Errorf("%w: expected %d, got %d", ErrFunctionSignature, expected, got)
}

func wrapErrArgumentType(expected, got *types.DataType) error {
	return fmt.Errorf("%w: expected %s, got %s", ErrType, expected.String(), got.String())
}

// ParseNotice parses a log raised from a notice() function.
// It returns an error if the log is not in the expected format.
func ParseNotice(log string) (txID string, notice string, err error) {
	_, after, found := strings.Cut(log, "txid:")
	if !found {
		return "", "", fmt.Errorf("notice log does not contain txid prefix: %s", log)
	}

	parts := strings.SplitN(after, " ", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("notice log does not contain txid and notice separated by space: %s", log)
	}

	return parts[0], parts[1], nil
}
