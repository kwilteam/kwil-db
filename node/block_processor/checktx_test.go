package blockprocessor

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/trufnetwork/kwil-db/common"
	"github.com/trufnetwork/kwil-db/core/crypto"
	"github.com/trufnetwork/kwil-db/core/crypto/auth"
	"github.com/trufnetwork/kwil-db/core/log"
	"github.com/trufnetwork/kwil-db/core/types"
	nodetypes "github.com/trufnetwork/kwil-db/node/types"
)

// TestCheckTxRejectsIncompleteTransactions sends checkTx transactions that
// decode without a body, a fee or a signature, as a peer or an RPC client can
// send them. Each one must be rejected before anything reads the missing part.
func TestCheckTxRejectsIncompleteTransactions(t *testing.T) {
	bp := &BlockProcessor{
		chainCtx: &common.ChainContext{NetworkParameters: &types.NetworkParameters{}},
		log:      log.DiscardLogger,
	}

	privKey, _, err := crypto.GenerateSecp256k1Key(nil)
	require.NoError(t, err)
	signer := auth.GetUserSigner(privKey)

	newBody := func() *types.TransactionBody {
		return &types.TransactionBody{
			Payload:     []byte("payload"),
			PayloadType: types.PayloadTypeExecute,
			Fee:         big.NewInt(0),
			Nonce:       1,
		}
	}

	tests := []struct {
		name string
		tx   func(t *testing.T) *types.Transaction
	}{
		{
			name: "no body and no signature",
			tx: func(t *testing.T) *types.Transaction {
				return &types.Transaction{}
			},
		},
		{
			name: "signed body without a fee",
			tx: func(t *testing.T) *types.Transaction {
				tx := &types.Transaction{Body: newBody(), Serialization: types.SignedMsgConcat}
				tx.Body.Fee = nil
				require.NoError(t, tx.Sign(signer))
				return tx
			},
		},
		{
			name: "body without a signature",
			tx: func(t *testing.T) *types.Transaction {
				return &types.Transaction{
					Body:          newBody(),
					Serialization: types.SignedMsgConcat,
					Sender:        signer.CompactID(),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Decode the transaction from its serialized form, the way a node
			// receives it from a peer.
			var tx types.Transaction
			require.NoError(t, tx.UnmarshalBinary(tt.tx(t).Bytes()))

			for _, recheck := range []bool{false, true} {
				err := bp.checkTx(context.Background(), nil, nodetypes.NewTx(&tx), 1, time.Now(), recheck)
				require.ErrorIs(t, err, types.ErrMalformedTx, "recheck %v", recheck)
			}
		})
	}
}
