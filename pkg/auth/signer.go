package auth

import (
	ethAccount "github.com/ethereum/go-ethereum/accounts"
	"github.com/kwilteam/kwil-db/pkg/crypto"
)

const (
	EthPersonalSignAuth = "secp256k1_ep"
	Ed25519Auth         = "ed25519"
)

// Signature is a signature with a designated AuthType, which should
// be used to determine how to verify the signature.
// It seems a bit weird to have a field "Signature" inside a struct called "Signature",
// but I am keeping it like this for compatibility with the old code.
type Signature struct {
	// Signature is the raw signature bytes
	Signature []byte `json:"signature_bytes"`
	Type      string `json:"signature_type"`
}

// Verify verifies the signature against the given message and public key.
func (s *Signature) Verify(senderPubKey, msg []byte) error {
	a, err := getAuthenticator(s.Type)
	if err != nil {
		return err
	}

	return a.Verify(senderPubKey, msg, s.Signature)
}

// Signer is an interface for something that can sign messages.
// It returns signatures with a designated AuthType, which should
// be used to determine how to verify the signature.
type Signer interface {
	// Sign signs a message and returns the signature
	Sign(msg []byte) (*Signature, error)

	// PublicKey returns the public key of the signer
	PublicKey() []byte
}

// EthPersonalSecp256k1Signer is a signer that signs messages using the
// secp256k1 curve, using ethereum's personal_sign signature scheme.
type EthPersonalSigner struct {
	crypto.Secp256k1PrivateKey
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
	signatureBts, err := e.Secp256k1PrivateKey.SignWithRecoveryID(ethAccount.TextHash(msg))
	if err != nil {
		return nil, err
	}

	return &Signature{
		Signature: signatureBts,
		Type:      EthPersonalSignAuth,
	}, nil
}

// PublicKey returns the public key of the signer
func (e *EthPersonalSigner) PublicKey() []byte {
	return e.Secp256k1PrivateKey.PubKey().Bytes()
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
