package auth

import (
	"github.com/kwilteam/kwil-db/core/crypto"

	ethAccount "github.com/ethereum/go-ethereum/accounts"
)

// Signature is a signature with a designated AuthType, which should
// be used to determine how to verify the signature.
// It seems a bit weird to have a field "Signature" inside a struct called "Signature",
// but I am keeping it like this for compatibility with the old code.
type Signature struct {
	// Signature is the raw signature bytes
	Signature []byte `json:"signature_bytes"`
	// Type is the signature type, which must have a registered Authenticator of
	// the same name for the Verify method to be usable.
	Type string `json:"signature_type"`
}

// Signer is an interface for something that can sign messages.
// It returns signatures with a designated AuthType, which should
// be used to determine how to verify the signature.
type Signer interface {
	// Sign signs a message and returns the signature
	Sign(msg []byte) (*Signature, error)

	// PublicKey returns the public key of the signer
	PublicKey() []byte

	// Authenticator returns the corresponding Authenticator that must work with
	// signatures produced by this signer. Not all Authenticators have a
	// corresponding Signer, but all Signers must have an Authenticator, which
	// this method guarantees.
	Authenticator() Authenticator
}

// EthPersonalSecp256k1Signer is a signer that signs messages using the
// secp256k1 curve, using ethereum's personal_sign signature scheme.
type EthPersonalSigner struct {
	Key crypto.Secp256k1PrivateKey
}

var _ Signer = (*EthPersonalSigner)(nil)

// Sign sign given message according to EIP-191 personal_sign.
// EIP-191 personal_sign prefix the message with "\x19Ethereum Signed Message:\n"
// and the message length, then hash the message with 'legacy' keccak256.
// The signature is in [R || S || V] format, 65 bytes.
// This method is used to sign an arbitrary message in the same manner in which
// a wallet like MetaMask would sign a text message. The message is defined by
// the object that is being serialized e.g. a Kwil Transaction.
func (e *EthPersonalSigner) Sign(msg []byte) (*Signature, error) {
	signatureBts, err := e.Key.SignWithRecoveryID(ethAccount.TextHash(msg))
	if err != nil {
		return nil, err
	}

	return &Signature{
		Signature: signatureBts,
		Type:      EthPersonalSignAuth,
	}, nil
}

// Address generates an ethereum address from a public key.
func (e *EthPersonalSigner) Address() (string, error) {
	pubBytes := e.PublicKey()
	return EthSecp256k1Authenticator{}.Address(pubBytes)
}

// Authenticator returns the Authenticator capable of Verifying signatures
// produced by this Signer.
func (e *EthPersonalSigner) Authenticator() Authenticator {
	return EthSecp256k1Authenticator{}
}

// PublicKey returns the public key of the signer
func (e *EthPersonalSigner) PublicKey() []byte {
	return e.Key.PubKey().Bytes()
}

// Ed25519Signer is a signer that signs messages using the
// ed25519 curve, using the standard signature scheme.
type Ed25519Signer struct {
	crypto.Ed25519PrivateKey
}

var _ Signer = (*Ed25519Signer)(nil)

// Sign signs the given message(not hashed) according to standard signature scheme.
// It does not apply any special digests to the message.
func (e *Ed25519Signer) Sign(msg []byte) (*Signature, error) {
	signatureBts, err := e.Ed25519PrivateKey.Sign(msg)
	if err != nil {
		return nil, err
	}

	return &Signature{
		Signature: signatureBts,
		Type:      Ed25519Auth,
	}, nil
}

// PublicKey returns the public key of the signer
func (e *Ed25519Signer) PublicKey() []byte {
	return e.Ed25519PrivateKey.PubKey().Bytes()
}

// Address generates an ethereum address from a public key.
func (e *Ed25519Signer) Address() (string, error) {
	pubBytes := e.PublicKey()
	return Ed25519Authenticator{}.Address(pubBytes)
}

// Authenticator returns the Authenticator capable of Verifying signatures
// produced by this Signer.
func (e *Ed25519Signer) Authenticator() Authenticator {
	return Ed25519Authenticator{}
}
