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

// Key represents a crypto key that can be compared to another key
type Key interface {
	// Equals checks whether two PubKeys are the same
	Equals(Key) bool

	// Bytes returns the raw bytes of the key.
	Bytes() []byte

	// Type returns the protobuf key type.
	Type() KeyType
}

type KeyType int32

const (
	KeyTypeSecp256k1 KeyType = 0
	KeyTypeEd25519   KeyType = 1
	KeyTypeRSA       KeyType = 2
)

// PrivateKey represents a private key that can be used to generate a public key and sign data
type PrivateKey interface {
	Key

	// Cryptographically sign the given bytes
	Sign([]byte) ([]byte, error)

	// Public a public key paired with this private key
	Public() PublicKey
}

// PublicKey is a public key
type PublicKey interface {
	Key

	// Verify that 'sig' is the signed hash of 'data'
	Verify(data []byte, sig []byte) (bool, error)
}

var panicWriter io.Writer = os.Stderr

func HandlePanic(rerr interface{}, err *error, where string) {
	if rerr != nil {
		fmt.Fprintf(panicWriter, "caught panic: %s\n%s\n", rerr, debug.Stack())
		*err = fmt.Errorf("panic in %s: %s", where, rerr)
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
	if k1 == k2 {
		return true
	}

	return k1.Equals(k2)
}

func MarshalPrivateKey(key PrivateKey) []byte {
	// encode the type before the raw bytes
	b := binary.LittleEndian.AppendUint32(nil, uint32(key.Type()))
	return append(b, key.Bytes()...)
}

func UnmarshalPrivateKey(b []byte) (PrivateKey, error) {
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
