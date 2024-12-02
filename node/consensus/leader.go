package consensus

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"time"

	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
)

var lastReset int64 = 0

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
	ce.log.Info("Starting a new round", "height", ce.state.lc.height+1)
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	blkProp, err := ce.createBlockProposal()
	if err != nil {
		ce.log.Errorf("Error creating a block proposal: %v", err)
		return err
	}

	ce.log.Info("Created a new block proposal", "height", blkProp.height, "hash", blkProp.blkHash, "header", blkProp.blk.Header)

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

	// Execute the block and generate the appHash
	if err := ce.executeBlock(); err != nil {
		ce.log.Errorf("Error executing the block: %v", err)
		return err
	}

	// Add its own vote to the votes map
	ce.state.votes[string(ce.pubKey.Bytes())] = &vote{
		ack:     true,
		appHash: &ce.state.blockRes.appHash,
		blkHash: blkProp.blkHash,
		height:  blkProp.height,
	}

	// TODO: test resetState
	if ce.state.blkProp.height%10 == 0 && lastReset != ce.state.blkProp.height {
		lastReset = ce.state.blkProp.height
		ce.log.Info("Resetting the state (for testing purposes)", "height", lastReset, " blkHash", ce.state.blkProp.blkHash)
		if err := ce.resetState(ctx); err != nil {
			ce.log.Errorf("Error resetting the state: %v", err)
			return err
		}
		go ce.rstStateBroadcaster(ce.state.lc.height)
		go ce.startNewRound(ctx)
		return nil
	}

	ce.processVotes(ctx)
	return nil
}

// Create Proposal creates a new block proposal for the leader
// by reaping the transactions from the mempool. This also adds the
// proposer transactions such as ValidatorVoteBodies.
// This method orders the transactions in the nonce order and also
// does basic gas and balance checks and enforces the block size limits.
func (ce *ConsensusEngine) createBlockProposal() (*blockProposal, error) {
	// TODO: ReapN shouldn't remove the transactions from the mempool
	_, txns := ce.mempool.ReapN(blockTxCount)
	blk := types.NewBlock(ce.state.lc.height+1, ce.state.lc.blkHash, ce.state.lc.appHash, ce.ValidatorSetHash(), time.Now(), txns)

	// ValSet + valUpdatesHash

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
		return errors.New("not processing any block proposal at the moment")
	}

	// check if the vote is for the current height
	if ce.state.blkProp.height != vote.height {
		return errors.New("vote received for a different block height, ignore it")
	}

	// check if the vote is for the current block and from a validator
	if ce.state.blkProp.blkHash != vote.blkHash {
		return fmt.Errorf("vote received for an incorrect block %s", vote.blkHash.String())
	}

	// Check if the vote is from a validator
	if _, ok := ce.validatorSet[sender]; !ok {
		return fmt.Errorf("vote received from an unknown validator %s", sender)
	}

	ce.log.Info("Adding vote", "height", vote.height, "blkHash", vote.blkHash, "appHash", vote.appHash, "sender", sender)
	if _, ok := ce.state.votes[sender]; !ok {
		ce.state.votes[sender] = vote
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
	ce.log.Info("Processing votes", "height", ce.state.lc.height+1)

	if ce.state.blkProp == nil || ce.state.blockRes == nil {
		// Moved onto the next round or leader still processing the current block
		return nil
	}

	// Count the votes
	var acks, nacks int
	expectedHash := ce.state.blockRes.appHash
	for _, vote := range ce.state.votes {
		if vote.ack && vote.appHash != nil && *vote.appHash == expectedHash {
			acks++
		} else {
			nacks++
		}
	}

	if ce.hasMajority(acks) {
		ce.log.Info("Majority of the validators have accepted the block, proceeding to commit the block",
			"height", ce.state.blkProp.blk.Header.Height, "hash", ce.state.blkProp.blkHash, "acks", acks, "nacks", nacks)
		// Commit the block and broadcast the blockAnn message
		if err := ce.commit(); err != nil {
			ce.log.Errorf("Error committing the block (process votes): %v", err)
			return err
		}

		ce.log.Infoln("Announce committed block", ce.state.blkProp.blk.Header.Height, ce.state.blkProp.blkHash)
		// Broadcast the blockAnn message
		go ce.blkAnnouncer(ctx, ce.state.blkProp.blk, ce.state.blockRes.appHash)

		// start the next round
		ce.nextState()

		go func() { // must not sleep with ce.state mutex locked
			// Wait for the timeout to start the next round
			select {
			case <-ctx.Done():
				return
			case <-time.After(ce.proposeTimeout):
			}
			ce.startNewRound(ctx)
		}()
	} else if ce.hasMajority(nacks) {
		ce.log.Warnln("Majority of the validators have rejected the block, halting the network",
			ce.state.blkProp.blk.Header.Height, acks, nacks)
		close(ce.haltChan)
		return nil
	}

	// No majority yet, wait for more votes
	return nil
}

func (ce *ConsensusEngine) ValidatorSetHash() types.Hash {
	hasher := ktypes.NewHasher()

	keys := make([]string, 0, len(ce.validatorSet))
	for _, v := range ce.validatorSet {
		keys = append(keys, v.PubKey.String())
	}

	// sort the keys
	slices.Sort(keys)

	for _, k := range keys {
		val := ce.validatorSet[k]
		hasher.Write(val.PubKey)
		binary.Write(hasher, binary.BigEndian, val.Power)
	}

	return hasher.Sum(nil)
}
