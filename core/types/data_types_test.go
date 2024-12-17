package types

import (
	"testing"
)

func TestDataTypeBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal valid data type", func(t *testing.T) {
		original := DataType{
			Name:     "test_type",
			IsArray:  true,
			Metadata: [2]uint16{42, 123},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded DataType
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != original.Name {
			t.Errorf("got name %s, want %s", decoded.Name, original.Name)
		}
		if decoded.IsArray != original.IsArray {
			t.Errorf("got isArray %v, want %v", decoded.IsArray, original.IsArray)
		}
		if decoded.Metadata != original.Metadata {
			t.Errorf("got metadata %v, want %v", decoded.Metadata, original.Metadata)
		}
	})

	t.Run("unmarshal with insufficient data length", func(t *testing.T) {
		data := []byte{0, 0, 0, 0}
		var dt DataType
		err := dt.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for insufficient data length")
		}
	})

	t.Run("unmarshal with invalid version", func(t *testing.T) {
		data := []byte{0, 1, 0, 0, 0, 0}
		var dt DataType
		err := dt.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid version")
		}
	})

	t.Run("unmarshal with invalid name length", func(t *testing.T) {
		data := []byte{0, 0, 255, 255, 255, 255}
		var dt DataType
		err := dt.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid name length")
		}
	})

	t.Run("marshal empty name", func(t *testing.T) {
		original := DataType{
			Name:     "",
			IsArray:  false,
			Metadata: [2]uint16{0, 0},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded DataType
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded != original {
			t.Errorf("got %v, want %v", decoded, original)
		}
	})

	t.Run("marshal with maximum metadata values", func(t *testing.T) {
		original := DataType{
			Name:     "test",
			IsArray:  true,
			Metadata: [2]uint16{65535, 65535},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded DataType
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded != original {
			t.Errorf("got %v, want %v", decoded, original)
		}
	})
}

func Test_ParseDataTypes(t *testing.T) {
	type testcase struct {
		in        string
		out       DataType
		wantError bool
	}

	tests := []testcase{
		{
			in: "int8",
			out: DataType{
				Name: "int8",
			},
		},
		{
			in: "int8[]",
			out: DataType{
				Name:    "int8",
				IsArray: true,
			},
		},
		{
			in: "text[]",
			out: DataType{
				Name:    "text",
				IsArray: true,
			},
		},
		{
			in: "decimal(10, 2)",
			out: DataType{
				Name:     "decimal",
				Metadata: [2]uint16{10, 2},
			},
		},
		{
			in: "decimal(10, 2)[]",
			out: DataType{
				Name:     "decimal",
				Metadata: [2]uint16{10, 2},
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
			res, err := ParseDataType(tt.in)
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
