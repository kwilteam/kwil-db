package crypto

type KeyType string

type PrivateKey interface {
	Bytes() []byte
	Type() KeyType
	Sign(msg []byte, signatureType SignatureType) ([]byte, error)
	PubKey() PublicKey
}

type PublicKey interface {
	Bytes() []byte
	Type() KeyType
	Verify(sign *Signature) error
	Address() Address
}

type Address interface {
	Bytes() []byte
	String() string
	// do we need to know the key type?
	Type() KeyType
}
