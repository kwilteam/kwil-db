package types

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math/big"

	"github.com/holiman/uint256"
)

// Uint256 is a 256-bit unsigned integer.
// It is mostly a wrapper around github.com/holiman/uint256.Int, but includes
// extra methods for usage in Postgres.
type Uint256 struct {
	base uint256.Int // not exporting massive method set, which also has params and returns of holiman types
	Null bool
}

// Uint256FromInt creates a new Uint256 from an int.
func Uint256FromInt(i uint64) *Uint256 {
	return &Uint256{base: *uint256.NewInt(i)}
}

// Uint256FromString creates a new Uint256 from a string.
func Uint256FromString(s string) (*Uint256, error) {
	i, err := uint256.FromDecimal(s)
	if err != nil {
		return nil, err
	}
	return &Uint256{base: *i}, nil
}

// Uint256FromBig creates a new Uint256 from a big.Int.
func Uint256FromBig(i *big.Int) (*Uint256, error) {
	return Uint256FromString(i.String())
}

// Uint256FromBytes creates a new Uint256 from a byte slice.
func Uint256FromBytes(b []byte) (*Uint256, error) {
	bigInt := new(big.Int).SetBytes(b)
	return Uint256FromBig(bigInt)
}

func (u Uint256) String() string {
	return u.base.String()
}

func (u Uint256) Bytes() []byte {
	return u.base.Bytes()
}

func (u Uint256) ToBig() *big.Int {
	return u.base.ToBig()
}

func (u Uint256) MarshalJSON() ([]byte, error) {
	return []byte(u.base.String()), nil // ? json ?
}

func (u *Uint256) Clone() *Uint256 {
	v := *u
	return &v
}

func (u *Uint256) Cmp(v *Uint256) int {
	return u.base.Cmp(&v.base)
}

func CmpUint256(u, v *Uint256) int {
	return u.Cmp(v)
}

func (u *Uint256) UnmarshalJSON(b []byte) error {
	u2, err := Uint256FromString(string(b))
	if err != nil {
		return err
	}

	u.base = u2.base
	return nil
}

// Value implements the driver.Valuer interface.
func (u Uint256) Value() (driver.Value, error) {
	if u.Null {
		return nil, nil
	}
	return u.String(), nil
}

var _ driver.Valuer = Uint256{}
var _ driver.Valuer = (*Uint256)(nil)

// Scan implements the sql.Scanner interface.
func (u *Uint256) Scan(src interface{}) error {
	switch s := src.(type) {
	case string:
		u2, err := Uint256FromString(s)
		if err != nil {
			return err
		}

		u.base = u2.base
		u.Null = false
		return nil

	case nil:
		u.Null = true
		u.base.Clear()
		return nil
	}

	return fmt.Errorf("cannot convert %T to Uint256", src)
}

var _ sql.Scanner = (*Uint256)(nil)

// Uint256Array is an array of Uint256s.
type Uint256Array []*Uint256

// Value implements the driver.Valuer interface.
func (ua Uint256Array) Value() (driver.Value, error) {
	// Even when implementing pgtype.ArrayGetter we still need this, so that the
	// pgx driver can use it's wrapSliceEncodePlan.
	strs := make([]string, len(ua))
	for i, u := range ua {
		strs[i] = u.String()
	}

	return strs, nil
}

var _ driver.Valuer = (*Uint256Array)(nil)

// Uint256Array is a slice of Scanners. pgx at least is smart enough to make
// this work automatically!
// Another approach is to implement pgx.ArraySetter and pgx.ArrayGetter like
// similar in effect to:
//   type Uint256Array pgtype.FlatArray[*Uint256]
