package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/holiman/uint256"
)

// Uint256 is a 256-bit unsigned integer.
// It is mostly a wrapper around github.com/holiman/uint256.Int, but includes
// extra methods for usage in Postgres.
type Uint256 struct {
	base uint256.Int // not exporting massive method set, which also has params and returns of holiman types
	// Null indicates if this is a NULL value in a SQL table. This approach is
	// typical in most sql.Valuers, which precludes using a nil pointer to
	// indicate a NULL value.
	Null bool
}

// Uint256FromInt creates a new Uint256 from an int.
func Uint256FromInt(i uint64) *Uint256 {
	return &Uint256{base: *uint256.NewInt(i)}
}

// Uint256FromString creates a new Uint256 from a string. A Uint256 representing
// a NULL value should be created with a literal (&Uint256{ Null: true }) or via
// of the unmarshal / scan methods.
func Uint256FromString(s string) (*Uint256, error) {
	i, err := uint256.FromDecimal(s)
	if err != nil {
		return nil, err
	}
	return &Uint256{base: *i}, nil
}

// Uint256FromBig creates a new Uint256 from a big.Int.
func Uint256FromBig(i *big.Int) (*Uint256, error) {
	if i == nil {
		return &Uint256{Null: true}, nil
	}
	return Uint256FromString(i.String())
}

// Uint256FromBytes creates a new Uint256 from a byte slice.
func Uint256FromBytes(b []byte) (*Uint256, error) {
	if b == nil {
		return &Uint256{Null: true}, nil
	} // zero length non-null is for the actual value 0
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

func (u *Uint256) Clone() *Uint256 {
	v := *u
	return &v
}

func (u *Uint256) Add(v *Uint256) *Uint256 {
	z := uint256.NewInt(0)

	return &Uint256{
		base: *z.Add(&u.base, &v.base),
	}
}

func (u *Uint256) Sub(v *Uint256) (*Uint256, error) {
	z := uint256.NewInt(0)
	res, overflow := z.SubOverflow(&u.base, &v.base)
	if overflow {
		return nil, fmt.Errorf("overflow")
	}

	return &Uint256{
		base: *res,
	}, nil
}

func (u *Uint256) Mul(v *Uint256) (*Uint256, error) {
	z := uint256.NewInt(0)

	res, overflow := z.MulOverflow(&u.base, &v.base)
	if overflow {
		return nil, fmt.Errorf("overflow")
	}

	return &Uint256{
		base: *res,
	}, nil
}

func (u *Uint256) Div(v *Uint256) *Uint256 {
	z := uint256.NewInt(0)

	return &Uint256{
		base: *z.Div(&u.base, &v.base),
	}
}

func (u *Uint256) DivMod(v *Uint256) (*Uint256, *Uint256) {
	z := uint256.NewInt(0)
	mod := uint256.NewInt(0)
	z.DivMod(&u.base, &v.base, mod)

	return &Uint256{
			base: *z,
		}, &Uint256{
			base: *mod,
		}
}

func (u *Uint256) Mod(v *Uint256) *Uint256 {
	z := uint256.NewInt(0)

	return &Uint256{
		base: *z.Mod(&u.base, &v.base),
	}
}

func (u *Uint256) Cmp(v *Uint256) int {
	return u.base.Cmp(&v.base)
}

func CmpUint256(u, v *Uint256) int {
	return u.Cmp(v)
}

var _ json.Marshaler = Uint256{}
var _ json.Marshaler = (*Uint256)(nil)

func (u Uint256) MarshalJSON() ([]byte, error) {
	if u.Null {
		return []byte("null"), nil
	}
	return []byte(`"` + u.base.String() + `"`), nil
}

var _ json.Unmarshaler = (*Uint256)(nil)

func (u *Uint256) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	if str == "" { // JSON data was null or ""
		u.Null = true
		u.base.Clear()
		return nil
	}
	u2, err := Uint256FromString(str)
	if err != nil {
		return err
	}

	u.base = u2.base
	return nil
}

var _ encoding.BinaryMarshaler = Uint256{}
var _ encoding.BinaryMarshaler = (*Uint256)(nil)

func (u Uint256) MarshalBinary() ([]byte, error) {
	if u.Null {
		return nil, nil
	}
	return u.base.Bytes(), nil
}

var _ encoding.BinaryUnmarshaler = (*Uint256)(nil)

func (u *Uint256) UnmarshalBinary(data []byte) error {
	if data == nil {
		*u = Uint256{Null: true}
		return nil
	} // len(data) == 0 is the actual value 0
	u.base.SetBytes(data) // u.base, _ = uint256.FromBig(new(big.Int).SetBytes(buf))
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
