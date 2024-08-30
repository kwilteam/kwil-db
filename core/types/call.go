package types

// This file defines CallMessage, which is the type used for RPCs that perform a
// (read-only) action call. Also defined here is the ActionCall struct, which
// contains the arguments to the action call. This type is a BinaryMarshaler and
// BinaryUnmarshaler as it is transmitted serialized in CallMessage.Body.

import (
	"encoding"

	"github.com/kwilteam/kwil-db/core/types/serialize"
)

// CallMessage represents a message could be used to call an action. This is
// meant to work like transactions.Transaction, with a signed body and the
// signer's public key, except that it is not a transaction.
type CallMessage struct {
	// Body is the body of the actual message. This should be unmarshaled into
	// an ActionCall.
	Body serialize.SerializedData `json:"body"`

	// AuthType is the type of authenticator, which will be used to derive
	// 'identifier' from the 'sender`
	AuthType string `json:"auth_type"`

	// Sender is the public key of the sender
	Sender HexBytes `json:"sender"`
}

// CreateCallMessage creates a new call message from a ActionCall payload.
func CreateCallMessage(payload *ActionCall) (*CallMessage, error) {
	bts, err := payload.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &CallMessage{
		Body: bts,
	}, nil
}

// ActionCall models the arguments of an action call. It would be serialized
// into CallMessage.Body. This is not a transaction payload. See
// transactions.ActionExecution for the transaction payload used for executing
// an action.
type ActionCall struct {
	DBID      string
	Action    string
	Arguments []*EncodedValue
}

func (a *ActionCall) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(a)
}

func (a *ActionCall) UnmarshalBinary(b serialize.SerializedData) error {
	return serialize.Decode(b, a)
}

var _ encoding.BinaryUnmarshaler = (*ActionCall)(nil)
var _ encoding.BinaryMarshaler = (*ActionCall)(nil)
