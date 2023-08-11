package crypto

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

const Secp256k1 KeyType = "secp256k1"

type Secp256k1PrivateKey struct {
	privateKey *ecdsa.PrivateKey
}

func (s *Secp256k1PrivateKey) Bytes() []byte {
	return ethCrypto.FromECDSA(s.privateKey)
}

func (s *Secp256k1PrivateKey) PubKey() PublicKey {
	return &Secp256k1PublicKey{
		publicKey: &s.privateKey.PublicKey,
	}
}

func (s *Secp256k1PrivateKey) Sign(msg []byte, signatureType SignatureType) ([]byte, error) {
	// TODO: implement
	panic("TODO")
}

func (s *Secp256k1PrivateKey) Type() KeyType {
	return Secp256k1
}

type Secp256k1PublicKey struct {
	publicKey *ecdsa.PublicKey
}

func (s *Secp256k1PublicKey) Address() Address {
	return &Secp256k1Address{
		address: ethCrypto.PubkeyToAddress(*s.publicKey),
	}
}

func (s *Secp256k1PublicKey) Bytes() []byte {
	return ethCrypto.FromECDSAPub(s.publicKey)
}

func (s *Secp256k1PublicKey) Type() KeyType {
	return Secp256k1
}

func (s *Secp256k1PublicKey) Verify(sig *Signature2, data []byte) error {
	// TODO: implement
	panic("TODO")
}

type Secp256k1Address struct {
	address common.Address
}

func (s *Secp256k1Address) Bytes() []byte {
	return s.address.Bytes()
}

func (s *Secp256k1Address) Type() KeyType {
	return Secp256k1
}

func (s *Secp256k1Address) String() string {
	return s.address.String()
}
