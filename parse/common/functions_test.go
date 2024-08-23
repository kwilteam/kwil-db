package common_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/parse/common"
	"github.com/stretchr/testify/require"
)

// tests that we have implemented all functions
func Test_AllFunctionsImplemented(t *testing.T) {
	for name, fn := range common.Functions {
		scalar, ok := fn.(*common.ScalarFunctionDefinition)
		if ok {
			if scalar.EvaluateFunc == nil {
				t.Errorf("function %s has no EvaluateFunc", name)
			}
			if scalar.PGFormatFunc == nil {
				t.Errorf("function %s has no PGFormatFunc", name)
			}
			if scalar.ValidateArgsFunc == nil {
				t.Errorf("function %s has no ValidateArgsFunc", name)
			}
		} else {
			agg, ok := fn.(*common.AggregateFunctionDefinition)
			if !ok {
				t.Errorf("function %s is not a scalar or aggregate function", name)
			}
			if agg.PGFormatFunc == nil {
				t.Errorf("function %s has no PGFormatFunc", name)
			}
			if agg.ValidateArgsFunc == nil {
				t.Errorf("function %s has no ValidateArgsFunc", name)
			}
		}
	}
}

func Test_ScalarFunctions(t *testing.T) {
	type testcase struct {
		name         string
		functionName string
		input        []any // will be converted to []Value
		expected     any   // result of Value.Value()
		err          error
	}

	tests := []testcase{
		{
			name:         "parse_unix_timestamp",
			functionName: "parse_unix_timestamp",
			input:        []any{"2023-05-15 14:30:45", "YYYY-MM-DD HH:MI:SS"},
			expected:     mustDecimal("1684161045.000000"),
		},
		{
			name:         "format_unix_timestamp",
			functionName: "format_unix_timestamp",
			input:        []any{mustDecimal("1684161045.000000"), "YYYY-MM-DD HH:MI:SS"},
			expected:     "2023-05-15 14:30:45",
		},
		{
			// checking that this matches Postgres's.
			// select uuid_generate_v5('9ed26752-08bc-44c5-83ef-3c6df734c4e7'::uuid, 'a');
			// yields 32d43be3-8591-5849-946b-6f5268aef4ae
			name:         "uuid_v5",
			functionName: "uuid_generate_v5",
			input:        []any{mustUUID("9ed26752-08bc-44c5-83ef-3c6df734c4e7"), "a"},
			expected:     mustUUID("32d43be3-8591-5849-946b-6f5268aef4ae"),
		},
		{
			name:         "array_append",
			functionName: "array_append",
			input:        []any{[]int64{1, 2, 3}, int64(4)},
			expected:     []*int64{intRef(1), intRef(2), intRef(3), intRef(4)},
		},
		{
			name:         "array_prepend",
			functionName: "array_prepend",
			input:        []any{int64(1), []int64{2, 3, 4}},
			expected:     []*int64{intRef(1), intRef(2), intRef(3), intRef(4)},
		},
		{
			name:         "array_cat",
			functionName: "array_cat",
			input:        []any{[]int64{1, 2, 3}, []int64{4, 5, 6}},
			expected:     []*int64{intRef(1), intRef(2), intRef(3), intRef(4), intRef(5), intRef(6)},
		},
		{
			name:         "array_length",
			functionName: "array_length",
			input:        []any{[]int64{1, 2, 3}},
			expected:     int64(3),
		},
		{
			name:         "bit_length",
			functionName: "bit_length",
			input:        []any{"hello"},
			expected:     int64(40), // 5 characters * 8 bits
		},
		{
			name:         "char_length",
			functionName: "char_length",
			input:        []any{"hello世界"},
			expected:     int64(7),
		},
		{
			name:         "character_length",
			functionName: "character_length",
			input:        []any{"hello世界"},
			expected:     int64(7),
		},
		{
			name:         "length",
			functionName: "length",
			input:        []any{"hello世界"},
			expected:     int64(7),
		},
		{
			name:         "lower",
			functionName: "lower",
			input:        []any{"HeLLo"},
			expected:     "hello",
		},
		{
			name:         "lpad",
			functionName: "lpad",
			input:        []any{"hi", int64(5), "xy"},
			expected:     "xyxhi",
		},
		{
			name:         "lpad_with_space",
			functionName: "lpad",
			input:        []any{"hi", int64(4)},
			expected:     "  hi",
		},
		{
			name:         "lpad_longer_input",
			functionName: "lpad",
			input:        []any{"hello", int64(4), "xy"},
			expected:     "hell",
		},
		{
			name:         "lpad_empty_pad_string",
			functionName: "lpad",
			input:        []any{"hi", int64(5), ""},
			expected:     "hi",
		},
		{
			name:         "lpad_single_char_pad",
			functionName: "lpad",
			input:        []any{"hi", int64(5), "x"},
			expected:     "xxxhi",
		},
		{
			name:         "ltrim",
			functionName: "ltrim",
			input:        []any{"  hello  "},
			expected:     "hello  ",
		},
		{
			name:         "ltrim_with_chars",
			functionName: "ltrim",
			input:        []any{"xxhelloxx", "x"},
			expected:     "helloxx",
		},
		{
			name:         "octet_length",
			functionName: "octet_length",
			input:        []any{"hello世界"},
			expected:     int64(11),
		},
		{
			name:         "rpad",
			functionName: "rpad",
			input:        []any{"hi", int64(5), "xy"},
			expected:     "hixyx",
		},
		{
			name:         "rpad_with_space",
			functionName: "rpad",
			input:        []any{"hi", int64(4)},
			expected:     "hi  ",
		},
		{
			name:         "rtrim",
			functionName: "rtrim",
			input:        []any{"  hello  "},
			expected:     "  hello",
		},
		{
			name:         "rtrim_with_chars",
			functionName: "rtrim",
			input:        []any{"xxhelloxx", "x"},
			expected:     "xxhello",
		},
		{
			name:         "substring",
			functionName: "substring",
			input:        []any{"hello", int64(2), int64(3)},
			expected:     "ell",
		},
		{
			name:         "substring_without_length",
			functionName: "substring",
			input:        []any{"hello", int64(2)},
			expected:     "ello",
		},
		{
			name:         "substring_full_string",
			functionName: "substring",
			input:        []any{"hello", int64(1), int64(5)},
			expected:     "hello",
		},
		{
			name:         "substring_beyond_end",
			functionName: "substring",
			input:        []any{"hello", int64(2), int64(10)},
			expected:     "ello",
		},
		{
			name:         "substring_zero_length",
			functionName: "substring",
			input:        []any{"hello", int64(2), int64(0)},
			expected:     "",
		},
		{
			name:         "substring_negative_start",
			functionName: "substring",
			input:        []any{"hello", int64(-3), int64(2)},
			expected:     "",
		},
		{
			name:         "substring_negative_length",
			functionName: "substring",
			input:        []any{"hello1", int64(2), int64(-1)},
			err:          common.ErrNegativeSubstringLength,
		},
		{
			name:         "substring_start_beyond_end",
			functionName: "substring",
			input:        []any{"hello", int64(10), int64(2)},
			expected:     "",
		},
		{
			name:         "substring_unicode",
			functionName: "substring",
			input:        []any{"hello世界", int64(6), int64(2)},
			expected:     "世界",
		},
		{
			name:         "substring_unicode_partial",
			functionName: "substring",
			input:        []any{"hello世界", int64(7), int64(1)},
			expected:     "界",
		},
		{
			name:         "trim",
			functionName: "trim",
			input:        []any{"  hello  "},
			expected:     "hello",
		},
		{
			name:         "trim_with_chars",
			functionName: "trim",
			input:        []any{"xxhelloxx", "x"},
			expected:     "hello",
		},
		{
			name:         "upper",
			functionName: "upper",
			input:        []any{"HeLLo"},
			expected:     "HELLO",
		},
		{
			name:         "format",
			functionName: "format",
			input:        []any{"Hello %s, %1$s", "World"},
			expected:     "Hello World, World",
		},
		// Overlay tests
		{
			name:         "overlay_basic",
			functionName: "overlay",
			input:        []any{"Txxxxas", "hom", int64(2), int64(4)},
			expected:     "Thomas",
		},
		{
			name:         "overlay_without_length",
			functionName: "overlay",
			input:        []any{"Txxxxas", "hom", int64(2)},
			expected:     "Thomxas",
		},
		{
			name:         "overlay_beyond_end",
			functionName: "overlay",
			input:        []any{"Hello", "world", int64(6)},
			expected:     "Helloworld",
		},
		{
			name:         "overlay_at_start",
			functionName: "overlay",
			input:        []any{"ello", "H", int64(1)},
			expected:     "Hllo",
		},
		{
			name:         "overlay_empty_string",
			functionName: "overlay",
			input:        []any{"Hello", "", int64(2), int64(2)},
			expected:     "Hlo",
		},
		{
			name:         "overlay_entire_string",
			functionName: "overlay",
			input:        []any{"Hello", "World", int64(1), int64(5)},
			expected:     "World",
		},
		{
			name:         "overlay_zero_length",
			functionName: "overlay",
			input:        []any{"Hello", "x", int64(3), int64(0)},
			expected:     "Hexllo",
		},
		{
			name:         "overlay_negative_start",
			functionName: "overlay",
			input:        []any{"Hello", "x", int64(-1), int64(2)},
			err:          common.ErrNegativeSubstringLength,
		},
		{
			name:         "overlay_negative_length",
			functionName: "overlay",
			input:        []any{"Hello", "x", int64(2), int64(-1)},
			expected:     "HxHello",
		},
		{
			name:         "overlay_long_replacement",
			functionName: "overlay",
			input:        []any{"Hello", "Beautiful World", int64(2), int64(4)},
			expected:     "HBeautiful World",
		},
		{
			name:         "overlay_unicode_mixed",
			functionName: "overlay",
			input:        []any{"Hello 世界！", "こんにちは", int64(7), int64(2)},
			expected:     "Hello こんにちは！",
		},
		// Position tests
		{
			name:         "position_basic",
			functionName: "position",
			input:        []any{"hi", "hello world"},
			expected:     int64(0), // PostgreSQL returns 0 when substring is not found
		},
		{
			name:         "position_found",
			functionName: "position",
			input:        []any{"world", "hello world"},
			expected:     int64(7),
		},
		{
			name:         "position_at_start",
			functionName: "position",
			input:        []any{"hello", "hello world"},
			expected:     int64(1),
		},
		{
			name:         "position_at_end",
			functionName: "position",
			input:        []any{"world", "hello world"},
			expected:     int64(7),
		},
		{
			name:         "position_empty_substring",
			functionName: "position",
			input:        []any{"", "hello world"},
			expected:     int64(1), // PostgreSQL returns 1 for empty substring
		},
		{
			name:         "position_empty_string",
			functionName: "position",
			input:        []any{"hello", ""},
			expected:     int64(0),
		},
		{
			name:         "position_both_empty",
			functionName: "position",
			input:        []any{"", ""},
			expected:     int64(1),
		},
		{
			name:         "position_case_sensitive",
			functionName: "position",
			input:        []any{"WORLD", "hello world"},
			expected:     int64(0),
		},
		{
			name:         "position_multiple_occurrences",
			functionName: "position",
			input:        []any{"o", "hello world"},
			expected:     int64(5), // Returns the first occurrence
		},
		{
			name:         "position_unicode",
			functionName: "position",
			input:        []any{"世界", "你好世界"},
			expected:     int64(3),
		},
		{
			name:         "position_substring_longer",
			functionName: "position",
			input:        []any{"hello world plus", "hello world"},
			expected:     int64(0),
		},
		{
			name:         "position_special_chars",
			functionName: "position",
			input:        []any{"lo_", "hello_world"},
			expected:     int64(4),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := make([]common.Value, len(tt.input))

			var err error
			for i, in := range tt.input {
				vars[i], err = common.NewVariable(in)
				require.NoError(t, err)
			}

			fn, ok := common.Functions[tt.functionName]
			require.True(t, ok)

			scalar, ok := fn.(*common.ScalarFunctionDefinition)
			require.True(t, ok)

			res, err := scalar.EvaluateFunc(&mockInterpreter{}, vars)
			if tt.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
				require.EqualValues(t, tt.expected, res.Value())
			}
		})
	}
}

func mustDecimal(s string) *decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func mustUUID(s string) *types.UUID {
	u, err := types.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}

func intRef(i int64) *int64 {
	return &i
}

type mockInterpreter struct{}

func (m *mockInterpreter) Spend(_ int64) error {
	return nil
}

func (m *mockInterpreter) Notice(_ string) {
}
