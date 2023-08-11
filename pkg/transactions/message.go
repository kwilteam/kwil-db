/*
Package transactions contains all the logic for creating and validating
transactions and signed messages.
*/
package transactions

import (
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/serialize/rlp"
)

// SignedMessage is any message that has been signed by a private key
// It contains a signature and a payload.  The message in the signature
// should be the hash of the payload.
type SignedMessage struct {
	Signature *crypto.Signature
	Message   rlp.SerializedData
	Sender    crypto.PublicKey
}

// Verify verifies the authenticity of a signed message.
// It does this by reconstructing the hash of the payload and comparing
// it to the message in the signature.
// It then uses the public key in the signature to verify the signature.
func (s *SignedMessage) Verify() error {
	return s.Sender.Verify(s.Signature, s.Message)
}
