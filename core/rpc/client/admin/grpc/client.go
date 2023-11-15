// package grpc implements a grpc client for the Kwil admin client.
package grpc

import (
	"context"
	"time"

	admpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/admin/v0"
	"github.com/kwilteam/kwil-db/core/types"
	adminTypes "github.com/kwilteam/kwil-db/core/types/admin"
	"google.golang.org/grpc"
)

// GrpcAdminClient is an grpc client for the Kwil admin service.
type GrpcAdminClient struct {
	client admpb.AdminServiceClient
}

// NewAdminClient creates a grpc client for the Kwil admin service.
func NewAdminClient(conn *grpc.ClientConn) *GrpcAdminClient {
	return &GrpcAdminClient{
		client: admpb.NewAdminServiceClient(conn),
	}
}

func (c *GrpcAdminClient) Version(ctx context.Context) (string, error) {
	resp, err := c.client.Version(ctx, &admpb.VersionRequest{})
	if err != nil {
		return "", err
	}
	return resp.VersionString, nil
}

func convertNodeInfo(ni *admpb.NodeInfo) *adminTypes.NodeInfo {
	return &adminTypes.NodeInfo{
		ChainID:         ni.ChainId,
		Name:            ni.NodeName,
		NodeID:          ni.NodeId,
		ProtocolVersion: ni.ProtocolVersion,
		AppVersion:      ni.AppVersion,
		BlockVersion:    ni.BlockVersion,
		ListenAddr:      ni.ListenAddr,
		RPCAddr:         ni.RpcAddr,
	}
}

func (c *GrpcAdminClient) Status(ctx context.Context) (*adminTypes.Status, error) {
	resp, err := c.client.Status(ctx, &admpb.StatusRequest{})
	if err != nil {
		return nil, err
	}
	return &adminTypes.Status{
		Node: convertNodeInfo(resp.Node),
		Sync: &adminTypes.SyncInfo{
			AppHash:         resp.Sync.AppHash,
			BestBlockHash:   resp.Sync.BestBlockHash,
			BestBlockHeight: resp.Sync.BestBlockHeight,
			BestBlockTime:   time.UnixMilli(resp.Sync.BestBlockTime),
			Syncing:         resp.Sync.Syncing,
		},
		Validator: &adminTypes.ValidatorInfo{
			PubKey: resp.Validator.Pubkey,
			Power:  resp.Validator.Power,
		},
	}, nil
}

func (c *GrpcAdminClient) Peers(ctx context.Context) ([]*adminTypes.PeerInfo, error) {
	resp, err := c.client.Peers(ctx, &admpb.PeersRequest{})
	if err != nil {
		return nil, err
	}
	peers := make([]*adminTypes.PeerInfo, len(resp.Peers))
	for i, pbPeer := range resp.Peers {
		peers[i] = &adminTypes.PeerInfo{
			NodeInfo:   convertNodeInfo(pbPeer.Node),
			Inbound:    pbPeer.Inbound,
			RemoteAddr: pbPeer.RemoteAddr,
		}
	}
	return peers, nil
}

// Approve approves a node to join the network.
// It returns a transaction hash.
func (c *GrpcAdminClient) Approve(ctx context.Context, publicKey []byte) ([]byte, error) {
	resp, err := c.client.Approve(ctx, &admpb.ApproveRequest{Pubkey: publicKey})
	if err != nil {
		return nil, err
	}
	return resp.TxHash, nil
}

// Join sends a node join request to the network from the connected node.
// It returns a transaction hash.
func (c *GrpcAdminClient) Join(ctx context.Context) ([]byte, error) {
	resp, err := c.client.Join(ctx, &admpb.JoinRequest{})
	if err != nil {
		return nil, err
	}
	return resp.TxHash, nil
}

// Leave sends a node leave request to the network from the connected node.
// It returns a transaction hash.
func (c *GrpcAdminClient) Leave(ctx context.Context) ([]byte, error) {
	resp, err := c.client.Leave(ctx, &admpb.LeaveRequest{})
	if err != nil {
		return nil, err
	}
	return resp.TxHash, nil
}

// Remove votes to remove a node from the network.
// It returns a transaction hash.
func (c *GrpcAdminClient) Remove(ctx context.Context, publicKey []byte) ([]byte, error) {
	resp, err := c.client.Remove(ctx, &admpb.RemoveRequest{Pubkey: publicKey})
	if err != nil {
		return nil, err
	}
	return resp.TxHash, nil
}

// JoinStatus returns the status of a node's join request.
func (c *GrpcAdminClient) JoinStatus(ctx context.Context, pubkey []byte) (*types.JoinRequest, error) {
	resp, err := c.client.JoinStatus(ctx, &admpb.JoinStatusRequest{Pubkey: pubkey})
	if err != nil {
		return nil, err
	}

	total := len(resp.ApprovedValidators) + len(resp.PendingValidators)
	join := &types.JoinRequest{
		Candidate: pubkey,
		Power:     resp.Power,
		Board:     make([][]byte, 0, total),
		Approved:  make([]bool, 0, total),
		ExpiresAt: resp.Expiration,
	}
	for _, vi := range resp.ApprovedValidators {
		join.Board = append(join.Board, vi)
		join.Approved = append(join.Approved, true) // approved
	}
	for _, vi := range resp.PendingValidators {
		join.Board = append(join.Board, vi)
		join.Approved = append(join.Approved, false) // pending
	}

	return join, nil
}

func (c *GrpcAdminClient) ListValidators(ctx context.Context) ([]*types.Validator, error) {
	resp, err := c.client.ListValidators(ctx, &admpb.ListValidatorsRequest{})
	if err != nil {
		return nil, err
	}
	validators := make([]*types.Validator, len(resp.Validators))
	for i, v := range resp.Validators {
		validators[i] = &types.Validator{
			PubKey: v.Pubkey,
			Power:  v.Power,
		}
	}
	return validators, nil
}
