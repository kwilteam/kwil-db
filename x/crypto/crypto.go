package crypto

import (
	"crypto/ecdsa"
	c "crypto/sha512"
	"encoding/hex"

	"kwil/x/utils"

	ec "github.com/ethereum/go-ethereum/crypto"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

// LoadPrivateKey loads a private key from a file relative to the root directory
func LoadPrivateKey(path string) (*PrivateKey, error) {
	hKey, err := utils.LoadFileFromRoot(path)
	if err != nil {
		return nil, err
	}

	key, err := ec.HexToECDSA(string(hKey))
	if err != nil {
		return nil, err
	}

	return &PrivateKey{key}, nil
}

// Sha384 returns the sha384 hash of the data.
func Sha384(data []byte) []byte { // I wrapped this in a function so that we know it is standard
	h := c.New384()
	h.Write(data)
	return h.Sum(nil)
}

func Sha384Str(data []byte) string {
	return hex.EncodeToString(Sha384(data))
}
