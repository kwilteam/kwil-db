package crypto

import (
	"crypto/sha256"
	"fmt"

	ethAccount "github.com/ethereum/go-ethereum/accounts"
)

// Signer represents an interface that signs (raw) message.
// You're supposed to use `Signature` to verify the signature.
type Signer interface {
	Sign(msg []byte) (*Signature, error)
	PubKey() PublicKey
}

// trivialSigner is a signer that does nothing, you cannot use it to sign messages.
// This serves as a placeholder for the signer in the case that the private key is not supported.
type trivialSigner struct {
	key PrivateKey
}

func NewTrivialSigner(key PrivateKey) *trivialSigner {
	return &trivialSigner{
		key: key,
	}
}

// Sign just complains.
func (t *trivialSigner) Sign(msg []byte) (*Signature, error) {
	// We can also return a signature with invalid signature type, but caller won't notice this until the signature hit
	// the chain, which is not ideal.
	return nil, fmt.Errorf("you got a trivial signer from unsupported private key, it cannot sign messages")
}

func (t *trivialSigner) PubKey() PublicKey {
	return t.key.PubKey()
}

// CometbftSecp256k1Signer is a signer that signs messages using the secp256k1 curve, using cometbft's signature scheme.
type CometbftSecp256k1Signer struct {
	key *Secp256k1PrivateKey
}

func NewCometbftSecp256k1Signer(key *Secp256k1PrivateKey) *CometbftSecp256k1Signer {
	return &CometbftSecp256k1Signer{
		key: key,
	}
}

func (c *CometbftSecp256k1Signer) PubKey() PublicKey {
	return c.key.PubKey()
}

// Sign signs the given message(not hashed) according to cometbft's signature scheme.
// It use sha256 to hash the message.
// The signature is in [R || S] format, 64 bytes.
func (c *CometbftSecp256k1Signer) Sign(msg []byte) (*Signature, error) {
	hash := Sha256(msg)
	sig, err := c.key.Sign(hash)
	if err != nil {
		return nil, err
	}
	return &Signature{
		Signature: sig[:len(sig)-1],
		Type:      SignatureTypeSecp256k1Cometbft,
	}, nil
}

// EthPersonalSecp256k1Signer is a signer that signs messages using the
// secp256k1 curve, using ethereum's personal_sign signature scheme.
type EthPersonalSecp256k1Signer struct {
	key *Secp256k1PrivateKey
}

func NewEthPersonalSecp256k1Signer(key *Secp256k1PrivateKey) *EthPersonalSecp256k1Signer {
	return &EthPersonalSecp256k1Signer{
		key: key,
	}
}

func (e *EthPersonalSecp256k1Signer) PubKey() PublicKey {
	return e.key.PubKey()
}

// Sign sign given message according to EIP-191 personal_sign.
// EIP-191 personal_sign prefix the message with "\x19Ethereum Signed Message:\n"
// and the message length, then hash the message with 'legacy' keccak256.
// The signature is in [R || S || V] format, 65 bytes.
// This method is used to sign an arbitrary message in the same manner in which
// a wallet like MetaMask would sign a text message. The message is defined by
// the object that is being serialized e.g. a kwil Transaction.
func (e *EthPersonalSecp256k1Signer) Sign(msg []byte) (*Signature, error) {
	hash := ethAccount.TextHash(msg) // prefix and hash
	sig, err := e.key.Sign(hash)
	if err != nil {
		return nil, err
	}
	return &Signature{
		Signature: sig,
		Type:      SignatureTypeSecp256k1Personal,
	}, nil
}

//// EthEIP712Secp256k1Signer is a signer that signs messages using the
//// secp256k1 curve, using ethereum's EIP-712 signature scheme.
//type EthEIP712Secp256k1Signer struct {
//	key *Secp256k1PrivateKey
//}
//
//func NewEthEIP712Secp256k1Signer(key *Secp256k1PrivateKey) *EthEIP712Secp256k1Signer {
//	return &EthEIP712Secp256k1Signer{
//		key: key,
//	}
//}
//
//func (e *EthEIP712Secp256k1Signer) PubKey() PublicKey {
//	return e.key.PubKey()
//}
//
//// Sign sign given message according to EIP-712.
//// EIP-712 prefix the message with `\x19\x01`, then hash the message with keccak256.
//// The signature is in [R || S || V] format, 65 bytes.
//// This method is used to sign structured message. The message is defined by
//// the object that is being serialized e.g. a kwil Transaction. According to
//// EIP-712, the message is `hashStruct(eip712Domain) || hashStruct(message)`.
//func (e *EthEIP712Secp256k1Signer) Sign(msg []byte) (*Signature, error) {
//	rawData := fmt.Sprintf("\x19\x01%s", msg)
//	hash := ethCrypto.Keccak256([]byte(rawData))
//	sig, err := e.key.Sign(hash)
//	if err != nil {
//		return nil, err
//	}
//	return &Signature{
//		Signature: sig,
//		Type:      SignatureTypeSecp256k1Eip712,
//	}, nil
//}

// StdEd25519Signer is a signer that signs messages using the ed25519 curve.
// Vanilla implementation.
type StdEd25519Signer struct {
	key *Ed25519PrivateKey
}

func NewStdEd25519Signer(key *Ed25519PrivateKey) *StdEd25519Signer {
	return &StdEd25519Signer{
		key: key,
	}
}

func (e *StdEd25519Signer) PubKey() PublicKey {
	return e.key.PubKey()
}

// Sign signs the given message(not hashed).
// ed25519 is kind special that it's also an EdDSA signing schema,
// which require sha512 as hashing algorithm(which is handled in downstream lib).
// It returns 64 bytes signature.
func (e *StdEd25519Signer) Sign(msg []byte) (*Signature, error) {
	sig, err := e.key.Sign(msg)
	if err != nil {
		return nil, err
	}
	return &Signature{
		Signature: sig,
		Type:      SignatureTypeEd25519,
	}, nil
}

// NewNearSigner is a signer that signs messages using the ed25519 curve,
// using Near's signature scheme.
func NewNearSigner(key *Ed25519PrivateKey) *NearEd25519Signer {
	return &NearEd25519Signer{
		key: key,
	}
}

type NearEd25519Signer struct {
	key *Ed25519PrivateKey
}

func (n *NearEd25519Signer) PubKey() PublicKey {
	return n.key.PubKey()
}

// Sign signs the given message(not hashed) according to Near's signature scheme.
// It first hash the message with sha256, then sign the hash.
// It returns 64 bytes signature.
func (n *NearEd25519Signer) Sign(msg []byte) (*Signature, error) {
	hash := sha256.Sum256(msg)

	sig, err := n.key.Sign(hash[:])
	if err != nil {
		return nil, err
	}

	return &Signature{
		Signature: sig,
		Type:      SignatureTypeEd25519Near,
	}, nil
}

// DefaultSigner returns a default signer for the given private key.
func DefaultSigner(key PrivateKey) Signer {
	switch key.Type() {
	case Secp256k1:
		return NewEthPersonalSecp256k1Signer(key.(*Secp256k1PrivateKey))
	case Ed25519:
		return NewStdEd25519Signer(key.(*Ed25519PrivateKey))
	default:
		return NewTrivialSigner(key)
	}
}
