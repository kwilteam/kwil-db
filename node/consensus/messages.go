package consensus

import (
	"bytes"
	"fmt"

	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
)

// consensusMessageType is the type of messages used to trigger the state changes in the consensus engine.
type consensusMsgType string

const (
	BlockProposal consensusMsgType = "block_proposal"
	BlockAnnounce consensusMsgType = "block_announce"
	Vote          consensusMsgType = "vote"
)

func (mt consensusMsgType) String() string {
	return string(mt)
}

type consensusMessage struct {
	Sender  []byte
	MsgType consensusMsgType
	Msg     any
}

func (ce *ConsensusEngine) sendConsensusMessage(msg *consensusMessage) {
	ce.msgChan <- *msg
}

// BlockProposal is a message that is sent to the consensus engine to notify
// that a new block proposal has been received from the leader.
// Ensure that the source of the block proposal is the leader.
type blockProposal struct {
	height  int64
	blkHash types.Hash
	blk     *ktypes.Block
}

func (bpm *blockProposal) Type() consensusMsgType {
	return BlockProposal
}

func (bpm *blockProposal) String() string {
	return fmt.Sprintf("BlockProposal {height: %d, blkHash: %s, prevAppHash: %s}", bpm.height, bpm.blkHash.String(), bpm.blk.Header.PrevAppHash.String())
}

type vote struct {
	ack            bool
	nackStatus     *types.NackStatus
	blkHash        types.Hash
	appHash        *types.Hash
	height         int64
	outOfSyncProof *types.OutOfSyncProof
	signature      *types.Signature
}

func (vm *vote) Type() consensusMsgType {
	return Vote
}

func (vm *vote) String() string {
	if vm.ack {
		return fmt.Sprintf("Vote {ack: %t, height: %d, blkHash: %s, appHash: %s}",
			vm.ack, vm.height, vm.blkHash, vm.appHash)
	}
	return fmt.Sprintf("Vote {ack: %t, height: %d, blkHash: %s}", vm.ack, vm.height, vm.blkHash)
}

// BlockAnnounce is a message that is sent to the consensus engine to notify
// that a new block has been committed to the blockchain.
// Ensure that the source of the block announce is the leader.
type blockAnnounce struct {
	blk *ktypes.Block
	ci  *types.CommitInfo
}

func (bam *blockAnnounce) Type() consensusMsgType {
	return BlockAnnounce
}

func (bam *blockAnnounce) String() string {
	return fmt.Sprintf("BlockAnnounce {height: %d, blkHash: %s, appHash: %s}", bam.blk.Header.Height, bam.blk.Hash().String(), bam.ci.AppHash.String())
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
func (ce *ConsensusEngine) sendResetMsg(msg *resetMsg) {
	ce.resetChan <- msg
}

// NotifyBlockProposal is used by the p2p stream handler to notify the consensus engine of a block proposal.
// Only a validator should receive block proposals and notify the consensus engine, whereas others should ignore this message.
func (ce *ConsensusEngine) NotifyBlockProposal(blk *ktypes.Block) {
	if ce.role.Load() == types.RoleLeader {
		return
	}

	blkProp := &blockProposal{
		height:  blk.Header.Height,
		blkHash: blk.Header.Hash(),
		blk:     blk,
	}

	go ce.sendConsensusMessage(&consensusMessage{
		MsgType: blkProp.Type(),
		Msg:     blkProp,
		Sender:  ce.leader.Bytes(),
	})
}

// NotifyBlockCommit is used by the p2p stream handler to notify the consensus engine of a committed block.
// Leader should ignore this message.
func (ce *ConsensusEngine) NotifyBlockCommit(blk *ktypes.Block, ci *types.CommitInfo) {
	if ce.role.Load() == types.RoleLeader {
		return
	}

	blkCommit := &blockAnnounce{
		blk: blk,
		ci:  ci,
	}

	go ce.sendConsensusMessage(&consensusMessage{
		MsgType: blkCommit.Type(),
		Msg:     blkCommit,
		Sender:  ce.leader.Bytes(),
	})
}

// NotifyACK notifies the consensus engine about the ACK received from the validator.
func (ce *ConsensusEngine) NotifyACK(validatorPK []byte, ack types.AckRes) {
	if ce.role.Load() != types.RoleLeader {
		return
	}

	voteMsg := &vote{
		ack:            ack.ACK,
		appHash:        ack.AppHash,
		blkHash:        ack.BlkHash,
		height:         ack.Height,
		nackStatus:     ack.NackStatus,
		outOfSyncProof: ack.OutOfSyncProof,
		signature:      ack.Signature,
	}

	ce.sendConsensusMessage(&consensusMessage{
		MsgType: voteMsg.Type(),
		Msg:     voteMsg,
		Sender:  validatorPK,
	})
}

func (v *vote) ToAckRes() *types.AckRes {
	return &types.AckRes{
		ACK:            v.ack,
		AppHash:        v.appHash,
		BlkHash:        v.blkHash,
		Height:         v.height,
		NackStatus:     v.nackStatus,
		OutOfSyncProof: v.outOfSyncProof,
		Signature:      v.signature,
	}
}

type resetMsg struct {
	height int64
	txIDs  []types.Hash
}

// NotifyResetState is used by the p2p stream handler to notify the consensus engine to reset the state to the specified height.
// Only a validator should receive this message to abort the current block execution.
func (ce *ConsensusEngine) NotifyResetState(height int64, txIDs []types.Hash, leaderPubKey []byte) {
	if ce.role.Load() != types.RoleValidator {
		return
	}

	// check if the sender is the leader
	if !bytes.Equal(leaderPubKey, ce.leader.Bytes()) {
		ce.log.Warn("Received reset state message from non-leader", "sender", leaderPubKey)
		return
	}

	go ce.sendResetMsg(&resetMsg{
		height: height,
		txIDs:  txIDs,
	})
}

type discoveryMsg struct {
	BestHeight int64
	Sender     []byte
}

func (ce *ConsensusEngine) NotifyDiscoveryMessage(sender []byte, height int64) {
	if ce.role.Load() != types.RoleLeader {
		return
	}

	dm := &discoveryMsg{
		BestHeight: height,
		Sender:     sender,
	}

	ce.bestHeightCh <- dm
}
