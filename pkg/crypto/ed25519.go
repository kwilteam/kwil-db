package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
)

const Ed25519 KeyType = "Ed25519"

type Ed25519PrivateKey struct {
	key ed25519.PrivateKey
}

func (pv *Ed25519PrivateKey) Bytes() []byte {
	return pv.key
}

func (pv *Ed25519PrivateKey) PubKey() PublicKey {
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

func (pv *Ed25519PrivateKey) Type() KeyType {
	return Ed25519
}

type Ed25519PublicKey struct {
	key ed25519.PublicKey
}

func (pub *Ed25519PublicKey) Address() Address {
	return Ed25519Address(pub.key[:20])
}

func (pub *Ed25519PublicKey) Bytes() []byte {
	return pub.key
}

func (pub *Ed25519PublicKey) Type() KeyType {
	return Ed25519
}

// Verify verifies the given signature against the given message(not hashed).
// This expects a standard ed25519 signature, 64 bytes long.
func (pub *Ed25519PublicKey) Verify(sig []byte, msg []byte) error {
	if len(sig) != ed25519.SignatureSize {
		return errInvalidSignature
	}

	ok := ed25519.Verify(pub.key, msg, sig)
	if !ok {
		return errVerifySignatureFailed
	}
	return nil
}

type Ed25519Address [20]byte

func (s Ed25519Address) Bytes() []byte {
	return s[:]
}

func (s Ed25519Address) Type() KeyType {
	return Ed25519
}

func (s Ed25519Address) String() string {
	// TODO: need an address format
	return hex.EncodeToString(s.Bytes())
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
