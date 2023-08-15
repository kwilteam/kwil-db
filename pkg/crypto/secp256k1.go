package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"
	ethAccount "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

const Secp256k1 KeyType = "secp256k1"

type Secp256k1PrivateKey struct {
	privateKey *ecdsa.PrivateKey
}

func (s *Secp256k1PrivateKey) Bytes() []byte {
	return ethCrypto.FromECDSA(s.privateKey)
}

func (s *Secp256k1PrivateKey) Hex() string {
	return hex.EncodeToString(s.Bytes())
}

func (s *Secp256k1PrivateKey) PubKey() PublicKey {
	return &Secp256k1PublicKey{
		publicKey: &s.privateKey.PublicKey,
	}
}

// SignMsg signs the given message(not hashed) according to EIP-191 personal_sign.
// This is default signature type for sec256k1.
func (s *Secp256k1PrivateKey) SignMsg(msg []byte) (*Signature, error) {
	hash := ethAccount.TextHash(msg)
	sig, err := s.Sign(hash)
	if err != nil {
		return nil, err
	}
	return &Signature{
		Signature: sig,
		Type:      SIGNATURE_TYPE_SECP256K1_PERSONAL,
	}, nil
}

// Sign signs the given hash utilizing go-ethereum's Sign function.
func (s *Secp256k1PrivateKey) Sign(hash []byte) ([]byte, error) {
	return ethCrypto.Sign(hash, s.privateKey)
}

func (s *Secp256k1PrivateKey) Type() KeyType {
	return Secp256k1
}

type Secp256k1PublicKey struct {
	publicKey *ecdsa.PublicKey
}

func (s *Secp256k1PublicKey) Address() Address {
	return &Secp256k1Address{
		address: ethCrypto.PubkeyToAddress(*s.publicKey),
	}
}

func (s *Secp256k1PublicKey) Bytes() []byte {
	return ethCrypto.FromECDSAPub(s.publicKey)
}

func (s *Secp256k1PublicKey) Type() KeyType {
	return Secp256k1
}

// Verify verifies the given signature against the given message according to EIP-191
// personal sign.
func (s *Secp256k1PublicKey) Verify(sig []byte, hash []byte) error {
	if len(sig) != 64 {
		return errInvalidSignature
	}

	// signature should have the 64 byte [R || S] format
	if !ethCrypto.VerifySignature(s.Bytes(), hash, sig) {
		return errVerifySignatureFailed
	}

	return nil
}

type Secp256k1Address struct {
	address common.Address
}

func (s *Secp256k1Address) Bytes() []byte {
	return s.address.Bytes()
}

func (s *Secp256k1Address) Type() KeyType {
	return Secp256k1
}

func (s *Secp256k1Address) String() string {
	return s.address.Hex()
}
