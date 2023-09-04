package crypto

import (
	"crypto/ed25519"
	c256 "crypto/sha256"
	c512 "crypto/sha512"
	"encoding/hex"
	"fmt"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

var (
	errInvalidPrivateKey = fmt.Errorf("invalid private key")
	errInvalidPublicKey  = fmt.Errorf("invalid public key")
)

// Sha384 returns the sha384 hash of the data.
func Sha384(data []byte) []byte { // I wrapped this in a function so that we know it is standard
	h := c512.New384()
	h.Write(data)
	return h.Sum(nil)
}

func Sha384Hex(data []byte) string {
	return hex.EncodeToString(Sha384(data))
}

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
		return nil, errInvalidPrivateKey
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
		return nil, errInvalidPublicKey
	}
	return &Ed25519PublicKey{key: key}, nil
}

func PrivateKeyFromHex(keyType KeyType, key string) (PrivateKey, error) {
	switch keyType {
	case Secp256k1:
		return Secp256k1PrivateKeyFromHex(key)
	case Ed25519:
		return Ed25519PrivateKeyFromHex(key)
	default:
		return nil, fmt.Errorf("invalid key type: %s", keyType)
	}
}

func PublicKeyFromBytes(keyType KeyType, key []byte) (PublicKey, error) {
	switch keyType {
	case Secp256k1:
		return Secp256k1PublicKeyFromBytes(key)
	case Ed25519:
		return Ed25519PublicKeyFromBytes(key)
	default:
		return nil, fmt.Errorf("invalid key type %s", keyType)
	}
}
