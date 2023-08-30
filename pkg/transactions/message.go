/*
Package transactions contains all the logic for creating and validating
transactions and signed messages.
*/
package transactions

import (
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/serialize"
)

// CreateSignedMessage creates a signed message from a message.
// This message is used for non-transactional messages.
func CreateSignedMessage(message Payload) (*SignedMessage, error) {
	bts, err := message.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &SignedMessage{
		Message: bts,
	}, nil
}

// SignedMessage is any message that has been signed by a private key
// It contains a signature and a payload.  The message in the signature
// should be the hash of the payload.
type SignedMessage struct {
	Signature *crypto.Signature
	Message   serialize.SerializedData
	Sender    crypto.PublicKey
}

// Verify verifies the authenticity of a signed message.
// It does this by reconstructing the hash of the payload and comparing
// it to the message in the signature.
// It then uses the public key in the signature to verify the signature.
func (s *SignedMessage) Verify() error {
	return s.Signature.Verify(s.Sender, s.Message)
}

// Sign signs a message with a private key.
func (s *SignedMessage) Sign(signer crypto.Signer) error {
	signature, err := signer.Sign(s.Message)
	if err != nil {
		return err
	}
	s.Signature = signature
	s.Sender = signer.PubKey()
	return nil
}
