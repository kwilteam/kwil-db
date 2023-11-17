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
	NodeAddress        string
}

func NewAbciApp(cfg *AbciConfig, accounts AccountsModule, database DatasetsModule, vldtrs ValidatorModule, kv KVStore,
	committer AtomicCommitter, snapshotter SnapshotModule, bootstrapper DBBootstrapModule, bridgeEvents BridgeEventsModule, opts ...AbciOpt) *AbciApp {
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
		chainEvents:  bridgeEvents,

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

	chainEvents BridgeEventsModule

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
func (a *AbciApp) AccountInfo(ctx context.Context, pubkey []byte) (balance *big.Int, nonce int64, err error) {
	// If we have any unconfirmed transactions for the user, report that info
	// without even checking the account store.
	ua := a.mempool.peekAccountInfo(ctx, pubkey)
	if ua.nonce > 0 {
		return ua.balance, ua.nonce, nil
	}
	// Nothing in mempool, check account store. Changes to the account store are
	// committed before the mempool is cleared, so there should not be any race.
	acct, err := a.accounts.GetAccount(ctx, pubkey)
	if err != nil {
		return nil, 0, err
	}
	return acct.Balance, acct.Nonce, nil
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

	// Reject any governance transaction that should be inserted by the block
	// proposer rather than broadcasted. This would be like broadcasting a
	// coinbase transaction.
	switch tx.Body.PayloadType {
	case transactions.PayloadTypeAccountCredit: // list others here
		code = codeInvalidTxType
		return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: "governance transaction not valid in mempool"}, nil
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
			return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: "wrong chain ID"}, nil
		}
		// Verify Payload type
		if !tx.Body.PayloadType.Valid() {
			code = codeInvalidTxType
			logger.Debug("invalid payload type", zap.String("payloadType", tx.Body.PayloadType.String()))
			return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: "invalid payload type"}, nil
		}

		// Verify Signature
		err = ident.VerifyTransaction(tx)
		if err != nil {
			code = codeInvalidSignature
			logger.Debug("failed to verify transaction", zap.Error(err))
			return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: err.Error()}, nil
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
		return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: err.Error()}, nil
	}

	return &abciTypes.ResponseCheckTx{Code: code.Uint32()}, nil
}

func (a *AbciApp) executeTx(ctx context.Context, rawTx, blockProposerAddr []byte, logger *log.Logger) *abciTypes.ExecTxResult {
	var events []abciTypes.Event
	var gasUsed int64

	newExecuteTxRes := func(code transactions.TxCode, err error) *abciTypes.ExecTxResult {
		res := &abciTypes.ExecTxResult{
			Code:    code.Uint32(),
			GasUsed: gasUsed,
			Events:  events,
			Log:     "success",
			// Data, GasWanted, Info, Codespace
		}

		if err != nil {
			res.Log = fmt.Sprintf("FAILED TRANSACTION: %v", err) // may be too much info in err
		}

		return res
	}

	tx := &transactions.Transaction{}
	err := tx.UnmarshalBinary(rawTx)
	if err != nil {
		logger.Error("failed to unmarshal transaction", zap.Error(err))
		return newExecuteTxRes(codeEncodingError, err)
	}

	logger = logger.With(zap.String("Sender", hex.EncodeToString(tx.Sender)),
		zap.String("PayloadType", tx.Body.PayloadType.String()))

	switch tx.Body.PayloadType {
	case transactions.PayloadTypeAccountCredit:
		var creditPayload transactions.AccountCredit
		err = creditPayload.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			return newExecuteTxRes(codeEncodingError, err)
		}

		// TODO: check that tx.Sender was the proposer of the block? or just let that be in ProcessProposal...
		// if !bytes.Equal(cometAddrFromPubKey(tx.Sender), blockProposerAddr) {}

		amt, ok := big.NewInt(0).SetString(creditPayload.Amount, 10)
		if !ok {
			return newExecuteTxRes(codeUnknownError, fmt.Errorf("bad amount %q", creditPayload.Amount))
		}
		addr := a.chainEvents.Address(creditPayload.Account)
		err = a.accounts.Credit(ctx, addr, creditPayload.Account, amt)
		if err != nil { // this is probably worse then just a non-zero code (panic?)
			return newExecuteTxRes(codeUnknownError, err)
		}

		events = []abciTypes.Event{
			{
				Type: transactions.PayloadTypeAccountCredit.String(),
				Attributes: []abciTypes.EventAttribute{
					{Key: "Proposer", Value: hex.EncodeToString(tx.Sender), Index: true},
					{Key: "Result", Value: "Success", Index: true},
					{Key: "Account", Value: hex.EncodeToString(creditPayload.Account), Index: true},
					{Key: "Amount", Value: creditPayload.Amount, Index: true},
				},
			},
		}

		logger.Debug("credited account for deposit", zap.String("account", hex.EncodeToString(creditPayload.Account)),
			zap.String("amount", creditPayload.Amount))

	case transactions.PayloadTypeDeploySchema:
		var schemaPayload transactions.Schema
		err = schemaPayload.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			return newExecuteTxRes(codeEncodingError, err)
		}

		var schema *engineTypes.Schema
		schema, err = modDataset.ConvertSchemaToEngine(&schemaPayload)
		if err != nil {
			return newExecuteTxRes(codeEncodingError, err)
		}

		var res *modDataset.ExecutionResponse
		res, err = a.database.Deploy(ctx, schema, tx)
		if err != nil {
			return newExecuteTxRes(codeUnknownError, err)
		}

		dbID := utils.GenerateDBID(schema.Name, tx.Sender)
		gasUsed = res.GasUsed
		events = []abciTypes.Event{
			{
				Type: transactions.PayloadTypeDeploySchema.String(),
				Attributes: []abciTypes.EventAttribute{
					{Key: "Sender", Value: hex.EncodeToString(tx.Sender), Index: true},
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
			return newExecuteTxRes(codeEncodingError, err)
		}

		logger.Debug("drop database", zap.String("DBID", drop.DBID))

		var res *modDataset.ExecutionResponse
		res, err = a.database.Drop(ctx, drop.DBID, tx)
		if err != nil {
			return newExecuteTxRes(codeUnknownError, err)
		}

		gasUsed = res.GasUsed

	case transactions.PayloadTypeExecuteAction:
		execution := &transactions.ActionExecution{}
		err = execution.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			return newExecuteTxRes(codeEncodingError, err)
		}

		logger.Debug("execute action",
			zap.String("DBID", execution.DBID), zap.String("Action", execution.Action),
			zap.Any("Args", execution.Arguments))

		var res *modDataset.ExecutionResponse
		res, err = a.database.Execute(ctx, execution.DBID, execution.Action, convertArgs(execution.Arguments), tx)
		if err != nil {
			return newExecuteTxRes(codeUnknownError, err)
		}

		gasUsed = res.GasUsed

	case transactions.PayloadTypeValidatorJoin:
		var join transactions.ValidatorJoin
		err = join.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			return newExecuteTxRes(codeEncodingError, err)
		}

		logger.Debug("join validator",
			zap.String("pubkey", hex.EncodeToString(tx.Sender)),
			zap.Int64("power", int64(join.Power)))

		var res *modVal.ExecutionResponse
		res, err = a.validators.Join(ctx, int64(join.Power), tx)
		if err != nil {
			return newExecuteTxRes(codeUnknownError, err)
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
			return newExecuteTxRes(codeEncodingError, err)
		}

		logger.Debug("leave validator", zap.String("pubkey", hex.EncodeToString(tx.Sender)))

		var res *modVal.ExecutionResponse
		res, err = a.validators.Leave(ctx, tx)
		if err != nil {
			return newExecuteTxRes(codeUnknownError, err)
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
			return newExecuteTxRes(codeEncodingError, err)
		}

		logger.Debug("approve validator", zap.String("pubkey", hex.EncodeToString(approve.Candidate)))

		var res *modVal.ExecutionResponse
		res, err = a.validators.Approve(ctx, approve.Candidate, tx)
		if err != nil {
			return newExecuteTxRes(codeUnknownError, err)
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
			return newExecuteTxRes(codeEncodingError, err)
		}

		logger.Debug("remove validator", zap.String("pubkey", hex.EncodeToString(remove.Validator)))

		var res *modVal.ExecutionResponse
		res, err = a.validators.Remove(ctx, remove.Validator, tx)
		if err != nil {
			return newExecuteTxRes(codeUnknownError, err)
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
		return newExecuteTxRes(codeUnknownError, err)
	}

	return newExecuteTxRes(codeOk, nil)
}

// FinalizeBlock is on the consensus connection
func (a *AbciApp) FinalizeBlock(ctx context.Context, req *abciTypes.RequestFinalizeBlock) (*abciTypes.ResponseFinalizeBlock, error) {
	logger := a.log.With(zap.String("stage", "ABCI FinalizeBlock"), zap.Int("height", int(req.Height)))

	res := &abciTypes.ResponseFinalizeBlock{}

	// BeginBlock was this part

	idempotencyKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(idempotencyKey, uint64(req.Height))
	a.blockHeight = req.Height

	err := a.committer.Begin(ctx, idempotencyKey)
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
		execRes := a.executeTx(ctx, tx, req.ProposerAddress, logger)
		res.TxResults = append(res.TxResults, execRes)
	}

	// EndBlock was this part

	a.valUpdates, err = a.validators.Finalize(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to finalize validator updates: %v", err))
	}

	res.ValidatorUpdates = make([]abciTypes.ValidatorUpdate, len(a.valUpdates))
	for i, up := range a.valUpdates {
		res.ValidatorUpdates[i] = abciTypes.Ed25519ValidatorUpdate(up.PubKey, up.Power)
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

	id, err := a.committer.Commit(ctx, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("failed to commit atomic commit: %w", err)
	}

	appHash, err := a.createNewAppHash(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create new app hash: %w", err)
	}
	res.AppHash = appHash

	return res, nil
}

func (a *AbciApp) Commit(ctx context.Context, _ *abciTypes.RequestCommit) (*abciTypes.ResponseCommit, error) {
	err := a.metadataStore.IncrementBlockHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to increment block height: %w", err)
	}

	defer a.mempool.reset()

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

	height, err := a.metadataStore.GetBlockHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get block height: %w", err)
	}
	a.validators.UpdateBlockHeight(ctx, height)

	// snapshotting
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

	if err := a.validators.GenesisInit(context.Background(), vldtrs, req.InitialHeight); err != nil {
		return nil, fmt.Errorf("validators.GenesisInit failed: %w", err)
	}

	valUpdates := make([]abciTypes.ValidatorUpdate, len(vldtrs))
	for i, validator := range vldtrs {
		valUpdates[i] = abciTypes.Ed25519ValidatorUpdate(validator.PubKey, validator.Power)
	}

	err := a.metadataStore.SetAppHash(ctx, a.cfg.GenesisAppHash)
	if err != nil {
		panic(fmt.Sprintf("failed to set app hash: %v", err))
	}

	logger.Debug("initialized chain", zap.String("app hash", fmt.Sprintf("%x", a.cfg.GenesisAppHash)))

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
	a.log.Info("extend vote", zap.Int64("height", req.Height))

	dv := &TestVoteExt{
		Msg: "howdy",
	}
	data, err := dv.MarshalBinary()
	if err != nil {
		return nil, err
	}
	ve := []*VoteExtensionSegment{{
		Version: 0,
		Type:    VoteExtensionTypeDeposit,
		Data:    data,
	}}

	// Fetch the unannounced deposits from the DepositStore and encode them into the vote extension.
	depEventIDs, depAmts, depAccts, err := a.chainEvents.DepositsToReport(ctx, a.cfg.NodeAddress)
	if err != nil {
		// TODO: Is this a panic? or just return nil extension?
		return &abciTypes.ResponseExtendVote{
			VoteExtension: nil,
		}, nil
	}

	for i := range depEventIDs {
		// EncodeDeposits(...)
		dep := &DepositVoteExt{
			EventID: depEventIDs[i],
			Account: depAccts[i],
			Amount:  depAmts[i],
		}
		ve = append(ve, dep.EncodeSegment())
	}

	return &abciTypes.ResponseExtendVote{
		VoteExtension: EncodeVoteExtension(ve),
	}, nil
}

// registerVoteExtension is the handler for incoming extensions
func (a *AbciApp) registerVoteExtension(ctx context.Context, validator, b []byte) error {
	ve, err := DecodeVoteExtension(b)
	if err != nil {
		return fmt.Errorf("failed to decode vote extension: %w", err)
	}
	if len(ve) == 0 {
		return nil // ok
	}

	a.log.Info("decoded vote extension", zap.Int("segments", len(ve)))

	for i := range ve {
		// Decode the segment and invoke the appropriate module method with the
		// data. This code could submit the opaque extension data, but the
		// module interfaces are more sensible and clear this way.
		switch ve[i].Type {
		case VoteExtensionTypeDeposit:
			dep := &DepositVoteExt{}
			err = dep.UnmarshalBinary(ve[i].Data)
			if err != nil {
				// a.log.Warn("failed to decode deposit vote", zap.Error(err))
				return fmt.Errorf("failed to decode deposit vote: %w", err)
			}
			a.log.Info("decoded a val deposit vote extension segment",
				zap.String("account", dep.Account), zap.String("amount", dep.Amount))

			amt := big.NewInt(0)
			amt.SetString(dep.Amount, 10)

			// Record the deposit events from other validators in the DepositStore.
			a.chainEvents.AddDeposit(ctx, dep.EventID, dep.Account, amt, hex.EncodeToString(validator))

		default:
			a.log.Error("ignoring unknown vote extension type", zap.Uint32("type", ve[i].Type))
			// just ignore it, return ACCEPT
			continue
		}
	}

	return nil
}

// VerifyVoteExtension verifies the vote extension data from another validator.
func (a *AbciApp) VerifyVoteExtension(ctx context.Context, req *abciTypes.RequestVerifyVoteExtension) (*abciTypes.ResponseVerifyVoteExtension, error) {
	a.log.Info("verify vote extension", zap.Int64("height", req.Height))

	// handle the vote extension by dispatching to appropriate modules depending on type
	err := a.registerVoteExtension(ctx, req.ValidatorAddress, req.VoteExtension)
	if err != nil {
		a.log.Warn("bad vote extension", zap.Error(err))
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

func (a *AbciApp) PrepareProposal(ctx context.Context, req *abciTypes.RequestPrepareProposal) (*abciTypes.ResponsePrepareProposal, error) {
	logger := a.log.With(zap.String("stage", "ABCI PrepareProposal"),
		zap.Int64("height", req.Height),
		zap.Int("txs", len(req.Txs)))

	okTxns := prepareMempoolTxns(req.Txs, int(req.MaxTxBytes), &a.log)

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

	/*
		for _, vote := range req.LocalLastCommit.Votes {
			a.log.Info("ExtendedCommitInfo", zap.String("validator", hex.EncodeToString(vote.Validator.Address)),
				zap.Int("extension_length", len(vote.VoteExtension)))
			err := a.registerVoteExtension(ctx, vote.Validator.Address, vote.VoteExtension)
			if err != nil {
				a.log.Error("bad vote extension", zap.Error(err))
				continue // just ignore it
			}
		}
	*/

	// For any deposit events that have reached the deposit threshold, create an
	// account credit transaction (and clear those events or wait until the tx
	// executes in case the proposal is rejected?).

	for eventID, dep := range a.thresholdDeposits(ctx) {
		a.log.Debug("preparing account credit txn", zap.String("acct_addr", dep.acct),
			zap.String("amount", dep.amt.String()), zap.String("event", eventID))
		// newAccountCreditTxn(eventID, dep.Account, dep.Amount)
	}

	// Create consensus transaction type and include in okTxns (prepend?)

	return &abciTypes.ResponsePrepareProposal{
		Txs: okTxns,
	}, nil
}

func haveValidatorKey(b []byte, vs []*validators.Validator) bool {
	for _, vi := range vs {
		if bytes.Equal(b, vi.PubKey) {
			return true
		}
	}
	return false
}

type depEvt struct {
	acct string // account address, not pubkey
	amt  *big.Int
}

// thresholdDeposits returns the deposit events that have reached the threshold
// of attestations from current validators.
func (a *AbciApp) thresholdDeposits(ctx context.Context) map[string]depEvt {
	deps := make(map[string]depEvt)

	vals := make(map[string]bool)
	for addr, _ := range a.valAddrToKey {
		vals[addr] = true
	}

	deposits, err := a.chainEvents.DepositEvents(ctx, vals)
	if err != nil {
		// TODO: How to handle the error?
		return nil
	}

	for eventID, dep := range deposits {
		amt := big.NewInt(0)
		amt.SetString(dep.Amount, 10)
		deps[eventID] = depEvt{dep.Sender, amt}
		// attestations := dep.Attestations
		// thresh := validators.Threshold(len(currentValidators))
		// if len(attestations) < thresh {
		// 	a.log.Info("not enough attestations", zap.Int("attestations", len(attestations)))
		// 	continue
		// }
		// var count int
		// for _, ai := range attestations {
		// 	pubkey, ok := a.valAddrToKey[string(ai)] // attestations are address, get pubkey
		// 	if !ok {
		// 		continue // not a current validator or address unknown
		// 	}
		// 	if haveValidatorKey(pubkey, currentValidators) {
		// 		count++
		// 	}
		// }

		// fmt.Println(eventID)
		// // TODO: Check the logic for this
		// amt, ok := big.NewInt(0).SetString(dep.Amount, 10)
		// if !ok { // that shouldn't have happened
		// 	a.log.Error("invalid deposit amount from event store!", zap.String("amount", dep.Amount))
		// 	continue
		// }

		// if count >= thresh {
		// 	deps[eventID] = depEvt{dep.Account, amt}
		// }
	}
	return deps
}

func (a *AbciApp) validateProposalTransactions(ctx context.Context, txns [][]byte, proposerAddr string) error {
	logger := a.log.With(zap.String("stage", "ABCI ProcessProposal"))
	grouped, err := groupTxsBySender(txns)
	if err != nil {
		return fmt.Errorf("failed to group transaction by sender: %w", err)
	}

	// Deposit events at threshold (allowable ot have account credit txn).
	var thresholdDeps map[string]depEvt

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

			// If account credit, ensure there is a threshold of attestations
			// warranting the transaction, and that it is authored by the block
			// proposer.
			if tx.Body.PayloadType == transactions.PayloadTypeAccountCredit {
				if cometAddrFromPubKey(tx.Sender) != proposerAddr {
					return fmt.Errorf("account credit not authored by block proposer")
				}
				if thresholdDeps == nil {
					thresholdDeps = a.thresholdDeposits(ctx)
				}
				eventID, acct, amt, err := extractDepInfo(tx.Body)
				if err != nil {
					return fmt.Errorf("invalid account credit tx payload: %w", err)
				}
				if dep, ok := thresholdDeps[eventID]; !ok {
					return fmt.Errorf("deposit not at threshold (event ID %v)", eventID)
				} else if a.chainEvents.Address(acct) != dep.acct {
					return fmt.Errorf("deposit account mismatch (%x != %x)", acct, dep.acct)
				} else if dep.amt.Cmp(amt) != 0 {
					return fmt.Errorf("deposit amount mismatch (%v != %v)", acct, dep.amt)
				}
				// it's valid!
			}
		}
	}
	return nil
}

func extractDepInfo(txBody *transactions.TransactionBody) (id string, acct []byte, amt *big.Int, err error) {
	if txBody.PayloadType != transactions.PayloadTypeAccountCredit {
		err = errors.New("not an account credit transaction")
		return
	}
	var creditPayload transactions.AccountCredit
	err = creditPayload.UnmarshalBinary(txBody.Payload)
	if err != nil {
		return
	}
	var ok bool
	amt, ok = big.NewInt(0).SetString(creditPayload.Amount, 10)
	if !ok {
		err = fmt.Errorf("bad amount %q", creditPayload.Amount)
		return
	}
	return creditPayload.DepositEventID, creditPayload.Account, amt, nil
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

	if err := a.validateProposalTransactions(ctx, req.Txs, string(req.ProposerAddress)); err != nil {
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
