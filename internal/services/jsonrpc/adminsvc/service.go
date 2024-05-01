package adminsvc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
	adminjson "github.com/kwilteam/kwil-db/core/rpc/json/admin"
	types "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	rpcserver "github.com/kwilteam/kwil-db/internal/services/jsonrpc"
	"github.com/kwilteam/kwil-db/internal/version"
	"github.com/kwilteam/kwil-db/internal/voting"

	cmtCoreTypes "github.com/cometbft/cometbft/rpc/core/types"
	"go.uber.org/zap"
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

type Service struct {
	log log.Logger

	blockchain BlockchainTransactor // node is the local node that can accept transactions.
	TxApp      TxApp
	db         sql.ReadTxMaker

	cfg     *config.KwildConfig
	chainID string
	signer  auth.Signer // ed25519 signer derived from the node's private key
}

const (
	apiVerMajor = 0
	apiVerMinor = 1
	apiVerPatch = 0
)

var (
	apiSemver = fmt.Sprintf("%d.%d.%d", apiVerMajor, apiVerMinor, apiVerPatch)
)

// The admin Service must be usable as a Svc registered with a JSON-RPC Server.
var _ rpcserver.Svc = (*Service)(nil)

func (svc *Service) Handlers() map[jsonrpc.Method]rpcserver.MethodHandler {
	return map[jsonrpc.Method]rpcserver.MethodHandler{
		adminjson.MethodVersion: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &jsonrpc.VersionRequest{}
			return req, func() (any, *jsonrpc.Error) {
				return &jsonrpc.VersionResponse{
					Service:     "user",
					Version:     apiSemver,
					Major:       apiVerMajor,
					Minor:       apiVerMinor,
					Patch:       apiVerPatch,
					KwilVersion: version.KwilVersion,
				}, nil
			}
		},
		adminjson.MethodStatus: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &adminjson.StatusRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.Status(ctx, req) }
		},
		adminjson.MethodPeers: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &adminjson.PeersRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.Peers(ctx, req) }
		},
		adminjson.MethodConfig: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &adminjson.GetConfigRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.GetConfig(ctx, req) }
		},
		adminjson.MethodValApprove: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &adminjson.ApproveRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.Approve(ctx, req) }
		},
		adminjson.MethodValJoin: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &adminjson.JoinRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.Join(ctx, req) }
		},
		adminjson.MethodValJoinStatus: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &adminjson.JoinStatusRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.JoinStatus(ctx, req) }
		},
		adminjson.MethodValRemove: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &adminjson.RemoveRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.Remove(ctx, req) }
		},
		adminjson.MethodValLeave: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &adminjson.LeaveRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.Leave(ctx, req) }
		},
		adminjson.MethodValList: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &adminjson.ListValidatorsRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.ListValidators(ctx, req) }
		},
		adminjson.MethodValListJoins: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &adminjson.ListJoinRequestsRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.ListPendingJoins(ctx, req) }
		},
	}
}

// NewService constructs a new Service.
func NewService(db sql.ReadTxMaker, blockchain BlockchainTransactor, txApp TxApp, signer auth.Signer, cfg *config.KwildConfig,
	chainID string, logger log.Logger) *Service {
	return &Service{
		blockchain: blockchain,
		TxApp:      txApp,
		signer:     signer,
		chainID:    chainID,
		cfg:        cfg,
		log:        logger,
		db:         db,
	}
}

func convertSyncInfo(si *types.SyncInfo) *adminjson.SyncInfo {
	return &adminjson.SyncInfo{
		AppHash:         si.AppHash,
		BestBlockHash:   si.BestBlockHash,
		BestBlockHeight: si.BestBlockHeight,
		BestBlockTime:   si.BestBlockTime.UnixMilli(), // this is why we dup this
		Syncing:         si.Syncing,
	}
}

func (svc *Service) Status(ctx context.Context, req *adminjson.StatusRequest) (*adminjson.StatusResponse, *jsonrpc.Error) {
	status, err := svc.blockchain.Status(ctx)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorNodeInternal, "node status unavailable", nil)
	}
	return &adminjson.StatusResponse{
		Node: status.Node,
		Sync: convertSyncInfo(status.Sync),
		Validator: &adminjson.Validator{ // TODO: weed out the type dups
			PubKey: status.Validator.PubKey,
			Power:  status.Validator.Power,
		},
	}, nil
}

func (svc *Service) Peers(ctx context.Context, _ *adminjson.PeersRequest) (*adminjson.PeersResponse, *jsonrpc.Error) {
	peers, err := svc.blockchain.Peers(ctx)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorNodeInternal, "node peers unavailable", nil)
	}
	// pbPeers := make([]*types.PeerInfo, len(peers))
	// for i, p := range peers {
	// 	pbPeers[i] = &types.PeerInfo{
	// 		NodeInfo:   p.NodeInfo,
	// 		Inbound:    p.Inbound,
	// 		RemoteAddr: p.RemoteAddr,
	// 	}
	// }
	return &adminjson.PeersResponse{
		Peers: peers,
	}, nil
}

// sendTx makes a transaction and sends it to the local node.
func (svc *Service) sendTx(ctx context.Context, payload transactions.Payload) (*jsonrpc.BroadcastResponse, *jsonrpc.Error) {
	// Get the latest nonce for the account, if it exists.
	_, nonce, err := svc.TxApp.AccountInfo(ctx, svc.signer.Identity(), true)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorAccountInternal, "account info error", nil)
	}

	tx, err := transactions.CreateTransaction(payload, svc.chainID, uint64(nonce+1))
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "unable to create transaction", nil)
	}

	fee, err := svc.TxApp.Price(ctx, tx)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorTxInternal, "unable to price transaction", nil)
	}

	tx.Body.Fee = fee

	// Sign the transaction.
	err = tx.Sign(svc.signer)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "signing transaction failed", nil)
	}
	encodedTx, err := tx.MarshalBinary()
	if err != nil {
		svc.log.Error("failed to serialize transaction data", log.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorInvalidParams, "failed to serialize transaction data", nil)
	}

	res, err := svc.blockchain.BroadcastTx(ctx, encodedTx, uint8(jsonrpc.BroadcastSyncSync))
	if err != nil {
		svc.log.Error("failed to broadcast tx", log.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorTxInternal, "failed to broadcast transaction", nil)
	}

	code, txHash := res.Code, res.Hash.Bytes()

	if txCode := transactions.TxCode(code); txCode != transactions.CodeOk {
		errData := &jsonrpc.BroadcastError{
			TxCode:  txCode.Uint32(), // e.g. invalid nonce, wrong chain, etc.
			Hash:    hex.EncodeToString(txHash),
			Message: res.Log,
		}
		data, _ := json.Marshal(errData)
		return nil, jsonrpc.NewError(jsonrpc.ErrorTxExecFailure, "broadcast error", data)
	}

	svc.log.Info("broadcast transaction", log.String("TxHash", hex.EncodeToString(txHash)), log.Uint("nonce", tx.Body.Nonce))
	return &jsonrpc.BroadcastResponse{
		TxHash: txHash,
	}, nil

}

func (svc *Service) Approve(ctx context.Context, req *adminjson.ApproveRequest) (*jsonrpc.BroadcastResponse, *jsonrpc.Error) {
	return svc.sendTx(ctx, &transactions.ValidatorApprove{
		Candidate: req.PubKey,
	})
}

func (svc *Service) Join(ctx context.Context, req *adminjson.JoinRequest) (*jsonrpc.BroadcastResponse, *jsonrpc.Error) {
	return svc.sendTx(ctx, &transactions.ValidatorJoin{
		Power: 1,
	})
}

func (svc *Service) Remove(ctx context.Context, req *adminjson.RemoveRequest) (*jsonrpc.BroadcastResponse, *jsonrpc.Error) {
	return svc.sendTx(ctx, &transactions.ValidatorRemove{
		Validator: req.PubKey,
	})
}

func (svc *Service) JoinStatus(ctx context.Context, req *adminjson.JoinStatusRequest) (*adminjson.JoinStatusResponse, *jsonrpc.Error) {
	readTx, err := svc.db.BeginReadTx(ctx)
	if err != nil {
		svc.log.Error("failed to start read transaction", zap.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to start read transaction", nil)
	}
	defer readTx.Rollback(ctx) // always rollback, the readTx is read-only

	ids, err := voting.GetResolutionIDsByTypeAndProposer(ctx, readTx, voting.ValidatorJoinEventType, req.PubKey)
	if err != nil {
		svc.log.Error("failed to retrieve join request", zap.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to retrieve join request", nil)
	}
	if len(ids) == 0 {
		return nil, jsonrpc.NewError(jsonrpc.ErrorValidatorNotFound, "no active join request", nil)
	}

	resolution, err := voting.GetResolutionInfo(ctx, readTx, ids[0])
	if err != nil {
		svc.log.Error("failed to retrieve join request", zap.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to retrieve join request details", nil)
	}

	pendingJoin, err := toPendingInfo(ctx, readTx, resolution)
	if err != nil {
		svc.log.Error("failed to convert join request", zap.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorResultEncoding, "failed to convert join request", nil)
	}

	return &adminjson.JoinStatusResponse{
		JoinRequest: pendingJoin,
	}, nil
}

func (svc *Service) Leave(ctx context.Context, req *adminjson.LeaveRequest) (*jsonrpc.BroadcastResponse, *jsonrpc.Error) {
	return svc.sendTx(ctx, &transactions.ValidatorLeave{})
}

func (svc *Service) ListValidators(ctx context.Context, req *adminjson.ListValidatorsRequest) (*adminjson.ListValidatorsResponse, *jsonrpc.Error) {
	readTx, err := svc.db.BeginReadTx(ctx)
	if err != nil {
		svc.log.Error("failed to start read transaction", zap.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to start read transaction", nil)
	}
	defer readTx.Rollback(ctx) // always rollback, the readTx is read-only

	vals, err := voting.GetValidators(ctx, readTx)
	if err != nil {
		svc.log.Error("failed to retrieve voters", zap.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to retrieve voters", nil)
	}

	pbValidators := make([]*adminjson.Validator, len(vals))
	for i, vi := range vals {
		pbValidators[i] = &adminjson.Validator{
			PubKey: vi.PubKey,
			Power:  vi.Power,
		}
	}

	return &adminjson.ListValidatorsResponse{
		Validators: pbValidators,
	}, nil
}

func (svc *Service) ListPendingJoins(ctx context.Context, req *adminjson.ListJoinRequestsRequest) (*adminjson.ListJoinRequestsResponse, *jsonrpc.Error) {
	readTx, err := svc.db.BeginReadTx(ctx)
	if err != nil {
		svc.log.Error("failed to start read transaction", zap.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to start read transaction", nil)
	}
	defer readTx.Rollback(ctx) // always rollback, the readTx is read-only

	activeJoins, err := voting.GetResolutionsByType(ctx, readTx, voting.ValidatorJoinEventType)
	if err != nil {
		svc.log.Error("failed to retrieve active join requests", zap.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to retrieve active join requests", nil)
	}

	pbJoins := make([]*adminjson.PendingJoin, len(activeJoins))
	for i, ji := range activeJoins {
		pbJoins[i], err = toPendingInfo(ctx, readTx, ji)
		if err != nil {
			svc.log.Error("failed to convert join request", zap.Error(err))
			return nil, jsonrpc.NewError(jsonrpc.ErrorResultEncoding, "failed to convert join request", nil)
		}
	}

	return &adminjson.ListJoinRequestsResponse{
		JoinRequests: pbJoins,
	}, nil
}

// toPendingInfo gets the pending information for an active join from a resolution
func toPendingInfo(ctx context.Context, db sql.DB, resolution *resolutions.Resolution) (*adminjson.PendingJoin, error) {
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

	return &adminjson.PendingJoin{
		Candidate: resolution.Proposer,
		Power:     resolutionBody.Power,
		ExpiresAt: resolution.ExpirationHeight,
		Board:     board,
		Approved:  approvals,
	}, nil
}

func (svc *Service) GetConfig(ctx context.Context, req *adminjson.GetConfigRequest) (*adminjson.GetConfigResponse, *jsonrpc.Error) {
	bts, err := svc.cfg.MarshalBinary()
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorResultEncoding, "failed to encode node config", nil)
	}

	return &adminjson.GetConfigResponse{
		Config: bts,
	}, nil
}
