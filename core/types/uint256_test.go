package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// func intHex(v int64) []byte

const hugeIntStrP1 = "18446744073709551616"   // 1 + math.MaxUint64
const hugeIntStrX10 = "184467440737095516150" // 10 * math.MaxUint64

func TestUint256BinaryMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		val      string
		expected []byte
	}{
		{
			name:     "small int",
			val:      "123",
			expected: []byte{0x7b},
		},
		{
			name:     "zero",
			val:      "0",
			expected: []byte{}, // optimized to empty slice
		},
		{
			name:     "null",
			val:      "", // special case
			expected: nil,
		},
		{
			name:     "just bigger than uint64",
			val:      hugeIntStrP1,
			expected: []byte{0x01, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "much than uint64",
			val:      hugeIntStrX10,
			expected: []byte{0x09, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xf6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u *Uint256
			if tt.val == "" {
				u = &Uint256{Null: true}
			} else {
				var err error
				u, err = Uint256FromString(tt.val)
				require.NoError(t, err)
			}

			marshaled, err := u.MarshalBinary()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, marshaled)

			var unmarshaled Uint256
			err = unmarshaled.UnmarshalBinary(marshaled)
			require.NoError(t, err)

			assert.Equal(t, u.String(), unmarshaled.String())
		})
	}
}

func TestUint256JSONRoundTrip(t *testing.T) {
	for _, str := range []string{"12345", "0", "", hugeIntStrX10} {
		var original *Uint256
		if str == "" {
			original = &Uint256{Null: true}
		} else {
			var err error
			original, err = Uint256FromString(str)
			require.NoError(t, err)
		}

		marshaled, err := original.MarshalJSON()
		require.NoError(t, err)
		if len(str) > 0 {
			require.Equal(t, `"`+str+`"`, string(marshaled))
		} else {
			require.Equal(t, string(marshaled), "null")
		}

		var unmarshaled Uint256
		err = unmarshaled.UnmarshalJSON(marshaled)
		require.NoError(t, err)

		assert.Equal(t, original.String(), unmarshaled.String())
	}

}

func Test_Uint256Math(t *testing.T) {
	// simply testing that the base number is not modified
	a, err := Uint256FromString("500")
	require.NoError(t, err)

	b, err := Uint256FromString("10000000000")
	require.NoError(t, err)

	c := a.Add(b)
	require.Equal(t, "10000000500", c.String())
	require.Equal(t, "500", a.String())
	require.Equal(t, "10000000000", b.String())

	// go underflow
	_, err = a.Sub(b)
	require.Error(t, err)

	// div without mod
	d, err := Uint256FromString("498")
	require.NoError(t, err)

	e := a.Div(d)
	require.Equal(t, "1", e.String())

	// div mod
	f, g := a.DivMod(d)
	require.Equal(t, "1", f.String())
	require.Equal(t, "2", g.String())
}
