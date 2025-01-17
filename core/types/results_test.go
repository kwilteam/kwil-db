package types

import (
	"encoding/binary"
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxResultMarshalUnmarshal(t *testing.T) {
	t.Run("empty result", func(t *testing.T) {
		tr := TxResult{
			Code:   0,
			Log:    "",
			Events: nil,
		}

		data, err := tr.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded TxResult
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Code != tr.Code {
			t.Errorf("got code %d, want %d", decoded.Code, tr.Code)
		}
		if decoded.Log != tr.Log {
			t.Errorf("got log %s, want %s", decoded.Log, tr.Log)
		}
		if len(decoded.Events) != 0 {
			t.Errorf("got %d events, want 0", len(decoded.Events))
		}
	})

	t.Run("with log and code", func(t *testing.T) {
		tr := TxResult{
			Code:   123,
			Log:    "test log message",
			Events: nil,
		}

		data, err := tr.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded TxResult
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Code != tr.Code {
			t.Errorf("got code %d, want %d", decoded.Code, tr.Code)
		}
		if decoded.Log != tr.Log {
			t.Errorf("got log %s, want %s", decoded.Log, tr.Log)
		}
	})

	t.Run("invalid data length", func(t *testing.T) {
		data := make([]byte, 3)
		var tr TxResult
		err := tr.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for insufficient data")
		}
	})

	t.Run("invalid log length", func(t *testing.T) {
		data := make([]byte, 8)
		binary.BigEndian.PutUint32(data, uint32(1))
		binary.BigEndian.PutUint32(data[2:], uint32(1000000))

		var tr TxResult
		err := tr.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid log length")
		}
	})

	t.Run("invalid events length", func(t *testing.T) {
		tr := TxResult{
			Code:   1,
			Log:    "test",
			Events: make([]Event, 65536),
		}

		_, err := tr.MarshalBinary()
		if err == nil {
			t.Error("expected error for too many events")
		}
	})

	// with events
	t.Run("with events", func(t *testing.T) {
		tr := TxResult{
			Code: 1,
			Log:  "test",
			Events: []Event{
				{},
			},
		}

		data, err := tr.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded TxResult
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Code != tr.Code {
			t.Errorf("got code %d, want %d", decoded.Code, tr.Code)
		}
		if decoded.Log != tr.Log {
			t.Errorf("got log %s, want %s", decoded.Log, tr.Log)
		}
		if len(decoded.Events) != len(tr.Events) {
			t.Errorf("got %d events, want 0", len(decoded.Events))
		}
	})
}

// errTestAny is a special error type used within tests if we want
// to signal that we just want any error, and dont care about the
// specific error type.
var errTestAny = errors.New("any test error")

func TestQueryResultScanScalars(t *testing.T) {
	type testcase struct {
		name   string
		rawval any // the value received from json unmarshalling
		// all of the "exp" (expect) values are the expected results
		// of scanning the rawval into the corresponding type.
		// They should be one of 3 values: the core type, nil, or error
		expString any
		expInt64  any
		expInt    any
		expBool   any
		expBytes  any
		expDec    any
		expUUID   any
	}

	tests := []testcase{
		{
			name:      "string",
			rawval:    "hello",
			expString: "hello",
			expInt64:  strconv.ErrSyntax,
			expInt:    strconv.ErrSyntax,
			expBool:   strconv.ErrSyntax,
			expBytes:  []byte("hello"),
			expDec:    strconv.ErrSyntax,
			expUUID:   errTestAny,
		},
		{
			name:      "int64",
			rawval:    int64(123),
			expString: "123",
			expInt64:  int64(123),
			expInt:    int(123),
			expBool:   strconv.ErrSyntax,
			expBytes:  []byte("123"),
			expDec:    *MustParseDecimal("123"),
			expUUID:   errTestAny,
		},
		{
			name:      "int",
			rawval:    int(123),
			expString: "123",
			expInt64:  int64(123),
			expInt:    int(123),
			expBool:   strconv.ErrSyntax,
			expBytes:  []byte("123"),
			expDec:    *MustParseDecimal("123"),
			expUUID:   errTestAny,
		},
		{
			name: "int string",
			// this is a string that looks like an int
			rawval:    "123",
			expString: "123",
			expInt64:  int64(123),
			expInt:    int(123),
			expBool:   strconv.ErrSyntax,
			expBytes:  []byte("123"),
			expDec:    *MustParseDecimal("123"),
			expUUID:   errTestAny,
		},
		{
			name:      "bool",
			rawval:    true,
			expString: "true",
			expInt64:  strconv.ErrSyntax,
			expInt:    strconv.ErrSyntax,
			expBool:   true,
			expBytes:  []byte("true"),
			expDec:    strconv.ErrSyntax,
			expUUID:   errTestAny,
		},
		{
			name:      "bytes",
			rawval:    []byte("hello"),
			expString: "hello",
			expInt64:  strconv.ErrSyntax,
			expInt:    strconv.ErrSyntax,
			expBool:   strconv.ErrSyntax,
			expBytes:  []byte("hello"),
			expDec:    strconv.ErrSyntax,
			expUUID:   errTestAny,
		},
		{
			name:      "bytes (16 bytes)",
			rawval:    MustParseUUID("12345678-1234-1234-1234-123456789abc").Bytes(),
			expString: string(MustParseUUID("12345678-1234-1234-1234-123456789abc").Bytes()),
			expInt64:  strconv.ErrSyntax,
			expInt:    strconv.ErrSyntax,
			expBool:   strconv.ErrSyntax,
			expBytes:  MustParseUUID("12345678-1234-1234-1234-123456789abc").Bytes(),
			expDec:    errTestAny,
			expUUID:   *MustParseUUID("12345678-1234-1234-1234-123456789abc"),
		},
		{
			name:      "decimal",
			rawval:    "123.456",
			expString: "123.456",
			expInt64:  strconv.ErrSyntax,
			expInt:    strconv.ErrSyntax,
			expBool:   strconv.ErrSyntax,
			expBytes:  []byte("123.456"),
			expDec:    *MustParseDecimal("123.456"),
			expUUID:   errTestAny,
		},
		{
			name:      "uuid",
			rawval:    "12345678-1234-1234-1234-123456789abc",
			expString: "12345678-1234-1234-1234-123456789abc",
			expInt64:  strconv.ErrSyntax,
			expInt:    strconv.ErrSyntax,
			expBool:   strconv.ErrSyntax,
			expBytes:  []byte("12345678-1234-1234-1234-123456789abc"),
			expDec:    errTestAny,
			expUUID:   *MustParseUUID("12345678-1234-1234-1234-123456789abc"),
		},
		{
			name: "nil",
			// this is a nil value
			rawval:    nil,
			expString: nil,
			expInt64:  nil,
			expInt:    nil,
			expBool:   nil,
			expBytes:  nil,
			expDec:    nil,
			expUUID:   nil,
		},
		{
			name:      "float",
			rawval:    float64(123.456),
			expString: "123.456",
			expInt64:  strconv.ErrSyntax,
			expInt:    strconv.ErrSyntax,
			expBool:   strconv.ErrSyntax,
			expBytes:  []byte("123.456"),
			expDec:    *MustParseDecimal("123.456"),
			expUUID:   errTestAny,
		},
		{
			name:      "round float",
			rawval:    float32(123),
			expString: "123",
			expInt64:  int64(123),
			expInt:    int(123),
			expBool:   strconv.ErrSyntax,
			expBytes:  []byte("123"),
			expDec:    *MustParseDecimal("123"),
			expUUID:   errTestAny,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qr := &QueryResult{
				Values: [][]any{{tt.rawval}},
			}
			checkType[string](t, qr, tt.expString)
			checkType[int64](t, qr, tt.expInt64)
			checkType[int](t, qr, tt.expInt)
			checkType[bool](t, qr, tt.expBool)
			checkType[[]byte](t, qr, tt.expBytes)
			checkType[Decimal](t, qr, tt.expDec)
			checkType[UUID](t, qr, tt.expUUID)
		})
	}
}

func checkType[T any](t *testing.T, q *QueryResult, want any) {
	var name string
	_, wantErr := want.(error)
	if want != nil && !wantErr {
		typeof := reflect.TypeOf(want)
		isPtr := false
		if typeof.Kind() == reflect.Ptr {
			isPtr = true
			typeof = typeof.Elem()
		}
		name = typeof.String()
		if isPtr {
			name = "*" + name
		}
	} else if wantErr {
		name = "error"
	} else {
		name = "nil"
	}
	t.Logf("testing type %T, expecting %s", *new(T), name)

	v := new(T)
	err := q.Scan([]any{v}, func() error {
		return nil
	})

	switch want := want.(type) {
	case nil:
		assert.NoError(t, err)

		newNil := new(T)
		assert.EqualValues(t, newNil, v)
	case error:
		if want == errTestAny {
			assert.Error(t, err)
		} else {
			assert.ErrorIs(t, err, want)
		}

		newNil := new(T)
		assert.EqualValues(t, newNil, v)
	case T:
		assert.NoError(t, err)
		assert.EqualValues(t, want, *v)
	default:
		t.Fatalf("unexpected want type %T", want)
	}
}

func TestQueryResultScanArrays(t *testing.T) {
	type testcase struct {
		name   string
		rawval any // the value received from json unmarshalling
		// all of the "exp" (expect) values are the expected results
		// of scanning the rawval into the corresponding type.
		// They should be one of 3 values: the core type, nil, or error.
		expStringArr    any
		expStringArrPtr any
		expInt64Arr     any
		expInt64ArrPtr  any
		expIntArr       any
		expIntArrPtr    any
		expBoolArr      any
		expBoolArrPtr   any
		expBytesArr     any
		expBytesArrPtr  any
		expDecArr       any
		expDecArrPtr    any
		expUUIDArr      any
		expUUIDArrPtr   any
	}

	tests := []testcase{
		{
			name:            "string",
			rawval:          []any{"hello", "world", nil},
			expStringArr:    []string{"hello", "world", ""},
			expStringArrPtr: ptrArr[string]("hello", "world", nil),
			expInt64Arr:     errTestAny,
			expInt64ArrPtr:  errTestAny,
			expIntArr:       errTestAny,
			expIntArrPtr:    errTestAny,
			expBoolArr:      errTestAny,
			expBoolArrPtr:   errTestAny,
			expBytesArr:     [][]byte{[]byte("hello"), []byte("world"), nil},
			expBytesArrPtr:  ptrArr[[]byte]([]byte("hello"), []byte("world"), nil),
			expDecArr:       errTestAny,
			expDecArrPtr:    errTestAny,
			expUUIDArr:      errTestAny,
			expUUIDArrPtr:   errTestAny,
		},
		{
			name:            "int64",
			rawval:          []any{int64(123), int64(456), nil},
			expStringArr:    []string{"123", "456", ""},
			expStringArrPtr: ptrArr[string]("123", "456", nil),
			expInt64Arr:     []int64{int64(123), int64(456), 0},
			expInt64ArrPtr:  ptrArr[int64](int64(123), int64(456), nil),
			expIntArr:       []int{123, 456, 0},
			expIntArrPtr:    ptrArr[int](123, 456, nil),
			expBoolArr:      errTestAny,
			expBoolArrPtr:   errTestAny,
			expBytesArr:     [][]byte{[]byte("123"), []byte("456"), nil},
			expBytesArrPtr:  ptrArr[[]byte]([]byte("123"), []byte("456"), nil),
			expDecArr:       []Decimal{*MustParseDecimal("123"), *MustParseDecimal("456"), {}},
			expDecArrPtr:    ptrArr[Decimal](*MustParseDecimal("123"), *MustParseDecimal("456"), nil),
			expUUIDArr:      errTestAny,
			expUUIDArrPtr:   errTestAny,
		},
		{
			name:            "int",
			rawval:          []any{int(123), int(456), nil},
			expStringArr:    []string{"123", "456", ""},
			expStringArrPtr: ptrArr[string]("123", "456", nil),
			expInt64Arr:     []int64{int64(123), int64(456), 0},
			expInt64ArrPtr:  ptrArr[int64](int64(123), int64(456), nil),
			expIntArr:       []int{123, 456, 0},
			expIntArrPtr:    ptrArr[int](123, 456, nil),
			expBoolArr:      errTestAny,
			expBoolArrPtr:   errTestAny,
			expBytesArr:     [][]byte{[]byte("123"), []byte("456"), nil},
			expBytesArrPtr:  ptrArr[[]byte]([]byte("123"), []byte("456"), nil),
			expDecArr:       []Decimal{*MustParseDecimal("123"), *MustParseDecimal("456"), {}},
			expDecArrPtr:    ptrArr[Decimal](*MustParseDecimal("123"), *MustParseDecimal("456"), nil),
			expUUIDArr:      errTestAny,
			expUUIDArrPtr:   errTestAny,
		},
		{
			name:            "bool",
			rawval:          []any{true, false, nil},
			expStringArr:    []string{"true", "false", ""},
			expStringArrPtr: ptrArr[string]("true", "false", nil),
			expInt64Arr:     errTestAny,
			expInt64ArrPtr:  errTestAny,
			expIntArr:       errTestAny,
			expIntArrPtr:    errTestAny,
			expBoolArr:      []bool{true, false, false},
			expBoolArrPtr:   ptrArr[bool](true, false, nil),
			expBytesArr:     [][]byte{[]byte("true"), []byte("false"), nil},
			expBytesArrPtr:  ptrArr[[]byte]([]byte("true"), []byte("false"), nil),
			expDecArr:       errTestAny,
			expDecArrPtr:    errTestAny,
			expUUIDArr:      errTestAny,
			expUUIDArrPtr:   errTestAny,
		},
		{
			name:            "bytes",
			rawval:          []any{[]byte("hello"), []byte("world"), nil},
			expStringArr:    []string{"hello", "world", ""},
			expStringArrPtr: ptrArr[string]("hello", "world", nil),
			expInt64Arr:     errTestAny,
			expInt64ArrPtr:  errTestAny,
			expIntArr:       errTestAny,
			expIntArrPtr:    errTestAny,
			expBoolArr:      errTestAny,
			expBoolArrPtr:   errTestAny,
			expBytesArr:     [][]byte{[]byte("hello"), []byte("world"), nil},
			expBytesArrPtr:  ptrArr[[]byte]([]byte("hello"), []byte("world"), nil),
			expDecArr:       errTestAny,
			expDecArrPtr:    errTestAny,
			expUUIDArr:      errTestAny,
			expUUIDArrPtr:   errTestAny,
		},
		{
			name:            "decimal",
			rawval:          []any{"123.456", "789.012", nil},
			expStringArr:    []string{"123.456", "789.012", ""},
			expStringArrPtr: ptrArr[string]("123.456", "789.012", nil),
			expInt64Arr:     errTestAny,
			expInt64ArrPtr:  errTestAny,
			expIntArr:       errTestAny,
			expIntArrPtr:    errTestAny,
			expBoolArr:      errTestAny,
			expBoolArrPtr:   errTestAny,
			expBytesArr:     [][]byte{[]byte("123.456"), []byte("789.012"), nil},
			expBytesArrPtr:  ptrArr[[]byte]([]byte("123.456"), []byte("789.012"), nil),
			expDecArr:       []Decimal{*MustParseDecimal("123.456"), *MustParseDecimal("789.012"), {}},
			expDecArrPtr:    ptrArr[Decimal](*MustParseDecimal("123.456"), *MustParseDecimal("789.012"), nil),
			expUUIDArr:      errTestAny,
			expUUIDArrPtr:   errTestAny,
		},
		{
			name:            "uuid",
			rawval:          []any{"12345678-1234-1234-1234-123456789abc", "12345678-1234-1234-1234-123456789def", nil},
			expStringArr:    []string{"12345678-1234-1234-1234-123456789abc", "12345678-1234-1234-1234-123456789def", ""},
			expStringArrPtr: ptrArr[string]("12345678-1234-1234-1234-123456789abc", "12345678-1234-1234-1234-123456789def", nil),
			expInt64Arr:     errTestAny,
			expInt64ArrPtr:  errTestAny,
			expIntArr:       errTestAny,
			expIntArrPtr:    errTestAny,
			expBoolArr:      errTestAny,
			expBoolArrPtr:   errTestAny,
			expBytesArr:     [][]byte{[]byte("12345678-1234-1234-1234-123456789abc"), []byte("12345678-1234-1234-1234-123456789def"), nil},
			expBytesArrPtr:  ptrArr[[]byte]([]byte("12345678-1234-1234-1234-123456789abc"), []byte("12345678-1234-1234-1234-123456789def"), nil),
			expDecArr:       errTestAny,
			expDecArrPtr:    errTestAny,
			expUUIDArr:      []UUID{*MustParseUUID("12345678-1234-1234-1234-123456789abc"), *MustParseUUID("12345678-1234-1234-1234-123456789def"), UUID{}},
			expUUIDArrPtr:   ptrArr[UUID](*MustParseUUID("12345678-1234-1234-1234-123456789abc"), *MustParseUUID("12345678-1234-1234-1234-123456789def"), nil),
		},
		{
			name:            "all nil values",
			rawval:          []any{nil, nil, nil},
			expStringArr:    []string{"", "", ""},
			expStringArrPtr: ptrArr[string](nil, nil, nil),
			expInt64Arr:     []int64{0, 0, 0},
			expInt64ArrPtr:  ptrArr[int64](nil, nil, nil),
			expIntArr:       []int{0, 0, 0},
			expIntArrPtr:    ptrArr[int](nil, nil, nil),
			expBoolArr:      []bool{false, false, false},
			expBoolArrPtr:   ptrArr[bool](nil, nil, nil),
			expBytesArr:     [][]byte{nil, nil, nil},
			expBytesArrPtr:  ptrArr[[]byte](nil, nil, nil),
			expDecArr:       []Decimal{{}, {}, {}},
			expDecArrPtr:    ptrArr[Decimal](nil, nil, nil),
			expUUIDArr:      []UUID{UUID{}, UUID{}, UUID{}},
			expUUIDArrPtr:   ptrArr[UUID](nil, nil, nil),
		},
		{
			name:            "nil",
			rawval:          nil,
			expStringArr:    nil,
			expStringArrPtr: nil,
			expInt64Arr:     nil,
			expInt64ArrPtr:  nil,
			expIntArr:       nil,
			expIntArrPtr:    nil,
			expBoolArr:      nil,
			expBoolArrPtr:   nil,
			expBytesArr:     nil,
			expBytesArrPtr:  nil,
			expDecArr:       nil,
			expDecArrPtr:    nil,
			expUUIDArr:      nil,
			expUUIDArrPtr:   nil,
		},
		{
			name:            "float",
			rawval:          []any{float64(123.456), float64(789), nil},
			expStringArr:    []string{"123.456", "789", ""},
			expStringArrPtr: ptrArr[string]("123.456", "789", nil),
			expInt64Arr:     errTestAny,
			expInt64ArrPtr:  errTestAny,
			expIntArr:       errTestAny,
			expIntArrPtr:    errTestAny,
			expBoolArr:      errTestAny,
			expBoolArrPtr:   errTestAny,
			expBytesArr:     [][]byte{[]byte("123.456"), []byte("789"), nil},
			expBytesArrPtr:  ptrArr[[]byte]([]byte("123.456"), []byte("789"), nil),
			expDecArr:       []Decimal{*MustParseDecimal("123.456"), *MustParseDecimal("789"), {}},
			expDecArrPtr:    ptrArr[Decimal](*MustParseDecimal("123.456"), *MustParseDecimal("789"), nil),
			expUUIDArr:      errTestAny,
			expUUIDArrPtr:   errTestAny,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qr := &QueryResult{
				Values: [][]any{{tt.rawval}},
			}
			checkType[[]string](t, qr, tt.expStringArr)
			checkType[[]*string](t, qr, tt.expStringArrPtr)
			checkType[[]int64](t, qr, tt.expInt64Arr)
			checkType[[]*int64](t, qr, tt.expInt64ArrPtr)
			checkType[[]int](t, qr, tt.expIntArr)
			checkType[[]*int](t, qr, tt.expIntArrPtr)
			checkType[[]bool](t, qr, tt.expBoolArr)
			checkType[[]*bool](t, qr, tt.expBoolArrPtr)
			checkType[[][]byte](t, qr, tt.expBytesArr)
			checkType[[]*[]byte](t, qr, tt.expBytesArrPtr)
			checkType[[]Decimal](t, qr, tt.expDecArr)
			checkType[[]*Decimal](t, qr, tt.expDecArrPtr)
			checkType[[]UUID](t, qr, tt.expUUIDArr)
			checkType[[]*UUID](t, qr, tt.expUUIDArrPtr)
		})
	}
}

// Im checking here that users are capable of detecting zero length
// arrays vs null arrays
func TestScanArrayNullability(t *testing.T) {
	v := new([]string)
	qr := &QueryResult{
		Values: [][]any{{[]any{}}},
	}
	err := qr.Scan([]any{v}, func() error {
		return nil
	})
	assert.NoError(t, err)

	assert.True(t, *v != nil)
	assert.Len(t, *v, 0)

	v = new([]string)
	v2 := new([]string)
	qr = &QueryResult{
		Values: [][]any{{[]any{"a"}, nil}},
	}
	err = qr.Scan([]any{v, v2}, func() error {
		return nil
	})
	assert.NoError(t, err)

	assert.True(t, *v != nil)
	assert.Len(t, *v, 1)

	assert.False(t, *v2 != nil)
}

func ptrArr[T any](v ...any) []*T {
	out := make([]*T, len(v))
	for i, b := range v {
		if b == nil {
			out[i] = nil
			continue
		}

		convV, ok := b.(T)
		if !ok {
			panic("invalid type")
		}

		out[i] = &convV
	}
	return out
}
