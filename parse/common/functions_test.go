package common

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/stretchr/testify/require"
)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := make([]Value, len(tt.input))

			var err error
			for i, in := range tt.input {
				vars[i], err = NewVariable(in)
				require.NoError(t, err)
			}

			fn, ok := Functions[tt.functionName]
			require.True(t, ok)

			scalar, ok := fn.(*ScalarFunctionDefinition)
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
