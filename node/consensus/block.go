package consensus

import (
	"context"
	"errors"
	"fmt"
	"time"

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

	// Verify the current validator set for the block
	valSetHash := ce.validatorSetHash()
	if valSetHash != blk.Header.ValidatorSetHash {
		return fmt.Errorf("validator set hash mismatch, expected %s, got %s", valSetHash.String(), blk.Header.ValidatorSetHash.String())
	}

	return nil
}

func (ce *ConsensusEngine) CheckTx(ctx context.Context, tx *ktypes.Transaction) error {
	ce.stateInfo.mtx.RLock()
	height, timestamp := ce.stateInfo.height, ce.stateInfo.lastCommit.blk.Header.Timestamp
	ce.stateInfo.mtx.RUnlock()

	ce.mempoolMtx.Lock()
	defer ce.mempoolMtx.Unlock()

	return ce.blockProcessor.CheckTx(ctx, tx, height, timestamp, false)
}

func (ce *ConsensusEngine) recheckTx(ctx context.Context, tx *ktypes.Transaction) error {
	ce.stateInfo.mtx.RLock()
	height, timestamp := ce.stateInfo.height, ce.stateInfo.lastCommit.blk.Header.Timestamp
	ce.stateInfo.mtx.RUnlock()

	return ce.blockProcessor.CheckTx(ctx, tx, height, timestamp, true)
}

// BroadcastTx checks the Tx with the mempool and if the verification is successful, broadcasts the Tx to the network.
// If sync is set to 1, the BroadcastTx returns only after the Tx is successfully committed in a block.
func (ce *ConsensusEngine) BroadcastTx(ctx context.Context, tx *ktypes.Transaction, sync uint8) (*ktypes.ResultBroadcastTx, error) {
	if err := ce.CheckTx(ctx, tx); err != nil {
		return nil, err
	}

	rawTx := tx.Bytes()
	txHash := types.HashBytes(rawTx)

	// add the transaction to the mempool
	ce.mempool.Store(txHash, tx)

	// Announce the transaction to the network
	if ce.txAnnouncer != nil {
		ce.log.Infof("broadcasting new tx %v", txHash)
		go ce.txAnnouncer(context.Background(), txHash, rawTx)
	}

	res := &ktypes.ResultBroadcastTx{
		Hash: txHash, // Code and Log are set only if sync is set to 1
	}

	// If sync is set to 1, wait for the transaction to be committed in a block.
	if sync == 1 { // Blocking code
		subChan, err := ce.SubscribeTx(txHash)
		if err != nil {
			return nil, err
		}
		defer ce.UnsubscribeTx(txHash) // Unsubscribe tx if BroadcastTx returns

		select {
		case txRes := <-subChan:
			return &ktypes.ResultBroadcastTx{
				Code: txRes.Code,
				Hash: txHash,
				Log:  txRes.Log,
			}, nil
		case <-ctx.Done():
			return res, ctx.Err()
		case <-time.After(ce.broadcastTxTimeout):
			return res, errors.New("timed out waiting for tx to be included in a block")
		}
	}

	return res, nil
}

func (ce *ConsensusEngine) ConsensusParams() *ktypes.NetworkParameters {
	return ce.blockProcessor.ConsensusParams()
}

// executeBlock uses the block processor to execute the block and stores the
// results in the state field.
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
		Proposer: ce.leader,
	}
	results, err := ce.blockProcessor.ExecuteBlock(ctx, req)
	if err != nil {
		ce.log.Warn("Error executing block", "height", blkProp.height, "hash", blkProp.blkHash, "error", err)
		return errors.Join(fmt.Errorf("Error executing block: height %d, hash: %s", blkProp.height, blkProp.blkHash.String()), err)
	}

	ce.state.blockRes = &blockResult{
		ack:       true,
		appHash:   results.AppHash,
		txResults: results.TxResults,
		// vote is set in processBlockProposal
		paramUpdates: results.ParamUpdates,
	}

	// reset the catchup timer as we have successfully processed a new block proposal
	ce.catchupTicker.Reset(ce.catchupTimeout)

	ce.log.Info("Executed block", "height", blkProp.height, "blkID", blkProp.blkHash, "numTxs", blkProp.blk.Header.NumTxns, "appHash", results.AppHash.String())
	return nil
}

// Commit method commits the block to the blockstore and postgres database.
// It also updates the txIndexer and mempool with the transactions in the block.
func (ce *ConsensusEngine) commit(ctx context.Context) error {
	ce.mempoolMtx.Lock()
	defer ce.mempoolMtx.Unlock()

	blkProp := ce.state.blkProp
	height, appHash := ce.state.blkProp.height, ce.state.blockRes.appHash

	if err := ce.blockStore.Store(blkProp.blk, ce.state.commitInfo); err != nil {
		return err
	}

	if err := ce.blockStore.StoreResults(blkProp.blkHash, ce.state.blockRes.txResults); err != nil {
		return err
	}

	req := &ktypes.CommitRequest{
		Height:  height,
		AppHash: appHash,
		// To indicate if the node is syncing, used by the blockprocessor to decide if it should create snapshots.
		Syncing: ce.InCatchup(),
	}
	if err := ce.blockProcessor.Commit(ctx, req); err != nil { // clears the mempool cache
		return err
	}

	// remove transactions from the mempool
	for idx, txn := range blkProp.blk.Txns {
		txHash := txn.Hash()
		ce.mempool.Remove(txHash)

		txRes := ce.state.blockRes.txResults[idx]
		subChan, ok := ce.txSubscribers[txHash]
		if ok { // Notify the subscribers about the transaction result
			subChan <- txRes
		}
	}

	// recheck the transactions in the mempool
	ce.mempool.RecheckTxs(ctx, ce.recheckTx)

	// update the role of the node based on the final validator set at the end of the commit.
	ce.updateValidatorSetAndRole()

	// reset the catchup timer as we have successfully processed a new block proposal
	ce.catchupTicker.Reset(ce.catchupTimeout)

	ce.log.Info("Committed Block", "height", height, "hash", blkProp.blkHash.String(),
		"appHash", appHash.String(), "updates", ce.state.blockRes.paramUpdates)
	return nil
}

func (ce *ConsensusEngine) nextState() {
	ce.state.lc = &lastCommit{
		height:     ce.state.blkProp.height,
		blkHash:    ce.state.blkProp.blkHash,
		appHash:    ce.state.blockRes.appHash,
		blk:        ce.state.blkProp.blk,
		commitInfo: ce.state.commitInfo,
	}

	ce.resetState()
}

func (ce *ConsensusEngine) rollbackState(ctx context.Context) error {
	// Revert back any state changes occurred due to the current block
	if err := ce.blockProcessor.Rollback(ctx, ce.state.lc.height, ce.state.lc.appHash); err != nil {
		return err
	}

	ce.resetState()
	return nil
}

func (ce *ConsensusEngine) resetState() {
	ce.state.blkProp = nil
	ce.state.blockRes = nil
	ce.state.votes = make(map[string]*types.VoteInfo)
	ce.state.commitInfo = nil

	// update the stateInfo
	ce.stateInfo.mtx.Lock()
	ce.stateInfo.status = Committed
	ce.stateInfo.blkProp = nil
	ce.stateInfo.height = ce.state.lc.height
	ce.stateInfo.lastCommit = *ce.state.lc
	ce.stateInfo.mtx.Unlock()

	ce.cancelFnMtx.Lock()
	ce.blkExecCancelFn = nil
	ce.longRunningTxs = make([]ktypes.Hash, 0)
	ce.cancelFnMtx.Unlock()
}
