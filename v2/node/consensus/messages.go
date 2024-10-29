package consensus

import (
	"fmt"

	"p2p/node/types"
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

// TODO: do we need a resetState message to rollback or stop processing certain block>
// type resetState struct {
// 	height int64
//  ignoreBlk types.Hash
// }
