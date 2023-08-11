package crypto

import (
	"crypto/ecdsa"
	c256 "crypto/sha256"
	c512 "crypto/sha512"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

func ECDSAFromHex(hex string) (*ecdsa.PrivateKey, error) {
	return ec.HexToECDSA(hex)
}

func AddressFromPrivateKey(key *ecdsa.PrivateKey) string {
	caddr := ec.PubkeyToAddress(key.PublicKey)
	return caddr.Hex()
}

func IsValidAddress(addr string) bool {
	return common.IsHexAddress(addr)
}

func HexFromECDSAPrivateKey(key *ecdsa.PrivateKey) string {
	return hexutil.Encode(ec.FromECDSA(key))[2:]
}
