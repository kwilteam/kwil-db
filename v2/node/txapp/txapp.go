// package tx_router routes transactions to the appropriate module(s)
package txapp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"kwil/crypto/auth"
	"kwil/extensions/resolutions"
	"kwil/log"
	"kwil/node/accounts"
	"kwil/node/types/sql"
	"kwil/node/voting"
	"kwil/types"
	"kwil/utils/order"
	"math/big"
	"slices"
)

// TxApp is the transaction processer for the Kwil node.
// It is responsible for interpreting payload bodies and routing them properly,
// maintaining a mempool for uncommitted accounts, pricing transactions,
// managing atomicity of the database, and managing the validator set.
type TxApp struct {
	// Engine     types.Engine // tracks deployed schemas
	Accounts   Accounts   // tracks account balances and nonces
	Validators Validators // tracks validator power

	service *types.Service
	// forks forks.Forks

	events Rebroadcaster
	signer *auth.Ed25519Signer

	mempool *mempool

	// channels to notify the subscriber about validator updates
	// TODO: ListenerMgr is the only subscriber for now. Potentially ConsensusEngine could be another subscriber? or not as it can decide to update the validators based on the updates.
	valChans []chan []*types.Validator

	// list of resolution types
	resTypes []string // How do these get updated runtime?

	// Tracks spends during migration
	spends []*Spend
}

// NewTxApp creates a new router.
func NewTxApp(ctx context.Context, db sql.Executor, engine types.Engine, signer *auth.Ed25519Signer,
	events Rebroadcaster, service *types.Service, accounts Accounts, validators Validators) (*TxApp, error) {
	resTypes := resolutions.ListResolutions()
	slices.Sort(resTypes)

	t := &TxApp{
		// Engine: engine,
		Accounts:   accounts,
		Validators: validators,

		events: events,
		mempool: &mempool{
			accounts:     make(map[string]*types.Account),
			nodeAddr:     signer.Identity(),
			accountMgr:   accounts,
			validatorMgr: validators,
		},
		signer:   signer,
		resTypes: resTypes,
		service:  service,
	}
	// t.forks.FromMap(service.GenesisConfig.ForkHeights)
	return t, nil
}

// GenesisInit initializes the TxApp. It must be called outside of a session,
// and before any session is started.
// It can assign the initial validator set and initial account balances.
// It is only called once for a new chain.
func (r *TxApp) GenesisInit(ctx context.Context, db sql.DB, validators []*types.Validator, genesisAccounts []*types.Account,
	initialHeight int64, chain *types.ChainContext) error {

	// Add Genesis Validators
	var voters []*types.Validator

	for _, validator := range validators {
		err := r.Validators.SetValidatorPower(ctx, db, validator.PubKey, validator.Power)
		if err != nil {
			return err
		}
		voters = append(voters, validator)
	}

	// Fund Genesis Accounts
	for _, account := range genesisAccounts {
		err := r.Accounts.Credit(ctx, db, account.Identifier, account.Balance)
		if err != nil {
			return err
		}
	}

	// genesis hooks
	// for _, hook := range hooks.ListGenesisHooks() {
	// 	err := hook.Hook(ctx, &types.App{
	// 		Service: r.service.NamedLogger(hook.Name),
	// 		DB:      db,
	// 		Engine:  r.Engine,
	// 	}, chain)
	// 	if err != nil {
	// 		return fmt.Errorf("error running genesis hook: %w", err)
	// 	}
	// }

	return nil
}

// Begin signals that a new block has begun. This creates an outer database
// transaction that may be committed, or rolled back on error or crash.
// It is given the starting networkParams, and is expected to use them to
// use them to store any changes to the network parameters in the database during Finalize.
func (r *TxApp) Begin(ctx context.Context, height int64) error {
	// Before executing transaction in this block, add/remove/update functionality.
	// TODO:
	return nil
}

// Execute executes a transaction.  It will route the transaction to the
// appropriate module(s) for execution and return the response.
// This method must only be called from the consensus engine,
// sequentially, when executing transactions in a block.
func (r *TxApp) Execute(ctx *types.TxContext, db sql.DB, tx *types.Transaction) *TxResponse {
	// RegisterRoute call is not concurrent
	route, ok := routes[tx.Body.PayloadType.String()]
	if !ok {
		return txRes(nil, types.CodeInvalidTxType, fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String()))
	}

	r.service.Logger.Debug("executing transaction", "tx", tx)
	return route.Execute(ctx, r, db, tx)
}

// Finalize signals that a block has been finalized. No more changes can be
// applied to the database. It returns the apphash and the validator set. And
// state modifications specified by hardforks activating at this height are
// applied. It is given the old and new network parameters, and is expected to
// use them to store any changes to the network parameters in the database.
// TODO: Also send updates, so that the CE doesn't have to regenerate the updates.
func (r *TxApp) Finalize(ctx context.Context, db sql.DB, block *types.BlockContext) (finalValidators []*types.Validator, err error) {
	err = r.processVotes(ctx, db, block)
	if err != nil {
		return nil, err
	}

	finalValidators, err = r.Validators.GetValidators()
	if err != nil {
		return nil, err
	}

	// TODO: forks and endblock hooks

	return finalValidators, nil
}

// Commit signals that a block's state changes should be committed.
func (r *TxApp) Commit(ctx context.Context) {
	r.Accounts.Commit()
	r.Validators.Commit()

	r.announceValidators()
	r.mempool.reset()

	r.spends = nil // reset spends for the next block
}

// processVotes confirms resolutions that have been approved by the network,
// expires resolutions that have expired, and properly credits proposers and voters.
func (r *TxApp) processVotes(ctx context.Context, db sql.DB, block *types.BlockContext) error {
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
		ResolveFunc func(ctx context.Context, app *types.App, resolution *resolutions.Resolution, block *types.BlockContext) error
	}

	totalPower, err := r.validatorSetPower()
	if err != nil {
		return err
	}

	for _, resolutionType := range r.resTypes {
		cfg, err := resolutions.GetResolution(resolutionType)
		if err != nil {
			return err
		}

		finalized, err := getResolutionsByThresholdAndType(ctx, db, cfg.ConfirmationThreshold, resolutionType, totalPower)
		if err != nil {
			return err
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
				ResolveFunc func(ctx context.Context, app *types.App, resolution *resolutions.Resolution, block *types.BlockContext) error
			}{
				Resolution:  resolution,
				ResolveFunc: cfg.ResolveFunc,
			})
		}
	}

	// apply all resolutions
	for _, resolveFunc := range resolveFuncs {
		r.service.Logger.Debug("resolving resolution", "type", resolveFunc.Resolution.Type, "id", resolveFunc.Resolution.ID.String())

		tx, err := db.BeginTx(ctx)
		if err != nil {
			return err
		}

		err = resolveFunc.ResolveFunc(ctx, &types.App{
			Service: r.service.NamedLogger(resolveFunc.Resolution.Type),
			DB:      tx,
			// Engine:  r.Engine,
		}, resolveFunc.Resolution, block)
		if err != nil {
			err2 := tx.Rollback(ctx)
			if err2 != nil {
				return fmt.Errorf("error rolling back transaction: %s, error: %s", err.Error(), err2.Error())
			}

			// if the resolveFunc fails, we should still continue on, since it simply means
			// some business logic failed in a deployed schema.
			r.service.Logger.Warn("error resolving resolution", "type", resolveFunc.Resolution.Type, "id", resolveFunc.Resolution.ID.String(), "error", err)
			continue
		}

		err = tx.Commit(ctx)
		if err != nil {
			return err
		}
	}

	// now we will expire resolutions
	expired, err := getExpired(ctx, db, block.Height)
	if err != nil {
		return err
	}

	expiredIDs := make([]*types.UUID, 0, len(expired))
	requiredPowerMap := make(map[string]int64) // map of resolution type to required power

	for _, resolution := range expired {
		expiredIDs = append(expiredIDs, resolution.ID)
		if resolution.Type != voting.ValidatorJoinEventType && resolution.Type != voting.ValidatorRemoveEventType {
			markProcessedIDs = append(markProcessedIDs, resolution.ID)
		}

		threshold, ok := requiredPowerMap[resolution.Type]
		if !ok {
			cfg, err := resolutions.GetResolution(resolution.Type)
			if err != nil {
				return err
			}

			// we need to use each configured resolutions refund threshold
			requiredPowerMap[resolution.Type] = requiredPower(ctx, db, cfg.RefundThreshold, totalPower)
		}
		// if it has enough power, we will still refund
		refunded := resolution.ApprovedPower >= threshold
		if refunded {
			credits.applyResolution(resolution)
		}

		r.service.Logger.Debug("expiring resolution", "type", resolution.Type, "id", resolution.ID.String(), "refunded", refunded)
	}

	allIDs := append(finalizedIDs, expiredIDs...)
	err = deleteResolutions(ctx, db, allIDs...)
	if err != nil {
		return err
	}

	err = markProcessed(ctx, db, markProcessedIDs...)
	if err != nil {
		return err
	}

	// This is to ensure that the nodes that never get to vote on this event due to limitation
	// per block vote sizes, they never get to vote and essentially delete the event
	// So this is handled instead when the nodes are approved.
	err = deleteEvents(ctx, db, markProcessedIDs...)
	if err != nil {
		return err
	}

	// now we will apply credits if gas is enabled.
	// Since it is a map, we need to order it for deterministic results.
	if !block.ChainContext.NetworkParameters.DisabledGasCosts {
		for _, kv := range order.OrderMap(credits) {
			err = r.Accounts.Credit(ctx, db, []byte(kv.Key), kv.Value)
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

// TxResponse is the response from a transaction.
// It contains information about the transaction, such as the amount spent.
type TxResponse struct {
	// ResponseCode is the response code from the transaction
	ResponseCode types.TxCode

	// Spend is the amount of tokens spent by the transaction
	Spend int64

	// Error is the error returned by the transaction, if any
	Error error
}

// txRes wraps a spend, tx code, and error into a tx response.
// the spend amount is included because an error can occur after the tokens
// are spent.
func txRes(spend *big.Int, code types.TxCode, err error) *TxResponse {
	if spend == nil {
		spend = big.NewInt(0)
	}

	return &TxResponse{
		ResponseCode: code,
		Spend:        spend.Int64(),
		Error:        err,
	}
}

// Price estimates the price of a transaction.
// It returns the estimated price in tokens.
func (r *TxApp) Price(ctx context.Context, dbTx sql.DB, tx *types.Transaction, chainContext *types.ChainContext) (*big.Int, error) {
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

// recordSpend records a spend occurred during the block execution.
// This only records spends during migrations.
func (r *TxApp) recordSpend(ctx *types.TxContext, spend *Spend) {
	if ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationInProgress {
		r.spends = append(r.spends, spend)
	}
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
func (r *TxApp) checkAndSpend(ctx *types.TxContext, tx *types.Transaction, pricer Pricer, dbTx sql.DB) (*big.Int, types.TxCode, error) {
	amt := big.NewInt(0)
	var err error

	if !ctx.BlockContext.ChainContext.NetworkParameters.DisabledGasCosts {
		amt, err = pricer.Price(ctx.Ctx, r, dbTx, tx)
		if err != nil {
			return nil, types.CodeUnknownError, err
		}
	}

	// check if the transaction consented to spending enough tokens
	if tx.Body.Fee.Cmp(amt) < 0 {
		// If the transaction does not consent to spending required tokens for the transaction execution,
		// spend the approved tx fee and terminate the transaction
		err = r.Accounts.Spend(ctx.Ctx, dbTx, tx.Sender, tx.Body.Fee, int64(tx.Body.Nonce))
		if errors.Is(err, accounts.ErrInsufficientFunds) {
			// spend as much as possible
			account, err := r.Accounts.GetAccount(ctx.Ctx, dbTx, tx.Sender)
			if err != nil { // account will just be empty if not found
				return nil, types.CodeUnknownError, err
			}

			err2 := r.Accounts.Spend(ctx.Ctx, dbTx, tx.Sender, account.Balance, int64(tx.Body.Nonce))
			if err2 != nil {
				if errors.Is(err2, accounts.ErrAccountNotFound) {
					return nil, types.CodeInsufficientBalance, errors.New("account has zero balance")
				}
				return nil, types.CodeUnknownError, err2
			}

			// Record spend here as a spend has occurred
			r.recordSpend(ctx, &Spend{Account: tx.Sender, Amount: account.Balance, Nonce: tx.Body.Nonce})

			return account.Balance, types.CodeInsufficientBalance, fmt.Errorf("transaction tries to spend %s tokens, but account only has %s tokens", amt.String(), tx.Body.Fee.String())
		}
		if err != nil {
			if errors.Is(err, accounts.ErrAccountNotFound) {
				return nil, types.CodeInsufficientBalance, errors.New("account has zero balance")
			}
			return nil, types.CodeUnknownError, err
		}

		// Record spend here if in a migration
		r.recordSpend(ctx, &Spend{Account: tx.Sender, Amount: tx.Body.Fee, Nonce: tx.Body.Nonce})

		return tx.Body.Fee, types.CodeInsufficientFee, fmt.Errorf("transaction does not consent to spending enough tokens. transaction fee: %s, required fee: %s", tx.Body.Fee.String(), amt.String())
	}

	// spend the tokens
	err = r.Accounts.Spend(ctx.Ctx, dbTx, tx.Sender, amt, int64(tx.Body.Nonce))
	if errors.Is(err, accounts.ErrInsufficientFunds) {
		// spend as much as possible
		account, err := r.Accounts.GetAccount(ctx.Ctx, dbTx, tx.Sender)
		if err != nil {
			return nil, types.CodeUnknownError, err
		}

		err2 := r.Accounts.Spend(ctx.Ctx, dbTx, tx.Sender, account.Balance, int64(tx.Body.Nonce))
		if err2 != nil {
			return nil, types.CodeUnknownError, err2
		}

		// Record spend here
		r.recordSpend(ctx, &Spend{Account: tx.Sender, Amount: account.Balance, Nonce: tx.Body.Nonce})

		return account.Balance, types.CodeInsufficientBalance, fmt.Errorf("transaction tries to spend %s tokens, but account has %s tokens", amt.String(), account.Balance.String())
	}
	if err != nil {
		if errors.Is(err, accounts.ErrAccountNotFound) { // probably wouldn't have passed the fee check
			return nil, types.CodeInsufficientBalance, errors.New("account has zero balance")
		}
		return nil, types.CodeUnknownError, err
	}

	// Record spend here
	r.recordSpend(ctx, &Spend{Account: tx.Sender, Amount: amt, Nonce: tx.Body.Nonce})
	return amt, types.CodeOk, nil
}

// GetBlockSpends returns the spends that occurred during the block.
func (r *TxApp) GetBlockSpends() []*Spend { // If we track spends in the account store, can we simplify this?
	return r.spends
}

// ApplyMempool applies the transactions in the mempool.
// If it returns an error, then the transaction is invalid.
func (r *TxApp) ApplyMempool(ctx *types.TxContext, db sql.DB, tx *types.Transaction) error {
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
		a, err = r.Accounts.GetAccount(ctx, db, acctID)
	}
	if err != nil {
		return nil, 0, err
	}

	return a.Balance, a.Nonce, nil
}

// UpdateValidator updates a validator's power.
// It can only be called in between Begin and Finalize.
// The value passed as power will simply replace the current power.
func (r *TxApp) UpdateValidator(ctx context.Context, db sql.DB, validator []byte, power int64) error {
	return r.Validators.SetValidatorPower(ctx, db, validator, power)
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
	c := make(chan []*types.Validator, 1)
	r.valChans = append(r.valChans, c)
	return c
}

// announceValidators sends the current validator list to subscribers from
// ReceiveValidators.
func (r *TxApp) announceValidators() {
	// dev note: this method should not be blocked by receivers. Keep a default
	// case and create buffered channels.

	if len(r.valChans) == 0 {
		return // no subscribers, skip the slice clone
	}

	vals, err := r.Validators.GetValidators()
	if err != nil {
		r.service.Logger.Error("error getting validators", "error", err)
		return
	}

	for _, c := range r.valChans {
		select {
		case c <- vals:
		default: // they'll get the next one... this is just supposed to be better than polling
			r.service.Logger.Warn("Validator update channel is blocking")
		}
	}
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
func (r *TxApp) validatorSetPower() (int64, error) {
	validators, err := r.Validators.GetValidators()
	if err != nil {
		return 0, err
	}
	return validatorSetPower(validators), nil
}

// logErr logs an error to TxApp if it is not nil.
// it should be used when committing or rolling back a transaction.
func logErr(l log.Logger, err error) {
	if err != nil {
		l.Error("error committing/rolling back transaction", "error", err)
	}
}
