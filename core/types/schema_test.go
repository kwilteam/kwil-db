package types_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
)

func Test_ParseDataTypes(t *testing.T) {
	type testcase struct {
		in        string
		out       types.DataType
		wantError bool
	}

	tests := []testcase{
		{
			in: "int8",
			out: types.DataType{
				Name: "int8",
			},
		},
		{
			in: "int8[]",
			out: types.DataType{
				Name:    "int8",
				IsArray: true,
			},
		},
		{
			in: "text[]",
			out: types.DataType{
				Name:    "text",
				IsArray: true,
			},
		},
		{
			in: "decimal(10, 2)",
			out: types.DataType{
				Name:     "decimal",
				Metadata: &[2]uint16{10, 2},
			},
		},
		{
			in: "decimal(10, 2)[]",
			out: types.DataType{
				Name:     "decimal",
				Metadata: &[2]uint16{10, 2},
				IsArray:  true,
			},
		},
		{
			in:        "decimal(10, 2)[][]",
			wantError: true,
		},
		{
			in:        "text(10, 2)",
			wantError: true,
		},
		{
			in:        "text(10)",
			wantError: true,
		},
		{
			in:        "decimal(10)",
			wantError: true,
		},
		{
			in:        "decimal(10, 2, 3)",
			wantError: true,
		},
		{
			in:        "decimal(10, a)",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			res, err := types.ParseDataType(tt.in)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tt.out.Equals(res) {
				t.Fatalf("expected %v, got %v", tt.out, res)
			}
		})
	}
}
