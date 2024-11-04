package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"       // key/curve
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa" // signature algorithm
)

// Secp256k1PrivateKey is a Secp256k1 private key
type Secp256k1PrivateKey secp256k1.PrivateKey

// Secp256k1PublicKey is a Secp256k1 public key
type Secp256k1PublicKey secp256k1.PublicKey

// GenerateSecp256k1Key generates a new Secp256k1 private and public key pair
func GenerateSecp256k1Key(src io.Reader) (PrivateKey, PublicKey, error) {
	if src == nil {
		src = rand.Reader
	}
	privk, err := secp256k1.GeneratePrivateKeyFromRand(src)
	if err != nil {
		return nil, nil, err
	}

	k := (*Secp256k1PrivateKey)(privk)
	return k, k.Public(), nil
}

// UnmarshalSecp256k1PrivateKey returns a private key from bytes
func UnmarshalSecp256k1PrivateKey(data []byte) (k PrivateKey, err error) {
	if len(data) != secp256k1.PrivKeyBytesLen {
		return nil, fmt.Errorf("expected secp256k1 data size to be %d", secp256k1.PrivKeyBytesLen)
	}
	defer func() { HandlePanic(recover(), &err, "secp256k1 private-key unmarshal") }()

	privk := secp256k1.PrivKeyFromBytes(data)
	return (*Secp256k1PrivateKey)(privk), nil
}

// UnmarshalSecp256k1PublicKey returns a public key from bytes
func UnmarshalSecp256k1PublicKey(data []byte) (k PublicKey, err error) {
	defer func() { HandlePanic(recover(), &err, "secp256k1 public-key unmarshal") }()
	key, err := secp256k1.ParsePubKey(data)
	if err != nil {
		return nil, err
	}

	return (*Secp256k1PublicKey)(key), nil
}

var _ PrivateKey = (*Secp256k1PrivateKey)(nil)

// Type returns the private key type
func (k *Secp256k1PrivateKey) Type() KeyType {
	return KeyTypeSecp256k1
}

// Bytes returns the bytes of the key
func (k *Secp256k1PrivateKey) Bytes() []byte {
	return (*secp256k1.PrivateKey)(k).Serialize()
}

// Equals compares two private keys
func (k *Secp256k1PrivateKey) Equals(o Key) bool {
	sk, ok := o.(*Secp256k1PrivateKey)
	if !ok {
		return keyEquals(k, o)
	}

	return k.Public().Equals(sk.Public())
}

// Sign returns a signature from input data. The signature is of the sha256 hash
// of the data, not data itself. This is to  match the other key types that
// internally use a hash function, unlike secp256k1, which does not.
func (k *Secp256k1PrivateKey) Sign(data []byte) (rawSig []byte, err error) {
	defer func() { HandlePanic(recover(), &err, "secp256k1 signing") }()
	key := (*secp256k1.PrivateKey)(k)
	hash := sha256.Sum256(data)
	sig := ecdsa.Sign(key, hash[:])

	return sig.Serialize(), nil
}

// Public returns a public key
func (k *Secp256k1PrivateKey) Public() PublicKey {
	return (*Secp256k1PublicKey)((*secp256k1.PrivateKey)(k).PubKey())
}

var _ PublicKey = (*Secp256k1PublicKey)(nil)

// Type returns the public key type
func (k *Secp256k1PublicKey) Type() KeyType {
	return KeyTypeSecp256k1
}

// Bytes returns the bytes of the key
func (k *Secp256k1PublicKey) Bytes() []byte {
	var err error
	defer func() { HandlePanic(recover(), &err, "secp256k1 public key marshaling") }()
	return (*secp256k1.PublicKey)(k).SerializeCompressed()
}

// Equals compares two public keys
func (k *Secp256k1PublicKey) Equals(o Key) bool {
	sk, ok := o.(*Secp256k1PublicKey)
	if !ok {
		return keyEquals(k, o)
	}

	return (*secp256k1.PublicKey)(k).IsEqual((*secp256k1.PublicKey)(sk))
}

// Verify compares a signature against the input data. The data is hashed with
// sha256 internally.
func (k *Secp256k1PublicKey) Verify(data, sigStr []byte) (success bool, err error) {
	defer func() {
		HandlePanic(recover(), &err, "secp256k1 signature verification")

		// To be extra safe.
		if err != nil {
			success = false
		}
	}()
	sig, err := ecdsa.ParseDERSignature(sigStr)
	if err != nil {
		return false, err
	}

	hash := sha256.Sum256(data)
	return sig.Verify(hash[:], (*secp256k1.PublicKey)(k)), nil
}
