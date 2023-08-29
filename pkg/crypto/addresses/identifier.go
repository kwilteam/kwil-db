package addresses

import (
	"bytes"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/crypto"
)

// KeyIdentifier identifies a public key type, the key itself,
// and the desired address format.
type KeyIdentifier struct {
	KeyType       KeyType
	AddressFormat AddressFormat
	PublicKey     []byte
}

// CreateKeyIdentifier creates a KeyIdentifier from a public key and address format.
// It will check to make sure the address format is compatible with the key type.
func CreateKeyIdentifier(pubkey crypto.PublicKey, format AddressFormat) (*KeyIdentifier, error) {
	var keyType KeyType
	switch pubkey.Type() {
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidKeyType, pubkey.Type())
	case crypto.Secp256k1:
		keyType = Secp256k1
	case crypto.Ed25519:
		keyType = Ed25519
	}

	k := &KeyIdentifier{
		KeyType:       keyType,
		AddressFormat: format,
		PublicKey:     pubkey.Bytes(),
	}

	// check that it is valid
	if err := k.Check(); err != nil {
		return nil, err
	}

	return k, nil
}

// MarshalBinary marshals the KeyIdentifier into a binary representation.
// It simply prepends the key type and address format to the public key.
func (k *KeyIdentifier) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	if err := buf.WriteByte(byte(k.KeyType)); err != nil {
		return nil, err
	}

	if err := buf.WriteByte(byte(k.AddressFormat)); err != nil {
		return nil, err
	}

	if _, err := buf.Write(k.PublicKey); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary unmarshals the KeyIdentifier from a binary representation.
func (k *KeyIdentifier) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	k.KeyType = KeyType(data[0])
	k.AddressFormat = AddressFormat(data[1])
	k.PublicKey = data[2:]

	return nil
}

// Check checks that the key types and address formats are valid,
// as well as compatible with each other.
func (k *KeyIdentifier) Check() error {
	if err := k.KeyType.Valid(); err != nil {
		return err
	}

	if err := k.AddressFormat.Valid(); err != nil {
		return err
	}

	pubKey, err := k.PubKey()
	if err != nil {
		return err
	}

	_, err = GenerateAddress(pubKey, k.AddressFormat)
	if err != nil {
		return err
	}

	return nil
}

// PubKey returns a public key from the KeyIdentifier
func (k *KeyIdentifier) PubKey() (crypto.PublicKey, error) {
	var keyType crypto.KeyType
	switch k.KeyType {
	default:
		return nil, fmt.Errorf("%w: %d", ErrInvalidKeyType, k.KeyType)
	case Secp256k1:
		keyType = crypto.Secp256k1
	case Ed25519:
		keyType = crypto.Ed25519
	}

	return crypto.PublicKeyFromBytes(keyType, k.PublicKey)
}

// Address returns the address of the KeyIdentifier.
func (k *KeyIdentifier) Address() (crypto.Address, error) {
	pubKey, err := k.PubKey()
	if err != nil {
		return nil, err
	}

	return GenerateAddress(pubKey, k.AddressFormat)
}

// KeyType is a uint8 representation of key types.
// Since this is commonly duplicated data in a database,
// we use a uint8 to save space.
type KeyType uint8

const (
	// Secp256k1 is the key type for secp256k1 keys.
	Secp256k1 KeyType = iota
	// Ed25519 is the key type for ed25519 keys.
	Ed25519
)

// Valid returns an error if the key type is an invalid enum
func (k KeyType) Valid() error {
	switch k {
	default:
		return fmt.Errorf("%w: %d", ErrInvalidKeyType, k)
	case Secp256k1, Ed25519:
		return nil
	}
}
