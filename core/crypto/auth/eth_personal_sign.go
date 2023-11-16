package auth

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/crypto"

	ethAccounts "github.com/ethereum/go-ethereum/accounts"
	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

const (
	// EthPersonalSignAuth is the Ethereum "personal sign" authentication type,
	// which uses the secp256k1 signature scheme with a prefixed message and the
	// legacy 256-bit Keccak hash function to mimic most Ethereum wallets. This
	// is intended as the authenticator for the SDK-provided EthPersonalSigner,
	// and must be registered with that name.
	EthPersonalSignAuth = "secp256k1_ep"

	// ethPersonalSignSignatureLength is the expected length of a signature
	ethPersonalSignSignatureLength = 65
)

// EthSecp256k1Authenticator is the authenticator for the Ethereum "personal
// sign" signature type, which is the default signer for Kwil. As such, it is a
// default authenticator.
type EthSecp256k1Authenticator struct{}

var _ Authenticator = EthSecp256k1Authenticator{}

// Identifier returns an ethereum address hex string from address bytes.
// It will include the 0x prefix, and the address will be checksum-able.
func (EthSecp256k1Authenticator) Identifier(ident []byte) (string, error) {
	return ethCommon.BytesToAddress(ident).Hex(), nil
}

// Verify verifies applies the Ethereum TextHash digest and verifies the signature
func (EthSecp256k1Authenticator) Verify(identity []byte, msg []byte, signature []byte) error {
	// signature is 65 bytes, [R || S || V] format
	if len(signature) != ethPersonalSignSignatureLength {
		return fmt.Errorf("invalid signature length: expected %d, received %d", ethPersonalSignSignatureLength, len(signature))
	}
	hash := ethAccounts.TextHash(msg)

	pubkeyBytes, err := ethCrypto.Ecrecover(hash, signature)
	if err != nil {
		return err
	}

	pubkey, err := crypto.Secp256k1PublicKeyFromBytes(pubkeyBytes)
	if err != nil {
		return err
	}

	// The contract of (*Secp256k1PublicKey).Verify is to have any recovery byte
	// at the end, if it is present. If verification fails here, it is possible
	// that the recovery byte was the first byte, as is common in other domains.
	// The EthPersonalSigner provided by the SDK will always work with this.
	return pubkey.Verify(signature, hash)
}
