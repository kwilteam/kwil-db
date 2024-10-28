package consensus

import (
	"context"
	"fmt"
	"p2p/node/types"
)

// AcceptProposal checks if the node should download the block corresponding to the proposal.
// This should not be processed by the leader and the sentry nodes.
// Validator should only accept the proposal if it is not already processing a block and
// the proposal is for the next block to be processed.
func (ce *ConsensusEngine) AcceptProposal(height int64, prevBlockID types.Hash) bool {
	fmt.Println("Accept proposal?", height, prevBlockID)
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	ce.updateNetworkHeight(height - 1)

	if ce.role != types.RoleValidator {
		return false
	}

	// initial block must precede genesis
	if height == 1 || prevBlockID.IsZero() {
		return ce.state.lc.blkHash.IsZero()
	}

	// Check if the validator is busy processing a block.
	if ce.state.blkProp != nil {
		return false
	}

	// Check if this is the next block to be processed
	if height != ce.state.lc.height+1 {
		return false
	}

	return prevBlockID == ce.state.lc.blkHash
}

// NotifyBlockProposal is used by the p2p stream handler to notify the consensus engine of a new block proposal.
// Only a validator should use this method, not leader or sentry. This method does it's best to ensure that this
// is the next block to be processed, only then it notifies the consensus engine of the block proposal.
// respCb is a callback function used to send the VoteMessage(ack/nack) back to the leader.
func (ce *ConsensusEngine) NotifyBlockProposal(blk *types.Block) {
	// fmt.Println("Notify block proposal", blk.Header.Height, blk.Header.Hash())
	if ce.role == types.RoleLeader {
		return
	}

	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if ce.state.blkProp != nil {
		fmt.Println("block proposal already exists")
		return
	}

	if blk.Header.Height != ce.state.lc.height+1 {
		fmt.Printf("proposal for height %d does not follow %d", blk.Header.Height, ce.state.lc.height)
		return
	}

	blkProp := &blockProposal{
		height:  blk.Header.Height,
		blkHash: blk.Header.Hash(),
		blk:     blk,
	}

	ce.sendConsensusMessage(&consensusMessage{
		MsgType: blkProp.Type(),
		Msg:     blkProp,
		Sender:  ce.pubKey,
	})
}

// AcceptCommit handles the blockAnnounce message from the leader.
// This should be processed only if this is the next block to be committed by the node.
// This also checks if the node should request the block from its peers. This can happen
// 1. If the node is a sentry node and doesn't have the block.
// 2. If the node is a validator and missed the block proposal message.
func (ce *ConsensusEngine) AcceptCommit(height int64, blkID types.Hash, appHash types.Hash) bool {
	// fmt.Println("Accept commit?", height, blkID, appHash)
	if ce.role == types.RoleLeader {
		return false
	}

	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	ce.updateNetworkHeight(height)

	fmt.Println("Accept commit?", height, blkID, appHash, ce.state.lc.height+1, ce.state.lc.blkHash)
	fmt.Println("lc: ", ce.state.blkProp, ce.state.blockRes)
	if ce.state.lc.height+1 != height {
		// This is not the next block to be committed by the node.
		return false
	}

	blkCommit := &blockAnnounce{
		appHash: appHash,
	}

	if ce.state.blkProp != nil {
		if ce.state.blockRes == nil {
			// still processing the block, ignore the commit message for now and commit when ready.
			return false
		} else {
			// Waiting for the block to be committed, notify the consensus engine to commit the block.
			blkCommit.blk = ce.state.blkProp.blk
			ce.sendConsensusMessage(&consensusMessage{
				MsgType: blkCommit.Type(),
				Msg:     blkCommit,
				Sender:  ce.pubKey,
			})
		}
	} else {
		// either sentry node or slow validator
		// check if this is the first time we are hearing about this block and not already downloaded it.
		blk, _, err := ce.blockStore.Get(blkID)
		if err != nil {
			return true
		}

		blkCommit.blk = blk
		ce.sendConsensusMessage(&consensusMessage{
			MsgType: blkCommit.Type(),
			Msg:     blkCommit,
			Sender:  ce.pubKey,
		})
	}

	return false
}

// TODO: Can we club this and AcceptCommit into a single method?
// NotifyBlockCommit is used by the p2p stream handler to notify the consensus engine of a new block commit.
// It validates blk height, appHash and blkHash and only then notifies the consensus engine to commit the block.
func (ce *ConsensusEngine) NotifyBlockCommit(blk *types.Block, appHash types.Hash) {
	// fmt.Println("Notify block commit", blk.Header.Height, blk.Header.Hash(), appHash)
	if ce.role == types.RoleLeader {
		// Leader can also use this in blocksync mode, when it tries to replay the blocks or catchup with the network.
		return
	}

	ce.state.mtx.Lock()
	ce.state.mtx.Unlock()

	if ce.state.lc.height+1 != blk.Header.Height {
		return
	}

	if ce.state.blkProp != nil && ce.state.blockRes == nil {
		fmt.Printf("still processing the block, ignore the commit message for now and commit when ready")
		return
	}

	// Notify the consensus engine to commit the block in the below scenarios:
	// 1. Sentry node: Execute the block, validate the appHash and commit the block.
	// 2. Validator:
	// - No blockProposal received: Execute the block, validate the appHash and commit the block.
	// - Incorrect Block received: Rollback and reprocess the block sent as part of the commit message.
	// - Incorrect AppHash: Halt the node.
	blkCommit := &blockAnnounce{
		blk:     blk,
		appHash: appHash,
	}

	ce.sendConsensusMessage(&consensusMessage{
		MsgType: blkCommit.Type(),
		Msg:     blkCommit,
		Sender:  ce.pubKey,
	})
	// fmt.Println("Notified consensus engine to commit the block", blk.Header.Height, blk.Header.Hash(), appHash)
}

// ProcessBlockProposal is used by the validator's consensus engine to process the new block proposal message.
// This method is used to validate the received block, execute the block and generate appHash and
// report the result back to the leader.
func (ce *ConsensusEngine) processBlockProposal(_ context.Context, blkPropMsg *blockProposal) error {
	fmt.Println("Processing block proposal", blkPropMsg.blk.Header.Height, blkPropMsg.blk.Header.Hash())
	if ce.role != types.RoleValidator {
		fmt.Println("Only validators can process block proposals")
		return nil
	}

	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if ce.state.blkProp != nil {
		return fmt.Errorf("We are in the process of executing a block, can't accept a new block proposal.")
	}

	if err := ce.validateBlock(blkPropMsg.blk); err != nil {
		go ce.ackBroadcaster(false, blkPropMsg.height, blkPropMsg.blkHash, nil)
		fmt.Println("Error validating block, sending NACK", err)
		return err
	}
	ce.state.blkProp = blkPropMsg

	if err := ce.executeBlock(); err != nil {
		// TODO: what to do if the block execution fails? Send NACK?
		fmt.Println("Error executing block", err)
		return err
	}

	// Broadcast the result back to the leader
	fmt.Println("Sending ack to the leader", blkPropMsg.height, blkPropMsg.blkHash.String(), ce.state.blockRes.appHash.String())
	go ce.ackBroadcaster(true, blkPropMsg.height, blkPropMsg.blkHash, &ce.state.blockRes.appHash)

	return nil
}

// This is triggered in response to the blockAnn message from the leader.
// This method is used by the sentry and the validators nodes to commit the specified block.
// If the validator node processed a different block, it should rollback and reprocess the block.
// Validator nodes can skip the block execution and directly commit the block if they have already processed the block.
// The nodes should only commit the block if the appHash is valid, else halt the node.
func (ce *ConsensusEngine) commitBlock(blk *types.Block, appHash types.Hash) error {
	// fmt.Println("processing Commit block", blk.Header.Height, blk.Header.Hash(), appHash)
	if ce.role == types.RoleLeader {
		return nil
	}

	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	// Three different scenarios are possible here:
	// 1. Sentry node: Execute the block, validate the appHash and commit the block.
	// 2. Validator:
	// - No blockProposal received: Execute the block, validate the appHash and commit the block.
	// - Incorrect Block received: Rollback and reprocess the block sent as part of the commit message.
	// - Incorrect AppHash: Halt the node.

	if ce.role == types.RoleSentry {
		go ce.processAndCommit(blk, appHash)
		return nil
	}

	// You are a validator
	if ce.state.blkProp == nil {
		// No block proposal received, execute the block, validate the appHash and commit the block.
		go ce.processAndCommit(blk, appHash)
		return nil
	}

	// ensure that you are processing the correct block
	if ce.state.blkProp.blkHash != blk.Header.Hash() {
		// Rollback and reprocess the block sent as part of the commit message.
		ce.resetState()
		// TODO: somehow signal the current block processing to halt and reprocess the new block.
		go ce.processAndCommit(blk, appHash)
	}

	if ce.state.blockRes == nil {
		// Still processing the block, return and commit when ready.
		return nil
	}

	if ce.state.blockRes.appHash != appHash {
		fmt.Println("Incorrect AppHash, halt the node.", appHash.String(), ce.state.blockRes.appHash.String())
		close(ce.haltChan)
		return nil
	}

	// Commit the block
	if err := ce.commit(); err != nil {
		fmt.Println("Error committing block", err)
		return err
	}

	// Move to the next state
	ce.nextState()
	return nil
}

// processAndCommit: used by the sentry nodes and slow validators to process and commit the block.
// This is used when the acks are not required to be sent back to the leader, essentially in catchup mode.
func (ce *ConsensusEngine) processAndCommit(blk *types.Block, appHash types.Hash) error {
	fmt.Println("Processing committed block", blk.Header.Height, blk.Header.Hash().String(), appHash.String())
	if err := ce.validateBlock(blk); err != nil {
		fmt.Println("Error validating block", err)
		return err
	}
	ce.state.blkProp = &blockProposal{
		height:  blk.Header.Height,
		blkHash: blk.Header.Hash(),
		blk:     blk,
		// respCb is not required here as we are not sending acks back to the leader.
	}

	if err := ce.executeBlock(); err != nil {
		fmt.Println("Error executing block", err)
		return err
	}

	if ce.state.blockRes.appHash != appHash {
		// Incorrect AppHash, halt the node.
		fmt.Println("Incorrect AppHash, processAndCommit.", appHash.String(), ce.state.blockRes.appHash.String())
		// ce.haltChan <- struct{}{}
		close(ce.haltChan)
		return fmt.Errorf("appHash mismatch, expected: %s, received: %s", appHash.String(), ce.state.blockRes.appHash.String())
	}

	// Commit the block if the appHash is valid
	if err := ce.commit(); err != nil {
		fmt.Println("Error committing block", err)
		return err
	}

	// Move to the next state
	ce.nextState()
	return nil
}
