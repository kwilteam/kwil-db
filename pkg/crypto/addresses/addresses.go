/*
Package addresses contains functions for generating addresses for various ecosystems

It is kept separate from the Kwil crypto package to encourage usage only when necessary.
This package has a lot of dependencies that are not stable, and therefore should be
avoided when possible.  It should really only be used to provide functions in Kuneiform,
as well as in tooling around Kwil, and should not be used in other core Kwil functionality.
*/
package addresses

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	cometCrypto "github.com/cometbft/cometbft/crypto"
	cometEd25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cometSecp256k1 "github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"golang.org/x/crypto/ripemd160" //nolint: staticcheck // necessary for Bitcoin address format
)

// CreateEthereumAddress returns the ethereum address of the given public key.
// This is the same as secp256k1.Address, but is kept here to be explicit about
// the functionality.  If Kwil decides to adopt its own default address format,
// thos function will continue to return the ethereum address.
func CreateEthereumAddress(pub *crypto.Secp256k1PublicKey) (*EthereumAddress, error) {
	ecdsaPub, err := ethCrypto.UnmarshalPubkey(pub.Bytes())
	if err != nil {
		return nil, err
	}

	return &EthereumAddress{
		address: ethCrypto.PubkeyToAddress(*ecdsaPub),
	}, nil
}

// EthereumAddress is an address for Ethereum
type EthereumAddress struct {
	address common.Address
}

func (addr *EthereumAddress) Bytes() []byte {
	return addr.address.Bytes()
}

func (addr *EthereumAddress) Type() crypto.KeyType {
	return crypto.Secp256k1
}

func (addr *EthereumAddress) String() string {
	return addr.address.Hex()
}

// CreateCometBFTAddress returns the cosmos address of the given public key.
func CreateCometBFTAddress(pub crypto.PublicKey) (*CometBFTAddress, error) {
	switch pub.Type() {
	default:
		return nil, fmt.Errorf("unsupported public key type: %s", pub.Type())
	case crypto.Secp256k1:
		secpKey, ok := pub.(*crypto.Secp256k1PublicKey)
		if !ok {
			return nil, fmt.Errorf("invalid secp256k1 implementation for generating cosmos address: %T", pub)
		}

		publicKeyBytes := secpKey.CompressedBytes()

		if len(publicKeyBytes) != cometSecp256k1.PubKeySize {
			return nil, fmt.Errorf("invalid secp256k1 public key size for generating cosmos address: public key length %d", len(pub.Bytes()))
		}

		sha := sha256.Sum256(publicKeyBytes[:])
		hasherRIPEMD160 := ripemd160.New()
		_, err := hasherRIPEMD160.Write(sha[:])
		if err != nil {
			return nil, err
		}

		return &CometBFTAddress{
			address: cometCrypto.Address(hasherRIPEMD160.Sum(nil)),
		}, nil
	case crypto.Ed25519:
		if len(pub.Bytes()) != cometEd25519.PubKeySize {
			return nil, fmt.Errorf("invalid ed25519 public key size for generating cosmos address: %d", len(pub.Bytes()))
		}

		return &CometBFTAddress{
			address: cometCrypto.Address(tmhash.SumTruncated(pub.Bytes())),
			keyType: crypto.Ed25519,
		}, nil
	}
}

// CometBFTAddress is an address for CometBFT
// It is distinctly different from Cosmos addresses, and is only used for CometBFT
type CometBFTAddress struct {
	address cometCrypto.Address
	keyType crypto.KeyType
}

func (c *CometBFTAddress) Bytes() []byte {
	return c.address.Bytes()
}

func (c *CometBFTAddress) String() string {
	return c.address.String()
}

func (c *CometBFTAddress) Type() crypto.KeyType {
	return c.keyType
}

// CreateNearAddress returns the near address of the given public key.
func CreateNearAddress(pub *crypto.Ed25519PublicKey) (*NEARAddress, error) {
	// NEAR simply uses the public key as hex for their implicit addresses
	// https://docs.near.org/concepts/basics/accounts/creating-accounts#local-implicit-account
	return &NEARAddress{
		address: pub.Bytes(),
	}, nil
}

// NEARAddress is an address for NEAR protocol
// It generates the implicit address from the public key
type NEARAddress struct {
	address []byte
}

func (n *NEARAddress) Bytes() []byte {
	return n.address
}

func (n *NEARAddress) String() string {
	return hex.EncodeToString(n.address)
}

func (n *NEARAddress) Type() crypto.KeyType {
	return crypto.Ed25519
}

type AddressFormat uint8

const (
	// AddressFormatEthereum is the address format for Ethereum
	AddressFormatEthereum AddressFormat = iota
	// AddressFormatCometBFT is the address format for CometBFT
	AddressFormatCometBFT
	// AddressFormatNEAR is the address format for NEAR
	AddressFormatNEAR
)

// Valid returns an error if the address format is an invalid enum
func (a AddressFormat) Valid() error {
	switch a {
	default:
		return fmt.Errorf("invalid address format: %d", a)
	case AddressFormatEthereum, AddressFormatCometBFT, AddressFormatNEAR:
		return nil
	}
}

// GenerateAddress generates the specified address format from the given public key
func GenerateAddress(pubkey crypto.PublicKey, format AddressFormat) (crypto.Address, error) {
	if err := format.Valid(); err != nil {
		return nil, err
	}

	switch format {
	default:
		return nil, fmt.Errorf("unsupported address format: %d", format) // this should get handled by format.Valid()
	case AddressFormatEthereum:
		secp256PubKey, ok := pubkey.(*crypto.Secp256k1PublicKey)
		if !ok {
			return nil, fmt.Errorf("%w: ethereum address format only supports secp256k1 public keys. got %T", ErrIncompatibleAddress, pubkey)
		}

		return CreateEthereumAddress(secp256PubKey)
	case AddressFormatCometBFT:
		return CreateCometBFTAddress(pubkey)
	case AddressFormatNEAR:
		ed25519PubKey, ok := pubkey.(*crypto.Ed25519PublicKey)
		if !ok {
			return nil, fmt.Errorf("%w: near address format only supports ed25519 public keys. got %T", ErrIncompatibleAddress, pubkey)
		}

		return CreateNearAddress(ed25519PubKey)
	}
}
