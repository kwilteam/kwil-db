package crypto

import (
	"bytes"
	"crypto/ed25519"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
)

// Ed25519PrivateKey is an ed25519 private key.
type Ed25519PrivateKey struct {
	k ed25519.PrivateKey
}

// Ed25519PublicKey is an ed25519 public key.
type Ed25519PublicKey struct {
	k ed25519.PublicKey
}

// GenerateEd25519Key generates a new ed25519 private and public key pair.  The
// returned keys may be cast to *Ed25519PrivateKey and *Ed25519PublicKey.
func GenerateEd25519Key(src io.Reader) (PrivateKey, PublicKey, error) {
	pub, priv, err := ed25519.GenerateKey(src) // crypto/ed25519 will use crypto/rand.Reader if src is nil
	if err != nil {
		return nil, nil, err
	}

	return &Ed25519PrivateKey{
			k: priv,
		},
		&Ed25519PublicKey{
			k: pub,
		},
		nil
}

var _ PrivateKey = (*Ed25519PrivateKey)(nil)

// Type of the private key (Ed25519).
func (k *Ed25519PrivateKey) Type() KeyType {
	return KeyTypeEd25519
}

// Bytes private key bytes.
func (k *Ed25519PrivateKey) Bytes() []byte {
	// The Ed25519 private key contains two 32-bytes curve points, the private
	// key and the public key.
	// It makes it more efficient to get the public key without re-computing an
	// elliptic curve multiplication.
	buf := make([]byte, len(k.k))
	copy(buf, k.k)

	return buf
}

func (k *Ed25519PrivateKey) pubKeyBytes() []byte {
	return k.k[ed25519.PrivateKeySize-ed25519.PublicKeySize:]
}

// Equals compares two ed25519 private keys.
func (k *Ed25519PrivateKey) Equals(o Key) bool {
	edk, ok := o.(*Ed25519PrivateKey)
	if !ok { // if different concrete type, test based on returns form the interface's Type and Bytes
		return keyEquals(k, o)
	}

	return subtle.ConstantTimeCompare(k.k, edk.k) == 1
}

// Public returns an ed25519 public key from a private key.
func (k *Ed25519PrivateKey) Public() PublicKey {
	return &Ed25519PublicKey{k: k.pubKeyBytes()}
}

// Sign returns a signature from an input message.
func (k *Ed25519PrivateKey) Sign(msg []byte) (res []byte, err error) {
	defer func() { handlePanic(recover(), &err, "ed15519 signing") }()

	return ed25519.Sign(k.k, msg), nil
}

var _ PublicKey = (*Ed25519PublicKey)(nil)

// Type of the public key (Ed25519).
func (k *Ed25519PublicKey) Type() KeyType {
	return KeyTypeEd25519
}

// Bytes public key bytes.
func (k *Ed25519PublicKey) Bytes() []byte {
	return k.k
}

// Equals compares two ed25519 public keys.
func (k *Ed25519PublicKey) Equals(o Key) bool {
	edk, ok := o.(*Ed25519PublicKey)
	if !ok {
		return keyEquals(k, o)
	}

	return bytes.Equal(k.k, edk.k)
}

// Verify checks a signature against the input data.
func (k *Ed25519PublicKey) Verify(data []byte, sig []byte) (success bool, err error) {
	defer func() {
		handlePanic(recover(), &err, "ed15519 signature verification")
		success = success && err == nil // to be safe
	}()
	return ed25519.Verify(k.k, data, sig), nil
}

// UnmarshalEd25519PublicKey returns a public key from input bytes.
func UnmarshalEd25519PublicKey(data []byte) (*Ed25519PublicKey, error) {
	if len(data) != 32 {
		return nil, errors.New("expect ed25519 public key data size to be 32")
	}

	return &Ed25519PublicKey{
		k: ed25519.PublicKey(data),
	}, nil
}

// UnmarshalEd25519PrivateKey returns a private key from input bytes.
func UnmarshalEd25519PrivateKey(data []byte) (*Ed25519PrivateKey, error) {
	switch len(data) {
	/*	case ed25519.PrivateKeySize + ed25519.PublicKeySize: // ?? no coverage!
		// Remove the redundant public key. See issue #36.
		redundantPk := data[ed25519.PrivateKeySize:]
		pk := data[ed25519.PrivateKeySize-ed25519.PublicKeySize : ed25519.PrivateKeySize]
		if subtle.ConstantTimeCompare(pk, redundantPk) == 0 {
			return nil, errors.New("expected redundant ed25519 public key to be redundant")
		}

		// No point in storing the extra data.
		newKey := make([]byte, ed25519.PrivateKeySize)
		copy(newKey, data[:ed25519.PrivateKeySize])
		data = newKey
	*/
	case ed25519.PrivateKeySize:
	default:
		return nil, fmt.Errorf(
			"expected ed25519 data size to be %d or %d, got %d",
			ed25519.PrivateKeySize,
			ed25519.PrivateKeySize+ed25519.PublicKeySize,
			len(data),
		)
	}

	return &Ed25519PrivateKey{
		k: ed25519.PrivateKey(data),
	}, nil
}
