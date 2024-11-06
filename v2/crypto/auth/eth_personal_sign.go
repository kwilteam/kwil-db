package auth

import (
	"bytes"
	"fmt"
	"kwil/crypto"

	ethCommon "github.com/ethereum/go-ethereum/common"
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
	pubkey, err := crypto.RecoverSecp256k1Key(msg, signature)
	if err != nil {
		return err
	}

	addr := crypto.EthereumAddressFromPubKey(pubkey)

	if !bytes.Equal(addr, identity) {
		return fmt.Errorf("invalid signature: expected address %x, received %x", identity, addr)
	}

	return nil
}
