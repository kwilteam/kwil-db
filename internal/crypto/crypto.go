package crypto

import (
<<<<<<< HEAD
	"crypto"
	"crypto/ecdsa"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
=======
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
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

<<<<<<< HEAD
	key, err := ethcrypto.HexToECDSA(string(hKey))
=======
	key, err := crypto.HexToECDSA(string(hKey))
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	if err != nil {
		return nil, err
	}

	return &PrivateKey{key}, nil
}
<<<<<<< HEAD

// Sha384 returns the sha384 hash of the data.
func Sha384(data []byte) []byte { // I wrapped this in a function so that we know it is standard
	return crypto.SHA384.New().Sum(data)
}
=======
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
