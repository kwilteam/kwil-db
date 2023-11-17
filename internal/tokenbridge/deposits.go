package tokenbridge

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/log"
	"go.uber.org/zap"
)

type Datastore interface {
	Execute(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error)
	Query(ctx context.Context, query string, args map[string]any) ([]map[string]any, error)
}

type DepositStore struct {
	db  Datastore
	log log.Logger
}

type DepositEvent struct {
	EventID      string
	Amount       string
	Sender       string
	Attestations []string
}

func Threshold(numValidators int) int {
	return threshold(numValidators)
}

func threshold(numValidators int) int {
	return int(intDivUp(2*int64(numValidators), 3)) // float64(valSet.Count*2) / 3.
}

func NewDepositStore(ctx context.Context, datastore Datastore, log log.Logger) (*DepositStore, error) {
	ds := &DepositStore{
		db:  datastore,
		log: log,
	}

	err := ds.initTables(ctx)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func isValidObserver(observer string, validators map[string]bool) bool {
	_, ok := validators[observer]
	return ok
}

// func (ds *DepositStore) AddValidator(validator string) error {
// 	ds.validators[validator] = true
// 	return nil
// }

// func (ds *DepositStore) RemoveValidator(validator string) error {
// 	delete(ds.validators, validator)
// 	// Should we purge all the events that the validator has attested to?
// 	err := ds.purgeObserverEvents(context.Background(), validator)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// AddDeposit adds a deposit event observed by the observer to the deposit_events table
// and adds the observer as the event attester to the deposit_event_attesters table.
// Interms of conflict where the eventID already exists in the deposit_events table,
// Should the event data be validated? Or should the event data be ignored?
func (ds *DepositStore) AddDeposit(ctx context.Context, eventID, spender, amount, observer string) error {
	// Check if the observer is a valid validator
	// TODO: Should we check if the observer is a valid validator? If a node just started, should we start the token bridge when it get validator status? or just reject the events observed by the node? or refetch the events from the starting height and add attestation to the existing events?
	// if !ds.isValidObserver(observer) {
	// 	return fmt.Errorf("deposit Event [%s] not added as observer is not a validator", eventID)
	// }

	amt, ok := big.NewInt(0).SetString(amount, 10)
	if !ok {
		return errors.New("invalid amount")
	}

	// Register the event to the database.
	if err := ds.addDepositEvent(ctx, eventID, spender, amt); err != nil {
		return fmt.Errorf("failed to add deposit event: %w", err)
	}

	// Add the observer as the event attester and set broadcast to true
	if err := ds.addDepositEventAttester(ctx, eventID, observer); err != nil {
		return fmt.Errorf("failed to add deposit event attester: %w", err)
	}
	return nil
}

// DepositsToReport returns the eventID's, amounts, and senders for all the deposit events
// that should be broadcasted to the other validators.
func (ds *DepositStore) DepositsToReport(ctx context.Context, observer string) (eventIDs, amts, accts []string, err error) {
	// Get all the eventID's that haven't been broadcasted yet
	eventIDs, err = ds.getDepositEventsToBroadcast(ctx, observer)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get deposit events to broadcast: %w", err)
	}

	// Get the amount and sender for each eventID
	for _, eventID := range eventIDs {
		// Get the amount and sender for each eventID
		amt, acct, err := ds.getDepositEvent(ctx, eventID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to get deposit event: %w", err)
		}
		amts = append(amts, amt)
		accts = append(accts, acct)
	}

	return eventIDs, amts, accts, nil
}

// Check if the deposit event is attested by threshold number of validators
func (ds *DepositStore) hasThresholdAttestations(ctx context.Context, eventID string, validators map[string]bool) (bool, error) {
	// Get all the eventID's that haven't been broadcasted yet
	attesters, err := ds.getEventObservers(ctx, eventID)
	if err != nil {
		return false, fmt.Errorf("failed to get deposit event attesters: %w", err)
	}

	// Check if the attesters are valid validators
	cnt := 0
	for _, attester := range attesters {
		if !isValidObserver(attester, validators) {
			continue
		}
		cnt++
	}
	if cnt < threshold(len(validators)) {
		return false, nil
	}
	return true, nil
}

// DepositEvents is used when preparing a block to determine which events
// should be acted upon (i.e. by creation of a governance transaction that
// credits and account's balance). This give the application enough
// information to decide if an event has sufficient attestation, and to
// create a transaction to credit the account.
func (ds *DepositStore) DepositEvents(ctx context.Context, validators map[string]bool) (map[string]*DepositEvent, error) {
	// Get all the eventID's that haven't been finalized yet
	deposits, err := ds.depositEvents(ctx)
	if err != nil {
		ds.log.Error("Failed to get deposit events", zap.Error(err))
		return nil, fmt.Errorf("failed to get deposit events: %w", err)
	}

	evts := make(map[string]*DepositEvent)
	for _, deposit := range deposits {
		valid, err := ds.hasThresholdAttestations(ctx, deposit.EventID, validators)
		if err != nil {
			ds.log.Error("Failed to check if deposit event is attested", zap.Error(err))
			return nil, fmt.Errorf("failed to check if deposit event is attested: %w", err)
		}
		if valid {
			evts[deposit.EventID] = deposit
			continue
		}

		// Are observers valid and enough?
		// cnt := 0
		// for _, observer := range observers {
		// 	if !isValidObserver(observer, validators) {
		// 		ds.log.Error("Observer is not a validator", zap.String("observer", observer))
		// 		// Remove all the attestation for the observer
		// 		// TODO: Should we remove the observer from the attestation list? or just ignore the observer? WHat if this is a node that just joined but haven't got the validator status yet? If we remove these entries from the list, it will lose access to them Once it get the validator status. So purging should happen only when the deposit event is finalized ? WIll that lead to resource exhaustion issues?
		// 		// if err := ds.purgeObserverEvents(ctx, observer); err != nil {
		// 		// 	ds.log.Error("Failed to purge observer events", zap.Error(err))
		// 		// 	return nil, fmt.Errorf("failed to purge observer events: %w", err)
		// 		// }
		// 		continue
		// 	}
		// 	cnt++
		// }
		// evts[deposit.EventID] = deposit
	}
	return evts, nil
}

// MarkDepositActuated is used to mark a deposit as applied to an account store.
// This should remove all the entries corresponding to an eventID both from the
// deposit_events table and the deposit_event_attesters table.
func (ds *DepositStore) MarkDepositActuated(ctx context.Context, eventID string) error {
	return ds.purgeEvent(ctx, eventID)
}

// InvalidateObserverEvents is used to invalidate all the events  that the observer
// has attested to. This should remove all the entries corresponding to the observer
// from the deposit_event_attesters table.
// Used when the observer is no longer a validator. So it shouldn't be considered for
// threshold calculations  .
func (ds *DepositStore) InvalidateObserverEvents(ctx context.Context, observer string) error {
	return ds.purgeObserverEvents(ctx, observer)
}

func (ds *DepositStore) LastProcessedBlock(ctx context.Context) (int64, error) {
	return ds.getLastHeight(ctx)
}

func (ds *DepositStore) SetLastProcessedBlock(ctx context.Context, height int64) error {
	return ds.setLastHeight(ctx, height)
}

// TODO: Is this something that this store should support??
func (ds *DepositStore) Address(pubKey []byte) string {
	return ""
}

// intDivUp divides two integers, rounding up, without using floating point
// conversion. This will panic if the denominator is zero, just like regular
// integer division.
func intDivUp(val, div int64) int64 {
	// https://github.com/rust-lang/rust/blob/343889b7234bf786e2bc673029467052f22fca08/library/core/src/num/uint_macros.rs#L2061
	q, rem := val/div, val%div
	if (rem > 0 && div > 0) || (rem < 0 && div < 0) {
		q++
	}
	return q
	// rumor is that this is just as good: (val + div - 1) / div
}
