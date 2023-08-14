package crypto

import (
	"crypto/ecdsa"
	c256 "crypto/sha256"
	c512 "crypto/sha512"
	"encoding/hex"

	ec "github.com/ethereum/go-ethereum/crypto"
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

func PublicKeyFromBytes(key []byte) (PublicKey, error) {
	panic("not implemented")
}

func PrivateKeyFromHex(h string) (PrivateKey, error) {
	panic("not implemented")
}

func PrivateKeyFromBytes(key []byte) (PrivateKey, error) {
	panic("not implemented")
}

func ECDSAFromHex(hex string) (*ecdsa.PrivateKey, error) {
	return ec.HexToECDSA(hex)
}
