package rlp_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/pkg/serialize/rlp"
	"github.com/stretchr/testify/assert"
)

func Test_Encoding(t *testing.T) {
	type testCase struct {
		name        string
		input       any
		output      any
		encodingErr bool
	}

	testCases := []testCase{
		{
			name: "valid struct",
			input: TestStruct1{
				Val1: 1,
				Val3: []byte("test"),
				Val2: "test",
				Val4: true,
			},
			output: TestStruct1{
				Val1: 1,
				Val2: "test",
				Val3: []byte("test"),
				Val4: true,
			},
		},
		{
			name: "unexported field",
			input: StructUnexportedField{
				val1: 1,
				Val2: "test",
			},
			output: StructUnexportedField{
				Val2: "test",
			},
		},
		{
			name: "invalid struct - signed int",
			input: InvalidStructSignedInt{
				Val1: -1,
			},
			encodingErr: true,
		},
		{
			name: "invalid struct - map",
			input: InvalidStructMap{
				Val1: map[string]string{
					"test": "test",
				},
			},
			encodingErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := rlp.Encode(tc.input)
			if tc.encodingErr && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tc.encodingErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if tc.encodingErr {
				return
			}

			decoded, err := rlp.Decode[any](output)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if err != nil {
				return
			}

			if *decoded == tc.output {
				assert.Equal(t, tc.output, *decoded)
			}
		})
	}
}

type TestStruct1 struct {
	Val1 uint64
	Val2 string
	Val3 []byte
	Val4 bool
}

type InvalidStructSignedInt struct {
	Val1 int64 // RLPEncode only supports unsigned integers
}

type InvalidStructMap struct {
	Val1 map[string]string // RLPEncode does not support maps
}

// will not error, but will not encode
type StructUnexportedField struct {
	val1 uint64 // RLPEncode only supports exported fields
	Val2 string
}
