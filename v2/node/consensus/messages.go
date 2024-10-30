package consensus

import (
	"fmt"

	"kwil/node/types"
)

// This file implements the consensus messages that are exchanged between the
// node's p2p receiver and the consensus engine that trigger the state changes.
// There are three types of consensus messages that the node can receive:
// 1. BlockProposal
// 2. Ack
// 3. BlockAnn
// NOTE: only send these messages to the consensus engine if the state machine is
// expecting them.
type consensusMessage struct {
	Sender  []byte
	MsgType string
	Msg     any
}

func (ce *ConsensusEngine) sendConsensusMessage(msg *consensusMessage) {
	// should we validate the msg types?
	ce.msgChan <- *msg
}

// BlockProposal is a message that is sent to the consensus engine to notify
// that a new block proposal has been received from the leader.
// Ensure that the source of the block proposal is the leader.
type blockProposal struct {
	height  int64
	blkHash types.Hash
	blk     *types.Block
	// respCb is a callback function used to send the VoteMessage(ack/nack) back to the leader.
	// respCb func(ack bool, appHash types.Hash) error
}

func (bpm *blockProposal) Type() string {
	return "block_proposal"
}

func (bpm *blockProposal) String() string {
	return fmt.Sprintf("BlockProposal {height: %d, blkHash: %s, prevAppHash: %s}", bpm.height, bpm.blkHash.String(), bpm.blk.Header.PrevAppHash.String())
}

type vote struct {
	ack     bool
	blkHash types.Hash
	appHash *types.Hash
	height  int64
}

func (vm *vote) Type() string {
	return "vote"
}

func (vm *vote) String() string {
	return fmt.Sprintf("Vote {height: %d, ack: %t, blkHash: %s, appHash: %s}", vm.height, vm.ack, vm.blkHash.String(), vm.appHash.String())
}

// BlockAnnounce is a message that is sent to the consensus engine to notify
// that a new block has been committed to the blockchain.
// Ensure that the source of the block announce is the leader.
type blockAnnounce struct {
	appHash types.Hash
	blk     *types.Block
}

func (bam *blockAnnounce) Type() string {
	return "block_ann"
}

func (bam *blockAnnounce) String() string {
	return fmt.Sprintf("BlockAnnounce {height: %d, blkHash: %s, appHash: %s}", bam.blk.Header.Height, bam.blk.Hash().String(), bam.appHash.String())
}

// resetState is a message that is sent to the consensus engine to
// abort any ongoing block proposal at height + 1 and reset to the
// state at height.
// This message can be triggered in the following scenarios:
// 1. Leader explicitly sends a resetState message to the nodes.
// 2. Nodes receive conflicting block proposals from the leader probably
// due to amnesia after leader restart.
// 3. Nodes receive a blockAnn message from the leader for a different blk
// than the one the node is currently processing or waiting on.
type resetState struct {
	height int64
	// ignoreBlk types.Hash
}

func (rs *resetState) Type() string {
	return "reset_state"
}

func (rs *resetState) String() string {
	return fmt.Sprintf("ResetState {height: %d}", rs.height)
}
