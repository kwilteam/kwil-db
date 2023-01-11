package transactions

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"kwil/x/crypto"
	"kwil/x/transactions"
)

type Transaction struct {
	Hash        []byte                   `json:"hash"`
	PayloadType transactions.PayloadType `json:"payload_type"`
	Payload     []byte                   `json:"payload"`
	Fee         string                   `json:"fee"`
	Nonce       int64                    `json:"nonce"`
	Signature   crypto.Signature         `json:"signature"`
	Sender      string                   `json:"sender"`
}

func NewTx(txType transactions.PayloadType, data []byte, nonce int64) *Transaction {
	return &Transaction{
		PayloadType: txType,
		Payload:     data,
		Fee:         "0",
		Nonce:       nonce,
	}
}

func (t *Transaction) Verify() error {
	if !bytes.Equal(t.Hash, t.generateHash()) {
		return fmt.Errorf("invalid hash")
	}

	// verify valid payload type
	if t.PayloadType <= transactions.INVALID_PAYLOAD_TYPE || t.PayloadType >= transactions.END_PAYLOAD_TYPE {
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

// generateHash generates a hash of the transaction
// it does this by hashing the payload type, payload, fee, and nonce
func (t *Transaction) generateHash() []byte {
	var data []byte

	// convert payload type to bytes
	payloadTypeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(payloadTypeBytes, uint32(t.PayloadType))
	data = append(data, payloadTypeBytes...)

	// hash payload
	payloadHash := crypto.Sha384(t.Payload)
	data = append(data, payloadHash...)

	// add fee
	data = append(data, []byte(t.Fee)...)

	// convert nonce to bytes
	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, uint64(t.Nonce))
	data = append(data, nonceBytes...)

	return crypto.Sha384(data)
}

func (t *Transaction) Sign(p *ecdsa.PrivateKey) error {
	hash := t.generateHash()
	sig, err := crypto.Sign(hash, p)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	address, err := crypto.AddressFromPrivateKey(p)
	if err != nil {
		return fmt.Errorf("failed to get address from private key: %v", err)
	}

	t.Hash = hash
	t.Signature = sig
	t.Sender = address

	return nil
}
