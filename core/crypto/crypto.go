// Package crypto implements asymmetric (public/private key) cryptographic
// signing and verification for Kwil.
//
// This package is based on the go-libp2p crypto package. This uses the pure Go
// dcrec module for secp256k1 keys and ecdsa signing and verification, and the
// standard library's ed25519 package for Ed25519 keys and signing.
package crypto

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
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
