//go:build auth_ed25519_sha256 || ext_test

package auth

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/auth"
	"github.com/kwilteam/kwil-db/pkg/crypto"
)

func init() {
	err := auth.RegisterAuthenticator(Ed25519Sha256Auth, Ed22519Sha256Authenticator{})
	if err != nil {
		panic(err)
	}
}

const (
	// Ed25519Sha256Auth is the authenticator name
	// the "nr" suffix is for NEAR, and provides backwards compatibility
	Ed25519Sha256Auth = "ed25519_nr"
	// ed25519SignatureLength is the expected length of a signature
	ed25519SignatureLength = 64
)

// Ed22519Sha256Authenticator is an authenticator that applies the sha256 hash to the message
// before verifying the signature. This is a common standard in ecosystems like NEAR.
type Ed22519Sha256Authenticator struct{}

var _ auth.Authenticator = Ed22519Sha256Authenticator{}

// Address generates a NEAR implicit address from a public key
func (e Ed22519Sha256Authenticator) Address(publicKey []byte) (string, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return "", fmt.Errorf("invalid ed25519 public key size for generating near address: %d", len(publicKey))
	}

	return hex.EncodeToString(publicKey), nil
}

// Verify verifies the signature against the given public key and data.
func (e Ed22519Sha256Authenticator) Verify(publicKey []byte, msg []byte, signature []byte) error {
	pubkey, err := crypto.Ed25519PublicKeyFromBytes(publicKey)
	if err != nil {
		return err
	}

	if len(signature) != ed25519SignatureLength {
		return fmt.Errorf("invalid signature length: expected %d, received %d", ed25519SignatureLength, len(signature))
	}

	hash := sha256.Sum256(msg)
	return pubkey.Verify(signature, hash[:])
}
