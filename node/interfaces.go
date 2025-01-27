package node

import (
	"context"

	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/consensus"
	"github.com/kwilteam/kwil-db/node/snapshotter"
	"github.com/kwilteam/kwil-db/node/types"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

type ConsensusEngine interface {
	Status() *types.NodeStatus // includes: role, inCatchup, consensus params, last commit info and block header

	Role() types.Role
	InCatchup() bool

	AcceptProposal(height int64, blkID, prevBlkID types.Hash, leaderSig []byte, timestamp int64) bool
	NotifyBlockProposal(blk *ktypes.Block)

	AcceptCommit(height int64, blkID types.Hash, ci *types.CommitInfo, leaderSig []byte) bool
	NotifyBlockCommit(blk *ktypes.Block, ci *types.CommitInfo)

	NotifyACK(validatorPK []byte, ack types.AckRes)

	NotifyResetState(height int64, txIDs []types.Hash, senderPubKey []byte)

	NotifyDiscoveryMessage(validatorPK []byte, height int64)

	Start(ctx context.Context, fns consensus.BroadcastFns, peerFns consensus.WhitelistFns) error

	CheckTx(ctx context.Context, tx *ktypes.Transaction) error
	BroadcastTx(ctx context.Context, tx *ktypes.Transaction, sync uint8) (*ktypes.ResultBroadcastTx, error)

	ConsensusParams() *ktypes.NetworkParameters
	CancelBlockExecution(height int64, txIDs []types.Hash) error
}

type BlockProcessor interface {
	GetValidators() []*ktypes.Validator
	SubscribeValidators() <-chan []*ktypes.Validator
}

type SnapshotStore interface {
	Enabled() bool
	GetSnapshot(height uint64, format uint32) *snapshotter.Snapshot
	ListSnapshots() []*snapshotter.Snapshot
	LoadSnapshotChunk(height uint64, format uint32, chunk uint32) ([]byte, error)
}

type DB interface {
	sql.ReadTxMaker
}
