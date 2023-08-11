package crypto

import (
	oasisEd25519 "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
)

const Ed25519 KeyType = "Ed25519"

type Ed25519PrivateKey struct {
	privateKey []byte
}

func (s *Ed25519PrivateKey) Bytes() []byte {
	return s.privateKey
}

func (s *Ed25519PrivateKey) PubKey() PublicKey {
	panic("implement me")
}

func (s *Ed25519PrivateKey) Sign(msg []byte, signatureType SignatureType) ([]byte, error) {
	return oasisEd25519.Sign(oasisEd25519.PrivateKey(s.privateKey), msg), nil
}

func (s *Ed25519PrivateKey) Type() KeyType {
	return Ed25519
}

type Ed25519PublicKey struct {
}

func (s *Ed25519PublicKey) Address() Address {
	panic("implement me")
}

func (s *Ed25519PublicKey) Bytes() []byte {
	panic("implement me")
}

func (s *Ed25519PublicKey) Type() KeyType {
	return Ed25519
}

func (s *Ed25519PublicKey) Verify(sig *Signature) error {
	panic("implement me")
}

type Ed25519Address struct {
}

func (s *Ed25519Address) Bytes() []byte {
	panic("implement me")
}

func (s *Ed25519Address) Type() KeyType {
	return Ed25519
}

func (s *Ed25519Address) String() string {
	panic("implement me")
}
