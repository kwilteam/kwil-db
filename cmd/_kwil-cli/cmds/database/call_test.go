package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SplitArray(t *testing.T) {
	type testCase struct {
		input    string
		expected []string
		wantErr  bool
	}

	tests := []testCase{
		{
			input: `name,"id",'age'`,
			expected: []string{
				"name",
				"id",
				"age",
			},
		},
		{
			input:   `name,"id",'age`,
			wantErr: true,
		},
		{
			input: `"val1,val1",val2,val3`,
			expected: []string{
				"val1,val1",
				"val2",
				"val3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := splitIgnoringQuotedCommas(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitIgnoringQuotedCommas() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.expected, got)
		})
	}
}
