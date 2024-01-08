// package broadcast contains logic for broadcasting events to the Kwil network
package broadcast

/*
	I'm a bit torn on this package; it is very much a micro-package. Its only purpose
	is to implement the abci.CommitHook function signature, and broadcast events to the network.

	It seems like it could be
	in the events package, however this then creates a circular dependency, since
	- txapp needs events
	- abci needs txapp
	- cometbft node needs abci
	- cometbft client needs cometbft node
	- this package needs cometbft client

	This package is also a coupling point between abci and the event store
*/

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/abci"
)

// EventStore allows the EventBroadcaster to read events
// from the event store.
type EventStore interface {
	// GetUnbroadcastedEvents gets events that this node has not yet broadcasted.
	// Events are only marked as "broadcasted" when they have been included in a block.
	GetUnreceivedEvents(ctx context.Context) ([]*types.VotableEvent, error)
}

// Broadcaster is an interface for broadcasting to the Kwil network.
type Broadcaster interface {
	BroadcastTx(ctx context.Context, tx []byte, sync uint8) (code uint32, txHash []byte, err error)
}

// AccountInfoer gets uncommitted information about an account.
// It can be used for building transactions.
type AccountInfoer interface {
	// AccountInfo gets uncommitted information about an account.
	AccountInfo(ctx context.Context, acctID []byte, getUncommitted bool) (balance *big.Int, nonce int64, err error)
}

// ValidatorStore gets data about the local validators.
type ValidatorStore interface {
	IsCurrent(ctx context.Context, validator []byte) (bool, error)
}

func NewEventBroadcaster(store EventStore, broadcaster Broadcaster, accountInfo AccountInfoer, validatorStore ValidatorStore, signer *auth.Ed25519Signer, chainID string) *EventBroadcaster {
	return &EventBroadcaster{
		store:          store,
		broadcaster:    broadcaster,
		accountInfo:    accountInfo,
		validatorStore: validatorStore,
		signer:         signer,
		chainID:        chainID,
	}
}

// EventBroadcaster manages broadcasting events to the Kwil network.
type EventBroadcaster struct {
	store          EventStore
	broadcaster    Broadcaster
	accountInfo    AccountInfoer
	validatorStore ValidatorStore
	signer         *auth.Ed25519Signer
	chainID        string
}

// RunBroadcast tells the EventBroadcaster to broadcast any events it wishes.
// It implements Kwil's abci.CommitHook function signature.
// If the node is not a validator, it will do nothing.
func (e *EventBroadcaster) RunBroadcast(ctx context.Context, _ *abci.BlockInfo) error {
	isCurrent, err := e.validatorStore.IsCurrent(ctx, e.signer.Identity())
	if err != nil {
		return err
	}
	if !isCurrent {
		return nil
	}

	events, err := e.store.GetUnreceivedEvents(ctx)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return nil
	}

	ids := make([]types.UUID, len(events))
	for i, event := range events {
		ids[i] = event.ID()
	}

	_, nonce, err := e.accountInfo.AccountInfo(ctx, e.signer.Identity(), true)
	if err != nil {
		return err
	}

	tx, err := transactions.CreateTransaction(&transactions.ValidatorVoteIDs{ResolutionIDs: ids}, e.chainID, uint64(nonce)+1)
	if err != nil {
		return err
	}

	err = tx.Sign(e.signer)
	if err != nil {
		return err
	}

	bts, err := tx.MarshalBinary()
	if err != nil {
		return err
	}

	_, _, err = e.broadcaster.BroadcastTx(ctx, bts, 0)
	if err != nil {
		return err
	}

	fmt.Println("Broadcast events: ", ids, " with nonce: ", nonce+1, " For account: ", hex.EncodeToString(e.signer.Identity()))
	return nil
}
