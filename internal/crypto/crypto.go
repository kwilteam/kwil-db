package crypto

import (
	"crypto"
	"crypto/ecdsa"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/kwilteam/kwil-db/internal/utils/files"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

// LoadPrivateKey loads a private key from a file relative to the root directory
func LoadPrivateKey(path string) (*PrivateKey, error) {
	hKey, err := files.LoadFileFromRoot(path)
	if err != nil {
		return nil, err
	}

	key, err := ethcrypto.HexToECDSA(string(hKey))
	if err != nil {
		return nil, err
	}

	return &PrivateKey{key}, nil
}

// Sha384 returns the sha384 hash of the data.
func Sha384(data []byte) []byte { // I wrapped this in a function so that we know it is standard
	return crypto.SHA384.New().Sum(data)
}
