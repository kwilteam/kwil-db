// package broadcast contains logic for broadcasting events to the Kwil network
package broadcast

/*
	I'm a bit torn on this package; it is very much a micro-package. Its only purpose
	is to implement the abci.CommitHook function signature, and broadcast events to the network.

	It seems like it could be
	in the voting package, however this then creates a circular dependency, since
	- txapp needs voting
	- abci needs txapp
	- cometbft node needs abci
	- cometbft client needs cometbft node
	- this package needs cometbft client

	This package is also a coupling point between abci and the event store
*/

import (
	"bytes"
	"context"
	"math/big"

	cmtCoreTypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

var (
	// Until the votestore is properly optimized, we limit the number of voteIDs that can be included in a single transaction.
	// This is to limit the long external roundtrips to the postgres database
	// 10k voteIDs in a block takes around 30s to process, which is too long.
	maxVoteIDsPerTx = 100
)

// EventStore allows the EventBroadcaster to read events
// from the event store.
type EventStore interface {
	// GetUnbroadcastedEvents filters out the events observed by the validator
	// that are not previously broadcasted.
	GetUnbroadcastedEvents(ctx context.Context) ([]types.UUID, error)

	// MarkBroadcasted marks list of events as broadcasted.
	MarkBroadcasted(ctx context.Context, ids []types.UUID) error
}

// Broadcaster is an interface for broadcasting to the Kwil network.
type Broadcaster interface {
	BroadcastTx(ctx context.Context, tx []byte, sync uint8) (res *cmtCoreTypes.ResultBroadcastTx, err error)
}

// TxApp is the main Kwil application.
type TxApp interface {
	// AccountInfo gets uncommitted information about an account.
	AccountInfo(ctx context.Context, acctID []byte, getUncommitted bool) (balance *big.Int, nonce int64, err error)
	// Price gets the estimated fee for a transaction.
	Price(ctx context.Context, tx *transactions.Transaction) (*big.Int, error)
}

// ValidatorStore gets data about the local validators.
type ValidatorStore interface {
	GetValidators(ctx context.Context) ([]*types.Validator, error)
}

func NewEventBroadcaster(store EventStore, broadcaster Broadcaster, app TxApp, validatorStore ValidatorStore, signer *auth.Ed25519Signer, chainID string) *EventBroadcaster {
	return &EventBroadcaster{
		store:          store,
		broadcaster:    broadcaster,
		validatorStore: validatorStore,
		signer:         signer,
		chainID:        chainID,
		app:            app,
	}
}

// EventBroadcaster manages broadcasting events to the Kwil network.
type EventBroadcaster struct {
	store          EventStore
	broadcaster    Broadcaster
	validatorStore ValidatorStore
	signer         *auth.Ed25519Signer
	chainID        string
	app            TxApp
}

// RunBroadcast tells the EventBroadcaster to broadcast any events it wishes.
// It implements Kwil's abci.CommitHook function signature.
// If the node is not a validator, it will do nothing.
// It broadcasts votes for the existing resolutions.
func (e *EventBroadcaster) RunBroadcast(ctx context.Context, Proposer []byte) error {
	validators, err := e.validatorStore.GetValidators(ctx)
	if err != nil {
		return err
	}

	var isCurrent bool
	for _, v := range validators {
		if bytes.Equal(v.PubKey, e.signer.Identity()) {
			isCurrent = true
			break
		}
	}

	if !isCurrent {
		return nil
	}

	// Proposers are not allowed to broadcast voteID transactions.
	// This is to avoid complexities around the nonce tracking arising from
	// proposer introducing voteBody transactions in the on-going block.
	// As the nonces for proposer induced transactions are currently not tracked
	// by the mempool, it is not safe to introduce new transactions by proposer
	// in the on-going block. This probably is a temporary restriction until
	// we figure out a better way to track both
	// mempool(uncommitted), committed and proposer introduced txns.
	if bytes.Equal(Proposer, e.signer.Identity()) {
		return nil
	}

	// Vote only if the voter observed the event corresponding to the resolution.
	// ids are the resolution ids that the validator witnessed the events for and can vote on.
	ids, err := e.store.GetUnbroadcastedEvents(ctx)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	// consider only the first maxVoteIDsPerTx events, to limit the postgres access roundtrips per block execution.
	if len(ids) > maxVoteIDsPerTx {
		ids = ids[:maxVoteIDsPerTx]
	}

	bal, nonce, err := e.app.AccountInfo(ctx, e.signer.Identity(), true)
	if err != nil {
		return err
	}

	tx, err := transactions.CreateTransaction(&transactions.ValidatorVoteIDs{ResolutionIDs: ids}, e.chainID, uint64(nonce)+1)
	if err != nil {
		return err
	}

	// Get the fee estimate
	fee, err := e.app.Price(ctx, tx)
	if err != nil {
		return err
	}

	tx.Body.Fee = fee

	if bal.Cmp(fee) < 0 {
		// Not enough balance to pay for the tx fee
		return nil
	}

	err = tx.Sign(e.signer)
	if err != nil {
		return err
	}

	bts, err := tx.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = e.broadcaster.BroadcastTx(ctx, bts, 0)
	if err != nil {
		return err
	}

	// mark these events as broadcasted
	err = e.store.MarkBroadcasted(ctx, ids)
	if err != nil {
		return err
	}

	return nil
}
