package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

const (
	Secp256k1CompressedPublicKeySize       = 33
	Secp256k1UncompressedPublicKeySize     = 65
	secp256k1SignatureLength               = 64
	secp256k1SignatureWithRecoveryIDLength = 65
)

type Secp256k1PrivateKey struct {
	key *ecdsa.PrivateKey
}

func (pv *Secp256k1PrivateKey) Bytes() []byte {
	return ethCrypto.FromECDSA(pv.key)
}

func (pv *Secp256k1PrivateKey) Hex() string {
	return hex.EncodeToString(pv.Bytes())
}

func (pv *Secp256k1PrivateKey) PubKey() *Secp256k1PublicKey {
	return &Secp256k1PublicKey{
		publicKey: &pv.key.PublicKey,
	}
}

// Sign signs the given hash directly utilizing go-ethereum's Sign function.
// go-ethereum returns a secp256k1 signature, in [R || S || V] format where V is 0 or 1, 65 bytes long.
// We want to remove the recovery ID, so we return a 64 byte signature, in [R || S] format.
func (pv *Secp256k1PrivateKey) Sign(hash []byte) ([]byte, error) {
	signature, err := ethCrypto.Sign(hash, pv.key)
	if err != nil {
		return nil, err
	}

	// remove recovery ID
	return signature[:len(signature)-1], nil
}

// SignWithRecoveryID signs the given hash directly utilizing go-ethereum's Sign function.
// It includes go-ethereum's recovery ID, which while it is non-standard for Secp256k1,
// is very common in Bitcoin and Ethereum
func (pv *Secp256k1PrivateKey) SignWithRecoveryID(hash []byte) ([]byte, error) {
	return ethCrypto.Sign(hash, pv.key)
}

type Secp256k1PublicKey struct {
	publicKey *ecdsa.PublicKey
}

func (pub *Secp256k1PublicKey) Bytes() []byte {
	return ethCrypto.FromECDSAPub(pub.publicKey)
}

// Verify verifies the standard secp256k1 signature against the given hash.
// Caller of this function should make sure the signature is in one of the following two formats:
// - 65 bytes, [R || S || V] format. This is the standard format.
// - 64 bytes, [R || S] format.
//
// Since `Verify` suppose to verify the signature produced from `Sign` function, it expects the signature to be
// 65 bytes long, and in [R || S || V] format where V is 0 or 1.
// In this implementation, we use `VerifySignature`, which doesn't care about the recovery ID, so it can
// also support 64 bytes [R || S] format signature like cometbft.
// e.g. this `Verify` function is able to verify multi-signature-schema like personal_sign, eip712, cometbft, etc.,
// as long as the given signature is in supported format.
func (pub *Secp256k1PublicKey) Verify(sig []byte, hash []byte) error {
	if len(sig) == secp256k1SignatureWithRecoveryIDLength {
		// we choose `VerifySignature` since it doesn't care recovery ID
		// it expects signature in 64 byte [R || S] format
		sig = sig[:len(sig)-1]
	}

	if len(sig) != secp256k1SignatureLength {
		return fmt.Errorf("secp256k1: %w: expected: %d received: %d", ErrInvalidSignatureLength, secp256k1SignatureLength, len(sig))
	}

	if !ethCrypto.VerifySignature(pub.Bytes(), hash, sig) {
		return ErrInvalidSignature
	}

	return nil
}

// GenerateSecp256k1Key generates a new secp256k1 private key.
func GenerateSecp256k1Key() (*Secp256k1PrivateKey, error) {
	key, err := ethCrypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return &Secp256k1PrivateKey{
		key: key,
	}, nil
}
