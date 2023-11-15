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
}
