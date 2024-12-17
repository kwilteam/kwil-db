package userjson

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBroadcastError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      BroadcastError
		expected string
	}{
		{
			name: "standard error",
			err: BroadcastError{
				TxCode:  1,
				Hash:    "abc123",
				Message: "test error",
			},
			expected: "broadcast error: code = 1, hash = abc123, msg = test error",
		},
		{
			name: "zero code error",
			err: BroadcastError{
				TxCode:  0,
				Hash:    "def456",
				Message: "another error",
			},
			expected: "broadcast error: code = 0, hash = def456, msg = another error",
		},
		{
			name: "empty hash error",
			err: BroadcastError{
				TxCode:  2,
				Hash:    "",
				Message: "empty hash",
			},
			expected: "broadcast error: code = 2, hash = , msg = empty hash",
		},
		{
			name: "empty message error",
			err: BroadcastError{
				TxCode:  3,
				Hash:    "ghi789",
				Message: "",
			},
			expected: "broadcast error: code = 3, hash = ghi789, msg = ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			require.Equal(t, tt.expected, result)
		})
	}
}
