package consensus

import (
	"context"
	"fmt"

	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
)

// Block processing methods
func (ce *ConsensusEngine) validateBlock(blk *types.Block) error {
	// Validate if this is the correct block proposal to be processed.
	if blk.Header.Version != types.BlockVersion {
		return fmt.Errorf("block version mismatch, expected %d, got %d", types.BlockVersion, blk.Header.Version)
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

func (ce *ConsensusEngine) CheckTx(tx []byte) error {
	return nil
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
		return fmt.Errorf("error executing block: %v", err)
	}

	ce.state.blockRes = &blockResult{
		ack:       true,
		appHash:   results.AppHash,
		txResults: results.TxResults,
	}

	ce.log.Info("Executed block", "height", blkProp.height, "hash", blkProp.blkHash, "appHash", results.AppHash.String())
	return nil
}

// Commit method commits the block to the blockstore and postgres database.
// It also updates the txIndexer and mempool with the transactions in the block.
func (ce *ConsensusEngine) commit(ctx context.Context) error {
	// TODO: Lock mempool and update the mempool to remove the transactions in the block
	// Mempool should not receive any new transactions until this Commit is done as
	// we are updating the state and the tx checks should be done against the new state.
	blkProp := ce.state.blkProp
	height, appHash := ce.state.blkProp.height, ce.state.blockRes.appHash

	if err := ce.blockStore.Store(blkProp.blk, appHash); err != nil {
		return err
	}

	if err := ce.blockStore.StoreResults(blkProp.blkHash, ce.state.blockRes.txResults); err != nil {
		return err
	}

	if err := ce.blockProcessor.Commit(ctx, height, appHash, false); err != nil {
		return err
	}

	// remove transactions from the mempool
	for _, txn := range blkProp.blk.Txns {
		txHash := types.HashBytes(txn)
		ce.mempool.Store(txHash, nil)
	}

	// TODO: set the role based on the final validators

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
