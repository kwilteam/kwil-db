package precompiles

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/node/engine"
)

// valueMapping maps Go types and Kwil native types.
type valueMapping struct {
	// KwilType is the Kwil type that the value maps to.
	// It will ignore the metadata of the type.
	KwilType *types.DataType
	// ZeroValue creates a zero-value of the type.
	ZeroValue func(t *types.DataType) (Value, error)
	// NullValue creates a null-value of the type.
	NullValue func(t *types.DataType) (Value, error)
}

var (
	kwilTypeToValue = map[struct {
		name    string
		isArray bool
	}]valueMapping{}
)

func registerValueMapping(ms ...valueMapping) {
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
		valueMapping{
			KwilType: types.IntType,
			ZeroValue: func(t *types.DataType) (Value, error) {
				return MakeInt8(0), nil
			},
			NullValue: func(t *types.DataType) (Value, error) {
				return &Int8Value{
					Int8: pgtype.Int8{
						Valid: false,
					},
				}, nil
			},
		},
		valueMapping{
			// TODO: we can get rid of this
			KwilType: types.NullType,
			ZeroValue: func(t *types.DataType) (Value, error) {
				return nil, fmt.Errorf("cannot create zero value of null type")
			},
			NullValue: func(t *types.DataType) (Value, error) {
				return &TextValue{
					Text: pgtype.Text{
						Valid: false,
					},
				}, nil
			},
		},
		valueMapping{
			KwilType: types.TextType,
			ZeroValue: func(t *types.DataType) (Value, error) {
				return MakeText(""), nil
			},
			NullValue: func(t *types.DataType) (Value, error) {
				return &TextValue{
					Text: pgtype.Text{
						Valid: false,
					},
				}, nil
			},
		},
		valueMapping{
			KwilType: types.BoolType,
			ZeroValue: func(t *types.DataType) (Value, error) {
				return MakeBool(false), nil
			},
			NullValue: func(t *types.DataType) (Value, error) {
				return &BoolValue{
					Bool: pgtype.Bool{
						Valid: false,
					},
				}, nil
			},
		},
		valueMapping{
			KwilType: types.BlobType,
			ZeroValue: func(t *types.DataType) (Value, error) {
				return MakeBlob([]byte{}), nil
			},
			NullValue: func(t *types.DataType) (Value, error) {
				return &BlobValue{}, nil
			},
		},
		valueMapping{
			KwilType: types.UUIDType,
			ZeroValue: func(t *types.DataType) (Value, error) {
				return MakeUUID(&types.UUID{}), nil
			},
			NullValue: func(t *types.DataType) (Value, error) {
				return &UUIDValue{
					UUID: pgtype.UUID{
						Valid: false,
					},
				}, nil
			},
		},
		valueMapping{
			KwilType: types.DecimalType,
			ZeroValue: func(t *types.DataType) (Value, error) {
				if !t.HasMetadata() {
					return nil, fmt.Errorf("cannot create zero value of decimal type with zero precision and scale")
				}

				dec, err := decimal.NewFromString("0")
				if err != nil {
					return nil, err
				}
				dec2 := MakeDecimal(dec)

				prec := t.Metadata[0]
				scale := t.Metadata[1]
				dec2.metadata = &precAndScale{prec, scale}

				return dec2, nil
			},
			NullValue: func(t *types.DataType) (Value, error) {
				if !t.HasMetadata() {
					return nil, fmt.Errorf("cannot create null value of decimal type with zero precision and scale")
				}
				prec := t.Metadata[0]
				scale := t.Metadata[1]
				d := MakeDecimal(nil)
				d.metadata = &precAndScale{prec, scale}
				return d, nil
			},
		},
		valueMapping{
			KwilType: types.IntArrayType,
			ZeroValue: func(t *types.DataType) (Value, error) {
				return &IntArrayValue{
					OneDArray: newValidArr([]pgtype.Int8{}),
				}, nil
			},
			NullValue: func(t *types.DataType) (Value, error) {
				return &IntArrayValue{
					OneDArray: newNullArray[pgtype.Int8](),
				}, nil
			},
		},
		valueMapping{
			KwilType: types.TextArrayType,
			ZeroValue: func(t *types.DataType) (Value, error) {
				return &TextArrayValue{
					OneDArray: newValidArr([]pgtype.Text{}),
				}, nil
			},
			NullValue: func(t *types.DataType) (Value, error) {
				return &TextArrayValue{
					OneDArray: newNullArray[pgtype.Text](),
				}, nil
			},
		},
		valueMapping{
			KwilType: types.BoolArrayType,
			ZeroValue: func(t *types.DataType) (Value, error) {
				return &BoolArrayValue{
					OneDArray: newValidArr([]pgtype.Bool{}),
				}, nil
			},
			NullValue: func(t *types.DataType) (Value, error) {
				return &BoolArrayValue{
					OneDArray: newNullArray[pgtype.Bool](),
				}, nil
			},
		},
		valueMapping{
			KwilType: types.BlobArrayType,
			ZeroValue: func(t *types.DataType) (Value, error) {
				return &BlobArrayValue{
					OneDArray: newValidArr([]*BlobValue{}),
				}, nil
			},
			NullValue: func(t *types.DataType) (Value, error) {
				return &BlobArrayValue{
					OneDArray: newNullArray[*BlobValue](),
				}, nil
			},
		},
		valueMapping{
			KwilType: types.DecimalArrayType,
			ZeroValue: func(t *types.DataType) (Value, error) {
				if !t.HasMetadata() {
					return nil, fmt.Errorf("cannot create zero value of decimal type with zero precision and scale")
				}

				prec := t.Metadata[0]
				scale := t.Metadata[1]

				arr := &DecimalArrayValue{
					OneDArray: newValidArr([]pgtype.Numeric{}),
					metadata:  &precAndScale{prec, scale},
				}
				return arr, nil
			},
			NullValue: func(t *types.DataType) (Value, error) {
				if !t.HasMetadata() {
					return nil, fmt.Errorf("cannot create null value of decimal array type with zero precision and scale")
				}

				prec := t.Metadata[0]
				scale := t.Metadata[1]

				arr := newNullDecArr(types.DecimalArrayType)
				arr.metadata = &precAndScale{prec, scale}
				return arr, nil
			},
		},
		valueMapping{
			KwilType: types.UUIDArrayType,
			ZeroValue: func(t *types.DataType) (Value, error) {
				return &UuidArrayValue{
					OneDArray: newValidArr([]pgtype.UUID{}),
				}, nil
			},
			NullValue: func(t *types.DataType) (Value, error) {
				return &UuidArrayValue{
					OneDArray: newNullArray[pgtype.UUID](),
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
		return nil, fmt.Errorf("type %s not found", t.String())
	}

	return m.ZeroValue(t)
}

// Value is a value that can be compared, used in arithmetic operations,
// and have unary operations applied to it.
type Value interface {
	common.EngineValue
	// Compare compares the variable with another variable using the given comparison operator.
	// It will return a boolean value or null, depending on the comparison and the values.
	Compare(v Value, op engine.ComparisonOp) (*BoolValue, error)
	// Cast casts the variable to the given type.
	// It is meant to mirror Postgres's type casting behavior.
	Cast(t *types.DataType) (Value, error)
}

// ScalarValue is a scalar value that can be computed on and have unary operations applied to it.
type ScalarValue interface {
	Value
	// Arithmetic performs an arithmetic operation on the variable with another variable.
	Arithmetic(v ScalarValue, op engine.ArithmeticOp) (ScalarValue, error)
	// Unary applies a unary operation to the variable.
	Unary(op engine.UnaryOp) (ScalarValue, error)
	// Array creates an array from this scalar value and any other scalar values.
	Array(v ...ScalarValue) (ArrayValue, error)
}

// ArrayValue is an array value that can be compared and have unary operations applied to it.
type ArrayValue interface {
	Value
	// Len returns the length of the array.
	Len() int32
	// Get returns the value at the given index.
	// If the index is out of bounds, an error is returned.
	// All indexing is 1-based.
	Get(i int32) (ScalarValue, error)
	// Set sets the value at the given index.
	// If the index is out of bounds, enough space is allocated to set the value.
	// This matches the behavior of Postgres.
	// All indexing is 1-based.
	Set(i int32, v ScalarValue) error
}

func newValidArr[T any](a []T) OneDArray[T] {
	return OneDArray[T]{
		Array: pgtype.Array[T]{
			Elements: a,
			Dims:     []pgtype.ArrayDimension{{Length: int32(len(a)), LowerBound: 1}},
			Valid:    true,
		},
	}
}

// NewValue creates a new Value from the given any val.
func NewValue(v any) (Value, error) {
	switch v := v.(type) {
	case Value:
		return v, nil
	case int64:
		return MakeInt8(v), nil
	case int:
		return MakeInt8(int64(v)), nil
	case string:
		return MakeText(v), nil
	case bool:
		return MakeBool(v), nil
	case []byte:
		return MakeBlob(v), nil
	case *types.UUID:
		return MakeUUID(v), nil
	case types.UUID:
		return MakeUUID(&v), nil
	case *decimal.Decimal:
		return MakeDecimal(v), nil
	case decimal.Decimal:
		return MakeDecimal(&v), nil
	case []int64:
		pgInts := make([]pgtype.Int8, len(v))
		for i, val := range v {
			pgInts[i].Int64 = val
			pgInts[i].Valid = true
		}

		return &IntArrayValue{
			OneDArray: newValidArr(pgInts),
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
			OneDArray: newValidArr(pgInts),
		}, nil
	case []int:
		pgInts := make([]pgtype.Int8, len(v))
		for i, val := range v {
			pgInts[i].Int64 = int64(val)
			pgInts[i].Valid = true
		}

		return &IntArrayValue{
			OneDArray: newValidArr(pgInts),
		}, nil
	case []*int:
		pgInts := make([]pgtype.Int8, len(v))
		for i, val := range v {
			if val == nil {
				pgInts[i].Valid = false
			} else {
				pgInts[i].Int64 = int64(*val)
				pgInts[i].Valid = true
			}
		}
		return &IntArrayValue{
			OneDArray: newValidArr(pgInts),
		}, nil
	case []string:
		pgTexts := make([]pgtype.Text, len(v))
		for i, val := range v {
			pgTexts[i].String = val
			pgTexts[i].Valid = true
		}

		return &TextArrayValue{
			OneDArray: newValidArr(pgTexts),
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
			OneDArray: newValidArr(pgTexts),
		}, nil
	case []bool:
		pgBools := make([]pgtype.Bool, len(v))
		for i, val := range v {
			pgBools[i].Bool = val
			pgBools[i].Valid = true
		}

		return &BoolArrayValue{
			OneDArray: newValidArr(pgBools),
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
			OneDArray: newValidArr(pgBools),
		}, nil
	case [][]byte:
		pgBlobs := make([]*BlobValue, len(v))
		for i, val := range v {
			pgBlobs[i] = MakeBlob(val)
		}

		return &BlobArrayValue{
			OneDArray: newValidArr(pgBlobs),
		}, nil
	case []*[]byte:
		pgBlobs := make([]*BlobValue, len(v))
		for i, val := range v {
			if val == nil {
				pgBlobs[i] = &BlobValue{}
			} else {
				pgBlobs[i] = MakeBlob(*val)
			}
		}

		return &BlobArrayValue{
			OneDArray: newValidArr(pgBlobs),
		}, nil
	case []*decimal.Decimal:
		pgDecs := make([]pgtype.Numeric, len(v))
		for i, val := range v {
			pgDecs[i] = pgTypeFromDec(val)
		}

		var metadata *precAndScale
		if len(v) > 0 {
			precCopy := v[0].Precision()
			scaleCopy := v[0].Scale()
			metadata = &precAndScale{precCopy, scaleCopy}
		}

		return &DecimalArrayValue{
			OneDArray: newValidArr(pgDecs),
			metadata:  metadata,
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
			OneDArray: newValidArr(pgUUIDs),
		}, nil
	case nil:
		return &TextValue{
			Text: pgtype.Text{
				Valid: false,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected type %T", v)
	}
}

func makeTypeErr(left, right Value) error {
	return fmt.Errorf("%w: left: %s right: %s", engine.ErrType, left.Type(), right.Type())
}

// makeArrTypeErr returns an error for when an array operation is performed on a non-array type.
func makeArrTypeErr(arrVal Value, newVal Value) error {
	return fmt.Errorf("%w: cannot create an array of different types %s and %s", engine.ErrType, arrVal.Type(), newVal.Type())
}

func MakeInt8(i int64) *Int8Value {
	return &Int8Value{
		Int8: pgtype.Int8{
			Int64: i,
			Valid: true,
		},
	}
}

type Int8Value struct {
	pgtype.Int8
}

func (i *Int8Value) Null() bool {
	return !i.Valid
}

func (v *Int8Value) Compare(v2 Value, op engine.ComparisonOp) (*BoolValue, error) {
	if res, early := nullCmp(v, v2, op); early {
		return res, nil
	}

	val2, ok := v2.(*Int8Value)
	if !ok {
		return nil, makeTypeErr(v, v2)
	}

	var b bool
	switch op {
	case engine.EQUAL:
		b = v.Int64 == val2.Int64
	case engine.LESS_THAN:
		b = v.Int64 < val2.Int64
	case engine.GREATER_THAN:
		b = v.Int64 > val2.Int64
	case engine.IS_DISTINCT_FROM:
		b = v.Int64 != val2.Int64
	default:
		return nil, fmt.Errorf("%w: cannot compare int with operator %s", engine.ErrComparison, op)
	}

	return MakeBool(b), nil
}

// nullCmp is a helper function for comparing null values.
// It takes two values and a comparison operator.
// If the operator is IS or IS DISTINCT FROM, it will return a boolean value
// based on the comparison of the two values.
// If the operator is any other operator and either of the values is null,
// it will return a null value.
func nullCmp(a, b Value, op engine.ComparisonOp) (*BoolValue, bool) {
	// if it is is_DISTINCT_FROM or is, we should handle nulls
	// Otherwise, if either is a null, we return early because we cannot compare
	// a null value with a non-null value.
	if op == engine.IS_DISTINCT_FROM {
		if a.Null() && b.Null() {
			return MakeBool(false), true
		}
		if a.Null() || b.Null() {
			return MakeBool(true), true
		}

		// otherwise, we let equality handle it
	}

	if op == engine.IS {
		if a.Null() && b.Null() {
			return MakeBool(true), true
		}
		if a.Null() || b.Null() {
			return MakeBool(false), true
		}
	}

	if a.Null() || b.Null() {
		nv, err := MakeNull(types.BoolType)
		if err != nil {
			panic(err) // should never happen, MakeNull(types.BoolType) should never return an error
		}
		boolType, ok := nv.(*BoolValue)
		if !ok {
			panic("MakeNull(types.BoolType) did not return a *BoolValue") // should never happen
		}

		return boolType, true
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

func (i *Int8Value) Arithmetic(v ScalarValue, op engine.ArithmeticOp) (ScalarValue, error) {
	if res, early := checkScalarNulls(i, v); early {
		return res, nil
	}

	val2, ok := v.(*Int8Value)
	if !ok {
		return nil, makeTypeErr(i, v)
	}

	var r int64

	switch op {
	case engine.ADD:
		r = i.Int64 + val2.Int64
	case engine.SUB:
		r = i.Int64 - val2.Int64
	case engine.MUL:
		r = i.Int64 * val2.Int64
	case engine.DIV:
		if val2.Int64 == 0 {
			return nil, fmt.Errorf("%w: cannot divide by zero", engine.ErrArithmetic)
		}
		r = i.Int64 / val2.Int64
	case engine.MOD:
		if val2.Int64 == 0 {
			return nil, fmt.Errorf("%w: cannot modulo by zero", engine.ErrArithmetic)
		}
		r = i.Int64 % val2.Int64
	case engine.EXP:
		p := math.Pow(float64(i.Int64), float64(val2.Int64))
		if p > math.MaxInt64 {
			return nil, fmt.Errorf("%w: result of exponentiation is too large", engine.ErrArithmetic)
		}
		r = int64(p)
	default:
		return nil, fmt.Errorf("%w: cannot perform arithmetic operation %s on type int", engine.ErrArithmetic, op)
	}

	return &Int8Value{
		Int8: pgtype.Int8{
			Int64: r,
			Valid: true,
		},
	}, nil
}

func (i *Int8Value) Unary(op engine.UnaryOp) (ScalarValue, error) {
	if i.Null() {
		return i, nil
	}

	switch op {
	case engine.NEG:
		return &Int8Value{Int8: pgtype.Int8{Int64: -i.Int64, Valid: true}}, nil
	case engine.NOT:
		return nil, fmt.Errorf("%w: cannot apply logical NOT to an integer", engine.ErrUnary)
	case engine.POS:
		return i, nil
	default:
		return nil, fmt.Errorf("%w: unknown unary operator: %s", engine.ErrUnary, op)
	}
}

func (i *Int8Value) Type() *types.DataType {
	return types.IntType
}

func (i *Int8Value) RawValue() any {
	if !i.Valid {
		return nil
	}

	return i.Int64
}

func (i *Int8Value) Array(v ...ScalarValue) (ArrayValue, error) {
	pgtArr := make([]pgtype.Int8, len(v)+1)
	pgtArr[0] = i.Int8
	for j, val := range v {
		if intVal, ok := val.(*Int8Value); !ok {
			return nil, makeArrTypeErr(i, val)
		} else {
			pgtArr[j+1] = intVal.Int8
		}
	}

	return &IntArrayValue{
		OneDArray: newValidArr(pgtArr),
	}, nil
}

func (i *Int8Value) Cast(t *types.DataType) (Value, error) {
	if i.Null() {
		return MakeNull(t)
	}

	// we check for decimal first since type switching on it
	// doesn't work, since it has precision and scale
	if t.Name == types.NumericStr {
		if t.IsArray {
			return nil, castErr(errors.New("cannot cast int to decimal array"))
		}

		dec, err := decimal.NewFromString(fmt.Sprint(i.Int64))
		if err != nil {
			return nil, castErr(err)
		}

		return MakeDecimal(dec), nil
	}

	switch *t {
	case *types.IntType:
		return i, nil
	case *types.TextType:
		return MakeText(fmt.Sprint(i.Int64)), nil
	case *types.BoolType:
		return MakeBool(i.Int64 != 0), nil
	default:
		return nil, castErr(fmt.Errorf("cannot cast int to %s", t))
	}
}

// MakeNull creates a new null value of the given type.
func MakeNull(t *types.DataType) (Value, error) {
	m, ok := kwilTypeToValue[struct {
		name    string
		isArray bool
	}{
		name:    t.Name,
		isArray: t.IsArray,
	}]
	if !ok {
		return nil, fmt.Errorf("type %s not found", t.String())
	}

	return m.NullValue(t)
}

func newNullArray[T any]() OneDArray[T] {
	return OneDArray[T]{
		Array: pgtype.Array[T]{
			Valid: false,
		},
	}
}

func MakeText(s string) *TextValue {
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

func (s *TextValue) Compare(v Value, op engine.ComparisonOp) (*BoolValue, error) {
	if res, early := nullCmp(s, v, op); early {
		return res, nil
	}

	val2, ok := v.(*TextValue)
	if !ok {
		return nil, makeTypeErr(s, v)
	}

	var b bool
	switch op {
	case engine.EQUAL:
		b = s.String == val2.String
	case engine.LESS_THAN:
		b = s.String < val2.String
	case engine.GREATER_THAN:
		b = s.String > val2.String
	case engine.IS_DISTINCT_FROM:
		b = s.String != val2.String
	default:
		return nil, fmt.Errorf("%w: cannot use comparison operator %s with type %s", engine.ErrComparison, s.Type(), op)
	}

	return MakeBool(b), nil
}

func (s *TextValue) Arithmetic(v ScalarValue, op engine.ArithmeticOp) (ScalarValue, error) {
	if res, early := checkScalarNulls(s, v); early {
		return res, nil
	}

	val2, ok := v.(*TextValue)
	if !ok {
		return nil, makeTypeErr(s, v)
	}

	if op == engine.CONCAT {
		return MakeText(s.String + val2.String), nil
	}

	return nil, fmt.Errorf("%w: cannot perform arithmetic operation %s on type string", engine.ErrArithmetic, op)
}

func (s *TextValue) Unary(op engine.UnaryOp) (ScalarValue, error) {
	return nil, fmt.Errorf("%w: cannot perform unary operation on string", engine.ErrUnary)
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
			return nil, makeArrTypeErr(s, val)
		} else {
			pgtArr[j+1] = textVal.Text
		}
	}

	arr := newValidArr(pgtArr)

	return &TextArrayValue{
		OneDArray: arr,
	}, nil
}

func (s *TextValue) Cast(t *types.DataType) (Value, error) {
	if s.Null() {
		return MakeNull(t)
	}

	if t.Name == types.NumericStr {
		if t.IsArray {
			return nil, castErr(errors.New("cannot cast text to decimal array"))
		}

		dec, err := decimal.NewFromString(s.String)
		if err != nil {
			return nil, castErr(err)
		}

		return MakeDecimal(dec), nil
	}

	switch *t {
	case *types.IntType:
		i, err := strconv.ParseInt(s.String, 10, 64)
		if err != nil {
			return nil, castErr(err)
		}

		return MakeInt8(i), nil
	case *types.TextType:
		return s, nil
	case *types.BoolType:
		b, err := strconv.ParseBool(s.String)
		if err != nil {
			return nil, castErr(err)
		}

		return MakeBool(b), nil
	case *types.UUIDType:
		u, err := types.ParseUUID(s.String)
		if err != nil {
			return nil, castErr(err)
		}

		return MakeUUID(u), nil
	case *types.BlobType:
		return MakeBlob([]byte(s.String)), nil
	default:
		return nil, castErr(fmt.Errorf("cannot cast text to %s", t))
	}
}

func MakeBool(b bool) *BoolValue {
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

func (b *BoolValue) Compare(v Value, op engine.ComparisonOp) (*BoolValue, error) {
	if res, early := nullCmp(b, v, op); early {
		return res, nil
	}

	val2, ok := v.(*BoolValue)
	if !ok {
		return nil, makeTypeErr(b, v)
	}

	var b2 bool
	switch op {
	case engine.EQUAL:
		b2 = b.Bool.Bool == val2.Bool.Bool
	case engine.IS_DISTINCT_FROM:
		b2 = b.Bool.Bool != val2.Bool.Bool
	case engine.LESS_THAN:
		b2 = !b.Bool.Bool && val2.Bool.Bool
	case engine.GREATER_THAN:
		b2 = b.Bool.Bool && !val2.Bool.Bool
	case engine.IS:
		b2 = b.Bool.Bool == val2.Bool.Bool
	default:
		return nil, fmt.Errorf("%w: cannot use comparison operator %s with type %s", engine.ErrComparison, b.Type(), op)
	}

	return MakeBool(b2), nil
}

func (b *BoolValue) Arithmetic(v ScalarValue, op engine.ArithmeticOp) (ScalarValue, error) {
	return nil, fmt.Errorf("%w: cannot perform arithmetic operation on bool", engine.ErrArithmetic)
}

func (b *BoolValue) Unary(op engine.UnaryOp) (ScalarValue, error) {
	if b.Null() {
		return b, nil
	}

	switch op {
	case engine.NOT:
		return MakeBool(!b.Bool.Bool), nil
	case engine.NEG, engine.POS:
		return nil, fmt.Errorf("%w: cannot perform unary operation %s on bool", engine.ErrUnary, op)
	default:
		return nil, fmt.Errorf("%w: unexpected operator id %s for bool", engine.ErrUnary, op)
	}
}

func (b *BoolValue) Type() *types.DataType {
	return types.BoolType
}

func (b *BoolValue) RawValue() any {
	if !b.Valid {
		return nil
	}

	return b.Bool.Bool
}

func (b *BoolValue) Array(v ...ScalarValue) (ArrayValue, error) {
	pgtArr := make([]pgtype.Bool, len(v)+1)
	pgtArr[0] = b.Bool
	for j, val := range v {
		if boolVal, ok := val.(*BoolValue); !ok {
			return nil, makeArrTypeErr(b, val)
		} else {
			pgtArr[j+1] = boolVal.Bool
		}
	}

	arr := newValidArr(pgtArr)

	return &BoolArrayValue{
		OneDArray: arr,
	}, nil
}

func (b *BoolValue) Cast(t *types.DataType) (Value, error) {
	if b.Null() {
		return MakeNull(t)
	}

	switch *t {
	case *types.IntType:
		if b.Bool.Bool {
			return MakeInt8(1), nil
		}

		return MakeInt8(0), nil
	case *types.TextType:
		return MakeText(strconv.FormatBool(b.Bool.Bool)), nil
	case *types.BoolType:
		return b, nil
	default:
		return nil, castErr(fmt.Errorf("cannot cast bool to %s", t))
	}
}

func MakeBlob(b []byte) *BlobValue {
	return &BlobValue{
		bts: b,
	}
}

type BlobValue struct {
	bts []byte
}

func (b *BlobValue) Null() bool {
	return b.bts == nil
}

func (b *BlobValue) Compare(v Value, op engine.ComparisonOp) (*BoolValue, error) {
	if res, early := nullCmp(b, v, op); early {
		return res, nil
	}

	val2, ok := v.(*BlobValue)
	if !ok {
		return nil, makeTypeErr(b, v)
	}

	var b2 bool
	switch op {
	case engine.EQUAL:
		b2 = string(b.bts) == string(val2.bts)
	case engine.IS_DISTINCT_FROM:
		b2 = string(b.bts) != string(val2.bts)
	default:
		return nil, fmt.Errorf("%w: cannot use comparison operator %s with type %s", engine.ErrComparison, b.Type(), op)
	}

	return MakeBool(b2), nil
}

func (b *BlobValue) Arithmetic(v ScalarValue, op engine.ArithmeticOp) (ScalarValue, error) {
	if res, early := checkScalarNulls(b, v); early {
		return res, nil
	}

	val2, ok := v.(*BlobValue)
	if !ok {
		return nil, makeTypeErr(b, v)
	}

	if op == engine.CONCAT {
		return MakeBlob(append(b.bts, val2.bts...)), nil
	}

	return nil, fmt.Errorf("%w: cannot perform arithmetic operation %s on blob", engine.ErrArithmetic, op)
}

func (b *BlobValue) Unary(op engine.UnaryOp) (ScalarValue, error) {
	return nil, fmt.Errorf("%w: cannot perform unary operation on blob", engine.ErrUnary)
}

func (b *BlobValue) Type() *types.DataType {
	return types.BlobType
}

func (b *BlobValue) RawValue() any {
	return b.bts
}

func (b *BlobValue) Array(v ...ScalarValue) (ArrayValue, error) {
	pgtArr := make([]*BlobValue, len(v)+1)
	pgtArr[0] = b
	for j, val := range v {
		if blobVal, ok := val.(*BlobValue); !ok {
			return nil, makeArrTypeErr(b, val)
		} else {
			pgtArr[j+1] = blobVal
		}
	}

	arr := newValidArr(pgtArr)

	return &BlobArrayValue{
		OneDArray: arr,
	}, nil
}

func (b *BlobValue) Cast(t *types.DataType) (Value, error) {
	switch *t {
	case *types.IntType:
		i, err := strconv.ParseInt(string(b.bts), 10, 64)
		if err != nil {
			return nil, castErr(err)
		}

		return MakeInt8(i), nil
	case *types.TextType:
		return MakeText(string(b.bts)), nil
	case *types.BlobType:
		return b, nil
	default:
		return nil, castErr(fmt.Errorf("cannot cast blob to %s", t))
	}
}

var _ pgtype.BytesScanner = (*BlobValue)(nil)
var _ pgtype.BytesValuer = (*BlobValue)(nil)

// ScanBytes implements the pgtype.BytesScanner interface.
func (b *BlobValue) ScanBytes(src []byte) error {
	if src == nil {
		b.bts = nil
		return nil
	}

	// copy the src bytes into the prealloc bytes
	b.bts = make([]byte, len(src))
	copy(b.bts, src)
	return nil
}

// Value implements the driver.Valuer interface.
func (b *BlobValue) Value() (driver.Value, error) {
	if b.Null() {
		return nil, nil
	}

	return b.bts, nil
}

// BytesValue implements the pgtype.BytesValuer interface.
func (b *BlobValue) BytesValue() ([]byte, error) {
	if b.Null() {
		return nil, nil
	}

	return b.bts, nil
}

func MakeUUID(u *types.UUID) *UUIDValue {
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

func (u *UUIDValue) Compare(v Value, op engine.ComparisonOp) (*BoolValue, error) {
	if res, early := nullCmp(u, v, op); early {
		return res, nil
	}

	val2, ok := v.(*UUIDValue)
	if !ok {
		return nil, makeTypeErr(u, v)
	}

	var b bool
	switch op {
	case engine.EQUAL:
		b = u.Bytes == val2.Bytes
	case engine.IS_DISTINCT_FROM:
		b = u.Bytes != val2.Bytes
	default:
		return nil, fmt.Errorf("%w: cannot use comparison operator %s with type %s", engine.ErrComparison, u.Type(), op)
	}

	return MakeBool(b), nil
}

func (u *UUIDValue) Arithmetic(v ScalarValue, op engine.ArithmeticOp) (ScalarValue, error) {
	return nil, fmt.Errorf("%w: cannot perform arithmetic operation on uuid", engine.ErrArithmetic)
}

func (u *UUIDValue) Unary(op engine.UnaryOp) (ScalarValue, error) {
	return nil, fmt.Errorf("%w: cannot perform unary operation on uuid", engine.ErrUnary)
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
			return nil, makeArrTypeErr(u, val)
		} else {
			pgtArr[j+1] = uuidVal.UUID
		}
	}

	arr := newValidArr(pgtArr)

	return &UuidArrayValue{
		OneDArray: arr,
	}, nil
}

func (u *UUIDValue) Cast(t *types.DataType) (Value, error) {
	if u.Null() {
		return MakeNull(t)
	}

	switch *t {
	case *types.TextType:
		return MakeText(types.UUID(u.Bytes).String()), nil
	case *types.BlobType:
		return MakeBlob(u.Bytes[:]), nil
	case *types.UUIDType:
		return u, nil
	default:
		return nil, castErr(fmt.Errorf("cannot cast uuid to %s", t))
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

	return decimal.NewFromBigInt(n.Int, n.Exp)
}

func MakeDecimal(d *decimal.Decimal) *DecimalValue {
	if d == nil {
		return &DecimalValue{
			Numeric: pgtype.Numeric{
				Valid: false,
			},
		}
	}

	prec := d.Precision()
	scale := d.Scale()
	return &DecimalValue{
		Numeric:  pgTypeFromDec(d),
		metadata: &precAndScale{prec, scale},
	}
}

type DecimalValue struct {
	pgtype.Numeric
	metadata *precAndScale // can be nil
}

type precAndScale [2]uint16

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

	d2, err := decimal.NewFromBigInt(d.Int, d.Exp)
	if err != nil {
		return nil, err
	}

	if d.metadata != nil {
		meta := *d.metadata
		err = d2.SetPrecisionAndScale(meta[0], meta[1])
		if err != nil {
			return nil, err
		}
	}

	return d2, nil
}

func (d *DecimalValue) Compare(v Value, op engine.ComparisonOp) (*BoolValue, error) {
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

func (d *DecimalValue) Arithmetic(v ScalarValue, op engine.ArithmeticOp) (ScalarValue, error) {
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
	case engine.ADD:
		d2, err = decimal.Add(dec1, dec2)
	case engine.SUB:
		d2, err = decimal.Sub(dec1, dec2)
	case engine.MUL:
		d2, err = decimal.Mul(dec1, dec2)
	case engine.DIV:
		d2, err = decimal.Div(dec1, dec2)
	case engine.EXP:
		d2, err = decimal.Pow(dec1, dec2)
	case engine.MOD:
		d2, err = decimal.Mod(dec1, dec2)
	default:
		return nil, fmt.Errorf("%w: unexpected operator id %d for decimal", engine.ErrArithmetic, op)
	}
	if err != nil {
		return nil, err
	}

	err = d2.SetPrecisionAndScale(dec1.Precision(), dec1.Scale())
	if err != nil {
		return nil, err
	}

	return MakeDecimal(d2), nil
}

func (d *DecimalValue) Unary(op engine.UnaryOp) (ScalarValue, error) {
	if d.Null() {
		return d, nil
	}

	switch op {
	case engine.NEG:
		dec, err := d.dec()
		if err != nil {
			return nil, err
		}

		err = dec.Neg()
		if err != nil {
			return nil, err
		}

		return MakeDecimal(dec), nil
	case engine.POS:
		return d, nil
	default:
		return nil, fmt.Errorf("%w: unexpected operator id %s for decimal", engine.ErrUnary, op)
	}
}

func (d *DecimalValue) Type() *types.DataType {
	if d.metadata == nil {
		return types.DecimalType
	}

	t := types.DecimalType.Copy()
	t.Metadata = *d.metadata
	return t
}

func (d *DecimalValue) RawValue() any {
	if !d.Valid {
		return nil
	}
	dec, err := d.dec()
	if err != nil {
		panic(err)
	}

	return dec
}

func (d *DecimalValue) Array(v ...ScalarValue) (ArrayValue, error) {
	pgtArr := make([]pgtype.Numeric, len(v)+1)
	pgtArr[0] = d.Numeric
	for j, val := range v {
		if decVal, ok := val.(*DecimalValue); !ok {
			return nil, makeArrTypeErr(d, val)
		} else {
			pgtArr[j+1] = decVal.Numeric
		}
	}

	metaCopy := *d.metadata

	return &DecimalArrayValue{
		OneDArray: newValidArr(pgtArr),
		metadata:  &metaCopy,
	}, nil
}

func (d *DecimalValue) Cast(t *types.DataType) (Value, error) {
	if t.Name == types.NumericStr {
		if t.IsArray {
			return nil, castErr(errors.New("cannot cast decimal to decimal array"))
		}

		// otherwise, we need to alter the precision and scale

		dec, err := d.dec()
		if err != nil {
			return nil, castErr(err)
		}

		err = dec.SetPrecisionAndScale(t.Metadata[0], t.Metadata[1])
		if err != nil {
			return nil, castErr(err)
		}

		return MakeDecimal(dec), nil
	}

	switch *t {
	case *types.IntType:
		dec, err := d.dec()
		if err != nil {
			return nil, castErr(err)
		}

		i, err := dec.Int64()
		if err != nil {
			return nil, castErr(err)
		}

		return MakeInt8(i), nil
	case *types.TextType:
		dec, err := d.dec()
		if err != nil {
			return nil, castErr(err)
		}

		return MakeText(dec.String()), nil
	default:
		return nil, castErr(fmt.Errorf("cannot cast decimal to %s", t))
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
		OneDArray: newValidArr(pgInts),
	}
}

// OneDArray array intercepts the pgtype SetDimensions method to ensure that all arrays we scan are
// 1D arrays. This is because we do not support multi-dimensional arrays.
type OneDArray[T any] struct {
	pgtype.Array[T]
}

var _ pgtype.ArraySetter = (*OneDArray[any])(nil)
var _ pgtype.ArrayGetter = (*OneDArray[any])(nil)

func (a *OneDArray[T]) SetDimensions(dims []pgtype.ArrayDimension) error {
	// if len(dims) is 0, it is null.
	// if len(dims) is 1, it is a 1D array.
	// Kwil does not support multi-dimensional arrays.
	switch len(dims) {
	case 0, 1:
		return a.Array.SetDimensions(dims)
	default:
		return fmt.Errorf("%w: expected 1 dimension, got %d", engine.ErrArrayDimensionality, len(dims))
	}
}

func (a *OneDArray[T]) Value() (driver.Value, error) {
	// for some reason, not having this Value method causes the OneDArray type
	// to not function despite implementing the pgtype.ArrayGetter interface.
	return a.Array, nil
}

type IntArrayValue struct {
	OneDArray[pgtype.Int8]
}

func (a *IntArrayValue) Null() bool {
	return !a.Valid
}

func (a *IntArrayValue) Compare(v Value, op engine.ComparisonOp) (*BoolValue, error) {
	return cmpArrs(a, v, op)
}

func (a *IntArrayValue) Len() int32 {
	return int32(len(a.Elements))
}

func (a *IntArrayValue) Get(i int32) (ScalarValue, error) {
	if i < 1 || i > a.Len() {
		return nil, engine.ErrIndexOutOfBounds
	}

	return &Int8Value{a.Elements[i-1]}, nil // indexing is 1-based
}

// allocArr checks that the array has index i, and if NOT, it allocates enough space to set the value.
func allocArr[T any](p *pgtype.Array[T], i int32) error {
	if i < 1 {
		return engine.ErrIndexOutOfBounds
	}

	if i > int32(len(p.Elements)) {
		// Allocate enough space to set the value.
		// This matches the behavior of Postgres.
		newVal := make([]T, i)
		copy(newVal, p.Elements)
		p.Elements = newVal
		p.Dims[0] = pgtype.ArrayDimension{
			Length:     i,
			LowerBound: 1,
		}
	}

	return nil
}

func (a *IntArrayValue) Set(i int32, v ScalarValue) error {
	// we do NOT need to worry about nulls here. Postgres will automatically make an array
	// NOT null if we set a value in it.
	// to test it:
	// CREATE TABLE test (arr int[]);
	// INSERT INTO test VALUES (NULL);
	// UPDATE test SET arr[1] = 1;
	err := allocArr(&a.Array, i)
	if err != nil {
		return err
	}

	val, ok := v.(*Int8Value)
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
		return MakeNull(t)
	}

	if t.Name == types.NumericStr {
		if !t.IsArray {
			return nil, castErr(errors.New("cannot cast int array to decimal"))
		}

		return castArrWithPtr(a, func(i int64) (*decimal.Decimal, error) {
			return decimal.NewExplicit(strconv.FormatInt(i, 10), t.Metadata[0], t.Metadata[1])
		}, newDecArrFn(t))
	}

	switch *t {
	case *types.IntArrayType:
		return a, nil
	case *types.TextArrayType:
		return castArr(a, func(i int64) (string, error) { return strconv.FormatInt(i, 10), nil }, newTextArrayValue)
	case *types.BoolArrayType:
		return castArr(a, func(i int64) (bool, error) { return i != 0, nil }, newBoolArrayValue)
	default:
		return nil, castErr(fmt.Errorf("cannot cast int array to %s", t))
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
		OneDArray: newValidArr(vals),
	}
}

type TextArrayValue struct {
	OneDArray[pgtype.Text]
}

func (a *TextArrayValue) Null() bool {
	return !a.Valid
}

func (a *TextArrayValue) Compare(v Value, op engine.ComparisonOp) (*BoolValue, error) {
	return cmpArrs(a, v, op)
}

func (a *TextArrayValue) Len() int32 {
	return int32(len(a.Elements))
}

func (a *TextArrayValue) Get(i int32) (ScalarValue, error) {
	if i < 1 || i > a.Len() {
		return nil, engine.ErrIndexOutOfBounds
	}

	return &TextValue{a.Elements[i-1]}, nil
}

func (a *TextArrayValue) Set(i int32, v ScalarValue) error {
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
	if t.Name == types.NumericStr {
		if !t.IsArray {
			return nil, castErr(errors.New("cannot cast text array to decimal"))
		}

		return castArrWithPtr(a, func(s string) (*decimal.Decimal, error) {
			return decimal.NewExplicit(s, t.Metadata[0], t.Metadata[1])
		}, newDecArrFn(t))
	}

	switch *t {
	case *types.IntArrayType:
		return castArr(a, func(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) }, newIntArr)
	case *types.BoolArrayType:
		return castArr(a, strconv.ParseBool, newBoolArrayValue)
	case *types.UUIDArrayType:
		return castArrWithPtr(a, types.ParseUUID, newUUIDArrayValue)
	case *types.TextArrayType:
		return a, nil
	case *types.BlobArrayType:
		return castArr(a, func(s string) ([]byte, error) { return []byte(s), nil }, newBlobArrayValue)
	default:
		return nil, castErr(fmt.Errorf("cannot cast text array to %s", t))
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
		v, err := c.Get(i + 1) // SQL Indexes are 1-based
		if err != nil {
			return *new(D), castErr(err)
		}

		// if the value is nil, we dont need to do anything; a nil value is already
		// in the array
		if !v.Null() {
			raw, ok := v.RawValue().(A)
			if !ok {
				// should never happen unless I messed up the usage of generics or implementation
				// of RawValue
				return *new(D), castErr(fmt.Errorf("internal bug: unexpected type %T", v.RawValue()))
			}

			res[i], err = get(raw)
			if err != nil {
				return *new(D), castErr(err)
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
		OneDArray: newValidArr(vals),
	}
}

type BoolArrayValue struct {
	OneDArray[pgtype.Bool]
}

func (a *BoolArrayValue) Null() bool {
	return !a.Valid
}

func (a *BoolArrayValue) Compare(v Value, op engine.ComparisonOp) (*BoolValue, error) {
	return cmpArrs(a, v, op)
}

func (a *BoolArrayValue) Len() int32 {
	return int32(len(a.Elements))
}

func (a *BoolArrayValue) Get(i int32) (ScalarValue, error) {
	if i < 1 || i > a.Len() {
		return nil, engine.ErrIndexOutOfBounds
	}

	return &BoolValue{a.Elements[i-1]}, nil
}

func (a *BoolArrayValue) Set(i int32, v ScalarValue) error {
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
	switch *t {
	case *types.TextArrayType:
		return castArr(a, func(b bool) (string, error) { return strconv.FormatBool(b), nil }, newTextArrayValue)
	case *types.IntArrayType:
		return castArr(a, func(b bool) (int64, error) {
			if b {
				return 1, nil
			} else {
				return 0, nil
			}
		}, newIntArr)
	case *types.BoolArrayType:
		return a, nil
	default:
		return nil, castErr(fmt.Errorf("cannot cast bool array to %s", t))
	}
}

func newNullDecArr(t *types.DataType) *DecimalArrayValue {
	if t.Name != types.NumericStr {
		panic("internal bug: expected decimal type")
	}
	if !t.IsArray {
		panic("internal bug: expected array type")
	}
	precCopy := t.Metadata[0]
	scaleCopy := t.Metadata[1]
	return &DecimalArrayValue{
		OneDArray: OneDArray[pgtype.Numeric]{
			Array: pgtype.Array[pgtype.Numeric]{Valid: false},
		},
		metadata: &precAndScale{precCopy, scaleCopy},
	}
}

// newDecArrFn returns a function that creates a new DecimalArrayValue.
// It is used for type casting.
func newDecArrFn(t *types.DataType) func(d []*decimal.Decimal) *DecimalArrayValue {
	return func(d []*decimal.Decimal) *DecimalArrayValue {
		return newDecimalArrayValue(d, t)
	}
}

func newDecimalArrayValue(d []*decimal.Decimal, t *types.DataType) *DecimalArrayValue {
	vals := make([]pgtype.Numeric, len(d))
	for i, v := range d {
		var newDec pgtype.Numeric
		if v == nil {
			newDec = pgtype.Numeric{Valid: false}
		} else {
			newDec = pgTypeFromDec(v)
		}

		vals[i] = newDec
	}

	precCopy := t.Metadata[0]
	scaleCopy := t.Metadata[1]

	return &DecimalArrayValue{
		OneDArray: newValidArr(vals),
		metadata:  &precAndScale{precCopy, scaleCopy},
	}
}

type DecimalArrayValue struct {
	OneDArray[pgtype.Numeric]               // we embed decimal value here because we need to track the precision and scale
	metadata                  *precAndScale // can be nil
}

func (a *DecimalArrayValue) Null() bool {
	return !a.Valid
}

func (a *DecimalArrayValue) Compare(v Value, op engine.ComparisonOp) (*BoolValue, error) {
	return cmpArrs(a, v, op)
}

// cmpArrs compares two Kwil array types.
func cmpArrs[M ArrayValue](a M, b Value, op engine.ComparisonOp) (*BoolValue, error) {
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

		for i := int32(1); i <= a.Len(); i++ {
			v1, err := a.Get(i)
			if err != nil {
				return false, err
			}

			v2, err := b.Get(i)
			if err != nil {
				return false, err
			}

			if v1.Null() && v2.Null() {
				continue
			}

			if v1.Null() || v2.Null() {
				return false, nil
			}

			res, err := v1.Compare(v2, engine.EQUAL)
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
	case engine.EQUAL:
		return MakeBool(eq), nil
	case engine.IS_DISTINCT_FROM:
		return MakeBool(!eq), nil
	default:
		return nil, fmt.Errorf("%w: only =, IS DISTINCT FROM are supported for array comparison", engine.ErrComparison)
	}
}

func (a *DecimalArrayValue) Len() int32 {
	return int32(len(a.Elements))
}

func (a *DecimalArrayValue) Get(i int32) (ScalarValue, error) {
	if i < 1 || i > a.Len() {
		return nil, engine.ErrIndexOutOfBounds
	}

	return &DecimalValue{Numeric: a.Elements[i-1]}, nil
}

func (a *DecimalArrayValue) Set(i int32, v ScalarValue) error {
	err := allocArr(&a.Array, i)
	if err != nil {
		return err
	}

	val, ok := v.(*DecimalValue)
	if !ok {
		return fmt.Errorf("cannot set non-decimal value in decimal array")
	}

	if val.metadata != nil && a.metadata != nil && *val.metadata != *a.metadata {
		valMeta := *val.metadata
		aMeta := *a.metadata
		return fmt.Errorf("cannot set decimal with precision %d and scale %d in array with precision %d and scale %d", valMeta[0], valMeta[1], aMeta[0], aMeta[1])
	}

	a.Elements[i-1] = val.Numeric
	return nil
}

func (a *DecimalArrayValue) Type() *types.DataType {
	if a.metadata == nil {
		return types.DecimalArrayType
	}

	t := types.DecimalArrayType.Copy()
	t.Metadata = *a.metadata
	return t
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
	if t.Name == types.NumericStr {
		if !t.IsArray {
			return nil, castErr(errors.New("cannot cast decimal array to decimal"))
		}

		// otherwise, we need to alter the precision and scale
		res := make([]*decimal.Decimal, a.Len())
		for i := int32(1); i <= a.Len(); i++ {
			v, err := a.Get(i)
			if err != nil {
				return nil, err
			}

			dec, err := v.(*DecimalValue).dec()
			if err != nil {
				return nil, err
			}

			// we need to make a copy of the decimal because SetPrecisionAndScale
			// will modify the decimal in place.
			dec2, err := decimal.NewExplicit(dec.String(), dec.Precision(), dec.Scale())
			if err != nil {
				return nil, err
			}

			err = dec2.SetPrecisionAndScale(t.Metadata[0], t.Metadata[1])
			if err != nil {
				return nil, err
			}

			res[i-1] = dec
		}

		return newDecimalArrayValue(res, t), nil
	}

	switch *t {
	case *types.TextArrayType:
		return castArr(a, func(d *decimal.Decimal) (string, error) { return d.String(), nil }, newTextArrayValue)
	case *types.IntArrayType:
		return castArr(a, func(d *decimal.Decimal) (int64, error) { return d.Int64() }, newIntArr)
	case *types.DecimalArrayType:
		return a, nil
	default:
		return nil, castErr(fmt.Errorf("cannot cast decimal array to %s", t))
	}
}

func newBlobArrayValue(b []*[]byte) *BlobArrayValue {
	vals := make([]*BlobValue, len(b))
	for i, v := range b {
		if v == nil {
			vals[i] = &BlobValue{bts: nil}
		} else {
			vals[i] = &BlobValue{bts: *v}
		}
	}

	return &BlobArrayValue{
		OneDArray: newValidArr(vals),
	}
}

type BlobArrayValue struct {
	// we embed BlobValue because unlike other types, there is no native pgtype embedded within
	// blob value that allows pgx to scan the value into the struct.
	OneDArray[*BlobValue]
}

func (a *BlobArrayValue) Null() bool {
	return !a.Valid
}

// A special Value method is needed since pgx handles byte slices differently than other types.
func (a *BlobArrayValue) Value() (driver.Value, error) {
	var btss [][]byte
	for _, v := range a.Elements {
		if v != nil {
			btss = append(btss, v.bts)
		} else {
			btss = append(btss, nil)
		}
	}

	return btss, nil
}

func (a *BlobArrayValue) Compare(v Value, op engine.ComparisonOp) (*BoolValue, error) {
	return cmpArrs(a, v, op)
}

func (a *BlobArrayValue) Len() int32 {
	return int32(len(a.Elements))
}

func (a *BlobArrayValue) Get(i int32) (ScalarValue, error) {
	if i < 1 || i > a.Len() {
		return nil, engine.ErrIndexOutOfBounds
	}

	return a.Elements[i-1], nil
}

func (a *BlobArrayValue) Set(i int32, v ScalarValue) error {
	err := allocArr(&a.Array, i)
	if err != nil {
		return err
	}

	val, ok := v.(*BlobValue)
	if !ok {
		return fmt.Errorf("cannot set non-blob value in blob array")
	}

	a.Elements[i-1] = val
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
			res[i] = make([]byte, len(v.bts))
			copy(res[i], v.bts)
		}
	}

	return res
}

func (a *BlobArrayValue) Cast(t *types.DataType) (Value, error) {
	switch *t {
	case *types.TextArrayType:
		return castArr(a, func(b []byte) (string, error) { return string(b), nil }, newTextArrayValue)
	case *types.BlobArrayType:
		return a, nil
	default:
		return nil, castErr(fmt.Errorf("cannot cast blob array to %s", t))
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
		OneDArray: newValidArr(vals),
	}
}

type UuidArrayValue struct {
	OneDArray[pgtype.UUID]
}

func (a *UuidArrayValue) Null() bool {
	return !a.Valid
}

func (a *UuidArrayValue) Compare(v Value, op engine.ComparisonOp) (*BoolValue, error) {
	return cmpArrs(a, v, op)
}

func (a *UuidArrayValue) Len() int32 {
	return int32(len(a.Elements))
}

func (a *UuidArrayValue) Get(i int32) (ScalarValue, error) {
	if i < 1 || i > a.Len() {
		return nil, engine.ErrIndexOutOfBounds
	}

	return &UUIDValue{a.Elements[i-1]}, nil
}

func (a *UuidArrayValue) Set(i int32, v ScalarValue) error {
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
	switch *t {
	case *types.TextArrayType:
		return castArr(a, func(u *types.UUID) (string, error) { return u.String(), nil }, newTextArrayValue)
	case *types.UUIDArrayType:
		return a, nil
	case *types.BlobArrayType:
		return castArr(a, func(u *types.UUID) ([]byte, error) { return u.Bytes(), nil }, newBlobArrayValue)
	default:
		return nil, castErr(fmt.Errorf("cannot cast uuid array to %s", t))
	}
}

// EmptyRecordValue creates a new empty record value.
func EmptyRecordValue() *RecordValue {
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

func (o *RecordValue) Compare(v Value, op engine.ComparisonOp) (*BoolValue, error) {
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

			eq, err := o.Fields[field].Compare(v2, engine.EQUAL)
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
	case engine.EQUAL:
		return MakeBool(isSame), nil
	default:
		return nil, fmt.Errorf("%w: cannot use comparison operator %s with record type", engine.ErrComparison, op)
	}
}

func (o *RecordValue) Type() *types.DataType {
	return &types.DataType{
		Name: "record", // special type that is NOT in the types package
	}
}

func (o *RecordValue) RawValue() any {
	return o.Fields
}

func (o *RecordValue) Cast(t *types.DataType) (Value, error) {
	return nil, castErr(fmt.Errorf("cannot cast record to %s", t))
}

func cmpIntegers(a, b int, op engine.ComparisonOp) (*BoolValue, error) {
	switch op {
	case engine.EQUAL:
		return MakeBool(a == b), nil
	case engine.LESS_THAN:
		return MakeBool(a < b), nil
	case engine.GREATER_THAN:
		return MakeBool(a > b), nil
	case engine.IS_DISTINCT_FROM:
		return MakeBool(a != b), nil
	default:
		return nil, fmt.Errorf("%w: cannot use comparison operator %s with numeric types", engine.ErrComparison, op)
	}
}

// StringifyValue converts a value to a string.
// It can be reversed using ParseValue.
func StringifyValue(v Value) (string, error) {
	if v.Null() {
		return "NULL", nil
	}

	array, ok := v.(ArrayValue)
	if ok {
		// we will convert each element to a string and join them with a comma
		strs := make([]string, array.Len())
		for i := int32(1); i <= array.Len(); i++ {
			val, err := array.Get(i)
			if err != nil {
				return "", err
			}

			str, err := StringifyValue(val)
			if err != nil {
				return "", err
			}

			strs[i-1] = str
		}

		return strings.Join(strs, ","), nil
	}

	switch val := v.(type) {
	case *TextValue:
		return val.Text.String, nil
	case *Int8Value:
		return strconv.FormatInt(val.Int64, 10), nil
	case *BoolValue:
		return strconv.FormatBool(val.Bool.Bool), nil
	case *UUIDValue:
		return types.UUID(val.UUID.Bytes).String(), nil
	case *DecimalValue:
		dec, err := val.dec()
		if err != nil {
			return "", err
		}

		return dec.String(), nil
	case *BlobValue:
		return string(val.bts), nil
	case *RecordValue:
		return "", fmt.Errorf("cannot convert record to string")
	default:
		return "", fmt.Errorf("unexpected type %T", v)
	}
}

// ParseValue parses a string into a value.
// It is the reverse of StringifyValue.
func ParseValue(s string, t *types.DataType) (Value, error) {
	if s == "NULL" {
		return MakeNull(t)
	}

	if t.IsArray {
		return parseArray(s, t)
	}

	if t.Name == types.NumericStr {
		dec, err := decimal.NewFromString(s)
		if err != nil {
			return nil, err
		}

		return MakeDecimal(dec), nil
	}

	switch *t {
	case *types.TextType:
		return MakeText(s), nil
	case *types.IntType:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}

		return MakeInt8(i), nil
	case *types.BoolType:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return nil, err
		}

		return MakeBool(b), nil
	case *types.UUIDType:
		u, err := types.ParseUUID(s)
		if err != nil {
			return nil, err
		}

		return MakeUUID(u), nil
	case *types.BlobType:
		return MakeBlob([]byte(s)), nil
	default:
		return nil, fmt.Errorf("unexpected type %s", t)
	}
}

// parseArray parses a string into an array value.
func parseArray(s string, t *types.DataType) (ArrayValue, error) {
	if s == "NULL" {
		nv, err := MakeNull(t)
		if err != nil {
			return nil, err
		}

		nva, ok := nv.(ArrayValue)
		if !ok {
			return nil, fmt.Errorf("unexpected type for null array %T", nv)
		}

		return nva, nil
	}

	// we will parse the string into individual values and then cast them to the
	// correct type
	strs := strings.Split(s, ",")
	fields := make([]ScalarValue, len(strs))
	scalarType := t.Copy()
	scalarType.IsArray = false
	for i, str := range strs {
		val, err := ParseValue(str, scalarType)
		if err != nil {
			return nil, err
		}

		scalar, ok := val.(ScalarValue)
		if !ok {
			return nil, fmt.Errorf("unexpected type %T", val)
		}

		fields[i] = scalar
	}

	if len(fields) == 0 {
		// if 0-length, then we return a new zero-length array
		zv, err := NewZeroValue(t)
		if err != nil {
			return nil, err
		}

		zva, ok := zv.(ArrayValue)
		if !ok {
			return nil, fmt.Errorf("unexpected type %T", zv)
		}

		return zva, nil
	}

	arrType, err := fields[0].Array(fields[1:]...)
	if err != nil {
		return nil, err
	}

	return arrType, nil
}

func castErr(e error) error {
	return fmt.Errorf("%w: %w", engine.ErrCast, e)
}
