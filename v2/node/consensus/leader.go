package consensus

import (
	"context"
	"fmt"
	"time"

	"p2p/node/types"
)

// Leader is the node that proposes the block and drives the consensus process:
// 1. Prepare Phase:
//   - Create a block proposal
//   - Broadcast the block proposal
//   - Process the block and generate the appHash
//   - Wait for the votes from the validators
//
// 2. Commit Phase:
//   - Commit the block and start next prepare phase

// startNewRound starts a new round of consensus process.
func (ce *ConsensusEngine) startNewRound(ctx context.Context) error {
	fmt.Println("Starting a new round", ce.state.lc.height+1)
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	blkProp, err := ce.createBlockProposal()
	if err != nil {
		fmt.Println("Error creating a block proposal", err)
		return err
	}

	// Validate the block proposal before announcing it to the network
	if err := ce.validateBlock(blkProp.blk); err != nil {
		fmt.Println("Error validating the block proposal", err)
		return err
	}
	ce.state.blkProp = blkProp

	// Broadcast the block proposal to the network
	go ce.proposalBroadcaster(ctx, blkProp.blk, ce.host)

	// Execute the block and generate the appHash
	if err := ce.executeBlock(); err != nil {
		fmt.Println("Error executing the block", err)
		return err
	}

	// Add its own vote to the votes map
	ce.state.votes[string(ce.pubKey)] = &vote{
		ack:     true,
		appHash: &ce.state.blockRes.appHash,
	}

	ce.processVotes(ctx)
	return nil
}

// NotifyACK notifies the consensus engine about the ACK received from the validator.
// This only notifies if leader is still processing the block the vote is for.
func (ce *ConsensusEngine) NotifyACK(validatorPK []byte, ack types.AckRes) {
	// fmt.Println("NotifyACK: Received ACK from validator", string(validatorPK), ack.Height, ack.ACK, ack.BlkHash, ack.AppHash)
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if ce.role != types.RoleLeader {
		return
	}

	// Check if the vote is for the current block
	if ce.state.blkProp == nil {
		fmt.Println("NotifyACK: Not processing any block proposal at the moment")
		return
	}

	if ce.state.blkProp.height != ack.Height {
		fmt.Println("NotifyACK: Vote received for a different block, ignore it.", ce.state.blkProp.height, ack.Height)
		return
	}

	// If the ack is for the current height, but the block hash is different? Ignore it.
	if ce.state.blkProp.blkHash != ack.BlkHash {
		fmt.Println("NotifyACK: Vote received for an incorrect block", ce.state.blkProp.blkHash, ack.BlkHash)
		return
	}

	// else notify the vote to the consensus engine
	voteMsg := &vote{
		ack:     ack.ACK,
		appHash: ack.AppHash,
		blkHash: ack.BlkHash,
		height:  ack.Height,
	}
	ce.sendConsensusMessage(&consensusMessage{
		MsgType: voteMsg.Type(),
		Msg:     voteMsg,
		Sender:  validatorPK,
	})
	return
}

// Create Proposal creates a new block proposal for the leader
// by reaping the transactions from the mempool. This also adds the
// proposer transactions such as ValidatorVoteBodies.
// This method orders the transactions in the nonce order and also
// does basic gas and balance checks and enforces the block size limits.
func (ce *ConsensusEngine) createBlockProposal() (*blockProposal, error) {
	// fmt.Println("Creating a new block proposal")
	_, txns := ce.mempool.ReapN(blockTxCount)
	blk := types.NewBlock(ce.state.lc.height+1, ce.state.lc.blkHash, ce.state.lc.appHash, time.Now(), txns)
	return &blockProposal{
		height:  ce.state.lc.height + 1,
		blkHash: blk.Header.Hash(),
		blk:     blk,
	}, nil
}

// addVote registers the vote received from the validator if it is for the current block.
func (ce *ConsensusEngine) addVote(ctx context.Context, vote *vote, sender string) error {
	// fmt.Println("Adding vote", vote, sender)
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if ce.state.blkProp == nil {
		return fmt.Errorf("not processing any block proposal at the moment")
	}

	// check if the vote is for the current height
	if ce.state.blkProp.height != vote.height {
		return fmt.Errorf("vote received for a different block height, ignore it")
	}

	// check if the vote is for the current block and from a validator
	if ce.state.blkProp.blkHash != vote.blkHash {
		return fmt.Errorf("vote received for an incorrect block %s", vote.blkHash.String())
	}

	// Check if the vote is from a validator
	// TODO: accept vote from anyone, till we have access to validator set
	// if _, ok := ce.validatorSet[sender]; !ok {
	// 	return fmt.Errorf("vote received from an unknown validator %s", sender)
	// }

	// Add the vote to the votes map
	if _, ok := ce.state.votes[sender]; !ok {
		ce.state.votes[sender] = vote
	}

	// Good to do this sequentially, so that we only trigger one nextRound goroutine
	ce.processVotes(ctx)
	return nil
}

// ProcessVotes processes the votes received from the validators.
// Depending on the votes, leader will trigger one of the following:
// 1. Commit the block
// 2. Re-announce the block proposal
// 3. Halt the network (should there be a message to halt the network?)
// Leaders will re-announce the blkProp and blkRes for every reannounceTimer interval
// for the slow valdiators to catchup incase they missed the event.
// Validators will peridically reannounce the votes to the leader.
func (ce *ConsensusEngine) processVotes(ctx context.Context) error {
	fmt.Println("Processing votes")

	if ce.state.blkProp == nil || ce.state.blockRes == nil { // Moved onto the next round
		return nil
	}

	threshold := ce.requiredThreshold()
	if len(ce.state.votes) < int(threshold) {
		fmt.Println("Not enough votes received yet", len(ce.state.votes), threshold)
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

	if acks >= threshold {
		fmt.Println("Majority of the validators have accepted the block, proceeding to commit the block", ce.state.blkProp.blk.Header.Height, ce.state.blkProp.blkHash, acks, nacks)
		// Commit the block and broadcast the blockAnn message
		if err := ce.commit(); err != nil {
			fmt.Println("Error committing the block: process votes", err)
			return err
		}

		fmt.Println("Announce committed block", ce.state.blkProp.blk.Header.Height, ce.state.blkProp.blkHash)
		// Broadcast the blockAnn message
		go ce.blkAnnouncer(ctx, ce.state.blkProp.blk, ce.state.blockRes.appHash, ce.host)

		// start the next round
		ce.nextState()
		// Wait for the timeout to start the next round
		time.Sleep(1 * time.Second) // TODO: Is there a better way to handle this timeout?
		go ce.startNewRound(ctx)
	} else if nacks >= threshold {
		// halt the network
		fmt.Println("Majority of the validators have rejected the block, halting the network",
			ce.state.blkProp.blk.Header.Height, acks, nacks)
		close(ce.haltChan)
		return nil
	}

	// No majority yet, wait for more votes
	return nil
}
