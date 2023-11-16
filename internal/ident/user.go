package ident

import "github.com/kwilteam/kwil-db/core/types/serialize"

// User is an end user of the engine, identified by a public key.
// It includes an authentication type, which is used to determine how to
// authenticate the and how to generate an address for the user.
type User struct {
	// PublicKey is the public key of the user
	PublicKey []byte

	// AuthType is the type of authentication used by the user
	AuthType string
}

func (u *User) MarshalBinary() ([]byte, error) {
	return serialize.Encode(u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	return serialize.DecodeInto(data, u)
}
