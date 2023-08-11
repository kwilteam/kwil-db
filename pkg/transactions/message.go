/*
Package transactions contains all the logic for creating and validating
transactions and signed messages.
*/
package transactions

import (
	"bytes"
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/crypto"
)

// SignedMessage is any message that has been signed by a private key
// It contains a signature and a payload.  The message in the signature
// should be the hash of the payload.
type SignedMessage struct {
	Signature *crypto.Signature
	Message   Serializable
}

// Verify verifies the authenticity of a signed message.
// It does this by reconstructing the hash of the payload and comparing
// it to the message in the signature.
// It then uses the public key in the signature to verify the signature.
func (s *SignedMessage) Verify() error {
	messageBytes, err := s.Message.Bytes()
	if err != nil {
		return err
	}

	if !bytes.Equal(s.Signature.Message, crypto.Sha256(messageBytes)) {
		return ErrFailedHashReconstruction
	}

	return s.Verify()
}

// Serializable is any message that can be hashed
// This hash is used to both sign the message, as well as reconstruct the
// hash (if we are checking the message) to verify the signature
type Serializable interface {
	Bytes() ([]byte, error)
}

// TransactionMessageis the payload of a transaction
// It contains information on the nonce, fee, and the transaction payload
// The type of the transaction can be derived from the payload
type TransactionMessage struct {
	Nonce int64
	Fee   *big.Int
}

// AuthenticatedMessage is a message that has been signed by a private key
// it is not a valid blockchain transaction, but can be used to authenticate
