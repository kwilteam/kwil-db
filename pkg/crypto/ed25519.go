package crypto

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

func (s *Ed25519PrivateKey) Sign(msg []byte) (*Signature, error) {
	panic("implement me")
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

func (s *Ed25519PublicKey) Verify(sig *Signature, data []byte) error {
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
