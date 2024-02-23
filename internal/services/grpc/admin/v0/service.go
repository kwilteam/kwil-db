package admin

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	admpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/admin/v0"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	types "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/internal/version"
	"github.com/kwilteam/kwil-db/internal/voting"

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
	db         sql.ReadTxMaker

	cfg *config.KwildConfig

	log     log.Logger
	chainId string

	signer auth.Signer // signer is an ed25519 signer derived from the nodes private key.
}

var _ admpb.AdminServiceServer = (*Service)(nil)

// NewService constructs a new Service.
func NewService(db sql.ReadTxMaker, blockchain BlockchainTransactor, txApp TxApp, signer auth.Signer, cfg *config.KwildConfig, chainId string, opts ...AdminSvcOpt) *Service {
	s := &Service{
		blockchain: blockchain,
		TxApp:      txApp,
		signer:     signer,
		chainId:    chainId,
		cfg:        cfg,
		log:        log.NewNoOp(),
		db:         db,
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
	readTx, err := s.db.BeginReadTx(ctx)
	if err != nil {
		s.log.Error("failed to start read transaction", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to start read transaction")
	}
	defer readTx.Rollback(ctx) // always rollback, the readTx is read-only

	ids, err := voting.GetResolutionIDsByTypeAndProposer(ctx, readTx, voting.ValidatorJoinEventType, req.Pubkey)
	if err != nil {
		s.log.Error("failed to retrieve join request", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to retrieve join request")
	}
	if len(ids) == 0 {
		return nil, status.Errorf(codes.NotFound, "no active join request")
	}

	resolution, err := voting.GetResolutionInfo(ctx, readTx, ids[0])
	if err != nil {
		s.log.Error("failed to retrieve join request", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to retrieve join request")
	}

	pendingJoin, err := toPendingInfo(ctx, readTx, resolution)
	if err != nil {
		s.log.Error("failed to convert join request", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to convert join request")
	}

	return &admpb.JoinStatusResponse{
		JoinRequest: pendingJoin,
	}, nil
}

func (s *Service) Leave(ctx context.Context, req *admpb.LeaveRequest) (*txpb.BroadcastResponse, error) {
	return s.sendTx(ctx, &transactions.ValidatorLeave{})
}

func (s *Service) ListValidators(ctx context.Context, req *admpb.ListValidatorsRequest) (*admpb.ListValidatorsResponse, error) {
	readTx, err := s.db.BeginReadTx(ctx)
	if err != nil {
		s.log.Error("failed to start read transaction", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to start read transaction")
	}
	defer readTx.Rollback(ctx) // always rollback, the readTx is read-only

	vals, err := voting.GetValidators(ctx, readTx)
	if err != nil {
		s.log.Error("failed to retrieve voters", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to retrieve voters")
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
	readTx, err := s.db.BeginReadTx(ctx)
	if err != nil {
		s.log.Error("failed to start read transaction", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to start read transaction")
	}
	defer readTx.Rollback(ctx) // always rollback, the readTx is read-only

	activeJoins, err := voting.GetResolutionsByType(ctx, readTx, voting.ValidatorJoinEventType)
	if err != nil {
		s.log.Error("failed to retrieve active join requests", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to retrieve active join requests")
	}

	pbJoins := make([]*admpb.PendingJoin, len(activeJoins))
	for i, ji := range activeJoins {
		pbJoins[i], err = toPendingInfo(ctx, readTx, ji)
		if err != nil {
			s.log.Error("failed to convert join request", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to convert join request")
		}
	}

	return &admpb.ListJoinRequestsResponse{
		JoinRequests: pbJoins,
	}, nil
}

// toPendingInfo gets the pending information for an active join from a resolution
func toPendingInfo(ctx context.Context, db sql.DB, resolution *resolutions.Resolution) (*admpb.PendingJoin, error) {
	resolutionBody := &voting.UpdatePowerRequest{}
	if err := resolutionBody.UnmarshalBinary(resolution.Body); err != nil {
		return nil, fmt.Errorf("failed to unmarshal join request")
	}

	allVoters, err := voting.GetValidators(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve voters")
	}

	// to create the board, we will take a list of all approvers and append the voters.
	// we will then remove any duplicates the second time we see them.
	// this will result with all approvers at the start of the list, and all voters at the end.
	// finally, the approvals will be true for the length of the approvers, and false for found.length - voters.length
	board := make([][]byte, 0, len(allVoters))
	approvals := make([]bool, len(allVoters))
	for i, v := range resolution.Voters {
		board = append(board, v.PubKey)
		approvals[i] = true
	}
	for _, v := range allVoters {
		board = append(board, v.PubKey)
	}

	// we will now remove duplicates from the board.
	found := make(map[string]struct{})
	for i := 0; i < len(board); i++ {
		if _, ok := found[string(board[i])]; ok {
			board = append(board[:i], board[i+1:]...)
			i--
			continue
		}
		found[string(board[i])] = struct{}{}
	}

	return &admpb.PendingJoin{
		Candidate: resolution.Proposer,
		Power:     resolutionBody.Power,
		ExpiresAt: resolution.ExpirationHeight,
		Board:     board,
		Approved:  approvals,
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
