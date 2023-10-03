package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type Ed25519PrivateKey struct {
	key ed25519.PrivateKey
}

func (pv *Ed25519PrivateKey) Bytes() []byte {
	return pv.key
}

func (pv *Ed25519PrivateKey) PubKey() *Ed25519PublicKey {
	publicKey := make([]byte, ed25519.PublicKeySize)
	copy(publicKey, pv.key[32:])
	return &Ed25519PublicKey{
		key: publicKey,
	}
}

func (pv *Ed25519PrivateKey) Hex() string {
	return hex.EncodeToString(pv.Bytes())
}

// Sign signs the given message(not hashed). This returns a standard ed25519 signature, 64 bytes long.
func (pv *Ed25519PrivateKey) Sign(msg []byte) ([]byte, error) {
	return ed25519.Sign(pv.key, msg), nil
}

type Ed25519PublicKey struct {
	key ed25519.PublicKey
}

func (pub *Ed25519PublicKey) Bytes() []byte {
	return pub.key
}

// Verify verifies the given signature against the given message(not hashed).
// This expects a standard ed25519 signature, 64 bytes long.
func (pub *Ed25519PublicKey) Verify(sig []byte, msg []byte) error {
	if len(sig) != ed25519.SignatureSize {
		return fmt.Errorf("ed25519: %w: expected: %d, got: %d", ErrInvalidSignatureLength, ed25519.SignatureSize, len(sig))
	}

	ok := ed25519.Verify(pub.key, msg, sig)
	if !ok {
		return ErrInvalidSignature
	}
	return nil
}

// GenerateEd25519Key generates a new ed25519 key pair.
func GenerateEd25519Key() (*Ed25519PrivateKey, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &Ed25519PrivateKey{
		key: priv,
	}, nil
}
