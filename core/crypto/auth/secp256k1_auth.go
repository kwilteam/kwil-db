package auth

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/kwilteam/kwil-db/core/crypto"
)

const (
	Secp256k1Auth = "secp256k1"
)

type Secp25k1Authenticator struct{}

var _ Authenticator = Secp25k1Authenticator{}

// Identifier simply returns the hexadecimal encoded compressed public key.
func (e Secp25k1Authenticator) Identifier(publicKey []byte) (string, error) {
	pub, err := crypto.UnmarshalSecp256k1PublicKey(publicKey)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(pub.Bytes()), nil // compressed pubkey
}

// Verify verifies the signature against the given identifier (pubkey bytes) and
// data. The identifier must be the secp256k1 public key bytes.
func (e Secp25k1Authenticator) Verify(publicKey []byte, msg []byte, signature []byte) error {
	hash := sha256.Sum256(msg)
	pubkey, err := crypto.UnmarshalSecp256k1PublicKey(publicKey)
	if err != nil {
		return err
	}

	// pubkeyBack, err := crypto.RecoverSecp256k1KeyFromSigHash(hash[:], signature)
	// if err != nil {
	// 	return err
	// }
	// if !pubkey.Equals(pubkeyBack) {crypto.ErrInvalidSignature }
	valid, err := pubkey.VerifyRaw(hash[:], signature)
	if err != nil {
		return err
	}

	if !valid {
		return crypto.ErrInvalidSignature
	}

	return nil
}
