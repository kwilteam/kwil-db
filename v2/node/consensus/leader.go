package consensus

import (
	"context"
	"fmt"
	"time"

	"p2p/node/types"
)

// Leader Modes:
// 1. Prepare Phase:
//   - Create a block proposal
//   - Broadcast the block proposal
//   - Process the block and generate the appHash
//   - Wait for the votes from the validators
//
// 2. Commit Phase:
//   - Commit the block and start next prepare phase
func (ce *ConsensusEngine) startNewRound(ctx context.Context) error {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	blkProp, err := ce.createProposal()
	if err != nil {
		// Log error
		return err
	}

	// Validate the block proposal before announcing it to the network
	if err := ce.validateBlock(blkProp.blk); err != nil {
		return err
	}
	ce.state.blkProp = blkProp

	// Broadcast the block proposal to the network
	go ce.proposalBroadcaster(ctx, blkProp.blk)

	// Execute the block and generate the appHash
	if err := ce.executeBlock(); err != nil {
		return err
	}

	// Add its own vote to the votes map
	ce.state.votes[ce.nodeID] = &vote{
		ack:     true,
		appHash: &ce.state.blockRes.appHash,
	}

	ce.processVotes(ctx)
	return nil
}

// Create Proposal creates a new block proposal for the leader
// by reaping the transactions from the mempool. This also adds the
// proposer transactions such as ValidatorVoteBodies.
// This method orders the transactions in the nonce order and also
// does basic gas and balance checks and enforces the block size limits.
func (ce *ConsensusEngine) createProposal() (*blockProposal, error) {
	_, txns := ce.mempool.ReapN(blockTxCount)
	blk := types.NewBlock(ce.state.lc.height+1, ce.state.lc.blkHash, ce.state.lc.appHash, time.Now(), txns)
	return &blockProposal{
		height:  ce.state.lc.height + 1,
		blkHash: blk.Header.Hash(),
		blk:     blk,
	}, nil
}

// addVote registers the vote received from the validator if it is for the current block.
func (ce *ConsensusEngine) addVote(vote *vote, sender string) error {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if ce.state.blkProp == nil {
		return fmt.Errorf("not processing any block proposal at the moment")
	}

	// check if the vote is for the current block and from a validator
	if ce.state.blkProp.blkHash != vote.blkHash {
		return fmt.Errorf("vote received for an incorrect block %s", vote.blkHash.String())
	}

	// Check if the vote is from a validator
	if _, ok := ce.validatorSet[sender]; !ok {
		return fmt.Errorf("vote received from an unknown validator %s", sender)
	}

	// Add the vote to the votes map
	if _, ok := ce.state.votes[sender]; !ok {
		ce.state.votes[sender] = vote
	}
	return nil
}

// ProcessVotes processes the votes received from the validators.
// If threshold Acks are received with matching appHashes, then the block is committed.
// If threshold Nacks are received, then the network stops.
// If enough votes are not received, wait for them. Validators are programmed to
// repeatedly send votes in regular intervals until they receive a BlockAnn or BlockProp msg.
// If validators missed blockProp messages probably because they are catching up with the network,
// Leader will re-annonuce the blockProp messages at regular intervals until it receives the votes(threshold).
func (ce *ConsensusEngine) processVotes(ctx context.Context) error {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	// check if the votes are already processed and moved to the next round
	if ce.state.blkProp == nil || ce.state.blockRes == nil {
		return nil
	}

	// Count the votes
	var acks, nacks int64
	expectedHash := ce.state.blockRes.appHash
	for _, vote := range ce.state.votes {
		if vote.ack && vote.appHash != nil && *vote.appHash == expectedHash {
			acks++
		} else {
			nacks++
		}
	}

	// Depending on the votes, leader will trigger one of the following:
	// 1. Commit the block
	// 2. Re-announce the block proposal
	// 3. Halting the network (should there be a message to halt the network?)
	threshold := ce.requiredThreshold()

	if acks >= threshold {
		// Commit the block and broadcast the blockAnn message
		if err := ce.commit(); err != nil {
			return err
		}

		// Broadcast the blockAnn message
		go ce.blkAnnouncer(ctx, ce.state.lc.height, ce.state.lc.blkHash, ce.state.lc.appHash)

		// start the next round
		ce.nextState()
		ce.startNewRound(ctx)
	} else if nacks >= threshold {
		// Majority of the validators have either rejected the block or disagreed on the appHash
		// halt the network
		ce.haltChan <- struct{}{}
		return nil
	}

	// If the threshold is not reached, leader will re-announce the block proposal at regular intervals
	// until it receives the votes(threshold)
	return nil
}

func (ce *ConsensusEngine) NotifyACK(validatorPK []byte, ack types.AckRes) {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	// Check if the vote is for the current block
	if ce.state.blkProp.blkHash != ack.BlkHash {
		return
	}

	// else notify the vote to the consensus engine
	voteMsg := &vote{
		ack:     ack.ACK,
		appHash: ack.AppHash,
	}
	ce.sendConsensusMessage(&consensusMessage{
		MsgType: voteMsg.Type(),
		Msg:     voteMsg,
	})
	return
}
