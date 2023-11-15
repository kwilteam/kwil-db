package admin

import (
	"bytes"
	"context"
	"encoding/hex"
	"math/big"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	admpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/admin/v0"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	types "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/validators"
	"github.com/kwilteam/kwil-db/internal/version"
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
	BroadcastTx(ctx context.Context, tx []byte, sync uint8) (code uint32, txHash []byte, err error)
}

// NodeApplication is the abci application that is running on the node.
type NodeApplication interface {
	ChainID() string
	// AccountInfo returns the unconfirmed account info for the given identifier.
	AccountInfo(ctx context.Context, identifier []byte) (balance *big.Int, nonce int64, err error)
}

// ValidatorReader reads data about the validator store.
type ValidatorReader interface {
	CurrentValidators(ctx context.Context) ([]*validators.Validator, error)
	ActiveVotes(ctx context.Context) ([]*validators.JoinRequest, []*validators.ValidatorRemoveProposal, error)
	// JoinStatus(ctx context.Context, joiner []byte) ([]*JoinRequest, error)
	PriceJoin(ctx context.Context) (*big.Int, error)
	PriceLeave(ctx context.Context) (*big.Int, error)
	PriceApprove(ctx context.Context) (*big.Int, error)
	PriceRemove(ctx context.Context) (*big.Int, error)
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
	nodeApp    NodeApplication
	validators ValidatorReader

	log log.Logger

	signer auth.Signer // signer is an ed25519 signer derived from the nodes private key.
}

var _ admpb.AdminServiceServer = (*Service)(nil)

// NewService constructs a new Service.
func NewService(blockchain BlockchainTransactor, node NodeApplication, validators ValidatorReader, signer auth.Signer, opts ...AdminSvcOpt) *Service {
	s := &Service{
		blockchain: blockchain,
		nodeApp:    node,
		validators: validators,
		signer:     signer,
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
func (s *Service) sendTx(ctx context.Context, payload transactions.Payload, price *big.Int) (*txpb.BroadcastResponse, error) {
	// Get the latest nonce for the account, if it exists.
	_, nonce, err := s.nodeApp.AccountInfo(ctx, s.signer.Identity())
	if err != nil {
		return nil, err
	}

	tx, err := transactions.CreateTransaction(payload, s.nodeApp.ChainID(), uint64(nonce+1))
	if err != nil {
		return nil, err
	}

	tx.Body.Fee = price

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
	code, txHash, err := s.blockchain.BroadcastTx(ctx, encodedTx, 1)
	if err != nil {
		return nil, err
	}

	if txCode := transactions.TxCode(code); txCode != transactions.CodeOk {
		stat := &spb.Status{
			Code:    int32(codes.InvalidArgument),
			Message: "broadcast error",
		}
		if details, err := anypb.New(&txpb.BroadcastErrorDetails{
			Code:    code, // e.g. invalid nonce, wrong chain, etc.
			Hash:    hex.EncodeToString(txHash),
			Message: txCode.String(),
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
	price, err := s.validators.PriceApprove(ctx)
	if err != nil {
		return nil, err
	}

	return s.sendTx(ctx, &transactions.ValidatorApprove{
		Candidate: req.Pubkey,
	}, price)
}

func (s *Service) Join(ctx context.Context, req *admpb.JoinRequest) (*txpb.BroadcastResponse, error) {
	price, err := s.validators.PriceJoin(ctx)
	if err != nil {
		return nil, err
	}

	return s.sendTx(ctx, &transactions.ValidatorJoin{
		Power: 1,
	}, price)
}

func (s *Service) Remove(ctx context.Context, req *admpb.RemoveRequest) (*txpb.BroadcastResponse, error) {
	price, err := s.validators.PriceRemove(ctx)
	if err != nil {
		return nil, err
	}

	return s.sendTx(ctx, &transactions.ValidatorRemove{
		Validator: req.Pubkey,
	}, price)
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
	price, err := s.validators.PriceLeave(ctx)
	if err != nil {
		return nil, err
	}

	return s.sendTx(ctx, &transactions.ValidatorLeave{}, price)
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
