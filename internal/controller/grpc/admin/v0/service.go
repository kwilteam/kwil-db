package admin

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/pkg/version"
	"github.com/kwilteam/kwil-db/pkg/admin/types"
	"github.com/kwilteam/kwil-db/pkg/log"

	admpb "github.com/kwilteam/kwil-db/api/protobuf/admin/v0"
)

// Node specifies the methods required for the admin service to access
// information from the network node.
type Node interface {
	Status(context.Context) (*types.Status, error)
	Peers(context.Context) ([]*types.PeerInfo, error)
}

type AdminSvcOpt func(*Service)

func WithLogger(logger log.Logger) AdminSvcOpt {
	return func(s *Service) {
		s.log = logger
	}
}

// Service is the implementation of the admpb.AdminServiceServer methods.
type Service struct {
	admpb.UnimplementedAdminServiceServer
	node Node

	log log.Logger
}

// NewService constructs a new Service.
func NewService(node Node, opts ...AdminSvcOpt) *Service {
	s := &Service{
		node: node,
		log:  log.NewNoOp(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Ping responds to any ping request with "pong".
func (svc *Service) Ping(ctx context.Context, req *admpb.PingRequest) (*admpb.PingResponse, error) {
	return &admpb.PingResponse{Message: "pong"}, nil
}

// Version reports the compile-time kwild version.
func (svc *Service) Version(ctx context.Context, req *admpb.VersionRequest) (*admpb.VersionResponse, error) {
	return &admpb.VersionResponse{
		VersionString: version.KwilVersion,
	}, nil
}

func convertNodeInfo(ni *types.NodeInfo) *admpb.NodeInfo {
	return &admpb.NodeInfo{
		ChainId:         ni.ChainID,
		NodeName:        ni.Name,
		NodeId:          ni.NodeID,
		ProtocolVersion: ni.ProtocolVersion,
		AppVersion:      ni.AppVersion,
		BlockVersion:    ni.BlockVersion,
		ListenAddr:      ni.ListenAddr,
		RpcAddr:         ni.RPCAddr,
	}
}

func convertValidatorInfo(vi *types.ValidatorInfo) *admpb.ValidatorInfo {
	return &admpb.ValidatorInfo{
		Pubkey:     vi.PubKey,
		PubkeyType: vi.PubKeyType,
		Power:      vi.Power,
	}
}

func convertSyncInfo(si *types.SyncInfo) *admpb.SyncInfo {
	return &admpb.SyncInfo{
		AppHash:         si.AppHash,
		BestBlockHash:   si.BestBlockHash,
		BestBlockHeight: si.BestBlockHeight,
		BestBlockTime:   si.BestBlockTime.UnixMilli(),
		Syncing:         si.Syncing,
	}
}

func (svc *Service) Status(ctx context.Context, req *admpb.StatusRequest) (*admpb.StatusResponse, error) {
	status, err := svc.node.Status(ctx)
	if err != nil {
		return nil, err
	}
	return &admpb.StatusResponse{
		Node:      convertNodeInfo(status.Node),
		Sync:      convertSyncInfo(status.Sync),
		Validator: convertValidatorInfo(status.Validator),
	}, nil
}

func (svc *Service) Peers(ctx context.Context, req *admpb.PeersRequest) (*admpb.PeersResponse, error) {
	peers, err := svc.node.Peers(ctx)
	if err != nil {
		return nil, err
	}
	pbPeers := make([]*admpb.Peer, len(peers))
	for i, p := range peers {
		pbPeers[i] = &admpb.Peer{
			Node:       convertNodeInfo(p.NodeInfo),
			Inbound:    p.Inbound,
			RemoteAddr: p.RemoteAddr,
		}
	}
	return &admpb.PeersResponse{
		Peers: pbPeers,
	}, nil
}

// Peers(context.Context, *PeersRequest) (*PeersResponse, error)
