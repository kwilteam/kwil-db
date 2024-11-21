package node

import (
	"context"

	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/consensus"
	"github.com/kwilteam/kwil-db/node/types"
	"github.com/libp2p/go-libp2p/core/network"
)

type ConsensusEngine interface {
	AcceptProposal(height int64, blkID, prevBlkID types.Hash, leaderSig []byte, timestamp int64) bool
	NotifyBlockProposal(blk *types.Block)

	AcceptCommit(height int64, blkID types.Hash, appHash types.Hash, leaderSig []byte) bool
	NotifyBlockCommit(blk *types.Block, appHash types.Hash)

	NotifyACK(validatorPK []byte, ack types.AckRes)
	NotifyResetState(height int64)

	// Gonna remove this once we have the commit results such as app hash and the tx results stored in the block store.

	Start(ctx context.Context, proposerBroadcaster consensus.ProposalBroadcaster,
		blkAnnouncer consensus.BlkAnnouncer, ackBroadcaster consensus.AckBroadcaster,
		blkRequester consensus.BlkRequester, stateResetter consensus.ResetStateBroadcaster)

	// Note: Not sure if these are needed here, just for separate of concerns:
	// p2p stream handlers role is to download the messages and pass it to the
	// respective modules to process it and we probably should not be triggering any consensus
	// affecting methods.

	// ProcessProposal(blk *types.Block, cb func(ack bool, appHash types.Hash) error)
	// ProcessACK(validatorPK []byte, ack types.AckRes)
	// CommitBlock(blk *types.Block, appHash types.Hash) error
}

type PeerManager interface {
	network.Notifiee
	Start(context.Context) error
	ConnectedPeers() []types.PeerInfo
	KnownPeers() []types.PeerInfo
}
type TxApp interface {
	SubscribeValidators() <-chan []*ktypes.Validator
}
