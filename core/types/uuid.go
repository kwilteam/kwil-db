package types

import (
	"database/sql"
	"database/sql/driver"
	"errors"

	"github.com/google/uuid"
)

var namespace = uuid.MustParse("cc1cd90f-b4db-47f4-b6df-4bbe5fca88eb")

// UUID is a rfc4122 compliant uuidv5
type UUID [16]byte

// NewUUIDV5 generates a uuidv5 from a byte slice.
// This is used to deterministically generate uuids.
func NewUUIDV5(from []byte) UUID {
	u := uuid.NewSHA1(namespace, from)
	return UUID(u)
}

// ParseUUID parses a uuid from a string
func ParseUUID(s string) (*UUID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return &UUID{}, err
	}
	u2 := UUID(u)
	return &u2, nil
}

// String returns the string representation of the uuid
func (u UUID) String() string {
	return uuid.UUID(u).String()
}

func (u UUID) Value() (driver.Value, error) {
	return u[:], nil // []byte works for sql
}

func (u UUID) Bytes() []byte {
	return u[:]
}

var _ driver.Valuer = UUID{}
var _ driver.Valuer = (*UUID)(nil)

func (u *UUID) Scan(src any) error {
	switch s := src.(type) {
	case []byte:
		copy(u[:], s)
		return nil
	}
	return errors.New("not a byte slice")
}

var _ sql.Scanner = (*UUID)(nil)

// pgx seems to work alright with any slice of Valuers (like a []UUID), but
// explicitly defining the Valuer for a custom type saves some reflection

type UUIDArray []UUID

func (u UUIDArray) Value() (driver.Value, error) {
	v := make([][]byte, len(u))
	for i, ui := range u {
		vi := make([]byte, 16)
		copy(vi, ui[:])
		v[i] = vi
	}
	return v, nil
}

var _ driver.Valuer = UUIDArray{}
var _ driver.Valuer = (*UUIDArray)(nil)

func (u *UUIDArray) Scan(src any) error {
	switch s := src.(type) {
	case [][]byte:
		ux := make(UUIDArray, len(s))
		for i, si := range s {
			var vi UUID
			copy(vi[:], si)
			ux[i] = vi
		}
		return nil
	}
	return errors.New("not a byte slice slice")
}

var _ sql.Scanner = (*UUIDArray)(nil)
