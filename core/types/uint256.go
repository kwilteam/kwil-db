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
var _ driver.Valuer = (*Uint256)(nil)

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

// Well, as long as Uint256Array is just an underlying []*Uint256, pgx works fine
// without implementing these interfaces. Also, these methods have pgtype types on
// them, so I'm commenting it for now, may delete.
/*
var _ pgtype.ArrayGetter = Uint256Array{}
var _ pgtype.ArrayGetter = (*Uint256Array)(nil)

func (ua Uint256Array) Dimensions() []pgtype.ArrayDimension {
	if ua == nil {
		return nil
	}
	return []pgtype.ArrayDimension{{Length: int32(len(ua)), LowerBound: 1}}
}

func (ua Uint256Array) Index(i int) any { return ua[i] }

func (ua Uint256Array) IndexType() any {
	return (*Uint256)(nil) // &Uint256{}
}

func pgDimCard(dimensions []pgtype.ArrayDimension) int {
	if len(dimensions) == 0 {
		return 0
	}
	elementCount := int(dimensions[0].Length)
	for _, d := range dimensions[1:] {
		elementCount *= int(d.Length)
	}
	return elementCount
}

var _ pgtype.ArraySetter = (*Uint256Array)(nil)

func (ua *Uint256Array) SetDimensions(dimensions []pgtype.ArrayDimension) error {
	if dimensions == nil {
		*ua = nil
		return nil
	}
	*ua = make(Uint256Array, pgDimCard(dimensions))
	return nil
}

func (ua Uint256Array) ScanIndex(i int) any { return &ua[i] }

func (ua Uint256Array) ScanIndexType() any { return new(Uint256) }
*/
