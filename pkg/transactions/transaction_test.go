package transactions_test

import (
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/transactions"
	"github.com/stretchr/testify/require"
)

// testing serialization of a transaction, since Luke found a bug
func Test_TransactionMarshal(t *testing.T) {
	tx := &transactions.Transaction{
		Signature: &crypto.Signature{
			Signature: []byte("signature"),
			Type:      crypto.SignatureTypeSecp256k1Cometbft,
		},
		Body: &transactions.TransactionBody{
			Payload:     []byte("payload"),
			PayloadType: transactions.PayloadTypeDeploySchema,
			Fee:         big.NewInt(100),
			Nonce:       1,
			Salt:        []byte("salt"),
		},
		Sender: []byte("sender"),
	}

	serialized, err := tx.MarshalBinary()
	require.NoError(t, err)

	tx2 := &transactions.Transaction{}
	err = tx2.UnmarshalBinary(serialized)
	require.NoError(t, err)

	require.Equal(t, tx, tx2)
}
