/*
Package transactions contains all the logic for creating and validating
transactions and call messages.
*/
package transactions

import (
	"github.com/kwilteam/kwil-db/core/types/serialize"
)

// CallMessageBody is the body of a call message.
type CallMessageBody struct {
	// Payload is the payload of the message, it is RLP encoded
	Payload serialize.SerializedData
}

// CallMessage represents a message could be used to call an action.
// This is meant to work like transactions.Transaction, except that it is not a transaction.
type CallMessage struct {
	// Body is the body of the actual message
	Body *CallMessageBody

	// the type of authenticator, which will be used to derive 'identifier'
	// from the 'sender`
	AuthType string

	// Sender is the public key of the sender
	Sender []byte
}

// CreateCallMessage creates a new call message from a ActionCall payload.
func CreateCallMessage(payload *ActionCall) (*CallMessage, error) {
	bts, err := payload.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &CallMessage{
		Body: &CallMessageBody{
			Payload: bts,
		},
	}, nil
}
