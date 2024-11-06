// Package crypto implements asymmetric (public/private key) cryptographic
// signing and verification for Kwil.
//
// This package is based on the go-libp2p crypto package. This uses the pure Go
// dcrec module for secp256k1 keys and ecdsa signing and verification, and the
// standard library's ed25519 package for Ed25519 keys and signing.
package crypto

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
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

func Sha256(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}
