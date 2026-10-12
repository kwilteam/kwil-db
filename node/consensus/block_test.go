package consensus

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ktypes "github.com/trufnetwork/kwil-db/core/types"
	"github.com/trufnetwork/kwil-db/node/mempool"
	"github.com/trufnetwork/kwil-db/node/types"
)

// panickingCheckBP is a block processor whose CheckTx panics.
type panickingCheckBP struct{ BlockProcessor }

func (panickingCheckBP) CheckTx(context.Context, *types.Tx, int64, time.Time, bool) error {
	panic("check failed")
}

// TestQueueTxRemovesTxWhenCheckPanics checks that a transaction whose check
// panics does not stay in the mempool. The JSON-RPC server recovers panics in
// its handlers, and a transaction left behind would be rechecked after every
// block and announced to every peer.
func TestQueueTxRemovesTxWhenCheckPanics(t *testing.T) {
	mp := mempool.New(1<<20, 1<<20)
	ce := &ConsensusEngine{mempool: mp, blockProcessor: panickingCheckBP{}}
	ce.role.Store(types.RoleSentry)

	tx := types.NewTx(&ktypes.Transaction{})
	require.Panics(t, func() {
		_ = ce.QueueTx(context.Background(), tx)
	})

	_, numTxns := mp.Size()
	require.Zero(t, numTxns)
	require.Empty(t, mp.PeekN(10, 1<<20))
}
