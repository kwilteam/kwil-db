package abci

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/core/utils"
	engineTypes "github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/ident"
	"github.com/kwilteam/kwil-db/internal/kv"
	modDataset "github.com/kwilteam/kwil-db/internal/modules/datasets"
	modVal "github.com/kwilteam/kwil-db/internal/modules/validators"
	"github.com/kwilteam/kwil-db/internal/validators"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tendermintTypes "github.com/cometbft/cometbft/proto/tendermint/types"
	cmtTypes "github.com/cometbft/cometbft/types"
	"go.uber.org/zap"
)

// FatalError is a type that can be used in an explicit panic so that the nature
// of the failure may bubble up through the cometbft Node to the top level
// kwil server type.
type FatalError struct {
	AppMethod string
	Request   fmt.Stringer // entire request for debugging
	Message   string
}

// appState is an in-memory representation of the state of the application.
type appState struct {
	height  int64
	appHash []byte
}

func (fe FatalError) String() string {
	return fmt.Sprintf("Application Method: %s\nError: %s\nRequest (%T): %v",
		fe.AppMethod, fe.Message, fe.Request, fe.Request)
}

func newFatalError(method string, request fmt.Stringer, message string) FatalError {
	if request == nil {
		request = nilStringer{}
	}

	return FatalError{
		AppMethod: method,
		Request:   request,
		Message:   message,
	}
}

type nilStringer struct{}

func (ds nilStringer) String() string {
	return "no message"
}

// AbciConfig includes data that defines the chain and allow the application to
// satisfy the ABCI Application interface.
type AbciConfig struct {
	GenesisAppHash     []byte
	ChainID            string
	ApplicationVersion uint64
}

func NewAbciApp(cfg *AbciConfig, accounts AccountsModule, database DatasetsModule, vldtrs ValidatorModule, kv KVStore,
	committer AtomicCommitter, snapshotter SnapshotModule, bootstrapper DBBootstrapModule, opts ...AbciOpt) *AbciApp {
	app := &AbciApp{
		cfg:        *cfg,
		database:   database,
		validators: vldtrs,
		committer:  committer,
		metadataStore: &metadataStore{
			kv: kv,
		},
		bootstrapper: bootstrapper,
		snapshotter:  snapshotter,
		accounts:     accounts,

		valAddrToKey: make(map[string][]byte),

		mempool: &mempool{
			accounts:     make(map[string]*userAccount),
			accountStore: accounts,
		},

		log: log.NewNoOp(),
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

// pubkeyToAddr converts an Ed25519 public key as used to identify nodes in
// CometBFT into an address, which for ed25519 in comet is an upper case
// truncated sha256 hash of the pubkey. For secp256k1, they do like BTC with
// RIPEMD160(SHA256(pubkey)).  If we support both (if either), we'll need a type
// flag.
func pubkeyToAddr(pubkey []byte) (string, error) {
	if len(pubkey) != ed25519.PubKeySize {
		return "", errors.New("invalid public key")
	}
	publicKey := ed25519.PubKey(pubkey)
	return publicKey.Address().String(), nil
}

type AbciApp struct {
	cfg AbciConfig

	// database is the database module that handles database deployment, dropping, and execution
	database DatasetsModule

	// validators is the validators module that handles joining and approving validators
	validators ValidatorModule
	// comet punishes by address, so we maintain an address=>pubkey map.
	valAddrToKey map[string][]byte // NOTE: includes candidates
	// Validator updates obtained in EndBlock, applied to valAddrToKey in Commit
	valUpdates []*validators.Validator

	// committer is the atomic committer that handles atomic commits across multiple stores
	committer AtomicCommitter

	// snapshotter is the snapshotter module that handles snapshotting
	snapshotter SnapshotModule

	// bootstrapper is the bootstrapper module that handles bootstrapping the database
	bootstrapper DBBootstrapModule

	// metadataStore to track the app hash and block height
	metadataStore *metadataStore

	// accountStore is the store that maintains the account state
	accounts AccountsModule

	// mempool maintains in-memory account state to validate the unconfirmed transactions against.
	mempool *mempool

	log log.Logger

	// Expected AppState after bootstrapping the node with a given snapshot,
	// state gets updated with the bootupState after bootstrapping
	bootupState appState

	// blockHeight is the current block height
	blockHeight int64
}

func (a *AbciApp) ChainID() string {
	return a.cfg.ChainID
}

// AccountInfo gets the pending balance and nonce for an account. If there are
// no unconfirmed transactions for the account, the latest confirmed values are
// returned.
func (a *AbciApp) AccountInfo(ctx context.Context, identifier []byte) (balance *big.Int, nonce int64, err error) {
	// If we have any unconfirmed transactions for the user, report that info
	// without even checking the account store.
	ua := a.mempool.peekAccountInfo(ctx, identifier)
	if ua.nonce > 0 {
		return ua.balance, ua.nonce, nil
	}
	// Nothing in mempool, check account store. Changes to the account store are
	// committed before the mempool is cleared, so there should not be any race.
	acct, err := a.accounts.GetAccount(ctx, identifier)
	if err != nil {
		return nil, 0, err
	}
	return acct.Balance, acct.Nonce, nil
}

var _ abciTypes.Application = &AbciApp{}

// BeginBlock begins a block.
// If the previous commit is not finished, it will wait for the previous commit to finish.
func (a *AbciApp) BeginBlock(req abciTypes.RequestBeginBlock) abciTypes.ResponseBeginBlock {
	logger := a.log.With(zap.String("stage", "ABCI BeginBlock"), zap.Int("height", int(req.Header.Height)))
	logger.Debug("begin block")

	ctx := context.Background()

	idempotencyKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(idempotencyKey, uint64(req.Header.Height))
	a.blockHeight = req.Header.Height

	err := a.committer.Begin(ctx, idempotencyKey)
	if err != nil {
		panic(newFatalError("BeginBlock", &req, err.Error()))
	}

	// Punish bad validators.
	for _, ev := range req.ByzantineValidators {
		addr := string(ev.Validator.Address) // comet example app confirms this conversion... weird
		// if ev.Type == abciTypes.MisbehaviorType_DUPLICATE_VOTE { // ?
		// 	a.log.Error("Wanted to punish val, but can't find it", zap.String("val", addr))
		// 	continue
		// }
		logger.Info("punish validator", zap.String("addr", addr))

		// This is why we need the addr=>pubkey map. Why, comet, why?
		pubkey, ok := a.valAddrToKey[addr]
		if !ok {
			logger.Error("unknown validator address", zap.String("addr", addr))
			panic(newFatalError("BeginBlock", &req, fmt.Sprintf("unknown validator address %v", addr)))
		}
		const punishDelta = 1
		newPower := ev.Validator.Power - punishDelta
		if err = a.validators.Punish(ctx, pubkey, newPower); err != nil {
			logger.Error("failed to punish validator", zap.Error(err))
			panic(newFatalError("BeginBlock", &req, fmt.Sprintf("failed to punish validator %v", addr)))
		}
	}

	return abciTypes.ResponseBeginBlock{}
}

// CheckTx is the "Guardian of the mempool: every node runs CheckTx before
// letting a transaction into its local mempool". Also "The transaction may come
// from an external user or another node". Further "CheckTx validates the
// transaction against the current state of the application, for example,
// checking signatures and account balances, but does not apply any of the state
// changes described in the transaction."
//
// This method must reject transactions that are invalid and/or may be crafted
// to attack the network by flooding the mempool or filling blocks with rejected
// transactions.
//
// This method is also used to re-check mempool transactions after blocks are
// mined. This is used to *evict* previously accepted transactions that become
// invalid, which may happen for a variety of reason only the application can
// decide, such as changes in account balance and last mined nonce.
//
// It is important to use this method rather than include failing transactions
// in blocks, particularly if the failure mode involves the transaction author
// spending no gas or achieving including in the block with little effort.
func (a *AbciApp) CheckTx(incoming abciTypes.RequestCheckTx) abciTypes.ResponseCheckTx {
	newTx := incoming.Type == abciTypes.CheckTxType_New
	logger := a.log.With(zap.Bool("recheck", !newTx))
	logger.Debug("check tx")
	ctx := context.Background()
	var err error
	code := codeOk

	// NOTE about the error logging here: These transactions are from users, so
	// most of these are not server errors, but client errors, so we ideally do
	// not want to log them at all in production. We'll keep a few for now to
	// help debugging.

	tx := &transactions.Transaction{}
	err = tx.UnmarshalBinary(incoming.Tx)
	if err != nil {
		code = codeEncodingError
		logger.Debug("failed to unmarshal transaction", zap.Error(err))
		return abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: err.Error()}
	}

	logger.Debug("",
		zap.String("sender", hex.EncodeToString(tx.Sender)),
		zap.String("PayloadType", tx.Body.PayloadType.String()),
		zap.Uint64("nonce", tx.Body.Nonce))

	// For a new transaction (not re-check), before looking at execution cost or
	// checking nonce validity, ensure the payload is recognized and signature is valid.
	if newTx {
		// Verify the correct chain ID is set, if it is set.
		if protected := tx.Body.ChainID != ""; protected && tx.Body.ChainID != a.cfg.ChainID {
			code = codeWrongChain
			logger.Info("wrong chain ID", zap.String("payloadType", tx.Body.PayloadType.String()))
			return abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: "wrong chain ID"}
		}
		// Verify Payload type
		if !tx.Body.PayloadType.Valid() {
			code = codeInvalidTxType
			logger.Debug("invalid payload type", zap.String("payloadType", tx.Body.PayloadType.String()))
			return abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: "invalid payload type"}
		}

		// Verify Signature
		err = ident.VerifyTransaction(tx)
		if err != nil {
			code = codeInvalidSignature
			logger.Debug("failed to verify transaction", zap.Error(err))
			return abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: err.Error()}
		}
	} else {
		txHash := cmtTypes.Tx(incoming.Tx).Hash()
		logger.Info("Recheck", zap.String("hash", hex.EncodeToString(txHash)), zap.Uint64("nonce", tx.Body.Nonce))
	}

	err = a.mempool.applyTransaction(ctx, tx)
	if err != nil {
		if errors.Is(err, transactions.ErrInvalidNonce) {
			code = codeInvalidNonce
			logger.Info("received transaction with invalid nonce", zap.Uint64("nonce", tx.Body.Nonce), zap.Error(err))
		} else {
			code = codeUnknownError
			logger.Warn("unexpected failure to verify transaction against local mempool state", zap.Error(err))
		}
		return abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: err.Error()}
	}

	return abciTypes.ResponseCheckTx{Code: code.Uint32()}
}

func (a *AbciApp) DeliverTx(req abciTypes.RequestDeliverTx) abciTypes.ResponseDeliverTx {
	logger := a.log.With(zap.String("stage", "ABCI DeliverTx"))

	ctx := context.Background()
	tx := &transactions.Transaction{}
	err := tx.UnmarshalBinary(req.Tx)
	if err != nil {
		logger.Error("failed to unmarshal transaction",
			zap.Error(err))
		return abciTypes.ResponseDeliverTx{
			Code: codeEncodingError.Uint32(),
			Log:  err.Error(),
		}
	}

	// Fail transactions with invalid signatures.
	if err = ident.VerifyTransaction(tx); err != nil {
		return abciTypes.ResponseDeliverTx{
			Code: codeInvalidSignature.Uint32(),
			Log:  err.Error(),
		}
	}

	var events []abciTypes.Event
	gasUsed := int64(0)
	txCode := codeOk

	logger = logger.With(zap.String("sender", hex.EncodeToString(tx.Sender)),
		zap.String("PayloadType", tx.Body.PayloadType.String()))

	switch tx.Body.PayloadType {
	case transactions.PayloadTypeDeploySchema:
		var schemaPayload transactions.Schema
		err = schemaPayload.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			txCode = codeEncodingError
			break
		}

		var schema *engineTypes.Schema
		schema, err = modDataset.ConvertSchemaToEngine(&schemaPayload)
		if err != nil {
			txCode = codeEncodingError
			break
		}

		var res *modDataset.ExecutionResponse
		res, err = a.database.Deploy(ctx, schema, tx)
		if err != nil {
			txCode = codeUnknownError
			break
		}

		dbID := utils.GenerateDBID(schema.Name, tx.Sender)
		gasUsed = res.GasUsed
		events = []abciTypes.Event{
			{
				Type: transactions.PayloadTypeDeploySchema.String(),
				Attributes: []abciTypes.EventAttribute{
					{Key: "sender", Value: hex.EncodeToString(tx.Sender), Index: true},
					{Key: "Result", Value: "Success", Index: true},
					{Key: "DBID", Value: dbID, Index: true},
				},
			},
		}
		logger.Debug("deployed database", zap.String("DBID", dbID))

	case transactions.PayloadTypeDropSchema:
		drop := &transactions.DropSchema{}
		err = drop.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			break
		}

		logger.Debug("drop database", zap.String("DBID", drop.DBID))

		var res *modDataset.ExecutionResponse
		res, err = a.database.Drop(ctx, drop.DBID, tx)
		if err != nil {
			txCode = codeUnknownError
			break
		}

		gasUsed = res.GasUsed
	case transactions.PayloadTypeExecuteAction:
		execution := &transactions.ActionExecution{}
		err = execution.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			txCode = codeEncodingError
			break
		}

		logger.Debug("execute action",
			zap.String("DBID", execution.DBID), zap.String("Action", execution.Action),
			zap.Any("Args", execution.Arguments))

		var res *modDataset.ExecutionResponse
		res, err = a.database.Execute(ctx, execution.DBID, execution.Action, convertArgs(execution.Arguments), tx)
		if err != nil {
			txCode = codeUnknownError
			break
		}

		gasUsed = res.GasUsed

	case transactions.PayloadTypeValidatorJoin:
		var join transactions.ValidatorJoin
		err = join.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			txCode = codeEncodingError
			break
		}

		logger.Debug("join validator",
			zap.String("pubkey", hex.EncodeToString(tx.Sender)),
			zap.Int64("power", int64(join.Power)))

		var res *modVal.ExecutionResponse
		res, err = a.validators.Join(ctx, int64(join.Power), tx)
		if err != nil {
			txCode = codeUnknownError
			break
		}

		events = []abciTypes.Event{
			{
				Type: "validator_join",
				Attributes: []abciTypes.EventAttribute{
					{Key: "Result", Value: "Success", Index: true},
					{Key: "ValidatorPubKey", Value: hex.EncodeToString(tx.Sender), Index: true},
					{Key: "ValidatorPower", Value: fmt.Sprintf("%d", join.Power), Index: true},
				},
			},
		}

		gasUsed = res.GasUsed

	case transactions.PayloadTypeValidatorLeave:
		var leave transactions.ValidatorLeave
		err = leave.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			txCode = codeEncodingError
			break
		}

		logger.Debug("leave validator", zap.String("pubkey", hex.EncodeToString(tx.Sender)))

		var res *modVal.ExecutionResponse
		res, err = a.validators.Leave(ctx, tx)
		if err != nil {
			txCode = codeUnknownError
			break
		}

		events = []abciTypes.Event{
			{
				Type: "validator_leave",
				Attributes: []abciTypes.EventAttribute{
					{Key: "Result", Value: "Success", Index: true},
					{Key: "ValidatorPubKey", Value: hex.EncodeToString(tx.Sender), Index: true},
					{Key: "ValidatorPower", Value: "0", Index: true},
				},
			},
		}

		gasUsed = res.GasUsed

	case transactions.PayloadTypeValidatorApprove:
		var approve transactions.ValidatorApprove
		err = approve.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			txCode = codeEncodingError
			break
		}

		logger.Debug("approve validator", zap.String("pubkey", hex.EncodeToString(approve.Candidate)))

		var res *modVal.ExecutionResponse
		res, err = a.validators.Approve(ctx, approve.Candidate, tx)
		if err != nil {
			txCode = codeUnknownError
			break
		}

		events = []abciTypes.Event{
			{
				Type: "validator_approve",
				Attributes: []abciTypes.EventAttribute{
					{Key: "Result", Value: "Success", Index: true},
					{Key: "CandidatePubKey", Value: hex.EncodeToString(approve.Candidate), Index: true},
					{Key: "ApproverPubKey", Value: hex.EncodeToString(tx.Sender), Index: true},
				},
			},
		}

		gasUsed = res.GasUsed

	case transactions.PayloadTypeValidatorRemove:
		var remove transactions.ValidatorRemove
		err = remove.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			txCode = codeEncodingError
			break
		}

		logger.Debug("remove validator", zap.String("pubkey", hex.EncodeToString(remove.Validator)))

		var res *modVal.ExecutionResponse
		res, err = a.validators.Remove(ctx, remove.Validator, tx)
		if err != nil {
			txCode = codeUnknownError
			break
		}

		events = []abciTypes.Event{
			{
				Type: "validator_remove",
				Attributes: []abciTypes.EventAttribute{
					{Key: "Result", Value: "Success", Index: true},
					{Key: "TargetPubKey", Value: hex.EncodeToString(remove.Validator), Index: true},
					{Key: "RemoverPubKey", Value: hex.EncodeToString(tx.Sender), Index: true},
				},
			},
		}

		gasUsed = res.GasUsed

	default:
		err = fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String())
	}
	if err != nil {
		a.log.Warn("failed to deliver tx", zap.Error(err))
		return abciTypes.ResponseDeliverTx{
			Code:    txCode.Uint32(),
			Log:     err.Error(),
			GasUsed: gasUsed,
		}
	}

	return abciTypes.ResponseDeliverTx{
		Code:    abciTypes.CodeTypeOK,
		GasUsed: gasUsed,
		Events:  events,
		Log:     "success",
	}
}

func (a *AbciApp) EndBlock(e abciTypes.RequestEndBlock) abciTypes.ResponseEndBlock {
	logger := a.log.With(zap.String("stage", "ABCI EndBlock"), zap.Int("height", int(e.Height)))
	logger.Debug("", zap.Int64("height", e.Height))

	var err error
	a.valUpdates, err = a.validators.Finalize(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to finalize validator updates: %v", err))
	}

	valUpdates := make([]abciTypes.ValidatorUpdate, len(a.valUpdates))
	for i, up := range a.valUpdates {
		valUpdates[i] = abciTypes.Ed25519ValidatorUpdate(up.PubKey, up.Power)
	}

	return abciTypes.ResponseEndBlock{
		ValidatorUpdates: valUpdates,
		ConsensusParamUpdates: &tendermintTypes.ConsensusParams{ // why are we "updating" these on every block? Should be nil for no update.
			// we can include evidence in here for malicious actors, but this is not important this release
			Version: &tendermintTypes.VersionParams{
				App: a.cfg.ApplicationVersion, // how would we change the application version?
			},
			Validator: &tendermintTypes.ValidatorParams{
				PubKeyTypes: []string{"ed25519"},
			},
		},
	}
}

func (a *AbciApp) Commit() abciTypes.ResponseCommit {
	logger := a.log.With(zap.String("stage", "ABCI Commit"))
	logger.Debug("start commit")
	ctx := context.Background()

	defer a.mempool.reset()

	idempotencyKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(idempotencyKey, uint64(a.blockHeight))

	id, err := a.committer.Commit(ctx, idempotencyKey)
	if err != nil {
		panic(newFatalError("Commit", nil, fmt.Sprintf("failed to commit atomic commit: %v", err)))
	}

	// Update AppHash and Block Height in metadata store.
	appHash, err := a.createNewAppHash(ctx, id)
	if err != nil {
		panic(newFatalError("Commit", nil, fmt.Sprintf("failed to create new app hash: %v", err)))
	}

	err = a.metadataStore.IncrementBlockHeight(ctx)
	if err != nil {
		panic(newFatalError("Commit", nil, fmt.Sprintf("failed to increment block height: %v", err)))
	}

	// Update the validator address=>pubkey map used by Penalize.
	for _, up := range a.valUpdates {
		if up.Power < 1 { // leave or punish
			delete(a.valAddrToKey, cometAddrFromPubKey(up.PubKey))
		} else { // add or update without remove
			a.valAddrToKey[cometAddrFromPubKey(up.PubKey)] = up.PubKey
		}
	}
	a.valUpdates = nil

	// snapshotting
	height, err := a.metadataStore.GetBlockHeight(ctx)
	if err != nil {
		a.log.Error("failed to get block height", zap.Error(err))
		return abciTypes.ResponseCommit{}
	}
	a.validators.UpdateBlockHeight(ctx, height)

	if a.snapshotter != nil && a.snapshotter.IsSnapshotDue(uint64(height)) {
		// TODO: Lock all DBs
		err = a.snapshotter.CreateSnapshot(uint64(height))
		if err != nil {
			a.log.Error("snapshot creation failed", zap.Error(err))
		}
		// Unlock all the DBs
	}

	return abciTypes.ResponseCommit{
		Data: appHash,
	}
}

func (a *AbciApp) Info(p0 abciTypes.RequestInfo) abciTypes.ResponseInfo {
	ctx := context.Background()

	// Load the current validator set from our store.
	vals, err := a.validators.CurrentSet(ctx)
	if err != nil { // TODO error return
		panic(newFatalError("Info", &p0, fmt.Sprintf("failed to load current validators: %v", err)))
	}
	// NOTE: We can check against cometbft/rpc/core.Validators(), but that only
	// works with an *in-process* node and after the node is started.

	// Prepare the validator addr=>pubkey map.
	a.valAddrToKey = make(map[string][]byte, len(vals))
	for _, vi := range vals {
		addr, err := pubkeyToAddr(vi.PubKey)
		if err != nil {
			panic(newFatalError("Info", &p0, fmt.Sprintf("invalid validator pubkey: %v", err)))
		}
		a.valAddrToKey[addr] = vi.PubKey
	}

	height, err := a.metadataStore.GetBlockHeight(ctx)
	if err != nil {
		panic(newFatalError("Info", &p0, fmt.Sprintf("failed to get block height: %v", err)))
	}

	appHash, err := a.metadataStore.GetAppHash(ctx)
	if err != nil {
		panic(newFatalError("Info", &p0, fmt.Sprintf("failed to get app hash: %v", err)))
	}

	a.log.Info("ABCI application is ready", zap.Int64("height", height))

	return abciTypes.ResponseInfo{
		LastBlockHeight:  height,
		LastBlockAppHash: appHash,
		// Version: kwildVersion, // the *software* semver string
		AppVersion: a.cfg.ApplicationVersion, // app protocol version, must match "the version in the current height’s block header"
	}
}

func (a *AbciApp) InitChain(p0 abciTypes.RequestInitChain) abciTypes.ResponseInitChain {
	logger := a.log.With(zap.String("stage", "ABCI InitChain"), zap.Int64("height", p0.InitialHeight))
	logger.Debug("", zap.String("ChainId", p0.ChainId))
	// maybe verify a.cfg.ChainID against the one in the request

	ctx := context.Background()

	// Initialize the validator module with the genesis validators.
	vldtrs := make([]*validators.Validator, len(p0.Validators))
	for i := range p0.Validators {
		vi := &p0.Validators[i]
		// pk := vi.PubKey.GetEd25519()
		// if pk == nil { panic("only ed25519 validator keys are supported") }
		pk := vi.PubKey.GetEd25519()
		vldtrs[i] = &validators.Validator{
			PubKey: pk,
			Power:  vi.Power,
		}
	}

	if err := a.validators.GenesisInit(context.Background(), vldtrs, p0.InitialHeight); err != nil {
		panic(fmt.Sprintf("GenesisInit failed: %v", err))
	}

	valUpdates := make([]abciTypes.ValidatorUpdate, len(vldtrs))
	for i, validator := range vldtrs {
		valUpdates[i] = abciTypes.Ed25519ValidatorUpdate(validator.PubKey, validator.Power)
	}

	err := a.metadataStore.SetAppHash(ctx, a.cfg.GenesisAppHash)
	if err != nil {
		panic(fmt.Sprintf("failed to set app hash: %v", err))
	}

	logger.Info("initialized chain", zap.String("app hash", fmt.Sprintf("%x", a.cfg.GenesisAppHash)))

	return abciTypes.ResponseInitChain{
		Validators: valUpdates,
		AppHash:    a.cfg.GenesisAppHash,
	}
}

func (a *AbciApp) ApplySnapshotChunk(p0 abciTypes.RequestApplySnapshotChunk) abciTypes.ResponseApplySnapshotChunk {
	if a.bootstrapper == nil {
		return abciTypes.ResponseApplySnapshotChunk{Result: abciTypes.ResponseApplySnapshotChunk_ABORT, RefetchChunks: nil}
	}

	refetchChunks, status, err := a.bootstrapper.ApplySnapshotChunk(p0.Chunk, p0.Index)
	if err != nil {
		return abciTypes.ResponseApplySnapshotChunk{Result: abciStatus(status), RefetchChunks: refetchChunks}
	}

	ctx := context.Background()

	if a.bootstrapper.IsDBRestored() {
		err = a.metadataStore.SetAppHash(ctx, a.bootupState.appHash)
		if err != nil {
			return abciTypes.ResponseApplySnapshotChunk{Result: abciTypes.ResponseApplySnapshotChunk_ABORT, RefetchChunks: nil}
		}

		err = a.metadataStore.SetBlockHeight(ctx, a.bootupState.height)
		if err != nil {
			return abciTypes.ResponseApplySnapshotChunk{Result: abciTypes.ResponseApplySnapshotChunk_ABORT, RefetchChunks: nil}
		}

		a.log.Info("Bootstrapped database successfully")
	}
	return abciTypes.ResponseApplySnapshotChunk{Result: abciTypes.ResponseApplySnapshotChunk_ACCEPT, RefetchChunks: nil}
}

func (a *AbciApp) ListSnapshots(p0 abciTypes.RequestListSnapshots) abciTypes.ResponseListSnapshots {
	if a.snapshotter == nil {
		return abciTypes.ResponseListSnapshots{Snapshots: nil}
	}

	snapshots, err := a.snapshotter.ListSnapshots()
	if err != nil {
		return abciTypes.ResponseListSnapshots{Snapshots: nil}
	}

	var res []*abciTypes.Snapshot
	for _, snapshot := range snapshots {
		abcisnapshot, err := convertToABCISnapshot(&snapshot)
		if err != nil {
			return abciTypes.ResponseListSnapshots{Snapshots: nil}
		}
		res = append(res, abcisnapshot)
	}
	return abciTypes.ResponseListSnapshots{Snapshots: res}
}

func (a *AbciApp) LoadSnapshotChunk(p0 abciTypes.RequestLoadSnapshotChunk) abciTypes.ResponseLoadSnapshotChunk {
	if a.snapshotter == nil {
		return abciTypes.ResponseLoadSnapshotChunk{Chunk: nil}
	}

	chunk := a.snapshotter.LoadSnapshotChunk(p0.Height, p0.Format, p0.Chunk)
	return abciTypes.ResponseLoadSnapshotChunk{Chunk: chunk}
}

func (a *AbciApp) OfferSnapshot(p0 abciTypes.RequestOfferSnapshot) abciTypes.ResponseOfferSnapshot {
	if a.bootstrapper == nil {
		return abciTypes.ResponseOfferSnapshot{Result: abciTypes.ResponseOfferSnapshot_REJECT}
	}

	snapshot := convertABCISnapshots(p0.Snapshot)
	if a.bootstrapper.OfferSnapshot(snapshot) != nil {
		return abciTypes.ResponseOfferSnapshot{Result: abciTypes.ResponseOfferSnapshot_REJECT}
	}
	a.bootupState.appHash = p0.Snapshot.Hash
	a.bootupState.height = int64(snapshot.Height)
	return abciTypes.ResponseOfferSnapshot{Result: abciTypes.ResponseOfferSnapshot_ACCEPT}
}

// txSubList implements sort.Interface to perform in-place sorting of a slice
// that is a subset of another slice, reordering in both while staying within
// the subsets positions in the parent slice.
//
// For example:
//
//	parent slice: {a0, b2, b0, a1, b1}
//	b's subset: {b2, b0, b1}
//	sorted subset: {b0, b1, b2}
//	parent slice: {a0, b0, b1, a1, b2}
//
// The set if locations used by b elements within the parent slice is unchanged,
// but the elements are sorted.
type txSubList struct {
	sub   []*indexedTxn // sort.Sort references only this with Len and Less
	super []*indexedTxn // sort.Sort also Swaps in super using the i field
}

func (txl txSubList) Len() int {
	return len(txl.sub)
}

func (txl txSubList) Less(i int, j int) bool {
	a, b := txl.sub[i], txl.sub[j]
	return a.Body.Nonce < b.Body.Nonce
}

func (txl txSubList) Swap(i int, j int) {
	// Swap elements in sub.
	txl.sub[i], txl.sub[j] = txl.sub[j], txl.sub[i]
	// Swap the elements in their positions in super.
	ip, jp := txl.sub[i].i, txl.sub[j].i
	txl.super[ip], txl.super[jp] = txl.super[jp], txl.super[ip]
}

// indexedTxn facilitates in-place sorting of transaction slices that are
// subsets of other larger slices using a txSubList. This is only used within
// prepareMempoolTxns, and is package-level rather than scoped to the function
// because we define methods to implement sort.Interface.
type indexedTxn struct {
	i int // index in superset slice
	*transactions.Transaction

	is int // not used for sorting, only referencing the marshalled txn slice
}

// prepareMempoolTxns prepares the transactions for the block we are proposing.
// The input transactions are from mempool direct from cometbft, and we modify
// the list for our purposes. This includes ensuring transactions from the same
// sender in ascending nonce-order, enforcing the max bytes limit, etc.
//
// NOTE: This is a plain function instead of an AbciApplication method so that
// it may be directly tested.
func prepareMempoolTxns(txs [][]byte, maxBytes int, log *log.Logger) [][]byte {
	// Unmarshal and index the transactions.
	var okTxns []*indexedTxn
	var i int
	for is, txB := range txs {
		tx := &transactions.Transaction{}
		err := tx.UnmarshalBinary(txB)
		if err != nil {
			log.Error("failed to unmarshal transaction that was previously accepted to mempool", zap.Error(err))
			continue // should not have passed CheckTx to get into our mempool
		}
		okTxns = append(okTxns, &indexedTxn{i, tx, is})
		i++
	}

	// Group by sender and stable sort each group by nonce.
	grouped := make(map[string][]*indexedTxn)
	for _, txn := range okTxns {
		key := string(txn.Sender)
		grouped[key] = append(grouped[key], txn)
	}
	for _, txns := range grouped {
		sort.Stable(txSubList{
			sub:   txns,
			super: okTxns,
		})
	}

	// TODO: truncate based on our max block size since we'll have to set
	// ConsensusParams.Block.MaxBytes to -1 so that we get ALL transactions even
	// if it goes beyond max_tx_bytes.  See:
	// https://github.com/cometbft/cometbft/pull/1003
	// https://docs.cometbft.com/v0.38/spec/abci/abci++_methods#prepareproposal
	// https://github.com/cometbft/cometbft/issues/980

	// Grab the bytes rather than re-marshalling.
	nonces := make([]uint64, 0, len(okTxns))
	finalTxns := make([][]byte, 0, len(okTxns))
	i = 0
	for _, tx := range okTxns {
		if i > 0 && tx.Body.Nonce == nonces[i-1] && bytes.Equal(tx.Sender, okTxns[i-1].Sender) {
			log.Warn(fmt.Sprintf("Dropping tx with re-used nonce %d from block proposal", tx.Body.Nonce))
			continue // mempool recheck should have removed this
		}
		finalTxns = append(finalTxns, txs[tx.is])
		nonces = append(nonces, tx.Body.Nonce)
		i++
	}

	return finalTxns
}

func (a *AbciApp) PrepareProposal(p0 abciTypes.RequestPrepareProposal) abciTypes.ResponsePrepareProposal {
	logger := a.log.With(zap.String("stage", "ABCI PrepareProposal"),
		zap.Int64("height", p0.Height), zap.Int("txs", len(p0.Txs)))

	okTxns := prepareMempoolTxns(p0.Txs, int(p0.MaxTxBytes), logger)

	if len(okTxns) != len(p0.Txs) {
		logger.Info("PrepareProposal: number of transactions in proposed block has changed!",
			zap.Int("in", len(p0.Txs)), zap.Int("out", len(okTxns)))
	} else if logger.L.Level() <= log.DebugLevel { // spare the check if it wouldn't be logged
		for i, tx := range okTxns {
			if !bytes.Equal(tx, p0.Txs[i]) { //  not and error, just notable
				logger.Debug("transaction was moved or mutated", zap.Int("idx", i))
			}
		}
	}

	return abciTypes.ResponsePrepareProposal{
		Txs: okTxns,
	}
}

func (a *AbciApp) validateProposalTransactions(ctx context.Context, txns [][]byte) error {
	logger := a.log.With(zap.String("stage", "ABCI ProcessProposal"))
	grouped, err := groupTxsBySender(txns)
	if err != nil {
		return fmt.Errorf("failed to group transaction by sender: %w", err)
	}

	// ensure there are no gaps in an account's nonce, either from the
	// previous best confirmed or within this block. Our current transaction
	// execution does not update an accounts nonce in state unless it is the
	// next nonce. Delivering transactions to a block in that way cannot happen.
	for sender, txs := range grouped {
		acct, err := a.accounts.GetAccount(ctx, []byte(sender))
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}
		expectedNonce := uint64(acct.Nonce) + 1

		for _, tx := range txs {
			if tx.Body.Nonce != expectedNonce {
				logger.Warn("nonce mismatch", zap.Uint64("txNonce", tx.Body.Nonce),
					zap.Uint64("expectedNonce", expectedNonce), zap.String("nonces", fmt.Sprintf("%v", nonceList(txs))))
				return fmt.Errorf("nonce mismatch, expected: %d tx: %d", expectedNonce, tx.Body.Nonce)
			}
			expectedNonce++

			chainID := tx.Body.ChainID
			if protected := chainID != ""; protected && chainID != a.cfg.ChainID {
				return fmt.Errorf("protected transaction with mismatched chain ID")
			}

			// This block proposal may include transactions that did not pass
			// through our mempool, so we have to verify all signatures.
			if err = ident.VerifyTransaction(tx); err != nil {
				return fmt.Errorf("transaction signature verification failed: %w", err)
			}
		}
	}
	return nil
}

// ProcessProposal should validate the received blocks and reject the block if:
// 1. transactions are not ordered by nonces
// 2. nonce is less than the last committed nonce for the account
// 3. duplicates or gaps in the nonces
// 4. transaction size is greater than the max_tx_bytes
// else accept the proposed block.
func (a *AbciApp) ProcessProposal(p0 abciTypes.RequestProcessProposal) abciTypes.ResponseProcessProposal {
	logger := a.log.With(zap.String("stage", "ABCI ProcessProposal"),
		zap.Int64("height", p0.Height), zap.Int("txs", len(p0.Txs)))

	ctx := context.Background()
	if err := a.validateProposalTransactions(ctx, p0.Txs); err != nil {
		logger.Warn("rejecting block proposal", zap.Error(err))
		return abciTypes.ResponseProcessProposal{Status: abciTypes.ResponseProcessProposal_REJECT}
	}

	// TODO: Verify the Tx and Block sizes based on the genesis configuration
	return abciTypes.ResponseProcessProposal{Status: abciTypes.ResponseProcessProposal_ACCEPT}
}

func (a *AbciApp) Query(p0 abciTypes.RequestQuery) abciTypes.ResponseQuery {
	return abciTypes.ResponseQuery{}
}

// updateAppHash updates the app hash with the given app hash.
// It persists the app hash to the metadata store.
func (a *AbciApp) createNewAppHash(ctx context.Context, addition []byte) ([]byte, error) {
	oldHash, err := a.metadataStore.GetAppHash(ctx)
	if err != nil {
		return nil, err
	}

	newHash := crypto.Sha256(append(oldHash, addition...))

	err = a.metadataStore.SetAppHash(ctx, newHash)
	return newHash, err
}

// TODO: here should probably be other apphash computations such as the genesis
// config digest. The cmd/kwild/config package should probably not
// contain consensus-critical computations.

// convertArgs converts the string args to type any.
func convertArgs(args [][]string) [][]any {
	converted := make([][]any, len(args))
	for i, arg := range args {
		converted[i] = make([]any, len(arg))
		for j, a := range arg {
			converted[i][j] = a
		}
	}

	return converted
}

var (
	appHashKey     = []byte("a")
	blockHeightKey = []byte("b")
)

type metadataStore struct {
	kv KVStore
}

func (m *metadataStore) GetAppHash(ctx context.Context) ([]byte, error) {
	res, err := m.kv.Get(appHashKey)
	if err == kv.ErrKeyNotFound {
		return nil, nil
	}
	return res, err
}

func (m *metadataStore) SetAppHash(ctx context.Context, appHash []byte) error {
	return m.kv.Set(appHashKey, appHash)
}

func (m *metadataStore) GetBlockHeight(ctx context.Context) (int64, error) {
	height, err := m.kv.Get(blockHeightKey)
	if err == kv.ErrKeyNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return int64(binary.BigEndian.Uint64(height)), nil
}

func (m *metadataStore) SetBlockHeight(ctx context.Context, height int64) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(height))

	return m.kv.Set(blockHeightKey, buf)
}

func (m *metadataStore) IncrementBlockHeight(ctx context.Context) error {
	height, err := m.GetBlockHeight(ctx)
	if err != nil {
		return err
	}

	return m.SetBlockHeight(ctx, height+1)
}