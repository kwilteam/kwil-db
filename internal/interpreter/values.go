package interpreter

import (
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
)

// Value is a value that can be compared, used in arithmetic operations,
// and have unary operations applied to it.
type Value interface {
	// Compare compares the variable with another variable using the given comparison operator.
	// It will return a boolean value, or null either of the variables is null.
	Compare(v Value, op ComparisonOp) (Value, error)
	// Type returns the type of the variable.
	Type() *types.DataType
	// Value returns the value of the variable.
	Value() any
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

// NewVariable creates a new variable from the given value.
func NewVariable(v any) (Value, error) {
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
	case *types.Uint256:
		return &Uint256Value{Val: v}, nil
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
		return &BlobArrayValue{
			Val: v,
		}, nil
	case []*decimal.Decimal:
		if len(v) == 0 {
			return nil, fmt.Errorf("cannot infer type from decimal empty array")
		}

		dt2, err := types.NewDecimalType(v[0].Precision(), v[0].Scale())
		if err != nil {
			return nil, err
		}

		return &DecimalArrayValue{
			Val:      v,
			DataType: dt2,
		}, nil
	case []*types.Uint256:
		return &Uint256ArrayValue{
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

func (i *IntValue) Value() any {
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
	case types.Uint256Type:
		if i.Val < 0 {
			return nil, fmt.Errorf("cannot cast negative int to uint256")
		}

		return &Uint256Value{
			Val: types.Uint256FromInt(uint64(i.Val)),
		}, nil
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

func (s *TextValue) Value() any {
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
	case types.Uint256Type:
		u, err := types.Uint256FromString(s.Val)
		if err != nil {
			return nil, err
		}

		return &Uint256Value{Val: u}, nil
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

func (b *BoolValue) Value() any {
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

func (b *BlobValue) Value() any {
	return b.Val
}

func (b *BlobValue) Array(v ...ScalarValue) (ArrayValue, error) {
	res := make([][]byte, len(v)+1)
	res[0] = b.Val
	for i, val := range v {
		switch val := val.(type) {
		case *NullValue:
			res[i+1] = nil
		case *BlobValue:
			res[i+1] = val.Val
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

func (u *UUIDValue) Value() any {
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

func (d *DecimalValue) Value() any {
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
		Val:      res,
		DataType: d.Type(),
	}, nil
}

type Uint256Value struct {
	Val *types.Uint256
}

func (u *Uint256Value) Compare(v Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v, op); early {
		return res, nil
	}

	val2, ok := v.(*Uint256Value)
	if !ok {
		return nil, makeTypeErr(u, v)
	}

	c := u.Val.Cmp(val2.Val)

	return cmpIntegers(c, 0, op)
}

func (u *Uint256Value) Arithmetic(v ScalarValue, op ArithmeticOp) (ScalarValue, error) {
	if _, ok := v.(*NullValue); ok {
		return &NullValue{DataType: types.Uint256Type}, nil
	}

	val2, ok := v.(*Uint256Value)
	if !ok {
		return nil, makeTypeErr(u, v)
	}

	switch op {
	case add:
		res := u.Val.Add(val2.Val)
		return &Uint256Value{Val: res}, nil
	case sub:
		res, err := u.Val.Sub(val2.Val)
		return &Uint256Value{Val: res}, err
	case mul:
		res, err := u.Val.Mul(val2.Val)
		return &Uint256Value{Val: res}, err
	case div:
		if val2.Val.Cmp(types.Uint256FromInt(0)) == 0 {
			return nil, fmt.Errorf("cannot divide by zero")
		}
		res := u.Val.Div(val2.Val)
		return &Uint256Value{Val: res}, nil
	case mod:
		if val2.Val.Cmp(types.Uint256FromInt(0)) == 0 {
			return nil, fmt.Errorf("cannot divide by zero")
		}
		res := u.Val.Mod(val2.Val)
		return &Uint256Value{Val: res}, nil
	default:
		return nil, fmt.Errorf("cannot perform arithmetic operation id %d on type uint256", op)
	}
}

func (u *Uint256Value) Unary(op UnaryOp) (ScalarValue, error) {
	switch op {
	case neg:
		return nil, fmt.Errorf("cannot apply unary negation to a uint256")
	case not:
		return nil, fmt.Errorf("cannot apply logical NOT to a uint256")
	case pos:
		return u, nil
	default:
		return nil, fmt.Errorf("unknown unary operator: %d", op)
	}
}

func (u *Uint256Value) Type() *types.DataType {
	return types.Uint256Type
}

func (u *Uint256Value) Value() any {
	return u.Val
}

func (u *Uint256Value) Array(v ...ScalarValue) (ArrayValue, error) {
	res := make([]*types.Uint256, len(v)+1)
	res[0] = u.Val
	for i, val := range v {
		switch val := val.(type) {
		case *NullValue:
			res[i+1] = nil
		case *Uint256Value:
			res[i+1] = val.Val
		default:
			return nil, makeTypeErr(u, val)
		}
	}

	return &Uint256ArrayValue{
		Val: res,
	}, nil
}

func (u *Uint256Value) Cast(t *types.DataType) (Value, error) {
	switch t {
	case types.IntType:
		b := u.Val.ToBig()
		if b.Cmp(big.NewInt(math.MaxInt64)) > 0 {
			return nil, fmt.Errorf("cannot cast uint256 to int: value too large")
		}

		return &IntValue{Val: b.Int64()}, nil
	case types.TextType:
		return &TextValue{Val: u.Val.String()}, nil
	case types.Uint256Type:
		return u, nil
	default:
		return nil, fmt.Errorf("cannot cast uint256 to %s", t)
	}
}

func (u *Uint256Value) Size() int {
	return 32
}

type IntArrayValue struct {
	Val []*int64
}

func (a *IntArrayValue) Compare(v Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v, op); early {
		return res, nil
	}

	val2, ok := v.(*IntArrayValue)
	if !ok {
		return nil, makeTypeErr(a, v)
	}

	if len(a.Val) != len(val2.Val) {
		return nil, fmt.Errorf("cannot compare arrays of different lengths")
	}

	for i, v1 := range a.Val {
		if v1 == nil {
			return &NullValue{DataType: types.IntType}, nil
		}
		v2 := val2.Val[i]
		if v2 == nil {
			return &NullValue{DataType: types.IntType}, nil
		}

		var b bool
		switch op {
		case equal:
			b = *v1 == *v2
		case lessThan:
			b = *v1 < *v2
		case greaterThan:
			b = *v1 > *v2
		case isDistinctFrom:
			b = *v1 != *v2
		default:
			return nil, fmt.Errorf("unknown comparison operator: %d", op)
		}

		if !b {
			return &BoolValue{Val: false}, nil
		}
	}

	return &BoolValue{Val: true}, nil
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

func (a *IntArrayValue) Value() any {
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
	case types.Uint256ArrayType:
		res := make([]*types.Uint256, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				res[i] = types.Uint256FromInt(uint64(*v))
			}
		}

		return &Uint256ArrayValue{
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
			Val:      res,
			DataType: types.DecimalType,
		}, nil
	default:
		return nil, fmt.Errorf("cannot cast int array to %s", t)
	}
}

type TextArrayValue struct {
	Val []*string
}

func (a *TextArrayValue) Compare(v Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v, op); early {
		return res, nil
	}

	val2, ok := v.(*TextArrayValue)
	if !ok {
		return nil, makeTypeErr(a, v)
	}

	if len(a.Val) != len(val2.Val) {
		return nil, fmt.Errorf("cannot compare arrays of different lengths")
	}

	for i, v1 := range a.Val {
		if v1 == nil {
			return &NullValue{DataType: types.TextType}, nil
		}
		v2 := val2.Val[i]
		if v2 == nil {
			return &NullValue{DataType: types.TextType}, nil
		}

		var b bool
		switch op {
		case equal:
			b = *v1 == *v2
		case lessThan:
			b = *v1 < *v2
		case greaterThan:
			b = *v1 > *v2
		case isDistinctFrom:
			b = *v1 != *v2
		default:
			return nil, fmt.Errorf("unknown comparison operator: %d", op)
		}

		if !b {
			return &BoolValue{Val: false}, nil
		}
	}

	return &BoolValue{Val: true}, nil
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

func (a *TextArrayValue) Value() any {
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
	case types.Uint256ArrayType:
		res := make([]*types.Uint256, len(a.Val))
		for i, v := range a.Val {
			if v == nil {
				res[i] = nil
			} else {
				u, err := types.Uint256FromString(*v)
				if err != nil {
					return nil, fmt.Errorf("cannot cast text array to uint256 array: %w", err)
				}
				res[i] = u
			}
		}

		return &Uint256ArrayValue{
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
			Val:      res,
			DataType: types.DecimalType,
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

func (a *BoolArrayValue) Compare(v Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v, op); early {
		return res, nil
	}

	val2, ok := v.(*BoolArrayValue)
	if !ok {
		return nil, makeTypeErr(a, v)
	}

	if len(a.Val) != len(val2.Val) {
		return nil, fmt.Errorf("cannot compare arrays of different lengths")
	}

	for i, v1 := range a.Val {
		if v1 == nil {
			return &NullValue{DataType: types.BoolType}, nil
		}
		v2 := val2.Val[i]
		if v2 == nil {
			return &NullValue{DataType: types.BoolType}, nil
		}

		var b bool
		switch op {
		case equal:
			b = *v1 == *v2
		case isDistinctFrom:
			b = *v1 != *v2
		default:
			return nil, fmt.Errorf("unknown comparison operator: %d", op)
		}

		if !b {
			return &BoolValue{Val: false}, nil
		}
	}

	return &BoolValue{Val: true}, nil
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

func (a *BoolArrayValue) Value() any {
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
	Val      []*decimal.Decimal
	DataType *types.DataType
}

func (a *DecimalArrayValue) Compare(v Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v, op); early {
		return res, nil
	}

	val2, ok := v.(*DecimalArrayValue)
	if !ok {
		return nil, makeTypeErr(a, v)
	}

	if len(a.Val) != len(val2.Val) {
		return nil, fmt.Errorf("cannot compare arrays of different lengths")
	}

	for i, v1 := range a.Val {
		if v1 == nil {
			return &NullValue{DataType: a.DataType}, nil
		}
		v2 := val2.Val[i]
		if v2 == nil {
			return &NullValue{DataType: a.DataType}, nil
		}

		res, err := v1.Cmp(v2)
		if err != nil {
			return nil, err
		}

		b, err := cmpIntegers(res, 0, op)
		if err != nil {
			return nil, err
		}

		if !b.(*BoolValue).Val {
			return &BoolValue{Val: false}, nil
		}
	}

	return &BoolValue{Val: true}, nil
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
		return &NullValue{DataType: a.DataType}, nil
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
	case types.Uint256Type:
		if d.Dec.IsNegative() {
			return nil, fmt.Errorf("cannot cast negative decimal to uint256")
		}

		d2 := d.Dec.Copy()

		err := d2.Round(0)
		if err != nil {
			return nil, err
		}

		u, err := types.Uint256FromString(d2.String())
		if err != nil {
			return nil, err
		}

		return &Uint256Value{
			Val: u,
		}, nil
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
	return a.DataType
}

func (a *DecimalArrayValue) Value() any {
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

type Uint256ArrayValue struct {
	Val []*types.Uint256
}

func (a *Uint256ArrayValue) Compare(v Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v, op); early {
		return res, nil
	}

	val2, ok := v.(*Uint256ArrayValue)
	if !ok {
		return nil, makeTypeErr(a, v)
	}

	if len(a.Val) != len(val2.Val) {
		return nil, fmt.Errorf("cannot compare arrays of different lengths")
	}

	for i, v1 := range a.Val {
		if v1 == nil {
			return &NullValue{DataType: types.Uint256Type}, nil
		}
		v2 := val2.Val[i]
		if v2 == nil {
			return &NullValue{DataType: types.Uint256Type}, nil
		}

		var b bool
		switch op {
		case equal:
			b = v1.Cmp(v2) == 0
		case lessThan:
			b = v1.Cmp(v2) < 0
		case greaterThan:
			b = v1.Cmp(v2) > 0
		case isDistinctFrom:
			b = v1.Cmp(v2) != 0
		default:
			return nil, fmt.Errorf("unknown comparison operator: %d", op)
		}

		if !b {
			return &BoolValue{Val: false}, nil
		}
	}

	return &BoolValue{Val: true}, nil
}

func (a *Uint256ArrayValue) Len() int {
	return len(a.Val)
}

func (a *Uint256ArrayValue) Index(i int64) (ScalarValue, error) {
	if i < 1 || i > int64(len(a.Val)) {
		return nil, fmt.Errorf("index out of bounds")
	}

	val := a.Val[i-1]
	if val == nil {
		return &NullValue{DataType: types.Uint256Type}, nil
	}

	return &Uint256Value{Val: val}, nil
}

func (a *Uint256ArrayValue) Set(i int64, v ScalarValue) error {
	if i < 1 {
		return fmt.Errorf("index out of bounds")
	}

	if i > int64(len(a.Val)) {
		// Allocate enough space to set the value.
		// This matches the behavior of Postgres.
		newVal := make([]*types.Uint256, i)
		copy(newVal, a.Val)
		a.Val = newVal
	}

	val, ok := v.(*Uint256Value)
	if !ok {
		return fmt.Errorf("cannot set non-uint256 value in uint256 array")
	}

	a.Val[i-1] = val.Val
	return nil
}

func (a *Uint256ArrayValue) Type() *types.DataType {
	return types.Uint256ArrayType
}

func (a *Uint256ArrayValue) Value() any {
	return a.Val
}

func (a *Uint256ArrayValue) Size() int {
	size := 0
	for _, v := range a.Val {
		if v != nil {
			size += 32
		}
	}
	return size
}

func (a *Uint256ArrayValue) Cast(t *types.DataType) (Value, error) {
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
		return nil, fmt.Errorf("cannot cast uint256 array to %s", t)
	}
}

type BlobArrayValue struct {
	Val [][]byte
}

func (a *BlobArrayValue) Compare(v Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v, op); early {
		return res, nil
	}

	val2, ok := v.(*BlobArrayValue)
	if !ok {
		return nil, makeTypeErr(a, v)
	}

	if len(a.Val) != len(val2.Val) {
		return nil, fmt.Errorf("cannot compare arrays of different lengths")
	}

	for i, v1 := range a.Val {
		if v1 == nil {
			return &NullValue{DataType: types.BlobType}, nil
		}
		v2 := val2.Val[i]
		if v2 == nil {
			return &NullValue{DataType: types.BlobType}, nil
		}

		var b bool
		switch op {
		case equal:
			b = string(v1) == string(v2)
		case isDistinctFrom:
			b = string(v1) != string(v2)
		default:
			return nil, fmt.Errorf("unknown comparison operator: %d", op)
		}

		if !b {
			return &BoolValue{Val: false}, nil
		}
	}

	return &BoolValue{Val: true}, nil
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

	return &BlobValue{Val: val}, nil
}

func (a *BlobArrayValue) Set(i int64, v ScalarValue) error {
	if i < 1 {
		return fmt.Errorf("index out of bounds")
	}

	if i > int64(len(a.Val)) {
		// Allocate enough space to set the value.
		// This matches the behavior of Postgres.
		newVal := make([][]byte, i)
		copy(newVal, a.Val)
		a.Val = newVal
	}

	val, ok := v.(*BlobValue)
	if !ok {
		return fmt.Errorf("cannot set non-blob value in blob array")
	}

	a.Val[i-1] = val.Val
	return nil
}

func (a *BlobArrayValue) Type() *types.DataType {
	return types.BlobArrayType
}

func (a *BlobArrayValue) Value() any {
	return a.Val
}

func (a *BlobArrayValue) Size() int {
	size := 0
	for _, v := range a.Val {
		if v != nil {
			size += len(v)
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
				s := string(v)
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

func (a *UuidArrayValue) Compare(v Value, op ComparisonOp) (Value, error) {
	if res, early := nullCmp(v, op); early {
		return res, nil
	}

	val2, ok := v.(*UuidArrayValue)
	if !ok {
		return nil, makeTypeErr(a, v)
	}

	if len(a.Val) != len(val2.Val) {
		return nil, fmt.Errorf("cannot compare arrays of different lengths")
	}

	for i, v1 := range a.Val {
		if v1 == nil {
			return &NullValue{DataType: types.UUIDType}, nil
		}
		v2 := val2.Val[i]
		if v2 == nil {
			return &NullValue{DataType: types.UUIDType}, nil
		}

		var b bool
		switch op {
		case equal:
			b = *v1 == *v2
		case isDistinctFrom:
			b = *v1 != *v2
		default:
			return nil, fmt.Errorf("unknown comparison operator: %d", op)
		}

		if !b {
			return &BoolValue{Val: false}, nil
		}
	}

	return &BoolValue{Val: true}, nil
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

func (a *UuidArrayValue) Value() any {
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

func (n *NullValue) Value() any {
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

			if !eq.Value().(bool) {
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

func (o *RecordValue) Value() any {
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

func cmpIntegers(a, b int, op ComparisonOp) (Value, error) {
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
