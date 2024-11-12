package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

var namespace = uuid.MustParse("cc1cd90f-b4db-47f4-b6df-4bbe5fca88eb")

// UUID is a rfc4122 compliant uuidv5
type UUID [16]byte

// NewUUIDV5 generates a uuidv5 from a byte slice.
// This is used to deterministically generate uuids.
func NewUUIDV5(from []byte) *UUID {
	u := UUID(uuid.NewSHA1(namespace, from))
	return &u
}

// NewUUIDV5WithNamespace generates a uuidv5 from a byte slice and a namespace.
// This is used to deterministically generate uuids.
func NewUUIDV5WithNamespace(namespace UUID, from []byte) UUID {
	u := uuid.NewSHA1(uuid.UUID(namespace), from)
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
	return u.String(), nil
}

func (u *UUID) Bytes() []byte {
	return u[:]
}

var _ json.Marshaler = UUID{}
var _ json.Marshaler = (*UUID)(nil)

// MarshalJSON implements json.Marshaler.
func (u UUID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + u.String() + `"`), nil
}

var _ json.Unmarshaler = (*UUID)(nil)

func (u *UUID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	uu, err := ParseUUID(s)
	if err != nil {
		return err
	}

	copy(u[:], uu[:])
	return nil
}

var _ driver.Valuer = (*UUID)(nil)

func (u *UUID) Scan(src any) error {
	switch s := src.(type) {
	case []byte:
		copy(u[:], s)
		return nil
	case string:
		ui, err := ParseUUID(s)
		if err != nil {
			return err
		}
		copy(u[:], ui[:])
		return nil
	}
	return errors.New("not a byte slice")
}

var _ sql.Scanner = (*UUID)(nil)

// pgx seems to work alright with any slice of Valuers (like a []UUID), but
// explicitly defining the Valuer for a custom type saves some reflection

// UUIDArray is a slice of UUIDs.
// It is used to store arrays of UUIDs in the database.
type UUIDArray []*UUID

func (u UUIDArray) Value() (driver.Value, error) {
	// Postgres does not like []byte for uuid, so we convert to string
	v := make([]string, len(u))
	for i, ui := range u {
		v[i] = ui.String()
	}
	return v, nil
}

var _ driver.Valuer = (*UUIDArray)(nil)

func (u *UUIDArray) Scan(src any) error {
	switch s := src.(type) {
	case [][]byte:
		ux := make(UUIDArray, len(s))
		for i, si := range s {
			var vi UUID
			copy(vi[:], si)
			ux[i] = &vi
		}
		return nil
	case []string:
		ux := make(UUIDArray, len(s))
		for i, si := range s {
			ui, err := ParseUUID(si)
			if err != nil {
				return err
			}
			ux[i] = ui
		}
	}
	return errors.New("not a byte slice")
}

func (u UUIDArray) Bytes() [][]byte {
	v := make([][]byte, len(u))
	for i, ui := range u {
		v[i] = ui.Bytes()
	}
	return v
}
