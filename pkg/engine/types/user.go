package types

import "github.com/kwilteam/kwil-db/pkg/crypto"

// UserIdentifier is an interface for identifying a user by public key
type UserIdentifier interface {
	MarshalBinary() ([]byte, error)
	PubKey() (crypto.PublicKey, error)
	UnmarshalBinary(data []byte) error
}
