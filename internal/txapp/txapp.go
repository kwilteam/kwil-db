// package tx_router routes transactions to the appropriate module(s)
package txapp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"sync"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/common/chain/forks"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/serialize"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/core/utils/order"
	authExt "github.com/kwilteam/kwil-db/extensions/auth"
	"github.com/kwilteam/kwil-db/extensions/consensus"
	"github.com/kwilteam/kwil-db/extensions/hooks"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/voting"
)

// TxApp maintains the state for Kwil's ABCI application.
// It is responsible for interpreting payload bodies and routing them properly,
// maintaining a mempool for uncommitted accounts, pricing transactions,
// managing atomicity of the database, and managing the validator set.
type TxApp struct {
	Engine common.Engine // tracks deployed schemas
	// The various internal stores (accounts, votes, etc.) are accessed through
	// the Database via the functions defined in relevant packages.

	forks forks.Forks

	events Rebroadcaster

	chainID string
	signer  *auth.Ed25519Signer

	log log.Logger

	mempool    *mempool
	validators []*types.Validator // used to optimize reads, gets updated at the block boundaries
	valMtx     sync.RWMutex       // protects validators access
	valChans   []chan []*types.Validator

	extensionConfigs map[string]map[string]string

	// precomputed variables
	emptyVoteBodyTxSize int64
	resTypes            []string

	// Tracks spends during migration
	spends []*Spend

	// list of pubkeys of join candidates approved by this node in the current block
	approvedJoins [][]byte
}

// NewTxApp creates a new router.
func NewTxApp(ctx context.Context, db sql.Executor, engine common.Engine, signer *auth.Ed25519Signer,
	events Rebroadcaster, chainParams *chain.GenesisConfig,
	extensionConfigs map[string]map[string]string, log log.Logger) (*TxApp, error) {
	voteBodyTxSize, err := computeEmptyVoteBodyTxSize(chainParams.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to compute empty vote body tx size: %w", err)
	}

	resTypes := resolutions.ListResolutions()
	slices.Sort(resTypes)

	t := &TxApp{
		Engine: engine,
		events: events,
		log:    log,
		mempool: &mempool{accounts: make(map[string]*types.Account),
			nodeAddr: signer.Identity(),
		},
		signer:              signer,
		chainID:             chainParams.ChainID,
		extensionConfigs:    extensionConfigs,
		emptyVoteBodyTxSize: voteBodyTxSize,
		resTypes:            resTypes,
	}
	t.forks.FromMap(chainParams.ForkHeights)
	return t, nil
}

// GenesisInit initializes the TxApp. It must be called outside of a session,
// and before any session is started.
// It can assign the initial validator set and initial account balances.
// It is only called once for a new chain.
func (r *TxApp) GenesisInit(ctx context.Context, db sql.DB, validators []*types.Validator, genesisAccounts []*types.Account,
	initialHeight int64, chain *common.ChainContext) error {

	// Add Genesis Validators
	var voters []*types.Validator
	r.valMtx.Lock()
	defer r.valMtx.Unlock()

	for _, validator := range validators {
		err := setVoterPower(ctx, db, validator.PubKey, validator.Power)
		if err != nil {
			return err
		}
		voters = append(voters, validator)
	}
	r.validators = voters

	// Fund Genesis Accounts
	for _, account := range genesisAccounts {
		err := credit(ctx, db, account.Identifier, account.Balance)
		if err != nil {
			return err
		}
	}

	// genesis hooks
	for _, hook := range hooks.ListGenesisHooks() {
		err := hook(ctx, &common.App{
			Service: &common.Service{
				Logger:           r.log.Sugar(),
				ExtensionConfigs: r.extensionConfigs,
				Identity:         r.signer.Identity(),
			},
			DB:     db,
			Engine: r.Engine,
		}, chain)
		if err != nil {
			return fmt.Errorf("error running genesis hook: %w", err)
		}
	}

	return nil
}

// Reload reloads the database state into the engine.
func (r *TxApp) Reload(ctx context.Context, db sql.DB) error {
	// Reload the engine internal state from the updated database state
	return r.Engine.Reload(ctx, db)
}

// UpdateValidator updates a validator's power.
// It can only be called in between Begin and Finalize.
// The value passed as power will simply replace the current power.
func (r *TxApp) UpdateValidator(ctx context.Context, db sql.DB, validator []byte, power int64) error {
	return setVoterPower(ctx, db, validator, power)
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
// It will return ONLY committed changes.
func (r *TxApp) GetValidators(ctx context.Context, db sql.DB) ([]*types.Validator, error) {
	r.valMtx.Lock()
	defer r.valMtx.Unlock()

	// if we have a cached validator set, return it
	if r.validators != nil {
		return slices.Clone(r.validators), nil
	}

	// NOTE: we aren't saving this to r.validators, leaving that to next FinalizeBlock. We could though...
	return getAllVoters(ctx, db)
}

// CachedValidators returns the current validator set that is cached, and whether or not it is valid.
// This can be used by TxApp consumers to try to get a validator set without hitting the database.
func (r *TxApp) CachedValidators() ([]*types.Validator, bool) {
	r.valMtx.RLock()
	defer r.valMtx.RUnlock()
	if r.validators == nil {
		return nil, false
	}

	return slices.Clone(r.validators), true
}

func validatorSetPower(validators []*types.Validator) int64 {
	var totalPower int64
	for _, v := range validators {
		totalPower += v.Power
	}
	return totalPower
}

// validatorsPower returns the total power of the current validator set
// according to the DB.
func (r *TxApp) validatorSetPower(ctx context.Context, tx sql.Executor) (int64, error) {
	validators, err := getAllVoters(ctx, tx)
	if err != nil {
		return 0, err
	}
	return validatorSetPower(validators), nil
}

// Execute executes a transaction.  It will route the transaction to the
// appropriate module(s) and return the response. This method must only be
// called from the consensus engine, sequentially, when executing transactions
// in a block.
func (r *TxApp) Execute(ctx TxContext, db sql.DB, tx *transactions.Transaction) *TxResponse {
	route, ok := routes[tx.Body.PayloadType.String()] // and RegisterRoute call is not concurrent
	if !ok {
		return txRes(nil, transactions.CodeInvalidTxType, fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String()))
	}

	r.log.Debug("executing transaction", log.Any("tx", tx))

	// Check if the tx is a approval vote by this node for a validator join request
	// record the approval in the approvedJoins list
	r.logValidatorJoinApprovals(tx)

	return route.Execute(ctx, r, db, tx)
}

func (r *TxApp) logValidatorJoinApprovals(tx *transactions.Transaction) {
	if tx.Body.PayloadType != transactions.PayloadTypeValidatorApprove {
		return
	}

	if !bytes.Equal(tx.Sender, r.signer.Identity()) {
		return
	}

	approve := &transactions.ValidatorApprove{}
	if err := approve.UnmarshalBinary(tx.Body.Payload); err != nil {
		return
	}

	r.approvedJoins = append(r.approvedJoins, approve.Candidate)

}

// Begin signals that a new block has begun. This creates an outer database
// transaction that may be committed, or rolled back on error or crash.
// It is given the starting networkParams, and is expected to use them to
// use them to store any changes to the network parameters in the database during Finalize.
func (r *TxApp) Begin(ctx context.Context, height int64) error {
	// Before executing transaction in this block, add/remove/update functionality.
	forks := r.activations(height)
	if len(forks) > 0 {
		r.log.Infof("Forks activating at height %d: %v", height, len(forks))
	}
	for _, fork := range forks {
		r.log.Info("Hardfork activating", log.String("fork", fork.Name))

		// Update transaction payloads.
		for _, newPayload := range fork.TxPayloads {
			r.log.Infof("Registering transaction route for payload type %s", newPayload.Type)
			if err := RegisterRouteImpl(newPayload.Type, newPayload.Route); err != nil {
				return fmt.Errorf("failed to register route for payload %v: %w", newPayload.Type, err)
			}
		}
		// Update authenticators.
		for _, authMod := range fork.AuthUpdates {
			authExt.RegisterAuthenticator(authMod.Operation, authMod.Name, authMod.Authn)
		}
		// Update resolutions.
		for _, resMod := range fork.ResolutionUpdates {
			resolutions.RegisterResolution(resMod.Name, resMod.Operation, *resMod.Config)
		}
		// Update serialization codecs.
		for _, enc := range fork.Encoders {
			serialize.RegisterCodec(enc)
		}
	}

	return nil
}

// Activations consults chain config for the names of hard forks that activate
// *at* the given block height, and retrieves the associated changes from the
// consensus package that contains the canonical and extended fork definitions.
func (r *TxApp) activations(height int64) []*consensus.Hardfork {
	var activations []*consensus.Hardfork
	activationNames := r.forks.ActivatesAt(uint64(height)) // chain.Forks.ActivatesAt()
	for _, name := range activationNames {
		fork := consensus.Hardforks[name]
		if fork == nil {
			r.log.Errorf("hardfork %v at height %d has no definition", name, height)
			continue // really could be a panic
		}
		activations = append(activations, fork) // how to handle multiple at same height? alphabetical??
	}
	return activations
}

// Finalize signals that a block has been finalized. No more changes can be
// applied to the database. It returns the apphash and the validator set. And
// state modifications specified by hardforks activating at this height are
// applied. It is given the old and new network parameters, and is expected to
// use them to store any changes to the network parameters in the database.
func (r *TxApp) Finalize(ctx context.Context, db sql.DB, block *common.BlockContext) (finalValidators []*types.Validator, approvedJoins, expiredJoins [][]byte, err error) {
	expiredJoins, err = r.processVotes(ctx, db, block)
	if err != nil {
		return nil, nil, nil, err
	}

	finalValidators, err = getAllVoters(ctx, db)
	if err != nil {
		return nil, nil, nil, err
	}

	// Execute state modifications for the hard forks that activate at this
	// height. These changes are associated with other consensus logic or
	// parameters changes, otherwise a resolution might be more sensible.
	for _, fork := range r.activations(block.Height) {
		if fork.StateMod == nil {
			continue
		}
		r.log.Info("running StateMod", log.String("hardfork", fork.Name))
		if err := fork.StateMod(ctx, &common.App{
			Service: &common.Service{
				Logger:           r.log.Sugar(),
				ExtensionConfigs: r.extensionConfigs,
				Identity:         r.signer.Identity(),
			},
			DB:     db,
			Engine: r.Engine,
		}); err != nil {
			return nil, nil, nil, err
		}
	}

	// end block hooks
	for _, hook := range hooks.ListEndBlockHooks() {
		err := hook(ctx, &common.App{
			Service: &common.Service{
				Logger:           r.log.Sugar(),
				ExtensionConfigs: r.extensionConfigs,
				Identity:         r.signer.Identity(),
			},
			DB:     db,
			Engine: r.Engine,
		}, block)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error running end block hook: %w", err)
		}
	}

	r.valMtx.Lock()
	r.validators = finalValidators
	r.valMtx.Unlock()

	return finalValidators, r.approvedJoins, expiredJoins, nil
}

// processVotes confirms resolutions that have been approved by the network,
// expires resolutions that have expired, and properly credits proposers and voters.
func (r *TxApp) processVotes(ctx context.Context, db sql.DB, block *common.BlockContext) ([][]byte, error) {
	credits := make(creditMap)

	var finalizedIDs []*types.UUID
	// markedProcessedIDs is a separate list for marking processed, since we do not want to process validator resolutions
	// validator vote IDs are not unique, so we cannot mark them as processed, in case a validator leaves and joins again
	var markProcessedIDs []*types.UUID
	// resolveFuncs tracks the resolve function for each resolution, in the order they are queried.
	// we track this and execute all of these functions after we have found all confirmed resolutions
	// because a resolve function can change a validator's power. This would then change the required power
	// for subsequent resolutions in the same block, which should not happen.
	var resolveFuncs []*struct {
		Resolution  *resolutions.Resolution
		ResolveFunc func(ctx context.Context, app *common.App, resolution *resolutions.Resolution, block *common.BlockContext) error
	}

	totalPower, err := r.validatorSetPower(ctx, db)
	if err != nil {
		return nil, err
	}

	for _, resolutionType := range r.resTypes {
		cfg, err := resolutions.GetResolution(resolutionType)
		if err != nil {
			return nil, err
		}

		finalized, err := getResolutionsByThresholdAndType(ctx, db, cfg.ConfirmationThreshold, resolutionType, totalPower)
		if err != nil {
			return nil, err
		}

		for _, resolution := range finalized {
			credits.applyResolution(resolution)
			finalizedIDs = append(finalizedIDs, resolution.ID)

			// we do not want to mark processed for validator join and remove events, as they can occur again
			if resolution.Type != voting.ValidatorJoinEventType && resolution.Type != voting.ValidatorRemoveEventType {
				markProcessedIDs = append(markProcessedIDs, resolution.ID)
			}

			resolveFuncs = append(resolveFuncs, &struct {
				Resolution  *resolutions.Resolution
				ResolveFunc func(ctx context.Context, app *common.App, resolution *resolutions.Resolution, block *common.BlockContext) error
			}{
				Resolution:  resolution,
				ResolveFunc: cfg.ResolveFunc,
			})
		}
	}

	// apply all resolutions
	for _, resolveFunc := range resolveFuncs {
		r.log.Debug("resolving resolution", log.String("type", resolveFunc.Resolution.Type), log.String("id", resolveFunc.Resolution.ID.String()))

		tx, err := db.BeginTx(ctx)
		if err != nil {
			return nil, err
		}

		err = resolveFunc.ResolveFunc(ctx, &common.App{
			Service: &common.Service{
				Logger:           r.log.Named("resolution_" + resolveFunc.Resolution.Type).Sugar(),
				ExtensionConfigs: r.extensionConfigs,
				Identity:         r.signer.Identity(),
			},
			DB:     tx,
			Engine: r.Engine,
		}, resolveFunc.Resolution, block)
		if err != nil {
			err2 := tx.Rollback(ctx)
			if err2 != nil {
				return nil, fmt.Errorf("error rolling back transaction: %s, error: %s", err.Error(), err2.Error())
			}

			// if the resolveFunc fails, we should still continue on, since it simply means
			// some business logic failed in a deployed schema.
			r.log.Warn("error resolving resolution", log.String("type", resolveFunc.Resolution.Type), log.String("id", resolveFunc.Resolution.ID.String()), log.Error(err))
			continue
		}

		err = tx.Commit(ctx)
		if err != nil {
			return nil, err
		}
	}

	// now we will expire resolutions
	expired, err := getExpired(ctx, db, block.Height)
	if err != nil {
		return nil, err
	}

	expiredIDs := make([]*types.UUID, 0, len(expired))
	requiredPowerMap := make(map[string]int64) // map of resolution type to required power
	var expiredJoins [][]byte

	for _, resolution := range expired {
		expiredIDs = append(expiredIDs, resolution.ID)
		if resolution.Type != voting.ValidatorJoinEventType && resolution.Type != voting.ValidatorRemoveEventType {
			markProcessedIDs = append(markProcessedIDs, resolution.ID)
		}

		if resolution.Type == voting.ValidatorJoinEventType {
			req := &voting.UpdatePowerRequest{}
			if err := req.UnmarshalBinary(resolution.Body); err != nil {
				return nil, fmt.Errorf("failed to unmarshal join request: %w", err)
			}

			expiredJoins = append(expiredJoins, req.PubKey)
		}

		threshold, ok := requiredPowerMap[resolution.Type]
		if !ok {
			cfg, err := resolutions.GetResolution(resolution.Type)
			if err != nil {
				return nil, err
			}

			// we need to use each configured resolutions refund threshold
			requiredPowerMap[resolution.Type] = requiredPower(ctx, db, cfg.RefundThreshold, totalPower)
		}
		// if it has enough power, we will still refund
		refunded := resolution.ApprovedPower >= threshold
		if refunded {
			credits.applyResolution(resolution)
		}

		r.log.Debug("expiring resolution", log.String("type", resolution.Type),
			log.String("id", resolution.ID.String()), log.Bool("refunded", refunded))
	}

	allIDs := append(finalizedIDs, expiredIDs...)
	err = deleteResolutions(ctx, db, allIDs...)
	if err != nil {
		return nil, err
	}

	err = markProcessed(ctx, db, markProcessedIDs...)
	if err != nil {
		return nil, err
	}

	// This is to ensure that the nodes that never get to vote on this event due to limitation
	// per block vote sizes, they never get to vote and essentially delete the event
	// So this is handled instead when the nodes are approved.
	// TODO: We need to figure out the consequences of resolutions getting expired due to the vote limits set per block. There can be scenarios where the events are observed by the nodes, but before they can vote, the event gets expired. rare but possible in the situations with higher event traffic.
	err = deleteEvents(ctx, db, markProcessedIDs...)
	if err != nil {
		return nil, err
	}

	// now we will apply credits if gas is enabled.
	// Since it is a map, we need to order it for deterministic results.
	if !block.ChainContext.NetworkParameters.DisabledGasCosts {
		for _, kv := range order.OrderMap(credits) {
			err = credit(ctx, db, []byte(kv.Key), kv.Value)
			if err != nil {
				return nil, err
			}
		}
	}

	return expiredJoins, nil
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
func (r *TxApp) Commit(ctx context.Context) {
	r.announceValidators()
	r.mempool.reset()
	r.approvedJoins = nil
	r.spends = nil // reset spends for the next block
}

// ApplyMempool applies the transactions in the mempool.
// If it returns an error, then the transaction is invalid.
func (r *TxApp) ApplyMempool(ctx *common.TxContext, db sql.DB, tx *transactions.Transaction) error {
	// check that payload type is valid
	if getRoute(tx.Body.PayloadType.String()) == nil {
		return fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String())
	}

	return r.mempool.applyTransaction(ctx, tx, db, r.events)
}

// AccountInfo gets account info from either the mempool or the account store.
// It takes a flag to indicate whether it should check the mempool first.
func (r *TxApp) AccountInfo(ctx context.Context, db sql.DB, acctID []byte, getUnconfirmed bool) (balance *big.Int, nonce int64, err error) {
	var a *types.Account
	if getUnconfirmed {
		a, err = r.mempool.accountInfoSafe(ctx, db, acctID)
	} else {
		a, err = getAccount(ctx, db, acctID)
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
func (r *TxApp) ProposerTxs(ctx context.Context, db sql.DB, txNonce uint64, maxTxsSize int64, block *common.BlockContext) ([][]byte, error) {
	acct, err := getAccount(ctx, db, block.Proposer)
	if err != nil {
		return nil, err
	}
	bal, nonce := acct.Balance, acct.Nonce

	if !block.ChainContext.NetworkParameters.DisabledGasCosts && nonce == 0 && bal.Sign() == 0 {
		r.log.Debug("proposer account has no balance, not allowed to propose any new transactions", log.Int("height", block.Height))
		return nil, nil
	}

	if txNonce == 0 {
		txNonce = uint64(nonce) + 1
	}

	maxTxsSize -= r.emptyVoteBodyTxSize + 1000 // empty payload size + 1000 safety buffer
	events, err := getEvents(ctx, db)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		r.log.Debug("no events to propose", log.Int("height", block.Height))
		return nil, nil
	}

	ids := make([]*types.UUID, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID())
	}

	// Is thre any reason to check for notProcessed events here? Becase event store will never have events that are already processed.

	// Limit upto only 50 VoteBodies per block
	if len(ids) > int(block.ChainContext.NetworkParameters.MaxVotesPerTx) {
		ids = ids[:block.ChainContext.NetworkParameters.MaxVotesPerTx]
	}

	eventMap := make(map[types.UUID]*types.VotableEvent)
	for _, evt := range events {
		eventMap[*evt.ID()] = evt
	}

	var finalEvents []*transactions.VotableEvent
	for _, id := range ids {
		event, ok := eventMap[*id]
		if !ok { // this should never happen
			return nil, fmt.Errorf("internal bug: event with id %s not found", id.String())
		}

		evtSz := int64(len(event.Type)) + int64(len(event.Body)) + eventRLPSize
		if evtSz > maxTxsSize {
			r.log.Debug("reached maximum proposer tx size", log.Int("height", block.Height))
			break
		}
		maxTxsSize -= evtSz
		finalEvents = append(finalEvents, &transactions.VotableEvent{
			Type: event.Type,
			Body: event.Body,
		})
	}

	if len(finalEvents) == 0 {
		r.log.Debug("found proposer events to propose, but cannot fit them in a block",
			log.Int("height", block.Height),
			log.Int("maxTxsSize", maxTxsSize),
			log.Int("emptyVoteBodyTxSize", r.emptyVoteBodyTxSize),
			log.Int("foundEvents", len(events)),
			log.Int("maxVotesPerTx", block.ChainContext.NetworkParameters.MaxVotesPerTx),
		)
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
	amt, err := r.Price(ctx, db, tx, block.ChainContext)
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
var eventRLPSize = func() int64 {
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
func (r *TxApp) Price(ctx context.Context, dbTx sql.DB, tx *transactions.Transaction, chainContext *common.ChainContext) (*big.Int, error) {
	if chainContext.NetworkParameters.DisabledGasCosts {
		return big.NewInt(0), nil
	}

	route := getRoute(tx.Body.PayloadType.String())
	if route == nil {
		return nil, fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String())
	}

	return route.Price(ctx, r, dbTx, tx)
}

type Spend struct {
	Account []byte
	Amount  *big.Int
	Nonce   uint64
}

// ApplySpend applies a spend to the accounts database.
func (s *Spend) ApplySpend(ctx context.Context, db sql.DB) error {
	return applySpend(ctx, db, s.Account, s.Amount, int64(s.Nonce))
}

// recordSpend records a spend occurred during the block execution.
// This only records spends during migrations.
func (r *TxApp) recordSpend(ctx TxContext, spend *Spend) {
	if ctx.BlockContext.ChainContext.NetworkParameters.InMigration {
		r.spends = append(r.spends, spend)
	}
}

// GetBlockSpends returns the spends that occurred during the block.
func (r *TxApp) GetBlockSpends() []*Spend {
	return r.spends
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
func (r *TxApp) checkAndSpend(ctx TxContext, tx *transactions.Transaction, pricer Pricer, dbTx sql.DB) (*big.Int, transactions.TxCode, error) {
	amt := big.NewInt(0)
	var err error

	if !ctx.BlockContext.ChainContext.NetworkParameters.DisabledGasCosts {
		amt, err = pricer.Price(ctx.Ctx, r, dbTx, tx)
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

			// Record spend here as a spend has occurred
			r.recordSpend(ctx, &Spend{Account: tx.Sender, Amount: account.Balance, Nonce: tx.Body.Nonce})

			return account.Balance, transactions.CodeInsufficientBalance, fmt.Errorf("transaction tries to spend %s tokens, but account only has %s tokens", amt.String(), tx.Body.Fee.String())
		}
		if err != nil {
			if errors.Is(err, accounts.ErrAccountNotFound) {
				return nil, transactions.CodeInsufficientBalance, errors.New("account has zero balance")
			}
			return nil, transactions.CodeUnknownError, err
		}

		// Record spend here if in a migration
		r.recordSpend(ctx, &Spend{Account: tx.Sender, Amount: tx.Body.Fee, Nonce: tx.Body.Nonce})

		return tx.Body.Fee, transactions.CodeInsufficientFee, fmt.Errorf("transaction does not consent to spending enough tokens. transaction fee: %s, required fee: %s", tx.Body.Fee.String(), amt.String())
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

		// Record spend here
		r.recordSpend(ctx, &Spend{Account: tx.Sender, Amount: account.Balance, Nonce: tx.Body.Nonce})

		return account.Balance, transactions.CodeInsufficientBalance, fmt.Errorf("transaction tries to spend %s tokens, but account has %s tokens", amt.String(), account.Balance.String())
	}
	if err != nil {
		if errors.Is(err, accounts.ErrAccountNotFound) { // probably wouldn't have passed the fee check
			return nil, transactions.CodeInsufficientBalance, errors.New("account has zero balance")
		}
		return nil, transactions.CodeUnknownError, err
	}

	// Record spend here
	r.recordSpend(ctx, &Spend{Account: tx.Sender, Amount: amt, Nonce: tx.Body.Nonce})
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
		l.Error("error committing/rolling back transaction", log.Error(err))
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
