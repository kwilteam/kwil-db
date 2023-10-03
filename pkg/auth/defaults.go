package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	ethAccount "github.com/ethereum/go-ethereum/accounts"
	"github.com/kwilteam/kwil-db/pkg/crypto"

	"crypto/ed25519"

	cometCrypto "github.com/cometbft/cometbft/crypto"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/ripemd160" //nolint: staticcheck // necessary for Bitcoin address format
)

// init registers default authenticators
func init() {
	if err := errors.Join(
		RegisterAuthenticator(EthAuth, ethSecp256k1Authenticator{}),
		RegisterAuthenticator(CometBftSecp256k1Auth, cometBftSecp256k1Authenticator{}),
		RegisterAuthenticator(Ed25519Auth, ed25519Authenticator{}),
		RegisterAuthenticator(NearAuth, nearAuthenticator{}),
	); err != nil {
		panic(err)
	}
}

var (
	ErrInvalidSignatureLength = errors.New("invalid signature length")
)

// newInvalidSignatureLength returns an error for invalid signature length.
// It includes the expected and received length.
func newInvalidSignatureLength(expected, received int) error {
	return fmt.Errorf("%w: expected %d, received %d", ErrInvalidSignatureLength, expected, received)
}

// constants for eth personal sign
const (
	// ethAuth is the authenticator name
	EthAuth = "secp256k1_ep"
	// ethPersonalSignSignatureLength is the expected length of a signature
	ethPersonalSignSignatureLength = 65
)

// ethSecp256k1Authenticator is an authenticator for Ethereum secp256k1 keys
// It is provided as a default authenticator
type ethSecp256k1Authenticator struct{}

var _ Authenticator = ethSecp256k1Authenticator{}

// Address generates an ethereum address from a public key
func (e ethSecp256k1Authenticator) Address(publicKey []byte) (string, error) {
	ethKey, err := ethCrypto.UnmarshalPubkey(publicKey)
	if err != nil {
		return "", err
	}

	return ethCrypto.PubkeyToAddress(*ethKey).Hex(), nil
}

// Verify verifies applies the Ethereum TextHash digest and verifies the signature
func (e ethSecp256k1Authenticator) Verify(publicKey []byte, msg []byte, signature []byte) error {
	pubkey, err := crypto.Secp256k1PublicKeyFromBytes(publicKey)
	if err != nil {
		return err
	}

	// signature is 65 bytes, [R || S || V] format
	if len(signature) != ethPersonalSignSignatureLength {
		return newInvalidSignatureLength(ethPersonalSignSignatureLength, len(signature))
	}
	hash := ethAccount.TextHash(msg)

	// trim off the recovery id
	return pubkey.Verify(signature, hash)
}

// cometBftSecp245k1 constants
const (
	// CometBftSecp256k1Auth is the authenticator name
	CometBftSecp256k1Auth = "secp256k1_cmt"
	// cometBftSecp256k1SignatureLength is the expected length of a signature
	cometBftSecp256k1SignatureLength = 64
)

// cometBftSecp256k1Authenticator is an authenticator for CometBFT secp256k1 keys
type cometBftSecp256k1Authenticator struct{}

var _ Authenticator = cometBftSecp256k1Authenticator{}

// Address generates a CometBFT address from a public key
func (e cometBftSecp256k1Authenticator) Address(publicKey []byte) (string, error) {
	compressed, err := getCompressed(publicKey)
	if err != nil {
		return "", err
	}

	sha := sha256.Sum256(compressed[:])
	hasherRIPEMD160 := ripemd160.New()
	_, err = hasherRIPEMD160.Write(sha[:])
	if err != nil {
		return "", err
	}

	return cometCrypto.Address(hasherRIPEMD160.Sum(nil)).String(), nil
}

// Verify verifies applies a sha256 digest and verifies the signature
func (e cometBftSecp256k1Authenticator) Verify(publicKey []byte, msg []byte, signature []byte) error {
	pubkey, err := crypto.Secp256k1PublicKeyFromBytes(publicKey)
	if err != nil {
		return err
	}

	if len(signature) != cometBftSecp256k1SignatureLength {
		return newInvalidSignatureLength(cometBftSecp256k1SignatureLength, len(signature))
	}

	hash := sha256.Sum256(msg)
	return pubkey.Verify(signature, hash[:])
}

// ed25519 constants
const (
	// Ed25519Auth is the authenticator name
	Ed25519Auth = "ed25519"
	// ed25519SignatureLength is the expected length of a signature
	ed25519SignatureLength = 64
)

// ed25519Authenticator is an authenticator for ed25519 keys
type ed25519Authenticator struct{}

var _ Authenticator = ed25519Authenticator{}

// Address simply returns the public key as the address
func (e ed25519Authenticator) Address(publicKey []byte) (string, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return "", fmt.Errorf("invalid ed25519 public key size: %d", len(publicKey))
	}

	return hex.EncodeToString(publicKey), nil
}

// Verify verifies the signature against the given public key and data.
func (e ed25519Authenticator) Verify(publicKey []byte, msg []byte, signature []byte) error {
	pubkey, err := crypto.Ed25519PublicKeyFromBytes(publicKey)
	if err != nil {
		return err
	}

	if len(signature) != ed25519SignatureLength {
		return newInvalidSignatureLength(ed25519SignatureLength, len(signature))
	}

	return pubkey.Verify(signature, msg)
}

const (
	// NearAuth is the authenticator name
	NearAuth = "ed25519_nr"
	// Near uses the same signature length as ed25519
)

// nearAuthenticator is an authenticator for NEAR keys
type nearAuthenticator struct{}

var _ Authenticator = nearAuthenticator{}

// Address generates a NEAR address from a public key
func (e nearAuthenticator) Address(publicKey []byte) (string, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return "", fmt.Errorf("invalid ed25519 public key size for generating near address: %d", len(publicKey))
	}

	return hex.EncodeToString(publicKey), nil
}

// Verify verifies the signature against the given public key and data.
func (e nearAuthenticator) Verify(publicKey []byte, msg []byte, signature []byte) error {
	pubkey, err := crypto.Ed25519PublicKeyFromBytes(publicKey)
	if err != nil {
		return err
	}

	if len(signature) != ed25519SignatureLength {
		return newInvalidSignatureLength(ed25519SignatureLength, len(signature))
	}

	hash := sha256.Sum256(msg)
	return pubkey.Verify(signature, hash[:])
}

// getCompressed returns the compressed bytes of the secp256k1 public key.
// if it is already compressed, it returns the original bytes.
func getCompressed(pubkey []byte) ([]byte, error) {
	switch len(pubkey) {
	default:
		return nil, fmt.Errorf("invalid secp256k1 public key size: %d", len(pubkey))
	case crypto.Secp256k1CompressedPublicKeySize:
		return pubkey, nil
	case crypto.Secp256k1UncompressedPublicKeySize:
		ecdsaPubKey, err := ethCrypto.UnmarshalPubkey(pubkey)
		if err != nil {
			return nil, err
		}

		return ethCrypto.CompressPubkey(ecdsaPubKey), nil
	}
}
