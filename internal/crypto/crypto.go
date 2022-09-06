package crypto

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
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

	key, err := crypto.HexToECDSA(string(hKey))
	if err != nil {
		return nil, err
	}

	return &PrivateKey{key}, nil
}
