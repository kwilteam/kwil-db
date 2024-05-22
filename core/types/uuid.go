package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	return u[:], nil // []byte works for sql
}

func (u UUID) Bytes() []byte {
	return u[:]
}

// Over json, we want to send uuids as strings
func (u UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

func (u UUID) UnmarshalJSON(b []byte) error {
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

var _ driver.Valuer = UUID{}
var _ driver.Valuer = (*UUID)(nil)
var _ pgtype.Codec = &UUID{}

func (u *UUID) Scan(src any) error {
	switch s := src.(type) {
	case []byte:
		copy(u[:], s)
		return nil
	}
	return errors.New("not a byte slice")
}

func (u *UUID) FormatSupported(format int16) bool {
	return format == pgtype.TextFormatCode || format == pgtype.BinaryFormatCode
}

func (u *UUID) PreferredFormat() int16 {
	return pgtype.BinaryFormatCode
}

func (u *UUID) PlanEncode(m *pgtype.Map, oid uint32, format int16, value any) pgtype.EncodePlan {
	var val *UUID
	switch t := value.(type) {
	case UUID:
		val = &t
	case *UUID:
		val = t
	case []byte:
		if len(t) != 16 && len(t) != 0 {
			return nil
		}

		uuid := UUID{}
		copy(uuid[:], t)
		val = &uuid
	case [16]byte:
		uuid := UUID{}
		copy(uuid[:], t[:])
		val = &uuid
	default:
		return nil
	}

	switch format {
	// given our uuid type, I believe it will always come binary
	case pgtype.BinaryFormatCode:
		return encodePlanFunc(func(value any, buf []byte) (newBuf []byte, err error) {
			return append(buf, val[:]...), nil
		})
	case pgtype.TextFormatCode:
		return encodePlanFunc(func(value any, buf []byte) (newBuf []byte, err error) {
			return append(buf, val.String()...), nil
		})
	}

	return nil
}

type encodePlanFunc func(value any, buf []byte) (newBuf []byte, err error)

func (e encodePlanFunc) Encode(value any, buf []byte) (newBuf []byte, err error) {
	return e(value, buf)
}

func (u *UUID) PlanScan(m *pgtype.Map, oid uint32, format int16, target any) pgtype.ScanPlan {
	switch format {
	case pgtype.BinaryFormatCode:
		return scanPlanFunc(func(src []byte, target any) error {
			if target == nil {
				return nil
			}

			if len(src) == 0 {
				return nil
			}

			uuid, ok := target.(*UUID)
			if !ok {
				if len(src) != 16 {
					return fmt.Errorf("expected 16 bytes, got %d", len(src))
				}

				uuid := UUID{}
				copy(uuid[:], src)
				reflect.ValueOf(target).Elem().Set(reflect.ValueOf(uuid))

				return nil
			}

			copy(uuid[:], src)
			return nil
		})
	case pgtype.TextFormatCode:
		return scanPlanFunc(func(src []byte, target any) error {
			if target == nil {
				return nil
			}

			if len(src) == 0 {
				return nil
			}

			u, err := ParseUUID(string(src))
			if err != nil {
				return err
			}

			uuid, ok := target.(*UUID)
			if !ok {
				reflect.ValueOf(target).Elem().Set(reflect.ValueOf(uuid))
				return nil
			}
			copy(uuid[:], u[:])
			return nil
		})
	}

	return nil
}

type scanPlanFunc func(src []byte, target any) error

func (s scanPlanFunc) Scan(src []byte, target any) error {
	return s(src, target)
}

func (u *UUID) DecodeDatabaseSQLValue(m *pgtype.Map, oid uint32, format int16, src []byte) (driver.Value, error) {
	if src == nil {
		return nil, nil
	}

	uuid := UUID{}
	copy(uuid[:], src)

	return uuid, nil
}

func (u *UUID) DecodeValue(m *pgtype.Map, oid uint32, format int16, src []byte) (any, error) {
	if src == nil {
		return &UUID{}, nil
	}

	uuid := UUID{}
	copy(uuid[:], src)

	return &uuid, nil
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

func (u *UUIDArray) UnmarshalJSON(b []byte) error {
	var s []string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	ux := make(UUIDArray, len(s))
	for i, si := range s {
		uui, err := ParseUUID(si)
		if err != nil {
			return err
		}
		copy(ux[i][:], uui[:])
	}

	*u = ux
	return nil
}

func (u UUIDArray) MarshalJSON() ([]byte, error) {
	s := make([]string, len(u))
	for i, ui := range u {
		s[i] = ui.String()
	}
	return json.Marshal(s)
}

var _ sql.Scanner = (*UUIDArray)(nil)
