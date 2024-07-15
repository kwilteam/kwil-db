// package admin specifies the interface for the admin service client.
package admin

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types"
	adminTypes "github.com/kwilteam/kwil-db/core/types/admin"
)

type AdminClient interface {
	Approve(ctx context.Context, publicKey []byte) ([]byte, error)
	Join(ctx context.Context) ([]byte, error)
	JoinStatus(ctx context.Context, pubkey []byte) (*types.JoinRequest, error)
	Leave(ctx context.Context) ([]byte, error)
	ListValidators(ctx context.Context) ([]*types.Validator, error)
	Peers(ctx context.Context) ([]*adminTypes.PeerInfo, error)
	Remove(ctx context.Context, publicKey []byte) ([]byte, error)
	Status(ctx context.Context) (*adminTypes.Status, error)
	Version(ctx context.Context) (string, error)
	ListPendingJoins(ctx context.Context) ([]*types.JoinRequest, error)

	// GetConfig gets the current config from the node.
	// It returns the config serialized as JSON.
	GetConfig(ctx context.Context) ([]byte, error)

	// Migrations
	TriggerMigration(ctx context.Context, activationHeight *big.Int, migrationDuration *big.Int, chainID string) ([]byte, error)
	ApproveMigration(ctx context.Context, id string) ([]byte, error)
	ListMigrations(ctx context.Context) ([]*types.Migration, error)

	// Active Migration State
	GenesisState(ctx context.Context) (bool, []byte, error)
	GenesisSnapshotChunk(ctx context.Context, height uint64, chunkIdx uint32) ([]byte, error)

	// Changesets
	LoadChangeset(ctx context.Context, height int64, index int64) ([]byte, error)
	ChangesetMetadata(ctx context.Context, height int64) (int64, int64, error)
}
