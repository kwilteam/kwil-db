package sql

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"
)

// int64Valuer is for internal use so a pgtype.Numeric can be recognized by our
// Int64 helper below.
type int64Valuer interface {
	Int64Value() (pgtype.Int8, error)
}

// int64er is for internal use so our Numeric type can be recognized by the
// Int64 function in addition to a pgtype.Numeric.
type int64erer interface {
	Int64() (int64, error)
}
type int64er interface {
	Int64() int64
}

func Int64(val interface{}) (int64, bool) {
	switch v := val.(type) {
	case int64Valuer:
		iv, err := v.Int64Value()
		if err != nil {
			return 0, false
		}
		return iv.Int64, true
	case int64erer:
		iv, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return iv, true
	case int64er:
		return v.Int64(), true

	case int64:
		return v, true
	case int32:
		return int64(v), true
	case int16:
		return int64(v), true
	case int8:
		return int64(v), true
	case int:
		return int64(v), true

	// unsigned is not gonna happen from sql Scan, but for completeness...
	case uint64:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint:
		return int64(v), true
	}

	return 0, false
}

func ToNumeric(v any) (Numeric, bool) {
	if vn, ok := v.(pgtype.Numeric); ok {
		return Numeric{num: vn}, true
	}
	return Numeric{}, false
}

// TODO: register our Numeric with pgx's TypeMap so it scans into it (embedding
// pgtype.Numeric) instead of a pgtype.Numeric.

// Numeric provides access to the `numeric` values. This type should implement
// the Int64er, BigInter, and Float64er interfaces.
type Numeric struct {
	num pgtype.Numeric
}

func (n *Numeric) Int64() (int64, error) {
	// It could represent a float64, so we use Int64Value instead of just
	// checking if n.num.Valid && n.num.Int != nil.
	pgInt8, err := n.num.Int64Value()
	if err != nil {
		return 0, err
	}
	// pgInt8.Valid check would be redundant with Int64Value error
	return pgInt8.Int64, nil
}

func (n *Numeric) Float64() (float64, error) {
	pgFloat8, err := n.num.Float64Value()
	if err != nil {
		return 0, err
	}
	return pgFloat8.Float64, nil
}

var big0 *big.Int = big.NewInt(0)
var big10 *big.Int = big.NewInt(10)

func (n *Numeric) BigInt() (*big.Int, error) {
	if !n.num.Valid || n.num.Int == nil {
		return nil, errors.New("invalid numeric")
	}

	if n.num.Exp == 0 { // not float
		return n.num.Int, nil
	}

	// The rest of this function is the logic used by Int64Value => toBigInt
	// (unexported) to get an int for a convertible float.
	num := &big.Int{}
	num.Set(n.num.Int)
	if n.num.Exp > 0 {
		mul := &big.Int{}
		mul.Exp(big10, big.NewInt(int64(n.num.Exp)), nil)
		num.Mul(num, mul)
		return num, nil
	}

	div := &big.Int{}
	div.Exp(big10, big.NewInt(int64(-n.num.Exp)), nil)
	remainder := &big.Int{}
	num.DivMod(num, div, remainder)
	if remainder.Cmp(big0) != 0 {
		return nil, fmt.Errorf("cannot convert %v to integer", n)
	}
	return num, nil
}
