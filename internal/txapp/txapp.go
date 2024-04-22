// package tx_router routes transactions to the appropriate module(s)
package txapp

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"sync"

	"go.uber.org/zap"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/serialize"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/voting"
)

// NewTxApp creates a new router.
func NewTxApp(db DB, engine common.Engine, signer *auth.Ed25519Signer,
	events Rebroadcaster, chainID string, GasEnabled bool, extensionConfigs map[string]map[string]string, log log.Logger) (*TxApp, error) {
	voteBodyTxSize, err := computeEmptyVoteBodyTxSize(chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to compute empty vote body tx size: %w", err)
	}

	resTypes := resolutions.ListResolutions()
	slices.Sort(resTypes)

	t := &TxApp{
		Database: db,
		Engine:   engine,
		events:   events,
		log:      log,
		mempool: &mempool{
			accounts:   make(map[string]*types.Account),
			gasEnabled: GasEnabled,
			nodeAddr:   signer.Identity(),
		},
		signer:              signer,
		chainID:             chainID,
		GasEnabled:          GasEnabled,
		extensionConfigs:    extensionConfigs,
		emptyVoteBodyTxSize: voteBodyTxSize,
		resTypes:            resTypes,
	}
	t.height, t.appHash, err = t.ChainInfo(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain info: %w", err)
	}
	return t, nil
}

// TxApp maintains the state for Kwil's ABCI application.
// It is responsible for interpreting payload bodies and routing them properly,
// maintaining a mempool for uncommitted accounts, pricing transactions,
// managing atomicity of the database, and managing the validator set.
type TxApp struct {
	Database DB            // postgres database
	Engine   common.Engine // tracks deployed schemas
	// The various internal stores (accounts, votes, etc.) are accessed through
	// the Database via the functions defined in relevant packages.

	GasEnabled bool
	events     Rebroadcaster

	chainID string
	signer  *auth.Ed25519Signer

	log log.Logger

	mempool *mempool

	// appHash is the last block's apphash, set for genesis in GenesisInit
	// updated in FinalizeBlock by combining with new engine hash.
	appHash []byte
	height  int64

	validators []*types.Validator // used to optimize reads, gets updated at the block boundaries
	valMtx     sync.RWMutex       // protects validators access
	valChans   []chan []*types.Validator

	// transaction that exists between Begin and Commit
	currentTx sql.OuterTx

	// Abci.InitChain can be called multiple times from comet when the node fails
	// before the first block is committed.
	// Therefore any changes in the Genesis must be committed only
	// upon calling Commit at the end of the first block.
	// For more information, see: https://github.com/cometbft/cometbft/issues/203
	// genesisTx is the transaction that is used to apply the genesis state changes
	// along with the updates by the transactions in the first block.
	genesisTx sql.OuterTx

	extensionConfigs map[string]map[string]string

	// precomputed variables
	emptyVoteBodyTxSize int64
	resTypes            []string
}

func (r *TxApp) Log() *log.Logger {
	return &r.log
}

// GenesisInit initializes the TxApp. It must be called outside of a session,
// and before any session is started.
// It can assign the initial validator set and initial account balances.
// It is only called once for a new chain.
func (r *TxApp) GenesisInit(ctx context.Context, validators []*types.Validator, genesisAccounts []*types.Account,
	initialHeight int64, genesisAppHash []byte) error {
	tx, err := r.Database.BeginOuterTx(ctx)
	if err != nil {
		return err
	}
	r.genesisTx = tx

	// With the genesisTx not being committed until the first FinalizeBlock, we
	// expect no existing chain state in the application (postgres).
	height, appHash, err := getChainState(ctx, tx)
	if err != nil {
		err2 := tx.Rollback(ctx)
		if err2 != nil {
			return fmt.Errorf("error rolling back transaction: %s, error: %s", err.Error(), err2.Error())
		}
		return fmt.Errorf("error getting database height: %s", err.Error())
	}

	// First app hash and height are stored in FinalizeBlock for first block.
	if height != -1 {
		return fmt.Errorf("expected database to be uninitialized, but had height %d", height)
	}
	if len(appHash) != 0 {
		return fmt.Errorf("expected NULL app hash, got %x", appHash)
	}

	r.appHash = genesisAppHash // combined with first block's apphash and stored in FinalizeBlock

	// Add Genesis Validators
	var voters []*types.Validator
	r.valMtx.Lock()
	defer r.valMtx.Unlock()

	for _, validator := range validators {
		err := setVoterPower(ctx, tx, validator.PubKey, validator.Power)
		if err != nil {
			err2 := tx.Rollback(ctx)
			if err2 != nil {
				return fmt.Errorf("error rolling back transaction: %s, error: %s", err.Error(), err2.Error())
			}
			return err
		}
		voters = append(voters, validator)
	}
	r.validators = voters

	// Fund Genesis Accounts
	for _, account := range genesisAccounts {
		err := credit(ctx, tx, account.Identifier, account.Balance)
		if err != nil {
			err2 := tx.Rollback(ctx)
			if err2 != nil {
				return fmt.Errorf("error rolling back transaction: %s, error: %s", err.Error(), err2.Error())
			}
			return err
		}
	}
	return nil
}

// ChainInfo is called be the ABCI applications' Info method, which is called
// once on startup except when the node is at genesis, in which case GenesisInit
// is called by the application's ChainInit method.
func (r *TxApp) ChainInfo(ctx context.Context) (int64, []byte, error) {
	tx, err := r.Database.BeginReadTx(ctx)
	if err != nil {
		return 0, nil, err
	}
	defer tx.Rollback(ctx)

	// MAYBE: return r.height, r.appHash from the exported method and only query
	// the DB from an unexported method that c'tor uses. Needs mutex. Hitting DB
	// always may also be good to ensure the exported method gets committed.

	// return getChainState(ctx, tx)
	height, appHash, err := getChainState(ctx, tx)
	if err != nil {
		return 0, nil, err
	}
	// r.log.Debug("ChainInfo", log.Int("height", height), log.String("appHash", hex.EncodeToString(appHash)),
	// 	log.Int("height_x", r.height), log.String("appHash_x", hex.EncodeToString(r.appHash)))
	if height == -1 {
		height = 0 // for ChainInfo caller, non-negative is expected for genesis
	}
	return height, appHash, nil
}

// UpdateValidator updates a validator's power.
// It can only be called in between Begin and Finalize.
// The value passed as power will simply replace the current power.
func (r *TxApp) UpdateValidator(ctx context.Context, validator []byte, power int64) error {
	if r.currentTx == nil {
		return errors.New("txapp misuse: cannot update a validator without a transaction in progress")
	}

	// since validators are used for voting, we also must update the vote store. this should be atomic.

	sp, err := r.currentTx.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer sp.Rollback(ctx)

	err = setVoterPower(ctx, r.currentTx, validator, power)
	if err != nil {
		return err
	}

	return sp.Commit(ctx)
}

// SubscribeValidators creates and returns a new channel on which the current
// validator set will be sent for each block Commit. The receiver will miss
// updates if they are unable to receive fast enough. This should generally
// be used after catch-up is complete, and only called once by the receiving
// goroutine rather than repeatedly in a loop, for instance. The slice should
// not be modified by the receiver.
func (r *TxApp) SubscribeValidators() <-chan []*types.Validator {
	// There's only supposed to be one user of this method, and they should
	// only get one channel and listen, but play it safe and use a slice.
	r.valMtx.Lock()
	defer r.valMtx.Unlock()
	c := make(chan []*types.Validator, 1)
	r.valChans = append(r.valChans, c)
	return c
}

// announceValidators sends the current validator list to subscribers from
// ReceiveValidators.
func (r *TxApp) announceValidators() {
	// dev note: this method should not be blocked by receivers. Keep a default
	// case and create buffered channels.
	r.valMtx.RLock()
	defer r.valMtx.RUnlock()

	if len(r.valChans) == 0 {
		return // no subscribers, skip the slice clone
	}

	vals := slices.Clone(r.validators)

	for _, c := range r.valChans {
		select {
		case c <- vals:
		default: // they'll get the next one... this is just supposed to be better than polling
			r.log.Warn("Validator update channel is blocking")
		}
	}
}

// GetValidators returns a shallow copy of the current validator set.
// It will not return uncommitted changes.
func (r *TxApp) GetValidators(ctx context.Context) ([]*types.Validator, error) {
	r.valMtx.Lock()
	defer r.valMtx.Unlock()

	// if we have a cached validator set, return it
	if r.validators != nil {
		return slices.Clone(r.validators), nil
	}

	// otherwise, we need to get the validator set from the database
	// This is done especially when a node restarts
	readTx, err := r.Database.BeginReadTx(ctx)
	if err != nil {
		return nil, err
	}
	defer readTx.Rollback(ctx) // always rollback read tx
	// NOTE: we aren't saving this to r.validators, leaving that to next FinalizeBlock. We could though...
	return getAllVoters(ctx, readTx)
}

// GetValidatorsPower returns the total power of the current validator set.
func (r *TxApp) GetValidatorSetPower(ctx context.Context) (int64, error) {
	validators, err := r.GetValidators(ctx)
	if err != nil {
		return 0, err
	}

	var totalPower int64
	for _, v := range validators {
		totalPower += v.Power
	}

	return totalPower, nil
}

// Execute executes a transaction.  It will route the transaction to the
// appropriate module(s) and return the response.
func (r *TxApp) Execute(ctx TxContext, tx *transactions.Transaction) *TxResponse {
	route, ok := routes[tx.Body.PayloadType.String()]
	if !ok {
		return txRes(nil, transactions.CodeInvalidTxType, fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String()))
	}

	r.log.Debug("executing transaction", zap.Any("tx", tx))

	if r.currentTx == nil {
		return txRes(nil, transactions.CodeUnknownError, errors.New("txapp misuse: cannot execute a blockchain transaction without a db transaction in progress"))
	}

	res := route.Execute(ctx, r, tx)
	if res.Error != nil {
		return res
	}

	return res
}

// Begin signals that a new block has begun.
// It contains information on any validators whose power should be updated.
func (r *TxApp) Begin(ctx context.Context) error {
	if r.currentTx != nil {
		return errors.New("txapp misuse: cannot begin a new block while a transaction is in progress")
	}

	if r.genesisTx != nil {
		r.currentTx = r.genesisTx
		return nil
	}

	tx, err := r.Database.BeginOuterTx(ctx)
	if err != nil {
		return err
	}

	r.currentTx = tx

	return nil
}

// Finalize signals that a block has been finalized. No more changes can be
// applied to the database. It returns the apphash and the validator set.
func (r *TxApp) Finalize(ctx context.Context, blockHeight int64) (appHash []byte, finalValidators []*types.Validator, err error) {
	if r.currentTx == nil {
		return nil, nil, errors.New("txapp misuse: cannot finalize a block without a transaction in progress")
	}

	defer func() {
		if err != nil {
			err2 := r.currentTx.Rollback(ctx)
			r.currentTx = nil
			if err2 != nil {
				err = fmt.Errorf("error rolling back transaction: %s, error: %s", err.Error(), err2.Error())
			}

			return
		}
	}()

	r.log.Debug("Finalize(start)", log.Int("height", r.height), log.String("appHash", hex.EncodeToString(r.appHash)))

	if blockHeight != r.height+1 {
		return nil, nil, fmt.Errorf("Finalize was expecting height %d, got %d", r.height+1, blockHeight)
	}

	err = r.processVotes(ctx, blockHeight)
	if err != nil {
		return nil, nil, err
	}

	finalValidators, err = getAllVoters(ctx, r.currentTx)
	if err != nil {
		return nil, nil, err
	}

	// While still in the DB transaction, update to this next height but dummy
	// app hash. If we crash before Commit can store the next app hash that we
	// get after Precommit, the startup handshake's call to Info will detect the
	// mismatch. That requires manual recovery (drop state and reapply), but it
	// at least detects this recorded height rather than not recognizing that we
	// have committed the data for this block at all.
	err = setChainState(ctx, r.currentTx, blockHeight, []byte{0x42})
	if err != nil {
		return nil, nil, err
	}

	engineHash, err := r.currentTx.Precommit(ctx)
	if err != nil {
		return nil, nil, err
	}

	r.appHash = crypto.Sha256(append(r.appHash, engineHash...))
	r.height = blockHeight

	r.valMtx.Lock()
	r.validators = finalValidators
	r.valMtx.Unlock()

	r.log.Debug("Finalize(end)", log.Int("height", r.height), log.String("appHash", hex.EncodeToString(r.appHash)))

	// I'd really like to setChainState here with appHash, but we can't use
	// currentTx for anything now except Commit.

	return r.appHash, finalValidators, nil
}

// processVotes confirms resolutions that have been approved by the network,
// expires resolutions that have expired, and properly credits proposers and voters.
func (r *TxApp) processVotes(ctx context.Context, blockheight int64) error {
	credits := make(creditMap)

	var finalizedIds []types.UUID
	// markedProcessedIds is a separate list for marking processed, since we do not want to process validator resolutions
	// validator vote IDs are not unique, so we cannot mark them as processed, in case a validator leaves and joins again
	var markProcessedIds []types.UUID
	// resolveFuncs tracks the resolve function for each resolution, in the order they are queried.
	// we track this and execute all of these functions after we have found all confirmed resolutions
	// because a resolve function can change a validator's power. This would then change the required power
	// for subsequent resolutions in the same block, which should not happen.
	var resolveFuncs []*struct {
		Resolution  *resolutions.Resolution
		ResolveFunc func(ctx context.Context, app *common.App, resolution *resolutions.Resolution) error
	}

	totalPower, err := r.GetValidatorSetPower(ctx)
	if err != nil {
		return err
	}

	for _, resolutionType := range r.resTypes {
		cfg, err := resolutions.GetResolution(resolutionType)
		if err != nil {
			return err
		}

		finalized, err := getResolutionsByThresholdAndType(ctx, r.currentTx, cfg.ConfirmationThreshold, resolutionType, totalPower)
		if err != nil {
			return err
		}

		for _, resolution := range finalized {
			credits.applyResolution(resolution)
			finalizedIds = append(finalizedIds, resolution.ID)

			// we do not want to mark processed for validator join and remove events, as they can occur again
			if resolution.Type != voting.ValidatorJoinEventType && resolution.Type != voting.ValidatorRemoveEventType {
				markProcessedIds = append(markProcessedIds, resolution.ID)
			}

			resolveFuncs = append(resolveFuncs, &struct {
				Resolution  *resolutions.Resolution
				ResolveFunc func(ctx context.Context, app *common.App, resolution *resolutions.Resolution) error
			}{
				Resolution:  resolution,
				ResolveFunc: cfg.ResolveFunc,
			})
		}
	}

	// apply all resolutions
	for _, resolveFunc := range resolveFuncs {
		r.log.Debug("resolving resolution", zap.String("type", resolveFunc.Resolution.Type), zap.String("id", resolveFunc.Resolution.ID.String()))

		tx, err := r.currentTx.BeginTx(ctx)
		if err != nil {
			return err
		}

		err = resolveFunc.ResolveFunc(ctx, &common.App{
			Service: &common.Service{
				Logger:           r.log.Named("resolution_" + resolveFunc.Resolution.Type).Sugar(),
				ExtensionConfigs: r.extensionConfigs,
			},
			DB:     tx,
			Engine: r.Engine,
		}, resolveFunc.Resolution)
		if err != nil {
			err2 := tx.Rollback(ctx)
			if err2 != nil {
				return fmt.Errorf("error rolling back transaction: %s, error: %s", err.Error(), err2.Error())
			}
			return err
		}

		err = tx.Commit(ctx)
		if err != nil {
			return err
		}
	}

	// now we will expire resolutions
	expired, err := getExpired(ctx, r.currentTx, blockheight)
	if err != nil {
		return err
	}

	expiredIds := make([]types.UUID, 0, len(expired))
	requiredPowerMap := make(map[string]int64) // map of resolution type to required power
	for _, resolution := range expired {
		expiredIds = append(expiredIds, resolution.ID)
		if resolution.Type != voting.ValidatorJoinEventType && resolution.Type != voting.ValidatorRemoveEventType {
			markProcessedIds = append(markProcessedIds, resolution.ID)
		}

		threshold, ok := requiredPowerMap[resolution.Type]
		if !ok {
			cfg, err := resolutions.GetResolution(resolution.Type)
			if err != nil {
				return err
			}

			// we need to use each configured resolutions refund threshold
			requiredPowerMap[resolution.Type] = requiredPower(ctx, r.currentTx, cfg.RefundThreshold, totalPower)
		}
		refunded := false
		// if it has enough power, we will still refund
		if resolution.ApprovedPower >= threshold {
			refunded = true
			credits.applyResolution(resolution)
		}

		r.log.Debug("expiring resolution", zap.String("type", resolution.Type), zap.String("id", resolution.ID.String()), zap.Bool("refunded", refunded))
	}

	allIds := append(finalizedIds, expiredIds...)
	err = deleteResolutions(ctx, r.currentTx, allIds...)
	if err != nil {
		return err
	}

	err = markProcessed(ctx, r.currentTx, markProcessedIds...)
	if err != nil {
		return err
	}

	// This is to ensure that the nodes that never get to vote on this event due to limitation
	// per block vote sizes, they never get to vote and essentially delete the event
	// So this is handled instead when the nodes are approved.
	// TODO: We need to figure out the consequences of resolutions getting expired due to the vote limits set per block. There can be scenarios where the events are observed by the nodes, but before they can vote, the event gets expired. rare but possible in the situations with higher event traffic.
	err = deleteEvents(ctx, r.currentTx, markProcessedIds...)
	if err != nil {
		return err
	}

	// now we will apply credits if gas is enabled
	if r.GasEnabled {
		for pubKey, amount := range credits {
			err = credit(ctx, r.currentTx, []byte(pubKey), amount)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

var (
	ValidatorVoteBodyBytePrice int64 = 1000                  // Per byte cost
	ValidatorVoteIDPrice             = big.NewInt(1000 * 16) // 16 bytes for the UUID
)

// creditMap maps string(public_keys) to big.Int amounts that should be credited
type creditMap map[string]*big.Int

// applyResolution will calculate the rewards for the proposer and voters of a resolution.
// it will add the rewards to the credit map.
func (c creditMap) applyResolution(res *resolutions.Resolution) {
	// reward voters.
	// this will include the proposer, even if they did not submit a vote id
	for _, voter := range res.Voters {
		// if the voter is the proposer, then we will reward them below,
		// since extra logic is required if they did not submit a vote id
		if bytes.Equal(voter.PubKey, res.Proposer) {
			continue
		}

		currentBalance, ok := c[string(voter.PubKey)]
		if !ok {
			currentBalance = big.NewInt(0)
		}

		c[string(voter.PubKey)] = big.NewInt(0).Add(currentBalance, ValidatorVoteIDPrice)
	}

	bodyCost := big.NewInt(ValidatorVoteBodyBytePrice * int64(len(res.Body)))
	currentBalance, ok := c[string(res.Proposer)]
	if !ok {
		currentBalance = big.NewInt(0)
	}

	// reward proposer
	c[string(res.Proposer)] = big.NewInt(0).Add(currentBalance, bodyCost)
}

// Commit signals that a block's state changes should be committed.
func (r *TxApp) Commit(ctx context.Context) error {
	if r.currentTx == nil {
		return errors.New("txapp misuse: cannot commit a block without a transaction in progress")
	}
	defer r.mempool.reset()

	// r.log.Debug("Commit(start)", log.Int("height", r.height), log.String("appHash", hex.EncodeToString(r.appHash)))

	err := r.currentTx.Commit(ctx)
	if err != nil {
		return err
	}

	r.currentTx = nil
	r.genesisTx = nil

	// Now we can store the app hash computed in FinalizeBlock after Precommit.
	// Note that if we crash here, Info on startup will immediately detect an
	// unexpected app hash since we've saved this height in the Commit above.
	// While this takes manual recovery, it does not go undetected as if we had
	// not updated to the new height in that Commit. We could improve this with
	// some refactoring to pg.DB to allow multiple simultaneous uncommitted
	// prepared transactions to make this an actual two-phase commit, but it is
	// just a single row so the difference is minor.
	ctx = context.Background() // don't let them cancel us now, we need consistency with consensus tx commit
	tx, err := r.Database.BeginOuterTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = setChainState(ctx, tx, r.height, r.appHash) // unchanged height, known appHash
	if err != nil {
		return err
	}

	err = tx.Commit(ctx) // no Precommit for this one
	if err != nil {
		return err
	}

	r.announceValidators()

	return nil
}

// ApplyMempool applies the transactions in the mempool.
// If it returns an error, then the transaction is invalid.
func (r *TxApp) ApplyMempool(ctx context.Context, tx *transactions.Transaction) error {
	// check that payload type is valid
	_, ok := routes[tx.Body.PayloadType.String()]
	if !ok {
		return fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String())
	}
	readTx, err := r.Database.BeginReadTx(ctx)
	if err != nil {
		return err
	}
	defer readTx.Rollback(ctx) // always rollback read tx

	return r.mempool.applyTransaction(ctx, tx, readTx, r.events)
}

// AccountInfo gets account info from either the mempool or the account store.
// It takes a flag to indicate whether it should check the mempool first.
func (r *TxApp) AccountInfo(ctx context.Context, acctID []byte, getUncommitted bool) (balance *big.Int, nonce int64, err error) {
	readTx, err := r.Database.BeginReadTx(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer readTx.Rollback(ctx) // always rollback read tx

	var a *types.Account
	if getUncommitted {
		a, err = r.mempool.accountInfoSafe(ctx, readTx, acctID)
	} else {
		a, err = getAccount(ctx, readTx, acctID)
	}
	if err != nil {
		return nil, 0, err
	}

	return a.Balance, a.Nonce, nil
}

// ProposerTxs returns the transactions that the proposer should include in the
// next block.
// It takes txNonce as an argument because, the proposer may have its own transactions
// in the mempool that are included in the current block. Therefore, we need to know the
// largest nonce of the transactions included in the block that are authored by the proposer.
// This transaction only includes voteBodies for events whose bodies have not been received by the network.
// Therefore, there won't be more than 1 VoteBody transaction per event.
func (r *TxApp) ProposerTxs(ctx context.Context, txNonce uint64, maxTxsSize int64, proposerAddr []byte) ([][]byte, error) {
	bal, nonce, err := r.AccountInfo(ctx, proposerAddr, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposer account: %w", err)
	}

	if r.GasEnabled && nonce == 0 && bal.Sign() == 0 {
		r.log.Debug("proposer account has no balance, not allowed to propose any new transactions")
		return nil, nil
	}

	if txNonce == 0 {
		txNonce = uint64(nonce) + 1
	}

	readTx, err := r.Database.BeginReadTx(ctx)
	if err != nil {
		return nil, err
	}
	defer readTx.Rollback(ctx) // always rollback read tx

	maxTxsSize -= r.emptyVoteBodyTxSize + 1000 // empty payload size + 1000 safety buffer
	events, err := getEvents(ctx, readTx)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}

	ids := make([]types.UUID, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID())
	}

	// Is thre any reason to check for notProcessed events here? Becase event store will never have events that are already processed.

	// Limit upto only 50 VoteBodies per block
	if len(ids) > 50 {
		ids = ids[:50]
	}

	eventMap := make(map[types.UUID]*types.VotableEvent)
	for _, evt := range events {
		eventMap[evt.ID()] = evt
	}

	finalEvents := make([]*transactions.VotableEvent, 0)
	for _, id := range ids {
		event, ok := eventMap[id]
		if !ok {
			// this should never happen
			return nil, fmt.Errorf("internal bug: event with id %s not found", id.String())
		}

		evtSz := int64(len(event.Type)) + int64(len(event.Body)) + eventRLPSize
		if evtSz > maxTxsSize {
			break
		}
		maxTxsSize -= evtSz
		finalEvents = append(finalEvents, &transactions.VotableEvent{
			Type: event.Type,
			Body: event.Body,
		})
	}

	if len(finalEvents) == 0 {
		return nil, nil
	}

	r.log.Info("Creating new ValidatorVoteBodies transaction", log.Int("events", len(finalEvents)))

	tx, err := transactions.CreateTransaction(&transactions.ValidatorVoteBodies{
		Events: finalEvents,
	}, r.chainID, txNonce)
	if err != nil {
		return nil, err
	}

	// Fee Estimate
	amt, err := r.Price(ctx, tx)
	if err != nil {
		return nil, err
	}
	tx.Body.Fee = amt

	err = tx.Sign(r.signer)
	if err != nil {
		return nil, err
	}

	bts, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return [][]byte{bts}, nil // NOTE: may return more than one in the future.
}

/*
	In the two functions below, I am calculating the constants for overheads on RLP encoding.
	The overhead is simply the amount of extra bytes added to an event's size when it is RLP encoded.
	The first function calculates the overhead per event, while the second calculates the overhead
	of encoding a slice of events.
*/

// eventRLPSize is the extra size added to an event from RLP
// encoding. It is the same regardless of event data.
var eventRLPSize int64 = func() int64 {
	event := &types.VotableEvent{
		Body: []byte("body"),
		Type: "type",
	}

	eventSize := int64(len(event.Type)) + int64(len(event.Body))

	bts, err := serialize.Encode(event)
	if err != nil {
		panic(err)
	}

	return int64(len(bts)) - eventSize
}()

// TxResponse is the response from a transaction.
// It contains information about the transaction, such as the amount spent.
type TxResponse struct {
	// ResponseCode is the response code from the transaction
	ResponseCode transactions.TxCode

	// Spend is the amount of tokens spent by the transaction
	Spend int64

	// Error is the error returned by the transaction, if any
	Error error
}

// Price estimates the price of a transaction.
// It returns the estimated price in tokens.
func (r *TxApp) Price(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	if !r.GasEnabled {
		return big.NewInt(0), nil
	}

	route, ok := routes[tx.Body.PayloadType.String()]
	if !ok {
		return nil, fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String())
	}

	return route.Price(ctx, r, tx)
}

// checkAndSpend checks the price of a transaction.
// It requires a tx, so that spends can be made transactional with other database interactions.
// it returns the price it will cost to execute the transaction.
// if the transaction does not have enough tokens to pay for the transaction,
// it will return an error.
// if the caller does not have enough tokens to pay for the transaction,
// it will return an error.
// if the transaction does not have the correct nonce, it will return an error.
// it will spend the tokens if the caller has enough tokens.
// It also returns an error code.
// if we allow users to implement their own routes, this function will need to
// be exported.
func (r *TxApp) checkAndSpend(ctx TxContext, tx *transactions.Transaction, pricer Pricer, dbTx sql.Executor) (*big.Int, transactions.TxCode, error) {
	amt := big.NewInt(0)
	var err error

	if r.GasEnabled {
		amt, err = pricer.Price(ctx.Ctx, r, tx)
		if err != nil {
			return nil, transactions.CodeUnknownError, err
		}
	}

	// check if the transaction consented to spending enough tokens
	if tx.Body.Fee.Cmp(amt) < 0 {
		// If the transaction does not consent to spending required tokens for the transaction execution,
		// spend the approved tx fee and terminate the transaction
		err = spend(ctx.Ctx, dbTx, tx.Sender, tx.Body.Fee, int64(tx.Body.Nonce))
		if errors.Is(err, accounts.ErrInsufficientFunds) {
			// spend as much as possible
			account, err := getAccount(ctx.Ctx, dbTx, tx.Sender)
			if err != nil { // account will just be empty if not found
				return nil, transactions.CodeUnknownError, err
			}

			err2 := spend(ctx.Ctx, dbTx, tx.Sender, account.Balance, int64(tx.Body.Nonce))
			if err2 != nil {
				if errors.Is(err2, accounts.ErrAccountNotFound) {
					return nil, transactions.CodeInsufficientBalance, errors.New("account has zero balance")
				}
				return nil, transactions.CodeUnknownError, err2
			}

			return nil, transactions.CodeInsufficientBalance, fmt.Errorf("transaction tries to spend %s tokens, but account only has %s tokens", amt.String(), tx.Body.Fee.String())
		}
		if err != nil {
			if errors.Is(err, accounts.ErrAccountNotFound) {
				return nil, transactions.CodeInsufficientBalance, errors.New("account has zero balance")
			}
			return nil, transactions.CodeUnknownError, err
		}

		return nil, transactions.CodeInsufficientFee, fmt.Errorf("transaction does not consent to spending enough tokens. transaction fee: %s, required fee: %s", tx.Body.Fee.String(), amt.String())
	}

	// spend the tokens
	err = spend(ctx.Ctx, dbTx, tx.Sender, amt, int64(tx.Body.Nonce))
	if errors.Is(err, accounts.ErrInsufficientFunds) {
		// spend as much as possible
		account, err := getAccount(ctx.Ctx, dbTx, tx.Sender)
		if err != nil {
			return nil, transactions.CodeUnknownError, err
		}

		err2 := spend(ctx.Ctx, dbTx, tx.Sender, account.Balance, int64(tx.Body.Nonce))
		if err2 != nil {
			return nil, transactions.CodeUnknownError, err2
		}
		return nil, transactions.CodeInsufficientBalance, fmt.Errorf("transaction tries to spend %s tokens, but account has %s tokens", amt.String(), account.Balance.String())
	}
	if err != nil {
		if errors.Is(err, accounts.ErrAccountNotFound) { // probably wouldn't have passed the fee check
			return nil, transactions.CodeInsufficientBalance, errors.New("account has zero balance")
		}
		return nil, transactions.CodeUnknownError, err
	}

	return amt, transactions.CodeOk, nil
}

// txRes wraps a spend, tx code, and error into a tx response.
// the spend amount is included because an error can occur after the tokens
// are spent.
func txRes(spend *big.Int, code transactions.TxCode, err error) *TxResponse {
	if spend == nil {
		spend = big.NewInt(0)
	}

	return &TxResponse{
		ResponseCode: code,
		Spend:        spend.Int64(),
		Error:        err,
	}
}

// lofIfErr logs an error to TxApp if it is not nil.
// it should be used when committing or rolling back a transaction.
func logErr(l log.Logger, err error) {
	if err != nil {
		l.Error("error committing/rolling back transaction", zap.Error(err))
	}
}

func computeEmptyVoteBodyTxSize(chainID string) (int64, error) {
	// Create a transaction with an empty payload to measure the fixed size without the payload.
	tx, err := transactions.CreateTransaction(&transactions.ValidatorVoteBodies{
		Events: []*transactions.VotableEvent{},
	}, chainID, 1<<63) // large nonce using all 8 bytes of the uint64
	if err != nil {
		return 0, err
	}
	tx.Body.Fee, _ = big.NewInt(0).SetString("987654000000000000000000000000000", 10)
	sz, err := tx.MarshalBinary()
	if err != nil {
		return 0, err
	}
	return int64(len(sz)), nil
}
