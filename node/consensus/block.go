package consensus

import (
	"context"
	"fmt"

	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
)

// TODO: should include consensus params hash
func (ce *ConsensusEngine) validateBlock(blk *ktypes.Block) error {
	// Validate if this is the correct block proposal to be processed.
	if blk.Header.Version != ktypes.BlockVersion {
		return fmt.Errorf("block version mismatch, expected %d, got %d", ktypes.BlockVersion, blk.Header.Version)
	}

	if ce.state.lc.height+1 != blk.Header.Height {
		return fmt.Errorf("block proposal for height %d does not follow %d", blk.Header.Height, ce.state.lc.height)
	}

	if ce.state.lc.blkHash != blk.Header.PrevHash {
		return fmt.Errorf("prevBlockHash mismatch, expected %v, got %v", ce.state.lc.blkHash, blk.Header.PrevHash)
	}

	if blk.Header.PrevAppHash != ce.state.lc.appHash {
		return fmt.Errorf("apphash mismatch, expected %v, got %v", ce.state.lc.appHash, blk.Header.PrevAppHash)
	}

	if blk.Header.NumTxns != uint32(len(blk.Txns)) {
		return fmt.Errorf("transaction count mismatch, expected %d, got %d", blk.Header.NumTxns, len(blk.Txns))
	}

	// Verify the merkle root of the block transactions
	merkleRoot := blk.MerkleRoot()
	if merkleRoot != blk.Header.MerkleRoot {
		return fmt.Errorf("merkleroot mismatch, expected %v, got %v", merkleRoot, blk.Header.MerkleRoot)
	}

	// Verify other stuff such as validatorsetHash, signature of the block etc.
	return nil
}

func (ce *ConsensusEngine) CheckTx(ctx context.Context, tx []byte) error {
	ce.mempoolMtx.Lock()
	defer ce.mempoolMtx.Unlock()

	return ce.blockProcessor.CheckTx(ctx, tx, false)
}

func (ce *ConsensusEngine) executeBlock(ctx context.Context, blkProp *blockProposal) error {
	defer func() {
		ce.stateInfo.mtx.Lock()
		ce.stateInfo.status = Executed
		ce.stateInfo.mtx.Unlock()
	}()

	req := &ktypes.BlockExecRequest{
		Block:    blkProp.blk,
		Height:   blkProp.height,
		BlockID:  blkProp.blkHash,
		Proposer: ce.leader.Bytes(),
	}
	results, err := ce.blockProcessor.ExecuteBlock(ctx, req)
	if err != nil {
		ce.log.Warn("Error executing block", "height", blkProp.height, "hash", blkProp.blkHash, "error", err)
		return fmt.Errorf("error executing block: %v", err)
	}

	ce.state.blockRes = &blockResult{
		ack:       true,
		appHash:   results.AppHash,
		txResults: results.TxResults,
	}

	ce.log.Info("Executed block", "height", blkProp.height, "hash", blkProp.blkHash, "numTxs", blkProp.blk.Header.NumTxns, "appHash", results.AppHash.String())
	return nil
}

// Commit method commits the block to the blockstore and postgres database.
// It also updates the txIndexer and mempool with the transactions in the block.
func (ce *ConsensusEngine) commit(ctx context.Context) error {
	ce.mempoolMtx.Lock()
	defer ce.mempoolMtx.Unlock()

	blkProp := ce.state.blkProp
	height, appHash := ce.state.blkProp.height, ce.state.blockRes.appHash

	if err := ce.blockStore.Store(blkProp.blk, appHash); err != nil {
		return err
	}

	if err := ce.blockStore.StoreResults(blkProp.blkHash, ce.state.blockRes.txResults); err != nil {
		return err
	}

	req := &ktypes.CommitRequest{
		Height:  height,
		AppHash: appHash,
		// To indicate if the node is syncing, used by the blockprocessor to decide if it should create snapshots.
		Syncing: ce.inSync.Load(),
	}
	if err := ce.blockProcessor.Commit(ctx, req); err != nil { // clears the mempool cache
		return err
	}

	// remove transactions from the mempool
	for _, txn := range blkProp.blk.Txns {
		txHash := types.HashBytes(txn) // TODO: can this be saved instead of recalculating?
		ce.mempool.Store(txHash, nil)
	}

	// TODO: reapply existing transaction  (checkTX)
	// get all the transactions from mempool and recheck them, the transactions should be checked
	// in the order of nonce (stable sort to maintain relative order)
	// ce.blockProcessor.CheckTx(ctx, tx, true)

	// update the role of the node based on the final validator set at the end of the commit.
	ce.updateRole()

	ce.log.Info("Committed Block", "height", height, "hash", blkProp.blkHash, "appHash", appHash.String())
	return nil
}

func (ce *ConsensusEngine) nextState() {
	ce.state.lc = &lastCommit{
		height:  ce.state.blkProp.height,
		blkHash: ce.state.blkProp.blkHash,
		appHash: ce.state.blockRes.appHash,
		blk:     ce.state.blkProp.blk,
	}

	ce.state.blkProp = nil
	ce.state.blockRes = nil
	ce.state.votes = make(map[string]*vote)
	ce.state.consensusTx = nil

	// update the stateInfo
	ce.stateInfo.mtx.Lock()
	ce.stateInfo.status = Committed
	ce.stateInfo.blkProp = nil
	ce.stateInfo.height = ce.state.lc.height
	ce.stateInfo.mtx.Unlock()
}

func (ce *ConsensusEngine) resetState(ctx context.Context) error {
	// Revert back any state changes occurred due to the current block
	if err := ce.blockProcessor.Rollback(ctx, ce.state.lc.height, ce.state.lc.appHash); err != nil {
		return err
	}

	ce.state.blkProp = nil
	ce.state.blockRes = nil
	ce.state.votes = make(map[string]*vote)
	ce.state.consensusTx = nil

	// update the stateInfo
	ce.stateInfo.mtx.Lock()
	ce.stateInfo.status = Committed
	ce.stateInfo.blkProp = nil
	ce.stateInfo.height = ce.state.lc.height
	ce.stateInfo.mtx.Unlock()

	return nil
}
