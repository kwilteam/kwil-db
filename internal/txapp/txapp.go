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
func NewTxApp(db sql.OuterTxMaker, engine ExecutionEngine, acc AccountsStore, validators ValidatorStore,
	voteStore VoteStore, signer *auth.Ed25519Signer, chainID string, eventStore EventStore, GasEnabled bool, log log.Logger) *TxApp {
	return &TxApp{
		Database:   db,
		Engine:     engine,
		Accounts:   acc,
		Validators: validators,
		VoteStore:  voteStore,
		EventStore: eventStore,
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
// It is responsible for interpreting payload bodies and routing them properly.
// It also contains a mempool for uncommitted accounts, as well as pricing
// for transactions
type TxApp struct {
	Database   sql.OuterTxMaker // postgres database
	Engine     ExecutionEngine  // tracks deployed schemas
	Accounts   AccountsStore    // accounts
	Validators ValidatorStore   // validators
	VoteStore  VoteStore        // tracks resolutions, their votes, manages expiration
	EventStore EventStore       // tracks events, not part of consensus
	GasEnabled bool

	chainID string
	signer  *auth.Ed25519Signer

	log log.Logger

	mempool *mempool

	// transaction that exists between Begin and Commit
	currentTx sql.OuterTx
}

// GenesisInit initializes the VoteStore with the genesis validators.
func (r *TxApp) GenesisInit(ctx context.Context, validators []*types.Validator) error {
	for _, validator := range validators {
		err := r.VoteStore.UpdateVoter(ctx, validator.PubKey, validator.Power)
		if err != nil {
			return err
		}
	}
	return nil
}

// Execute executes a transaction.  It will route the transaction to the
// appropriate module(s) and return the response.
func (r *TxApp) Execute(ctx TxContext, tx *transactions.Transaction) *TxResponse {
	route, ok := routes[tx.Body.PayloadType.String()]
	if !ok {
		return txRes(nil, transactions.CodeInvalidTxType, fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String()))
	}

	r.log.Debug("executing transaction", zap.Any("tx", tx))

	return route.Execute(ctx, r, tx)
}

// Begin signals that a new block has begun.
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

	finalizedEvents, err := r.VoteStore.ProcessConfirmedResolutions(ctx)
	if err != nil {
		return nil, nil, err
	}

	for _, eventID := range finalizedEvents {
		err = r.EventStore.DeleteEvent(ctx, eventID)
		if err != nil {
			return nil, nil, err
		}
	}

	err = r.VoteStore.Expire(ctx, blockHeight)
	if err != nil {
		return nil, nil, err
	}

	// we need to set this before we finalize validators,
	// so that pending requests expire on Finalize()
	r.Validators.UpdateBlockHeight(blockHeight)

	validatorUpdates, err := r.Validators.Finalize(ctx)
	if err != nil {
		return nil, nil, err
	}

	// we intentionally update the validators after processing confirmed resolutions
	// if a vote passes and a validator is upgraded in the same block.
	for _, validator := range validatorUpdates {
		err = r.VoteStore.UpdateVoter(ctx, validator.PubKey, validator.Power)
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

	return r.mempool.applyTransaction(ctx, tx)
}

// GetAccount gets account info from either the mempool or the account store.
// It takes a flag to indicate whether it should check the mempool first.
func (r *TxApp) AccountInfo(ctx context.Context, acctID []byte, getUncommitted bool) (balance *big.Int, nonce int64, err error) {
	var a *accounts.Account
	if getUncommitted {
		a, err = r.mempool.accountInfoSafe(ctx, acctID)
	} else {
		a, err = r.Accounts.GetAccount(ctx, acctID)
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
func (r *TxApp) ProposerTxs(ctx context.Context, txNonce uint64) ([]*transactions.Transaction, error) {
	events, err := r.EventStore.GetEvents(ctx)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}

	tx, err := transactions.CreateTransaction(&transactions.ValidatorVoteBodies{
		Events: events,
	}, r.chainID, txNonce)
	if err != nil {
		return nil, err
	}

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
func (r *TxApp) checkAndSpend(ctx TxContext, tx *transactions.Transaction, pricer Pricer) (*big.Int, transactions.TxCode, error) {
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
		err = r.Accounts.Spend(ctx.Ctx, &accounts.Spend{
			AccountID: tx.Sender,
			Amount:    tx.Body.Fee,
			Nonce:     int64(tx.Body.Nonce),
		})
		if err != nil {
			return nil, transactions.CodeUnknownError, err
		}

		return nil, transactions.CodeInsufficientFee, fmt.Errorf("transaction does not consent to spending enough tokens. transaction fee: %s, required fee: %s", tx.Body.Fee.String(), amt.String())
	}

	// spend the tokens
	err = r.Accounts.Spend(ctx.Ctx, &accounts.Spend{
		AccountID: tx.Sender,
		Amount:    amt,
		Nonce:     int64(tx.Body.Nonce),
	})
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
