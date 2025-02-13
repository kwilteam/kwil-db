package types

import (
	"crypto/sha256"
	"fmt"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
)

// CallMessageBody is the body of a call message. The serialization of this body
// is signed when authenticated call RPCs are enabled.
type CallMessageBody struct {
	// Payload is the payload of the message, it is RLP encoded
	Payload []byte `json:"payload"`

	// Why not just use the type? There are no other possible payloads and no
	// payload type field anyway.
	// CallData *ActionCall `json:"params"`

	// Challenge is a random value for call authentication with replay
	// protection. It is provided by the authenticating RPC server, where it is
	// generated with a csPRNG.
	Challenge []byte `json:"challenge"`
}

// CallMessage represents a message could be used to call an action.
// This is meant to work like transactions.Transaction, except that it is not a transaction.
type CallMessage struct {
	// Body is the body of the actual message
	Body *CallMessageBody `json:"body"`

	// the type of authenticator, which will be used to derive 'identifier'
	// from the 'sender`
	AuthType string `json:"auth_type"`

	// Sender is the public key of the sender
	Sender HexBytes `json:"sender"`

	// SignatureData is the content of is the sender's signature of the
	// serialized call body. This is only set when using authenticated call
	// RPCs, in which case the Challenge field of the call body is also set.
	// Note that this was historically called Signature, which was an
	// *auth.Signature struct, but it is now a []byte that represents just the
	// signature data since the type is already in the AuthType field above.
	SignatureData []byte `json:"signature"`
}

const callMsgToSignTmplV0 = `Kwil view call.

Namespace: %s
Method: %s
Digest: %x
Challenge: %x
`

func CallSigText(namespace, action string, payload []byte, challenge []byte) string {
	digest := sha256.Sum256(payload)
	return fmt.Sprintf(callMsgToSignTmplV0, namespace, action, digest[:20], challenge)
}

// TODO: support authentication on execute SQL (ad hoc querys)
// kgw client must use StmtSigText

const stmtMsgToSignTmplV0 = `Kwil SQL statement.

Statement: %s
Digest: %x
Challenge: %x
`

func StmtSigText(stmt string, digest []byte, challenge []byte) string {
	return fmt.Sprintf(stmtMsgToSignTmplV0, stmt, digest[:20], challenge)
}

// CreateCallMessage creates a new call message from a ActionCall payload. If a
// signer is provided, the sender and authenticator type are set. If a challenge
// is also provided, it will also sign a serialization of the request that
// includes the challenge for replay protection. Thus, if a challenge is
// provided, a signer must also be provided.
func CreateCallMessage(ac *ActionCall, challenge []byte, signer auth.Signer) (*CallMessage, error) {
	if ac.Action == "" {
		return nil, fmt.Errorf("invalid action call")
	}
	bts, err := ac.MarshalBinary()
	if err != nil {
		return nil, err
	}

	msg := &CallMessage{
		Body: &CallMessageBody{
			Payload:   bts,
			Challenge: challenge,
		},
	}

	if signer != nil { // for @caller
		msg.AuthType = signer.AuthType()
		msg.Sender = signer.CompactID()
	}

	if len(challenge) > 0 {
		sigText := CallSigText(ac.Namespace, ac.Action, bts, challenge)
		sig, err := signer.Sign([]byte(sigText))
		if err != nil {
			return nil, err
		}

		msg.SignatureData = sig.Data
	}

	return msg, nil
}
