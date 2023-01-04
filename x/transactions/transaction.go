package dto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"kwil/x/crypto"
)

type Transaction struct {
	Hash        []byte      `json:"hash"`
	PayloadType PayloadType `json:"payload_type"`
	Payload     []byte      `json:"payload"`
	Fee         string      `json:"fee"`
	Nonce       int64       `json:"nonce"`
	Signature   string      `json:"signature"`
	Sender      string      `json:"from"`
}

// an interface for tx's sent over GRPC
type TxMsg interface {
	GetHash() []byte
	GetPayloadType() int32
	GetPayload() []byte
	GetFee() string
	GetNonce() int64
	GetSignature() string
	GetSender() string
}

func (t *Transaction) Convert(txmsg TxMsg) {
	t.Hash = txmsg.GetHash()
	t.Payload = txmsg.GetPayload()
	t.Fee = txmsg.GetFee()
	t.Nonce = txmsg.GetNonce()
	t.Signature = txmsg.GetSignature()
	t.Sender = txmsg.GetSender()
}

func (t *Transaction) Verify() error {
	if !bytes.Equal(t.Hash, t.GenerateHash()) {
		return fmt.Errorf("invalid hash")
	}

	// verify valid payload type
	if t.PayloadType <= INVALID_PAYLOAD_TYPE || t.PayloadType >= END_PAYLOAD_TYPE {
		return fmt.Errorf("invalid payload type")
	}

	// Not returning this function directly since I want specific error messages.
	ok, err := crypto.CheckSignature(t.Sender, t.Signature, t.Hash)
	if err != nil {
		return fmt.Errorf("unexpected error checking signature: %v", err)
	}
	if !ok {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (t *Transaction) GenerateHash() []byte {
	// hash payload
	payloadHash := crypto.Sha384(t.Payload)

	var data []byte
	data = append(data, payloadHash...)
	data = append(data, []byte(t.Fee)...)

	// convert payload type to bytes
	payloadTypeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(payloadTypeBytes, uint32(t.PayloadType))
	data = append(data, payloadTypeBytes...)

	// convert nonce to bytes
	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, uint64(t.Nonce))
	data = append(data, nonceBytes...)

	data = append(data, t.Signature...)
	data = append(data, []byte(t.Sender)...)

	return crypto.Sha384(data)
}
