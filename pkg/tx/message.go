package tx

import (
	"crypto/ecdsa"
	"encoding/json"

	kwilCrypto "github.com/kwilteam/kwil-db/pkg/crypto"
)

// SignedMessage is a signed message.
// This was made after Transaction, and is made to be more general.
// Unlike Transaction, SignedMessage contains a deserialized payload
type SignedMessage[T Serializable] struct {
	Payload   T // we use generic here to give access to the underlying struct fields
	Signature *kwilCrypto.Signature
	Sender    string
}

type Serializable interface {
	Bytes() ([]byte, error)
}

type Verifiable interface {
	Verify() error
}

func (s *SignedMessage[T]) generateHash() ([]byte, error) {
	data, err := s.Payload.Bytes()
	if err != nil {
		return nil, err
	}

	return kwilCrypto.Sha384(data), nil
}

func (s *SignedMessage[T]) Verify() error {
	hash, err := s.generateHash()
	if err != nil {
		return err
	}

	return s.Signature.Check(s.Sender, hash)
}

// CreateSignedMessage creates and signs a SignedMessage
func CreateSignedMessage[T Serializable](message T, privateKey *ecdsa.PrivateKey) (*SignedMessage[T], error) {
	msg := &SignedMessage[T]{
		Payload: message,
	}

	hash, err := msg.generateHash()
	if err != nil {
		return nil, err
	}

	msg.Sender = kwilCrypto.AddressFromPrivateKey(privateKey)

	sig, err := kwilCrypto.Sign(hash, privateKey)
	if err != nil {
		return nil, err
	}

	msg.Signature = sig

	return msg, nil
}

// CreateEmptySignedMessage creates a SignedMessage and does not sign it
func CreateEmptySignedMessage[T Serializable](payload T) *SignedMessage[T] {
	return &SignedMessage[T]{
		Payload:   payload,
		Signature: &kwilCrypto.Signature{},
		Sender:    "",
	}
}

type CallActionMessage SignedMessage[*CallActionPayload]

// CallActionPayload is a struct that represents the action call
type CallActionPayload struct {
	Action string         `json:"action"`
	DBID   string         `json:"dbid"`
	Params map[string]any `json:"params"`
}

func (c *CallActionPayload) Bytes() ([]byte, error) {
	return json.Marshal(c)
}

type JsonPayload []byte

func (j JsonPayload) Bytes() ([]byte, error) {
	return j, nil
}

// ExecuteActionPayload is a struct that represents the action execution
type ExecuteActionPayload struct {
	Action string           `json:"action"`
	DBID   string           `json:"dbid"`
	Params []map[string]any `json:"params"`
}
