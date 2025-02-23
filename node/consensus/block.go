package consensus

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
)

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

	// network params hash
	if blk.Header.NetworkParamsHash != ce.blockProcessor.ConsensusParams().Hash() {
		return fmt.Errorf("network params hash mismatch, expected %s, got %s", ce.blockProcessor.ConsensusParams().Hash().String(), blk.Header.NetworkParamsHash.String())
	}

	// Ensure that if any leader update is present, it is valid
	if blk.Header.NewLeader != nil {
		candidate := hex.EncodeToString(blk.Header.NewLeader.Bytes())
		if _, ok := ce.validatorSet[candidate]; !ok {
			return fmt.Errorf("leader update candidate %s is not a validator", candidate)
		}
	}

	maxBlockSize := ce.ConsensusParams().MaxBlockSize
	if blockTxnsSize := blk.TxnsSize(); blockTxnsSize > maxBlockSize {
		return fmt.Errorf("block size %d exceeds max block size %d", blockTxnsSize, maxBlockSize)
	}

	// Ensure that the number of event and resolution IDs within validator vote transactions votes
	// per transaction does not exceed the max consensus limit.
	maxVotesPerTx := ce.ConsensusParams().MaxVotesPerTx
	for _, txn := range blk.Txns {
		if txn.Body.PayloadType == ktypes.PayloadTypeValidatorVoteBodies {
			// unmarshal the payload
			vote := &ktypes.ValidatorVoteBodies{}
			err := vote.UnmarshalBinary(txn.Body.Payload)
			if err != nil {
				return fmt.Errorf("failed to unmarshal validator vote body: %v", err)
			}

			if int64(len(vote.Events)) > maxVotesPerTx {
				return fmt.Errorf("max votes exceeded in tx of type %s : %d > %d", txn.Body.PayloadType, len(vote.Events), maxVotesPerTx)
			}

		} else if txn.Body.PayloadType == ktypes.PayloadTypeValidatorVoteIDs {
			// unmarshal the payload
			vote := &ktypes.ValidatorVoteIDs{}
			err := vote.UnmarshalBinary(txn.Body.Payload)
			if err != nil {
				return fmt.Errorf("failed to unmarshal validator vote id: %v", err)
			}

			if int64(len(vote.ResolutionIDs)) > maxVotesPerTx {
				return fmt.Errorf("max votes exceeded in tx of type %s : %d > %d", txn.Body.PayloadType, len(vote.ResolutionIDs), maxVotesPerTx)
			}
		}
	}

	return nil
}

func (ce *ConsensusEngine) CheckTx(ctx context.Context, tx *ktypes.Transaction) error {
	ce.stateInfo.mtx.RLock()
	height := ce.stateInfo.height
	var timestamp time.Time
	if ce.stateInfo.lastCommit.blk != nil {
		timestamp = ce.stateInfo.lastCommit.blk.Header.Timestamp
	}
	ce.stateInfo.mtx.RUnlock()

	ce.mempoolMtx.Lock()
	defer ce.mempoolMtx.Unlock()

	return ce.blockProcessor.CheckTx(ctx, tx, height, timestamp, false)
}

func (ce *ConsensusEngine) recheckTx(ctx context.Context, tx *ktypes.Transaction) error {
	ce.stateInfo.mtx.RLock()
	height := ce.stateInfo.height
	var timestamp time.Time
	if ce.stateInfo.lastCommit.blk != nil {
		timestamp = ce.stateInfo.lastCommit.blk.Header.Timestamp
	}
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
	have, rejected := ce.mempool.Store(txHash, tx)
	if rejected {
		return &ktypes.ResultBroadcastTx{
			Hash: txHash,
		}, ktypes.ErrMempoolFull
	}

	// Announce the transaction to the network only if not previously announced
	if ce.txAnnouncer != nil && !have {
		// We can't use parent context 'cause it's canceled in the caller, which
		// could be the RPC request. handler.  This shouldn't be CE's problem...
		go ce.txAnnouncer(context.Background(), txHash, rawTx)
	}

	res := &ktypes.ResultBroadcastTx{
		Hash: txHash, // Code and Log are set only if sync is set to 1
	}

	// If sync is set to 1, wait for the transaction to be committed in a block.
	if sync == 1 { // Blocking code
		subChan, err := ce.SubscribeTx(txHash)
		if err != nil {
			return &ktypes.ResultBroadcastTx{
				Hash: txHash,
			}, err
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
		return err
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

	ce.log.Info("Executed block", "height", blkProp.height, "blockID", blkProp.blkHash, "numTxs", blkProp.blk.Header.NumTxns, "appHash", results.AppHash.String())
	return nil
}

// Commit method commits the block to the blockstore and postgres database.
// It also updates the txIndexer and mempool with the transactions in the block.
func (ce *ConsensusEngine) commit(ctx context.Context) error {
	ce.mempoolMtx.Lock()
	defer ce.mempoolMtx.Unlock()

	if ce.state.blockRes == nil {
		return errors.New("no block results to commit")
	}

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

	ce.log.Info("Committed Block", "height", height, "hash", blkProp.blkHash.String(),
		"appHash", appHash.String(), "numTxs", blkProp.blk.Header.NumTxns)

	// update and reset the state fields
	ce.nextState()

	// update the role of the node based on the final validator set at the end of the commit.
	ce.updateValidatorSetAndRole()

	// reset the catchup timer as we have successfully processed a new block proposal
	ce.catchupTicker.Reset(ce.catchupTimeout)

	return ctx.Err()
}

// nextState sets the lastCommit in state.lc from the current block proposal
// execution and commit results, resets the other state fields such as block
// proposal, block result, etc., and updates the status (stateInfo) to reflect
// the block that was just committed.
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
	ce.stateInfo.hasBlock.Store(ce.state.lc.height)

	return nil
}

func (ce *ConsensusEngine) resetState() {
	ce.state.blkProp = nil
	ce.state.blockRes = nil
	ce.state.votes = make(map[string]*types.VoteInfo)
	ce.state.commitInfo = nil
	if ce.state.leaderUpdate != nil {
		ce.state.leaderUpdate = nil
		ce.storeLeaderUpdates(nil) // clear the leader update once applied.
	}

	// update the stateInfo
	ce.stateInfo.mtx.Lock()
	ce.stateInfo.status = Committed
	ce.stateInfo.blkProp = nil
	ce.stateInfo.height = ce.state.lc.height
	ce.stateInfo.lastCommit = *ce.state.lc
	ce.stateInfo.mtx.Unlock()

	ce.stateInfo.hasBlock.Store(ce.state.lc.height)

	ce.cancelFnMtx.Lock()
	ce.blkExecCancelFn = nil
	ce.longRunningTxs = make([]ktypes.Hash, 0)
	ce.cancelFnMtx.Unlock()
}
