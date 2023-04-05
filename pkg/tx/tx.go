package tx

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	kwilCrypto "kwil/pkg/crypto"
)

type Transaction struct {
	Hash        []byte               `json:"hash"`
	PayloadType PayloadType          `json:"payload_type"`
	Payload     []byte               `json:"payload"`
	Fee         string               `json:"fee"`
	Nonce       int64                `json:"nonce"`
	Signature   kwilCrypto.Signature `json:"signature"`
	Sender      string               `json:"sender"`
}

func NewTx(txType PayloadType, data []byte, nonce int64) *Transaction {
	return &Transaction{
		PayloadType: txType,
		Payload:     data,
		Fee:         "0",
		Nonce:       nonce,
	}
}

func (t *Transaction) Verify() error {
	if !bytes.Equal(t.Hash, t.generateHash()) {
		return fmt.Errorf("invalid hash. received %s, expected %s", hex.EncodeToString(t.Hash), hex.EncodeToString(t.generateHash()))
	}

	// verify valid payload type
	if t.PayloadType <= INVALID_PAYLOAD_TYPE || t.PayloadType >= END_PAYLOAD_TYPE {
		return fmt.Errorf("invalid payload type")
	}

	// Not returning this function directly since I want specific error messages.
	ok, err := kwilCrypto.CheckSignature(t.Sender, t.Signature, t.Hash)
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
	payloadHash := kwilCrypto.Sha384(t.Payload)
	data = append(data, payloadHash...)

	// add fee
	data = append(data, []byte(t.Fee)...)

	// convert nonce to bytes
	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, uint64(t.Nonce))
	data = append(data, nonceBytes...)

	return kwilCrypto.Sha384(data)
}

func (t *Transaction) Sign(p *ecdsa.PrivateKey) error {
	hash := t.generateHash()
	sig, err := kwilCrypto.Sign(hash, p)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	address := kwilCrypto.AddressFromPrivateKey(p)

	t.Hash = hash
	t.Signature = sig
	t.Sender = address

	return nil
}

type Receipt struct {
	Hash []byte `json:"hash"`
	Fee  string `json:"fee"`
	Body []byte `json:"body"`
}
