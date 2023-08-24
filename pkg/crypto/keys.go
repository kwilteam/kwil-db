package crypto

type KeyType string

const UnknownKeyType KeyType = "unknown"

type PrivateKey interface {
	Bytes() []byte
	Type() KeyType
	// Sign generate signature on data. Data could be hashed or not, depends on the implementation
	Sign(data []byte) ([]byte, error)
	PubKey() PublicKey
	Hex() string
	Signer() Signer
}

type PublicKey interface {
	Bytes() []byte
	Type() KeyType
	// Verify verify signature against data. Data could be hashed or not, depends on the implementation
	Verify(sig []byte, data []byte) error
	Address() Address
}

type Address interface {
	Bytes() []byte
	String() string
	// do we need to know the key type?
	Type() KeyType
}