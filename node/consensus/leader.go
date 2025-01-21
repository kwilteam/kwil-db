package consensus

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/kwilteam/kwil-db/config"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
)

// The Leader is responsible for proposing blocks and managing the consensus process:
// 1. Prepare Phase:
//   - Create a block proposal
//   - Broadcast the block proposal
//   - Process the block and generate the appHash
//   - Wait for votes from validators
//   - Enter the commit phase if the majority of validators approve the block
//
// 2. Commit Phase:
//   - Commit the block and initiate the next prepare phase
//   - This phase includes committing the block to the block store, clearing the mempool,
//     updating the chain state, creating snapshots, committing the pg db state, etc.
//
// The Leader can also issue ResetState messages using "kwild block reset <reset-block-height> <longrunning-tx-list>"
// When a leader receives a ResetState request, it will broadcast the ResetState message to the network to
// halt the current block execution if the current block height equals reset-block-height + 1. The leader will stop processing
// the current block, revert any state changes made, and remove the problematic transactions from the mempool before
// reproposing the block.

func (ce *ConsensusEngine) newBlockRound(ctx context.Context) error {
	ticker := time.NewTicker(ce.proposeTimeout)
	now := time.Now()

	// if EmptyBlockTimeout = 0, leader doesn't propose empty blocks.
	// Behavior is similar to automine feature where the blocks are produced
	// the moment transactions are available once the proposeTimeout is elapsed.
	// if EmptyBlockTimeout is not 0, leader will propose an empty block
	// if no transactions or events are available for emptyBlockTimeout duration.
	allowEmptyBlocks := ce.emptyBlockTimeout != 0
	ce.log.Info("Starting a new consensus round", "height", ce.lastCommitHeight()+1)

	for {
		select {
		case <-ctx.Done():
			ce.log.Warn("Context cancelled, stopping the new block round")
			return nil
		case <-ticker.C:
			// check for the availability of transactions in the mempool or
			// if the leader has any new events to broadcast a voteID transaction
			if ce.mempool.TxsAvailable() || ce.blockProcessor.HasEvents() {
				ce.newBlockProposal <- struct{}{}
				return nil
			}

			// If the emptyBlockTimeout duration has elapsed, produce an empty block if
			// empty blocks are allowed
			if allowEmptyBlocks && time.Since(now) >= ce.emptyBlockTimeout {
				ce.newBlockProposal <- struct{}{}
				return nil
			}
		}

		// no transactions available, wait till the next tick to recheck the mempool
	}
}

// proposeBlock used by the leader to propose a new block to the network.
func (ce *ConsensusEngine) proposeBlock(ctx context.Context) error {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()
	// Check if the network is halted due to migration or other reasons
	params := ce.blockProcessor.ConsensusParams()
	if params.MigrationStatus == ktypes.MigrationCompleted {
		haltReason := "Network is halted for migration, cannot start a new round"
		ce.log.Warn(haltReason)
		ce.haltChan <- haltReason // signal the network to halt
		return nil
	}

	blkProp, err := ce.createBlockProposal(ctx)
	if err != nil {
		ce.log.Errorf("Error creating a block proposal: %v", err)
		return err
	}

	ce.log.Info("Created a new block proposal", "height", blkProp.height, "hash", blkProp.blkHash)

	// Validate the block proposal before announcing it to the network
	if err := ce.validateBlock(blkProp.blk); err != nil {
		ce.log.Errorf("Error validating the block proposal: %v", err)
		return err
	}
	ce.state.blkProp = blkProp

	// Broadcast the block proposal to the network
	go ce.proposalBroadcaster(ctx, blkProp.blk)

	// update the stateInfo
	ce.stateInfo.mtx.Lock()
	ce.stateInfo.status = Proposed
	ce.stateInfo.blkProp = blkProp
	ce.stateInfo.mtx.Unlock()

	// execCtx is applicable only for the duration of the block execution
	// This is used to give leader the ability to cancel the block execution.
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set the cancel function for the block execution
	ce.cancelFnMtx.Lock()
	ce.blkExecCancelFn = cancel
	ce.cancelFnMtx.Unlock()

	// Execute the block and generate the appHash
	if err := ce.executeBlock(execCtx, blkProp); err != nil {
		// check if the error is due to context cancellation
		ce.log.Errorf("Error executing the block: %v", err)
		if execCtx.Err() != nil && errors.Is(err, context.Canceled) {
			ce.log.Warn("Block execution cancelled by the leader", "height", blkProp.height, "hash", blkProp.blkHash)
			ce.cancelFnMtx.Lock()
			// trigger a reset state message to the network
			go ce.rstStateBroadcaster(ce.state.lc.height, ce.longRunningTxs)

			// Remove the long running transactions from the mempool
			ce.log.Info("Removing long running transactions from the mempool as per leader's request", "txIDs", ce.longRunningTxs)
			for _, txID := range ce.longRunningTxs {
				ce.mempool.Remove(txID)
			}
			ce.numResets++
			ce.cancelFnMtx.Unlock()

			if err := ce.rollbackState(ctx); err != nil {
				ce.log.Errorf("Error resetting the state: %v", err)
				return fmt.Errorf("error resetting the state: %v", err)
			}

			// Recheck the transactions in the mempool
			ce.mempoolMtx.Lock()
			ce.mempool.RecheckTxs(ctx, ce.recheckTx)
			ce.mempoolMtx.Unlock()

			// signal ce to start a new round
			// ce.newRound <- struct{}{}
			// repropse a new block
			ce.newBlockProposal <- struct{}{}
			return nil
		}

		ce.log.Errorf("Error executing the block: %v", err)
		return err
	}

	// Add its own vote to the votes map
	sig, err := ktypes.SignVote(blkProp.blkHash, true, &ce.state.blockRes.appHash, ce.privKey)
	if err != nil {
		ce.log.Errorf("Error signing the vote: %v", err)
		return err
	}

	ce.log.Info("Waiting for votes from the validators", "height", blkProp.height, "hash", blkProp.blkHash)

	ce.state.votes[string(ce.pubKey.Bytes())] = &ktypes.VoteInfo{
		AppHash:   &ce.state.blockRes.appHash,
		AckStatus: ktypes.AckStatusAgree,
		Signature: *sig,
	}

	ce.processVotes(ctx)
	return nil
}

// Create Proposal creates a new block proposal for the leader
// by reaping the transactions from the mempool. This also adds the
// proposer transactions such as ValidatorVoteBodies.
// This method orders the transactions in the nonce order and also
// does basic gas and balance checks and enforces the block size limits.
func (ce *ConsensusEngine) createBlockProposal(ctx context.Context) (*blockProposal, error) {
	nTxs := ce.mempool.PeekN(blockTxCount)
	txns := make([]*ktypes.Transaction, len(nTxs))
	for i, ntx := range nTxs {
		txns[i] = ntx.Tx
	}

	finalTxs, invalidTxs, err := ce.blockProcessor.PrepareProposal(ctx, txns)
	if err != nil {
		ce.log.Errorf("Error preparing the block proposal: %v", err)
		return nil, err
	}

	// remove invalid transactions from the mempool
	for _, tx := range invalidTxs {
		txid := tx.Hash()
		ce.mempool.Remove(txid)
	}

	valSetHash := ce.validatorSetHash()
	stamp := time.Now().Truncate(time.Millisecond).UTC()
	blk := ktypes.NewBlock(ce.state.lc.height+1, ce.state.lc.blkHash, ce.state.lc.appHash, valSetHash, stamp, finalTxs)

	// Sign the block
	if err := blk.Sign(ce.privKey); err != nil {
		return nil, err
	}

	return &blockProposal{
		height:  blk.Header.Height,
		blkHash: blk.Header.Hash(),
		blk:     blk,
	}, nil
}

// addVote registers the vote received from the validator if it is for the current block.
func (ce *ConsensusEngine) addVote(ctx context.Context, vote *vote, sender string) error {
	// ce.log.Debugln("Adding vote", vote, sender)
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if ce.state.blkProp == nil {
		ce.log.Debug("Error adding vote: not processing any block proposal at the moment")
		return nil
	}

	// check if the vote is for the current height
	if ce.state.blkProp.height != vote.height {
		ce.log.Debug("Error adding vote: Vote received for a different block height, ignore it", "height", vote.height)
		return nil
	}

	// check if the vote is for the current block and from a validator
	if ce.state.blkProp.blkHash != vote.blkHash {
		ce.log.Warn("Error adding vote: Vote received for a different block", "height", vote.height, "blkHash", vote.blkHash)
		return nil
	}

	// Check if the vote is from a validator
	if _, ok := ce.validatorSet[sender]; !ok {
		return fmt.Errorf("vote received from an unknown validator %s", sender)
	}

	if vote.ack && vote.appHash == nil {
		return errors.New("missing appHash in the vote")
	}

	if vote.signature == nil {
		return errors.New("missing signature in the vote")
	}

	ce.log.Info("Adding vote", "height", vote.height, "blkHash", vote.blkHash, "appHash", vote.appHash, "sender", sender)
	if _, ok := ce.state.votes[sender]; !ok {
		// verify the vote signature, before accepting the vote from the validator if ack is true
		var ackStatus ktypes.AckStatus
		var appHash *types.Hash

		if vote.ack {
			ackStatus = ktypes.AckStatusAgree
			if *vote.appHash != ce.state.blockRes.appHash {
				ackStatus = ktypes.AckStatusDiverge
				appHash = vote.appHash
			}
		}

		voteInfo := &ktypes.VoteInfo{
			Signature: *vote.signature,
			AckStatus: ackStatus,
			AppHash:   appHash,
		}

		// verify signature
		if err := voteInfo.Verify(vote.blkHash, ce.state.blockRes.appHash); err != nil {
			ce.log.Errorf("Error verifying the vote signature: %v", err)
			return fmt.Errorf("error verifying the vote signature: %v", err)
		}

		ce.state.votes[sender] = voteInfo
	}

	ce.processVotes(ctx)
	return nil
}

// ProcessVotes processes the votes received from the validators.
// Depending on the votes, leader will trigger one of the following:
// 1. Commit the block
// 2. Re-announce the block proposal
// 3. Halt the network (should there be a message to halt the network?)
func (ce *ConsensusEngine) processVotes(ctx context.Context) error {
	ce.log.Debug("Processing votes", "height", ce.state.lc.height+1)

	if ce.state.blkProp == nil || ce.state.blockRes == nil {
		// Moved onto the next round or leader still processing the current block
		return nil
	}

	// Count the votes
	var acks, nacks int
	for _, vote := range ce.state.votes {
		if vote.AckStatus == ktypes.AckStatusAgree {
			acks++
		} else {
			nacks++
		}
	}

	if ce.hasMajorityCeil(acks) {
		ce.log.Info("Majority of the validators have accepted the block, proceeding to commit the block",
			"height", ce.state.blkProp.blk.Header.Height, "hash", ce.state.blkProp.blkHash, "acks", acks, "nacks", nacks)

		votes := make([]*ktypes.VoteInfo, 0)
		for _, v := range ce.state.votes {
			votes = append(votes, v)
		}

		ci := &ktypes.CommitInfo{
			AppHash:      ce.state.blockRes.appHash,
			Votes:        votes,
			ParamUpdates: ce.state.blockRes.paramUpdates,
		}
		ce.state.commitInfo = ci

		// Commit the block and broadcast the blockAnn message
		if err := ce.commit(ctx); err != nil {
			ce.log.Errorf("Error committing the block (process votes): %v", err)
			return err
		}

		ce.log.Infoln("Announce committed block", ce.state.blkProp.blk.Header.Height, ce.state.blkProp.blkHash,
			ce.state.blockRes.paramUpdates)
		// Broadcast the blockAnn message
		go ce.blkAnnouncer(ctx, ce.state.blkProp.blk, ce.state.commitInfo)

		// start the next round
		ce.nextState()

		// signal ce to start a new round
		ce.newRound <- struct{}{}

	} else if ce.hasMajorityCeil(nacks) {
		haltReason := fmt.Sprintf("Majority of the validators have rejected the block, halting the network: %d acks, %d nacks", acks, nacks)
		ce.haltChan <- haltReason
		return nil
	}

	// No majority yet, wait for more votes
	return nil
}

func (ce *ConsensusEngine) validatorSetHash() types.Hash {
	hasher := ktypes.NewHasher()

	keys := make([]string, 0, len(ce.validatorSet))
	for _, v := range ce.validatorSet {
		keys = append(keys, config.EncodePubKeyAndType(v.Identifier, v.KeyType))
	}

	// sort the keys
	slices.Sort(keys)

	for _, k := range keys {
		val := ce.validatorSet[k]
		hasher.Write(val.Identifier)
		binary.Write(hasher, binary.BigEndian, val.KeyType)
		binary.Write(hasher, binary.BigEndian, val.Power)
	}

	return hasher.Sum(nil)
}

// CancelBlockExecution is used by the leader to manually cancel the block execution
// if it is taking too long to execute. This method takes the height of the block to
// be cancelled and the list of long transaction IDs to be evicted from the mempool.
// One concern is: what if the block finishes execution and the leader tries to cancel it,
// and the resolutions update some internal state that cannot be reverted?
func (ce *ConsensusEngine) CancelBlockExecution(height int64, txIDs []types.Hash) error {
	ce.log.Info("Block execution cancel request received", "height", height)
	// Ensure we are cancelling the block execution for the current block
	ce.stateInfo.mtx.RLock()
	defer ce.stateInfo.mtx.RUnlock()

	// Check if the height is the same as the current block height
	if height != ce.stateInfo.height+1 {
		ce.log.Warn("Cannot cancel block execution, block height does not match", "height", height, "current", ce.stateInfo.height+1)
		return fmt.Errorf("cannot cancel block execution for block height %d, currently executing %d", height, ce.stateInfo.height+1)
	}

	// Check if a block is proposed
	if ce.stateInfo.blkProp == nil {
		ce.log.Warn("Cannot cancel block execution, no block is proposed yet", "height", height)
		return fmt.Errorf("cannot cancel block execution, no block is proposed yet")
	}

	// Cannot cancel if the block is already finished executing or committed
	if ce.stateInfo.status != Proposed {
		ce.log.Warn("Cannot cancel block execution, block is already executed or committed", "height", height)
		return fmt.Errorf("cannot cancel block execution, block is already executed or committed")
	}

	// Cancel the block execution
	ce.cancelFnMtx.Lock()
	defer ce.cancelFnMtx.Unlock()

	ce.longRunningTxs = append([]ktypes.Hash{}, txIDs...)

	if ce.blkExecCancelFn != nil {
		ce.log.Info("Cancelling block execution", "height", height, "txIDs", txIDs)
		ce.blkExecCancelFn()
	} else {
		ce.log.Error("Block execution cancel function not set")
		return errors.New("block execution cancel function not set")
	}

	return nil
}
