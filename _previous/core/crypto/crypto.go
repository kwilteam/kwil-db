package crypto

import (
	"crypto/ed25519"
	c256 "crypto/sha256"
	"encoding/hex"
	"fmt"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

func Sha224(data []byte) []byte {
	h := c256.New224()
	h.Write(data)
	return h.Sum(nil)
}

func Sha224Hex(data []byte) string {
	return hex.EncodeToString(Sha224(data))
}

func Sha256(data []byte) []byte {
	h := c256.New()
	h.Write(data)
	return h.Sum(nil)
}

func Sha256Hex(data []byte) string {
	return hex.EncodeToString(Sha256(data))
}

func Secp256k1PublicKeyFromBytes(key []byte) (*Secp256k1PublicKey, error) {
	pk, err := ethCrypto.UnmarshalPubkey(key)
	if err != nil {
		return nil, err
	}
	return &Secp256k1PublicKey{publicKey: pk}, nil
}

func Secp256k1PrivateKeyFromHex(key string) (*Secp256k1PrivateKey, error) {
	pk, err := ethCrypto.HexToECDSA(key)
	if err != nil {
		return nil, err
	}
	return &Secp256k1PrivateKey{key: pk}, nil
}

func Ed25519PrivateKeyFromBytes(key []byte) (*Ed25519PrivateKey, error) {
	if len(key) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid ed25519 private key length: %d", len(key))
	}
	return &Ed25519PrivateKey{key: key}, nil
}

func Ed25519PrivateKeyFromHex(key string) (*Ed25519PrivateKey, error) {
	pkBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	return Ed25519PrivateKeyFromBytes(pkBytes)
}

func Ed25519PublicKeyFromBytes(key []byte) (*Ed25519PublicKey, error) {
	if len(key) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid ed25519 public key length: %d", len(key))
	}
	return &Ed25519PublicKey{key: key}, nil
}
