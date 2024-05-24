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
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
	"github.com/kwilteam/kwil-db/core/types"
	adminTypes "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/abci"             // errors from chainClient
	"github.com/kwilteam/kwil-db/internal/engine/execution" // errors from engine
	"github.com/kwilteam/kwil-db/internal/ident"
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

func (svc *Service) Handlers() map[jsonrpc.Method]rpcserver.MethodHandler {
	return map[jsonrpc.Method]rpcserver.MethodHandler{
		jsonrpc.MethodUserVersion: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &jsonrpc.VersionRequest{}
			return req, func() (any, *jsonrpc.Error) {
				return &jsonrpc.VersionResponse{
					Service:     "user",
					Version:     apiVerUserSemver,
					Major:       apiVerUserMajor,
					Minor:       apiVerUserMinor,
					Patch:       apiVerUserPatch,
					KwilVersion: version.KwilVersion,
				}, nil
			}
		},
		jsonrpc.MethodAccount: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &jsonrpc.AccountRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.Account(ctx, req) }
		},
		jsonrpc.MethodBroadcast: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &jsonrpc.BroadcastRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.Broadcast(ctx, req) }
		},
		jsonrpc.MethodCall: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &jsonrpc.CallRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.Call(ctx, req) }
		},
		jsonrpc.MethodChainInfo: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &jsonrpc.ChainInfoRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.ChainInfo(ctx, req) }
		},
		jsonrpc.MethodDatabases: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &jsonrpc.ListDatabasesRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.ListDatabases(ctx, req) }
		},
		jsonrpc.MethodPing: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &jsonrpc.PingRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.Ping(ctx, req) }
		},
		jsonrpc.MethodPrice: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &jsonrpc.EstimatePriceRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.EstimatePrice(ctx, req) }
		},
		jsonrpc.MethodQuery: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &jsonrpc.QueryRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.Query(ctx, req) }
		},
		jsonrpc.MethodSchema: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &jsonrpc.SchemaRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.Schema(ctx, req) }
		},
		jsonrpc.MethodTxQuery: func(ctx context.Context, s *rpcserver.Server) (any, func() (any, *jsonrpc.Error)) {
			req := &jsonrpc.TxQueryRequest{}
			return req, func() (any, *jsonrpc.Error) { return svc.TxQuery(ctx, req) }
		},
	}
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

func (svc *Service) ChainInfo(ctx context.Context, req *jsonrpc.ChainInfoRequest) (*jsonrpc.ChainInfoResponse, *jsonrpc.Error) {
	status, err := svc.chainClient.Status(ctx)
	if err != nil {
		svc.log.Error("chain status error", log.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorNodeInternal, "status failure", nil)
	}
	return &jsonrpc.ChainInfoResponse{
		ChainID:     status.Node.ChainID,
		BlockHeight: uint64(status.Sync.BestBlockHeight),
		BlockHash:   status.Sync.BestBlockHash,
	}, nil
}

func (svc *Service) Broadcast(ctx context.Context, req *jsonrpc.BroadcastRequest) (*jsonrpc.BroadcastResponse, *jsonrpc.Error) {
	logger := svc.log.With(log.String("rpc", "Broadcast"), // new logger each time, ick
		log.String("PayloadType", req.Tx.Body.PayloadType))
	svc.log.Debug("incoming transaction")

	logger = logger.With(log.String("from", hex.EncodeToString(req.Tx.Sender)))

	// NOTE: it's mostly pointless to have the structured transaction in the
	// request rather than the serialized transaction.
	encodedTx, err := req.Tx.MarshalBinary()
	if err != nil {
		logger.Error("failed to serialize transaction data", log.Error(err))
		return nil, jsonrpc.NewError(jsonrpc.ErrorInvalidParams, "failed to serialize transaction data", nil)
	}

	var sync = jsonrpc.BroadcastSyncSync // default to sync, not async or commit
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
		errData := &jsonrpc.BroadcastError{
			TxCode:  txCode.Uint32(), // e.g. invalid nonce, wrong chain, etc.
			Hash:    hex.EncodeToString(txHash),
			Message: res.Log,
		}
		data, _ := json.Marshal(errData)
		return nil, jsonrpc.NewError(jsonrpc.ErrorTxExecFailure, "broadcast error", data)
	}

	logger.Info("broadcast transaction", log.String("TxHash", hex.EncodeToString(txHash)),
		log.Uint("sync", sync), log.Uint("nonce", req.Tx.Body.Nonce))
	return &jsonrpc.BroadcastResponse{
		TxHash: txHash,
	}, nil
}

func (svc *Service) EstimatePrice(ctx context.Context, req *jsonrpc.EstimatePriceRequest) (*jsonrpc.EstimatePriceResponse, *jsonrpc.Error) {
	svc.log.Debug("Estimating price", log.String("payload_type", req.Tx.Body.PayloadType))

	price, err := svc.nodeApp.Price(ctx, req.Tx)
	if err != nil {
		svc.log.Error("failed to estimate price", log.Error(err)) // why not tell the client though?
		return nil, jsonrpc.NewError(jsonrpc.ErrorTxInternal, "failed to estimate price", nil)
	}

	return &jsonrpc.EstimatePriceResponse{
		Price: price.String(),
	}, nil
}

func (svc *Service) Query(ctx context.Context, req *jsonrpc.QueryRequest) (*jsonrpc.QueryResponse, *jsonrpc.Error) {
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

	return &jsonrpc.QueryResponse{
		Result: bts,
	}, nil
}

func (svc *Service) Account(ctx context.Context, req *jsonrpc.AccountRequest) (*jsonrpc.AccountResponse, *jsonrpc.Error) {
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

	return &jsonrpc.AccountResponse{
		Identifier: ident, // nil for non-existent account
		Nonce:      nonce,
		Balance:    balance.String(),
	}, nil
}

func (svc *Service) Ping(ctx context.Context, req *jsonrpc.PingRequest) (*jsonrpc.PingResponse, *jsonrpc.Error) {
	return &jsonrpc.PingResponse{
		Message: "pong",
	}, nil
}

func (svc *Service) ListDatabases(ctx context.Context, req *jsonrpc.ListDatabasesRequest) (*jsonrpc.ListDatabasesResponse, *jsonrpc.Error) {
	dbs, err := svc.engine.ListDatasets(req.Owner)
	if err != nil {
		svc.log.Error("ListDatasets failed", log.Error(err))
		return nil, engineError(err)
	}

	pbDatasets := make([]*jsonrpc.DatasetInfo, len(dbs))
	for i, db := range dbs {
		pbDatasets[i] = &jsonrpc.DatasetInfo{
			DBID:  db.DBID,
			Name:  db.Name,
			Owner: db.Owner,
		}
	}

	return &jsonrpc.ListDatabasesResponse{
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

func (svc *Service) Schema(ctx context.Context, req *jsonrpc.SchemaRequest) (*jsonrpc.SchemaResponse, *jsonrpc.Error) {
	logger := svc.log.With(log.String("rpc", "GetSchema"), log.String("dbid", req.DBID))
	schema, err := svc.engine.GetSchema(req.DBID)
	if err != nil {
		logger.Debug("failed to get schema", log.Error(err))
		return nil, engineError(err)
	}

	return &jsonrpc.SchemaResponse{
		Schema: schema,
	}, nil
}

func convertActionCall(req *jsonrpc.CallRequest) (*transactions.ActionCall, *transactions.CallMessage, error) {
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

func (svc *Service) Call(ctx context.Context, req *jsonrpc.CallRequest) (*jsonrpc.CallResponse, *jsonrpc.Error) {
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
			Signer: signer,
			Caller: caller,
			Height: -1, // not available
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

	return &jsonrpc.CallResponse{
		Result: btsResult,
	}, nil
}

func (svc *Service) TxQuery(ctx context.Context, req *jsonrpc.TxQueryRequest) (*jsonrpc.TxQueryResponse, *jsonrpc.Error) {
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

	return &jsonrpc.TxQueryResponse{
		Hash:     cmtResult.Hash.Bytes(),
		Height:   cmtResult.Height,
		Tx:       tx,
		TxResult: txResult,
	}, nil
}
