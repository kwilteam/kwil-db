package crypto

import "fmt"

type Tx struct {
	Id        string
	Payload   []byte
	Fee       string
	Nonce     string
	Signature string
	Sender    string
}

// An interface for all transaction types.
// This is primarily for converting the TxMsg protobuf message into a Tx type.
type TxMsg interface {
	GetId() string
	GetPayload() []byte
	GetFee() string
	GetNonce() string
	GetSignature() string
	GetSender() string
}

// Convert is meant to convert other types of transactions into a Tx type.
// For example, it will convert the TxMsg protobuf message into a Tx type.
func (t *Tx) Convert(txmsg TxMsg) {
	t.Id = txmsg.GetId()
	t.Payload = txmsg.GetPayload()
	t.Fee = txmsg.GetFee()
	t.Nonce = txmsg.GetNonce()
	t.Signature = txmsg.GetSignature()
	t.Sender = txmsg.GetSender()
}

func (t *Tx) Verify() error {
	if t.Id != t.GenerateId() {
		return fmt.Errorf("invalid id")
	}

	// Not returning this function directly since I want specific error messages.
	ok, err := CheckSignature(t.Sender, t.Signature, []byte(t.Id))
	if err != nil {
		return fmt.Errorf("unexpected error checking signature: %v", err)
	}
	if !ok {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (t *Tx) GenerateId() string {
	// hash payload
	payloadHash := Sha384Str(t.Payload)

	var data []byte
	data = append(data, []byte(payloadHash)...)
	data = append(data, []byte(t.Fee)...)
	data = append(data, []byte(t.Nonce)...)
	data = append(data, []byte(t.Sender)...)
	return Sha384Str(data)
}
