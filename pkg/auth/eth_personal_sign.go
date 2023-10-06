package auth

// eth_personal_sign is a default signer, and does not include a build tag

import (
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/crypto"

	ethAccounts "github.com/ethereum/go-ethereum/accounts"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

func init() {
	err := RegisterAuthenticator(EthPersonalSignAuth, EthSecp256k1Authenticator{})
	if err != nil {
		panic(err)
	}
}

const (
	// using EthPersonalSignAuth for the authenticator name
	EthPersonalSignAuth = "secp256k1_ep"

	// ethPersonalSignSignatureLength is the expected length of a signature
	ethPersonalSignSignatureLength = 65
)

// EthSecp256k1Authenticator is an authenticator for Ethereum secp256k1 keys
// It is provided as a default authenticator
type EthSecp256k1Authenticator struct{}

var _ Authenticator = EthSecp256k1Authenticator{}

// Address generates an ethereum address from a public key
func (e EthSecp256k1Authenticator) Address(publicKey []byte) (string, error) {
	ethKey, err := ethCrypto.UnmarshalPubkey(publicKey)
	if err != nil {
		return "", err
	}

	return ethCrypto.PubkeyToAddress(*ethKey).Hex(), nil
}

// Verify verifies applies the Ethereum TextHash digest and verifies the signature
func (e EthSecp256k1Authenticator) Verify(publicKey []byte, msg []byte, signature []byte) error {
	pubkey, err := crypto.Secp256k1PublicKeyFromBytes(publicKey)
	if err != nil {
		return err
	}

	// signature is 65 bytes, [R || S || V] format
	if len(signature) != ethPersonalSignSignatureLength {
		return fmt.Errorf("invalid signature length: expected %d, received %d", ethPersonalSignSignatureLength, len(signature))
	}
	hash := ethAccounts.TextHash(msg)

	// trim off the recovery id
	return pubkey.Verify(signature, hash)
}
