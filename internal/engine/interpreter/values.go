package interpreter

import (
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
)

// ValueMapping maps Go types and Kwil native types.
type ValueMapping struct {
	// KwilType is the Kwil type that the value maps to.
	// It will ignore the metadata of the type.
	KwilType *types.DataType
	// ZeroValue creates a zero-value of the type.
	ZeroValue func() (Value, error)
}

var (
	kwilTypeToValue = map[struct {
		name    string
		isArray bool
	}]ValueMapping{}
)

func registerValueMapping(ms ...ValueMapping) {
	for _, m := range ms {
		k := struct {
			name    string
			isArray bool
		}{
			name:    m.KwilType.Name,
			isArray: m.KwilType.IsArray,
		}

		_, ok := kwilTypeToValue[k]
		if ok {
			panic(fmt.Sprintf("type %s already registered", m.KwilType.Name))
		}

		kwilTypeToValue[k] = m
	}
}

func init() {
	registerValueMapping(
		ValueMapping{
			KwilType: types.IntType,
			ZeroValue: func() (Value, error) {
				return newInt(0), nil
			},
		},
		ValueMapping{
			KwilType: types.TextType,
			ZeroValue: func() (Value, error) {
				return newText(""), nil
			},
		},
		ValueMapping{
			KwilType: types.BoolType,
			ZeroValue: func() (Value, error) {
				return newBool(false), nil
			},
		},
		ValueMapping{
			KwilType: types.BlobType,
			ZeroValue: func() (Value, error) {
				return newBlob([]byte{}), nil
			},
		},
		ValueMapping{
			KwilType: types.UUIDType,
			ZeroValue: func() (Value, error) {
				return newUUID(&types.UUID{}), nil
			},
		},
		ValueMapping{
			KwilType: types.DecimalType,
			ZeroValue: func() (Value, error) {
				dec, err := decimal.NewFromString("0")
				if err != nil {
					return nil, err
				}
				return newDec(dec), nil
			},
		},
		ValueMapping{
			KwilType: types.IntArrayType,
			ZeroValue: func() (Value, error) {
				return &IntArrayValue{
					Array: pgtype.Array[pgtype.Int8]{},
				}, nil
			},
		},
		ValueMapping{
			KwilType: types.TextArrayType,
			ZeroValue: func() (Value, error) {
				return &TextArrayValue{
					Array: pgtype.Array[pgtype.Text]{},
				}, nil
			},
		},
		ValueMapping{
			KwilType: types.BoolArrayType,
			ZeroValue: func() (Value, error) {
				return &BoolArrayValue{
					Array: pgtype.Array[pgtype.Bool]{},
				}, nil
			},
		},
		ValueMapping{
			KwilType: types.BlobArrayType,
			ZeroValue: func() (Value, error) {
				return &BlobArrayValue{
					Array: pgtype.Array[pgtype.PreallocBytes]{},
				}, nil
			},
		},
		ValueMapping{
			KwilType: types.DecimalArrayType,
			ZeroValue: func() (Value, error) {
				return &DecimalArrayValue{
					Array: pgtype.Array[pgtype.Numeric]{},
				}, nil
			},
		},
	)
}

// NewZeroValue creates a new zero value of the given type.
func NewZeroValue(t *types.DataType) (Value, error) {
	m, ok := kwilTypeToValue[struct {
		name    string
		isArray bool
	}{
		name:    t.Name,
		isArray: t.IsArray,
	}]
	if !ok {
		return nil, fmt.Errorf("type %s not found", t.Name)
	}

	return m.ZeroValue()
}

// Value is a value that can be compared, used in arithmetic operations,
// and have unary operations applied to it.
type Value interface {
	// Compare compares the variable with another variable using the given comparison operator.
	// It will return a boolean value, or null either of the variables is null.
	Compare(v Value, op ComparisonOp) (*BoolValue, error)
	// Type returns the type of the variable.
	Type() *types.DataType
	// RawValue returns the value of the variable.
	// This is one of: nil, int64, string, bool, []byte, *types.UUID, *decimal.Decimal,
	// []*int64, []*string, []*bool, [][]byte, []*decimal.Decimal, []*types.UUID
	RawValue() any
	// Cast casts the variable to the given type.
	Cast(t *types.DataType) (Value, error)
	// Null returns true if the variable is null.
	Null() bool
}

// ScalarValue is a scalar value that can be computed on and have unary operations applied to it.
type ScalarValue interface {
	Value
	// Arithmetic performs an arithmetic operation on the variable with another variable.
	Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error)
	// Unary applies a unary operation to the variable.
	Unary(op UnaryOp) (ScalarValue, error)
	// Array creates an array from this scalar value and any other scalar values.
	Array(v ...ScalarValue) (ArrayValue, error)
}

// ArrayValue is an array value that can be compared and have unary operations applied to it.
type ArrayValue interface {
	Value
	// Len returns the length of the array.
	Len() int
	// Index returns the value at the given index.
	// If the index is out of bounds, an error is returned.
	// All indexing is 1-based.
	Index(i int64) (ScalarValue, error)
	// Set sets the value at the given index.
	// If the index is out of bounds, enough space is allocated to set the value.
	// This matches the behavior of Postgres.
	// All indexing is 1-based.
	Set(i int64, v ScalarValue) error
}

// safePtrArr makes a pointer array avoiding closure issues.
func safePtrArr[T any](a []T) []*T {
	res := make([]*T, len(a))
	for i, val := range a {
		val2 := val // create a new variable to avoid closure issues
		res[i] = &val2
	}

	return res
}

func newValidArr[T any](a []T) pgtype.Array[T] {
	return pgtype.Array[T]{
		Elements: a,
		Dims:     []pgtype.ArrayDimension{{Length: int32(len(a)), LowerBound: 1}},
		Valid:    true,
	}
}

// NewValue creates a new Value from the given any val.
func NewValue(v any) (Value, error) {
	switch v := v.(type) {
	case int64:
		return newInt(v), nil
	case int:
		return newInt(int64(v)), nil
	case string:
		return newText(v), nil
	case bool:
		return newBool(v), nil
	case []byte:
		return newBlob(v), nil
	case *types.UUID:
		return newUUID(v), nil
	case types.UUID:
		return newUUID(&v), nil
	case *decimal.Decimal:
		return newDec(v), nil
	case decimal.Decimal:
		return newDec(&v), nil
	case []int64:
		pgInts := make([]pgtype.Int8, len(v))
		for i, val := range v {
			pgInts[i].Int64 = val
			pgInts[i].Valid = true
		}

		return &IntArrayValue{
			Array: newValidArr(pgInts),
		}, nil
	case []*int64:
		pgInts := make([]pgtype.Int8, len(v))
		for i, val := range v {
			if val == nil {
				pgInts[i].Valid = false
			} else {
				pgInts[i].Int64 = *val
				pgInts[i].Valid = true
			}
		}
		return &IntArrayValue{
			Array: newValidArr(pgInts),
		}, nil
	case []string:
		pgTexts := make([]pgtype.Text, len(v))
		for i, val := range v {
			pgTexts[i].String = val
			pgTexts[i].Valid = true
		}

		return &TextArrayValue{
			Array: newValidArr(pgTexts),
		}, nil
	case []*string:
		pgTexts := make([]pgtype.Text, len(v))
		for i, val := range v {
			if val == nil {
				pgTexts[i].Valid = false
			} else {
				pgTexts[i].String = *val
				pgTexts[i].Valid = true
			}
		}

		return &TextArrayValue{
			Array: newValidArr(pgTexts),
		}, nil
	case []bool:
		pgBools := make([]pgtype.Bool, len(v))
		for i, val := range v {
			pgBools[i].Bool = val
			pgBools[i].Valid = true
		}

		return &BoolArrayValue{
			Array: newValidArr(pgBools),
		}, nil
	case []*bool:
		pgBools := make([]pgtype.Bool, len(v))
		for i, val := range v {
			if val == nil {
				pgBools[i].Valid = false
			} else {
				pgBools[i].Bool = *val
				pgBools[i].Valid = true
			}
		}

		return &BoolArrayValue{
			Array: newValidArr(pgBools),
		}, nil
	case [][]byte:
		pgBlobs := make([]pgtype.PreallocBytes, len(v))
		for i, val := range v {
			pgBlobs[i] = val
		}

		return &BlobArrayValue{
			Array: newValidArr(pgBlobs),
		}, nil
	case []*[]byte:
		pgBlobs := make([]pgtype.PreallocBytes, len(v))
		for i, val := range v {
			if val == nil {
				pgBlobs[i] = nil
			} else {
				pgBlobs[i] = *val
			}
		}

		return &BlobArrayValue{
			Array: newValidArr(pgBlobs),
		}, nil
	case []*decimal.Decimal:
		pgDecs := make([]pgtype.Numeric, len(v))
		for i, val := range v {
			pgDecs[i] = pgTypeFromDec(val)
		}

		return &DecimalArrayValue{
			Array: newValidArr(pgDecs),
		}, nil
	case []*types.UUID:
		pgUUIDs := make([]pgtype.UUID, len(v))
		for i, val := range v {
			if val == nil {
				pgUUIDs[i].Valid = false
			} else {
				pgUUIDs[i].Bytes = *val
				pgUUIDs[i].Valid = true
			}
		}

		return &UuidArrayValue{
			Array: newValidArr(pgUUIDs),
		}, nil
	default:
		return nil, fmt.Errorf("unexpected type %T", v)
	}
}

func makeTypeErr(left, right Value) error {
	return fmt.Errorf("%w: left: %s right: %s", ErrTypeMismatch, left.Type(), right.Type())
}

func newInt(i int64) *IntValue {
	return &IntValue{
		Int8: pgtype.Int8{
			Int64: i,
			Valid: true,
		},
	}
}

type IntValue struct {
	pgtype.Int8
}

func (i *IntValue) Null() bool {
	return !i.Valid
}

func (v *IntValue) Compare(v2 Value, op ComparisonOp) (*BoolValue, error) {
	if res, early := nullCmp(v, v2, op); early {
		return res, nil
	}

	val2, ok := v2.(*IntValue)
	if !ok {
		return nil, makeTypeErr(v, v2)
	}

	var b bool
	switch op {
	case equal:
		b = v.Int64 == val2.Int64
	case lessThan:
		b = v.Int64 < val2.Int64
	case greaterThan:
		b = v.Int64 > val2.Int64
	case isDistinctFrom:
		b = v.Int64 != val2.Int64
	default:
		return nil, fmt.Errorf("cannot compare int with operator id %d", op)
	}

	return newBool(b), nil
}

// nullCmp is a helper function for comparing null values.
// It takes two values and a comparison operator.
// If the operator is IS or IS DISTINCT FROM, it will return a boolean value
// based on the comparison of the two values.
// If the operator is any other operator and either of the values is null,
// it will return a null value.
func nullCmp(a, b Value, op ComparisonOp) (*BoolValue, bool) {
	// if it is isDistinctFrom or is, we should handle nulls
	// Otherwise, if either is a null, we return early because we cannot compare
	// a null value with a non-null value.
	if op == isDistinctFrom {
		if a.Null() && b.Null() {
			return newBool(false), true
		}
		if a.Null() || b.Null() {
			return newBool(true), true
		}

		// otherwise, we let equality handle it
	}

	if op == is {
		if a.Null() && b.Null() {
			return newBool(true), true
		}
		if a.Null() || b.Null() {
			return newBool(false), true
		}
	}

	if a.Null() || b.Null() {
		// the type of this null doesnt really matter.
		return newNull(types.BoolType).(*BoolValue), true
	}

	return nil, false
}

// checks if any value is null. If so, it will return the null value.
func checkScalarNulls(v ...ScalarValue) (ScalarValue, bool) {
	for _, val := range v {
		if val.Null() {
			return val, true
		}
	}

	return nil, false
}

func (i *IntValue) Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error) {
	if res, early := checkScalarNulls(i, v); early {
		return res, nil
	}

	val2, ok := v.(*IntValue)
	if !ok {
		return nil, makeTypeErr(i, v)
	}

	var r int64

	switch op {
	case add:
		r = i.Int64 + val2.Int64
	case sub:
		r = i.Int64 - val2.Int64
	case mul:
		r = i.Int64 * val2.Int64
	case div:
		if val2.Int64 == 0 {
			return nil, fmt.Errorf("cannot divide by zero")
		}
		r = i.Int64 / val2.Int64
	case mod:
		if val2.Int64 == 0 {
			return nil, fmt.Errorf("cannot modulo by zero")
		}
		r = i.Int64 % val2.Int64
	default:
		return nil, fmt.Errorf("cannot perform arithmetic operation id %d on type int", op)
	}

	return &IntValue{
		Int8: pgtype.Int8{
			Int64: r,
			Valid: true,
		},
	}, nil
}

func (i *IntValue) Unary(op UnaryOp) (ScalarValue, error) {
	if i.Null() {
		return i, nil
	}

	switch op {
	case neg:
		return &IntValue{Int8: pgtype.Int8{Int64: -i.Int64, Valid: true}}, nil
	case not:
		return nil, fmt.Errorf("cannot apply logical NOT to an integer")
	case pos:
		return i, nil
	default:
		return nil, fmt.Errorf("unknown unary operator: %d", op)
	}
}

func (i *IntValue) Type() *types.DataType {
	return types.IntType
}

func (i *IntValue) RawValue() any {
	if !i.Valid {
		return nil
	}

	return i.Int64
}

func (i *IntValue) Array(v ...ScalarValue) (ArrayValue, error) {
	pgtArr := make([]pgtype.Int8, len(v)+1)
	pgtArr[0] = i.Int8
	for j, val := range v {
		if intVal, ok := val.(*IntValue); !ok {
			return nil, makeTypeErr(i, val)
		} else {
			pgtArr[j+1] = intVal.Int8
		}
	}

	arr := newValidArr(pgtArr)

	return &IntArrayValue{
		Array: arr,
	}, nil
}

func (i *IntValue) Cast(t *types.DataType) (Value, error) {
	if i.Null() {
		return newNull(t), nil
	}

	// we check for decimal first since type switching on it
	// doesn't work, since it has precision and scale
	if t.Name == types.DecimalStr {
		dec, err := decimal.NewFromString(fmt.Sprint(i.Int64))
		if err != nil {
			return nil, err
		}

		return newDec(dec), nil
	}

	switch t {
	case types.IntType:
		return i, nil
	case types.TextType:
		return newText(fmt.Sprint(i.Int64)), nil
	case types.BoolType:
		return newBool(i.Int64 != 0), nil
	default:
		return nil, fmt.Errorf("cannot cast int to %s", t)
	}
}

// newNull creates a new null value of the given type.
func newNull(t *types.DataType) Value {
	switch t {
	case types.IntType:
		return &IntValue{
			Int8: pgtype.Int8{
				Valid: false,
			},
		}
	case types.TextType:
		return &TextValue{
			Text: pgtype.Text{
				Valid: false,
			},
		}
	case types.BoolType:
		return &BoolValue{
			Bool: pgtype.Bool{
				Valid: false,
			},
		}
	case types.BlobType:
		return &BlobValue{
			PreallocBytes: nil,
		}
	case types.UUIDType:
		return &UUIDValue{
			UUID: pgtype.UUID{
				Valid: false,
			},
		}
	case types.DecimalType:
		return newDec(nil)
	case types.IntArrayType:
		return &IntArrayValue{
			Array: pgtype.Array[pgtype.Int8]{Valid: false},
		}
	case types.TextArrayType:
		return &TextArrayValue{
			Array: pgtype.Array[pgtype.Text]{Valid: false},
		}
	case types.BoolArrayType:
		return &BoolArrayValue{
			Array: pgtype.Array[pgtype.Bool]{Valid: false},
		}
	case types.BlobArrayType:
		return &BlobArrayValue{
			Array: pgtype.Array[pgtype.PreallocBytes]{Valid: false},
		}
	case types.DecimalArrayType:
		return &DecimalArrayValue{
			Array: pgtype.Array[pgtype.Numeric]{Valid: false},
		}
	case types.UUIDArrayType:
		return &UuidArrayValue{
			Array: pgtype.Array[pgtype.UUID]{Valid: false},
		}
	default:
		panic("unknown type")
	}
}

func newText(s string) *TextValue {
	return &TextValue{
		Text: pgtype.Text{
			String: s,
			Valid:  true,
		},
	}
}

type TextValue struct {
	pgtype.Text
}

func (t *TextValue) Null() bool {
	return !t.Valid
}

func (s *TextValue) Compare(v Value, op ComparisonOp) (*BoolValue, error) {
	if res, early := nullCmp(s, v, op); early {
		return res, nil
	}

	val2, ok := v.(*TextValue)
	if !ok {
		return nil, makeTypeErr(s, v)
	}

	var b bool
	switch op {
	case equal:
		b = s.String == val2.String
	case lessThan:
		b = s.String < val2.String
	case greaterThan:
		b = s.String > val2.String
	case isDistinctFrom:
		b = s.String != val2.String
	default:
		return nil, fmt.Errorf("unknown comparison operator: %d", op)
	}

	return newBool(b), nil
}

func (s *TextValue) Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error) {
	if res, early := checkScalarNulls(s, v); early {
		return res, nil
	}

	val2, ok := v.(*TextValue)
	if !ok {
		return nil, makeTypeErr(s, v)
	}

	if op == concat {
		return newText(s.String + val2.String), nil
	}

	return nil, fmt.Errorf("cannot perform arithmetic operation id %d on type string", op)
}

func (s *TextValue) Unary(op UnaryOp) (ScalarValue, error) {
	return nil, fmt.Errorf("cannot perform unary operation on string")
}

func (s *TextValue) Type() *types.DataType {
	return types.TextType
}

func (s *TextValue) RawValue() any {
	if !s.Valid {
		return nil
	}

	return s.String
}

func (s *TextValue) Array(v ...ScalarValue) (ArrayValue, error) {
	pgtArr := make([]pgtype.Text, len(v)+1)
	pgtArr[0] = s.Text
	for j, val := range v {
		if textVal, ok := val.(*TextValue); !ok {
			return nil, makeTypeErr(s, val)
		} else {
			pgtArr[j+1] = textVal.Text
		}
	}

	arr := newValidArr(pgtArr)

	return &TextArrayValue{
		Array: arr,
	}, nil
}

func (s *TextValue) Cast(t *types.DataType) (Value, error) {
	if s.Null() {
		return newNull(t), nil
	}

	if t.Name == types.DecimalStr {
		dec, err := decimal.NewFromString(s.String)
		if err != nil {
			return nil, err
		}

		return newDec(dec), nil
	}

	switch t {
	case types.IntType:
		i, err := strconv.ParseInt(s.String, 10, 64)
		if err != nil {
			return nil, err
		}

		return newInt(int64(i)), nil
	case types.TextType:
		return s, nil
	case types.BoolType:
		b, err := strconv.ParseBool(s.String)
		if err != nil {
			return nil, err
		}

		return newBool(b), nil
	case types.UUIDType:
		u, err := types.ParseUUID(s.String)
		if err != nil {
			return nil, err
		}

		return newUUID(u), nil
	default:
		return nil, fmt.Errorf("cannot cast string to %s", t)
	}
}

func newBool(b bool) *BoolValue {
	return &BoolValue{
		Bool: pgtype.Bool{
			Bool:  b,
			Valid: true,
		},
	}
}

type BoolValue struct {
	pgtype.Bool
}

func (b *BoolValue) Null() bool {
	return !b.Valid
}

func (b *BoolValue) Compare(v Value, op ComparisonOp) (*BoolValue, error) {
	if res, early := nullCmp(b, v, op); early {
		return res, nil
	}

	val2, ok := v.(*BoolValue)
	if !ok {
		return nil, makeTypeErr(b, v)
	}

	var b2 bool
	switch op {
	case equal:
		b2 = b.Bool.Bool == val2.Bool.Bool
	case isDistinctFrom:
		b2 = b.Bool.Bool != val2.Bool.Bool
	default:
		return nil, fmt.Errorf("unknown comparison operator: %d", op)
	}

	return newBool(b2), nil
}

func (b *BoolValue) Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error) {
	return nil, fmt.Errorf("%w: cannot perform arithmetic operation on bool", ErrArithmetic)
}

func (b *BoolValue) Unary(op UnaryOp) (ScalarValue, error) {
	if b.Null() {
		return b, nil
	}

	switch op {
	case not:
		return newBool(!b.Bool.Bool), nil
	default:
		return nil, fmt.Errorf("unexpected operator id %d for bool", op)
	}
}

func (b *BoolValue) Type() *types.DataType {
	return types.BoolType
}

func (b *BoolValue) RawValue() any {
	return b.Bool.Bool
}

func (b *BoolValue) Array(v ...ScalarValue) (ArrayValue, error) {
	pgtArr := make([]pgtype.Bool, len(v)+1)
	pgtArr[0] = b.Bool
	for j, val := range v {
		if boolVal, ok := val.(*BoolValue); !ok {
			return nil, makeTypeErr(b, val)
		} else {
			pgtArr[j+1] = boolVal.Bool
		}
	}

	arr := newValidArr(pgtArr)

	return &BoolArrayValue{
		Array: arr,
	}, nil
}

func (b *BoolValue) Cast(t *types.DataType) (Value, error) {
	if b.Null() {
		return newNull(t), nil
	}

	switch t {
	case types.IntType:
		if b.Bool.Bool {
			return newInt(1), nil
		}

		return newInt(0), nil
	case types.TextType:
		return newText(strconv.FormatBool(b.Bool.Bool)), nil
	case types.BoolType:
		return b, nil
	default:
		return nil, fmt.Errorf("cannot cast bool to %s", t)
	}
}

func newBlob(b []byte) *BlobValue {
	return &BlobValue{
		PreallocBytes: b,
	}
}

type BlobValue struct {
	pgtype.PreallocBytes
}

func (b *BlobValue) Null() bool {
	return b.PreallocBytes == nil
}

func (b *BlobValue) Compare(v Value, op ComparisonOp) (*BoolValue, error) {
	if res, early := nullCmp(b, v, op); early {
		return res, nil
	}

	val2, ok := v.(*BlobValue)
	if !ok {
		return nil, makeTypeErr(b, v)
	}

	var b2 bool
	switch op {
	case equal:
		b2 = string(b.PreallocBytes) == string(val2.PreallocBytes)
	case isDistinctFrom:
		b2 = string(b.PreallocBytes) != string(val2.PreallocBytes)
	default:
		return nil, fmt.Errorf("unknown comparison operator: %d", op)
	}

	return newBool(b2), nil
}

func (b *BlobValue) Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error) {
	return nil, fmt.Errorf("%w: cannot perform arithmetic operation on blob", ErrArithmetic)
}

func (b *BlobValue) Unary(op UnaryOp) (ScalarValue, error) {
	return nil, fmt.Errorf("cannot perform unary operation on blob")
}

func (b *BlobValue) Type() *types.DataType {
	return types.BlobType
}

func (b *BlobValue) RawValue() any {
	return b.PreallocBytes
}

func (b *BlobValue) Array(v ...ScalarValue) (ArrayValue, error) {
	pgtArr := make([]pgtype.PreallocBytes, len(v)+1)
	pgtArr[0] = b.PreallocBytes
	for j, val := range v {
		if blobVal, ok := val.(*BlobValue); !ok {
			return nil, makeTypeErr(b, val)
		} else {
			pgtArr[j+1] = blobVal.PreallocBytes
		}
	}

	arr := newValidArr(pgtArr)

	return &BlobArrayValue{
		Array: arr,
	}, nil
}

func (b *BlobValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.IntType:
		i, err := strconv.ParseInt(string(b.PreallocBytes), 10, 64)
		if err != nil {
			return nil, err
		}

		return newInt(i), nil
	case types.TextType:
		return newText(string(b.PreallocBytes)), nil
	default:
		return nil, fmt.Errorf("cannot cast blob to %s", t)
	}
}

func newUUID(u *types.UUID) *UUIDValue {
	if u == nil {
		return &UUIDValue{
			UUID: pgtype.UUID{
				Valid: false,
			},
		}
	}
	return &UUIDValue{
		UUID: pgtype.UUID{
			Bytes: *u,
			Valid: true,
		},
	}
}

type UUIDValue struct {
	pgtype.UUID
}

func (u *UUIDValue) Null() bool {
	return !u.Valid
}

func (u *UUIDValue) Compare(v Value, op ComparisonOp) (*BoolValue, error) {
	if res, early := nullCmp(u, v, op); early {
		return res, nil
	}

	val2, ok := v.(*UUIDValue)
	if !ok {
		return nil, makeTypeErr(u, v)
	}

	var b bool
	switch op {
	case equal:
		b = u.Bytes == val2.Bytes
	case isDistinctFrom:
		b = u.Bytes != val2.Bytes
	default:
		return nil, fmt.Errorf("unknown comparison operator: %d", op)
	}

	return newBool(b), nil
}

func (u *UUIDValue) Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error) {
	return nil, fmt.Errorf("%w: cannot perform arithmetic operation on uuid", ErrArithmetic)
}

func (u *UUIDValue) Unary(op UnaryOp) (ScalarValue, error) {
	return nil, fmt.Errorf("cannot perform unary operation on uuid")
}

func (u *UUIDValue) Type() *types.DataType {
	return types.UUIDType
}

func (u *UUIDValue) RawValue() any {
	if !u.Valid {
		return nil
	}

	// kwil always handled uuids as pointers
	u2 := types.UUID(u.Bytes)
	return &u2
}

func (u *UUIDValue) Array(v ...ScalarValue) (ArrayValue, error) {
	pgtArr := make([]pgtype.UUID, len(v)+1)
	pgtArr[0] = u.UUID
	for j, val := range v {
		if uuidVal, ok := val.(*UUIDValue); !ok {
			return nil, makeTypeErr(u, val)
		} else {
			pgtArr[j+1] = uuidVal.UUID
		}
	}

	arr := newValidArr(pgtArr)

	return &UuidArrayValue{
		Array: arr,
	}, nil
}

func (u *UUIDValue) Cast(t *types.DataType) (Value, error) {
	if u.Null() {
		return newNull(t), nil
	}

	switch t {
	case types.TextType:
		return newText(types.UUID(u.Bytes).String()), nil
	case types.BlobType:
		return newBlob(u.Bytes[:]), nil
	default:
		return nil, fmt.Errorf("cannot cast uuid to %s", t)
	}
}

func pgTypeFromDec(d *decimal.Decimal) pgtype.Numeric {
	if d == nil {
		return pgtype.Numeric{
			Valid: false,
		}
	}
	if d.NaN() {
		return pgtype.Numeric{
			NaN:   true,
			Valid: true,
		}
	}

	bigint := d.BigInt()
	// cockroach's APD library tracks negativity outside of the BigInt,
	// so here we need to check if the decimal is negative, and if so,
	// apply it to the big int we are putting into the pgtype.
	if d.IsNegative() {
		bigint = bigint.Neg(bigint)
	}

	return pgtype.Numeric{
		Int:   bigint,
		Exp:   d.Exp(),
		Valid: true,
	}
}

func decFromPgType(n pgtype.Numeric) (*decimal.Decimal, error) {
	if n.NaN {
		return decimal.NewNaN(), nil
	}
	if !n.Valid {
		// we should never get here, but just in case
		return nil, fmt.Errorf("internal bug: null decimal")
	}

	return decimal.NewFromBigInt(n.Int, int32(n.Exp))
}

func newDec(d *decimal.Decimal) *DecimalValue {
	if d == nil {
		return &DecimalValue{
			Numeric: pgtype.Numeric{
				Valid: false,
			},
		}
	}

	return &DecimalValue{
		Numeric: pgTypeFromDec(d),
	}
}

type DecimalValue struct {
	pgtype.Numeric
}

func (d *DecimalValue) Null() bool {
	return !d.Valid
}

func (d *DecimalValue) dec() (*decimal.Decimal, error) {
	if d.NaN {
		return nil, fmt.Errorf("NaN")
	}
	if !d.Valid {
		// we should never get here, but just in case
		return nil, fmt.Errorf("internal bug: null decimal")
	}

	d2, err := decimal.NewFromBigInt(d.Int, int32(d.Exp))
	if err != nil {
		return nil, err
	}

	return d2, nil
}

func (d *DecimalValue) Compare(v Value, op ComparisonOp) (*BoolValue, error) {
	if res, early := nullCmp(d, v, op); early {
		return res, nil
	}

	val2, ok := v.(*DecimalValue)
	if !ok {
		return nil, makeTypeErr(d, v)
	}

	dec1, err := d.dec()
	if err != nil {
		return nil, err
	}

	dec2, err := val2.dec()
	if err != nil {
		return nil, err
	}

	res, err := dec1.Cmp(dec2)
	if err != nil {
		return nil, err
	}

	return cmpIntegers(res, 0, op)
}

func (d *DecimalValue) Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error) {
	if res, early := checkScalarNulls(d, v); early {
		return res, nil
	}

	// we check they are both decimal, but we don't check the precision and scale
	// because our decimal library will calculate with higher precision and scale anyways.
	if v.Type().Name != d.Type().Name {
		return nil, makeTypeErr(d, v)
	}

	val2, ok := v.(*DecimalValue)
	if !ok {
		return nil, makeTypeErr(d, v)
	}

	dec1, err := d.dec()
	if err != nil {
		return nil, err
	}

	dec2, err := val2.dec()
	if err != nil {
		return nil, err
	}

	var d2 *decimal.Decimal
	switch op {
	case add:
		d2, err = decimal.Add(dec1, dec2)
	case sub:
		d2, err = decimal.Sub(dec1, dec2)
	case mul:
		d2, err = decimal.Mul(dec1, dec2)
	case div:
		d2, err = decimal.Div(dec1, dec2)
	case mod:
		d2, err = decimal.Mod(dec1, dec2)
	default:
		return nil, fmt.Errorf("%w: unexpected operator id %d for decimal", ErrArithmetic, op)
	}
	if err != nil {
		return nil, err
	}

	return newDec(d2), nil
}

func (d *DecimalValue) Unary(op UnaryOp) (ScalarValue, error) {
	if d.Null() {
		return d, nil
	}

	switch op {
	case neg:
		dec, err := d.dec()
		if err != nil {
			return nil, err
		}

		err = dec.Neg()
		if err != nil {
			return nil, err
		}

		return newDec(dec), nil
	case pos:
		return d, nil
	default:
		return nil, fmt.Errorf("unexpected operator id %d for decimal", op)
	}
}

func (d *DecimalValue) Type() *types.DataType {
	dec, err := d.dec()
	if err != nil {
		panic(err)
	}

	res, err := types.NewDecimalType(dec.Precision(), dec.Scale())
	if err != nil {
		panic(err)
	}

	return res
}

func (d *DecimalValue) RawValue() any {
	dec, err := d.dec()
	if err != nil {
		return nil
	}

	return dec
}

func (d *DecimalValue) Array(v ...ScalarValue) (ArrayValue, error) {
	pgtArr := make([]pgtype.Numeric, len(v)+1)
	pgtArr[0] = d.Numeric
	for j, val := range v {
		if decVal, ok := val.(*DecimalValue); !ok {
			return nil, makeTypeErr(d, val)
		} else {
			pgtArr[j+1] = decVal.Numeric
		}
	}

	arr := newValidArr(pgtArr)

	return &DecimalArrayValue{
		Array: arr,
	}, nil
}

func (d *DecimalValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.IntType:
		dec, err := d.dec()
		if err != nil {
			return nil, err
		}

		i, err := dec.Int64()
		if err != nil {
			return nil, err
		}

		return newInt(i), nil
	case types.TextType:
		dec, err := d.dec()
		if err != nil {
			return nil, err
		}

		return newText(dec.String()), nil
	default:
		return nil, fmt.Errorf("cannot cast decimal to %s", t)
	}
}

func newIntArr(v []*int64) *IntArrayValue {
	pgInts := make([]pgtype.Int8, len(v))
	for i, val := range v {
		if val == nil {
			pgInts[i].Valid = false
		} else {
			pgInts[i].Int64 = *val
			pgInts[i].Valid = true
		}
	}

	return &IntArrayValue{
		Array: newValidArr(pgInts),
	}
}

type IntArrayValue struct {
	pgtype.Array[pgtype.Int8]
}

func (a *IntArrayValue) Null() bool {
	return !a.Valid
}

func (a *IntArrayValue) Compare(v Value, op ComparisonOp) (*BoolValue, error) {
	return cmpArrs(a, v, op)
}

func (a *IntArrayValue) Len() int {
	return len(a.Elements)
}

func (a *IntArrayValue) Index(i int64) (ScalarValue, error) {
	if i < 1 || i > int64(a.Len()) {
		return nil, fmt.Errorf("index out of bounds")
	}

	return &IntValue{a.Elements[i-1]}, nil // indexing is 1-based
}

// allocArr checks that the array has index i, and if not, it allocates enough space to set the value.
func allocArr[T any](p *pgtype.Array[T], i int64) error {
	if i < 1 {
		return fmt.Errorf("index out of bounds")
	}

	if i > int64(len(p.Elements)) {
		// Allocate enough space to set the value.
		// This matches the behavior of Postgres.
		newVal := make([]T, i)
		copy(newVal, p.Elements)
		p.Elements = newVal
		p.Dims[0] = pgtype.ArrayDimension{
			Length:     int32(i),
			LowerBound: 1,
		}
	}

	return nil
}

func (a *IntArrayValue) Set(i int64, v ScalarValue) error {
	// we do not need to worry about nulls here. Postgres will automatically make an array
	// not null if we set a value in it.
	// to test it:
	// CREATE TABLE test (arr int[]);
	// INSERT INTO test VALUES (NULL);
	// UPDATE test SET arr[1] = 1;
	err := allocArr(&a.Array, i)
	if err != nil {
		return err
	}

	val, ok := v.(*IntValue)
	if !ok {
		return fmt.Errorf("cannot set non-int value in int array")
	}

	a.Elements[i-1] = val.Int8
	return nil
}

func (a *IntArrayValue) Type() *types.DataType {
	return types.IntArrayType
}

func (a *IntArrayValue) RawValue() any {
	if !a.Valid {
		return nil
	}

	var res []*int64
	for _, v := range a.Elements {
		if v.Valid {
			res = append(res, &v.Int64)
		} else {
			res = append(res, nil)
		}
	}

	return res
}

func (a *IntArrayValue) Cast(t *types.DataType) (Value, error) {
	if a.Null() {
		return newNull(t), nil
	}

	switch t {
	case types.TextArrayType:
		return castArr(a, func(i int64) (string, error) { return strconv.FormatInt(i, 10), nil }, newTextArrayValue)
	case types.BoolArrayType:
		return castArr(a, func(i int64) (bool, error) { return i != 0, nil }, newBoolArrayValue)
	case types.DecimalArrayType:
		return castArrWithPtr(a, func(i int64) (*decimal.Decimal, error) { return decimal.NewFromString(strconv.FormatInt(i, 10)) }, newDecimalArrayValue)
	default:
		return nil, fmt.Errorf("cannot cast int array to %s", t)
	}
}

func newTextArrayValue(s []*string) *TextArrayValue {
	vals := make([]pgtype.Text, len(s))
	for i, v := range s {
		if v == nil {
			vals[i] = pgtype.Text{Valid: false}
		} else {
			vals[i] = pgtype.Text{String: *v, Valid: true}
		}
	}

	return &TextArrayValue{
		Array: newValidArr(vals),
	}
}

type TextArrayValue struct {
	pgtype.Array[pgtype.Text]
}

func (a *TextArrayValue) Null() bool {
	return !a.Valid
}

func (a *TextArrayValue) Compare(v Value, op ComparisonOp) (*BoolValue, error) {
	return cmpArrs(a, v, op)
}

func (a *TextArrayValue) Len() int {
	return len(a.Elements)
}

func (a *TextArrayValue) Index(i int64) (ScalarValue, error) {
	if i < 1 || i > int64(a.Len()) {
		return nil, fmt.Errorf("index out of bounds")
	}

	return &TextValue{a.Elements[i-1]}, nil
}

func (a *TextArrayValue) Set(i int64, v ScalarValue) error {
	err := allocArr(&a.Array, i)
	if err != nil {
		return err
	}

	val, ok := v.(*TextValue)
	if !ok {
		return fmt.Errorf("cannot set non-text value in text array")
	}

	a.Elements[i-1] = val.Text
	return nil
}

func (a *TextArrayValue) Type() *types.DataType {
	return types.TextArrayType
}

func (a *TextArrayValue) RawValue() any {
	if !a.Valid {
		return nil
	}

	res := make([]*string, len(a.Elements))
	for i, v := range a.Elements {
		if v.Valid {
			res[i] = &v.String
		}
	}

	return res
}

func (a *TextArrayValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.IntArrayType:
		return castArr(a, func(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) }, newIntArr)
	case types.BoolArrayType:
		return castArr(a, strconv.ParseBool, newBoolArrayValue)
	case types.DecimalArrayType:
		return castArrWithPtr(a, decimal.NewFromString, newDecimalArrayValue)
	case types.UUIDArrayType:
		return castArrWithPtr(a, types.ParseUUID, newUUIDArrayValue)
	default:
		return nil, fmt.Errorf("cannot cast text array to %s", t)
	}
}

// castArr casts an array of one type to an array of another type.
// Generics:
// A is the current scalar Kwil type
// B is the desired scalar Kwil type
// C is the current array Kwil type
// D is the desired array Kwil type
// Params:
// c: the current array
// get: a function that converts the current array's scalar type to the desired scalar type
// newArr: a function that creates a new array of the desired type
func castArr[A any, B any, C ArrayValue, D ArrayValue](c C, get func(a A) (B, error), newArr func([]*B) D) (D, error) {
	return castArrWithPtr(c, func(b A) (*B, error) {
		res, err := get(b)
		if err != nil {
			return nil, err
		}

		return &res, nil
	}, newArr)
}

// castArrWithPtr casts an array of one type to an array of another type.
// It expects that the get function will return a pointer to the desired type.
func castArrWithPtr[A any, B any, C ArrayValue, D ArrayValue](c C, get func(a A) (*B, error), newArr func([]*B) D) (D, error) {
	res := make([]*B, c.Len())
	for i := range c.Len() {
		v, err := c.Index(int64(i + 1)) // Index is 1-based
		if err != nil {
			return *new(D), err
		}

		// if the value is nil, we dont need to do anything
		if !v.Null() {
			raw, ok := v.RawValue().(A)
			if !ok {
				// should never happen unless I messed up the types
				return *new(D), fmt.Errorf("internal bug: unexpected type %T", v.RawValue())
			}

			res[i], err = get(raw)
			if err != nil {
				return *new(D), err
			}
		}
	}

	return newArr(res), nil
}

func newBoolArrayValue(b []*bool) *BoolArrayValue {
	vals := make([]pgtype.Bool, len(b))
	for i, v := range b {
		if v == nil {
			vals[i] = pgtype.Bool{Valid: false}
		} else {
			vals[i] = pgtype.Bool{Bool: *v, Valid: true}
		}
	}

	return &BoolArrayValue{
		Array: newValidArr(vals),
	}
}

type BoolArrayValue struct {
	pgtype.Array[pgtype.Bool]
}

func (a *BoolArrayValue) Null() bool {
	return !a.Valid
}

func (a *BoolArrayValue) Compare(v Value, op ComparisonOp) (*BoolValue, error) {
	return cmpArrs(a, v, op)
}

func (a *BoolArrayValue) Len() int {
	return len(a.Elements)
}

func (a *BoolArrayValue) Index(i int64) (ScalarValue, error) {
	if i < 1 || i > int64(a.Len()) {
		return nil, fmt.Errorf("index out of bounds")
	}

	return &BoolValue{a.Elements[i-1]}, nil
}

func (a *BoolArrayValue) Set(i int64, v ScalarValue) error {
	err := allocArr(&a.Array, i)
	if err != nil {
		return err
	}

	val, ok := v.(*BoolValue)
	if !ok {
		return fmt.Errorf("cannot set non-bool value in bool array")
	}

	a.Elements[i-1] = val.Bool
	return nil
}

func (a *BoolArrayValue) Type() *types.DataType {
	return types.BoolArrayType
}

func (a *BoolArrayValue) RawValue() any {
	if !a.Valid {
		return nil
	}

	barr := make([]*bool, len(a.Elements))
	for i, v := range a.Elements {
		if v.Valid {
			barr[i] = &v.Bool
		}
	}

	return barr
}

func (a *BoolArrayValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.TextArrayType:
		return castArr(a, func(b bool) (string, error) { return strconv.FormatBool(b), nil }, newTextArrayValue)
	case types.IntArrayType:
		return castArr(a, func(b bool) (int64, error) {
			if b {
				return 1, nil
			} else {
				return 0, nil
			}
		}, newIntArr)
	default:
		return nil, fmt.Errorf("cannot cast bool array to %s", t)
	}
}

func newDecimalArrayValue(d []*decimal.Decimal) *DecimalArrayValue {
	vals := make([]pgtype.Numeric, len(d))
	for i, v := range d {
		if v == nil {
			vals[i] = pgtype.Numeric{Valid: false}
		} else {
			vals[i] = pgTypeFromDec(v)
		}
	}

	return &DecimalArrayValue{
		Array: newValidArr(vals),
	}
}

type DecimalArrayValue struct {
	pgtype.Array[pgtype.Numeric]
}

func (a *DecimalArrayValue) Null() bool {
	return !a.Valid
}

// detectDecArrType detects the type of a decimal array.
// It returns the type, and a boolean indicating if the array does
// not have any non-null values.
func detectDecArrType(arr *DecimalArrayValue) (typ *types.DataType, containsOnlyNulls bool) {
	var firstFound *types.DataType
	for _, v := range arr.Elements {
		if v.Valid {

			dec, err := decFromPgType(v)
			if err != nil {
				panic(err)
			}

			if firstFound == nil {
				d, err := types.NewDecimalType(dec.Precision(), dec.Scale())
				if err != nil {
					panic(err)
				}

				firstFound = d
			} else {
				d2, err := types.NewDecimalType(dec.Precision(), dec.Scale())
				if err != nil {
					panic(err)
				}

				if !firstFound.EqualsStrict(d2) {
					// should never reach here
					panic("mixed types in decimal array")
				}
			}
		}
	}

	if firstFound == nil {
		return types.DecimalType, true
	}

	return firstFound, false
}

func (a *DecimalArrayValue) Compare(v Value, op ComparisonOp) (*BoolValue, error) {
	return cmpArrs(a, v, op)
}

// cmpArrs compares two Kwil array types.
func cmpArrs[M ArrayValue](a M, b Value, op ComparisonOp) (*BoolValue, error) {
	if res, early := nullCmp(a, b, op); early {
		return res, nil
	}

	val2, ok := b.(M)
	if !ok {
		return nil, makeTypeErr(a, b)
	}

	isEqual := func(a, b ArrayValue) (isEq bool, err error) {
		if a.Len() != b.Len() {
			return false, nil
		}

		for i := 1; i <= a.Len(); i++ {
			v1, err := a.Index(int64(i))
			if err != nil {
				return false, err
			}

			v2, err := b.Index(int64(i))
			if err != nil {
				return false, err
			}

			if v1.Null() && v2.Null() {
				continue
			}

			if v1.Null() || v2.Null() {
				return false, nil
			}

			res, err := v1.Compare(v2, equal)
			if err != nil {
				return false, err
			}

			if !res.Bool.Bool {
				return false, nil
			}
		}

		return true, nil
	}

	eq, err := isEqual(a, val2)
	if err != nil {
		return nil, err
	}

	switch op {
	case equal:
		return newBool(eq), nil
	case isDistinctFrom:
		return newBool(!eq), nil
	default:
		return nil, fmt.Errorf("only =, IS DISTINCT FROM are supported for array comparison")
	}
}

func (a *DecimalArrayValue) Len() int {
	return len(a.Elements)
}

func (a *DecimalArrayValue) Index(i int64) (ScalarValue, error) {
	if i < 1 || i > int64(len(a.Elements)) {
		return nil, fmt.Errorf("index out of bounds")
	}

	return &DecimalValue{Numeric: a.Elements[i-1]}, nil
}

func (a *DecimalArrayValue) Set(i int64, v ScalarValue) error {
	err := allocArr(&a.Array, i)
	if err != nil {
		return err
	}

	val, ok := v.(*DecimalValue)
	if !ok {
		return fmt.Errorf("cannot set non-decimal value in decimal array")
	}

	a.Elements[i-1] = val.Numeric
	return nil
}

func (a *DecimalArrayValue) Type() *types.DataType {
	typ, _ := detectDecArrType(a)
	return typ
}

func (a *DecimalArrayValue) RawValue() any {
	if !a.Valid {
		return nil
	}

	res := make([]*decimal.Decimal, len(a.Elements))
	for i, v := range a.Elements {
		if v.Valid {
			dec, err := decFromPgType(v)
			if err != nil {
				panic(err)
			}

			res[i] = dec
		}
	}

	return res
}

func (a *DecimalArrayValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.TextArrayType:
		return castArr(a, func(d *decimal.Decimal) (string, error) { return d.String(), nil }, newTextArrayValue)
	case types.IntArrayType:
		return castArr(a, func(d *decimal.Decimal) (int64, error) { return d.Int64() }, newIntArr)
	default:
		return nil, fmt.Errorf("cannot cast decimal array to %s", t)
	}
}

func newBlobArrayValue(b [][]byte) *BlobArrayValue {
	vals := make([]pgtype.PreallocBytes, len(b))
	for i, v := range b {
		if v == nil {
			vals[i] = nil
		} else {
			vals[i] = pgtype.PreallocBytes(v)
		}
	}

	return &BlobArrayValue{
		Array: newValidArr(vals),
	}
}

type BlobArrayValue struct {
	pgtype.Array[pgtype.PreallocBytes]
}

func (a *BlobArrayValue) Null() bool {
	return !a.Valid
}

func (a *BlobArrayValue) Compare(v Value, op ComparisonOp) (*BoolValue, error) {
	return cmpArrs(a, v, op)
}

func (a *BlobArrayValue) Len() int {
	return len(a.Elements)
}

func (a *BlobArrayValue) Index(i int64) (ScalarValue, error) {
	if i < 1 || i > int64(a.Len()) {
		return nil, fmt.Errorf("index out of bounds")
	}

	return &BlobValue{a.Elements[i-1]}, nil
}

func (a *BlobArrayValue) Set(i int64, v ScalarValue) error {
	err := allocArr(&a.Array, i)
	if err != nil {
		return err
	}

	val, ok := v.(*BlobValue)
	if !ok {
		return fmt.Errorf("cannot set non-blob value in blob array")
	}

	// copy the blob value to avoid mutation
	valCopy := make([]byte, len(val.PreallocBytes))
	copy(valCopy, val.PreallocBytes)

	a.Elements[i-1] = valCopy
	return nil
}

func (a *BlobArrayValue) Type() *types.DataType {
	return types.BlobArrayType
}

func (a *BlobArrayValue) RawValue() any {
	if !a.Valid {
		return nil
	}

	res := make([][]byte, len(a.Elements))
	for i, v := range a.Elements {
		if v != nil {
			res[i] = make([]byte, len(v))
			copy(res[i], v)
		}
	}

	return res
}

func (a *BlobArrayValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.TextArrayType:
		return castArr(a, func(b []byte) (string, error) { return string(b), nil }, newTextArrayValue)
	default:
		return nil, fmt.Errorf("cannot cast blob array to %s", t)
	}
}

func newUUIDArrayValue(u []*types.UUID) *UuidArrayValue {
	vals := make([]pgtype.UUID, len(u))
	for i, v := range u {
		if v == nil {
			vals[i] = pgtype.UUID{Valid: false}
		} else {
			vals[i] = pgtype.UUID{Bytes: *v, Valid: true}
		}
	}

	return &UuidArrayValue{
		Array: newValidArr(vals),
	}
}

type UuidArrayValue struct {
	pgtype.Array[pgtype.UUID]
}

func (a *UuidArrayValue) Null() bool {
	return !a.Valid
}

func (a *UuidArrayValue) Compare(v Value, op ComparisonOp) (*BoolValue, error) {
	return cmpArrs(a, v, op)
}

func (a *UuidArrayValue) Len() int {
	return len(a.Elements)
}

func (a *UuidArrayValue) Index(i int64) (ScalarValue, error) {
	if i < 1 || i > int64(a.Len()) {
		return nil, fmt.Errorf("index out of bounds")
	}

	return &UUIDValue{a.Elements[i-1]}, nil
}

func (a *UuidArrayValue) Set(i int64, v ScalarValue) error {
	err := allocArr(&a.Array, i)
	if err != nil {
		return err
	}

	val, ok := v.(*UUIDValue)
	if !ok {
		return fmt.Errorf("cannot set non-uuid value in uuid array")
	}

	a.Elements[i-1] = val.UUID
	return nil
}

func (a *UuidArrayValue) Type() *types.DataType {
	return types.UUIDArrayType
}

func (a *UuidArrayValue) RawValue() any {
	if !a.Valid {
		return nil
	}

	res := make([]*types.UUID, len(a.Elements))
	for i, v := range a.Elements {
		if v.Valid {
			u := types.UUID(v.Bytes)
			res[i] = &u
		}
	}

	return res
}

func (a *UuidArrayValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.TextArrayType:
		return castArr(a, func(u types.UUID) (string, error) { return u.String(), nil }, newTextArrayValue)
	default:
		return nil, fmt.Errorf("cannot cast uuid array to %s", t)
	}
}

func newRecordValue() *RecordValue {
	return &RecordValue{
		Fields: make(map[string]Value),
	}
}

// RecordValue is a special type that represents a row in a table.
type RecordValue struct {
	Fields map[string]Value
	Order  []string
}

func (r *RecordValue) Null() bool {
	return len(r.Fields) == 0
}

func (r *RecordValue) AddValue(k string, v Value) error {
	_, ok := r.Fields[k]
	if ok {
		// protecting against this since it would detect non-determinism,
		// but our query planner should already protect against this
		return fmt.Errorf("record already has field %s", k)
	}

	r.Fields[k] = v
	r.Order = append(r.Order, k)
	return nil
}

func (o *RecordValue) Compare(v Value, op ComparisonOp) (*BoolValue, error) {
	if res, early := nullCmp(o, v, op); early {
		return res, nil
	}

	val2, ok := v.(*RecordValue)
	if !ok {
		return nil, makeTypeErr(o, v)
	}

	isSame := true
	if len(o.Fields) != len(val2.Fields) {
		isSame = false
	}

	if isSame {
		for i, field := range o.Order {
			v2, ok := val2.Fields[field]
			if !ok {
				isSame = false
				break
			}

			eq, err := o.Fields[field].Compare(v2, equal)
			if err != nil {
				return nil, err
			}

			if !eq.RawValue().(bool) {
				isSame = false
				break
			}

			// check the order
			if field != val2.Order[i] {
				isSame = false
				break
			}
		}
	}

	switch op {
	case equal:
		return newBool(isSame), nil
	default:
		return nil, fmt.Errorf("unknown comparison operator: %d", op)
	}
}

func (o *RecordValue) Type() *types.DataType {
	return types.RecordType
}

func (o *RecordValue) RawValue() any {
	return o.Fields
}

func (o *RecordValue) Cast(t *types.DataType) (Value, error) {
	return nil, fmt.Errorf("cannot cast record to %s", t)
}

func cmpIntegers(a, b int, op ComparisonOp) (*BoolValue, error) {
	switch op {
	case equal:
		return newBool(a == b), nil
	case lessThan:
		return newBool(a < b), nil
	case greaterThan:
		return newBool(a > b), nil
	case isDistinctFrom:
		return newBool(a != b), nil
	default:
		return nil, fmt.Errorf("unknown comparison operator: %d", op)
	}
}
