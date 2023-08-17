package transactions

import (
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/serialize"
	"github.com/kwilteam/kwil-db/pkg/utils/random"
)

// CreateTransaction creates a new unsigned transaction.
func CreateTransaction(contents Payload, nonce uint64) (*Transaction, error) {
	data, err := contents.MarshalBinary()
	if err != nil {
		return nil, err
	}

	salt, err := generateRandomSalt()
	if err != nil {
		return nil, err
	}

	return &Transaction{
		Body: &TransactionBody{
			Payload:     data,
			PayloadType: contents.Type(),
			Fee:         big.NewInt(0),
			Nonce:       nonce,
			Salt:        salt[:],
		},
	}, nil
}

type Transaction struct {
	// Signature is the signature of the transaction
	// It can be nil if the transaction is unsigned
	Signature *crypto.Signature

	// Body is the body of the transaction
	// It gets serialized and signed
	Body *TransactionBody

	// Sender is the public key of the sender
	// It is not included in the signature
	Sender []byte

	// hash of the transaction that is signed.  it is kept here as a cache
	hash []byte
}

func (t *Transaction) GetSenderPubKey() (crypto.PublicKey, error) {
	return crypto.PublicKeyFromBytes(t.Signature.KeyType(), t.Sender)
}

// Verify verifies the signature of the transaction
func (t *Transaction) Verify() error {
	data, err := t.Body.MarshalBinary()
	if err != nil {
		return err
	}

	var pubKey crypto.PublicKey
	pubKey, err = crypto.PublicKeyFromBytes(t.Signature.KeyType(), t.Sender)
	if err != nil {
		return err
	}

	return t.Signature.Verify(pubKey, data)
}

func (t *Transaction) Sign(signer crypto.Signer) error {
	data, err := t.Body.MarshalBinary()
	if err != nil {
		return err
	}

	signature, err := signer.SignMsg(data)
	if err != nil {
		return err
	}

	t.Signature = signature
	t.Sender = signer.PubKey().Bytes()

	return nil
}

// GetHash gets the hash for the transaction
// If a hash has already been generated, it is returned
func (t *Transaction) GetHash() ([]byte, error) {
	if t.hash != nil {
		return t.hash, nil
	}

	return t.SetHash()
}

// SetHash re-hashes the transaction and caches the new hash
func (t *Transaction) SetHash() ([]byte, error) {
	bts, err := t.Body.MarshalBinary()
	if err != nil {
		return nil, err
	}

	t.hash = crypto.Sha256(bts)

	return t.hash, nil
}

func (t *Transaction) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(t)
}

func (t *Transaction) UnmarshalBinary(data serialize.SerializedData) error {
	res, err := serialize.Decode[Transaction](data)
	if err != nil {
		return err
	}

	*t = *res
	return nil
}

// TransactionBody is the body of a transaction that gets included in the signature
type TransactionBody struct {
	// Payload are the raw bytes of the payload data
	Payload serialize.SerializedData

	// PayloadType is the type of the payload
	// This can be used to determine how to decode the payload
	PayloadType PayloadType

	// Fee is the fee the sender is willing to pay for the transaction
	Fee *big.Int

	// Nonce is the next nonce of the sender
	Nonce uint64

	// Salt is a random value that is used to prevent replay attacks and hash collisions
	Salt []byte
}

func (t *TransactionBody) Verify() error {
	if !t.PayloadType.Valid() {
		return fmt.Errorf("invalid payload type: %s", t.PayloadType)
	}

	if t.Fee == nil {
		t.Fee = big.NewInt(0)
	}

	return nil
}

func (t *TransactionBody) MarshalBinary() ([]byte, error) {
	return serialize.Encode(t)
}

// generateRandomSalt generates a new random salt
// this salt is not used for any sort of security purpose;
// rather, it is just to prevent hash collisions
// therefore, we only need a small amount of entropy
func generateRandomSalt() ([8]byte, error) {
	var s [8]byte

	_, err := random.New().Read(s[:])
	if err != nil {
		return s, err
	}
	return s, nil
}

// TransactionStatus is used to show the status of a transaction
// It is returned to the client after a tx is submitted,
// and can also be used for querying the status of a tx
type TransactionStatus struct {
	ID     []byte
	Fee    *big.Int
	Status Status
	Errors []string
}

// Status is the status of a transaction
type Status string

func (s Status) String() string {
	return string(s)
}

// we can add / modify these as needed
const (
	StatusPending Status = "pending"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)
