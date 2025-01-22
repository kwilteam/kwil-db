package consensus

import (
	"context"
	"errors"
	"fmt"

	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
)

// AcceptProposal checks if the node should download the block corresponding to the proposal.
// This should not be processed by the leader and the sentry nodes.
// Validator should only accept the proposal if it is not already processing a block and
// the proposal is for the next block to be processed.
// If we receive a new proposal for the same height, abort the execution of the current proposal and
// start processing the new proposal.
func (ce *ConsensusEngine) AcceptProposal(height int64, blkID, prevBlockID types.Hash, leaderSig []byte, timestamp int64) bool {
	if ce.role.Load() != types.RoleValidator {
		return false
	}

	ce.stateInfo.mtx.RLock()
	defer ce.stateInfo.mtx.RUnlock()

	ce.log.Info("Accept proposal?", "height", height, "blkID", blkID, "prevHash", prevBlockID)

	if height != ce.stateInfo.height+1 {
		ce.log.Debug("Block proposal is not for the next height", "blkPropHeight", height, "expected", ce.stateInfo.height+1)
		// but we already did ce.updateNetworkHeight... (?)
		return false
	}

	// check if the blkProposal is from the leader
	valid, err := ce.leader.Verify(blkID[:], leaderSig)
	if err != nil {
		ce.log.Error("Error verifying leader signature", "error", err)
		return false
	}

	if !valid {
		ce.log.Debug("Invalid leader signature, ignoring the block proposal msg: ", "height", height)
		return false
	}

	// Check if the validator is busy processing a block.
	if ce.stateInfo.status != Committed {
		// check if we are processing a different block, if yes then reset the state.
		if ce.stateInfo.blkProp.blkHash != blkID && ce.stateInfo.blkProp.blk.Header.Timestamp.UnixMilli() < timestamp {
			ce.log.Debug("Conflicting block proposals, abort block execution and requesting the latest block: ", "height", height)
			// go ce.sendResetMsg(ce.stateInfo.height)
			return true
		}
		ce.log.Debug("Already processing the block proposal", "height", height, "blkID", blkID)
		return false
	}

	// not processing any block, accept the proposal
	return true
}

// AcceptCommit handles the blockAnnounce message from the leader.
// This should be processed only if this is the next block to be committed by the node.
// This also checks if the node should request the block from its peers. This can happen
// 1. If the node is a sentry node and doesn't have the block.
// 2. If the node is a validator and missed the block proposal message.
func (ce *ConsensusEngine) AcceptCommit(height int64, blkID types.Hash, ci *types.CommitInfo, leaderSig []byte) bool {
	if ce.role.Load() == types.RoleLeader {
		return false
	}

	ce.stateInfo.mtx.RLock()
	defer ce.stateInfo.mtx.RUnlock()

	ce.log.Debugf("Accept commit? height: %d,  blockID: %s, appHash: %s, lastCommitHeight: %d", height, blkID, ci.AppHash, ce.stateInfo.height)

	if ce.stateInfo.height+1 != height {
		return false
	}

	// Verify the leader signature
	valid, err := ce.leader.Verify(blkID[:], leaderSig)
	if err != nil {
		ce.log.Error("Error verifying leader signature", "error", err)
		return false
	}

	if !valid {
		ce.log.Info("Invalid leader signature, ignoring the blkAnn message ", "height", height)
		return false
	}

	blkCommit := &blockAnnounce{
		ci: ci,
	}
	if ce.stateInfo.status == Committed {
		// Sentry node or slow validator
		// check if this is the first time we are hearing about this block and not already downloaded it.
		blk, _, err := ce.blockStore.Get(blkID)
		if err != nil {
			return true
		}

		blkCommit.blk = blk
		go ce.sendConsensusMessage(&consensusMessage{
			MsgType: blkCommit.Type(),
			Msg:     blkCommit,
			Sender:  ce.leader.Bytes(),
		})
		return false
	}

	// We are currently processing a block proposal, ensure that it's the correct block proposal.
	if ce.stateInfo.blkProp.blkHash != blkID {
		// Rollback and reprocess the block sent as part of the commit message.
		ce.log.Debug("Processing incorrect block, notify consensus engine to abort: ", "height", height)
		go ce.sendResetMsg(&resetMsg{height: height})
		return true // fetch the correct block
	}

	if ce.stateInfo.status == Proposed {
		// still processing the block, ignore the commit message for now and commit when ready.
		return false
	}

	blkCommit.blk = ce.stateInfo.blkProp.blk
	go ce.sendConsensusMessage(&consensusMessage{
		MsgType: blkCommit.Type(),
		Msg:     blkCommit,
		Sender:  ce.leader.Bytes(),
	})

	return false
}

// ProcessBlockProposal is used by the validator's consensus engine to process the new block proposal message.
// This method is used to validate the received block, execute the block and generate appHash and
// report the result back to the leader.
func (ce *ConsensusEngine) processBlockProposal(ctx context.Context, blkPropMsg *blockProposal) error {
	if ce.role.Load() != types.RoleValidator {
		ce.log.Warn("Only validators can process block proposals")
		return nil
	}

	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if ce.state.lc.height+1 != blkPropMsg.height {
		ce.log.Info("Block proposal is not for the current height", "blkPropHeight", blkPropMsg.height, "expected", ce.state.lc.height+1)
		return nil
	}

	if ce.state.blkProp != nil {
		if ce.state.blkProp.blkHash == blkPropMsg.blkHash {
			ce.log.Info("Already processing the block proposal", "height", blkPropMsg.height)
			return nil
		}

		if ce.state.blkProp.blk.Header.Timestamp.After(blkPropMsg.blk.Header.Timestamp) {
			ce.log.Info("Received stale block proposal, Ignore", "height", blkPropMsg.height, "blkHash", blkPropMsg.blkHash)
			return nil
		}

		ce.log.Info("Aborting execution of stale block proposal", "height", blkPropMsg.height, "blkHash", ce.state.blkProp.blkHash)
		if err := ce.rollbackState(ctx); err != nil {
			ce.log.Error("Error aborting execution of block", "height", blkPropMsg.height, "blkID", ce.state.blkProp.blkHash, "error", err)
			return err
		}
	}

	ce.log.Info("Processing block proposal", "height", blkPropMsg.blk.Header.Height, "header", blkPropMsg.blk.Header)

	if err := ce.validateBlock(blkPropMsg.blk); err != nil {
		ce.log.Error("Error validating block, sending NACK", "error", err)
		go ce.ackBroadcaster(false, blkPropMsg.height, blkPropMsg.blkHash, nil, nil)
		return err
	}
	ce.state.blkProp = blkPropMsg
	ce.state.blockRes = nil

	// Update the stateInfo
	ce.stateInfo.mtx.Lock()
	ce.stateInfo.status = Proposed
	ce.stateInfo.blkProp = blkPropMsg
	ce.stateInfo.mtx.Unlock()

	// execCtx is applicable only for the duration of the block execution
	// This is used to react to the leader's reset message by cancelling the block execution.
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set the cancel function for the block execution
	ce.cancelFnMtx.Lock()
	ce.blkExecCancelFn = cancel
	ce.cancelFnMtx.Unlock()

	if err := ce.executeBlock(execCtx, blkPropMsg); err != nil {
		if errors.Is(err, context.Canceled) {
			ce.log.Info("Block execution cancelled", "height", blkPropMsg.height)
			return nil
		}

		// go ce.ackBroadcaster(false, blkPropMsg.height, blkPropMsg.blkHash, nil)
		return err
	}

	// Broadcast the result back to the leader
	ce.log.Info("Sending ack to the leader", "height", blkPropMsg.height,
		"hash", blkPropMsg.blkHash, "appHash", ce.state.blockRes.appHash)

	signature, err := types.SignVote(blkPropMsg.blkHash, true, &ce.state.blockRes.appHash, ce.privKey)
	if err != nil {
		ce.log.Error("Error signing the voteInfo", "error", err)
	}
	voteInfo := &vote{
		ack:       true,
		blkHash:   blkPropMsg.blkHash,
		height:    blkPropMsg.height,
		appHash:   &ce.state.blockRes.appHash,
		signature: signature,
	}
	ce.state.blockRes.vote = voteInfo

	go ce.ackBroadcaster(true, blkPropMsg.height, blkPropMsg.blkHash, &ce.state.blockRes.appHash, signature.Data)

	return nil
}

// This is triggered in response to the blockAnn message from the leader.
// This method is used by the sentry and the validators nodes to commit the specified block.
// If the validator node processed a different block, it should rollback and reprocess the block.
// Validator nodes can skip the block execution and directly commit the block if they have already processed the block.
// The nodes should only commit the block if the appHash is valid, else halt the node.
func (ce *ConsensusEngine) commitBlock(ctx context.Context, blk *ktypes.Block, ci *types.CommitInfo) error {
	if ce.role.Load() == types.RoleLeader {
		return nil
	}

	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if ce.state.lc.height+1 != blk.Header.Height { // only accept/process the block if it is for the next height
		return nil
	}

	// Three different scenarios are possible here:
	// 1. Sentry node: Execute the block, validate the appHash and commit the block.
	// 2. Validator:
	// - No blockProposal received: Execute the block, validate the appHash and commit the block.
	// - Incorrect Block received: Rollback and reprocess the block sent as part of the commit message.
	// - Incorrect AppHash: Halt the node.

	if ce.role.Load() == types.RoleSentry {
		return ce.processAndCommit(ctx, blk, ci)
	}

	// You are a validator
	if ce.state.blkProp == nil {
		return ce.processAndCommit(ctx, blk, ci)
	}

	if ce.state.blkProp.blkHash != blk.Header.Hash() {
		if err := ce.rollbackState(ctx); err != nil {
			ce.log.Error("error aborting execution of incorrect block proposal", "height", blk.Header.Height, "blkID", blk.Header.Hash(), "error", err)
		}
		return ce.processAndCommit(ctx, blk, ci)
	}

	if ce.state.blockRes == nil {
		// Still processing the block, return and commit when ready.
		return nil
	}

	if !ce.state.blockRes.paramUpdates.Equals(ci.ParamUpdates) { // this is absorbed in apphash anyway, but helps diagnostics
		haltR := fmt.Sprintf("Incorrect ParamUpdates, halting the node. received: %s, computed: %s", ci.ParamUpdates, ce.state.blockRes.paramUpdates)
		ce.haltChan <- haltR
		return nil
	}
	if ce.state.blockRes.appHash != ci.AppHash {
		haltR := fmt.Sprintf("Incorrect AppHash, halting the node. received: %s, computed: %s", ci.AppHash, ce.state.blockRes.appHash)
		ce.haltChan <- haltR
		return nil
	}

	if err := ce.acceptCommitInfo(ci, ce.state.blkProp.blkHash); err != nil {
		ce.log.Error("Error validating commitInfo", "error", err)
		return err
	}

	// Commit the block
	if err := ce.commit(ctx); err != nil {
		ce.log.Errorf("Error committing block: height: %d, error: %v", blk.Header.Height, err)
		return err
	}

	// Move to the next state
	ce.nextState()
	return nil
}

// processAndCommit: used by the sentry nodes and slow validators to process and commit the block.
// This is used when the acks are not required to be sent back to the leader, essentially in catchup mode.
func (ce *ConsensusEngine) processAndCommit(ctx context.Context, blk *ktypes.Block, ci *types.CommitInfo) error {
	blkID := blk.Header.Hash()
	if ci == nil {
		return fmt.Errorf("commitInfo is nil")
	}

	ce.log.Info("Processing committed block", "height", blk.Header.Height, "hash", blkID, "appHash", ci.AppHash)
	if err := ce.validateBlock(blk); err != nil {
		return err
	}

	ce.state.blkProp = &blockProposal{
		height:  blk.Header.Height,
		blkHash: blk.Header.Hash(),
		blk:     blk,
	}

	// Update the stateInfo
	ce.stateInfo.mtx.Lock()
	ce.stateInfo.status = Proposed
	ce.stateInfo.blkProp = ce.state.blkProp
	ce.stateInfo.mtx.Unlock()

	if err := ce.executeBlock(ctx, ce.state.blkProp); err != nil {
		return fmt.Errorf("error executing block: %v", err)
	}

	if !ce.state.blockRes.paramUpdates.Equals(ci.ParamUpdates) { // this is absorbed in apphash anyway, but helps diagnostics
		haltR := fmt.Sprintf("processAndCommit: Incorrect ParamUpdates, halting the node. received: %s, computed: %s", ci.ParamUpdates, ce.state.blockRes.paramUpdates)
		ce.haltChan <- haltR
		return errors.New(haltR)
	}
	if ce.state.blockRes.appHash != ci.AppHash { // do in acceptCommitInfo?
		haltR := fmt.Sprintf("processAndCommit: AppHash mismatch, halting the node. expected: %s, received: %s", ce.state.blockRes.appHash, ci.AppHash)
		ce.haltChan <- haltR
		return errors.New(haltR)
	}

	// Commit the block if the appHash and commitInfo is valid
	if err := ce.acceptCommitInfo(ci, blkID); err != nil {
		return fmt.Errorf("error validating commitInfo: %v", err)
	}

	if err := ce.commit(ctx); err != nil {
		return fmt.Errorf("error committing block: %v", err)
	}

	ce.nextState()
	return nil
}

func (ce *ConsensusEngine) acceptCommitInfo(ci *types.CommitInfo, blkID ktypes.Hash) error {
	if ci == nil {
		return fmt.Errorf("commitInfo is nil")
	}

	// Validate CommitInfo
	var acks int
	for _, vote := range ci.Votes {
		err := vote.Verify(blkID, ci.AppHash)
		if err != nil {
			return fmt.Errorf("error verifying vote: %v", err)
		}

		if vote.AckStatus == types.AckStatusAgree {
			acks++
		}
	}

	if err := ktypes.ValidateUpdates(ci.ParamUpdates); err != nil {
		return fmt.Errorf("paramUpdates failed validation: %v", err)
	}

	if !ce.hasMajorityCeil(acks) {
		return fmt.Errorf("invalid blkAnn message, not enough acks in the commitInfo, leader misbehavior: %d", acks)
	}

	ce.state.commitInfo = ci

	return nil
}
