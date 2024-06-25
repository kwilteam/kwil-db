package execution

import (
	"maps"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_OrderAndClean(t *testing.T) {
	type testcase struct {
		name   string
		values map[string]any
		keys   []string
		res    []any
	}

	tests := []testcase{
		{
			name: "using $",
			values: map[string]any{
				"$key1": "value1",
				"$key2": []byte("value2"),
			},
			keys: []string{"$key1", "$key2"},
			res:  []any{"value1", []byte("value2")},
		},
		{
			name: "using $ and without $",
			values: map[string]any{
				"$key1": "value1",
				"key2":  []byte("value2"),
			},
			keys: []string{"$key1", "$key2"},
			res:  []any{"value1", []byte("value2")},
		},
		{
			name: "missing key",
			values: map[string]any{
				"$key1": "value1",
				"key2":  []byte("value2"),
			},
			keys: []string{"$key1", "$key3"},
			res:  []any{"value1", nil},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// copy the values to check that the function does not modify the input
			oldVals := maps.Clone(test.values)

			res := orderAndCleanValueMap(test.values, test.keys)
			require.EqualValues(t, test.res, res)

			require.EqualValues(t, oldVals, test.values)
		})
	}
}
