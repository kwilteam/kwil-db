package types

import "github.com/google/uuid"

var namespace = uuid.MustParse("cc1cd90f-b4db-47f4-b6df-4bbe5fca88eb")

// UUID is a rfc4122 compliant uuidv5
type UUID [16]byte

// NewUUIDV5 generates a uuidv5 from a byte slice.
// This is used to deterministically generate uuids.
func NewUUIDV5(from []byte) UUID {
	u := uuid.NewSHA1(namespace, from)
	return UUID(u)
}

// String returns the string representation of the uuid
func (u UUID) String() string {
	return uuid.UUID(u).String()
}
