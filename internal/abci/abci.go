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
	"strings"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/ident"
	"github.com/kwilteam/kwil-db/internal/kv"
	"github.com/kwilteam/kwil-db/internal/txapp"
	"github.com/kwilteam/kwil-db/internal/validators"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tendermintTypes "github.com/cometbft/cometbft/proto/tendermint/types"
	cmtTypes "github.com/cometbft/cometbft/types"
	"go.uber.org/zap"
)

// appState is an in-memory representation of the state of the application.
type appState struct {
	height  int64
	appHash []byte
}

// AbciConfig includes data that defines the chain and allow the application to
// satisfy the ABCI Application interface.
type AbciConfig struct {
	GenesisAppHash     []byte
	ChainID            string
	ApplicationVersion uint64
	GenesisAllocs      map[string]*big.Int
	// genesisCft.Alloc
	// GasEnabled bool // for mempool policy modification
}

func NewAbciApp(cfg *AbciConfig, accounts AccountsModule, vldtrs ValidatorModule, kv KVStore,
	snapshotter SnapshotModule, bootstrapper DBBootstrapModule, txRouter TxApp, consensusParams ConsensusParams, log log.Logger) *AbciApp {
	app := &AbciApp{
		cfg:        *cfg,
		validators: vldtrs,
		metadataStore: &metadataStore{
			kv: kv,
		},
		bootstrapper:    bootstrapper,
		snapshotter:     snapshotter,
		accounts:        accounts,
		txApp:           txRouter,
		consensusParams: consensusParams,

		valAddrToKey: make(map[string][]byte),

		log: log,
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

// proposerAddrToString converts a proposer address to a string.
// This follows the semantics of comet's ed25519.Pubkey.Address() method,
// which hex encodes and upper cases the address
func proposerAddrToString(addr []byte) string {
	return strings.ToUpper(hex.EncodeToString(addr))
}

type AbciApp struct {
	cfg AbciConfig

	// validators is the validators module that handles joining and approving validators
	validators ValidatorModule
	// comet punishes by address, so we maintain an address=>pubkey map.
	valAddrToKey map[string][]byte // NOTE: includes candidates

	// snapshotter is the snapshotter module that handles snapshotting
	snapshotter SnapshotModule

	// bootstrapper is the bootstrapper module that handles bootstrapping the database
	bootstrapper DBBootstrapModule

	// metadataStore to track the app hash and block height
	metadataStore *metadataStore

	// accountStore is the store that maintains the account state
	accounts AccountsModule

	log log.Logger

	// Expected AppState after bootstrapping the node with a given snapshot,
	// state gets updated with the bootupState after bootstrapping
	bootupState appState

	txApp TxApp

	consensusParams ConsensusParams

	// commitHooks is a set of functions that are called after a block is committed.
	// If any of the hooks return an error, the rest of the hooks will be called,
	// and the node will then fail with a consensus error.
	// This can be used to build extra logic within a node, such as broadcasting validator txa,
	// sending data to Arweave, etc.
	// It is placed within abci's `Commit` method, since abci blocks mempool updates during commit.
	// This allows validators to automatically author transactions, and not conflict with potential
	// transactions authored by the same validator's administrator.
	commitHooks []CommitHook

	// blockInfo is information about the last block that was finalized.
	// it is set and at the end of FinalizeBlock.
	// It is primarily made to be used in Commit.  If used elsewhere,
	// the implementer should take care to ensure that they account for potential
	// nil values (in the case of the node not yet having finalized a block).
	blockInfo *BlockInfo
}

func (a *AbciApp) ChainID() string {
	return a.cfg.ChainID
}

var _ abciTypes.Application = &AbciApp{}

// The Application interface methods in four groups according to the
// "connection" used by CometBFT to interact with the application. Calls to the
// methods within a connection are synchronized. They are not synchronized
// between the connections. e.g. CheckTx calls from the mempool connection can
// occur concurrent to calls on the Consensus connection.

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
func (a *AbciApp) CheckTx(ctx context.Context, incoming *abciTypes.RequestCheckTx) (*abciTypes.ResponseCheckTx, error) {
	newTx := incoming.Type == abciTypes.CheckTxType_New
	logger := a.log.With(zap.Bool("recheck", !newTx))
	logger.Debug("check tx")

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
		return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: err.Error()}, nil // return error now or is it still all about code?
	}

	txHash := cmtTypes.Tx(incoming.Tx).Hash()
	logger.Debug("",
		zap.String("sender", hex.EncodeToString(tx.Sender)),
		zap.String("PayloadType", tx.Body.PayloadType.String()),
		zap.Uint64("nonce", tx.Body.Nonce),
		zap.String("hash", hex.EncodeToString(txHash)))

	// For a new transaction (not re-check), before looking at execution cost or
	// checking nonce validity, ensure the payload is recognized and signature is valid.
	if newTx {
		// Verify the correct chain ID is set, if it is set.
		if protected := tx.Body.ChainID != ""; protected && tx.Body.ChainID != a.cfg.ChainID {
			code = codeWrongChain
			logger.Info("wrong chain ID", zap.String("payloadType", tx.Body.PayloadType.String()))
			return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: "wrong chain ID"}, nil
		}
		// we no longer need to do this, it fails in the mempool itself
		// // Verify Payload type
		// if !tx.Body.PayloadType.Valid() {
		// 	code = codeInvalidTxType
		// 	logger.Debug("invalid payload type", zap.String("payloadType", tx.Body.PayloadType.String()))
		// 	return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: "invalid payload type"}, nil
		// }

		// Verify Signature
		err = ident.VerifyTransaction(tx)
		if err != nil {
			code = codeInvalidSignature
			logger.Debug("failed to verify transaction", zap.Error(err))
			return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: err.Error()}, nil
		}
	} else {
		logger.Info("Recheck", zap.String("hash", hex.EncodeToString(txHash)), zap.Uint64("nonce", tx.Body.Nonce))
	}

	err = a.txApp.ApplyMempool(ctx, tx)
	if err != nil {
		if errors.Is(err, transactions.ErrInvalidNonce) {
			code = codeInvalidNonce
			logger.Info("received transaction with invalid nonce", zap.Uint64("nonce", tx.Body.Nonce), zap.Error(err))
		} else if errors.Is(err, transactions.ErrInvalidAmount) {
			code = codeInvalidAmount
			logger.Info("received transaction with invalid amount", zap.Uint64("nonce", tx.Body.Nonce), zap.Error(err))
		} else if errors.Is(err, transactions.ErrInsufficientBalance) {
			code = codeInsufficientBalance
			logger.Info("transaction sender has insufficient balance", zap.Uint64("nonce", tx.Body.Nonce), zap.Error(err))
		} else {
			code = codeUnknownError
			logger.Warn("unexpected failure to verify transaction against local mempool state", zap.Error(err))
		}
		return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: err.Error()}, nil
	}

	return &abciTypes.ResponseCheckTx{Code: code.Uint32()}, nil
}

// FinalizeBlock is on the consensus connection
func (a *AbciApp) FinalizeBlock(ctx context.Context, req *abciTypes.RequestFinalizeBlock) (*abciTypes.ResponseFinalizeBlock, error) {
	logger := a.log.With(zap.String("stage", "ABCI FinalizeBlock"), zap.Int("height", int(req.Height)))

	res := &abciTypes.ResponseFinalizeBlock{}

	// BeginBlock was this part
	err := a.txApp.Begin(ctx, req.Height)
	if err != nil {
		return nil, fmt.Errorf("begin atomic commit failed: %w", err)
	}

	// Punish bad validators.
	for _, ev := range req.Misbehavior {
		addr := proposerAddrToString(ev.Validator.Address) // comet example app confirms this conversion... weird
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

	addr := proposerAddrToString(req.ProposerAddress)
	proposerPubKey, ok := a.valAddrToKey[addr]
	if !ok {
		a.log.Warn("received block proposal from unknown validator", zap.String("addr", addr))
		return nil, fmt.Errorf("failed to find proposer pubkey corresponding to address %v", addr)
	}

	for _, tx := range req.Txs {
		// DeliverTx was the part in this loop.
		decoded := &transactions.Transaction{}
		err := decoded.UnmarshalBinary(tx)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
		}

		txRes := a.txApp.Execute(txapp.TxContext{
			BlockHeight:  req.Height,
			Proposer:     proposerPubKey,
			VotingPeriod: a.consensusParams.VotingPeriod(),
			Ctx:          ctx,
		}, decoded)

		abciRes := &abciTypes.ExecTxResult{}
		if txRes.Error != nil {
			abciRes.Log = txRes.Error.Error()
			a.log.Warn("failed to execute transaction", zap.Error(txRes.Error))
		} else {
			abciRes.Log = "success"
		}
		abciRes.Code = txRes.ResponseCode.Uint32()
		abciRes.GasUsed = txRes.Spend

		res.TxResults = append(res.TxResults, abciRes)
	}

	// EndBlock was this part
	res.ConsensusParamUpdates = &tendermintTypes.ConsensusParams{ // why are we "updating" these on every block? Should be nil for no update.
		// we can include evidence in here for malicious actors, but this is not important this release
		Version: &tendermintTypes.VersionParams{
			App: a.cfg.ApplicationVersion, // how would we change the application version?
		},
		Validator: &tendermintTypes.ValidatorParams{
			PubKeyTypes: []string{"ed25519"},
		},
	}

	for _, hook := range a.commitHooks {
		err := hook(ctx, a.blockInfo)
		if err != nil {
			fmt.Println("commit hook failed", err)
			return nil, fmt.Errorf("commit hook failed: %w", err)
		}
	}

	// TODO: some of this commit logic would actually go in Commit, but we need
	// to reconcile some issues with our idempotent commit process first.
	// while we have idempotent commits, it is ok to have this here.
	newAppHash, validatorUpdates, err := a.txApp.Commit(ctx, req.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction router: %w", err)
	}

	res.ValidatorUpdates = make([]abciTypes.ValidatorUpdate, len(validatorUpdates))
	for i, up := range validatorUpdates {
		res.ValidatorUpdates[i] = abciTypes.Ed25519ValidatorUpdate(up.PubKey, up.Power)

		addr, err := pubkeyToAddr(up.PubKey)
		if err != nil {
			return nil, fmt.Errorf("invalid validator pubkey: %w", err)
		}

		if up.Power < 1 { // leave or punish
			delete(a.valAddrToKey, addr)
		} else { // add or update without remove
			a.valAddrToKey[addr] = up.PubKey
		}
	}

	appHash, err := a.createNewAppHash(ctx, newAppHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create new app hash: %w", err)
	}
	res.AppHash = appHash

	a.blockInfo = &BlockInfo{
		Height:    req.Height,
		BlockHash: req.Hash,
		Proposer:  proposerPubKey,
	}

	return res, nil
}

func (a *AbciApp) Commit(ctx context.Context, _ *abciTypes.RequestCommit) (*abciTypes.ResponseCommit, error) {
	err := a.metadataStore.IncrementBlockHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to increment block height: %w", err)
	}

	height, err := a.metadataStore.GetBlockHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get block height: %w", err)
	}

	// snapshotting
	if a.snapshotter != nil && a.snapshotter.IsSnapshotDue(uint64(height)) {
		// TODO: Lock all DBs
		err = a.snapshotter.CreateSnapshot(uint64(height))
		if err != nil {
			a.log.Error("snapshot creation failed", zap.Error(err))
		}
		// Unlock all the DBs
	}

	return &abciTypes.ResponseCommit{}, nil // RetainHeight stays 0 to not prune any blocks
}

// Info is part of the Info/Query connection.
func (a *AbciApp) Info(ctx context.Context, req *abciTypes.RequestInfo) (*abciTypes.ResponseInfo, error) {
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

	a.log.Info("ABCI application is ready", zap.Int64("height", height))

	return &abciTypes.ResponseInfo{
		LastBlockHeight:  height,
		LastBlockAppHash: appHash,
		// Version: kwildVersion, // the *software* semver string
		AppVersion: a.cfg.ApplicationVersion,
	}, nil
}

func (a *AbciApp) InitChain(ctx context.Context, req *abciTypes.RequestInitChain) (*abciTypes.ResponseInitChain, error) {
	logger := a.log.With(zap.String("stage", "ABCI InitChain"), zap.Int64("height", req.InitialHeight))
	logger.Debug("", zap.String("ChainId", req.ChainId))
	// maybe verify a.cfg.ChainID against the one in the request

	// Store the genesis account allocations to the datastore. These are
	// reflected in the genesis app hash.
	for acct, bal := range a.cfg.GenesisAllocs {
		acct, _ := strings.CutPrefix(acct, "0x") // special case for ethereum addresses
		identifier, err := hex.DecodeString(acct)
		if err != nil {
			return nil, fmt.Errorf("invalid hex pubkey: %w", err)
		}
		if err = a.accounts.Credit(ctx, identifier, bal); err != nil {
			return nil, fmt.Errorf("failed to credit genesis account: %w", err)
		}
	}
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

		addr, err := pubkeyToAddr(pk)
		if err != nil {
			return nil, fmt.Errorf("invalid validator pubkey: %w", err)
		}

		a.valAddrToKey[addr] = pk
	}

	if err := a.validators.GenesisInit(ctx, vldtrs, req.InitialHeight); err != nil {
		return nil, fmt.Errorf("validators.GenesisInit failed: %w", err)
	}

	if err := a.txApp.GenesisInit(ctx, vldtrs); err != nil {
		return nil, fmt.Errorf("txApp.GenesisInit failed: %w", err)
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

	return &abciTypes.ResponseInitChain{
		Validators: valUpdates,
		AppHash:    a.cfg.GenesisAppHash,
	}, nil
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

// ExtendVote creates an application specific vote extension.
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
	return &abciTypes.ResponseExtendVote{}, nil
}

// Verify application's vote extension data
func (a *AbciApp) VerifyVoteExtension(ctx context.Context, req *abciTypes.RequestVerifyVoteExtension) (*abciTypes.ResponseVerifyVoteExtension, error) {
	if len(req.VoteExtension) > 0 {
		// We recognize no vote extensions yet.
		return &abciTypes.ResponseVerifyVoteExtension{
			Status: abciTypes.ResponseVerifyVoteExtension_REJECT,
		}, nil
	}
	return &abciTypes.ResponseVerifyVoteExtension{
		Status: abciTypes.ResponseVerifyVoteExtension_ACCEPT,
	}, nil
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
func prepareMempoolTxns(txs [][]byte, maxBytes int, log *log.Logger, proposerAddr []byte) ([][]byte, uint64) {
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
	proposerNonce := uint64(0)

	for _, tx := range okTxns {
		if i > 0 && tx.Body.Nonce == nonces[i-1] && bytes.Equal(tx.Sender, okTxns[i-1].Sender) {
			log.Warn(fmt.Sprintf("Dropping tx with re-used nonce %d from block proposal", tx.Body.Nonce))
			continue // mempool recheck should have removed this
		}

		if bytes.Equal(tx.Sender, proposerAddr) {
			proposerNonce = tx.Body.Nonce + 1
		}

		finalTxns = append(finalTxns, txs[tx.is])
		nonces = append(nonces, tx.Body.Nonce)
		i++
	}

	return finalTxns, proposerNonce
}

func (a *AbciApp) PrepareProposal(ctx context.Context, req *abciTypes.RequestPrepareProposal) (*abciTypes.ResponsePrepareProposal, error) {
	logger := a.log.With(zap.String("stage", "ABCI PrepareProposal"),
		zap.Int64("height", req.Height),
		zap.Int("txs", len(req.Txs)))

	addr, ok := a.valAddrToKey[proposerAddrToString(req.ProposerAddress)]
	if !ok {
		return nil, fmt.Errorf("failed to find proposer pubkey corresponding to %s", proposerAddrToString(req.ProposerAddress))
	}

	okTxns, proposerNonce := prepareMempoolTxns(req.Txs, int(req.MaxTxBytes), &a.log, addr)
	if len(okTxns) != len(req.Txs) {
		logger.Info("PrepareProposal: number of transactions in proposed block has changed!",
			zap.Int("in", len(req.Txs)), zap.Int("out", len(okTxns)))
	} else if logger.L.Level() <= log.DebugLevel { // spare the check if it wouldn't be logged
		for i, tx := range okTxns {
			if !bytes.Equal(tx, req.Txs[i]) { //  not and error, just notable
				logger.Debug("transaction was moved or mutated", zap.Int("idx", i))
			}
		}
	}

	if proposerNonce == 0 {
		acct, err := a.accounts.GetAccount(ctx, addr)
		if err != nil {
			return nil, fmt.Errorf("failed to get proposer account: %w", err)
		}
		proposerNonce = uint64(acct.Nonce) + 1
	}

	proposerTxs, err := a.txApp.ProposerTxs(ctx, proposerNonce)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposer transactions: %w", err)
	}

	proposerTxBts := make([][]byte, 0)
	for _, tx := range proposerTxs {
		bts, err := tx.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal proposer transaction: %w", err)
		}
		proposerTxBts = append(proposerTxBts, bts)
		fmt.Println("Proposer inserted tx: ", tx.Body.Nonce, " proposer addr: ", hex.EncodeToString(tx.Sender))
	}

	txs := append(req.Txs, proposerTxBts...)

	return &abciTypes.ResponsePrepareProposal{
		Txs: txs,
	}, nil
}

func (a *AbciApp) validateProposalTransactions(ctx context.Context, txns [][]byte, proposer []byte) error {
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

			// if it is a vote body payload, then only the proposer can propose it
			// this is a hard consensus rule for block building, and is protected by
			// the mempool.
			// it seems like this should somehow be in the same package as mempool since this is inter-related
			// logically, but I am putting it here for now.
			if tx.Body.PayloadType == transactions.PayloadTypeValidatorVoteBodies && !bytes.Equal(proposer, tx.Sender) {
				return fmt.Errorf("only proposer can propose validator vote bodies")
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
func (a *AbciApp) ProcessProposal(ctx context.Context, req *abciTypes.RequestProcessProposal) (*abciTypes.ResponseProcessProposal, error) {
	logger := a.log.With(zap.String("stage", "ABCI ProcessProposal"),
		zap.Int64("height", req.Height),
		zap.Int("txs", len(req.Txs)))

	addr := proposerAddrToString(req.ProposerAddress)
	proposerPubKey, ok := a.valAddrToKey[addr]
	if !ok {
		a.log.Warn("received block proposal from unknown validator", zap.String("addr", addr))
		return &abciTypes.ResponseProcessProposal{Status: abciTypes.ResponseProcessProposal_REJECT}, nil
	}

	if err := a.validateProposalTransactions(ctx, req.Txs, proposerPubKey); err != nil {
		logger.Warn("rejecting block proposal", zap.Error(err))
		return &abciTypes.ResponseProcessProposal{Status: abciTypes.ResponseProcessProposal_REJECT}, nil
	}

	// TODO: Verify the Tx and Block sizes based on the genesis configuration
	return &abciTypes.ResponseProcessProposal{Status: abciTypes.ResponseProcessProposal_ACCEPT}, nil
}

func (a *AbciApp) Query(ctx context.Context, req *abciTypes.RequestQuery) (*abciTypes.ResponseQuery, error) {
	return &abciTypes.ResponseQuery{}, nil
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

// BlockInfo contains information about a block.
type BlockInfo struct {
	Height    int64  // block height
	Proposer  []byte // public key of the proposer
	BlockHash []byte // hash of the block
}

type CommitHook func(ctx context.Context, block *BlockInfo) error

// AddCommitHook adds a function to be called after a block is committed.
// We allow this to be done ad-hoc, to give the application control over when
// their commit hook begins taking place (e.g. before blocksync, after blocksync, etc.)
func (a *AbciApp) AddCommitHook(hook CommitHook) {
	a.commitHooks = append(a.commitHooks, hook)
}
