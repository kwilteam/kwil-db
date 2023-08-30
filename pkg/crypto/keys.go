package crypto

type KeyType string

func (kt KeyType) String() string {
	return string(kt)
}

const UnknownKeyType KeyType = "unknown"

type PrivateKey interface {
	Bytes() []byte
	Type() KeyType
	// Sign generate signature on data. Data could be hashed or not, depends on the implementation
	Sign(data []byte) ([]byte, error)
	PubKey() PublicKey
	Hex() string
}

type PublicKey interface {
	Bytes() []byte
	Type() KeyType
	// Verify verify signature against data. Data could be hashed or not, depends on the implementation
	Verify(sig []byte, data []byte) error
	Address() Address
}

type Address interface {
	String() string
}
