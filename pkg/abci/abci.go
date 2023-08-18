package abci

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tendermintTypes "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/modules/datasets"
	"github.com/kwilteam/kwil-db/pkg/transactions"
	"github.com/kwilteam/kwil-db/pkg/validators"
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
		valUpdates:   make([]*validators.Validator, 0),

		log: log.NewNoOp(),

		commitWaiter: sync.WaitGroup{},

		// state: appState{height, ...}, // TODO
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

	// commitWaiter is a waitgroup that waits for the commit to finish
	// when a block is begun, the commitWaiter waits until the previous commit is finished
	// it then increments and starts "begin block"
	// when a commit is finished, the commitWaiter is decremented
	commitWaiter sync.WaitGroup

	// Expected AppState after bootstrapping the node with a given snapshot,
	// state gets updated with the bootupState after bootstrapping
	bootupState appState

	applicationVersion uint64
}

var _ abciTypes.Application = &AbciApp{}

// BeginBlock begins a block.
// If the previous commit is not finished, it will wait for the previous commit to finish.
func (a *AbciApp) BeginBlock(req abciTypes.RequestBeginBlock) abciTypes.ResponseBeginBlock {
	// TODO: replace this waitGroup with something else. It's not a queue and
	// Wait/Add is not atomic. Fortunately all consensus connections are
	// synchronous so there won't be more than one BeginBlock waiting.
	a.commitWaiter.Wait()
	a.commitWaiter.Add(1)

	err := a.committer.Begin(context.Background())
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
		a.log.Info("punish validator", zap.String("addr", addr))

		// This is why we need the addr=>pubkey map. Why, comet, why?
		pubkey, ok := a.valAddrToKey[addr]
		if !ok {
			panic(newFatalError("BeginBlock", &req, fmt.Sprintf("unknown validator address %v", addr)))
		}
		const punishDelta = 1
		newPower := ev.Validator.Power - punishDelta
		if err = a.validators.Punish(context.Background(), pubkey, newPower); err != nil {
			panic(newFatalError("BeginBlock", &req, fmt.Sprintf("failed to punish validator %v", addr)))
		}
	}

	return abciTypes.ResponseBeginBlock{}
}

func (a *AbciApp) CheckTx(incoming abciTypes.RequestCheckTx) abciTypes.ResponseCheckTx {
	tx := &transactions.Transaction{}
	err := tx.UnmarshalBinary(incoming.Tx)
	if err != nil {
		a.log.Error("failed to unmarshal transaction", zap.Error(err))
		return abciTypes.ResponseCheckTx{Code: 1, Log: err.Error()}
	}

	err = tx.Verify()
	if err != nil {
		a.log.Error("failed to verify transaction", zap.Error(err))
		return abciTypes.ResponseCheckTx{Code: 1, Log: err.Error()}
	}

	return abciTypes.ResponseCheckTx{Code: 0}
}

func (a *AbciApp) DeliverTx(req abciTypes.RequestDeliverTx) abciTypes.ResponseDeliverTx {
	ctx := context.Background()

	tx := &transactions.Transaction{}
	err := tx.UnmarshalBinary(req.Tx)
	if err != nil {
		return abciTypes.ResponseDeliverTx{
			Code: 1,
			Log:  err.Error(),
		}
	}

	var res *transactions.TransactionStatus
	var events []abciTypes.Event
	var gasUsed int64 // for error path

	switch tx.Body.PayloadType {
	case transactions.PayloadTypeDeploySchema:
		var schemaPayload transactions.Schema
		err = schemaPayload.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			break
		}

		var schema *engineTypes.Schema
		schema, err = datasets.ConvertSchemaToEngine(&schemaPayload)
		if err != nil {
			break
		}

		res, err = a.database.Deploy(ctx, schema, tx)
	case transactions.PayloadTypeDropSchema:
		drop := &transactions.DropSchema{}
		err = drop.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			break
		}

		res, err = a.database.Drop(ctx, drop.DBID, tx)
	case transactions.PayloadTypeExecuteAction:
		execution := &transactions.ActionExecution{}
		// Concept:
		// if res.Error != "" {
		// 	err = errors.New(res.Error)
		// 	gasUsed = res.Fee.Int64()
		// 	break
		// }

		err = execution.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			break
		}

		res, err = a.database.Execute(ctx, execution.DBID, execution.Action, convertArgs(execution.Arguments), tx)
	case transactions.PayloadTypeValidatorJoin:
		var join transactions.ValidatorJoin
		err = join.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			break
		}

		res, err = a.validators.Join(ctx, join.Candidate, int64(join.Power), tx)
		if err != nil {
			break
		}
		// Concept:
		// if res.Error != "" {
		// 	err = errors.New(res.Error)
		// 	gasUsed = res.Fee.Int64()
		// 	break
		// }

		events = []abciTypes.Event{
			{
				Type: "validator_join",
				Attributes: []abciTypes.EventAttribute{
					{Key: "Result", Value: "Success", Index: true},
					{Key: "ValidatorPubKey", Value: encodeBase64(join.Candidate), Index: true},
					{Key: "ValidatorPower", Value: fmt.Sprintf("%d", join.Power), Index: true},
				},
			},
		}
	case transactions.PayloadTypeValidatorLeave:
		var leave transactions.ValidatorLeave
		err = leave.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			break
		}

		res, err = a.validators.Leave(ctx, leave.Validator, tx)
		if err != nil {
			break
		}

		events = []abciTypes.Event{
			{
				Type: "remove_validator", // is this name arbitrary? it should be "validator_leave" for consistency
				Attributes: []abciTypes.EventAttribute{
					{Key: "Result", Value: "Success", Index: true},
					{Key: "ValidatorPubKey", Value: encodeBase64(leave.Validator), Index: true},
					{Key: "ValidatorPower", Value: "0", Index: true},
				},
			},
		}
	case transactions.PayloadTypeValidatorApprove:
		var approve transactions.ValidatorApprove
		err = approve.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			break
		}

		res, err = a.validators.Approve(ctx, approve.Candidate, tx)
		if err != nil {
			break
		}

		events = []abciTypes.Event{
			{
				Type: "validator_approve",
				Attributes: []abciTypes.EventAttribute{
					{Key: "Result", Value: "Success", Index: true},
					{Key: "CandidatePubKey", Value: encodeBase64(approve.Candidate), Index: true},
					{Key: "ApproverPubKey", Value: hex.EncodeToString(tx.Sender), Index: true},
				},
			},
		}
	default:
		err = fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String())
	}
	if err != nil {
		return abciTypes.ResponseDeliverTx{
			Code: 1,
			Log:  err.Error(),
			// NOTE: some execution that returned an error may still have used
			// gas. What is the meaning of the "Code"?
			GasUsed: gasUsed,
		}
	}

	return abciTypes.ResponseDeliverTx{
		Code:    abciTypes.CodeTypeOK,
		GasUsed: res.Fee.Int64(),
		Events:  events,
	}
}

func (a *AbciApp) EndBlock(e abciTypes.RequestEndBlock) abciTypes.ResponseEndBlock {
	a.log.Info("end block", zap.Int64("height", e.Height))

	a.valUpdates = a.validators.Finalize(context.Background())

	valUpdates := make([]abciTypes.ValidatorUpdate, len(a.valUpdates))
	for i, up := range a.valUpdates {
		valUpdates[i] = abciTypes.Ed25519ValidatorUpdate(up.PubKey, up.Power)
	}

	return abciTypes.ResponseEndBlock{
		ValidatorUpdates: valUpdates,
		ConsensusParamUpdates: &tendermintTypes.ConsensusParams{
			// we can include evidence in here for malicious actors, but this is not important this release
			Version: &tendermintTypes.VersionParams{
				App: a.applicationVersion,
			},
			Validator: &tendermintTypes.ValidatorParams{
				PubKeyTypes: []string{"ed25519"},
			},
		},
	}
}

func (a *AbciApp) Commit() abciTypes.ResponseCommit {
	ctx := context.Background()

	// generate the unique id for all changes occurred thus far
	id, err := a.committer.ID(ctx)
	if err != nil {
		panic(newFatalError("Commit", nil, fmt.Sprintf("failed to get commit id: %v", err)))
	}

	appHash, err := a.createNewAppHash(ctx, id)
	if err != nil {
		panic(newFatalError("Commit", nil, fmt.Sprintf("failed to create new app hash: %v", err)))
	}

	err = a.metadataStore.IncrementBlockHeight(ctx)
	if err != nil {
		panic(newFatalError("Commit", nil, fmt.Sprintf("failed to increment block height: %v", err)))
	}

	err = a.committer.Commit(ctx, func(err error) {
		if err != nil {
			panic(newFatalError("Commit", nil, fmt.Sprintf("failed to commit atomic commit: %v", err)))
		}

		a.commitWaiter.Done()
	})
	if err != nil {
		panic(newFatalError("Commit", nil, fmt.Sprintf("failed to commit atomic commit: %v", err)))
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

	return abciTypes.ResponseInfo{
		LastBlockHeight:  height,
		LastBlockAppHash: appHash,
		AppVersion:       a.applicationVersion,
	}
}

func (a *AbciApp) InitChain(p0 abciTypes.RequestInitChain) abciTypes.ResponseInitChain {
	ctx := context.Background()

	// Initialize the validator module with the genesis validators.
	vldtrs := make([]*validators.Validator, len(p0.Validators))
	for i := range p0.Validators {
		vi := &p0.Validators[i]
		// pk := vi.PubKey.GetEd25519()
		// if pk == nil { panic("only ed25519 validator keys are supported") }
		pk, err := vi.PubKey.Marshal()
		if err != nil {
			panic(fmt.Sprintf("invalid validator pubkey: %v", err))
		}
		vldtrs[i] = &validators.Validator{
			PubKey: pk,
			Power:  vi.Power,
		}
	}

	if err := a.validators.GenesisInit(context.Background(), vldtrs); err != nil {
		panic(fmt.Sprintf("GenesisInit failed: %v", err))
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

	return abciTypes.ResponseInitChain{
		Validators: valUpdates,
		AppHash:    apphash,
	}
}

func (a *AbciApp) ApplySnapshotChunk(p0 abciTypes.RequestApplySnapshotChunk) abciTypes.ResponseApplySnapshotChunk {
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
	snapshot := convertABCISnapshots(p0.Snapshot)
	if a.bootstrapper.OfferSnapshot(snapshot) != nil {
		return abciTypes.ResponseOfferSnapshot{Result: abciTypes.ResponseOfferSnapshot_REJECT}
	}
	a.bootupState.appHash = p0.Snapshot.Hash
	a.bootupState.height = int64(snapshot.Height)
	return abciTypes.ResponseOfferSnapshot{Result: abciTypes.ResponseOfferSnapshot_ACCEPT}
}

func (a *AbciApp) PrepareProposal(p0 abciTypes.RequestPrepareProposal) abciTypes.ResponsePrepareProposal {
	return abciTypes.ResponsePrepareProposal{}
}

func (a *AbciApp) ProcessProposal(p0 abciTypes.RequestProcessProposal) abciTypes.ResponseProcessProposal {
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
	return m.kv.Get(ctx, appHashKey)
}

func (m *metadataStore) SetAppHash(ctx context.Context, appHash []byte) error {
	return m.kv.Set(ctx, appHashKey, appHash)
}

func (m *metadataStore) GetBlockHeight(ctx context.Context) (int64, error) {
	height, err := m.kv.Get(ctx, blockHeightKey)
	if err != nil {
		return 0, err
	}

	return int64(binary.BigEndian.Uint64(height)), nil
}

func (m *metadataStore) SetBlockHeight(ctx context.Context, height int64) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(height))

	return m.kv.Set(ctx, blockHeightKey, buf)
}

func (m *metadataStore) IncrementBlockHeight(ctx context.Context) error {
	height, err := m.GetBlockHeight(ctx)
	if err != nil {
		return err
	}

	return m.SetBlockHeight(ctx, height+1)
}

// pubkeys in event attributes returned to comet as strings are base64 encoded,
// apparently.
func encodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
