package abci

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/ident"
	"github.com/kwilteam/kwil-db/internal/txapp"

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
	GasEnabled         bool
}

func NewAbciApp(cfg *AbciConfig, snapshotter SnapshotModule, bootstrapper DBBootstrapModule,
	txRouter TxApp, consensusParams *txapp.ConsensusParams, log log.Logger) *AbciApp {
	app := &AbciApp{
		cfg:             *cfg,
		bootstrapper:    bootstrapper,
		snapshotter:     snapshotter,
		txApp:           txRouter,
		consensusParams: consensusParams,

		log: log,

		validatorAddressToPubKey: make(map[string][]byte),
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

	// snapshotter is the snapshotter module that handles snapshotting
	snapshotter SnapshotModule

	// bootstrapper is the bootstrapper module that handles bootstrapping the database
	bootstrapper DBBootstrapModule

	log log.Logger

	// Expected AppState after bootstrapping the node with a given snapshot,
	// state gets updated with the bootupState after bootstrapping
	bootupState appState

	txApp TxApp

	consensusParams *txapp.ConsensusParams

	// height carries the height from FinalizeBlock to Commit for the snapshotter.
	// Ultimately this may not be required, or TxApp could provide the info.
	height uint64

	broadcastFn EventBroadcaster

	// validatorAddressToPubKey is a map of validator addresses to their public keys
	validatorAddressToPubKey map[string][]byte
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
			logger.Info("wrong chain ID",
				zap.String("payloadType", tx.Body.PayloadType.String()))
			return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: "wrong chain ID"}, nil
		}

		// Verify Signature
		err = ident.VerifyTransaction(tx)
		if err != nil {
			code = codeInvalidSignature
			logger.Debug("failed to verify transaction", zap.Error(err))
			return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: err.Error()}, nil
		}
	} else {
		logger.Info("Recheck", zap.String("hash", hex.EncodeToString(txHash)), zap.Uint64("nonce", tx.Body.Nonce), zap.String("payloadType", tx.Body.PayloadType.String()), zap.String("sender", hex.EncodeToString(tx.Sender)))
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

// FinalizeBlock is on the consensus connection. Note that according to CometBFT
// docs, "ResponseFinalizeBlock.app_hash is included as the Header.AppHash in
// the next block."
func (a *AbciApp) FinalizeBlock(ctx context.Context, req *abciTypes.RequestFinalizeBlock) (*abciTypes.ResponseFinalizeBlock, error) {
	fmt.Printf("\n\n")
	logger := a.log.With(zap.String("stage", "ABCI FinalizeBlock"), zap.Int64("height", req.Height))

	res := &abciTypes.ResponseFinalizeBlock{}

	err := a.txApp.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx commit failed: %w", err)
	}

	initialValidators, err := a.txApp.GetValidators(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current validators: %w", err)
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
		pubkey, ok := a.validatorAddressToPubKey[addr]
		if !ok {
			return nil, fmt.Errorf("unknown validator address %v", addr)
		}

		const punishDelta = 1
		newPower := ev.Validator.Power - punishDelta
		if err = a.txApp.UpdateValidator(ctx, pubkey, newPower); err != nil {
			return nil, fmt.Errorf("failed to punish validator: %w", err)
		}
	}

	addr := proposerAddrToString(req.ProposerAddress)
	proposerPubKey, ok := a.validatorAddressToPubKey[addr]
	if !ok && len(req.Txs) > 0 {
		// ProcessProposal allows block proposals for untracked validators, but
		// only if the block has no transactions.
		return nil, fmt.Errorf("failed to find proposer pubkey corresponding to address %v", addr)
	}

	for _, tx := range req.Txs {
		decoded := &transactions.Transaction{}
		err := decoded.UnmarshalBinary(tx)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
		}

		txRes := a.txApp.Execute(txapp.TxContext{
			BlockHeight:     req.Height,
			Proposer:        proposerPubKey,
			ConsensusParams: *a.consensusParams,
			Ctx:             ctx,
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

	res.ConsensusParamUpdates = &tendermintTypes.ConsensusParams{ // why are we "updating" these on every block? Should be nil for no update.
		// we can include evidence in here for malicious actors, but this is not important this release
		Version: &tendermintTypes.VersionParams{
			App: a.cfg.ApplicationVersion, // how would we change the application version?
		},
		Validator: &tendermintTypes.ValidatorParams{
			PubKeyTypes: []string{"ed25519"},
		},
	}

	// Broadcast any events that have not been broadcasted yet
	if a.broadcastFn != nil && len(proposerPubKey) > 0 {
		err := a.broadcastFn(ctx, proposerPubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to broadcast events: %w", err)
		}
	}

	// Get the new validator set and apphash from txApp.
	appHash, finalValidators, err := a.txApp.Finalize(ctx, req.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to finalize transaction app: %w", err)
	}
	res.AppHash = appHash
	a.height = uint64(req.Height) // remember for Commit

	valUpdates := validatorUpdates(initialValidators, finalValidators)

	res.ValidatorUpdates = make([]abciTypes.ValidatorUpdate, len(valUpdates))
	for i, up := range valUpdates {
		addr, err := pubkeyToAddr(up.PubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pubkey to address: %w", err)
		}
		if up.Power == 0 {
			delete(a.validatorAddressToPubKey, addr)
		} else {
			a.validatorAddressToPubKey[addr] = up.PubKey // there may be new validators we need to add
		}

		res.ValidatorUpdates[i] = abciTypes.Ed25519ValidatorUpdate(up.PubKey, up.Power)
	}

	return res, nil
}

// validatorUpdates returns the added, removed, and updated validators in the given block.
func validatorUpdates(initial, final []*types.Validator) []*types.Validator {
	initialVals := make(map[string]*types.Validator)
	for _, val := range initial {
		initialVals[hex.EncodeToString(val.PubKey)] = val
	}

	finalVals := make(map[string]*types.Validator)
	for _, val := range final {
		finalVals[hex.EncodeToString(val.PubKey)] = val
	}

	var updates []*types.Validator
	// check for newly added and updated validators
	for key, val := range finalVals {
		if initialVal, ok := initialVals[key]; ok {
			if initialVal.Power != val.Power {
				// Validator is in the initial set, but has updated power
				updates = append(updates, val)
			}
		} else {
			// Validator is not in the initial set, so it must be new
			updates = append(updates, val)
		}
	}

	// check for removed validators
	for key, val := range initialVals {
		if _, ok := finalVals[key]; !ok {
			// Validator is in the initial set, but not in the final set
			updates = append(updates, &types.Validator{
				PubKey: val.PubKey,
				Power:  0,
			})
		}
	}

	return updates
}

// Commit persists the state changes. This is called under mempool lock in
// cometbft, unlike FinalizeBlock.
func (a *AbciApp) Commit(ctx context.Context, _ *abciTypes.RequestCommit) (*abciTypes.ResponseCommit, error) {
	err := a.txApp.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction app: %w", err)
	}

	// snapshotting
	if a.snapshotter != nil && a.snapshotter.IsSnapshotDue(a.height) {
		// TODO: Lock all DBs
		err = a.snapshotter.CreateSnapshot(a.height)
		if err != nil {
			a.log.Error("snapshot creation failed", zap.Error(err))
		}
		// Unlock all the DBs
	}

	return &abciTypes.ResponseCommit{}, nil // RetainHeight stays 0 to not prune any blocks
}

// Info is part of the Info/Query connection. CometBFT will call this during
// it's handshake, and if height 0 is returned it will then use InitChain. This
// method should also be usable at any time (and read only) as it is used for
// cometbft's /abci_info RPC endpoint.
//
// CometBFT docs say:
//   - LastBlockHeight is the "Latest height for which the app persisted its state"
//   - LastBlockAppHash is the "Latest AppHash returned by FinalizeBlock"
//   - "ResponseFinalizeBlock.app_hash is included as the Header.AppHash in the
//     next block." Notably, the *next* block's header. This is verifiable with
//     the /block RPC endpoint.
//
// Thus, LastBlockAppHash is not NOT AppHash in the header of the block at the
// returned height. Our meta data store has to track the above values, where the
// stored app hash corresponds to the block at height+1. This is simple, but the
// discrepancy is worth noting.
func (a *AbciApp) Info(ctx context.Context, _ *abciTypes.RequestInfo) (*abciTypes.ResponseInfo, error) {
	height, appHash, err := a.txApp.ChainInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("chainInfo: %w", err)
	}

	validators, err := a.txApp.GetValidators(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get validators: %w", err)
	}

	for _, val := range validators {
		addr, err := pubkeyToAddr(val.PubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pubkey to address: %w", err)
		}

		a.validatorAddressToPubKey[addr] = val.PubKey
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
	genesisAllocs := make([]*types.Account, 0, len(a.cfg.GenesisAllocs))
	for acct, bal := range a.cfg.GenesisAllocs {
		acct, _ := strings.CutPrefix(acct, "0x") // special case for ethereum addresses
		identifier, err := hex.DecodeString(acct)
		if err != nil {
			return nil, fmt.Errorf("invalid hex pubkey: %w", err)
		}

		genesisAllocs = append(genesisAllocs, &types.Account{
			Identifier: identifier,
			Balance:    bal,
		})
	}
	// Initialize the validator module with the genesis validators.
	vldtrs := make([]*types.Validator, len(req.Validators))
	for i := range req.Validators {
		vi := &req.Validators[i]
		pk := vi.PubKey.GetEd25519()
		vldtrs[i] = &types.Validator{
			PubKey: pk,
			Power:  vi.Power,
		}

		addr, err := pubkeyToAddr(pk)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pubkey to address: %w", err)
		}

		a.validatorAddressToPubKey[addr] = pk
	}

	if err := a.txApp.GenesisInit(ctx, vldtrs, genesisAllocs, req.InitialHeight,
		a.cfg.GenesisAppHash); err != nil {
		return nil, fmt.Errorf("txApp.GenesisInit failed: %w", err)
	}

	valUpdates := make([]abciTypes.ValidatorUpdate, len(vldtrs))
	for i, validator := range vldtrs {
		valUpdates[i] = abciTypes.Ed25519ValidatorUpdate(validator.PubKey, validator.Power)
	}

	logger.Info("initialized chain", zap.String("app hash", fmt.Sprintf("%x", a.cfg.GenesisAppHash)))

	return &abciTypes.ResponseInitChain{
		Validators: valUpdates,
		AppHash:    a.cfg.GenesisAppHash, // doesn't matter what we gave the node in GenesisDoc, this is it
	}, nil
}

// ApplySnapshotChunk is on the state sync connection
func (a *AbciApp) ApplySnapshotChunk(ctx context.Context, req *abciTypes.RequestApplySnapshotChunk) (*abciTypes.ResponseApplySnapshotChunk, error) {
	refetchChunks, status, err := a.bootstrapper.ApplySnapshotChunk(req.Chunk, req.Index)
	if err != nil {
		return &abciTypes.ResponseApplySnapshotChunk{Result: abciStatus(status), RefetchChunks: refetchChunks}, nil
	}

	if a.bootstrapper.IsDBRestored() {
		// NOTE: when snapshot is implemented, the bootstrapper should be able
		// to meta.SetChainState.
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

// prepareBlockTransactions prepares the transactions for the block we are proposing.
// The input transactions are from mempool direct from cometbft, and we modify
// the list for our purposes. This includes ensuring transactions from the same
// sender in ascending nonce-order, enforcing the max bytes limit, etc.
// This also includes the proposer's transactions, which are not in the mempool.
// The transaction ordering is as follows:
// MempoolProposerTxns, ProposerInjectedTxns, MempoolTxns by other senders
func (a *AbciApp) prepareBlockTransactions(txs [][]byte, log *log.Logger, maxTxBytes int64, proposerAddr []byte) [][]byte {
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

	// Grab the bytes rather than re-marshalling.
	nonces := make([]uint64, 0, len(okTxns))
	var propTxs, otherTxns []*indexedTxn
	// otherTxns := make(map[string][]*indexedTxn)
	i = 0
	proposerNonce := uint64(0)

	// Enforce nonce ordering and remove transactions from unfunded accounts.
	for _, tx := range okTxns {
		if i > 0 && tx.Body.Nonce == nonces[i-1] && bytes.Equal(tx.Sender, okTxns[i-1].Sender) {
			log.Warn(fmt.Sprintf("Dropping tx with re-used nonce %d from block proposal", tx.Body.Nonce))
			continue // mempool recheck should have removed this
		}

		// Drop transactions from unfunded accounts in gasEnabled mode
		if a.cfg.GasEnabled {
			balance, nonce, err := a.txApp.AccountInfo(context.Background(), tx.Sender, false)
			if err != nil {
				log.Error("failed to get account info", zap.Error(err))
				continue
			}
			if nonce == 0 && balance.Sign() == 0 {
				log.Warn("Dropping tx from unfunded account while preparing the block", zap.String("sender", hex.EncodeToString(tx.Sender)))
				continue
			}
		}

		if bytes.Equal(tx.Sender, proposerAddr) {
			proposerNonce = tx.Body.Nonce + 1
			propTxs = append(propTxs, tx)
		} else {
			// Append the transaction to the final list.
			// sender := string(tx.Sender)
			otherTxns = append(otherTxns, tx)
			// otherTxns[sender] = append(otherTxns[sender], tx)
		}
		nonces = append(nonces, tx.Body.Nonce)
		i++
	}

	// TODO: truncate based on our max block size since we'll have to set
	// ConsensusParams.Block.MaxBytes to -1 so that we get ALL transactions even
	// if it goes beyond max_tx_bytes.  See:
	// https://github.com/cometbft/cometbft/pull/1003
	// https://docs.cometbft.com/v0.38/spec/abci/abci++_methods#prepareproposal
	// https://github.com/cometbft/cometbft/issues/980

	// Enforce block size limits
	// Txs order: MempoolProposerTxns, ProposerInjectedTxns, MempoolTxns
	var finalTxs [][]byte

	for _, tx := range propTxs {
		txBts := txs[tx.is]
		txSize := int64(len(txBts))
		if maxTxBytes < txSize {
			break
		}
		maxTxBytes -= txSize
		finalTxs = append(finalTxs, txBts)
	}

	proposerTxs, err := a.txApp.ProposerTxs(context.Background(), proposerNonce, maxTxBytes, proposerAddr)
	if err != nil {
		log.Error("failed to get proposer transactions", zap.Error(err))
	}

	for _, tx := range proposerTxs {
		txSize := int64(len(tx))
		if maxTxBytes < txSize {
			break
		}
		maxTxBytes -= txSize
		finalTxs = append(finalTxs, tx)
	}

	// senders tracks the sender of transactions that has pushed over the bytes limit for the block.
	// If a sender is in the senders, skip all subsequent transactions from the sender
	// because nonces need to be sequential.
	// Keep checking transactions for other senders that may be smaller and fit in the remaining space.
	senders := make(map[string]bool)

	for _, tx := range otherTxns {
		sender := string(tx.Sender)
		// If we have already added a transaction from this sender, skip it.
		if _, ok := senders[sender]; ok {
			continue
		}

		txSize := int64(len(txs[tx.is]))
		if maxTxBytes < txSize {
			// Ignore the rest of the transactions by this sender
			senders[sender] = true
			break
		}
		maxTxBytes -= txSize
		finalTxs = append(finalTxs, txs[tx.is])
	}

	return finalTxs
}

func (a *AbciApp) PrepareProposal(ctx context.Context, req *abciTypes.RequestPrepareProposal) (*abciTypes.ResponsePrepareProposal, error) {
	logger := a.log.With(zap.String("stage", "ABCI PrepareProposal"),
		zap.Int64("height", req.Height),
		zap.Int("txs", len(req.Txs)))

	pubKey, ok := a.validatorAddressToPubKey[proposerAddrToString(req.ProposerAddress)]
	if !ok {
		// there is an edge case where cometbft will allow a node to PrepareProposal
		// even if it is not a validator, if it was a validator in the most recent block
		// we check for this here and simply propose an empty block, since it will be rejected
		a.log.Warn("local node was made proposer, but is no longer a validator")
		return &abciTypes.ResponsePrepareProposal{}, nil
	}

	okTxns := a.prepareBlockTransactions(req.Txs, &a.log, req.MaxTxBytes, pubKey)
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

	return &abciTypes.ResponsePrepareProposal{
		Txs: okTxns,
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
		_, nonce, err := a.txApp.AccountInfo(ctx, []byte(sender), false)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}
		expectedNonce := uint64(nonce) + 1

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

	proposerPubKey, ok := a.validatorAddressToPubKey[proposerAddrToString(req.ProposerAddress)]
	if !ok {
		// there is an edge case where cometbft will allow a node to PrepareProposal
		// even if it is not a validator, if it was a validator in the most recent block.
		// a well behaved validator will submit an empty block here, which we will accept.
		// if not an empty block, we will reject it.

		if len(req.Txs) == 0 {
			a.log.Info("proposer is not a validator and submitted an empty block, accepting proposal")
			return &abciTypes.ResponseProcessProposal{Status: abciTypes.ResponseProcessProposal_ACCEPT}, nil
		}

		a.log.Warn("proposer is not a validator and submitted a non-empty block, rejecting proposal", zap.String("proposer", proposerAddrToString(req.ProposerAddress)))
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

type EventBroadcaster func(ctx context.Context, proposer []byte) error

func (a *AbciApp) SetEventBroadcaster(fn EventBroadcaster) {
	a.broadcastFn = fn
}
