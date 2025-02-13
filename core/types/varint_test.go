package types

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_uvarintLen(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    uint64
		expected int
	}{
		{
			name:     "zero value",
			input:    0,
			expected: len(binary.AppendUvarint(nil, 0)), // 1,
		},
		{
			name:     "single byte value",
			input:    127,
			expected: len(binary.AppendUvarint(nil, 127)), // 1,
		},
		{
			name:     "two byte value lower bound",
			input:    128,
			expected: len(binary.AppendUvarint(nil, 128)), // 2,
		},
		{
			name:     "two byte value upper bound",
			input:    16383,
			expected: len(binary.AppendUvarint(nil, 16383)), // 2,
		},
		{
			name:     "three byte value",
			input:    16384,
			expected: len(binary.AppendUvarint(nil, 16384)), // 3,
		},
		{
			name:     "max uint32",
			input:    4294967295,
			expected: len(binary.AppendUvarint(nil, 4294967295)), // 5,
		},
		{
			name:     "max uint64",
			input:    18446744073709551615,
			expected: len(binary.AppendUvarint(nil, 18446744073709551615)), // 10,
		},
		{
			name:     "power of 2 - 64",
			input:    1 << 63,
			expected: len(binary.AppendUvarint(nil, 1<<63)), // 10,
		},
		{
			name:     "large number requiring multiple bytes",
			input:    1234567890123456789,
			expected: len(binary.AppendUvarint(nil, 1234567890123456789)), // 9,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := uvarintLen(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func Test_varintLen(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    int64
		expected int
	}{
		{
			name:     "zero value",
			input:    0,
			expected: len(binary.AppendVarint(nil, 0)),
		},
		{
			name:     "max positive int64",
			input:    9223372036854775807,
			expected: len(binary.AppendVarint(nil, 9223372036854775807)),
		},
		{
			name:     "min negative int64",
			input:    -9223372036854775808,
			expected: len(binary.AppendVarint(nil, -9223372036854775808)),
		},
		{
			name:     "small negative number",
			input:    -42,
			expected: len(binary.AppendVarint(nil, -42)),
		},
		{
			name:     "small positive number",
			input:    42,
			expected: len(binary.AppendVarint(nil, 42)),
		},
		{
			name:     "power of 2 - 32",
			input:    1 << 31,
			expected: len(binary.AppendVarint(nil, 1<<31)),
		},
		{
			name:     "negative power of 2 - 32",
			input:    -(1 << 31),
			expected: len(binary.AppendVarint(nil, -(1 << 31))),
		},
		{
			name:     "large positive requiring multiple bytes",
			input:    1234567890123456789,
			expected: len(binary.AppendVarint(nil, 1234567890123456789)),
		},
		{
			name:     "large negative requiring multiple bytes",
			input:    -1234567890123456789,
			expected: len(binary.AppendVarint(nil, -1234567890123456789)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := varintLen(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func Test_uvarintLen_range(t *testing.T) {
	for i := range uint64(10_000_000) {
		expected := len(binary.AppendUvarint(nil, i))
		got := uvarintLen(i)
		if got != expected {
			t.Errorf("uvarintLen(%d) = %d, want %d", i, got, expected)
		}
	}

	// Test some larger values
	largeVals := []uint64{
		1<<32 - 1,
		1 << 32,
		1<<32 + 1,
		1<<63 - 1,
		1 << 63,
		1<<64 - 1,
	}
	for _, v := range largeVals {
		expected := len(binary.AppendUvarint(nil, v))
		got := uvarintLen(v)
		if got != expected {
			t.Errorf("uvarintLen(%d) = %d, want %d", v, got, expected)
		}
	}
}

func Test_varintLen_range(t *testing.T) {
	for i := int64(-6_000_000); i < 6_000_000; i++ {
		expected := len(binary.AppendVarint(nil, i))
		got := varintLen(i)
		if got != expected {
			t.Errorf("varintLen(%d) = %d, want %d", i, got, expected)
		}
	}

	// Test some edge cases
	edgeCases := []int64{
		-(1 << 63),
		-(1 << 32),
		-(1 << 16),
		1<<16 - 1,
		1 << 32,
		1<<32 - 1,
		1<<32 + 1,
		1<<63 - 1,
	}
	for _, v := range edgeCases {
		expected := len(binary.AppendVarint(nil, v))
		got := varintLen(v)
		if got != expected {
			t.Errorf("varintLen(%d) = %d, want %d", v, got, expected)
		}
	}
}

func TestCompactBytesRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       []byte
		expectError bool
		ensureNil   bool
	}{
		{
			name:        "nil bytes",
			input:       nil,
			expectError: false,
			ensureNil:   true,
		},
		{
			name:        "empty bytes",
			input:       []byte{},
			expectError: false,
		},
		{
			name:        "small data",
			input:       []byte("hello"),
			expectError: false,
		},
		{
			name:        "large data",
			input:       bytes.Repeat([]byte("x"), 1000),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Write to buffer
			buf := &bytes.Buffer{}
			err := WriteCompactBytes(buf, tc.input)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Read back and verify
			result, err := ReadCompactBytes(buf)
			require.NoError(t, err)
			require.Equal(t, tc.input, result)

			require.Equal(t, tc.ensureNil, result == nil)

			// Verify buffer is empty
			require.Equal(t, 0, buf.Len(), "buffer should be empty after reading")
		})
	}
}
