package interpreter

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"

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
	goTypeToValue   = map[reflect.Type]ValueMapping{}
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
				return &IntValue{
					Val: 0,
				}, nil
			},
		},
		ValueMapping{
			KwilType: types.TextType,
			ZeroValue: func() (Value, error) {
				return &TextValue{
					Val: "",
				}, nil
			},
		},
		ValueMapping{
			KwilType: types.BoolType,
			ZeroValue: func() (Value, error) {
				return &BoolValue{
					Val: false,
				}, nil
			},
		},
		ValueMapping{
			KwilType: types.BlobType,
			ZeroValue: func() (Value, error) {
				return &BlobValue{
					Val: []byte{},
				}, nil
			},
		},
		ValueMapping{
			KwilType: types.UUIDType,
			ZeroValue: func() (Value, error) {
				return &UUIDValue{
					Val: types.UUID{},
				}, nil
			},
		},
		ValueMapping{
			KwilType: types.DecimalType,
			ZeroValue: func() (Value, error) {
				dec, err := decimal.NewFromString("0")
				if err != nil {
					return nil, err
				}
				return &DecimalValue{
					Dec: dec,
				}, nil
			},
		},
		ValueMapping{
			KwilType: types.IntArrayType,
			ZeroValue: func() (Value, error) {
				return &IntArrayValue{
					Val: []*int64{},
				}, nil
			},
		},
		ValueMapping{
			KwilType: types.TextArrayType,
			ZeroValue: func() (Value, error) {
				return &TextArrayValue{
					Val: []*string{},
				}, nil
			},
		},
		ValueMapping{
			KwilType: types.BoolArrayType,
			ZeroValue: func() (Value, error) {
				return &BoolArrayValue{
					Val: []*bool{},
				}, nil
			},
		},
		ValueMapping{
			KwilType: types.BlobArrayType,
			ZeroValue: func() (Value, error) {
				return &BlobArrayValue{
					Val: []*[]byte{},
				}, nil
			},
		},
		ValueMapping{
			KwilType: types.DecimalArrayType,
			ZeroValue: func() (Value, error) {
				return &DecimalArrayValue{
					Val: []*decimal.Decimal{},
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
	// DBValue returns a value that the database can read from and write to.
	DBValue() (any, error)
	// Compare compares the variable with another variable using the given comparison operator.
	// It will return a boolean value, or null either of the variables is null.
	Compare(v Value, op ComparisonOp) (Value, error)
	// Type returns the type of the variable.
	Type() *types.DataType
	// RawValue returns the value of the variable.
	RawValue() any
	// Size is the size of the variable in bytes.
	Size() int
	// Cast casts the variable to the given type.
	Cast(t *types.DataType) (Value, error)
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

// NewValue creates a new Value from the given any val.
func NewValue(v any) (Value, error) {
	switch v := v.(type) {
	case int64:
		return &IntValue{Val: v}, nil
	case int:
		return &IntValue{Val: int64(v)}, nil
	case string:
		return &TextValue{Val: v}, nil
	case bool:
		return &BoolValue{Val: v}, nil
	case []byte:
		return &BlobValue{Val: v}, nil
	case *types.UUID:
		return &UUIDValue{Val: *v}, nil
	case types.UUID:
		return &UUIDValue{Val: v}, nil
	case *decimal.Decimal:
		return &DecimalValue{Dec: v}, nil
	case []int64:
		return &IntArrayValue{
			Val: safePtrArr(v),
		}, nil
	case []*int64:
		return &IntArrayValue{
			Val: v,
		}, nil
	case []string:
		return &TextArrayValue{
			Val: safePtrArr(v),
		}, nil
	case []*string:
		return &TextArrayValue{
			Val: v,
		}, nil
	case []bool:
		return &BoolArrayValue{
			Val: safePtrArr(v),
		}, nil
	case []*bool:
		return &BoolArrayValue{
			Val: v,
		}, nil
	case [][]byte:
		var res []*[]byte
		for _, val := range v {
			val2 := val
			res = append(res, &val2)
		}

		return &BlobArrayValue{
			Val: res,
		}, nil
	case []*[]byte:
		return &BlobArrayValue{
			Val: v,
		}, nil
	case []*decimal.Decimal:
		if len(v) == 0 {
			return nil, fmt.Errorf("cannot infer type from decimal empty array")
		}

		_, err := types.NewDecimalType(v[0].Precision(), v[0].Scale())
		if err != nil {
			return nil, err
		}

		return &DecimalArrayValue{
			Val: v,
		}, nil
	case []*types.UUID:
		return &UuidArrayValue{
			Val: v,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected type %T", v)
	}
}

// NewNullValue creates a new null value of the given type.
func NewNullValue(t *types.DataType) Value {
	return &NullValue{DataType: t}
}

func makeTypeErr(left, right Value) error {
	return fmt.Errorf("%w: left: %s right: %s", ErrTypeMismatch, left.Type(), right.Type())
}

type IntValue struct {
	Val int64
}

func (i *IntValue) DBValue() (any, error) {
	return &i.Val, nil
}

func (v *IntValue) Compare(v2 Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v2, op); early {
		return res, nil
	}

	val2, ok := v2.(*IntValue)
	if !ok {
		return nil, makeTypeErr(v, v2)
	}

	var b bool
	switch op {
	case equal:
		b = v.Val == val2.Val
	case lessThan:
		b = v.Val < val2.Val
	case greaterThan:
		b = v.Val > val2.Val
	case isDistinctFrom:
		b = v.Val != val2.Val
	default:
		return nil, fmt.Errorf("cannot compare int with operator id %d", op)
	}

	return &BoolValue{Val: b}, nil
}

// nullCmp is a helper function for comparing null values.
// It returns a Value, and a boolean as to whether the caller should return early.
// It is meant to be called from methods for non-null values that might need to compare with null.
func nullCmp(v Value, op ComparisonOp) (Value, bool) {
	if _, ok := v.(*NullValue); !ok {
		return nil, false
	}

	if op == isDistinctFrom {
		return &BoolValue{Val: true}, true
	}

	if op == is {
		return &BoolValue{Val: false}, true
	}

	return &NullValue{DataType: v.Type()}, true
}

func (i *IntValue) Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error) {
	if _, ok := v.(*NullValue); ok {
		return &NullValue{DataType: types.IntType}, nil
	}

	val2, ok := v.(*IntValue)
	if !ok {
		return nil, makeTypeErr(i, v)
	}

	switch op {
	case add:
		return &IntValue{Val: i.Val + val2.Val}, nil
	case sub:
		return &IntValue{Val: i.Val - val2.Val}, nil
	case mul:
		return &IntValue{Val: i.Val * val2.Val}, nil
	case div:
		if val2.Val == 0 {
			return nil, fmt.Errorf("cannot divide by zero")
		}
		return &IntValue{Val: i.Val / val2.Val}, nil
	case mod:
		if val2.Val == 0 {
			return nil, fmt.Errorf("cannot modulo by zero")
		}
		return &IntValue{Val: i.Val % val2.Val}, nil
	default:
		return nil, fmt.Errorf("cannot perform arithmetic operation id %d on type int", op)
	}
}

func (i *IntValue) Unary(op UnaryOp) (ScalarValue, error) {
	switch op {
	case neg:
		return &IntValue{Val: -i.Val}, nil
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
	return i.Val
}

func (i *IntValue) Array(v ...ScalarValue) (ArrayValue, error) {
	res := make([]*int64, len(v)+1)
	res[0] = &i.Val
	for i2, val := range v {
		switch val := val.(type) {
		case *NullValue:
			res[i2+1] = nil
		case *IntValue:
			res[i2+1] = &val.Val
		default:
			return nil, makeTypeErr(i, val)
		}
	}

	return &IntArrayValue{
		Val: res,
	}, nil
}

func (i *IntValue) Cast(t *types.DataType) (Value, error) {
	// we check for decimal first since type switching on it
	// doesn't work, since it has precision and scale
	if t.Name == types.DecimalStr {
		dec, err := decimal.NewFromString(fmt.Sprint(i.Val))
		if err != nil {
			return nil, err
		}

		return &DecimalValue{
			Dec: dec,
		}, nil
	}

	switch t {
	case types.IntType:
		return i, nil
	case types.TextType:
		return &TextValue{Val: fmt.Sprint(i.Val)}, nil
	case types.BoolType:
		return &BoolValue{Val: i.Val != 0}, nil
	default:
		return nil, fmt.Errorf("cannot cast int to %s", t)
	}
}

func (i *IntValue) Size() int {
	return 8
}

type TextValue struct {
	Val string
}

func (s *TextValue) DBValue() (any, error) {
	return &s.Val, nil
}

func (s *TextValue) Compare(v Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v, op); early {
		return res, nil
	}

	val2, ok := v.(*TextValue)
	if !ok {
		return nil, makeTypeErr(s, v)
	}

	var b bool
	switch op {
	case equal:
		b = s.Val == val2.Val
	case lessThan:
		b = s.Val < val2.Val
	case greaterThan:
		b = s.Val > val2.Val
	case isDistinctFrom:
		b = s.Val != val2.Val
	default:
		return nil, fmt.Errorf("unknown comparison operator: %d", op)
	}

	return &BoolValue{Val: b}, nil
}

func (s *TextValue) Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error) {
	val2, ok := v.(*TextValue)
	if !ok {
		return nil, makeTypeErr(s, v)
	}

	if op == concat {
		return &TextValue{Val: s.Val + val2.Val}, nil
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
	return s.Val
}

func (s *TextValue) Array(v ...ScalarValue) (ArrayValue, error) {
	res := make([]*string, len(v)+1)
	res[0] = &s.Val
	for i, val := range v {
		switch val := val.(type) {
		case *NullValue:
			res[i+1] = nil
		case *TextValue:
			res[i+1] = &val.Val
		default:
			return nil, makeTypeErr(s, val)
		}
	}

	return &TextArrayValue{
		Val: res,
	}, nil
}

func (s *TextValue) Cast(t *types.DataType) (Value, error) {
	if t.Name == types.DecimalStr {
		dec, err := decimal.NewFromString(s.Val)
		if err != nil {
			return nil, err
		}

		return &DecimalValue{
			Dec: dec,
		}, nil
	}

	switch t {
	case types.IntType:
		i, err := strconv.Atoi(s.Val)
		if err != nil {
			return nil, err
		}

		return &IntValue{Val: int64(i)}, nil
	case types.TextType:
		return s, nil
	case types.BoolType:
		b, err := strconv.ParseBool(s.Val)
		if err != nil {
			return nil, err
		}

		return &BoolValue{Val: b}, nil
	case types.UUIDType:
		u, err := types.ParseUUID(s.Val)
		if err != nil {
			return nil, err
		}

		return &UUIDValue{Val: *u}, nil
	default:
		return nil, fmt.Errorf("cannot cast string to %s", t)
	}
}

func (s *TextValue) Size() int {
	return len(s.Val)
}

type BoolValue struct {
	Val bool
}

func (b *BoolValue) DBValue() (any, error) {
	return &b.Val, nil
}

func (b *BoolValue) Compare(v Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v, op); early {
		return res, nil
	}

	val2, ok := v.(*BoolValue)
	if !ok {
		return nil, makeTypeErr(b, v)
	}

	var b2 bool
	switch op {
	case equal, is:
		b2 = b.Val == val2.Val
	case isDistinctFrom:
		b2 = b.Val != val2.Val
	default:
		return nil, fmt.Errorf("unknown comparison operator: %d", op)
	}

	return &BoolValue{Val: b2}, nil
}

func (b *BoolValue) Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error) {
	return nil, fmt.Errorf("cannot perform arithmetic operation on bool")
}

func (b *BoolValue) Unary(op UnaryOp) (ScalarValue, error) {
	switch op {
	case not:
		return &BoolValue{Val: !b.Val}, nil
	default:
		return nil, fmt.Errorf("unexpected operator id %d for bool", op)
	}
}

func (b *BoolValue) Type() *types.DataType {
	return types.BoolType
}

func (b *BoolValue) RawValue() any {
	return b.Val
}

func (b *BoolValue) Array(v ...ScalarValue) (ArrayValue, error) {
	res := make([]*bool, len(v)+1)
	res[0] = &b.Val
	for i, val := range v {
		switch val := val.(type) {
		case *NullValue:
			res[i+1] = nil
		case *BoolValue:
			res[i+1] = &val.Val
		default:
			return nil, makeTypeErr(b, val)
		}
	}

	return &BoolArrayValue{
		Val: res,
	}, nil
}

func (b *BoolValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.IntType:
		if b.Val {
			return &IntValue{Val: 1}, nil
		}

		return &IntValue{Val: 0}, nil
	case types.TextType:
		return &TextValue{Val: fmt.Sprint(b.Val)}, nil
	case types.BoolType:
		return b, nil
	default:
		return nil, fmt.Errorf("cannot cast bool to %s", t)
	}
}

func (b *BoolValue) Size() int {
	return 1
}

type BlobValue struct {
	Val []byte
}

func (b *BlobValue) DBValue() (any, error) {
	return &b.Val, nil
}

func (b *BlobValue) Compare(v Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v, op); early {
		return res, nil
	}

	val2, ok := v.(*BlobValue)
	if !ok {
		return nil, makeTypeErr(b, v)
	}

	var b2 bool
	switch op {
	case equal:
		b2 = string(b.Val) == string(val2.Val)
	case isDistinctFrom:
		b2 = string(b.Val) != string(val2.Val)
	default:
		return nil, fmt.Errorf("unknown comparison operator: %d", op)
	}

	return &BoolValue{Val: b2}, nil
}

func (b *BlobValue) Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error) {
	return nil, fmt.Errorf("cannot perform arithmetic operation on blob")
}

func (b *BlobValue) Unary(op UnaryOp) (ScalarValue, error) {
	return nil, fmt.Errorf("cannot perform unary operation on blob")
}

func (b *BlobValue) Type() *types.DataType {
	return types.BlobType
}

func (b *BlobValue) RawValue() any {
	return b.Val
}

func (b *BlobValue) Array(v ...ScalarValue) (ArrayValue, error) {
	res := make([]*[]byte, len(v)+1)
	res[0] = &b.Val
	for i, val := range v {
		switch val := val.(type) {
		case *NullValue:
			res[i+1] = nil
		case *BlobValue:
			res[i+1] = &val.Val
		default:
			return nil, makeTypeErr(b, val)
		}
	}

	return &BlobArrayValue{
		Val: res,
	}, nil
}

func (b *BlobValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.IntType:
		i, err := strconv.ParseInt(string(b.Val), 10, 64)
		if err != nil {
			return nil, err
		}

		return &IntValue{Val: i}, nil
	case types.TextType:
		return &TextValue{Val: string(b.Val)}, nil
	default:
		return nil, fmt.Errorf("cannot cast blob to %s", t)
	}
}

func (b *BlobValue) Size() int {
	return len(b.Val)
}

type UUIDValue struct {
	Val types.UUID
}

func (u *UUIDValue) DBValue() (any, error) {
	return &u.Val, nil
}

func (u *UUIDValue) Compare(v Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v, op); early {
		return res, nil
	}

	val2, ok := v.(*UUIDValue)
	if !ok {
		return nil, makeTypeErr(u, v)
	}

	var b bool
	switch op {
	case equal:
		b = u.Val == val2.Val
	case isDistinctFrom:
		b = u.Val != val2.Val
	default:
		return nil, fmt.Errorf("unknown comparison operator: %d", op)
	}

	return &BoolValue{Val: b}, nil
}

func (u *UUIDValue) Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error) {
	return nil, fmt.Errorf("cannot perform arithmetic operation on uuid")
}

func (u *UUIDValue) Unary(op UnaryOp) (ScalarValue, error) {
	return nil, fmt.Errorf("cannot perform unary operation on uuid")
}

func (u *UUIDValue) Type() *types.DataType {
	return types.UUIDType
}

func (u *UUIDValue) RawValue() any {
	// kwil always handled uuids as pointers
	return &u.Val
}

func (u *UUIDValue) Array(v ...ScalarValue) (ArrayValue, error) {
	res := make([]*types.UUID, len(v)+1)
	res[0] = &u.Val
	for i, val := range v {
		switch val := val.(type) {
		case *NullValue:
			res[i+1] = nil
		case *UUIDValue:
			res[i+1] = &val.Val
		default:
			return nil, makeTypeErr(u, val)
		}
	}

	return &UuidArrayValue{
		Val: res,
	}, nil
}

func (u *UUIDValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.TextType:
		return &TextValue{Val: u.Val.String()}, nil
	case types.BlobType:
		return &BlobValue{Val: u.Val.Bytes()}, nil
	default:
		return nil, fmt.Errorf("cannot cast uuid to %s", t)
	}
}

func (u *UUIDValue) Size() int {
	return 16
}

type DecimalValue struct {
	Dec *decimal.Decimal
}

func (d *DecimalValue) DBValue() (any, error) {
	return d.Dec, nil
}

func (d *DecimalValue) Compare(v Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v, op); early {
		return res, nil
	}

	val2, ok := v.(*DecimalValue)
	if !ok {
		return nil, makeTypeErr(d, v)
	}

	res, err := d.Dec.Cmp(val2.Dec)
	if err != nil {
		return nil, err
	}

	return cmpIntegers(res, 0, op)
}

func (d *DecimalValue) Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error) {
	// we perform an extra check here to ensure scale and precision are the same
	if !v.Type().EqualsStrict(d.Type()) {
		return nil, makeTypeErr(d, v)
	}

	val2, ok := v.(*DecimalValue)
	if !ok {
		return nil, makeTypeErr(d, v)
	}

	var d2 *decimal.Decimal
	var err error
	switch op {
	case add:
		d2, err = decimal.Add(d.Dec, val2.Dec)
	case sub:
		d2, err = decimal.Sub(d.Dec, val2.Dec)
	case mul:
		d2, err = decimal.Mul(d.Dec, val2.Dec)
	case div:
		d2, err = decimal.Div(d.Dec, val2.Dec)
	case mod:
		d2, err = decimal.Mod(d.Dec, val2.Dec)
	default:
		return nil, fmt.Errorf("unexpected operator id %d for decimal", op)
	}
	if err != nil {
		return nil, err
	}

	return &DecimalValue{
		Dec: d2,
	}, nil

}

func (d *DecimalValue) Unary(op UnaryOp) (ScalarValue, error) {
	switch op {
	case neg:
		dec2 := d.Dec.Copy()
		err := dec2.Neg()
		return &DecimalValue{
			Dec: dec2,
		}, err
	case pos:
		return d, nil
	default:
		return nil, fmt.Errorf("unexpected operator id %d for decimal", op)
	}
}

func (d *DecimalValue) Type() *types.DataType {
	res, err := types.NewDecimalType(d.Dec.Precision(), d.Dec.Scale())
	if err != nil {
		panic(err)
	}

	return res
}

func (d *DecimalValue) RawValue() any {
	return d.Dec
}

func (d *DecimalValue) Array(v ...ScalarValue) (ArrayValue, error) {
	res := make([]*decimal.Decimal, len(v)+1)
	res[0] = d.Dec
	for i, val := range v {
		switch val := val.(type) {
		case *NullValue:
			res[i+1] = nil
		case *DecimalValue:
			res[i+1] = val.Dec
		default:
			return nil, makeTypeErr(d, val)
		}
	}

	return &DecimalArrayValue{
		Val: res,
	}, nil
}

type IntArrayValue struct {
	Val []*int64
}

func (a *IntArrayValue) DBValue() (any, error) {
	return &a.Val, nil
}

func (a *IntArrayValue) Compare(v Value, op ComparisonOp) (Value, error) {
	return cmpArrs(a, v, op)
}

func (a *IntArrayValue) Len() int {
	return len(a.Val)
}

func (a *IntArrayValue) Index(i int64) (ScalarValue, error) {
	if i < 1 || i > int64(len(a.Val)) {
		return nil, fmt.Errorf("index out of bounds")
	}

	val := a.Val[i-1]
	if val == nil {
		return &NullValue{DataType: types.IntType}, nil
	}

	return &IntValue{Val: *val}, nil
}

func (a *IntArrayValue) Set(i int64, v ScalarValue) error {
	if i < 1 {
		return fmt.Errorf("index out of bounds")
	}

	if i > int64(len(a.Val)) {
		// Allocate enough space to set the value.
		// This matches the behavior of Postgres.
		newVal := make([]*int64, i)
		copy(newVal, a.Val)
		a.Val = newVal
	}

	val, ok := v.(*IntValue)
	if !ok {
		return fmt.Errorf("cannot set non-int value in int array")
	}

	a.Val[i-1] = &val.Val
	return nil
}

func (a *IntArrayValue) Type() *types.DataType {
	return types.IntArrayType
}

func (a *IntArrayValue) RawValue() any {
	return a.Val
}

func (a *IntArrayValue) Size() int {
	size := 0
	for _, v := range a.Val {
		if v != nil {
			size += 8
		}
	}
	return size
}

func (a *IntArrayValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.TextArrayType:
		res := make([]*string, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				res[i] = new(string)
				*res[i] = strconv.FormatInt(*v, 10)
			}
		}

		return &TextArrayValue{
			Val: res,
		}, nil
	case types.BoolArrayType:
		res := make([]*bool, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				b := *v != 0
				res[i] = &b
			}
		}

		return &BoolArrayValue{
			Val: res,
		}, nil
	case types.DecimalArrayType:
		res := make([]*decimal.Decimal, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				dec, err := decimal.NewFromBigInt(big.NewInt(*v), 0)
				if err != nil {
					return nil, err
				}
				res[i] = dec
			}
		}

		return &DecimalArrayValue{
			Val: res,
		}, nil
	default:
		return nil, fmt.Errorf("cannot cast int array to %s", t)
	}
}

type TextArrayValue struct {
	Val []*string
}

func (a *TextArrayValue) DBValue() (any, error) {
	return &a.Val, nil
}

func (a *TextArrayValue) Compare(v Value, op ComparisonOp) (Value, error) {
	return cmpArrs(a, v, op)
}

func (a *TextArrayValue) Len() int {
	return len(a.Val)
}

func (a *TextArrayValue) Index(i int64) (ScalarValue, error) {
	if i < 1 || i > int64(len(a.Val)) {
		return nil, fmt.Errorf("index out of bounds")
	}

	val := a.Val[i-1]
	if val == nil {
		return &NullValue{DataType: types.TextType}, nil
	}

	return &TextValue{Val: *val}, nil
}

func (a *TextArrayValue) Set(i int64, v ScalarValue) error {
	if i < 1 {
		return fmt.Errorf("index out of bounds")
	}

	if i > int64(len(a.Val)) {
		// Allocate enough space to set the value.
		// This matches the behavior of Postgres.
		newVal := make([]*string, i)
		copy(newVal, a.Val)
		a.Val = newVal
	}

	val, ok := v.(*TextValue)
	if !ok {
		return fmt.Errorf("cannot set non-text value in text array")
	}

	a.Val[i-1] = &val.Val
	return nil
}

func (a *TextArrayValue) Type() *types.DataType {
	return types.TextArrayType
}

func (a *TextArrayValue) RawValue() any {
	return a.Val
}

func (a *TextArrayValue) Size() int {
	size := 0
	for _, v := range a.Val {
		if v != nil {
			size += len(*v)
		}
	}
	return size
}

func (a *TextArrayValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.IntArrayType:
		res := make([]*int64, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				i, err := strconv.ParseInt(*v, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("cannot cast text array to int array: %w", err)
				}
				res[i] = &i
			}
		}

		return &IntArrayValue{
			Val: res,
		}, nil
	case types.BoolArrayType:
		res := make([]*bool, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				b, err := strconv.ParseBool(*v)
				if err != nil {
					return nil, fmt.Errorf("cannot cast text array to bool array: %w", err)
				}
				res[i] = &b
			}
		}

		return &BoolArrayValue{
			Val: res,
		}, nil
	case types.DecimalArrayType:
		res := make([]*decimal.Decimal, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				dec, err := decimal.NewFromString(*v)
				if err != nil {
					return nil, fmt.Errorf("cannot cast text array to decimal array: %w", err)
				}
				res[i] = dec
			}
		}

		return &DecimalArrayValue{
			Val: res,
		}, nil
	case types.UUIDArrayType:
		res := make([]*types.UUID, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				u, err := types.ParseUUID(*v)
				if err != nil {
					return nil, fmt.Errorf("cannot cast text array to uuid array: %w", err)
				}
				res[i] = u
			}
		}

		return &UuidArrayValue{
			Val: res,
		}, nil
	default:
		return nil, fmt.Errorf("cannot cast text array to %s", t)
	}
}

type BoolArrayValue struct {
	Val []*bool
}

func (a *BoolArrayValue) DBValue() (any, error) {
	return &a.Val, nil
}

func (a *BoolArrayValue) Compare(v Value, op ComparisonOp) (Value, error) {
	return cmpArrs(a, v, op)
}

func (a *BoolArrayValue) Len() int {
	return len(a.Val)
}

func (a *BoolArrayValue) Index(i int64) (ScalarValue, error) {
	if i < 1 || i > int64(len(a.Val)) {
		return nil, fmt.Errorf("index out of bounds")
	}

	val := a.Val[i-1]
	if val == nil {
		return &NullValue{DataType: types.BoolType}, nil
	}

	return &BoolValue{Val: *val}, nil
}

func (a *BoolArrayValue) Set(i int64, v ScalarValue) error {
	if i < 1 {
		return fmt.Errorf("index out of bounds")
	}

	if i > int64(len(a.Val)) {
		// Allocate enough space to set the value.
		// This matches the behavior of Postgres.
		newVal := make([]*bool, i)
		copy(newVal, a.Val)
		a.Val = newVal
	}

	val, ok := v.(*BoolValue)
	if !ok {
		return fmt.Errorf("cannot set non-bool value in bool array")
	}

	a.Val[i-1] = &val.Val
	return nil
}

func (a *BoolArrayValue) Type() *types.DataType {
	return types.BoolArrayType
}

func (a *BoolArrayValue) RawValue() any {
	return a.Val
}

func (a *BoolArrayValue) Size() int {
	return len(a.Val)
}

func (a *BoolArrayValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.TextArrayType:
		res := make([]*string, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				s := strconv.FormatBool(*v)
				res[i] = &s
			}
		}
		return &TextArrayValue{Val: res}, nil
	case types.IntArrayType:
		res := make([]*int64, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				var i int64
				if *v {
					i = 1
				}
				res[i] = &i
			}
		}
		return &IntArrayValue{Val: res}, nil
	default:
		return nil, fmt.Errorf("cannot cast bool array to %s", t)
	}
}

type DecimalArrayValue struct {
	Val []*decimal.Decimal
}

func (a *DecimalArrayValue) DBValue() (any, error) {
	return &a.Val, nil
}

// detectDecArrType detects the type of a decimal array.
// It returns the type, and a boolean indicating if the array does
// not have any non-null values.
func detectDecArrType(arr *DecimalArrayValue) (typ *types.DataType, containsOnlyNulls bool) {
	var firstFound *types.DataType
	for _, v := range arr.Val {
		if v != nil {
			if firstFound == nil {
				d, err := types.NewDecimalType(v.Precision(), v.Scale())
				if err != nil {
					panic(err)
				}

				firstFound = d
			} else {
				d2, err := types.NewDecimalType(v.Precision(), v.Scale())
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

func (a *DecimalArrayValue) Compare(v Value, op ComparisonOp) (Value, error) {
	return cmpArrs(a, v, op)
}

// cmpArrs compares two Kwil array types.
func cmpArrs[M ArrayValue](a M, b Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(b, op); early {
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

			_, v1IsNull := v1.(*NullValue)
			_, v2IsNull := v2.(*NullValue)

			if v1IsNull && v2IsNull {
				continue
			}

			if v1IsNull || v2IsNull {
				return false, nil
			}

			res, err := v1.Compare(v2, equal)
			if err != nil {
				return false, err
			}

			resBool, ok := res.(*BoolValue)
			if !ok {
				return false, fmt.Errorf("unexpected value from comparison")
			}

			if !resBool.Val {
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
		return &BoolValue{Val: eq}, nil
	case isDistinctFrom:
		return &BoolValue{Val: !eq}, nil
	default:
		return nil, fmt.Errorf("only = and IS DISTINCT FROM are supported for array comparison")
	}
}

func (a *DecimalArrayValue) Len() int {
	return len(a.Val)
}

func (a *DecimalArrayValue) Index(i int64) (ScalarValue, error) {
	if i < 1 || i > int64(len(a.Val)) {
		return nil, fmt.Errorf("index out of bounds")
	}

	val := a.Val[i-1]
	if val == nil {
		typ, _ := detectDecArrType(a)
		return &NullValue{DataType: typ}, nil
	}

	return &DecimalValue{Dec: val}, nil
}

func (d *DecimalValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.IntType:
		i, err := d.Dec.Int64()
		if err != nil {
			return nil, err
		}

		return &IntValue{Val: i}, nil
	case types.TextType:
		return &TextValue{Val: d.Dec.String()}, nil
	default:
		return nil, fmt.Errorf("cannot cast decimal to %s", t)
	}
}

func (d *DecimalValue) Size() int {
	return int(d.Dec.Precision())
}

func (a *DecimalArrayValue) Set(i int64, v ScalarValue) error {
	if i < 1 {
		return fmt.Errorf("index out of bounds")
	}

	if i > int64(len(a.Val)) {
		// Allocate enough space to set the value.
		// This matches the behavior of Postgres.
		newVal := make([]*decimal.Decimal, i)
		copy(newVal, a.Val)
		a.Val = newVal
	}

	val, ok := v.(*DecimalValue)
	if !ok {
		return fmt.Errorf("cannot set non-decimal value in decimal array")
	}

	a.Val[i-1] = val.Dec
	return nil
}

func (a *DecimalArrayValue) Type() *types.DataType {
	typ, _ := detectDecArrType(a)
	return typ
}

func (a *DecimalArrayValue) RawValue() any {
	return a.Val
}

func (a *DecimalArrayValue) Size() int {
	size := 0
	for _, v := range a.Val {
		if v != nil {
			size += int(v.Precision())
		}
	}
	return size
}

func (a *DecimalArrayValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.TextArrayType:
		res := make([]*string, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				s := v.String()
				res[i] = &s
			}
		}
		return &TextArrayValue{Val: res}, nil
	case types.IntArrayType:
		res := make([]*int64, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				i, err := v.Int64()
				if err != nil {
					return nil, fmt.Errorf("cannot cast decimal to int: %w", err)
				}
				res[i] = &i
			}
		}
		return &IntArrayValue{Val: res}, nil
	default:
		return nil, fmt.Errorf("cannot cast decimal array to %s", t)
	}
}

type BlobArrayValue struct {
	Val []*[]byte
}

func (a *BlobArrayValue) DBValue() (any, error) {
	return &a.Val, nil
}

func (a *BlobArrayValue) Compare(v Value, op ComparisonOp) (Value, error) {
	return cmpArrs(a, v, op)
}

func (a *BlobArrayValue) Len() int {
	return len(a.Val)
}

func (a *BlobArrayValue) Index(i int64) (ScalarValue, error) {
	if i < 1 || i > int64(len(a.Val)) {
		return nil, fmt.Errorf("index out of bounds")
	}

	val := a.Val[i-1]
	if val == nil {
		return &NullValue{DataType: types.BlobType}, nil
	}

	return &BlobValue{Val: *val}, nil
}

func (a *BlobArrayValue) Set(i int64, v ScalarValue) error {
	if i < 1 {
		return fmt.Errorf("index out of bounds")
	}

	if i > int64(len(a.Val)) {
		// Allocate enough space to set the value.
		// This matches the behavior of Postgres.
		newVal := make([]*[]byte, i)
		copy(newVal, a.Val)
		a.Val = newVal
	}

	val, ok := v.(*BlobValue)
	if !ok {
		return fmt.Errorf("cannot set non-blob value in blob array")
	}

	// copy the blob value to avoid mutation
	valCopy := make([]byte, len(val.Val))
	copy(valCopy, val.Val)

	// subtract 1 because it is 1-indexed
	a.Val[i-1] = &valCopy
	return nil
}

func (a *BlobArrayValue) Type() *types.DataType {
	return types.BlobArrayType
}

func (a *BlobArrayValue) RawValue() any {
	return a.Val
}

func (a *BlobArrayValue) Size() int {
	size := 0
	for _, v := range a.Val {
		if v != nil {
			size += len(*v)
		}
	}
	return size
}

func (a *BlobArrayValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.TextArrayType:
		res := make([]*string, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				s := string(*v)
				res[i] = &s
			}
		}
		return &TextArrayValue{Val: res}, nil
	default:
		return nil, fmt.Errorf("cannot cast blob array to %s", t)
	}
}

type UuidArrayValue struct {
	Val []*types.UUID
}

func (a *UuidArrayValue) DBValue() (any, error) {
	return &a.Val, nil
}

func (a *UuidArrayValue) Compare(v Value, op ComparisonOp) (Value, error) {
	return cmpArrs(a, v, op)
}

func (a *UuidArrayValue) Len() int {
	return len(a.Val)
}

func (a *UuidArrayValue) Index(i int64) (ScalarValue, error) {
	if i < 1 || i > int64(len(a.Val)) {
		return nil, fmt.Errorf("index out of bounds")
	}

	val := a.Val[i-1]
	if val == nil {
		return &NullValue{DataType: types.UUIDType}, nil
	}

	return &UUIDValue{Val: *val}, nil
}

func (a *UuidArrayValue) Set(i int64, v ScalarValue) error {
	if i < 1 {
		return fmt.Errorf("index out of bounds")
	}

	if i > int64(len(a.Val)) {
		// Allocate enough space to set the value.
		// This matches the behavior of Postgres.
		newVal := make([]*types.UUID, i)
		copy(newVal, a.Val)
		a.Val = newVal
	}

	val, ok := v.(*UUIDValue)
	if !ok {
		return fmt.Errorf("cannot set non-uuid value in uuid array")
	}

	a.Val[i-1] = &val.Val
	return nil
}

func (a *UuidArrayValue) Type() *types.DataType {
	return types.UUIDArrayType
}

func (a *UuidArrayValue) RawValue() any {
	return a.Val
}

func (a *UuidArrayValue) Size() int {
	size := 0
	for _, v := range a.Val {
		if v != nil {
			size += 16
		}
	}
	return size
}

func (a *UuidArrayValue) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.TextArrayType:
		res := make([]*string, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				s := v.String()
				res[i] = &s
			}
		}
		return &TextArrayValue{Val: res}, nil
	default:
		return nil, fmt.Errorf("cannot cast uuid array to %s", t)
	}
}

type NullValue struct {
	DataType *types.DataType
}

func (n *NullValue) DBValue() (any, error) {
	return nil, fmt.Errorf("cannot convert null to db value")
}

func (n *NullValue) Compare(v Value, op ComparisonOp) (Value, error) {
	if _, ok := v.(*NullValue); !ok {
		return &NullValue{DataType: n.DataType}, nil
	}

	if op == isDistinctFrom {
		return &BoolValue{Val: false}, nil
	}

	if op == is {
		return &BoolValue{Val: true}, nil
	}

	return &NullValue{DataType: n.DataType}, nil
}

func (n *NullValue) Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error) {
	return &NullValue{DataType: n.DataType}, nil
}

func (n *NullValue) Unary(op UnaryOp) (ScalarValue, error) {
	return &NullValue{DataType: n.DataType}, nil
}

func (n *NullValue) Type() *types.DataType {
	return n.DataType
}

func (n *NullValue) RawValue() any {
	return nil
}

func (n *NullValue) Size() int {
	return 0
}

func (n *NullValue) Array(v ...ScalarValue) (ArrayValue, error) {
	return &NullValue{DataType: n.DataType}, nil
}

func (n *NullValue) Len() int {
	return 0
}

func (n *NullValue) Index(i int64) (ScalarValue, error) {
	return &NullValue{DataType: n.DataType}, nil
}

func (n *NullValue) Set(i int64, v ScalarValue) error {
	return fmt.Errorf("cannot set value in null array")
}

func (n *NullValue) Cast(t *types.DataType) (Value, error) {
	return &NullValue{DataType: t}, nil
}

type RecordValue struct {
	Fields map[string]Value
	Order  []string
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

func (o *RecordValue) DBValue() (any, error) {
	return nil, fmt.Errorf("cannot convert record to db value")
}

func (o *RecordValue) Compare(v Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v, op); early {
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
		return &BoolValue{Val: isSame}, nil
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

func (o *RecordValue) Size() int {
	size := 0
	for _, v := range o.Fields {
		size += v.Size()
	}

	return size
}

func (o *RecordValue) Cast(t *types.DataType) (Value, error) {
	return nil, fmt.Errorf("cannot cast record to %s", t)
}

func cmpIntegers(a, b int, op ComparisonOp) (*BoolValue, error) {
	switch op {
	case equal:
		return &BoolValue{Val: a == b}, nil
	case lessThan:
		return &BoolValue{Val: a < b}, nil
	case greaterThan:
		return &BoolValue{Val: a > b}, nil
	case isDistinctFrom:
		return &BoolValue{Val: a != b}, nil
	default:
		return nil, fmt.Errorf("unknown comparison operator: %d", op)
	}
}
