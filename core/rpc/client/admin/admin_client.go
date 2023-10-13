package admin

import (
	"context"
	"crypto/tls"
	"time"

	admpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/admin/v0"
	types "github.com/kwilteam/kwil-db/core/types/admin"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// AdminClient manages a connection to an authenticated node administrative gRPC
// service.
type AdminClient struct {
	admClient admpb.AdminServiceClient
	conn      *grpc.ClientConn
}

// New constructs an AdminClient with the provided TLS configuration
func New(target string, tlsCfg *tls.Config, opts ...grpc.DialOption) (*AdminClient, error) {
	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}
	return &AdminClient{
		admClient: admpb.NewAdminServiceClient(conn),
		conn:      conn,
	}, nil
}

func (c *AdminClient) Close() error {
	return c.conn.Close()
}

func (c *AdminClient) Ping(ctx context.Context) (string, error) {
	resp, err := c.admClient.Ping(ctx, &admpb.PingRequest{})
	if err != nil {
		return "", err
	}
	return resp.Message, nil
}

func (c *AdminClient) Version(ctx context.Context) (string, error) {
	resp, err := c.admClient.Version(ctx, &admpb.VersionRequest{})
	if err != nil {
		return "", err
	}
	return resp.VersionString, nil
}

func convertNodeInfo(ni *admpb.NodeInfo) *types.NodeInfo {
	return &types.NodeInfo{
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

func (c *AdminClient) Status(ctx context.Context) (*types.Status, error) {
	resp, err := c.admClient.Status(ctx, &admpb.StatusRequest{})
	if err != nil {
		return nil, err
	}
	return &types.Status{
		Node: convertNodeInfo(resp.Node),
		Sync: &types.SyncInfo{
			AppHash:         resp.Sync.AppHash,
			BestBlockHash:   resp.Sync.BestBlockHash,
			BestBlockHeight: resp.Sync.BestBlockHeight,
			BestBlockTime:   time.UnixMilli(resp.Sync.BestBlockTime),
			Syncing:         resp.Sync.Syncing,
		},
		Validator: &types.ValidatorInfo{
			PubKey:     resp.Validator.Pubkey,
			PubKeyType: resp.Validator.PubkeyType,
			Power:      resp.Validator.Power,
		},
	}, nil
}

func (c *AdminClient) Peers(ctx context.Context) ([]*types.PeerInfo, error) {
	resp, err := c.admClient.Peers(ctx, &admpb.PeersRequest{})
	if err != nil {
		return nil, err
	}
	peers := make([]*types.PeerInfo, len(resp.Peers))
	for i, pbPeer := range resp.Peers {
		peers[i] = &types.PeerInfo{
			NodeInfo:   convertNodeInfo(pbPeer.Node),
			Inbound:    pbPeer.Inbound,
			RemoteAddr: pbPeer.RemoteAddr,
		}
	}
	return peers, nil
}
