package serialize_test

import (
	"testing"

	serialize "github.com/kwilteam/kwil-db/pkg/serialize"
	"github.com/stretchr/testify/assert"
)

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

func Test_Encoding(t *testing.T) {
	type testCase struct {
		name string
		testable
	}

	testCases := []testCase{
		{
			name: "valid struct",
			testable: genericTestCase[TestStruct1]{
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
		},
		{
			name: "unexported field",
			testable: genericTestCase[StructUnexportedField]{
				input: StructUnexportedField{
					val1: 1,
					Val2: "test",
				},
				output: StructUnexportedField{
					Val2: "test",
				},
			},
		},
		{
			name: "invalid struct - signed int",
			testable: genericTestCase[InvalidStructSignedInt]{
				input: InvalidStructSignedInt{
					Val1: -1,
				},
				encodingErr: true,
			},
		},
		{
			name: "invalid struct - map",
			testable: genericTestCase[InvalidStructMap]{
				input: InvalidStructMap{
					Val1: map[string]string{
						"test": "test",
					},
				},
				encodingErr: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.testable.runTest(t)
		})
	}
}

type testable interface {
	runTest(t *testing.T)
}

type genericTestCase[T any] struct {
	input       T
	output      T
	encodingErr bool
}

func (g genericTestCase[T]) runTest(t *testing.T) {
	output, err := serialize.Encode(g.input)
	if g.encodingErr && err == nil {
		t.Errorf("Expected error, got nil")
	}
	if !g.encodingErr && err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if g.encodingErr {
		return
	}

	result := new(T)
	err = serialize.DecodeInto(output, result)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	assert.EqualValuesf(t, g.output, *result, "Expected result to be %v, got %v", g.output, *result)

}

func Test_EncodeSlice(t *testing.T) {

	// this is an atypical way of testing, made to work with generics
	type testCase struct {
		name     string
		testFunc testable
	}

	testCases := []testCase{
		{
			testFunc: genericSliceTestCase[TestStruct1]{
				input: []*TestStruct1{
					{
						Val1: 1,
						Val3: []byte("test"),
						Val2: "test",
						Val4: true,
					},
					{
						Val1: 2,
						Val3: []byte("test2"),
						Val2: "test2",
						Val4: false,
					},
				},
				output: []*TestStruct1{
					{
						Val1: 1,
						Val2: "test",
						Val3: []byte("test"),
						Val4: true,
					},
					{
						Val1: 2,
						Val2: "test2",
						Val3: []byte("test2"),
						Val4: false,
					},
				},
				isEqual: true,
			},
		},
		{
			name: "valid struct, invalid values",
			testFunc: genericSliceTestCase[TestStruct1]{
				input: []*TestStruct1{
					{
						Val1: 1,
						Val3: []byte("test"),
						Val2: "test",
						Val4: true,
					},
					{
						Val1: 2,
						Val3: []byte("test2"),
						Val2: "test2",
						Val4: false,
					},
				},
				output: []*TestStruct1{
					{
						Val1: 1,
						Val2: "test",
						Val3: []byte("test"),
						Val4: true,
					},
					{
						Val1: 2000000,
						Val2: "test2",
						Val3: []byte("test2"),
						Val4: false,
					},
				},
				isEqual: false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.testFunc.runTest)
	}
}

type genericSliceTestCase[T any] struct {
	input   []*T
	output  []*T
	isEqual bool
}

func (g genericSliceTestCase[T]) runTest(t *testing.T) {
	out, err := serialize.EncodeSlice(g.input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	result, err := serialize.DecodeSlice[T](out)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if g.isEqual {
		assert.EqualValuesf(t, g.output, result, "Expected result to be %v, got %v", g.output, result)
		return
	} else {
		assert.NotEqualValuesf(t, g.output, result, "Expected result to be %v, got %v", g.output, result)
		return
	}

}
