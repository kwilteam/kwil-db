package usersvc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
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
	rpcserver "github.com/kwilteam/kwil-db/internal/services/jsonrpc"
	"github.com/kwilteam/kwil-db/internal/version"
)

// Service is the "user" RPC service, also known as txsvc in other contexts.
type Service struct {
	log           log.Logger
	readTxTimeout time.Duration

	engine      EngineReader
	db          sql.ReadTxMaker // this should only ever make a read-only tx
	nodeApp     NodeApplication // so we don't have to do ABCIQuery (indirect)
	chainClient BlockchainTransactor
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
func NewService(db sql.ReadTxMaker, engine EngineReader, chainClient BlockchainTransactor,
	nodeApp NodeApplication, logger log.Logger, opts ...Opt) *Service {
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
		chainClient:   chainClient,
		db:            db,
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
	AccountInfo(ctx context.Context, identifier []byte, getUncommitted bool) (balance *big.Int, nonce int64, err error)
	Price(ctx context.Context, tx *transactions.Transaction) (*big.Int, error)
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

	price, err := svc.nodeApp.Price(ctx, req.Tx)
	if err != nil {
		svc.log.Error("failed to estimate price", log.Error(err)) // why not tell the client though?
		return nil, jsonrpc.NewError(jsonrpc.ErrorTxInternal, "failed to estimate price", nil)
	}

	return &userjson.EstimatePriceResponse{
		Price: price.String(),
	}, nil
}

func (svc *Service) Query(ctx context.Context, req *userjson.QueryRequest) (*userjson.QueryResponse, *jsonrpc.Error) {
	tx, err := svc.db.BeginReadTx(ctx)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to create read tx", nil)
	}
	defer tx.Rollback(ctx)

	ctxExec, cancel := context.WithTimeout(ctx, svc.readTxTimeout)
	defer cancel()

	result, err := svc.engine.Execute(ctxExec, tx, req.DBID, req.Query, nil)
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

	balance, nonce, err := svc.nodeApp.AccountInfo(ctx, req.Identifier, uncommitted)
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

	tx, err := svc.db.BeginReadTx(ctx)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorDBInternal, "failed to create read tx", nil)
	}
	defer tx.Rollback(ctx)

	ctxExec, cancel := context.WithTimeout(ctx, svc.readTxTimeout)
	defer cancel()

	executeResult, err := svc.engine.Procedure(ctxExec, tx, &common.ExecutionData{
		Dataset:   body.DBID,
		Procedure: body.Action,
		Args:      args,
		TransactionData: common.TransactionData{
			Signer:        signer,
			Caller:        caller,
			Height:        -1, // not available
			Authenticator: msg.AuthType,
		},
	})
	if err != nil {
		return nil, engineError(err)
	}

	// marshalling the map is less efficient, but necessary for backwards compatibility
	btsResult, err := json.Marshal(resultMap(executeResult))
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorResultEncoding, "failed to marshal call result", nil)
	}

	return &userjson.CallResponse{
		Result: btsResult,
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
