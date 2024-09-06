package adminsvc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	cmtCoreTypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/kwilteam/kwil-db/common/config"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
	adminjson "github.com/kwilteam/kwil-db/core/rpc/json/admin"
	userjson "github.com/kwilteam/kwil-db/core/rpc/json/user"
	coretypes "github.com/kwilteam/kwil-db/core/types"
	types "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/internal/migrations"
	rpcserver "github.com/kwilteam/kwil-db/internal/services/jsonrpc"
	"github.com/kwilteam/kwil-db/internal/version"
	"github.com/kwilteam/kwil-db/internal/voting"
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
	// AccountInfo returns the unconfirmed account info for the given identifier.
	// If unconfirmed is true, the account found in the mempool is returned.
	// Otherwise, the account found in the blockchain is returned.
	AccountInfo(ctx context.Context, db sql.DB, identifier []byte, unconfirmed bool) (balance *big.Int, nonce int64, err error)
}

type Pricer interface {
	Price(ctx context.Context, db sql.DB, tx *transactions.Transaction) (*big.Int, error)
}

type P2P interface {
	// AddPeer adds a peer to the node's peer list and persists it.
	AddAndPersistPeer(ctx context.Context, nodeID string) error
	// RemovePeer removes a peer from the node's peer list permanently.
	RemovePersistedPeer(ctx context.Context, nodeID string) error
	// ListPeers returns the list of peers in the node's whitelist.
	ListPeers(ctx context.Context) []string
}

type Migrator interface {
	GetChangesetMetadata(height int64) (*migrations.ChangesetMetadata, error)
	GetChangeset(height int64, index int64) ([]byte, error)
	GetMigrationMetadata() (*coretypes.MigrationMetadata, error)
	GetGenesisSnapshotChunk(height int64, format uint32, chunkIdx uint32) ([]byte, error)
}

type Service struct {
	log log.Logger

	blockchain BlockchainTransactor // node is the local node that can accept transactions.
	TxApp      TxApp
	db         sql.DelayedReadTxMaker
	pricer     Pricer
	p2p        P2P
	migrator   Migrator

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

func verHandler(context.Context, *userjson.VersionRequest) (*userjson.VersionResponse, *jsonrpc.Error) {
	return &userjson.VersionResponse{
		Service:     "user",
		Version:     apiSemver,
		Major:       apiVerMajor,
		Minor:       apiVerMinor,
		Patch:       apiVerPatch,
		KwilVersion: version.KwilVersion,
	}, nil
}

// The admin Service must be usable as a Svc registered with a JSON-RPC Server.
var _ rpcserver.Svc = (*Service)(nil)

func (svc *Service) Methods() map[jsonrpc.Method]rpcserver.MethodDef {
	return map[jsonrpc.Method]rpcserver.MethodDef{
		adminjson.MethodVersion: rpcserver.MakeMethodDef(verHandler,
			"retrieve the API version of the admin service",    // method description
			"service info including semver and kwild version"), // return value description
		adminjson.MethodStatus: rpcserver.MakeMethodDef(svc.Status,
			"retrieve node status",
			"node information including name, chain id, sync, identity, etc."),
		adminjson.MethodPeers: rpcserver.MakeMethodDef(svc.Peers,
			"get the current peers of the node",
			"a list of the node's current peers"),
		adminjson.MethodConfig: rpcserver.MakeMethodDef(svc.GetConfig,
			"retrieve the current effective node config",
			"the raw bytes of the effective config TOML document"),
		adminjson.MethodValApprove: rpcserver.MakeMethodDef(svc.Approve,
			"approve a validator join request",
			"the hash of the broadcasted validator approve transaction"),
		adminjson.MethodValJoin: rpcserver.MakeMethodDef(svc.Join,
			"request the node to become a validator",
			"the hash of the broadcasted validator join transaction"),
		adminjson.MethodValJoinStatus: rpcserver.MakeMethodDef(svc.JoinStatus,
			"query for the status of a validator join request",
			"the pending join request details, if it exists"),
		adminjson.MethodValListJoins: rpcserver.MakeMethodDef(svc.ListPendingJoins,
			"list active validator join requests",
			"all pending join requests including the current approvals and the join expiry"),
		adminjson.MethodValList: rpcserver.MakeMethodDef(svc.ListValidators,
			"list the current validators",
			"the list of current validators and their power"),
		adminjson.MethodValLeave: rpcserver.MakeMethodDef(svc.Leave,
			"leave the validator set",
			"the hash of the broadcasted validator leave transaction"),
		adminjson.MethodValRemove: rpcserver.MakeMethodDef(svc.Remove,
			"vote to remote a validator",
			"the hash of the broadcasted validator remove transaction"),
		adminjson.MethodAddPeer: rpcserver.MakeMethodDef(svc.AddPeer,
			"add a peer to the network", ""),
		adminjson.MethodRemovePeer: rpcserver.MakeMethodDef(svc.RemovePeer,
			"add a peer to the network",
			""),
		adminjson.MethodListPeers: rpcserver.MakeMethodDef(svc.ListPeers,
			"list the peers from the node's whitelist",
			"the list of peers from which the node can accept connections from."),
		adminjson.MethodCreateResolution: rpcserver.MakeMethodDef(svc.CreateResolution,
			"create a resolution",
			"the hash of the broadcasted create resolution transaction",
		),
		adminjson.MethodApproveResolution: rpcserver.MakeMethodDef(svc.ApproveResolution,
			"approve a resolution",
			"the hash of the broadcasted approve resolution transaction",
		),
		adminjson.MethodDeleteResolution: rpcserver.MakeMethodDef(svc.DeleteResolution,
			"delete a resolution",
			"the hash of the broadcasted delete resolution transaction",
		),
		adminjson.MethodResolutionStatus: rpcserver.MakeMethodDef(svc.ResolutionStatus,
			"get the status of a resolution",
			"the status of the resolution"),
	}
}

func (svc *Service) Handlers() map[jsonrpc.Method]rpcserver.MethodHandler {
	handlers := make(map[jsonrpc.Method]rpcserver.MethodHandler)
	for method, def := range svc.Methods() {
		handlers[method] = def.Handler
	}
	return handlers
}

// NewService constructs a new Service.
func NewService(db sql.DelayedReadTxMaker, blockchain BlockchainTransactor, txApp TxApp,
	pricer Pricer, p2p P2P, migrator Migrator, signer auth.Signer, cfg *config.KwildConfig,
	chainID string, logger log.Logger) *Service {
	return &Service{
		blockchain: blockchain,
		TxApp:      txApp,
		signer:     signer,
		chainID:    chainID,
		pricer:     pricer,
		p2p:        p2p,
		migrator:   migrator,
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
func (svc *Service) sendTx(ctx context.Context, payload transactions.Payload) (*userjson.BroadcastResponse, *jsonrpc.Error) {
	readTx := svc.db.BeginDelayedReadTx()
	defer readTx.Rollback(ctx)

	// Get the latest nonce for the account, if it exists.
	_, nonce, err := svc.TxApp.AccountInfo(ctx, readTx, svc.signer.Identity(), true)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorAccountInternal, "account info error", nil)
	}

	tx, err := transactions.CreateTransaction(payload, svc.chainID, uint64(nonce+1))
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "unable to create transaction", nil)
	}

	fee, err := svc.pricer.Price(ctx, readTx, tx)
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

	res, err := svc.blockchain.BroadcastTx(ctx, encodedTx, uint8(userjson.BroadcastSyncSync))
	if err != nil {
		svc.log.Error("failed to broadcast tx", log.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorTxInternal, "failed to broadcast transaction", nil)
	}

	code, txHash := res.Code, res.Hash.Bytes()

	if txCode := transactions.TxCode(code); txCode != transactions.CodeOk {
		errData := &userjson.BroadcastError{
			TxCode:  txCode.Uint32(), // e.g. invalid nonce, wrong chain, etc.
			Hash:    hex.EncodeToString(txHash),
			Message: res.Log,
		}
		data, _ := json.Marshal(errData)
		return nil, jsonrpc.NewError(jsonrpc.ErrorTxExecFailure, "broadcast error", data)
	}

	svc.log.Info("broadcast transaction", log.String("TxHash", hex.EncodeToString(txHash)), log.Uint("nonce", tx.Body.Nonce))
	return &userjson.BroadcastResponse{
		TxHash: txHash,
	}, nil

}

func (svc *Service) Approve(ctx context.Context, req *adminjson.ApproveRequest) (*userjson.BroadcastResponse, *jsonrpc.Error) {
	return svc.sendTx(ctx, &transactions.ValidatorApprove{
		Candidate: req.PubKey,
	})
}

func (svc *Service) Join(ctx context.Context, req *adminjson.JoinRequest) (*userjson.BroadcastResponse, *jsonrpc.Error) {
	return svc.sendTx(ctx, &transactions.ValidatorJoin{
		Power: 1,
	})
}

func (svc *Service) Remove(ctx context.Context, req *adminjson.RemoveRequest) (*userjson.BroadcastResponse, *jsonrpc.Error) {
	return svc.sendTx(ctx, &transactions.ValidatorRemove{
		Validator: req.PubKey,
	})
}

func (svc *Service) JoinStatus(ctx context.Context, req *adminjson.JoinStatusRequest) (*adminjson.JoinStatusResponse, *jsonrpc.Error) {
	readTx := svc.db.BeginDelayedReadTx()
	defer readTx.Rollback(ctx)
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

func (svc *Service) Leave(ctx context.Context, req *adminjson.LeaveRequest) (*userjson.BroadcastResponse, *jsonrpc.Error) {
	return svc.sendTx(ctx, &transactions.ValidatorLeave{})
}

func (svc *Service) ListValidators(ctx context.Context, req *adminjson.ListValidatorsRequest) (*adminjson.ListValidatorsResponse, *jsonrpc.Error) {
	readTx := svc.db.BeginDelayedReadTx()
	defer readTx.Rollback(ctx)
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
	readTx := svc.db.BeginDelayedReadTx()
	defer readTx.Rollback(ctx)

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

	expiresAt, board, approvals, err := voting.ResolutionStatus(ctx, db, resolution)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve join request status")
	}

	return &adminjson.PendingJoin{
		Candidate: resolution.Proposer,
		Power:     resolutionBody.Power,
		ExpiresAt: expiresAt,
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

func (svc *Service) AddPeer(ctx context.Context, req *adminjson.PeerRequest) (*adminjson.PeerResponse, *jsonrpc.Error) {
	err := svc.p2p.AddAndPersistPeer(ctx, req.PeerID)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "failed to add a peer. Reason: "+err.Error(), nil)
	}
	return &adminjson.PeerResponse{}, nil
}

func (svc *Service) RemovePeer(ctx context.Context, req *adminjson.PeerRequest) (*adminjson.PeerResponse, *jsonrpc.Error) {
	fmt.Println("RemovePeer : ", req.PeerID)
	err := svc.p2p.RemovePersistedPeer(ctx, req.PeerID)
	if err != nil {
		svc.log.Error("failed to remove peer", zap.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "failed to remove peer : "+err.Error(), nil)
	}
	return &adminjson.PeerResponse{}, nil
}

func (svc *Service) ListPeers(ctx context.Context, req *adminjson.PeersRequest) (*adminjson.ListPeersResponse, *jsonrpc.Error) {
	return &adminjson.ListPeersResponse{
		Peers: svc.p2p.ListPeers(ctx),
	}, nil
}

func (svc *Service) CreateResolution(ctx context.Context, req *adminjson.CreateResolutionRequest) (*userjson.BroadcastResponse, *jsonrpc.Error) {
	res := &transactions.CreateResolution{
		Resolution: &transactions.VotableEvent{
			Type: req.ResolutionType,
			Body: req.Resolution,
		},
	}

	return svc.sendTx(ctx, res)
}

func (svc *Service) ApproveResolution(ctx context.Context, req *adminjson.ApproveResolutionRequest) (*userjson.BroadcastResponse, *jsonrpc.Error) {
	res := &transactions.ApproveResolution{
		ResolutionID: req.ResolutionID,
	}

	return svc.sendTx(ctx, res)
}

func (svc *Service) DeleteResolution(ctx context.Context, req *adminjson.DeleteResolutionRequest) (*userjson.BroadcastResponse, *jsonrpc.Error) {
	res := &transactions.DeleteResolution{
		ResolutionID: req.ResolutionID,
	}

	return svc.sendTx(ctx, res)
}

func (svc *Service) ResolutionStatus(ctx context.Context, req *adminjson.ResolutionStatusRequest) (*adminjson.ResolutionStatusResponse, *jsonrpc.Error) {
	readTx := svc.db.BeginDelayedReadTx()
	defer readTx.Rollback(ctx)

	uuid := req.ResolutionID
	resolution, err := voting.GetResolutionInfo(ctx, readTx, uuid)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to retrieve resolution", nil)
	}

	// get the status of the resolution
	expiresAt, board, approvals, err := voting.ResolutionStatus(ctx, readTx, resolution)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to retrieve resolution status", nil)
	}

	return &adminjson.ResolutionStatusResponse{
		Status: &coretypes.PendingResolution{
			ResolutionID: req.ResolutionID,
			ExpiresAt:    expiresAt,
			Board:        board,
			Approved:     approvals,
		},
	}, nil
}
