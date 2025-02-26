package node

import (
	"context"

	"github.com/kwilteam/kwil-db/core/crypto"
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
	NotifyBlockProposal(blk *ktypes.Block, done func())

	AcceptCommit(height int64, blkID types.Hash, hdr *ktypes.BlockHeader, ci *types.CommitInfo, leaderSig []byte) bool
	NotifyBlockCommit(blk *ktypes.Block, ci *types.CommitInfo, blkID types.Hash, doneFn func())

	NotifyACK(validatorPK []byte, ack types.AckRes)

	NotifyResetState(height int64, txIDs []types.Hash, senderPubKey []byte)

	NotifyDiscoveryMessage(validatorPK []byte, height int64)

	Start(ctx context.Context, fns consensus.BroadcastFns, peerFns consensus.WhitelistFns) error

	QueueTx(ctx context.Context, tx *types.Tx) error
	BroadcastTx(ctx context.Context, tx *types.Tx, sync uint8) (ktypes.Hash, *ktypes.TxResult, error)

	ConsensusParams() *ktypes.NetworkParameters
	CancelBlockExecution(height int64, txIDs []types.Hash) error

	// PromoteLeader is used to promote a validator to leader starting from the specified height
	PromoteLeader(leader crypto.PublicKey, height int64) error
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
