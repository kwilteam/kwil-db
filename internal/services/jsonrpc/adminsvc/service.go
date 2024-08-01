package adminsvc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/kwilteam/kwil-db/cmd/kwild/config"
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
	// AccountInfo returns the unconfirmed account info for the given identifier.
	// If unconfirmed is true, the account found in the mempool is returned.
	// Otherwise, the account found in the blockchain is returned.
	AccountInfo(ctx context.Context, db sql.DB, identifier []byte, unconfirmed bool) (balance *big.Int, nonce int64, err error)
}

type Pricer interface {
	Price(ctx context.Context, db sql.DB, tx *transactions.Transaction) (*big.Int, error)
}

type Migrator interface {
	GetChangesetMetadata(height int64) (*migrations.ChangesetMetdata, error)
	GetChangeset(height int64, index int64) ([]byte, error)
	GetMigrationMetadata() (*migrations.MigrationMetadata, error)
	GetGenesisSnapshotChunk(height int64, format uint32, chunkIdx uint32) ([]byte, error)
}

type Service struct {
	log log.Logger

	blockchain BlockchainTransactor // node is the local node that can accept transactions.
	TxApp      TxApp
	db         sql.DelayedReadTxMaker
	pricer     Pricer
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

		// Migration methods
		adminjson.MethodTriggerMigration: rpcserver.MakeMethodDef(svc.TriggerMigration,
			"create a migration resolution",
			"the hash of the broadcasted migration transaction",
		),
		adminjson.MethodApproveMigration: rpcserver.MakeMethodDef(svc.ApproveMigration,
			"approve a migration resolution",
			"the hash of the broadcasted migration approval transaction",
		),
		adminjson.MethodMigrationStatus: rpcserver.MakeMethodDef(svc.MigrationStatus,
			"get the status of a migration resolution",
			"the status of the migration resolution",
		),
		adminjson.MethodListMigrations: rpcserver.MakeMethodDef(svc.ListPendingMigrations,
			"list active migration resolutions",
			"the list of all the pending migration resolutions",
		),
		adminjson.MethodLoadChangesetMetadata: rpcserver.MakeMethodDef(svc.LoadChangesetMetadata,
			"get the changeset metadata for a given height",
			"the changesets metadata for the given height",
		),
		adminjson.MethodLoadChangeset: rpcserver.MakeMethodDef(svc.LoadChangeset,
			"load a changeset for a given height and index",
			"the changeset for the given height and index",
		),
		adminjson.MethodMigrationMetadata: rpcserver.MakeMethodDef(svc.MigrationMetadata,
			"get the migration information",
			"the metadata for the given migration",
		),
		adminjson.MethodMigrationGenesisChunk: rpcserver.MakeMethodDef(svc.MigrationGenesisChunk,
			"get a genesis snapshot chunk of given idx",
			"the genesis chunk for the given index",
		),
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
func NewService(db sql.DelayedReadTxMaker, blockchain BlockchainTransactor, txApp TxApp, pricer Pricer, migrator Migrator, signer auth.Signer, cfg *config.KwildConfig,
	chainID string, logger log.Logger) *Service {
	return &Service{
		blockchain: blockchain,
		TxApp:      txApp,
		signer:     signer,
		chainID:    chainID,
		pricer:     pricer,
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

	expiresAt, board, approvals, err := resolutionStatus(ctx, db, resolution)
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

func (svc *Service) LoadChangeset(ctx context.Context, req *adminjson.ChangesetRequest) (*adminjson.ChangesetsResponse, *jsonrpc.Error) {
	bts, err := svc.migrator.GetChangeset(req.Height, req.Index)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "failed to load changesets", nil)
	}

	return &adminjson.ChangesetsResponse{
		Changesets: bts,
	}, nil
}

func (svc *Service) LoadChangesetMetadata(ctx context.Context, req *adminjson.ChangesetMetadataRequest) (*adminjson.ChangesetMetadataResponse, *jsonrpc.Error) {
	metadata, err := svc.migrator.GetChangesetMetadata(req.Height)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "failed to load changeset metadata", nil)
	}

	return &adminjson.ChangesetMetadataResponse{
		Height:        metadata.Height,
		Changesets:    metadata.Chunks,
		ChangesetSize: metadata.ChangesetSize,
	}, nil
}

func (svc *Service) MigrationMetadata(ctx context.Context, req *adminjson.MigrationMetadataRequest) (*adminjson.MigrationMetadataResponse, *jsonrpc.Error) {
	metadata, err := svc.migrator.GetMigrationMetadata()
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, fmt.Sprintf("failed to load migration metadata: %s", err.Error()), nil)
	}

	bts, err := metadata.MarshalBinary()
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorResultEncoding, "failed to encode migration metadata", nil)
	}

	return &adminjson.MigrationMetadataResponse{
		InMigration: true,
		Metadata:    bts,
	}, nil

}

func (svc *Service) MigrationGenesisChunk(ctx context.Context, req *adminjson.MigrationSnapshotChunkRequest) (*adminjson.MigrationSnapshotChunkResponse, *jsonrpc.Error) {
	bts, err := svc.migrator.GetGenesisSnapshotChunk(int64(req.Height), 0, req.ChunkIndex)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "failed to load genesis chunk", nil)
	}

	return &adminjson.MigrationSnapshotChunkResponse{
		Chunk: bts,
	}, nil
}

func (svc *Service) TriggerMigration(ctx context.Context, req *adminjson.TriggerMigrationRequest) (*userjson.BroadcastResponse, *jsonrpc.Error) {
	timestamp := time.Now().GoString()

	migrationEvt := &migrations.MigrationDeclaration{
		ActivationPeriod: req.Migration.ActivationHeight,
		Duration:         req.Migration.MigrationDuration,
		ChainID:          req.Migration.ChainID,
		Timestamp:        timestamp,
	}

	bts, err := migrationEvt.MarshalBinary()
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "failed to encode migration declaration", nil)
	}

	res := &transactions.CreateResolution{
		Resolution: &transactions.VotableEvent{
			Type: migrations.StartMigrationEventType,
			Body: bts,
		},
	}

	return svc.sendTx(ctx, res)
}

func (svc *Service) ApproveMigration(ctx context.Context, req *adminjson.ApproveMigrationRequest) (*userjson.BroadcastResponse, *jsonrpc.Error) {
	uuid, err := coretypes.ParseUUID(req.Id)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInvalidParams, "invalid migration ID", nil)
	}

	res := &transactions.VoteResolution{
		ResolutionType: migrations.StartMigrationEventType,
		ResolutionID:   uuid,
	}

	return svc.sendTx(ctx, res)
}

func (svc *Service) ListPendingMigrations(ctx context.Context, req *adminjson.ListMigrationsRequest) (*adminjson.ListMigrationsResponse, *jsonrpc.Error) {
	readTx := svc.db.BeginDelayedReadTx()
	defer readTx.Rollback(ctx)

	resolutions, err := voting.GetResolutionsByType(ctx, readTx, migrations.StartMigrationEventType)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to get migration resolutions", nil)
	}

	var pendingMigrations []*coretypes.Migration

	for _, res := range resolutions {
		mig := &migrations.MigrationDeclaration{}
		if err := mig.UnmarshalBinary(res.Body); err != nil {
			return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "failed to unmarshal migration declaration", nil)
		}
		pendingMigrations = append(pendingMigrations, &coretypes.Migration{
			ID:                res.ID.String(),
			ActivationHeight:  mig.ActivationPeriod,
			MigrationDuration: mig.Duration,
			ChainID:           mig.ChainID,
			Timestamp:         mig.Timestamp,
		})
	}

	return &adminjson.ListMigrationsResponse{
		Migrations: pendingMigrations,
	}, nil
}

func (svc *Service) MigrationStatus(ctx context.Context, req *adminjson.MigrationStatusRequest) (*adminjson.MigrationStatusResponse, *jsonrpc.Error) {
	uuid, err := coretypes.ParseUUID(req.Id)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInvalidParams, "invalid migration ID", nil)
	}

	readTx := svc.db.BeginDelayedReadTx()
	defer readTx.Rollback(ctx)

	resolution, err := voting.GetResolutionInfo(ctx, readTx, uuid)
	if err != nil {
		svc.log.Error("failed to retrieve migration resolution", zap.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to retrieve migration proposal status, it might have already been approved.", nil)
	}

	expiresAt, board, approvals, err := resolutionStatus(ctx, readTx, resolution)
	if err != nil {
		svc.log.Error("failed to retrieve migration resolution status", zap.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to retrieve migration proposal status", nil)
	}

	return &adminjson.MigrationStatusResponse{
		Status: &coretypes.MigrationStatus{
			Proposal:  req.Id,
			ExpiresAt: expiresAt,
			Board:     board,
			Approved:  approvals,
		},
	}, nil
}

func resolutionStatus(ctx context.Context, db sql.DB, resolution *resolutions.Resolution) (expiresAt int64, board [][]byte, approvals []bool, err error) {
	allVoters, err := voting.GetValidators(ctx, db)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to retrieve voters")
	}

	// to create the board, we will take a list of all approvers and append the voters.
	// we will then remove any duplicates the second time we see them.
	// this will result with all approvers at the start of the list, and all voters at the end.
	// finally, the approvals will be true for the length of the approvers, and false for found.length - voters.length
	board = make([][]byte, 0, len(allVoters))
	approvals = make([]bool, len(allVoters))
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

	return resolution.ExpirationHeight, board, approvals, nil
}
