package transactions

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"strings"

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
		Serialization: TxSerFullRLP,
		Body: &TransactionBody{
			Payload:     data,
			PayloadType: contents.Type(),
			Fee:         big.NewInt(0),
			Nonce:       nonce,
			Salt:        salt[:],
		},
	}, nil
}

/* seems unworkable, but we'd ideally keep description off the chain
type DescribedTransaction struct {
	Serialization TxSerializationType
	TxDescription string

	Transaction
}
*/

type Transaction struct {
	// Signature is the signature of the transaction
	// It can be nil if the transaction is unsigned
	Signature *crypto.Signature

	Serialization TxSerializationType

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

func (t *Transaction) GetSenderAddress() string {
	var pubKey crypto.PublicKey
	pubKey, err := crypto.PublicKeyFromBytes(t.Signature.KeyType(), t.Sender)
	if err != nil {
		return "unknown"
	}

	return pubKey.Address().String()
}

// Verify verifies the signature of the transaction
func (t *Transaction) Verify() error {
	data, err := t.Body.SerializeMsg(t.Serialization)
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
	data, err := t.Body.SerializeMsg(t.Serialization)
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

// MarshalBinary serializes the entire transaction for efficient storage and
// transmission of the body on the blockchain and network.
func (t *Transaction) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(t)
}

func (t *Transaction) UnmarshalBinary(data serialize.SerializedData) error {
	return serialize.DecodeInto(data, t)
}

// TransactionBody is the body of a transaction that gets included in the signature
type TransactionBody struct {
	Description string // boo

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

// MarshalBinary is the full RLP serialization of the transaction body. This is
// for computation of the tx hash.
func (t *TransactionBody) MarshalBinary() ([]byte, error) {
	return serialize.Encode(t)
}

type TxSerializationType uint32

const (
	TxSerFullRLP TxSerializationType = iota
	TxSerReadableWithPayloadHash
)

const msgTmpl = `Kwil Signed message:

🖋️🖋️🖋️🖋️🖋️

Description: %s

Payload type: %s
Fee: %s
Nonce: %d
Token: %x
`

// SerializeMsg prepares a message for signing or verification using a certain
// message construction format. This is done since a Kwil transaction is foreign
// to wallets, and it is signed as a message, not a transaction that is native
// to the wallet. As such we define conventions for constructing user-friendly
// messages. The Kwil frontend SDKs much implement these serialization schemes.
func (t *TransactionBody) SerializeMsg(ser TxSerializationType) ([]byte, error) {
	switch ser {
	case TxSerFullRLP:
		return serialize.Encode(t)
	case TxSerReadableWithPayloadHash:
		// Make a human readable message with the payload only in RLP.
		// In this message scheme, the displayed "token" is a hash of the
		// payload.
		token := sha256.Sum256(t.Payload)
		msgStr := fmt.Sprintf(msgTmpl, t.Description, t.PayloadType.String(),
			t.Fee.String(), t.Nonce, token)
		return []byte(msgStr), nil

		// case TxSerEIP...
	}
	return nil, errors.New("invalid serialization type")
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

// TxHash is the hash of a transaction that could be used to query the transaction
type TxHash []byte

func (h TxHash) Hex() string {
	return strings.ToUpper(fmt.Sprintf("%x", h))
}
