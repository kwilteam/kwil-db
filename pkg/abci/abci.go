package abci

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"github.com/kwilteam/kwil-db/pkg/kv"
	"github.com/kwilteam/kwil-db/pkg/log"
	modDataset "github.com/kwilteam/kwil-db/pkg/modules/datasets"
	modVal "github.com/kwilteam/kwil-db/pkg/modules/validators"
	"github.com/kwilteam/kwil-db/pkg/transactions"
	"github.com/kwilteam/kwil-db/pkg/validators"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tendermintTypes "github.com/cometbft/cometbft/proto/tendermint/types"
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

// func (fe FatalError) String() string {
// 	return fmt.Sprintf("Application Method: %s\nError: %s\nRequest (%T): %v",
// 		fe.AppMethod, fe.Message, fe.Request, fe.Request)
// }

// func newFatalError(method string, request fmt.Stringer, message string) FatalError {
// 	if request == nil {
// 		request = nilStringer{}
// 	}

// 	return FatalError{
// 		AppMethod: method,
// 		Request:   request,
// 		Message:   message,
// 	}
// }

type nilStringer struct{}

func (ds nilStringer) String() string {
	return "no message"
}

func NewAbciApp(database DatasetsModule, vldtrs ValidatorModule, kv KVStore, committer AtomicCommitter, snapshotter SnapshotModule,
	bootstrapper DBBootstrapModule, opts ...AbciOpt) *AbciApp {
	app := &AbciApp{
		database:   database,
		validators: vldtrs,
		committer:  committer,
		metadataStore: &metadataStore{
			kv: kv,
		},
		bootstrapper: bootstrapper,
		snapshotter:  snapshotter,

		valAddrToKey: make(map[string][]byte),

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
	bootstrapper  DBBootstrapModule
	metadataStore *metadataStore

	log log.Logger

	// Expected AppState after bootstrapping the node with a given snapshot,
	// state gets updated with the bootupState after bootstrapping
	bootupState appState

	applicationVersion uint64
}

var _ abciTypes.Application = &AbciApp{}

// The Application interface methods in four groups according to the
// "connection" used by CometBFT to interact with the application. Calls to the
// methods within a connection are synchronized. They are not synchronized
// between the connections. e.g. CheckTx calls from the mempool connection can
// occur concurrent to calls on the Consensus connection.

// CheckTx is on the mempool connection
func (a *AbciApp) CheckTx(ctx context.Context, incoming *abciTypes.RequestCheckTx) (*abciTypes.ResponseCheckTx, error) {
	logger := a.log.With(zap.String("stage", "ABCI CheckTx"))
	logger.Debug("check tx")

	tx := &transactions.Transaction{}
	err := tx.UnmarshalBinary(incoming.Tx)
	if err != nil {
		logger.Error("failed to unmarshal transaction", zap.Error(err))
		return &abciTypes.ResponseCheckTx{Code: 1, Log: err.Error()}, nil // return error now???
	}

	logger.Debug("",
		zap.String("sender", tx.GetSenderAddress()),
		zap.String("PayloadType", tx.Body.PayloadType.String()))

	err = tx.Verify()
	if err != nil {
		logger.Error("failed to verify transaction", zap.Error(err))
		return &abciTypes.ResponseCheckTx{Code: 1, Log: err.Error()}, nil
	}

	return &abciTypes.ResponseCheckTx{Code: 0}, nil
}

// FinalizeBlock is on the consensus connection
func (a *AbciApp) FinalizeBlock(ctx context.Context, req *abciTypes.RequestFinalizeBlock) (*abciTypes.ResponseFinalizeBlock, error) {
	logger := a.log.With(zap.String("stage", "ABCI FinalizeBlock"), zap.Int("height", int(req.Height)))

	res := &abciTypes.ResponseFinalizeBlock{}

	// BeginBlock was this part

	err := a.committer.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin atomic commit failed: %w", err)
	}

	// Punish bad validators.
	for _, ev := range req.Misbehavior {
		addr := string(ev.Validator.Address) // comet example app confirms this conversion... weird
		// if ev.Type == abciTypes.MisbehaviorType_DUPLICATE_VOTE { // ?
		// 	a.log.Error("Wanted to punish val, but can't find it", zap.String("val", addr))
		// 	continue
		// }
		logger.Info("punish validator", zap.String("addr", addr))

		// This is why we need the addr=>pubkey map. Why, comet, why?
		pubkey, ok := a.valAddrToKey[addr]
		if !ok {
			return nil, fmt.Errorf("unknown validator address %v", addr)
		}
		const punishDelta = 1
		newPower := ev.Validator.Power - punishDelta
		if err = a.validators.Punish(ctx, pubkey, newPower); err != nil {
			return nil, fmt.Errorf("failed to punish validator: %w", err)
		}
	}

	for _, tx := range req.Txs {
		// DeliverTx was the part in this loop.
		execRes := a.executeTx(ctx, tx, logger)
		res.TxResults = append(res.TxResults, execRes)
	}

	// EndBlock was this part

	a.valUpdates = a.validators.Finalize(ctx)

	res.ValidatorUpdates = make([]abciTypes.ValidatorUpdate, len(a.valUpdates))
	for i, up := range a.valUpdates {
		res.ValidatorUpdates[i] = abciTypes.Ed25519ValidatorUpdate(up.PubKey, up.Power)
	}

	res.ConsensusParamUpdates = &tendermintTypes.ConsensusParams{
		// we can include evidence in here for malicious actors, but this is not important this release
		Version: &tendermintTypes.VersionParams{
			App: a.applicationVersion,
		},
		Validator: &tendermintTypes.ValidatorParams{
			PubKeyTypes: []string{"ed25519"},
		},
	}

	// generate the unique id for all changes occurred thus far
	id, err := a.committer.ID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate atomic commit ID: %w", err)
	}

	appHash, err := a.createNewAppHash(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create new app hash: %w", err)
	}
	res.AppHash = appHash

	return res, nil
}

func (a *AbciApp) executeTx(ctx context.Context, rawTx []byte, logger *log.Logger) *abciTypes.ExecTxResult {
	var events []abciTypes.Event
	var gasUsed int64

	newExecuteTxRes := func(code TxCode, err error) *abciTypes.ExecTxResult {
		res := &abciTypes.ExecTxResult{
			Code:    code.Uint32(),
			GasUsed: gasUsed,
			Events:  events,
			Log:     "success",
			// Data, GasWanted, Info, Codespace
		}

		if err != nil {
			logger.Warn("failed to deliver tx", zap.Error(err))
			res.Log = fmt.Sprintf("FAILED TRANSACTION: %v", err) // may be too much info in err
		}

		return res
	}

	tx := &transactions.Transaction{}
	err := tx.UnmarshalBinary(rawTx)
	if err != nil {
		logger.Error("failed to unmarshal transaction", zap.Error(err))
		return newExecuteTxRes(CodeEncodingError, err)
	}

	logger = logger.With(zap.String("Sender", tx.GetSenderAddress()),
		zap.String("PayloadType", tx.Body.PayloadType.String()))

	switch tx.Body.PayloadType {
	case transactions.PayloadTypeDeploySchema:
		var schemaPayload transactions.Schema
		err = schemaPayload.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			return newExecuteTxRes(CodeEncodingError, err)
		}

		var schema *engineTypes.Schema
		schema, err = modDataset.ConvertSchemaToEngine(&schemaPayload)
		if err != nil {
			return newExecuteTxRes(CodeEncodingError, err)
		}

		var res *modDataset.ExecutionResponse
		res, err = a.database.Deploy(ctx, schema, tx)
		if err != nil {
			return newExecuteTxRes(CodeUnknownError, err)
		}

		dbID := utils.GenerateDBID(schema.Name, tx.Sender)
		gasUsed = res.GasUsed
		events = []abciTypes.Event{
			{
				Type: transactions.PayloadTypeDeploySchema.String(),
				Attributes: []abciTypes.EventAttribute{
					{Key: "Sender", Value: tx.GetSenderAddress(), Index: true},
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
			return newExecuteTxRes(CodeEncodingError, err)
		}

		logger.Debug("drop database", zap.String("DBID", drop.DBID))

		var res *modDataset.ExecutionResponse
		res, err = a.database.Drop(ctx, drop.DBID, tx)
		if err != nil {
			return newExecuteTxRes(CodeUnknownError, err)
		}

		gasUsed = res.GasUsed

	case transactions.PayloadTypeExecuteAction:
		execution := &transactions.ActionExecution{}
		err = execution.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			return newExecuteTxRes(CodeEncodingError, err)
		}

		logger.Debug("execute action",
			zap.String("DBID", execution.DBID), zap.String("Action", execution.Action),
			zap.Any("Args", execution.Arguments))

		var res *modDataset.ExecutionResponse
		res, err = a.database.Execute(ctx, execution.DBID, execution.Action, convertArgs(execution.Arguments), tx)
		if err != nil {
			return newExecuteTxRes(CodeUnknownError, err)
		}

		gasUsed = res.GasUsed

	case transactions.PayloadTypeValidatorJoin:
		var join transactions.ValidatorJoin
		err = join.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			return newExecuteTxRes(CodeEncodingError, err)
		}

		logger.Debug("join validator",
			zap.String("pubkey", hex.EncodeToString(join.Candidate)),
			zap.Int64("power", int64(join.Power)))

		var res *modVal.ExecutionResponse
		res, err = a.validators.Join(ctx, join.Candidate, int64(join.Power), tx)
		if err != nil {
			return newExecuteTxRes(CodeUnknownError, err)
		}

		events = []abciTypes.Event{
			{
				Type: "validator_join",
				Attributes: []abciTypes.EventAttribute{
					{Key: "Result", Value: "Success", Index: true},
					{Key: "ValidatorPubKey", Value: hex.EncodeToString(join.Candidate), Index: true},
					{Key: "ValidatorPower", Value: fmt.Sprintf("%d", join.Power), Index: true},
				},
			},
		}

		gasUsed = res.GasUsed

	case transactions.PayloadTypeValidatorLeave:
		var leave transactions.ValidatorLeave
		err = leave.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			return newExecuteTxRes(CodeEncodingError, err)
		}

		logger.Debug("leave validator", zap.String("pubkey", hex.EncodeToString(leave.Validator)))

		var res *modVal.ExecutionResponse
		res, err = a.validators.Leave(ctx, leave.Validator, tx)
		if err != nil {
			return newExecuteTxRes(CodeUnknownError, err)
		}

		events = []abciTypes.Event{
			{
				Type: "remove_validator", // is this name arbitrary? it should be "validator_leave" for consistency
				Attributes: []abciTypes.EventAttribute{
					{Key: "Result", Value: "Success", Index: true},
					{Key: "ValidatorPubKey", Value: hex.EncodeToString(leave.Validator), Index: true},
					{Key: "ValidatorPower", Value: "0", Index: true},
				},
			},
		}

		gasUsed = res.GasUsed

	case transactions.PayloadTypeValidatorApprove:
		var approve transactions.ValidatorApprove
		err = approve.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			return newExecuteTxRes(CodeEncodingError, err)
		}

		logger.Debug("approve validator", zap.String("pubkey", hex.EncodeToString(approve.Candidate)))

		var res *modVal.ExecutionResponse
		res, err = a.validators.Approve(ctx, approve.Candidate, tx)
		if err != nil {
			return newExecuteTxRes(CodeUnknownError, err)
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
	default:
		err = fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String())
		return newExecuteTxRes(CodeUnknownError, err)
	}

	return newExecuteTxRes(CodeOk, nil)
}

// FinalizeBlock is on the consensus connection
func (a *AbciApp) Commit(ctx context.Context, _ *abciTypes.RequestCommit) (*abciTypes.ResponseCommit, error) {
	err := a.metadataStore.IncrementBlockHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to increment block height: %w", err)
	}

	err = a.committer.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit atomic commit: %w", err)
	}

	// Update the validator address=>pubkey map used by Penalize.
	for _, up := range a.valUpdates {
		addr := cometAddrFromPubKey(up.PubKey)
		if up.Power < 1 { // leave or punish
			delete(a.valAddrToKey, addr)
		} else { // add or update without remove
			a.valAddrToKey[addr] = up.PubKey
		}
	}
	a.valUpdates = nil

	// snapshotting
	height, err := a.metadataStore.GetBlockHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get block height: %w", err)
	}

	if a.snapshotter != nil && a.snapshotter.IsSnapshotDue(uint64(height)) {
		// TODO: Lock all DBs
		err = a.snapshotter.CreateSnapshot(uint64(height))
		if err != nil {
			a.log.Error("snapshot creation failed", zap.Error(err))
		}
		// Unlock all the DBs
	}

	return &abciTypes.ResponseCommit{
		RetainHeight: height,
	}, nil
}

// FinalizeBlock is on the consensus connection
func (a *AbciApp) InitChain(ctx context.Context, req *abciTypes.RequestInitChain) (*abciTypes.ResponseInitChain, error) {
	logger := a.log.With(zap.String("stage", "ABCI InitChain"), zap.Int64("height", req.InitialHeight))
	logger.Debug("", zap.String("ChainId", req.ChainId))

	// Initialize the validator module with the genesis validators.
	vldtrs := make([]*validators.Validator, len(req.Validators))
	for i := range req.Validators {
		vi := &req.Validators[i]
		// pk := vi.PubKey.GetEd25519()
		// if pk == nil { panic("only ed25519 validator keys are supported") }
		pk := vi.PubKey.GetEd25519()
		vldtrs[i] = &validators.Validator{
			PubKey: pk,
			Power:  vi.Power,
		}
	}

	if err := a.validators.GenesisInit(context.Background(), vldtrs); err != nil {
		return nil, fmt.Errorf("validators.GenesisInit failed: %w", err)
	}

	valUpdates := make([]abciTypes.ValidatorUpdate, len(vldtrs))
	for i, validator := range vldtrs {
		valUpdates[i] = abciTypes.Ed25519ValidatorUpdate(validator.PubKey, validator.Power)
	}

	apphash, err := a.metadataStore.GetAppHash(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to get app hash: %v", err))
		// TODO: should we initialize with a genesis hash instead if it fails
		// TODO: apparently InitChain is only genesis, so yes it should only be genesis hash
		// in fact, I don't think we should be getting it from this store at all
	}

	return &abciTypes.ResponseInitChain{
		Validators: valUpdates,
		AppHash:    apphash,
	}, nil
}

func (a *AbciApp) ProcessProposal(ctx context.Context, req *abciTypes.RequestProcessProposal) (*abciTypes.ResponseProcessProposal, error) {
	a.log.Debug("",
		zap.String("stage", "ABCI ProcessProposal"),
		zap.Int64("height", req.Height),
		zap.Int("txs", len(req.Txs)))

	// TODO: do something with the txs?

	return &abciTypes.ResponseProcessProposal{Status: abciTypes.ResponseProcessProposal_ACCEPT}, nil
}

// ExtendVote create an application specific vote extension.
//
//   - ResponseExtendVote.vote_extension is application-generated information that
//     will be signed by CometBFT and attached to the Precommit message.
//   - The Application may choose to use an empty vote extension (0 length).
//   - The contents of RequestExtendVote correspond to the proposed block on which
//     the consensus algorithm will send the Precommit message.
//   - ResponseExtendVote.vote_extension will only be attached to a non-nil
//     Precommit message. If the consensus algorithm is to precommit nil, it will
//     not call RequestExtendVote.
//   - The Application logic that creates the extension can be non-deterministic.
func (a *AbciApp) ExtendVote(ctx context.Context, req *abciTypes.RequestExtendVote) (*abciTypes.ResponseExtendVote, error) {
	spew.Dump(req)
	return &abciTypes.ResponseExtendVote{}, nil
}

// Verify application's vote extension data
func (a *AbciApp) VerifyVoteExtension(ctx context.Context, req *abciTypes.RequestVerifyVoteExtension) (*abciTypes.ResponseVerifyVoteExtension, error) {
	spew.Dump(req)
	if len(req.VoteExtension) > 0 {
		return &abciTypes.ResponseVerifyVoteExtension{
			Status: abciTypes.ResponseVerifyVoteExtension_REJECT,
		}, nil
	}
	return &abciTypes.ResponseVerifyVoteExtension{
		Status: abciTypes.ResponseVerifyVoteExtension_ACCEPT,
	}, nil
}

// Info is part of the Info/Query connection.
func (a *AbciApp) Info(ctx context.Context, req *abciTypes.RequestInfo) (*abciTypes.ResponseInfo, error) {
	err := a.committer.ClearWal(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to clear WAL: %w", err)
	}

	// Load the current validator set from our store.
	vals, err := a.validators.CurrentSet(ctx)
	if err != nil { // TODO error return
		return nil, fmt.Errorf("failed to load current validators: %w", err)
	}
	// NOTE: We can check against cometbft/rpc/core.Validators(), but that only
	// works with an *in-process* node and after the node is started.

	// Prepare the validator addr=>pubkey map.
	a.valAddrToKey = make(map[string][]byte, len(vals))
	for _, vi := range vals {
		addr, err := pubkeyToAddr(vi.PubKey)
		if err != nil {
			return nil, fmt.Errorf("invalid validator pubkey: %w", err)
		}
		a.valAddrToKey[addr] = vi.PubKey
	}

	height, err := a.metadataStore.GetBlockHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get block height: %w", err)
	}

	appHash, err := a.metadataStore.GetAppHash(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get app hash: %w", err)
	}

	return &abciTypes.ResponseInfo{
		LastBlockHeight:  height,
		LastBlockAppHash: appHash,
		AppVersion:       a.applicationVersion,
	}, nil
}

// Query is part of the Info/Query connection.
func (a *AbciApp) Query(ctx context.Context, req *abciTypes.RequestQuery) (*abciTypes.ResponseQuery, error) {
	return &abciTypes.ResponseQuery{}, nil // TODO: handle state query???
}

// ApplySnapshotChunk is on the state sync connection
func (a *AbciApp) ApplySnapshotChunk(ctx context.Context, req *abciTypes.RequestApplySnapshotChunk) (*abciTypes.ResponseApplySnapshotChunk, error) {
	refetchChunks, status, err := a.bootstrapper.ApplySnapshotChunk(req.Chunk, req.Index)
	if err != nil {
		return &abciTypes.ResponseApplySnapshotChunk{Result: abciStatus(status), RefetchChunks: refetchChunks}, nil
	}

	if a.bootstrapper.IsDBRestored() {
		err = a.metadataStore.SetAppHash(ctx, a.bootupState.appHash)
		if err != nil {
			return &abciTypes.ResponseApplySnapshotChunk{Result: abciTypes.ResponseApplySnapshotChunk_ABORT, RefetchChunks: nil}, nil
		}

		err = a.metadataStore.SetBlockHeight(ctx, a.bootupState.height)
		if err != nil {
			return &abciTypes.ResponseApplySnapshotChunk{Result: abciTypes.ResponseApplySnapshotChunk_ABORT, RefetchChunks: nil}, nil
		}

		a.log.Info("Bootstrapped database successfully")
	}
	return &abciTypes.ResponseApplySnapshotChunk{Result: abciTypes.ResponseApplySnapshotChunk_ACCEPT, RefetchChunks: nil}, nil
}

// ListSnapshots is on the state sync connection
func (a *AbciApp) ListSnapshots(ctx context.Context, req *abciTypes.RequestListSnapshots) (*abciTypes.ResponseListSnapshots, error) {
	if a.snapshotter == nil {
		return &abciTypes.ResponseListSnapshots{}, nil
	}

	snapshots, err := a.snapshotter.ListSnapshots()
	if err != nil {
		return &abciTypes.ResponseListSnapshots{}, nil
	}

	var res []*abciTypes.Snapshot
	for _, snapshot := range snapshots {
		abciSnapshot, err := convertToABCISnapshot(&snapshot)
		if err != nil {
			return &abciTypes.ResponseListSnapshots{}, nil
		}
		res = append(res, abciSnapshot)
	}
	return &abciTypes.ResponseListSnapshots{Snapshots: res}, nil
}

// LoadSnapshotChunk is on the state sync connection
func (a *AbciApp) LoadSnapshotChunk(ctx context.Context, req *abciTypes.RequestLoadSnapshotChunk) (*abciTypes.ResponseLoadSnapshotChunk, error) {
	if a.snapshotter == nil {
		return &abciTypes.ResponseLoadSnapshotChunk{}, nil
	}

	chunk := a.snapshotter.LoadSnapshotChunk(req.Height, req.Format, req.Chunk)
	return &abciTypes.ResponseLoadSnapshotChunk{Chunk: chunk}, nil
}

// OfferSnapshot is on the state sync connection
func (a *AbciApp) OfferSnapshot(ctx context.Context, req *abciTypes.RequestOfferSnapshot) (*abciTypes.ResponseOfferSnapshot, error) {
	snapshot := convertABCISnapshots(req.Snapshot)
	if a.bootstrapper.OfferSnapshot(snapshot) != nil {
		return &abciTypes.ResponseOfferSnapshot{Result: abciTypes.ResponseOfferSnapshot_REJECT}, nil
	}
	a.bootupState.appHash = req.Snapshot.Hash
	a.bootupState.height = int64(snapshot.Height)
	return &abciTypes.ResponseOfferSnapshot{Result: abciTypes.ResponseOfferSnapshot_ACCEPT}, nil
}

func (a *AbciApp) PrepareProposal(ctx context.Context, req *abciTypes.RequestPrepareProposal) (*abciTypes.ResponsePrepareProposal, error) {
	a.log.Debug("",
		zap.String("stage", "ABCI PrepareProposal"),
		zap.Int64("height", req.Height),
		zap.Int("txs", len(req.Txs)))

	// TODO: do something with the txs?

	return &abciTypes.ResponsePrepareProposal{
		Txs: req.Txs,
	}, nil
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
