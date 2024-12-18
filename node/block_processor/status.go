package blockprocessor

import (
	"slices"
	"time"

	ktypes "github.com/kwilteam/kwil-db/core/types"
)

type blockExecStatus struct {
	startTime, endTime time.Time
	height             int64
	txIDs              []ktypes.Hash
	txStatus           map[string]bool
}

// Used by the rpc server to get the execution status of the block being processed.
// end_time is not set if the block is still being processed.
func (bp *BlockProcessor) BlockExecutionStatus() *ktypes.BlockExecutionStatus {
	bp.statusMu.RLock()
	defer bp.statusMu.RUnlock()

	if bp.status == nil {
		return nil
	}

	status := &ktypes.BlockExecutionStatus{
		StartTime: bp.status.startTime,
		EndTime:   bp.status.endTime,
		Height:    bp.status.height,
		TxIDs:     slices.Clone(bp.status.txIDs),
		TxStatus:  make(map[string]bool),
	}

	for k, v := range bp.status.txStatus {
		status.TxStatus[k] = v
	}

	return status
}

func (bp *BlockProcessor) initBlockExecutionStatus(blk *ktypes.Block) {
	bp.statusMu.Lock()
	defer bp.statusMu.Unlock()

	status := &blockExecStatus{
		startTime: time.Now(),
		height:    blk.Header.Height,
		txStatus:  make(map[string]bool),
		txIDs:     make([]ktypes.Hash, len(blk.Txns)),
	}

	for i, tx := range blk.Txns {
		txID := ktypes.HashBytes(tx)
		status.txIDs[i] = txID
		status.txStatus[txID.String()] = false // not needed, just for clarity
	}

	bp.status = status
}

func (bp *BlockProcessor) clearBlockExecutionStatus() {
	bp.statusMu.Lock()
	defer bp.statusMu.Unlock()

	bp.status = nil
}

func (bp *BlockProcessor) updateBlockExecutionStatus(txID ktypes.Hash) {
	bp.statusMu.Lock()
	defer bp.statusMu.Unlock()

	bp.status.txStatus[txID.String()] = true
}

func (bp *BlockProcessor) recordBlockExecEndTime() {
	bp.statusMu.Lock()
	defer bp.statusMu.Unlock()

	bp.status.endTime = time.Now()
}
