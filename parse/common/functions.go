package common

import (
	"crypto"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/core/utils"
)

var (
	Functions = map[string]FunctionDefinition{
		"abs": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.IntType) && args[0].Name != types.DecimalStr {
					return nil, fmt.Errorf("expected argument to be int or decimal, got %s", args[0].String())
				}

				return args[0], nil
			},
			PGFormatFunc: defaultFormat("abs"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				switch arg := args[0].(type) {
				case *IntValue:
					if arg.Val < 0 {
						return &IntValue{Val: -arg.Val}, nil
					}
					return arg, nil
				case *DecimalValue:
					if arg.Dec.Sign() < 0 {
						arg2 := arg.Dec.Copy()
						err := arg2.Neg()
						if err != nil {
							return nil, err
						}
						return &DecimalValue{Dec: arg2}, nil
					}
					return arg, nil
				}

				return nil, fmt.Errorf("unexpected type %T in abs", args[0])
			},
		},
		"error": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: defaultFormat("error"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("error function expects 1 argument, got %d", len(args))
				}

				text, ok := args[0].(*TextValue)
				if !ok {
					return nil, fmt.Errorf("error function expects a text argument, got %T", args[0])
				}

				return nil, fmt.Errorf("%s", text.Val)
			},
		},
		"parse_unix_timestamp": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: defaultFormat("parse_unix_timestamp"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				// Kwil's parseTimestamp takes a timestamp and a format string
				// The first arg is the timestamp, the second arg is the format string
				res, err := parseTimestamp(args[1].Value().(string), args[0].Value().(string))
				if err != nil {
					return nil, err
				}

				// we now need to convert the unix timestamp to a decimal(16, 6)
				// We start with 22,6 since the current int64 is in microseconds (16 digits).
				// We make this a decimal(22, 6), and then divide by 10^6 to get a decimal(16, 6)
				dec16, err := decimal.NewExplicit(fmt.Sprintf("%d", res), 22, 6)
				if err != nil {
					return nil, err
				}

				dec16, err = dec16.Div(dec16, dec10ToThe6th)
				if err != nil {
					return nil, err
				}

				err = dec16.SetPrecisionAndScale(16, 6)

				return &DecimalValue{Dec: dec16}, err
			},
		},
		"format_unix_timestamp": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: defaultFormat("format_unix_timestamp"),
			EvaluateFunc: func(spender Interpreter, args []Value) (Value, error) {
				// the inverse of parse_unix_timestamp, we need to convert a decimal(16, 6) to a unix timestamp
				// by multiplying by 10^6 and converting to an int64
				dec := args[0].(*DecimalValue).Dec

				err := dec.SetPrecisionAndScale(22, 6)
				if err != nil {
					return nil, err
				}

				dec, err = dec.Mul(dec, dec10ToThe6th)
				if err != nil {
					return nil, err
				}

				i64Microseconds, err := dec.Int64()
				if err != nil {
					return nil, err
				}

				ts := formatUnixMicro(i64Microseconds, args[1].Value().(string))
				return &TextValue{Val: ts}, nil
			},
		},
		"notice": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: func(inputs []string) (string, error) {
				// TODO: this is implicitly coupled to internal/engine/generate, and should be moved there.
				// we can only move this there once we move all PGFormat, which will also be affected by
				// v0.9 changes, so leaving it here for now.
				return fmt.Sprintf("notice('txid:' || current_setting('ctx.txid') || ' ' || %s)", inputs[0]), nil
			},
			EvaluateFunc: func(i Interpreter, args []Value) (Value, error) {
				i.Notice(args[0].Value().(string))
				return &NullValue{
					DataType: types.NullType.Copy(),
				}, nil
			},
		},
		"uuid_generate_v5": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: defaultFormat("uuid_generate_v5"),
			EvaluateFunc: func(interp Interpreter, args []Value) (Value, error) {
				// uuidv5 uses sha1 to hash the text input
				u := types.NewUUIDV5WithNamespace(types.UUID(args[0].(*UUIDValue).Val), []byte(args[1].(*TextValue).Val))
				return &UUIDValue{Val: u}, nil
			},
		},
		"encode": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: defaultFormat("encode"),
			EvaluateFunc: func(interp Interpreter, args []Value) (Value, error) {
				// postgres supports hex, base64, and escape.
				// we won't support escape.
				switch args[1].(*TextValue).Val {
				case "hex":
					return &TextValue{Val: hex.EncodeToString(args[0].Value().([]byte))}, nil
				case "base64":
					return &TextValue{Val: base64.StdEncoding.EncodeToString(args[0].Value().([]byte))}, nil
				case "escape":
					return nil, fmt.Errorf("procedures do not support escape encoding")
				default:
					return nil, fmt.Errorf("unknown encoding: %s", args[1].Value())
				}
			},
		},
		"decode": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: defaultFormat("decode"),
			EvaluateFunc: func(interp Interpreter, args []Value) (Value, error) {
				// postgres supports hex and base64.
				// we won't support escape.
				switch args[1].(*TextValue).Val {
				case "hex":
					b, err := hex.DecodeString(args[0].Value().(string))
					if err != nil {
						return nil, err
					}
					return &BlobValue{Val: b}, nil
				case "base64":
					b, err := base64.StdEncoding.DecodeString(args[0].Value().(string))
					if err != nil {
						return nil, err
					}
					return &BlobValue{Val: b}, nil
				case "escape":
					return nil, fmt.Errorf("procedures do not support escape encoding")
				default:
					return nil, fmt.Errorf("unknown encoding: %s", args[1].Value())
				}
			},
		},
		"digest": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: defaultFormat("digest"),
			EvaluateFunc: func(interp Interpreter, args []Value) (Value, error) {
				// supports md5, sha1, sha224, sha256, sha384 and sha512
				switch args[1].(*TextValue).Val {
				case "md5":
					return &BlobValue{Val: md5.New().Sum([]byte(args[0].Value().(string)))}, nil
				case "sha1":
					return &BlobValue{Val: sha1.New().Sum([]byte(args[0].Value().(string)))}, nil
				case "sha224":
					return &BlobValue{Val: crypto.SHA224.New().Sum([]byte(args[0].Value().(string)))}, nil
				case "sha256":
					return &BlobValue{Val: crypto.SHA256.New().Sum([]byte(args[0].Value().(string)))}, nil
				case "sha384":
					return &BlobValue{Val: crypto.SHA384.New().Sum([]byte(args[0].Value().(string)))}, nil
				case "sha512":
					return &BlobValue{Val: crypto.SHA512.New().Sum([]byte(args[0].Value().(string)))}, nil
				default:
					return nil, fmt.Errorf("unknown digest: %s", args[1].Value())
				}
			},
		},
		"generate_dbid": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: func(inputs []string) (string, error) {
				return fmt.Sprintf(`(select 'x' || encode(sha224(lower(%s)::bytea || %s), 'hex'))`, inputs[0], inputs[1]), nil
			},
			EvaluateFunc: func(interp Interpreter, args []Value) (Value, error) {
				return &TextValue{Val: utils.GenerateDBID(args[0].Value().(string), args[1].Value().([]byte))}, nil
			},
		},
		// array functions
		"array_append": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
					return nil, fmt.Errorf("append type must be equal to scalar array type: array type: %s append type: %s", args[0].Name, args[1].Name)
				}

				return args[0], nil
			},
			PGFormatFunc: defaultFormat("array_append"),
			EvaluateFunc: func(interp Interpreter, args []Value) (Value, error) {
				arr := args[0].(ArrayValue)
				// all Kuneiform arrays are 1-indexed
				err := arr.Set(int64(arr.Len()+1), args[1].(ScalarValue))
				return arr, err
			},
		},
		"array_prepend": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 2 {
					return nil, wrapErrArgumentNumber(2, len(args))
				}

				if args[0].IsArray {
					return nil, fmt.Errorf("expected first argument to be a scalar, got %s", args[0].String())
				}

				if !args[1].IsArray {
					return nil, fmt.Errorf("expected second argument to be an array, got %s", args[1].String())
				}

				if !strings.EqualFold(args[0].Name, args[1].Name) {
					return nil, fmt.Errorf("prepend type must be equal to scalar array type: array type: %s prepend type: %s", args[1].Name, args[0].Name)
				}

				return args[1], nil
			},
			PGFormatFunc: defaultFormat("array_prepend"),
			EvaluateFunc: func(interp Interpreter, args []Value) (Value, error) {
				scal := args[0].(ScalarValue)
				arr := args[1].(ArrayValue)

				var scalars []ScalarValue
				// 1-indexed
				for i := 1; i <= arr.Len(); i++ {
					newScal, err := arr.Index(int64(i))
					if err != nil {
						return nil, err
					}
					scalars = append(scalars, newScal)
				}

				return scal.Array(scalars...)
			},
		},
		"array_cat": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
					return nil, fmt.Errorf("expected both arrays to be of the same scalar type, got %s and %s", args[0].Name, args[1].Name)
				}

				return args[0], nil
			},
			PGFormatFunc: defaultFormat("array_cat"),
			EvaluateFunc: func(interp Interpreter, args []Value) (Value, error) {
				arr1 := args[0].(ArrayValue)
				arr2 := args[1].(ArrayValue)

				startIdx := arr1.Len()
				for i := 1; i <= arr2.Len(); i++ {
					newScal, err := arr2.Index(int64(i))
					if err != nil {
						return nil, err
					}
					err = arr1.Set(int64(startIdx+i), newScal)
					if err != nil {
						return nil, err
					}
				}

				return arr1, nil
			},
		},
		"array_length": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].IsArray {
					return nil, fmt.Errorf("expected argument to be an array, got %s", args[0].String())
				}

				return types.IntType, nil
			},
			PGFormatFunc: func(inputs []string) (string, error) {
				return fmt.Sprintf("array_length(%s, 1)", inputs[0]), nil
			},
			EvaluateFunc: func(interp Interpreter, args []Value) (Value, error) {
				arr := args[0].(ArrayValue)
				return &IntValue{Val: int64(arr.Len())}, nil
			},
		},
		// string functions
		// the main SQL string functions defined here: https://www.postgresql.org/docs/16.1/functions-string.html
		"bit_length": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.IntType, nil
			},
			PGFormatFunc: defaultFormat("bit_length"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				text := args[0].(*TextValue).Val
				return &IntValue{Val: int64(len(text) * 8)}, nil
			},
		},
		"char_length": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.IntType, nil
			},
			PGFormatFunc: defaultFormat("char_length"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				text := args[0].(*TextValue).Val
				return &IntValue{Val: int64(utf8.RuneCountInString(text))}, nil
			},
		},
		"character_length": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.IntType, nil
			},
			PGFormatFunc: defaultFormat("character_length"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				text := args[0].(*TextValue).Val
				return &IntValue{Val: int64(utf8.RuneCountInString(text))}, nil
			},
		},
		"length": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.IntType, nil
			},
			PGFormatFunc: defaultFormat("length"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				text := args[0].(*TextValue).Val
				return &IntValue{Val: int64(utf8.RuneCountInString(text))}, nil
			},
		},
		"lower": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.TextType, nil
			},
			PGFormatFunc: defaultFormat("lower"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				text := args[0].(*TextValue).Val
				return &TextValue{Val: strings.ToLower(text)}, nil
			},
		},
		"lpad": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: defaultFormat("lpad"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				text := args[0].(*TextValue).Val
				length := args[1].(*IntValue).Val
				padStr := " "
				if len(args) == 3 {
					padStr = args[2].(*TextValue).Val
				}

				return &TextValue{Val: pad(text, int(length), padStr, true)}, nil
			},
		},
		"ltrim": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: defaultFormat("ltrim"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				text := args[0].(*TextValue).Val
				chars := " "
				if len(args) == 2 {
					chars = args[1].(*TextValue).Val
				}
				return &TextValue{Val: strings.TrimLeft(text, chars)}, nil
			},
		},
		"octet_length": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.IntType, nil
			},
			PGFormatFunc: defaultFormat("octet_length"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				text := args[0].(*TextValue).Val
				return &IntValue{Val: int64(len(text))}, nil
			},
		},
		"overlay": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: func(inputs []string) (string, error) {
				str := strings.Builder{}
				str.WriteString("overlay(")
				str.WriteString(inputs[0])
				str.WriteString(" placing ")
				str.WriteString(inputs[1])
				str.WriteString(" from ")
				str.WriteString(inputs[2])
				if len(inputs) == 4 {
					str.WriteString(" for ")
					str.WriteString(inputs[3])
				}
				str.WriteString(")")

				return str.String(), nil
			},
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				input := args[0].(*TextValue).Val
				replace := args[1].(*TextValue).Val
				start := args[2].(*IntValue).Val

				if start < 0 {
					return nil, ErrNegativeSubstringLength
				}

				length := int64(len(replace))
				if len(args) == 4 {
					length = args[3].(*IntValue).Val
				}

				return &TextValue{Val: overlay(input, replace, int(start), int(length))}, nil
			},
		},
		"position": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: func(inputs []string) (string, error) {
				return fmt.Sprintf("position(%s in %s)", inputs[0], inputs[1]), nil
			},
			EvaluateFunc: func(interp Interpreter, args []Value) (Value, error) {
				substr := args[0].(*TextValue).Val
				str := args[1].(*TextValue).Val

				pos := strings.Index(str, substr)

				var res int64
				if pos == -1 {
					res = 0
				} else {
					res = int64(utf8.RuneCountInString(str[:pos])) + 1
				}

				return &IntValue{Val: res}, nil
			},
		},
		"rpad": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: defaultFormat("rpad"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				text := args[0].(*TextValue).Val
				length := args[1].(*IntValue).Val
				padStr := " "
				if len(args) == 3 {
					padStr = args[2].(*TextValue).Val
				}

				return &TextValue{Val: pad(text, int(length), padStr, false)}, nil
			},
		},
		"rtrim": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: defaultFormat("rtrim"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				text := args[0].(*TextValue).Val
				chars := " "
				if len(args) == 2 {
					chars = args[1].(*TextValue).Val
				}
				return &TextValue{Val: strings.TrimRight(text, chars)}, nil
			},
		},
		"substring": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: func(inputs []string) (string, error) {
				str := strings.Builder{}
				str.WriteString("substring(")
				str.WriteString(inputs[0])
				str.WriteString(" from ")
				str.WriteString(inputs[1])
				if len(inputs) == 3 {
					str.WriteString(" for ")
					str.WriteString(inputs[2])
				}
				str.WriteString(")")

				return str.String(), nil
			},
			EvaluateFunc: func(_ Interpreter, args []Value) (v Value, err error) {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("panic: %v", r)
					}
				}()

				text := args[0].(*TextValue).Val
				start := args[1].(*IntValue).Val

				if start > int64(len(text)) {
					// not sure why Postgres does this, but it does.
					return &TextValue{Val: ""}, nil
				}

				length := int64(len(text))

				if len(args) == 3 {
					length = args[2].(*IntValue).Val
				}

				if length < 0 {
					return nil, ErrNegativeSubstringLength
				}

				runes := []rune(text)
				if start < 1 {
					// if start is negative, then we subtract the difference from 1
					// from the length. I don't know why Postgres does this, but it does.
					length -= 1 - start
					start = 1
				}
				if length < 0 {
					// if length is negative, then we set it to 0.
					// Not sure why Postgres does this, but it does.
					length = 0
				}
				end := min(int64(len(runes)), start-1+length)
				return &TextValue{Val: string(runes[start-1 : end])}, nil
			},
		},
		"trim": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: defaultFormat("trim"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				text := args[0].(*TextValue).Val
				chars := " "
				if len(args) == 2 {
					chars = args[1].(*TextValue).Val
				}
				return &TextValue{Val: strings.Trim(text, chars)}, nil
			},
		},
		"upper": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.TextType, nil
			},
			PGFormatFunc: defaultFormat("upper"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				text := args[0].(*TextValue).Val
				return &TextValue{Val: strings.ToUpper(text)}, nil
			},
		},
		"format": &ScalarFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("invalid number of arguments: expected at least 1, got %d", len(args))
				}

				if !args[0].EqualsStrict(types.TextType) {
					return nil, wrapErrArgumentType(types.TextType, args[0])
				}

				return types.TextType, nil
			},
			PGFormatFunc: defaultFormat("format"),
			EvaluateFunc: func(_ Interpreter, args []Value) (Value, error) {
				format := args[0].(*TextValue).Val

				values := []any{}
				for _, arg := range args[1:] {
					values = append(values, arg.Value())
				}

				return &TextValue{Val: positionalSprintf(format, values...)}, nil
			},
		},
		// Aggregate functions
		"count": &AggregateFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
				if len(args) > 1 {
					return nil, fmt.Errorf("invalid number of arguments: expected at most 1, got %d", len(args))
				}

				return types.IntType, nil
			},
			PGFormatFunc: func(inputs []string, distinct bool) (string, error) {
				if len(inputs) == 0 {
					if distinct {
						return "", fmt.Errorf("count(DISTINCT *) is not supported")
					}
					return "count(*)", nil
				}
				if distinct {
					return fmt.Sprintf("count(DISTINCT %s)", inputs[0]), nil
				}

				return fmt.Sprintf("count(%s)", inputs[0]), nil
			},
		},
		"sum": &AggregateFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
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
			PGFormatFunc: func(inputs []string, distinct bool) (string, error) {
				if distinct {
					return "sum(DISTINCT %s)", nil
				}

				return fmt.Sprintf("sum(%s)", inputs[0]), nil
			},
		},
		"min": &AggregateFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
				// as per postgres docs, min can take any numeric or string type: https://www.postgresql.org/docs/8.0/functions-aggregate.html
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].IsNumeric() && !args[0].EqualsStrict(types.TextType) {
					return nil, fmt.Errorf("expected argument to be numeric or text, got %s", args[0].String())
				}

				return args[0], nil
			},
			PGFormatFunc: func(inputs []string, distinct bool) (string, error) {
				if distinct {
					return "min(DISTINCT %s)", nil
				}

				return fmt.Sprintf("min(%s)", inputs[0]), nil
			},
		},
		"max": &AggregateFunctionDefinition{
			ValidateArgsFunc: func(args []*types.DataType) (*types.DataType, error) {
				// as per postgres docs, max can take any numeric or string type: https://www.postgresql.org/docs/8.0/functions-aggregate.html
				if len(args) != 1 {
					return nil, wrapErrArgumentNumber(1, len(args))
				}

				if !args[0].IsNumeric() && !args[0].EqualsStrict(types.TextType) {
					return nil, fmt.Errorf("expected argument to be numeric or text, got %s", args[0].String())
				}

				return args[0], nil
			},
			PGFormatFunc: func(inputs []string, distinct bool) (string, error) {
				if distinct {
					return "max(DISTINCT %s)", nil
				}

				return fmt.Sprintf("max(%s)", inputs[0]), nil
			},
		},
	}
)

// pad pads either side of a string. The side can be specified with the side parameter (left is true, right is false)
func pad(input string, length int, padStr string, side bool) string {
	inputLength := len(input)
	if inputLength >= length {
		return input[:length] // Truncate if the input string is longer than the desired length
	}

	padLength := len(padStr)
	if padLength == 0 {
		return input // If padStr is empty, return the input as is
	}

	// Calculate the number of times the padStr needs to be repeated
	repeatCount := (length - inputLength) / padLength
	remainder := (length - inputLength) % padLength

	// Build the left padding
	p := strings.Repeat(padStr, repeatCount) + padStr[:remainder]

	if side {
		return p + input
	}
	return input + p
}

// overlay function mimics the behavior of the PostgreSQL overlay function
func overlay(input, replace string, start, forInt int) string {
	if start < 1 {
		start = 1
	}

	// Convert start and length to rune-based indices
	startIndex := start - 1
	endIndex := startIndex + forInt

	// Get the slice indices in bytes
	inputRunes := []rune(input)
	replaceRunes := []rune(replace)

	// Adjust indices if they go beyond the string length
	if startIndex > len(inputRunes) {
		startIndex = len(inputRunes)
	}
	if endIndex > len(inputRunes) {
		endIndex = len(inputRunes)
	}

	// Replace the specified section of the string with the replacement string
	resultRunes := append(inputRunes[:startIndex], append(replaceRunes, inputRunes[endIndex:]...)...)
	return string(resultRunes)
}

// positionalSprintf is a version of fmt.Sprintf that supports positional arguments.
// It mimics Postgres's "format"
func positionalSprintf(format string, args ...interface{}) string {
	for i, arg := range args {
		placeholder := fmt.Sprintf("%%%d$s", i+1)
		format = strings.ReplaceAll(format, placeholder, fmt.Sprintf("%v", arg))
	}
	return fmt.Sprintf(format, args...)
}

// defaultFormat is the default PGFormat function for functions that do not have a custom one.
func defaultFormat(name string) func(inputs []string) (string, error) {
	return func(inputs []string) (string, error) {
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
	// dec10ToThe6th is 10^6
	dec10ToThe6th *decimal.Decimal
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

	dec10ToThe6th, err = decimal.NewFromString("1000000")
	if err != nil {
		panic(fmt.Sprintf("failed to create decimal type: 10^6: %v", err))
	}
}

// FunctionDefinition if a definition of a function.
// It has two implementations: ScalarFuncDef and AggregateFuncDef.
type FunctionDefinition interface {
	// ValidateArgs is a function that checks the arguments passed to the function.
	// It can check the argument type and amount of arguments.
	// It returns the expected return type based on the arguments.
	ValidateArgs(args []*types.DataType) (*types.DataType, error)
	funcdef()
}

// ScalarFunctionDefinition is a definition of a scalar function.
type ScalarFunctionDefinition struct {
	ValidateArgsFunc func(args []*types.DataType) (*types.DataType, error)
	PGFormatFunc     func(inputs []string) (string, error)
	EvaluateFunc     func(interp Interpreter, args []Value) (Value, error)
}

func (s *ScalarFunctionDefinition) ValidateArgs(args []*types.DataType) (*types.DataType, error) {
	return s.ValidateArgsFunc(args)
}

func (s *ScalarFunctionDefinition) funcdef() {}

// AggregateFunctionDefinition is a definition of an aggregate function.
type AggregateFunctionDefinition struct {
	// ValidateArgs is a function that checks the arguments passed to the function.
	// It can check the argument type and amount of arguments.
	ValidateArgsFunc func(args []*types.DataType) (*types.DataType, error)
	// PGFormat is a function that formats the inputs to the function in Postgres format.
	// For example, the function `sum` would format the inputs as `sum($1)`.
	// It can also format the inputs with DISTINCT. If no inputs are given, it is a *.
	PGFormatFunc func(inputs []string, distinct bool) (string, error)
	// We currently don't need to evaluate aggregates since they are handled by the engine.
}

func (a *AggregateFunctionDefinition) ValidateArgs(args []*types.DataType) (*types.DataType, error) {
	return a.ValidateArgsFunc(args)
}

func (a *AggregateFunctionDefinition) funcdef() {}

// FormatFunc is a function that formats a string of inputs for a SQL function.
type FormatFunc func(inputs []string) (string, error)

func wrapErrArgumentNumber(expected, got int) error {
	return fmt.Errorf("expected %d, got %d", expected, got)
}

func wrapErrArgumentType(expected, got *types.DataType) error {
	return fmt.Errorf("expected %s, got %s", expected.String(), got.String())
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

// Interpreter allows functions to interact with the interpreter.
type Interpreter interface {
	Spend(amount int64) error
	Notice(notice string)
}
