package pg

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ArrayEncodeDecodeNULLs(t *testing.T) {
	arr := []string{"a", "b", "c", "NULL"}
	res, err := serializeArray(arr, 4, func(s string) ([]byte, error) {
		if s == "NULL" {
			return nil, nil
		}
		return []byte(s), nil
	})
	require.NoError(t, err)

	// use deserializePtrArray to handle NULL values
	res2, err := deserializePtrArray[string](res, 4, func(b []byte) (any, error) {
		if b == nil {
			return nil, nil
		}
		return string(b), nil
	})
	require.NoError(t, err)

	require.Equal(t, len(arr), len(res2))

	for i := range res2 {
		if res2[i] == nil {
			require.Equal(t, arr[i], "NULL")
			continue
		}

		require.Equal(t, arr[i], *res2[i])
	}
}

// don't use this in product code, it doesn't handle NULL values.
// this is only a test helper when not testing arrays with NULLs.
func deserializeTestArray[T any](buf []byte, lengthSize uint8, deserialize func([]byte) (any, error)) ([]T, error) {
	ptrs, err := deserializePtrArray[T](buf, lengthSize, deserialize)
	if err != nil {
		return nil, err
	}
	var vals []T
	for _, ptr := range ptrs {
		if ptr == nil {
			var vt T
			vals = append(vals, vt)
		} else {
			vals = append(vals, *ptr)
		}
	}
	return vals, nil
}

func Test_ArrayEncodeDecode(t *testing.T) {
	arr := []string{"a", "b", "c"}
	res, err := serializeArray(arr, 4, func(s string) ([]byte, error) {
		return []byte(s), nil
	})
	require.NoError(t, err)

	res2, err := deserializeTestArray[string](res, 4, func(b []byte) (any, error) {
		return string(b), nil
	})
	require.NoError(t, err)

	require.EqualValues(t, arr, res2)

	arr2 := []int64{1, 2, 3}
	res, err = serializeArray(arr2, 1, func(i int64) ([]byte, error) {
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(i))
		return buf, nil
	})
	require.NoError(t, err)

	res3, err := deserializeTestArray[int64](res, 1, func(b []byte) (any, error) {
		return int64(binary.LittleEndian.Uint64(b)), nil
	})
	require.NoError(t, err)

	require.EqualValues(t, arr2, res3)
}

func TestSplitString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "single value",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "simple values",
			input:    "a,b,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "quoted strings with commas",
			input:    `"hello,world",next,"another,value"`,
			expected: []string{`hello,world`, "next", `another,value`},
		},
		{
			name:     "escaped quotes",
			input:    `value1,"escaped\"quote",value3`,
			expected: []string{"value1", `escaped"quote`, "value3"},
		},
		{
			name:     "escaped backslashes",
			input:    `normal,with\\backslashes,"quoted\\with\\backslashes"`,
			expected: []string{"normal", `with\backslashes`, `quoted\with\backslashes`},
		},
		{
			name:     "trailing backslash",
			input:    `value1,value2\`,
			expected: []string{"value1", "value2\\"},
		},
		{
			name:     "mixed escapes and quotes",
			input:    `simple,"quoted,value",escaped\\comma\,,"quoted\"escape\\chars"`,
			expected: []string{"simple", "quoted,value", `escaped\comma,`, `quoted"escape\chars`},
		},
		{ // well formed arrays from pg should no have whitespace outside of quotes...
			name:     "whitespace handling",
			input:    ` spaced , "quoted space" ,nospace`,
			expected: []string{" spaced ", " quoted space ", "nospace"},
		},
		{
			name:     "empty elements",
			input:    "first,,last",
			expected: []string{"first", "", "last"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pgStringArraySplit(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
