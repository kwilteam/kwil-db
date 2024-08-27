package usersvc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	// BlockchainTransactor returns have some big structs from cometbft.
	cmtCoreTypes "github.com/cometbft/cometbft/rpc/core/types" // :(

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/ident"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
	userjson "github.com/kwilteam/kwil-db/core/rpc/json/user"
	"github.com/kwilteam/kwil-db/core/types"
	adminTypes "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/abci"             // errors from chainClient
	"github.com/kwilteam/kwil-db/internal/engine/execution" // errors from engine
	"github.com/kwilteam/kwil-db/internal/migrations"
	rpcserver "github.com/kwilteam/kwil-db/internal/services/jsonrpc"
	"github.com/kwilteam/kwil-db/internal/version"
	"github.com/kwilteam/kwil-db/internal/voting"
	"github.com/kwilteam/kwil-db/parse"
)

// Service is the "user" RPC service, also known as txsvc in other contexts.
type Service struct {
	log           log.Logger
	readTxTimeout time.Duration

	engine      EngineReader
	db          DB              // this should only ever make a read-only tx
	nodeApp     NodeApplication // so we don't have to do ABCIQuery (indirect)
	chainClient BlockchainTransactor
	pricer      Pricer
	migrator    Migrator
}

type DB interface {
	sql.ReadTxMaker
	sql.DelayedReadTxMaker
}

type serviceCfg struct {
	readTxTimeout time.Duration
}

// Opt is a Service option.
type Opt func(*serviceCfg)

// WithReadTxTimeout sets a timeout for read-only DB transactions, as used by
// the Query and Call methods of Service.
func WithReadTxTimeout(timeout time.Duration) Opt {
	return func(cfg *serviceCfg) {
		cfg.readTxTimeout = timeout
	}
}

const defaultReadTxTimeout = 5 * time.Second

// NewService creates a new instance of the user RPC service.
func NewService(db DB, engine EngineReader, chainClient BlockchainTransactor,
	nodeApp NodeApplication, pricer Pricer, migrator Migrator, logger log.Logger, opts ...Opt) *Service {
	cfg := &serviceCfg{
		readTxTimeout: defaultReadTxTimeout,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return &Service{
		log:           logger,
		readTxTimeout: cfg.readTxTimeout,
		engine:        engine,
		nodeApp:       nodeApp,
		pricer:        pricer,
		chainClient:   chainClient,
		db:            db,
		migrator:      migrator,
	}
}

// The "user" service is versioned by these values. However, despite this API
// level versioning, methods can be versioned. For example "user.account.v2".
// The APIs minor version can indicate which new methods (or method versions)
// are available, while the API major version would be bumped for method removal
// or any other breaking changes.
const (
	apiVerUserMajor = 0
	apiVerUserMinor = 1
	apiVerUserPatch = 0
)

var (
	apiVerUserSemver = fmt.Sprintf("%d.%d.%d", apiVerUserMajor, apiVerUserMinor, apiVerUserPatch)
)

// The user Service must be usable as a Svc registered with a JSON-RPC Server.
var _ rpcserver.Svc = (*Service)(nil)

func (svc *Service) Methods() map[jsonrpc.Method]rpcserver.MethodDef {
	return map[jsonrpc.Method]rpcserver.MethodDef{
		userjson.MethodUserVersion: rpcserver.MakeMethodDef(
			verHandler,
			"retrieve the API version of the user service",
			"service info including semver and kwild version",
		),
		userjson.MethodAccount: rpcserver.MakeMethodDef(
			svc.Account,
			"get an account's status",
			"balance and nonce of an accounts",
		),
		userjson.MethodBroadcast: rpcserver.MakeMethodDef(
			svc.Broadcast,
			"broadcast a transaction",
			"the hash of the transaction",
		),
		userjson.MethodCall: rpcserver.MakeMethodDef(
			svc.Call,
			"call an action or procedure",
			"the result of the action/procedure call as a encoded records",
		),
		userjson.MethodChainInfo: rpcserver.MakeMethodDef(
			svc.ChainInfo,
			"get current blockchain info",
			"chain info including chain ID and best block",
		),
		userjson.MethodDatabases: rpcserver.MakeMethodDef(
			svc.ListDatabases,
			"list databases",
			"an array of matching databases",
		),
		userjson.MethodPing: rpcserver.MakeMethodDef(
			svc.Ping,
			"ping the server",
			"a message back from the server",
		),
		userjson.MethodPrice: rpcserver.MakeMethodDef(
			svc.EstimatePrice,
			"estimate the price of a transaction",
			"balance and nonce of an accounts",
		),
		userjson.MethodQuery: rpcserver.MakeMethodDef(
			svc.Query,
			"perform an ad-hoc SQL query",
			"the result of the query as a encoded records",
		),
		userjson.MethodSchema: rpcserver.MakeMethodDef(
			svc.Schema,
			"get a deployed database's kuneiform schema definition",
			"the kuneiform schema",
		),
		userjson.MethodTxQuery: rpcserver.MakeMethodDef(
			svc.TxQuery,
			"query for the status of a transaction",
			"the execution status of a transaction",
		),

		// Migration methods
		userjson.MethodListMigrations: rpcserver.MakeMethodDef(svc.ListPendingMigrations,
			"list active migration resolutions",
			"the list of all the pending migration resolutions",
		),
		userjson.MethodLoadChangesetMetadata: rpcserver.MakeMethodDef(svc.LoadChangesetMetadata,
			"get the changeset metadata for a given height",
			"the changesets metadata for the given height",
		),
		userjson.MethodLoadChangeset: rpcserver.MakeMethodDef(svc.LoadChangeset,
			"load a changeset for a given height and index",
			"the changeset for the given height and index",
		),
		userjson.MethodMigrationMetadata: rpcserver.MakeMethodDef(svc.MigrationMetadata,
			"get the migration information",
			"the metadata for the given migration",
		),
		userjson.MethodMigrationGenesisChunk: rpcserver.MakeMethodDef(svc.MigrationGenesisChunk,
			"get a genesis snapshot chunk of given idx",
			"the genesis chunk for the given index",
		),
	}
}

func verHandler(context.Context, *userjson.VersionRequest) (*userjson.VersionResponse, *jsonrpc.Error) {
	return &userjson.VersionResponse{
		Service:     "user",
		Version:     apiVerUserSemver,
		Major:       apiVerUserMajor,
		Minor:       apiVerUserMinor,
		Patch:       apiVerUserPatch,
		KwilVersion: version.KwilVersion,
	}, nil
}

func (svc *Service) Handlers() map[jsonrpc.Method]rpcserver.MethodHandler {
	handlers := make(map[jsonrpc.Method]rpcserver.MethodHandler)
	for method, def := range svc.Methods() {
		handlers[method] = def.Handler
	}
	return handlers
}

type EngineReader interface {
	Procedure(ctx context.Context, tx sql.DB, options *common.ExecutionData) (*sql.ResultSet, error)
	GetSchema(dbid string) (*types.Schema, error)
	ListDatasets(owner []byte) ([]*types.DatasetIdentifier, error)
	Execute(ctx context.Context, tx sql.DB, dbid string, query string, values map[string]any) (*sql.ResultSet, error)
}

// NOTE:
// with ResultBroadcastTx, we only need Code/Hash/Log
// with ResultTx we need: Tx (a []byte), Hash, Height, and some fields of TxResult

type BlockchainTransactor interface {
	Status(ctx context.Context) (*adminTypes.Status, error)
	BroadcastTx(ctx context.Context, tx []byte, sync uint8) (*cmtCoreTypes.ResultBroadcastTx, error)
	TxQuery(ctx context.Context, hash []byte, prove bool) (*cmtCoreTypes.ResultTx, error)
}

type NodeApplication interface {
	AccountInfo(ctx context.Context, db sql.DB, identifier []byte, getUncommitted bool) (balance *big.Int, nonce int64, err error)
}

type Pricer interface {
	Price(ctx context.Context, db sql.DB, tx *transactions.Transaction) (*big.Int, error)
}

type Migrator interface {
	GetChangesetMetadata(height int64) (*migrations.ChangesetMetadata, error)
	GetChangeset(height int64, index int64) ([]byte, error)
	GetMigrationMetadata() (*types.MigrationMetadata, error)
	GetGenesisSnapshotChunk(height int64, format uint32, chunkIdx uint32) ([]byte, error)
}

func (svc *Service) ChainInfo(ctx context.Context, req *userjson.ChainInfoRequest) (*userjson.ChainInfoResponse, *jsonrpc.Error) {
	status, err := svc.chainClient.Status(ctx)
	if err != nil {
		svc.log.Error("chain status error", log.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorNodeInternal, "status failure", nil)
	}
	return &userjson.ChainInfoResponse{
		ChainID:     status.Node.ChainID,
		BlockHeight: uint64(status.Sync.BestBlockHeight),
		BlockHash:   status.Sync.BestBlockHash,
	}, nil
}

func (svc *Service) Broadcast(ctx context.Context, req *userjson.BroadcastRequest) (*userjson.BroadcastResponse, *jsonrpc.Error) {
	logger := svc.log.With(log.String("rpc", "Broadcast"), // new logger each time, ick
		log.String("PayloadType", req.Tx.Body.PayloadType))
	svc.log.Debug("incoming transaction")

	logger = logger.With(log.String("from", hex.EncodeToString(req.Tx.Sender)))

	// NOTE: it's mostly pointless to have the structured transaction in the
	// request rather than the serialized transaction, except that a client only
	// has to serialize the *body* to sign.
	encodedTx, err := req.Tx.MarshalBinary()
	if err != nil {
		logger.Error("failed to serialize transaction data", log.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorInvalidParams, "failed to serialize transaction data", nil)
	}

	var sync = userjson.BroadcastSyncSync // default to sync, not async or commit
	if req.Sync != nil {
		sync = *req.Sync
	}
	res, err := svc.chainClient.BroadcastTx(ctx, encodedTx, uint8(sync))
	if err != nil {
		logger.Error("failed to broadcast tx", log.Error(err))
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

	logger.Info("broadcast transaction", log.String("TxHash", hex.EncodeToString(txHash)),
		log.Uint("sync", sync), log.Uint("nonce", req.Tx.Body.Nonce))
	return &userjson.BroadcastResponse{
		TxHash: txHash,
	}, nil
}

/* Most broadcast capabilities are bytes, not an object. We should support the following:

type BroadcastRawRequest struct {
	Raw  []byte                 `json:"raw,omitempty"`
	Sync *jsonrpc.BroadcastSync `json:"sync,omitempty"`
}
type BroadcastRawResponse struct {
	TxHash types.HexBytes `json:"tx_hash,omitempty"`
}

func (svc *Service) BroadcastRaw(ctx context.Context, req *BroadcastRawRequest) (*BroadcastRawResponse, *jsonrpc.Error) {
	var sync = jsonrpc.BroadcastSyncSync // default to sync, not async or commit
	if req.Sync != nil {
		sync = *req.Sync
	}
	res, err := svc.chainClient.BroadcastTx(ctx, req.Raw, uint8(sync))
	if err != nil {
		svc.log.Error("failed to broadcast tx", log.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorTxInternal, "failed to broadcast transaction", nil)
	}

	// If we want details, like Sender, Nonce, etc.:
	// var tx transactions.Transaction
	// tx.UnmarshalBinary(req.Raw) //	serialize.Decode(req.Raw, &tx)

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

	svc.log.Info("broadcast transaction", log.String("TxHash", hex.EncodeToString(txHash)), log.Uint("sync", sync))
	return &BroadcastRawResponse{
		TxHash: txHash,
	}, nil
}
*/

func (svc *Service) EstimatePrice(ctx context.Context, req *userjson.EstimatePriceRequest) (*userjson.EstimatePriceResponse, *jsonrpc.Error) {
	svc.log.Debug("Estimating price", log.String("payload_type", req.Tx.Body.PayloadType))
	readTx := svc.db.BeginDelayedReadTx()
	defer readTx.Rollback(ctx)

	price, err := svc.pricer.Price(ctx, readTx, req.Tx)
	if err != nil {
		svc.log.Error("failed to estimate price", log.Error(err)) // why not tell the client though?
		return nil, jsonrpc.NewError(jsonrpc.ErrorTxInternal, "failed to estimate price", nil)
	}

	return &userjson.EstimatePriceResponse{
		Price: price.String(),
	}, nil
}

func (svc *Service) Query(ctx context.Context, req *userjson.QueryRequest) (*userjson.QueryResponse, *jsonrpc.Error) {
	ctxExec, cancel := context.WithTimeout(ctx, svc.readTxTimeout)
	defer cancel()

	readTx := svc.db.BeginDelayedReadTx()
	defer readTx.Rollback(ctx)

	result, err := svc.engine.Execute(ctxExec, readTx, req.DBID, req.Query, nil)
	if err != nil {
		// We don't know for sure that it's an invalid argument, but an invalid
		// user-provided query isn't an internal server error.
		return nil, engineError(err)
	}

	bts, err := json.Marshal(resultMap(result)) // marshalling the map is less efficient, but necessary for backwards compatibility
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorResultEncoding, "failed to marshal call result", nil)
	}

	return &userjson.QueryResponse{
		Result: bts,
	}, nil
}

func (svc *Service) Account(ctx context.Context, req *userjson.AccountRequest) (*userjson.AccountResponse, *jsonrpc.Error) {
	// Status is presently just 0 for confirmed and 1 for pending, but there may
	// be others such as finalized and safe.
	uncommitted := req.Status != nil && *req.Status > 0

	if len(req.Identifier) == 0 {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInvalidParams, "missing account identifier", nil)
	}

	readTx := svc.db.BeginDelayedReadTx()
	defer readTx.Rollback(ctx)

	balance, nonce, err := svc.nodeApp.AccountInfo(ctx, readTx, req.Identifier, uncommitted)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorAccountInternal, "account info error", nil)
	}

	ident := []byte(nil)
	if nonce > 0 { // return nil pubkey for non-existent account
		ident = req.Identifier
	}

	return &userjson.AccountResponse{
		Identifier: ident, // nil for non-existent account
		Nonce:      nonce,
		Balance:    balance.String(),
	}, nil
}

func (svc *Service) Ping(ctx context.Context, req *userjson.PingRequest) (*userjson.PingResponse, *jsonrpc.Error) {
	return &userjson.PingResponse{
		Message: "pong",
	}, nil
}

func (svc *Service) ListDatabases(ctx context.Context, req *userjson.ListDatabasesRequest) (*userjson.ListDatabasesResponse, *jsonrpc.Error) {
	dbs, err := svc.engine.ListDatasets(req.Owner)
	if err != nil {
		svc.log.Error("ListDatasets failed", log.Error(err))
		return nil, engineError(err)
	}

	pbDatasets := make([]*userjson.DatasetInfo, len(dbs))
	for i, db := range dbs {
		pbDatasets[i] = &userjson.DatasetInfo{
			DBID:  db.DBID,
			Name:  db.Name,
			Owner: db.Owner,
		}
	}

	return &userjson.ListDatabasesResponse{
		Databases: pbDatasets,
	}, nil
}

func checkEngineError(err error) (jsonrpc.ErrorCode, string) {
	if err == nil {
		return 0, "" // would not be constructing a jsonrpc.Error
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return jsonrpc.ErrorTimeout, "db timeout"
	}
	if errors.Is(err, execution.ErrDatasetExists) {
		return jsonrpc.ErrorEngineDatasetExists, execution.ErrDatasetExists.Error()
	}
	if errors.Is(err, execution.ErrDatasetNotFound) {
		return jsonrpc.ErrorEngineDatasetNotFound, execution.ErrDatasetNotFound.Error()
	}
	if errors.Is(err, execution.ErrInvalidSchema) {
		return jsonrpc.ErrorEngineInvalidSchema, execution.ErrInvalidSchema.Error()
	}

	return jsonrpc.ErrorEngineInternal, err.Error()
}

func engineError(err error) *jsonrpc.Error {
	if err == nil {
		return nil // would not be constructing a jsonrpc.Error
	}
	code, msg := checkEngineError(err)
	return &jsonrpc.Error{
		Code:    code,
		Message: msg,
	}
}

func (svc *Service) Schema(ctx context.Context, req *userjson.SchemaRequest) (*userjson.SchemaResponse, *jsonrpc.Error) {
	logger := svc.log.With(log.String("rpc", "GetSchema"), log.String("dbid", req.DBID))
	schema, err := svc.engine.GetSchema(req.DBID)
	if err != nil {
		logger.Debug("failed to get schema", log.Error(err))
		return nil, engineError(err)
	}

	return &userjson.SchemaResponse{
		Schema: schema,
	}, nil
}

func convertActionCall(req *userjson.CallRequest) (*transactions.ActionCall, *transactions.CallMessage, error) {
	var actionPayload transactions.ActionCall

	err := actionPayload.UnmarshalBinary(req.Body.Payload)
	if err != nil {
		return nil, nil, err
	}

	return &actionPayload, &transactions.CallMessage{
		Body: &transactions.CallMessageBody{
			Payload: req.Body.Payload,
		},
		AuthType: req.AuthType,
		Sender:   req.Sender,
	}, nil
}

func resultMap(r *sql.ResultSet) []map[string]any {
	m := make([]map[string]any, len(r.Rows))
	for i, row := range r.Rows {
		m2 := make(map[string]any)
		for j, col := range row {
			m2[r.Columns[j]] = col
		}

		m[i] = m2
	}

	return m
}

func (svc *Service) Call(ctx context.Context, req *userjson.CallRequest) (*userjson.CallResponse, *jsonrpc.Error) {
	body, msg, err := convertActionCall(req)
	if err != nil {
		// NOTE: http api needs to be able to get the error message
		return nil, jsonrpc.NewError(jsonrpc.ErrorInvalidParams, "failed to convert action call: "+err.Error(), nil)

	}

	args := make([]any, len(body.Arguments))
	for i, arg := range body.Arguments {
		args[i], err = arg.Decode()
		if err != nil {
			return nil, jsonrpc.NewError(jsonrpc.ErrorInvalidParams, "failed to decode argument: "+err.Error(), nil)
		}
	}

	signer := msg.Sender
	caller := "" // string representation of sender, if signed.  Otherwise, empty string
	if signer != nil && msg.AuthType != "" {
		caller, err = ident.Identifier(msg.AuthType, signer)
		if err != nil {
			return nil, jsonrpc.NewError(jsonrpc.ErrorIdentInvalid, "failed to get caller: "+err.Error(), nil)
		}
	}

	ctxExec, cancel := context.WithTimeout(ctx, svc.readTxTimeout)
	defer cancel()

	// we use a basic read tx since we are subscribing to notices,
	// and it is therefore pointless to use a delayed tx
	readTx, err := svc.db.BeginReadTx(ctx)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorNodeInternal, "failed to start read tx", nil)
	}
	defer readTx.Rollback(ctx)

	logCh, done, err := readTx.Subscribe(ctx)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorNodeInternal, "failed to subscribe to notices", nil)
	}
	defer done(ctx)

	var logs []string
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			select {
			case <-ctxExec.Done():
				wg.Done()
				return
			case logMsg, ok := <-logCh:
				if !ok {
					wg.Done()
					return
				}

				_, notc, err := parse.ParseNotice(logMsg)
				if err != nil {
					svc.log.Error("failed to parse notice", log.Error(err))
					continue
				}

				logs = append(logs, notc)
			}
		}
	}()

	executeResult, err := svc.engine.Procedure(ctxExec, readTx, &common.ExecutionData{
		Dataset:   body.DBID,
		Procedure: body.Action,
		Args:      args,
		TxCtx: &common.TxContext{
			Signer: signer,
			Caller: caller,
			BlockContext: &common.BlockContext{
				Height: -1, // cannot know the height here.
			},
			Authenticator: msg.AuthType,
		},
	})
	if err != nil {
		return nil, engineError(err)
	}

	err = done(ctx)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorNodeInternal, "failed to unsubscribe from notices", nil)
	}

	// marshalling the map is less efficient, but necessary for backwards compatibility
	btsResult, err := json.Marshal(resultMap(executeResult))
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorResultEncoding, "failed to marshal call result", nil)
	}

	wg.Wait()

	return &userjson.CallResponse{
		Result: btsResult,
		Logs:   logs,
	}, nil
}

func (svc *Service) TxQuery(ctx context.Context, req *userjson.TxQueryRequest) (*userjson.TxQueryResponse, *jsonrpc.Error) {
	logger := svc.log.With(log.String("rpc", "TxQuery"),
		log.String("TxHash", hex.EncodeToString(req.TxHash)))

	cmtResult, err := svc.chainClient.TxQuery(ctx, req.TxHash, false)
	if err != nil {
		if errors.Is(err, abci.ErrTxNotFound) {
			return nil, jsonrpc.NewError(jsonrpc.ErrorTxNotFound, "transaction not found", nil)
		}
		logger.Warn("failed to query tx", log.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorNodeInternal, "failed to query transaction", nil)
	}

	//cmtResult.Tx can be nil
	var tx *transactions.Transaction
	if cmtResult.Tx != nil {
		tx = &transactions.Transaction{}
		if err := tx.UnmarshalBinary(cmtResult.Tx); err != nil {
			logger.Error("failed to deserialize transaction", log.Error(err))
			return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "failed to deserialize transaction", nil)
		}
	}

	txResult := &transactions.TransactionResult{
		Code:      cmtResult.TxResult.Code,
		Log:       cmtResult.TxResult.Log,
		GasUsed:   cmtResult.TxResult.GasUsed,
		GasWanted: cmtResult.TxResult.GasWanted,
		//Data: cmtResult.TxResult.Data,
		//Events: cmtResult.TxResult.Events,
	}

	logger.Debug("tx query result", log.Any("result", txResult))

	return &userjson.TxQueryResponse{
		Hash:     cmtResult.Hash.Bytes(),
		Height:   cmtResult.Height,
		Tx:       tx,
		TxResult: txResult,
	}, nil
}

func (svc *Service) LoadChangeset(ctx context.Context, req *userjson.ChangesetRequest) (*userjson.ChangesetsResponse, *jsonrpc.Error) {
	bts, err := svc.migrator.GetChangeset(req.Height, req.Index)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "failed to load changesets", nil)
	}

	return &userjson.ChangesetsResponse{
		Changesets: bts,
	}, nil
}

func (svc *Service) LoadChangesetMetadata(ctx context.Context, req *userjson.ChangesetMetadataRequest) (*userjson.ChangesetMetadataResponse, *jsonrpc.Error) {
	metadata, err := svc.migrator.GetChangesetMetadata(req.Height)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "failed to load changeset metadata", nil)
	}

	return &userjson.ChangesetMetadataResponse{
		Height:     metadata.Height,
		Changesets: metadata.Chunks,
		ChunkSizes: metadata.ChunkSizes,
	}, nil
}

func (svc *Service) MigrationMetadata(ctx context.Context, req *userjson.MigrationMetadataRequest) (*userjson.MigrationMetadataResponse, *jsonrpc.Error) {
	metadata, err := svc.migrator.GetMigrationMetadata()
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, err.Error(), nil)
	}

	return &userjson.MigrationMetadataResponse{
		Metadata: metadata,
	}, nil
}

func (svc *Service) MigrationGenesisChunk(ctx context.Context, req *userjson.MigrationSnapshotChunkRequest) (*userjson.MigrationSnapshotChunkResponse, *jsonrpc.Error) {
	bts, err := svc.migrator.GetGenesisSnapshotChunk(int64(req.Height), 0, req.ChunkIndex)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "failed to load genesis chunk", nil)
	}

	return &userjson.MigrationSnapshotChunkResponse{
		Chunk: bts,
	}, nil
}

func (svc *Service) ListPendingMigrations(ctx context.Context, req *userjson.ListMigrationsRequest) (*userjson.ListMigrationsResponse, *jsonrpc.Error) {
	readTx := svc.db.BeginDelayedReadTx()
	defer readTx.Rollback(ctx)

	resolutions, err := voting.GetResolutionsByType(ctx, readTx, migrations.StartMigrationEventType)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to get migration resolutions", nil)
	}

	var pendingMigrations []*types.Migration

	for _, res := range resolutions {
		mig := &migrations.MigrationDeclaration{}
		if err := mig.UnmarshalBinary(res.Body); err != nil {
			return nil, jsonrpc.NewError(jsonrpc.ErrorInternal, "failed to unmarshal migration declaration", nil)
		}
		pendingMigrations = append(pendingMigrations, &types.Migration{
			ID:               res.ID,
			ActivationPeriod: (int64)(mig.ActivationPeriod),
			Duration:         (int64)(mig.Duration),
			ChainID:          mig.ChainID,
		})
	}

	return &userjson.ListMigrationsResponse{
		Migrations: pendingMigrations,
	}, nil
}
