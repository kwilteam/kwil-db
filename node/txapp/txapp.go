// package tx_router routes transactions to the appropriate module(s)
package txapp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils/order"
	"github.com/kwilteam/kwil-db/extensions/hooks"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/node/accounts"
	"github.com/kwilteam/kwil-db/node/types/sql"
	"github.com/kwilteam/kwil-db/node/voting"
)

// TxApp is the transaction processor for the Kwil node.
// It is responsible for interpreting payload bodies and routing them properly,
// maintaining a mempool for uncommitted accounts, pricing transactions,
// managing atomicity of the database, and managing the validator set.
type TxApp struct {
	Engine     Engine     // tracks deployed schemas
	Accounts   Accounts   // tracks account balances and nonces
	Validators Validators // tracks validator power

	service *common.Service
	// forks forks.Forks

	events Rebroadcaster
	signer auth.Signer

	mempool *mempool

	// list of resolution types
	resTypes []string // How do these get updated runtime?
}

// NewTxApp creates a new router.
func NewTxApp(ctx context.Context, db sql.Executor, engine common.Engine, signer auth.Signer,
	events Rebroadcaster, service *common.Service, accounts Accounts, validators Validators) (*TxApp, error) {
	resTypes := resolutions.ListResolutions()
	slices.Sort(resTypes)

	t := &TxApp{
		Engine:     engine,
		Accounts:   accounts,
		Validators: validators,

		events: events,
		mempool: &mempool{
			accounts:     make(map[string]*types.Account),
			accountMgr:   accounts,
			validatorMgr: validators,
			nodeAddr:     signer.Identity(),
			log:          service.Logger.New("mempool"),
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
	initialHeight int64, chainCtx *common.ChainContext) error {

	// Add Genesis Validators
	for _, validator := range validators {
		err := r.Validators.SetValidatorPower(ctx, db, validator.PubKey, validator.Power)
		if err != nil {
			return err
		}
	}

	// Fund Genesis Accounts
	for _, account := range genesisAccounts {
		err := r.Accounts.Credit(ctx, db, account.Identifier, account.Balance)
		if err != nil {
			return err
		}
	}

	// genesis hooks
	for _, hook := range hooks.ListGenesisHooks() {
		err := hook.Hook(ctx, &common.App{
			Service: r.service.NamedLogger(hook.Name),
			DB:      db,
			Engine:  r.Engine,
		}, chainCtx)
		if err != nil {
			return fmt.Errorf("error running genesis hook %s: %w", hook.Name, err)
		}
	}

	return nil
}

// Begin signals that a new block has begun. This creates an outer database
// transaction that may be committed, or rolled back on error or crash.
// It is given the starting networkParams, and is expected to use them to
// use them to store any changes to the network parameters in the database during Finalize.
func (r *TxApp) Begin(ctx context.Context, height int64) error {
	// Before executing transaction in this block, add/remove/update functionality.
	// TODO: active forks, not for the beta
	return nil
}

// Execute executes a transaction.  It will route the transaction to the
// appropriate module(s) for execution and return the response.
// This method must only be called from the consensus engine,
// sequentially, when executing transactions in a block.
func (r *TxApp) Execute(ctx *common.TxContext, db sql.DB, tx *types.Transaction) *TxResponse {
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
func (r *TxApp) Finalize(ctx context.Context, db sql.DB, block *common.BlockContext) (finalValidators []*types.Validator, err error) {
	err = r.processVotes(ctx, db, block)
	if err != nil {
		return nil, err
	}

	return r.Validators.GetValidators(), nil
}

// Commit signals that a block's state changes should be committed.
func (r *TxApp) Commit() error {
	r.Accounts.Commit()
	r.Validators.Commit()

	r.mempool.reset()

	return nil
}

func (r *TxApp) Rollback() {
	r.Accounts.Rollback()
	r.Validators.Rollback()

	r.mempool.reset() // will issue recheck before next block
}

// processVotes confirms resolutions that have been approved by the network,
// expires resolutions that have expired, and properly credits proposers and voters.
func (r *TxApp) processVotes(ctx context.Context, db sql.DB, block *common.BlockContext) error {
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
				ResolveFunc func(ctx context.Context, app *common.App, resolution *resolutions.Resolution, block *common.BlockContext) error
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

		err = resolveFunc.ResolveFunc(ctx, &common.App{
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
func (r *TxApp) Price(ctx context.Context, dbTx sql.DB, tx *types.Transaction, chainContext *common.ChainContext) (*big.Int, error) {
	if chainContext.NetworkParameters.DisabledGasCosts {
		return big.NewInt(0), nil
	}

	route := getRoute(tx.Body.PayloadType.String())
	if route == nil {
		return nil, fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String())
	}

	return route.Price(ctx, r, dbTx, tx)
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
func (r *TxApp) checkAndSpend(ctx *common.TxContext, tx *types.Transaction, pricer Pricer, dbTx sql.DB) (*big.Int, types.TxCode, error) {
	amt := big.NewInt(0)
	var err error

	if !ctx.BlockContext.ChainContext.NetworkParameters.DisabledGasCosts {
		amt, err = pricer.Price(ctx.Ctx, r, dbTx, tx)
		if err != nil {
			return nil, types.CodeUnknownError, err
		}
	}

	// Get account info
	account, err := r.Accounts.GetAccount(ctx.Ctx, dbTx, tx.Sender)
	if err == nil {
		r.service.Logger.Info("account info", "account", tx.Sender, "balance", account.Balance, "nonce", account.Nonce)
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
			return account.Balance, types.CodeInsufficientBalance, fmt.Errorf("transaction tries to spend %s tokens, but account only has %s tokens", amt.String(), tx.Body.Fee.String())
		}
		if err != nil {
			if errors.Is(err, accounts.ErrAccountNotFound) {
				return nil, types.CodeInsufficientBalance, errors.New("account has zero balance")
			}
			return nil, types.CodeUnknownError, err
		}

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

		return account.Balance, types.CodeInsufficientBalance, fmt.Errorf("transaction tries to spend %s tokens, but account has %s tokens", amt.String(), account.Balance.String())
	}
	if err != nil {
		if errors.Is(err, accounts.ErrAccountNotFound) { // probably wouldn't have passed the fee check
			return nil, types.CodeInsufficientBalance, errors.New("account has zero balance")
		}
		return nil, types.CodeUnknownError, err
	}

	return amt, types.CodeOk, nil
}

// ApplyMempool applies the transactions in the mempool.
// If it returns an error, then the transaction is invalid.
func (r *TxApp) ApplyMempool(ctx *common.TxContext, db sql.DB, tx *types.Transaction) error {
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

func (r *TxApp) GetValidators() []*types.Validator {
	return r.Validators.GetValidators()
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
	validators := r.Validators.GetValidators()
	return validatorSetPower(validators), nil
}

// logErr logs an error to TxApp if it is not nil.
// it should be used when committing or rolling back a transaction.
func logErr(l log.Logger, err error) {
	if err != nil {
		l.Error("error committing/rolling back transaction", "error", err)
	}
}
