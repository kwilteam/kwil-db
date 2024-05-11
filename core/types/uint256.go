package types

import (
	"database/sql/driver"
	"fmt"
	"math/big"

	"github.com/holiman/uint256"
)

// Uint256 is a 256-bit unsigned integer.
// It is mostly a wrapper around github.com/holiman/uint256.Int, but includes
// extra methods for usage in Postgres.
type Uint256 struct {
	uint256.Int
}

// Uint256FromInt creates a new Uint256 from an int.
func Uint256FromInt(i uint64) *Uint256 {
	return &Uint256{Int: *uint256.NewInt(i)}
}

// Uint256FromString creates a new Uint256 from a string.
func Uint256FromString(s string) (*Uint256, error) {
	i, err := uint256.FromDecimal(s)
	if err != nil {
		return nil, err
	}
	return &Uint256{Int: *i}, nil
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

func (u Uint256) MarshalJSON() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *Uint256) UnmarshalJSON(b []byte) error {
	u2, err := Uint256FromString(string(b))
	if err != nil {
		return err
	}

	u.Int = u2.Int
	return nil
}

// Value implements the driver.Valuer interface.
func (u Uint256) Value() (driver.Value, error) {
	return u.String(), nil
}

var _ driver.Valuer = Uint256{}

// Scan implements the sql.Scanner interface.
func (u *Uint256) Scan(src interface{}) error {
	switch s := src.(type) {
	case string:
		u2, err := Uint256FromString(s)
		if err != nil {
			return err
		}

		u.Int = u2.Int
		return nil
	}

	return fmt.Errorf("cannot convert %T to Uint256", src)
}

var _ driver.Valuer = (*Uint256)(nil)
var _ driver.Valuer = (*Uint256)(nil)

// Uint256Array is an array of Uint256s.
type Uint256Array []*Uint256

// Value implements the driver.Valuer interface.
func (ua Uint256Array) Value() (driver.Value, error) {
	strs := make([]string, len(ua))
	for i, u := range ua {
		strs[i] = u.String()
	}

	return strs, nil
}

var _ driver.Valuer = (*Uint256Array)(nil)

// Scan implements the sql.Scanner interface.
func (ua *Uint256Array) Scan(src interface{}) error {
	switch s := src.(type) {
	case []string:
		*ua = make(Uint256Array, len(s))
		for i, str := range s {
			u, err := Uint256FromString(str)
			if err != nil {
				return err
			}

			(*ua)[i] = u
		}
		return nil
	}

	return fmt.Errorf("cannot convert %T to Uint256Array", src)
}
