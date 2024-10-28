package consensus

import (
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

type vote struct {
	ack     bool
	blkHash types.Hash
	appHash *types.Hash
	height  int64
}

func (vm *vote) Type() string {
	return "vote"
}

type blockAnnounce struct {
	appHash types.Hash
	blk     *types.Block
}

func (bam *blockAnnounce) Type() string {
	return "block_ann"
}
