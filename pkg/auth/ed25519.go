package auth

// ed25519 is a standard signature scheme that uses the ed25519 curve, and does not have a build tag

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/crypto"
)

func init() {
	err := RegisterAuthenticator(Ed25519Auth, Ed25519Authenticator{})
	if err != nil {
		panic(err)
	}
}

// ed25519 constants
const (
	// using Ed25519Auth for the authenticator name

	// ed25519SignatureLength is the expected length of a signature
	ed25519SignatureLength = 64
)

// Ed25519Authenticator is an authenticator for ed25519 keys
type Ed25519Authenticator struct{}

var _ Authenticator = Ed25519Authenticator{}

// Address simply returns the public key as the address
func (e Ed25519Authenticator) Address(publicKey []byte) (string, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return "", fmt.Errorf("invalid ed25519 public key size: %d", len(publicKey))
	}

	return hex.EncodeToString(publicKey), nil
}

// Verify verifies the signature against the given public key and data.
func (e Ed25519Authenticator) Verify(publicKey []byte, msg []byte, signature []byte) error {
	pubkey, err := crypto.Ed25519PublicKeyFromBytes(publicKey)
	if err != nil {
		return err
	}

	if len(signature) != ed25519SignatureLength {
		return fmt.Errorf("invalid signature length: expected %d, received %d", ed25519SignatureLength, len(signature))
	}

	return pubkey.Verify(signature, msg)
}
