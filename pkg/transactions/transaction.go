package transactions

import (
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/serialize/rlp"
)

type Transaction struct {
	Signature *crypto.Signature
	Body      *TransactionBody
	Sender    crypto.PublicKey
}

// Verify verifies the signature of the transaction
func (t *Transaction) Verify() error {
	data, err := t.Body.MarshalBinary()
	if err != nil {
		return err
	}

	return t.Sender.Verify(t.Signature, data)
}

// TransactionBody is the body of a transaction that gets included in the signature
type TransactionBody struct {
	// Payload are the raw bytes of the payload data
	Payload rlp.SerializedData

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

func (t *TransactionBody) MarshalBinary() ([]byte, error) {
	return rlp.Encode(t)
}
