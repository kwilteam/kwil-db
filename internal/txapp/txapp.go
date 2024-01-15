// package tx_router routes transactions to the appropriate module(s)
package txapp

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/extensions/oracles"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"go.uber.org/zap"
)

// NewTxApp creates a new router.
func NewTxApp(db DatabaseEngine, acc AccountsStore, validators ValidatorStore, atomicCommitter AtomicCommitter,
	voteStore VoteStore, signer *auth.Ed25519Signer, chainID string, eventStore EventStore, log log.Logger) *TxApp {
	return &TxApp{
		// TODO: set the eventstore and votestore dependencies
		Database:   db,
		Accounts:   acc,
		Validators: validators,
		VoteStore:  voteStore,
		EventStore: eventStore,

		atomicCommitter: atomicCommitter,
		log:             log,
		mempool: &mempool{
			accountStore:   acc,
			accounts:       make(map[string]*accounts.Account),
			validatorStore: validators,
		},
		oraclesUp: false,
		signer:    signer,
		chainID:   chainID,
	}
}

// TxApp maintains the state for Kwil's ABCI application.
// It is responsible for interpreting payload bodies and routing them properly.
// It also contains a mempool for uncommitted accounts, as well as pricing
// for transactions
type TxApp struct {
	Database   DatabaseEngine // tracks deployed schemas
	Accounts   AccountsStore  // accounts
	Validators ValidatorStore // validators
	VoteStore  VoteStore      // tracks resolutions, their votes, manages expiration
	EventStore EventStore     // tracks events, not part of consensus

	CometNode *cometbft.CometBftNode // comet node
	chainID   string
	signer    *auth.Ed25519Signer
	oraclesUp bool

	log log.Logger

	atomicCommitter AtomicCommitter
	mempool         *mempool
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
func (r *TxApp) Begin(ctx context.Context, blockHeight int64) error {
	idempotencyKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(idempotencyKey, uint64(blockHeight))

	r.log.Debug("beginning block", zap.Int64("blockHeight", blockHeight))

	isValidator, err := r.Validators.IsCurrent(ctx, r.signer.Identity())
	if err != nil {
		return err
	}

	// Check if the node is in a catchup-mode
	// start oracles only if the node is a validator
	if isValidator && !r.CometNode.IsCatchup() && !r.oraclesUp {
		// Start the oracles for the current block
		r.log.Debug("starting oracles")
		r.oraclesUp = true
		regOracles := oracles.RegisteredOracles()
		for name, oracle := range regOracles {
			go func(name string, inst oracles.Oracle) {
				r.log.Debug("Starting oracle", zap.String("oracle", name))
				err := inst.Start(ctx)
				if err != nil {
					r.log.Error("error starting oracle", zap.String("oracle", name), zap.Error(err))
					return
				}
			}(name, oracle)
		}
	}

	return r.atomicCommitter.Begin(ctx, idempotencyKey)
}

// Commit signals that a block has been committed.
// TODO: once we use postgres, this will no longer be applicable
// we will need a separate function for getting end results and committing
// Right now, Commit is called in FinalizeBlock in abci. However, it should
// be called in Commit.  The reason we can get away with this is because
// we rely on idempotency keys to ensure we don't double execute to a datastore.
// With Postgres, we will simply rely on its cross-schema
// transaction support.  Therefore, we should have another method here called
// GetEndResults.
// Commit also clears the mempool.
// It takes a `syncMode` parameter, which signals whether or not the node is currently syncing data
func (r *TxApp) Commit(ctx context.Context, blockHeight int64) (apphash []byte, validatorUpgrades []*types.Validator, err error) {
	// this would go in Commit
	defer r.mempool.reset()

	r.log.Debug("committing block", zap.Int64("blockHeight", blockHeight))

	// this would go in finalize block
	// run all approved votes, and delete from local store
	finalizedEvents, err := r.VoteStore.ProcessConfirmedResolutions(ctx)
	if err != nil {
		return nil, nil, err
	}

	// this would also go in finalize block
	for _, eventID := range finalizedEvents {
		err = r.EventStore.DeleteEvent(ctx, eventID)
		if err != nil {
			return nil, nil, err
		}
	}

	// expire votes
	// this would go in FinalizeBlock
	err = r.VoteStore.Expire(ctx, blockHeight)
	if err != nil {
		return nil, nil, err
	}

	// this would go in GetEndResults
	validatorUpdates, err := r.Validators.Finalize(ctx)
	if err != nil {
		return nil, nil, err
	}

	// we intentionally update the validators after processing confirmed resolutions
	// if a vote passes and a validator is upgraded in the same block, it doesn't make sense
	// for that new validator's votes to have an impact an the otherwise "confirmed" resolution
	for _, validator := range validatorUpdates {
		err = r.VoteStore.UpdateVoter(ctx, validator.PubKey, validator.Power)
		if err != nil {
			return nil, nil, err
		}
	}

	// this would go in Commit
	idempotencyKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(idempotencyKey, uint64(blockHeight))

	// appHash would go in GetEndResults,
	// the commit would go in Commit
	appHash, err := r.atomicCommitter.Commit(ctx, idempotencyKey)
	if err != nil {
		return nil, nil, err
	}

	// this only updates an in-memory value. but it seems weird to me that the validator store needs to be aware
	// of the current block height and "keep it"
	// this would go in Commit
	r.Validators.UpdateBlockHeight(blockHeight)
	return appHash, validatorUpdates, nil
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

	// Check of any of thr resolution ids are already processed
	finalEvents := make([]*types.VotableEvent, 0)
	for _, event := range events {
		id := event.ID()
		processed, err := r.VoteStore.IsProcessed(ctx, id)
		if err != nil {
			return nil, err
		}
		if !processed {
			finalEvents = append(finalEvents, event)
		} else {
			// Event already processed, delete it from the event store
			err = r.EventStore.DeleteEvent(ctx, id)
			if err != nil {
				return nil, err
			}
		}
	}

	// TODO: Should we check if any of these events are not already processed??
	tx, err := transactions.CreateTransaction(&transactions.ValidatorVoteBodies{
		Events: finalEvents,
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
	amt, err := pricer.Price(ctx.Ctx, r, tx)
	if err != nil {
		return nil, transactions.CodeUnknownError, err
	}

	if amt.Cmp(tx.Body.Fee) < 0 {
		return nil, transactions.CodeInsufficientFee, fmt.Errorf("transaction fee is too low: %s", amt.String())
	}

	// check if the transaction consented to spending enough tokens
	if tx.Body.Fee.Cmp(amt) < 0 {
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
