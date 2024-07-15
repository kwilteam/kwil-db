// package admin specifies the interface for the admin service client.
package admin

import (
	"context"

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

	AddPeer(ctx context.Context, peerID string) error
	RemovePeer(ctx context.Context, peerID string) error
	ListPeers(ctx context.Context) ([]string, error)

	// Resolutions
	CreateResolution(ctx context.Context, resolution []byte, resolutionType string) ([]byte, error)
	ApproveResolution(ctx context.Context, resolutionID *types.UUID) ([]byte, error)
	DeleteResolution(ctx context.Context, resolutionID *types.UUID) ([]byte, error)
	ResolutionStatus(ctx context.Context, resolutionID *types.UUID) (*types.PendingResolution, error)
}
