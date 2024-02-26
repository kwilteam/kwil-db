package crypto

type KeyType string

type PrivateKey interface {
	Bytes() []byte
	Type() KeyType
	Sign(msg []byte) (*Signature, error)
	PubKey() PublicKey
	Hex() string
}

type PublicKey interface {
	Bytes() []byte
	Type() KeyType
	Verify(sig *Signature, data []byte) error
	Address() Address
}

type Address interface {
	Bytes() []byte
	String() string
	// do we need to know the key type?
	Type() KeyType
}
