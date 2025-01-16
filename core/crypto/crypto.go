// Package crypto implements asymmetric (public/private key) cryptographic
// signing and verification for Kwil.
//
// This package is based on the go-libp2p crypto package. This uses the pure Go
// dcrec module for secp256k1 keys and ecdsa signing and verification, and the
// standard library's ed25519 package for Ed25519 keys and signing.
package crypto

import (
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
)

const (
	// ReservedKeyTypes is the range of key types that are reserved for internal purposes.
	ReservedKeyTypes = KeyType(1 << 16) // 65536
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
	// ErrInvalidSignatureLength = errors.New("invalid signature length")
)

// Key represents a public or private key that can be serialized and compared to
// another key. The Type method will indicate the type of key used.  See
// [KeyType] for the supported key types.
type Key interface {
	// Equals checks whether two Keys are the same.
	Equals(Key) bool

	// Bytes returns the raw bytes of the key.
	Bytes() []byte

	// Type returns the key type.
	Type() KeyType
}

// KeyType is the type of key, which may be public or private depending on context.
type KeyType int32

// The supported key types are secp256k1 and ed25519.
const (
	KeyTypeSecp256k1 KeyType = 0
	KeyTypeEd25519   KeyType = 1
	// KeyTypeRSA       KeyType = 2
)

var (
	// keyTypes maps key types to their string representations
	// of all the supported key types.
	keyTypes map[KeyType]string = map[KeyType]string{
		KeyTypeSecp256k1: "secp256k1",
		KeyTypeEd25519:   "ed25519",
	}

	// maps of keyTypeStrings to KeyType
	keyTypeStrings map[string]KeyType = map[string]KeyType{
		"secp256k1": KeyTypeSecp256k1,
		"ed25519":   KeyTypeEd25519,
	}
)

// RegisterKeyType registers a new keyType. The KeyType and its string should be
// unique and the KeyType value must be greater than ReservedKeyTypes.
func RegisterKeyType(kt KeyType, name string) error {
	if kt <= ReservedKeyTypes {
		return fmt.Errorf("key type %d is reserved", kt)
	}

	if _, ok := keyTypes[kt]; ok {
		return fmt.Errorf("key type %d already registered with %s", kt, keyTypes[kt])
	}

	if _, ok := keyTypeStrings[name]; ok {
		return fmt.Errorf("key type string %s already registered with keyType: %d", name, keyTypeStrings[name])
	}

	keyTypes[kt] = name
	keyTypeStrings[name] = kt
	return nil
}

func (kt KeyType) Valid() bool {
	_, ok := keyTypes[kt]
	return ok
}

func (kt KeyType) String() string {
	if s, ok := keyTypes[kt]; ok {
		return s
	}
	return fmt.Sprintf("unknown key type %d", kt)
}

// ParseKeyType parses a string into a KeyType.
func ParseKeyType(s string) (KeyType, error) {
	if kt, ok := keyTypeStrings[s]; ok {
		return kt, nil
	}
	return 0, fmt.Errorf("unknown key type: %s", s)
}

// PrivateKey represents a private key that can be used to sign data.
type PrivateKey interface {
	Key

	// Sign cryptographically signs the given bytes.
	Sign([]byte) ([]byte, error)

	// Public returns the public key associated with this private key.
	Public() PublicKey
}

// PublicKey is a public key, which can verify signatures.
type PublicKey interface {
	Key

	// Verify that 'sig' is the signed hash of 'data'
	Verify(data []byte, sig []byte) (bool, error)
}

var panicWriter io.Writer = os.Stderr

// handlePanic allows for deferred one-liners that capture an error.
func handlePanic(rerr interface{}, err *error, where string) {
	if rerr != nil {
		fmt.Fprintf(panicWriter, "caught panic: %v\n%s\n", rerr, debug.Stack())
		*err = fmt.Errorf("panic in %v: %v", where, rerr)
	}
}

func keyEquals(k1, k2 Key) bool {
	if k1.Type() != k2.Type() {
		return false
	}

	a := k1.Bytes()
	b := k2.Bytes()
	return subtle.ConstantTimeCompare(a, b) == 1
}

// KeyEquals checks whether two Keys are equivalent (same type and bytes).
func KeyEquals(k1, k2 Key) bool {
	if k1 == k2 { // interface type and data both equal, possibly unnecessary shortcut
		return true
		// comparable: "Two interface values are equal if they have identical
		// dynamic types and equal dynamic values or if both have value nil."
		// https://go.dev/ref/spec#Comparison_operators
	}
	if k1 == nil || k2 == nil {
		return false
	}

	return k1.Equals(k2)
}

// WireEncodePrivateKey serializes a private key for transmitting the type on
// the wire or persistent storage, encoding the key type. See
// [WireDecodePrivateKey]. This is a different serialization from the raw bytes
// obtained by the key's Bytes method.
func WireEncodePrivateKey(key PrivateKey) []byte {
	// encode the type before the raw bytes
	b := binary.LittleEndian.AppendUint32(nil, uint32(key.Type()))
	return append(b, key.Bytes()...)
}

// WireDecodePrivateKey deserializes a private key from a network or disk
// source, decoding the key type. See [WireEncodePrivateKey]. This is a
// different serialization from the raw bytes obtained by the key's Bytes
// method.
func WireDecodePrivateKey(b []byte) (PrivateKey, error) {
	if len(b) <= 4 { // 4 bytes for key type, at least
		return nil, errors.New("insufficient data for private key")
	}

	// read the uint32 key type.
	keyType := KeyType(binary.LittleEndian.Uint32(b))

	// unmarshal the bytes of the specific key type
	switch keyType {
	case KeyTypeSecp256k1:
		return UnmarshalSecp256k1PrivateKey(b[4:])
	case KeyTypeEd25519:
		return UnmarshalEd25519PrivateKey(b[4:])
	default:
		return nil, fmt.Errorf("invalid key type %v", keyType)
	}
}

// WireEncodePublicKey encodes public key and key type into a byte slice.
// This is different than using the key's Bytes method, which returns the
// raw bytes of the key; this is a serialization of the key type and the
// raw bytes that is suitable for transmitting on the wire or storing on
// disk.
func WireEncodePublicKey(key PublicKey) []byte {
	b := binary.LittleEndian.AppendUint32(nil, uint32(key.Type()))
	return append(b, key.Bytes()...)
}

// WireDecodePubKey decodes a public key from a network or disk source.
func WireDecodePubKey(b []byte) (PublicKey, error) {
	if len(b) <= 4 {
		return nil, errors.New("insufficient data for public key")
	}

	keyType := KeyType(binary.LittleEndian.Uint32(b))

	// unmarshal the bytes of the specific key type
	switch keyType {
	case KeyTypeSecp256k1:
		return UnmarshalSecp256k1PublicKey(b[4:])
	case KeyTypeEd25519:
		return UnmarshalEd25519PublicKey(b[4:])
	default:
		return nil, fmt.Errorf("invalid key type %v", keyType)
	}
}

func UnmarshalPublicKey(data []byte, keyType KeyType) (PublicKey, error) {
	switch keyType {
	case KeyTypeSecp256k1:
		return UnmarshalSecp256k1PublicKey(data)
	case KeyTypeEd25519:
		return UnmarshalEd25519PublicKey(data)
	default:
		return nil, fmt.Errorf("invalid key type %v", keyType)
	}
}

func UnmarshalPrivateKey(data []byte, keyType KeyType) (PrivateKey, error) {
	switch keyType {
	case KeyTypeSecp256k1:
		return UnmarshalSecp256k1PrivateKey(data)
	case KeyTypeEd25519:
		return UnmarshalEd25519PrivateKey(data)
	default:
		return nil, fmt.Errorf("invalid key type %v", keyType)
	}
}

func GeneratePrivateKey(keyType KeyType) (PrivateKey, error) {
	switch keyType {
	case KeyTypeSecp256k1:
		priv, _, err := GenerateSecp256k1Key(nil)
		return priv, err
	case KeyTypeEd25519:
		priv, _, err := GenerateEd25519Key(nil)
		return priv, err
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}
