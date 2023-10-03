package types

import (
	"github.com/kwilteam/kwil-db/pkg/serialize"
)

// // UserIdentifier is an interface for identifying a user by public key
// type UserIdentifier interface {
// 	MarshalBinary() ([]byte, error)
// 	PubKey() (crypto.PublicKey, error)
// 	UnmarshalBinary(data []byte) error
// 	Address() (string, error)
// }

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
	u2 := &User{}
	err := serialize.DecodeInto(data, u2)
	if err != nil {
		return err
	}

	u.PublicKey = u2.PublicKey
	u.AuthType = u2.AuthType

	return nil
}
