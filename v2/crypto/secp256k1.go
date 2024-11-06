package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"       // key/curve
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa" // signature algorithm
	"golang.org/x/crypto/sha3"
)

// Secp256k1PrivateKey is a Secp256k1 private key.
type Secp256k1PrivateKey secp256k1.PrivateKey

// Secp256k1PublicKey is a Secp256k1 public key.
type Secp256k1PublicKey secp256k1.PublicKey

// GenerateSecp256k1Key generates a new Secp256k1 private and public key pair.
// If the provided io.Reader is nil, crypto/rand.Reader is used. The returned
// keys may be cast to *Secp256k1PrivateKey and *Secp256k1PublicKey.
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

// UnmarshalSecp256k1PrivateKey returns a private key from the key's raw bytes.
func UnmarshalSecp256k1PrivateKey(data []byte) (k *Secp256k1PrivateKey, err error) {
	if len(data) != secp256k1.PrivKeyBytesLen {
		return nil, fmt.Errorf("expected secp256k1 data size to be %d", secp256k1.PrivKeyBytesLen)
	}
	defer func() { handlePanic(recover(), &err, "secp256k1 private-key unmarshal") }()

	privk := secp256k1.PrivKeyFromBytes(data)
	return (*Secp256k1PrivateKey)(privk), nil
}

// UnmarshalSecp256k1PublicKey returns a public key from the key's raw bytes.
func UnmarshalSecp256k1PublicKey(data []byte) (k *Secp256k1PublicKey, err error) {
	defer func() { handlePanic(recover(), &err, "secp256k1 public-key unmarshal") }()
	key, err := secp256k1.ParsePubKey(data)
	if err != nil {
		return nil, err
	}

	return (*Secp256k1PublicKey)(key), nil
}

var _ PrivateKey = (*Secp256k1PrivateKey)(nil)

// Type returns the private key type.
func (k *Secp256k1PrivateKey) Type() KeyType {
	return KeyTypeSecp256k1
}

// Bytes returns the raw bytes of the key. To serialize for the wire or disk,
// use WireEncodePrivateKey to maintain the key type.
func (k *Secp256k1PrivateKey) Bytes() []byte {
	return (*secp256k1.PrivateKey)(k).Serialize()
}

// Equals compares two private keys. This accepts a Key to satisfy the
// PrivateKey interface.
func (k *Secp256k1PrivateKey) Equals(o Key) bool {
	sk, ok := o.(*Secp256k1PrivateKey)
	if !ok {
		return keyEquals(k, o) // if different concrete type, test based on returns form the interface's Type and Bytes
	}

	return k.Public().Equals(sk.Public())
}

// Sign returns a signature (uncompressed) from input data. The signature is of
// the sha256 hash of the data, not data itself. This is to  match the other key
// types that internally use a hash function, unlike secp256k1, which does not.
func (k *Secp256k1PrivateKey) Sign(data []byte) (rawSig []byte, err error) {
	defer func() { handlePanic(recover(), &err, "secp256k1 signing") }()
	key := (*secp256k1.PrivateKey)(k)
	hash := sha256.Sum256(data)
	// sig := ecdsa.Sign(key, hash[:])
	sig := ecdsa.SignCompact(key, hash[:], false)

	v := sig[0] - 27
	copy(sig, sig[1:])
	sig[RecoveryIDOffset] = v

	return sig, nil
}

// Public returns a public key. This is a Secp256k1PublicKey as a PublicKey to
// satisfy the PrivateKey interface.
func (k *Secp256k1PrivateKey) Public() PublicKey {
	return (*Secp256k1PublicKey)((*secp256k1.PrivateKey)(k).PubKey())
}

var _ PublicKey = (*Secp256k1PublicKey)(nil)

// Type returns the public key type.
func (k *Secp256k1PublicKey) Type() KeyType {
	return KeyTypeSecp256k1
}

// Bytes returns the bytes of the key.
func (k *Secp256k1PublicKey) Bytes() []byte {
	var err error // discarded, since SerializeCompressed returns no error
	defer func() { handlePanic(recover(), &err, "secp256k1 public key marshaling") }()
	return (*secp256k1.PublicKey)(k).SerializeCompressed()
}

// BytesUncompressed returns the bytes of the key.
func (k *Secp256k1PublicKey) BytesUncompressed() []byte {
	var err error // discarded, since SerializeUncompressed returns no error
	defer func() { handlePanic(recover(), &err, "secp256k1 public key marshaling") }()
	return (*secp256k1.PublicKey)(k).SerializeUncompressed()
}

// Equals compares two public keys. This accepts a Key to satisfy the PublicKey interface.
func (k *Secp256k1PublicKey) Equals(o Key) bool {
	sk, ok := o.(*Secp256k1PublicKey)
	if !ok {
		return keyEquals(k, o)
	}

	return (*secp256k1.PublicKey)(k).IsEqual((*secp256k1.PublicKey)(sk))
}

// Verify compares a signature against the input data. The data is hashed with
// sha256 internally.
func (k *Secp256k1PublicKey) Verify(data, rawSig []byte) (success bool, err error) {
	defer func() {
		handlePanic(recover(), &err, "secp256k1 signature verification")

		// To be extra safe.
		if err != nil {
			success = false
		}
	}()

	if len(rawSig) != 65 {
		return false, errors.New("invalid signature length")
	}

	rawSig = rawSig[:RecoveryIDOffset]

	var r, s secp256k1.ModNScalar
	if r.SetByteSlice(rawSig[:32]) {
		return false, errors.New("r value overflow")
	}
	if s.SetByteSlice(rawSig[32:]) {
		return false, errors.New("s value overflow")
	}
	sig := ecdsa.NewSignature(&r, &s)
	hash := sha256.Sum256(data)
	return sig.Verify(hash[:], (*secp256k1.PublicKey)(k)), nil
}

// SignatureLength indicates the byte length required to carry a signature with recovery id.
const SignatureLength = 64 + 1 // 64 bytes ECDSA signature + 1 byte recovery id
// RecoveryIDOffset points to the byte offset within the signature that contains the recovery id.
const RecoveryIDOffset = 64

// RecoverSecp256k1Key recovers a secp256k1 public key from a signature and
// signed message, which is hashed with sha256 internally.
func RecoverSecp256k1Key(msg, sig []byte) (*Secp256k1PublicKey, error) {
	if len(sig) != SignatureLength {
		return nil, errors.New("invalid signature")
	}

	hash := sha256.Sum256(msg)
	return RecoverSecp256k1KeyFromSigHash(hash[:], sig)
}

// RecoverSecp256k1KeyFromSigHash is like [RecoverSecp256k1Key], except the
// input is the sighash, which was directly signed. This is the more proper
// implementation of secp256k1 pubkey recovery since a specific hash function is
// not associated with secp256k1 or ecdsa.
func RecoverSecp256k1KeyFromSigHash(hash, sig []byte) (*Secp256k1PublicKey, error) {
	if len(sig) != SignatureLength {
		return nil, errors.New("invalid signature")
	}

	// Convert to secp256k1 input format with 'recovery id' v at the beginning.
	btcsig := make([]byte, SignatureLength)
	btcsig[0] = sig[RecoveryIDOffset] + 27
	copy(btcsig[1:], sig)
	pub, _, err := ecdsa.RecoverCompact(btcsig, hash)
	return (*Secp256k1PublicKey)(pub), err
}

func EthereumAddressFromPubKey(pubKey *Secp256k1PublicKey) []byte {
	// Serialize the public key to 65 bytes (uncompressed format).
	pubKeyBytes := (*secp256k1.PublicKey)(pubKey).SerializeUncompressed()
	// Remove the first byte (0x04), which indicates that this is an uncompressed public key.
	pubKeyBytes = pubKeyBytes[1:]
	// Apply Keccak256 (SHA3-256) hashing.
	hash := sha3.NewLegacyKeccak256()
	hash.Write(pubKeyBytes)
	fullHash := hash.Sum(nil)
	// Take the last 20 bytes of the hash as the Ethereum address.
	return fullHash[len(fullHash)-20:]
}
