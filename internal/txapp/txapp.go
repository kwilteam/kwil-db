// package tx_router routes transactions to the appropriate module(s)
package txapp

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/sql"

	"go.uber.org/zap"
)

// NewTxApp creates a new router.
func NewTxApp(db DB, engine ExecutionEngine, acc AccountsStore, validators ValidatorStore,
	voteStore VoteStore, signer *auth.Ed25519Signer, events Rebroadcaster, chainID string, GasEnabled bool, log log.Logger) *TxApp {
	return &TxApp{
		Database:   db,
		Engine:     engine,
		Accounts:   acc,
		Validators: validators,
		VoteStore:  voteStore,
		events:     events,
		log:        log,
		mempool: &mempool{
			accountStore:   acc,
			accounts:       make(map[string]*accounts.Account),
			validatorStore: validators,
		},
		signer:     signer,
		chainID:    chainID,
		GasEnabled: GasEnabled,
	}
}

// TxApp maintains the state for Kwil's ABCI application.
// It is responsible for interpreting payload bodies and routing them properly,
// maintaining a mempool for uncommitted accounts, pricing transactions,
// managing atomicity of the database, and managing the validator set.
type TxApp struct {
	Database   DB              // postgres database
	Engine     ExecutionEngine // tracks deployed schemas
	Accounts   AccountsStore   // accounts
	Validators ValidatorStore  // validators
	VoteStore  VoteStore       // tracks resolutions, their votes, manages expiration
	GasEnabled bool
	events     Rebroadcaster

	chainID string
	signer  *auth.Ed25519Signer

	log log.Logger

	mempool *mempool

	// transaction that exists between Begin and Commit
	currentTx sql.OuterTx
}

// GenesisInit initializes the TxApp. It must be called outside of a session,
// and before any session is started.
// It can assign the initial validator set and initial account balances.
// It is only called once for a new chain.
func (r *TxApp) GenesisInit(ctx context.Context, validators []*types.Validator, accounts []*accounts.Account, initialHeight int64) error {
	tx, err := r.Database.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = r.Validators.GenesisInit(ctx, tx, validators, initialHeight)
	if err != nil {
		return err
	}

	for _, validator := range validators {
		err := r.VoteStore.UpdateVoter(ctx, tx, validator.PubKey, validator.Power)
		if err != nil {
			return err
		}
	}

	for _, account := range accounts {
		err := r.Accounts.Credit(ctx, tx, account.Identifier, account.Balance)
		if err != nil {
			return err
		}
	}

	_, err = tx.Precommit(ctx)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// UpdateValidator updates a validator's power.
// It can only be called in between Begin and Finalize.
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

	err = r.Validators.Update(ctx, r.currentTx, validator, power)
	if err != nil {
		return err
	}

	err = r.VoteStore.UpdateVoter(ctx, r.currentTx, validator, power)
	if err != nil {
		return err
	}

	return sp.Commit(ctx)
}

// GetValidators returns the current validator set.
// It will not return uncommitted changes.
func (r *TxApp) GetValidators(ctx context.Context) ([]*types.Validator, error) {
	readTx, err := r.Database.BeginReadTx(ctx)
	if err != nil {
		return nil, err
	}
	defer readTx.Rollback(ctx) // always rollback read tx

	return r.Validators.CurrentSet(ctx, readTx)
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
// It contains information on any validators whos power should be updated.
func (r *TxApp) Begin(ctx context.Context) error {
	if r.currentTx != nil {
		return errors.New("txapp misuse: cannot begin a new block while a transaction is in progress")
	}

	tx, err := r.Database.BeginTx(ctx)
	if err != nil {
		return err
	}

	r.currentTx = tx

	return nil
}

// Finalize signals that a block has been finalized.
// No more changes can be applied to the database, and TxApp should return
// information on the apphash and validator updates.
func (r *TxApp) Finalize(ctx context.Context, blockHeight int64) (apphash []byte, validatorUpgrades []*types.Validator, err error) {
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

	finalizedEvents, err := r.VoteStore.ProcessConfirmedResolutions(ctx, r.currentTx)
	if err != nil {
		return nil, nil, err
	}

	for _, eventID := range finalizedEvents {
		err = deleteEvent(ctx, r.currentTx, eventID)
		if err != nil {
			return nil, nil, err
		}
	}

	err = r.VoteStore.Expire(ctx, r.currentTx, blockHeight)
	if err != nil {
		return nil, nil, err
	}

	// we need to set this before we finalize validators,
	// so that pending requests expire on Finalize()
	r.Validators.UpdateBlockHeight(blockHeight)

	validatorUpdates, err := r.Validators.Finalize(ctx, r.currentTx)
	if err != nil {
		return nil, nil, err
	}

	// we intentionally update the validators after processing confirmed resolutions
	// if a vote passes and a validator is upgraded in the same block.
	for _, validator := range validatorUpdates {
		err = r.VoteStore.UpdateVoter(ctx, r.currentTx, validator.PubKey, validator.Power)
		if err != nil {
			return nil, nil, err
		}
	}

	engineHash, err := r.currentTx.Precommit(ctx)
	if err != nil {
		return nil, nil, err
	}

	validatorHash := r.Validators.StateHash()

	appHash := sha256.Sum256(append(engineHash, validatorHash...))

	return appHash[:], validatorUpdates, nil
}

// Commit signals that a block has been committed.
func (r *TxApp) Commit(ctx context.Context) error {
	if r.currentTx == nil {
		return errors.New("txapp misuse: cannot commit a block without a transaction in progress")
	}
	defer r.mempool.reset()

	err := r.currentTx.Commit(ctx)
	if err != nil {
		return err
	}

	r.currentTx = nil
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

	var a *accounts.Account
	if getUncommitted {
		a, err = r.mempool.accountInfoSafe(ctx, readTx, acctID)
	} else {
		a, err = r.Accounts.GetAccount(ctx, readTx, acctID)
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
func (r *TxApp) ProposerTxs(ctx context.Context, txNonce uint64) ([]*transactions.Transaction, error) {
	readTx, err := r.Database.BeginReadTx(ctx)
	if err != nil {
		return nil, err
	}
	defer readTx.Rollback(ctx) // always rollback read tx

	events, err := getEvents(ctx, readTx)
	if err != nil {
		return nil, err
	}

	// Final events are the events whose bodies have not been received by the network
	var finalEvents []*types.VotableEvent
	for _, event := range events {
		// Check if the event body is already received by the network
		containsBody, err := r.VoteStore.ContainsBodyOrFinished(ctx, readTx, event.ID())
		if err != nil {
			return nil, err
		}
		if !containsBody {
			finalEvents = append(finalEvents, event)
		}
	}

	if len(finalEvents) == 0 {
		return nil, nil
	}

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

	return []*transactions.Transaction{tx}, nil
}

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
func (r *TxApp) checkAndSpend(ctx TxContext, tx *transactions.Transaction, pricer Pricer, dbTx sql.DB) (*big.Int, transactions.TxCode, error) {
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
		err = r.Accounts.Spend(ctx.Ctx, dbTx, &accounts.Spend{
			AccountID: tx.Sender,
			Amount:    tx.Body.Fee,
			Nonce:     int64(tx.Body.Nonce),
		})
		if errors.Is(err, accounts.ErrInsufficientFunds) {
			return nil, transactions.CodeInsufficientBalance, fmt.Errorf("transaction tries to spend %s tokens, but account only has %s tokens", amt.String(), tx.Body.Fee.String())
		}
		if err != nil {
			return nil, transactions.CodeUnknownError, err
		}

		return nil, transactions.CodeInsufficientFee, fmt.Errorf("transaction does not consent to spending enough tokens. transaction fee: %s, required fee: %s", tx.Body.Fee.String(), amt.String())
	}

	// spend the tokens
	err = r.Accounts.Spend(ctx.Ctx, dbTx, &accounts.Spend{
		AccountID: tx.Sender,
		Amount:    amt,
		Nonce:     int64(tx.Body.Nonce),
	})
	if errors.Is(err, accounts.ErrInsufficientFunds) {
		return nil, transactions.CodeInsufficientBalance, fmt.Errorf("transaction tries to spend %s tokens, but account only has %s tokens", amt.String(), tx.Body.Fee.String())
	}
	if err != nil {
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
