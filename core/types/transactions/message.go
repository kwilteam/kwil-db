/*
Package transactions contains all the logic for creating and validating
transactions and call messages.
*/
package transactions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types/serialize"
)

// callMsgToSignTmplV0 is the template for the message to be signed.
const callMsgToSignTmplV0 = `%s

DBID: %s
Action: %s
PayloadDigest: %x
`

// CallMessageBody is the body of a call message.
type CallMessageBody struct {
	// Description is a human-readable description of the message
	Description string

	// Payload is the payload of the message, it is RLP encoded
	Payload serialize.SerializedData
}

// SerializeMsg serializes the message body and returns a result for signing
// and verification.
func (b *CallMessageBody) SerializeMsg(mst SignedMsgSerializationType) ([]byte, error) {
	if len(b.Description) > MsgDescriptionMaxLength {
		return nil, fmt.Errorf("description is too long")
	}

	// restore the payload first, we need it for the template
	var payload ActionCall
	err := payload.UnmarshalBinary(b.Payload)
	if err != nil {
		return nil, fmt.Errorf("unable to restore payload: %w", err)
	}

	switch mst {
	case SignedMsgConcat:
		// NOTE: this is kind silly, since we use both RLP encoded payload(for
		// digest) and raw payload in the message(for dbid and action).
		payloadDigest := crypto.Sha256(b.Payload)[:20]
		msgStr := fmt.Sprintf(callMsgToSignTmplV0,
			b.Description,
			strings.ToLower(payload.DBID),
			payload.Action, // action name is case-sensitive
			payloadDigest)
		return []byte(msgStr), nil
	}

	return nil, errors.New("invalid serialization type")
}

// CallMessage represents a message could be used to call an action.
// This is meant to work like transactions.Transaction, except that it is not a transaction.
type CallMessage struct {
	// Body is the body of the actual message
	Body *CallMessageBody

	// Signature is the signature of the message
	// optional, only required if the action requires authentication
	Signature *auth.Signature

	// Sender is the public key of the sender
	// optional, only required if the action requires authentication
	Sender []byte

	// Serialization is the serialization performed on `Body`
	// inorder to generate the message that being signed
	Serialization SignedMsgSerializationType
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
		Serialization: DefaultSignedMsgSerType,
	}, nil
}

// SerializeMsg produces the serialization of the message that is to be used in
// both signing and verification of message.
func (s *CallMessage) SerializeMsg() ([]byte, error) {
	return s.Body.SerializeMsg(s.Serialization)
}

// Sign signs message body with given signer. It will serialize the message
// body to get message-to-be-sign first, then sign it.
func (s *CallMessage) Sign(signer auth.Signer) error {
	msg, err := s.SerializeMsg()
	if err != nil {
		return err
	}

	signature, err := signer.Sign(msg)
	if err != nil {
		return err
	}

	s.Signature = signature
	s.Sender = signer.Identity()
	return nil
}

// IsSigned returns true if the message is signed.
func (s *CallMessage) IsSigned() bool {
	return s.Signature != nil && s.Sender != nil
}
