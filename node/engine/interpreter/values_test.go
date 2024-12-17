package interpreter

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Arithmetic(t *testing.T) {
	type testcase struct {
		name   string
		a      any
		b      any
		add    any
		sub    any
		mul    any
		div    any
		mod    any
		concat any
	}

	tests := []testcase{
		{
			name:   "int",
			a:      int64(10),
			b:      int64(5),
			add:    int64(15),
			sub:    int64(5),
			mul:    int64(50),
			div:    int64(2),
			mod:    int64(0),
			concat: ErrArithmetic,
		},
		{
			name:   "decimal",
			a:      mustDec("10.00"),
			b:      mustDec("5.00"),
			add:    mustDec("15.00"),
			sub:    mustDec("5.00"),
			mul:    mustDec("50.00"),
			div:    mustDec("2.00"),
			mod:    mustDec("0.00"),
			concat: ErrArithmetic,
		},
		{
			name:   "text",
			a:      "hello",
			b:      "world",
			add:    ErrArithmetic,
			sub:    ErrArithmetic,
			mul:    ErrArithmetic,
			div:    ErrArithmetic,
			mod:    ErrArithmetic,
			concat: "helloworld",
		},
		{
			name:   "uuid",
			a:      mustUUID("550e8400-e29b-41d4-a716-446655440000"),
			b:      mustUUID("550e8400-e29b-41d4-a716-446655440000"),
			add:    ErrArithmetic,
			sub:    ErrArithmetic,
			mul:    ErrArithmetic,
			div:    ErrArithmetic,
			mod:    ErrArithmetic,
			concat: ErrArithmetic,
		},
		{
			name:   "blob",
			a:      []byte("hello"),
			b:      []byte("world"),
			add:    ErrArithmetic,
			sub:    ErrArithmetic,
			mul:    ErrArithmetic,
			div:    ErrArithmetic,
			mod:    ErrArithmetic,
			concat: []byte("helloworld"),
		},
		{
			name:   "bool",
			a:      true,
			b:      false,
			add:    ErrArithmetic,
			sub:    ErrArithmetic,
			mul:    ErrArithmetic,
			div:    ErrArithmetic,
			mod:    ErrArithmetic,
			concat: ErrArithmetic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			makeVal := func(v any) ScalarValue {
				val, err := NewValue(v)
				require.NoError(t, err)
				return val.(ScalarValue)
			}

			a := makeVal(tt.a)
			b := makeVal(tt.b)

			isErrOrResult := func(a, b ScalarValue, op ArithmeticOp, want any) {
				res, err := a.Arithmetic(b, op)
				if wantErr, ok := want.(error); ok {
					require.Error(t, err)
					require.ErrorIs(t, err, wantErr)
					return
				}
				require.NoError(t, err)

				raw := res.RawValue()

				eq(t, want, raw)

				// operations on null values should always return null
				null := newNull(a.Type()).(ScalarValue)

				res, err = a.Arithmetic(null, op)
				require.NoError(t, err)

				require.True(t, res.Null())
				require.Nil(t, res.RawValue())
			}

			isErrOrResult(a, b, add, tt.add)
			isErrOrResult(a, b, sub, tt.sub)
			isErrOrResult(a, b, mul, tt.mul)
			isErrOrResult(a, b, div, tt.div)
			isErrOrResult(a, b, mod, tt.mod)
			isErrOrResult(a, b, concat, tt.concat)

			// test rountripping strings
			testRoundTripParse(t, a)
			testRoundTripParse(t, b)
		})
	}
}

// eq is a helper function that checks if two values are equal.
// It handles the semantics of comparing decimal values.
func eq(t *testing.T, a, b any) {
	// if the values are decimals, we need to compare them manually
	if aDec, ok := a.(*decimal.Decimal); ok {
		bDec, ok := b.(*decimal.Decimal)
		require.True(t, ok)

		rec, err := aDec.Cmp(bDec)
		require.NoError(t, err)
		assert.Zero(t, rec)
		return
	}

	if aDec, ok := a.([]*decimal.Decimal); ok {
		bDec, ok := b.([]*decimal.Decimal)
		require.True(t, ok)

		require.Len(t, aDec, len(bDec))
		for i := range aDec {
			eq(t, aDec[i], bDec[i])
		}
		return
	}

	assert.EqualValues(t, a, b)
}

func Test_Comparison(t *testing.T) {
	type testcase struct {
		name         string
		a            any
		b            any
		gt           any
		lt           any
		eq           any
		is           any
		distinctFrom any
	}

	// there are 6 types: int, text, bool, blob, uuid, decimal
	// Each type can also have a one dimensional array of that type
	// We need tests for each type and each array type, testing comparison against each other
	// as well as against null values.
	tests := []testcase{
		{
			name:         "int",
			a:            int64(10),
			b:            int64(5),
			eq:           false,
			gt:           true,
			lt:           false,
			is:           ErrComparison,
			distinctFrom: true,
		},
		{
			name:         "decimal",
			a:            mustDec("10.00"),
			b:            mustDec("5.00"),
			eq:           false,
			gt:           true,
			lt:           false,
			is:           ErrComparison,
			distinctFrom: true,
		},
		{
			name:         "text",
			a:            "hello",
			b:            "world",
			eq:           false,
			gt:           false,
			lt:           true,
			is:           ErrComparison,
			distinctFrom: true,
		},
		{
			name:         "uuid",
			a:            mustUUID("550e8400-e29b-41d4-a716-446655440000"),
			b:            mustUUID("550e8400-e29b-41d4-a716-446655440000"),
			eq:           true,
			gt:           ErrComparison,
			lt:           ErrComparison,
			is:           ErrComparison,
			distinctFrom: false,
		},
		{
			name:         "blob",
			a:            []byte("hello"),
			b:            []byte("world"),
			eq:           false,
			gt:           ErrComparison,
			lt:           ErrComparison,
			is:           ErrComparison,
			distinctFrom: true,
		},
		{
			name:         "bool",
			a:            true,
			b:            false,
			eq:           false,
			gt:           true,
			lt:           false,
			is:           false,
			distinctFrom: true,
		},
		{
			name:         "int-null",
			a:            int64(10),
			b:            nil,
			eq:           nil,
			gt:           nil,
			lt:           nil,
			is:           false,
			distinctFrom: true,
		},
		{
			name:         "null-null",
			a:            nil,
			b:            nil,
			eq:           nil,
			gt:           nil,
			lt:           nil,
			is:           true,
			distinctFrom: false,
		},
		// array tests
		{
			name:         "int-array",
			a:            []int64{1, 2, 3},
			b:            []int64{1, 2, 3},
			eq:           true,
			gt:           ErrComparison,
			lt:           ErrComparison,
			is:           ErrComparison,
			distinctFrom: false,
		},
		{
			name:         "text-array",
			a:            []string{"hello", "world"},
			b:            []string{"hello", "world"},
			eq:           true,
			gt:           ErrComparison,
			lt:           ErrComparison,
			is:           ErrComparison,
			distinctFrom: false,
		},
		{
			name:         "decimal-array",
			a:            []*decimal.Decimal{mustDec("1.00"), mustDec("2.00"), mustDec("3.00")},
			b:            []*decimal.Decimal{mustDec("1.00"), mustDec("2.00"), mustDec("3.00")},
			eq:           true,
			gt:           ErrComparison,
			lt:           ErrComparison,
			is:           ErrComparison,
			distinctFrom: false,
		},
		{
			name:         "text array not equal",
			a:            []string{"hello", "world"},
			b:            []string{"world", "hello"},
			eq:           false,
			gt:           ErrComparison,
			lt:           ErrComparison,
			is:           ErrComparison,
			distinctFrom: true,
		},
		{
			name:         "uuid-array",
			a:            []*types.UUID{mustUUID("550e8400-e29b-41d4-a716-446655440000")},
			b:            []*types.UUID{mustUUID("550e8400-e29b-41d4-a716-446655440000")},
			eq:           true,
			gt:           ErrComparison,
			lt:           ErrComparison,
			is:           ErrComparison,
			distinctFrom: false,
		},
		{
			name:         "blob-array",
			a:            [][]byte{[]byte("hello"), []byte("world")},
			b:            [][]byte{[]byte("hello"), []byte("world")},
			eq:           true,
			gt:           ErrComparison,
			lt:           ErrComparison,
			is:           ErrComparison,
			distinctFrom: false,
		},
		{
			name:         "bool-array",
			a:            []bool{true, false},
			b:            []bool{true, false},
			eq:           true,
			gt:           ErrComparison,
			lt:           ErrComparison,
			is:           ErrComparison,
			distinctFrom: false,
		},
		{
			name:         "int-array-null",
			a:            []int64{1, 2, 3},
			b:            nil,
			eq:           nil,
			gt:           nil,
			lt:           nil,
			is:           false,
			distinctFrom: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			makeVal := func(v any) Value {
				val, err := NewValue(v)
				require.NoError(t, err)
				return val
			}

			a := makeVal(tt.a)
			b := makeVal(tt.b)

			isErrOrResult := func(a, b Value, op ComparisonOp, want any) {
				t.Log(op.String())
				res, err := a.Compare(b, op)
				if wantErr, ok := want.(error); ok {
					require.Error(t, err)
					require.ErrorIs(t, err, wantErr)
					return
				}
				require.NoError(t, err)

				switch wantVal := want.(type) {
				default:
					require.EqualValues(t, wantVal, res.RawValue())
				case nil:
					require.True(t, res.Null())
					require.Nil(t, res.RawValue())
				case bool:
					require.Equal(t, wantVal, res.RawValue())
				case *bool:
					require.Equal(t, *wantVal, res.RawValue())
				}
			}

			isErrOrResult(a, b, lessThan, tt.lt)
			isErrOrResult(a, b, greaterThan, tt.gt)
			isErrOrResult(a, b, equal, tt.eq)
			isErrOrResult(a, b, is, tt.is)
			isErrOrResult(a, b, isDistinctFrom, tt.distinctFrom)

			// test rountripping strings
			testRoundTripParse(t, a)
			testRoundTripParse(t, b)
		})
	}
}

func Test_Cast(t *testing.T) {
	// for this test, we want to test each type and array type,
	// and ensure it can be casted to each other type and array type
	// all numerics will be precision 10, scale 5.
	// If a value is left as nil, it will expect an error when casted to that type.
	type testcase struct {
		name       string
		val        any
		intVal     any
		text       any
		boolVal    any
		decimalVal any
		uuidVal    any
		blobVal    any
		intArr     any
		textArr    any
		boolArr    any
		decimalArr any
		uuidArr    any
		blobArr    any
	}

	mDec := func(dec string) *decimal.Decimal {
		// all decimals will be precision 10, scale 5
		d, err := decimal.NewFromString(dec)
		require.NoError(t, err)

		err = d.SetPrecisionAndScale(10, 5)
		require.NoError(t, err)
		return d
	}

	mDecArr := func(decimals ...string) []*decimal.Decimal {
		var res []*decimal.Decimal
		for _, dec := range decimals {
			res = append(res, mDec(dec))
		}
		return res
	}

	tests := []testcase{
		{
			name:       "int",
			val:        int64(10),
			intVal:     int64(10),
			text:       "10",
			boolVal:    true,
			decimalVal: mDec("10.00000"),
		},
		{
			name:    "text",
			val:     "hello",
			text:    "hello",
			blobVal: []byte("hello"),
		},
		{
			name:       "text (number)",
			val:        "10",
			intVal:     10,
			text:       "10",
			decimalVal: mDec("10.00000"),
			blobVal:    []byte("10"),
		},
		{
			name:    "text (bool)",
			val:     "true",
			boolVal: true,
			text:    "true",
			blobVal: []byte("true"),
		},
		{
			name:       "text (decimal)",
			val:        "10.5",
			decimalVal: mDec("10.50000"),
			text:       "10.5",
			blobVal:    []byte("10.5"),
		},
		{
			name:    "text (uuid)",
			val:     "550e8400-e29b-41d4-a716-446655440000",
			uuidVal: mustUUID("550e8400-e29b-41d4-a716-446655440000"),
			text:    "550e8400-e29b-41d4-a716-446655440000",
			blobVal: []byte("550e8400-e29b-41d4-a716-446655440000"),
		},
		{
			name:    "bool",
			val:     true,
			boolVal: true,
			text:    "true",
			intVal:  int64(1),
		},
		{
			name:       "decimal",
			val:        mDec("10.00000"),
			decimalVal: mDec("10.00000"),
			text:       "10.00000",
			intVal:     int64(10),
		},
		{
			name:    "uuid",
			val:     mustUUID("550e8400-e29b-41d4-a716-446655440000"),
			uuidVal: mustUUID("550e8400-e29b-41d4-a716-446655440000"),
			text:    "550e8400-e29b-41d4-a716-446655440000",
			blobVal: mustUUID("550e8400-e29b-41d4-a716-446655440000").Bytes(),
		},
		{
			name:    "blob",
			val:     []byte("hello"),
			blobVal: []byte("hello"),
			text:    "hello",
		},
		{
			name:       "int-array",
			val:        []int64{1, 2, 3},
			intArr:     []int64{1, 2, 3},
			textArr:    []string{"1", "2", "3"},
			boolArr:    []bool{true, true, true},
			decimalArr: mDecArr("1", "2", "3"),
		},
		{
			name:    "text-array",
			val:     []string{"hello", "world"},
			textArr: []string{"hello", "world"},
			blobArr: [][]byte{[]byte("hello"), []byte("world")},
		},
		{
			name:    "text-array (uuid)",
			val:     []string{"550e8400-e29b-41d4-a716-446655440000"},
			uuidArr: []*types.UUID{mustUUID("550e8400-e29b-41d4-a716-446655440000")},
			textArr: []string{"550e8400-e29b-41d4-a716-446655440000"},
			blobArr: [][]byte{[]byte("550e8400-e29b-41d4-a716-446655440000")},
		},
		{
			name:    "bool-array",
			val:     []bool{true, false},
			boolArr: []bool{true, false},
			textArr: []string{"true", "false"},
			intArr:  []int64{1, 0},
		},
		{
			name:       "decimal-array",
			val:        mDecArr("1", "2", "3"),
			decimalArr: mDecArr("1", "2", "3"),
			textArr:    []string{"1.00000", "2.00000", "3.00000"},
			intArr:     []int64{1, 2, 3},
		},
		{
			name:    "uuid-array",
			val:     []*types.UUID{mustUUID("550e8400-e29b-41d4-a716-446655440000")},
			uuidArr: []*types.UUID{mustUUID("550e8400-e29b-41d4-a716-446655440000")},
			textArr: []string{"550e8400-e29b-41d4-a716-446655440000"},
			blobArr: [][]byte{mustUUID("550e8400-e29b-41d4-a716-446655440000").Bytes()},
		},
		{
			name:    "blob-array",
			val:     [][]byte{[]byte("hello"), []byte("world")},
			blobArr: [][]byte{[]byte("hello"), []byte("world")},
			textArr: []string{"hello", "world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := NewValue(tt.val)
			require.NoError(t, err)

			check := func(dataType *types.DataType, want any) {
				t.Log(dataType.String())
				if want == nil {
					want = ErrCast
				}

				res, err := val.Cast(dataType)
				if wantErr, ok := want.(error); ok {
					assert.Error(t, err)
					assert.ErrorIs(t, err, wantErr)
					return
				}
				require.NoError(t, err)

				eq(t, want, res.RawValue())
			}

			decimalType, err := types.NewDecimalType(10, 5)
			require.NoError(t, err)

			decArrType := decimalType.Copy()
			decArrType.IsArray = true

			check(types.IntType, tt.intVal)
			check(types.TextType, tt.text)
			check(types.BoolType, tt.boolVal)
			check(decimalType, tt.decimalVal)
			check(types.UUIDType, tt.uuidVal)
			check(types.BlobType, tt.blobVal)

			if intArr, ok := tt.intArr.([]int64); ok {
				tt.intArr = ptrArr(intArr)
			}
			if textArr, ok := tt.textArr.([]string); ok {
				tt.textArr = ptrArr(textArr)
			}
			if boolArr, ok := tt.boolArr.([]bool); ok {
				tt.boolArr = ptrArr(boolArr)
			}

			check(types.IntArrayType, tt.intArr)
			check(types.TextArrayType, tt.textArr)
			check(types.BoolArrayType, tt.boolArr)
			check(decArrType, tt.decimalArr)
			check(types.UUIDArrayType, tt.uuidArr)
			check(types.BlobArrayType, tt.blobArr)

			// test rountripping strings
			testRoundTripParse(t, val)
		})
	}
}

func Test_Unary(t *testing.T) {
	type testcase struct {
		name string
		val  any
		pos  any
		neg  any
		not  any
	}

	// any values left nil will expect an error when the unary operator is applied
	tests := []testcase{
		{
			name: "int",
			val:  int64(10),
			pos:  int64(10),
			neg:  int64(-10),
		},
		{
			name: "decimal",
			val:  mustDec("10.00"),
			pos:  mustDec("10.00"),
			neg:  mustDec("-10.00"),
		},
		{
			name: "text",
			// text values should not be able to be used with unary operators
		},
		{
			name: "uuid",
			val:  mustUUID("550e8400-e29b-41d4-a716-446655440000"),
			// uuid values should not be able to be used with unary operators
		},
		{
			name: "blob",
			// blob values should not be able to be used with unary operators
			val: []byte("hello"),
		},
		{
			name: "bool",
			val:  true,
			not:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := NewValue(tt.val)
			require.NoError(t, err)
			scal, ok := val.(ScalarValue)
			require.True(t, ok)

			check := func(op UnaryOp, want any) {
				if want == nil {
					want = ErrUnary
				}

				t.Log(op.String())
				res, err := scal.Unary(op)
				if wantErr, ok := want.(error); ok {
					require.Error(t, err)
					require.ErrorIs(t, err, wantErr)
					return
				}

				require.NoError(t, err)
				eq(t, want, res.RawValue())
			}

			check(pos, tt.pos)
			check(neg, tt.neg)
			check(not, tt.not)

			// test rountripping strings
			testRoundTripParse(t, val)
		})
	}
}

func Test_Array(t *testing.T) {
	type testcase struct {
		name    string
		vals    []any
		wantErr error
	}

	// all values will be put into an array.
	// unless the wantErr is specified, it will expect the array to be created successfully

	tests := []testcase{
		{
			name: "int",
			vals: []any{int64(1), int64(2), int64(3)},
		},
		{
			name: "decimal",
			vals: []any{mustDec("1.00"), mustDec("2.00"), mustDec("3.00")},
		},
		{
			name: "text",
			vals: []any{"hello", "world"},
		},
		{
			name: "uuid",
			vals: []any{mustUUID("550e8400-e29b-41d4-a716-446655440000"), mustUUID("550e8400-e29b-41d4-a716-446655440001")},
		},
		{
			name: "blob",
			vals: []any{[]byte("hello"), []byte("world")},
		},
		{
			name: "bool",
			vals: []any{true, false},
		},
		{
			name:    "mixed",
			vals:    []any{int64(1), "hello"},
			wantErr: ErrArrayMixedTypes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.vals) == 0 {
				t.Fatal("no values provided")
			}

			var vals []ScalarValue
			for _, v := range tt.vals {
				val, err := NewValue(v)
				require.NoError(t, err)
				vals = append(vals, val.(ScalarValue))
			}

			res, err := vals[0].Array(vals[1:]...)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			for i := range res.Len() {
				s, err := res.Index(i + 1) // 1-indexed
				require.NoError(t, err)

				eq(t, tt.vals[i], s.RawValue())
			}

			// we will now set them all to nulls and test that the array is created successfully
			dt := vals[0].Type()
			for i := range vals {
				err = res.Set(int32(i+1), newNull(dt).(ScalarValue))
				require.NoError(t, err)
			}

			for i := range res.Len() {
				s, err := res.Index(i + 1) // 1-indexed
				require.NoError(t, err)

				isNull := s.Null()
				_ = isNull
				require.True(t, s.Null())
				require.Nil(t, s.RawValue())
			}

			// test rountripping strings
			testRoundTripParse(t, res)
		})
	}
}

// ptrArr is a helper function that converts a slice of values to a slice of pointers to those values.
// Since Kwil returns pointers to account for nulls, we need to convert the slice of values to pointers
func ptrArr[T any](arr []T) []*T {
	var res []*T
	for i := range arr {
		res = append(res, &arr[i])
	}
	return res
}

func mustDec(dec string) *decimal.Decimal {
	d, err := decimal.NewFromString(dec)
	if err != nil {
		panic(err)
	}
	return d
}

func mustUUID(s string) *types.UUID {
	u, err := types.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}

// testRoundTripParse is a helper function that formats a value to a string, then parses it back to a value.
// It is meant to be used within these other tests.
func testRoundTripParse(t *testing.T, v Value) {
	if v.Null() {
		return
	}
	str, err := valueToString(v)
	require.NoError(t, err)

	val2, err := parseValue(str, v.Type())
	require.NoError(t, err)

	equal, err := v.Compare(val2, equal)
	require.NoError(t, err)

	if !equal.RawValue().(bool) {
		t.Fatalf("values not equal: %v != %v", v.RawValue(), val2.RawValue())
	}
}
