package admin

import (
	"bytes"
	"context"
	"encoding/hex"
	"math/big"

	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	admpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/admin/v0"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	types "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/validators"
	"github.com/kwilteam/kwil-db/internal/version"

	cmtCoreTypes "github.com/cometbft/cometbft/rpc/core/types"
	"go.uber.org/zap"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

// BlockchainTransactor specifies the methods required for the admin service to
// interact with the blockchain.
type BlockchainTransactor interface {
	Status(context.Context) (*types.Status, error)
	Peers(context.Context) ([]*types.PeerInfo, error)
	BroadcastTx(ctx context.Context, tx []byte, sync uint8) (*cmtCoreTypes.ResultBroadcastTx, error)
}

type TxApp interface {
	Price(ctx context.Context, tx *transactions.Transaction) (*big.Int, error)
	// AccountInfo returns the unconfirmed account info for the given identifier.
	// If unconfirmed is true, the account found in the mempool is returned.
	// Otherwise, the account found in the blockchain is returned.
	AccountInfo(ctx context.Context, identifier []byte, unconfirmed bool) (balance *big.Int, nonce int64, err error)
}

// ValidatorReader reads data about the validator store.
type ValidatorReader interface {
	CurrentValidators(ctx context.Context) ([]*validators.Validator, error)
	ActiveVotes(ctx context.Context) ([]*validators.JoinRequest, []*validators.ValidatorRemoveProposal, error)
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
	blockchain BlockchainTransactor // node is the local node that can accept transactions.
	TxApp      TxApp
	validators ValidatorReader

	cfg *config.KwildConfig

	log     log.Logger
	chainId string

	signer auth.Signer // signer is an ed25519 signer derived from the nodes private key.
}

var _ admpb.AdminServiceServer = (*Service)(nil)

// NewService constructs a new Service.
func NewService(blockchain BlockchainTransactor, txApp TxApp, validators ValidatorReader, signer auth.Signer, cfg *config.KwildConfig, chainId string, opts ...AdminSvcOpt) *Service {
	s := &Service{
		blockchain: blockchain,
		TxApp:      txApp,
		validators: validators,
		signer:     signer,
		chainId:    chainId,
		cfg:        cfg,
		log:        log.NewNoOp(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
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

func convertValidatorInfo(vi *types.ValidatorInfo) *admpb.Validator {
	return &admpb.Validator{
		Pubkey: vi.PubKey,
		Power:  vi.Power,
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
	status, err := svc.blockchain.Status(ctx)
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
	peers, err := svc.blockchain.Peers(ctx)
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

// sendTx makes a transaction and sends it to the local node.
func (s *Service) sendTx(ctx context.Context, payload transactions.Payload) (*txpb.BroadcastResponse, error) {
	// Get the latest nonce for the account, if it exists.
	_, nonce, err := s.TxApp.AccountInfo(ctx, s.signer.Identity(), true)
	if err != nil {
		return nil, err
	}

	tx, err := transactions.CreateTransaction(payload, s.chainId, uint64(nonce+1))
	if err != nil {
		return nil, err
	}

	fee, err := s.TxApp.Price(ctx, tx)
	if err != nil {
		return nil, err
	}

	tx.Body.Fee = fee

	// Sign the transaction.
	err = tx.Sign(s.signer)
	if err != nil {
		return nil, err
	}

	encodedTx, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// Broadcast the transaction.
	res, err := s.blockchain.BroadcastTx(ctx, encodedTx, 1)
	if err != nil {
		return nil, err
	}
	code, txHash := res.Code, res.Hash.Bytes()

	if txCode := transactions.TxCode(code); txCode != transactions.CodeOk {
		stat := &spb.Status{
			Code:    int32(codes.InvalidArgument),
			Message: "broadcast error",
		}
		if details, err := anypb.New(&txpb.BroadcastErrorDetails{
			Code:    code, // e.g. invalid nonce, wrong chain, etc.
			Hash:    hex.EncodeToString(txHash),
			Message: res.Log,
		}); err != nil {
			s.log.Error("failed to marshal broadcast error details", zap.Error(err))
		} else {
			stat.Details = append(stat.Details, details)
		}
		return nil, status.ErrorProto(stat)
	}

	return &txpb.BroadcastResponse{
		TxHash: txHash,
	}, nil

}

func (s *Service) Approve(ctx context.Context, req *admpb.ApproveRequest) (*txpb.BroadcastResponse, error) {
	return s.sendTx(ctx, &transactions.ValidatorApprove{
		Candidate: req.Pubkey,
	})
}

func (s *Service) Join(ctx context.Context, req *admpb.JoinRequest) (*txpb.BroadcastResponse, error) {

	return s.sendTx(ctx, &transactions.ValidatorJoin{
		Power: 1,
	})
}

func (s *Service) Remove(ctx context.Context, req *admpb.RemoveRequest) (*txpb.BroadcastResponse, error) {
	return s.sendTx(ctx, &transactions.ValidatorRemove{
		Validator: req.Pubkey,
	})
}

func (s *Service) JoinStatus(ctx context.Context, req *admpb.JoinStatusRequest) (*admpb.JoinStatusResponse, error) {
	joiner := req.Pubkey
	allJoins, _, err := s.validators.ActiveVotes(ctx)
	if err != nil {
		s.log.Error("failed to retrieve active join requests", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to retrieve active join requests")
	}
	for _, ji := range allJoins {
		if bytes.Equal(ji.Candidate, joiner) {
			return &admpb.JoinStatusResponse{
				JoinRequest: convertJoinRequest(ji),
			}, nil
		}
	}

	vals, err := s.validators.CurrentValidators(ctx)
	if err != nil {
		s.log.Error("failed to retrieve current validators", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to retrieve current validators")
	}
	for _, vi := range vals {
		if bytes.Equal(vi.PubKey, joiner) {
			return nil, status.Errorf(codes.NotFound, "already a validator") // maybe FailedPrecondition?
		}
	}

	return nil, status.Errorf(codes.NotFound, "no active join request")
}

func convertJoinRequest(join *validators.JoinRequest) *admpb.PendingJoin {
	resp := &admpb.PendingJoin{
		Candidate: join.Candidate,
		Power:     join.Power,
		ExpiresAt: join.ExpiresAt,
		Board:     join.Board,
		Approved:  join.Approved,
	}
	return resp
}

func (s *Service) Leave(ctx context.Context, req *admpb.LeaveRequest) (*txpb.BroadcastResponse, error) {
	return s.sendTx(ctx, &transactions.ValidatorLeave{})
}

func (s *Service) ListValidators(ctx context.Context, req *admpb.ListValidatorsRequest) (*admpb.ListValidatorsResponse, error) {
	vals, err := s.validators.CurrentValidators(ctx)
	if err != nil {
		s.log.Error("failed to retrieve current validators", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to retrieve current validators")
	}

	pbValidators := make([]*admpb.Validator, len(vals))
	for i, vi := range vals {
		pbValidators[i] = &admpb.Validator{
			Pubkey: vi.PubKey,
			Power:  vi.Power,
		}
	}

	return &admpb.ListValidatorsResponse{
		Validators: pbValidators,
	}, nil
}

func (s *Service) ListPendingJoins(ctx context.Context, req *admpb.ListJoinRequestsRequest) (*admpb.ListJoinRequestsResponse, error) {
	joins, _, err := s.validators.ActiveVotes(ctx)
	if err != nil {
		s.log.Error("failed to retrieve active join requests", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to retrieve active join requests")
	}

	pbJoins := make([]*admpb.PendingJoin, len(joins))
	for i, ji := range joins {
		pbJoins[i] = convertJoinRequest(ji)
	}

	return &admpb.ListJoinRequestsResponse{
		JoinRequests: pbJoins,
	}, nil
}

func (s *Service) GetConfig(ctx context.Context, req *admpb.GetConfigRequest) (*admpb.GetConfigResponse, error) {
	bts, err := s.cfg.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &admpb.GetConfigResponse{
		Config: bts,
	}, nil
}
