package transactions

import (
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/serialize/rlp"
)

type Transaction struct {
	Signature *crypto.Signature
	Payload   *TransactionPayload
	Sender    crypto.PublicKey
}

// Verify verifies the signature of the transaction
func (t *Transaction) Verify() error {
	data, err := t.Payload.MarshalBinary()
	if err != nil {
		return err
	}

	return t.Sender.Verify(t.Signature, data)
}

// TransactionPayload is a generic payload type that can be used to send binary data
type TransactionPayload struct {
	// Payload are the raw bytes of the payload data
	Payload rlp.SerializedData

	// PayloadType is the type of the payload
	// This can be used to determine how to decode the payload
	PayloadType PayloadType

	// Fee is the fee the sender is willing to pay for the transaction
	Fee *big.Int

	// Nonce is the next nonce of the sender
	Nonce uint64

	// Salt is a random number used to prevent replay attacks, as well as help guarantee uniqueness
	Salt uint64
}

func (t *TransactionPayload) MarshalBinary() ([]byte, error) {
	return rlp.Encode(t)
}
