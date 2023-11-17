package tokenbridge

import (
	"context"
	"fmt"
	"math/big"
	"strings"
)

const (
	/*
		DepositEvent: {
			eventID
			amount
			sender
		}
		Observer
	*/
	sqlInitTables = `
		-- deposit_events is used to keep track of all the deposit events that have been observed.
		-- Amount is used to indicate the amount of tokens that were deposited in the event.
		-- Sender is used to indicate the sender of the deposit event.
		-- ID is the unique identifier for the event: Hash("deposits", amount, sender, txHash, BlockHash, ChainID).
		CREATE TABLE IF NOT EXISTS deposit_events (
			EventID TEXT PRIMARY KEY,
			Amount TEXT NOT NULL,
			Sender TEXT NOT NULL
		) WITHOUT ROWID, STRICT;

		-- deposit_event_attesters is used to keep track of the validators that have attested to the deposit event.
		-- Broadcast is used to keep track of whether the event need to be broadcasted to the other validators
		CREATE TABLE IF NOT EXISTS deposit_event_attesters (
			EventID TEXT REFERENCES deposit_events(EventID) ON DELETE CASCADE,
			Observer TEXT NOT NULL,
			Broadcast INTEGER NOT NULL,
			PRIMARY KEY (EventID, Observer)
		) WITHOUT ROWID, STRICT;
		
		-- last_synced_height is used to keep track of the last height that was processed
		-- There should only be one row in this table identified by the ID "height"
		-- using ID just for simplicity of the code
		CREATE TABLE IF NOT EXISTS last_synced_height (
			ID TEXT PRIMARY KEY,
			HEIGHT INTEGER NOT NULL
		);`

	sqlAddDepositEvent = `INSERT INTO deposit_events (EventID, Amount, Sender) VALUES ($EventID, $Amount, $Sender) ON CONFLICT DO NOTHING;`

	sqlAddDepositEventAttester = `INSERT INTO deposit_event_attesters (EventID, Observer, Broadcast) VALUES ($EventID, $Observer, $Broadcast) ON CONFLICT  DO UPDATE SET Broadcast = $Broadcast;`

	sqlDepositEventsToBroadcast = `SELECT EventID FROM deposit_event_attesters WHERE Broadcast = 1 AND Observer = $Observer;`

	sqlGetDepositEvent = `SELECT Amount, Sender FROM deposit_events WHERE EventID = $EventID;`

	sqlGetLastHeight = `SELECT HEIGHT FROM last_synced_height WHERE ID = $ID;`

	sqlSetLastHeight = `INSERT INTO last_synced_height (ID, HEIGHT) VALUES ($ID, $HEIGHT) ON CONFLICT  DO UPDATE SET HEIGHT = $HEIGHT;`

	sqlRemoveObserver = `DELETE FROM deposit_event_attesters WHERE Observer = $Observer;`

	sqlRemoveEvent = `DELETE FROM deposit_events WHERE EventID = $EventID;`

	sqlGetDepositEvents = `SELECT EventID, Amount, Sender FROM deposit_events;`

	sqlGetEventObservers = `SELECT Observer FROM deposit_event_attesters WHERE EventID = $EventID;`
)

func getTableInits() []string {
	inits := strings.Split(sqlInitTables, ";")
	return inits[:len(inits)-1]
}

func (ds *DepositStore) initTables(ctx context.Context) error {
	inits := getTableInits()
	for _, init := range inits {
		err := ds.db.Execute(ctx, init, nil)
		if err != nil {
			return fmt.Errorf("failed to initialize tables: %w", err)
		}
	}

	// Get the last height
	_, err := ds.getLastHeight(ctx)
	if err != nil {
		// Height doesn't exist
		if err := ds.setLastHeight(ctx, 0); err != nil {
			return fmt.Errorf("failed to set last height: %w", err)
		}
	}
	return nil
}

func (ds *DepositStore) addDepositEvent(ctx context.Context, eventId string, sender string, amount *big.Int) error {
	if err := ds.db.Execute(ctx, sqlAddDepositEvent, map[string]interface{}{
		"$EventID": eventId,
		"$Amount":  amount.String(),
		"$Sender":  sender,
	}); err != nil {
		return fmt.Errorf("failed to add deposit event: %w", err)
	}
	return nil
}

func (ds *DepositStore) addDepositEventAttester(ctx context.Context, eventId string, observer string) error {
	if err := ds.db.Execute(ctx, sqlAddDepositEventAttester, map[string]interface{}{
		"$EventID":   eventId,
		"$Observer":  observer,
		"$Broadcast": 1,
	}); err != nil {
		return fmt.Errorf("failed to add deposit event attester: %w", err)
	}
	return nil
}
func (ds *DepositStore) getDepositEvent(ctx context.Context, eventId string) (amount, sender string, err error) {
	results, err := ds.db.Query(ctx, sqlGetDepositEvent, map[string]any{
		"$EventID": eventId,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to get deposit event: %w", err)
	}

	if len(results) == 0 {
		return "", "", fmt.Errorf("no deposit event found")
	}

	amount, ok := results[0]["Amount"].(string)
	if !ok {
		return "", "", fmt.Errorf("failed to convert amount to string")
	}

	sender, ok = results[0]["Sender"].(string)
	if !ok {
		return "", "", fmt.Errorf("failed to convert sender to string")
	}

	return amount, sender, nil
}

func (ds *DepositStore) getDepositEventsToBroadcast(ctx context.Context, observer string) ([]string, error) {
	results, err := ds.db.Query(ctx, sqlDepositEventsToBroadcast, map[string]any{
		"$Observer": observer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get deposit events to broadcast: %w", err)
	}

	eventIds := make([]string, len(results))
	for i, result := range results {
		eventId, ok := result["EventID"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert event id to string")
		}
		eventIds[i] = eventId

		err := ds.markDepositEventAsBroadcasted(ctx, eventIds[0], observer)
		if err != nil {
			return nil, fmt.Errorf("failed to mark deposit event as broadcasted: %w", err)
		}
	}

	return eventIds, nil
}

func (ds *DepositStore) markDepositEventAsBroadcasted(ctx context.Context, eventId string, observer string) error {
	_, err := ds.db.Query(ctx, sqlDepositEventsToBroadcast, map[string]any{
		"$EventID":   eventId,
		"$Observer":  observer,
		"$Broadcast": 0,
	})
	if err != nil {
		return fmt.Errorf("failed to get deposit events to broadcast: %w", err)
	}
	return nil
}

func (ds *DepositStore) purgeObserverEvents(ctx context.Context, observer string) error {
	if err := ds.db.Execute(ctx, sqlRemoveObserver, map[string]interface{}{
		"$Observer": observer,
	}); err != nil {
		return fmt.Errorf("failed to purge observer events: %w", err)
	}
	return nil
}

func (ds *DepositStore) purgeEvent(ctx context.Context, eventId string) error {
	if err := ds.db.Execute(ctx, sqlRemoveEvent, map[string]interface{}{
		"$EventID": eventId,
	}); err != nil {
		return fmt.Errorf("failed to purge event: %s due to error: %w", eventId, err)
	}
	return nil
}

func (ds *DepositStore) setLastHeight(ctx context.Context, height int64) error {
	if err := ds.db.Execute(ctx, sqlSetLastHeight, map[string]interface{}{
		"$ID":     "height",
		"$HEIGHT": height,
	}); err != nil {
		return fmt.Errorf("failed to set last height: %w", err)
	}
	return nil
}

func (ds *DepositStore) getEventObservers(ctx context.Context, eventId string) ([]string, error) {
	results, err := ds.db.Query(ctx, sqlGetEventObservers, map[string]any{
		"$EventID": eventId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get event observers: %w", err)
	}

	observers := make([]string, len(results))
	for i, result := range results {
		observer, ok := result["Observer"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert observer to string")
		}
		observers[i] = observer
	}

	return observers, nil
}

// TODO: is int64 the right type for height? change it to big.Int?
func (ds *DepositStore) getLastHeight(ctx context.Context) (int64, error) {
	results, err := ds.db.Query(ctx, sqlGetLastHeight, map[string]any{
		"$ID": "height",
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get last height: %w", err)
	}

	if len(results) == 0 {
		return 0, fmt.Errorf("no last height found")
	}

	resHeight, ok := results[0]["HEIGHT"]
	if !ok {
		return 0, fmt.Errorf("failed to get last height from result")
	}
	height, ok := resHeight.(int64)
	if !ok {
		return 0, fmt.Errorf("failed to convert last height to int64")
	}

	return height, nil
}

func (ds *DepositStore) depositEvents(ctx context.Context) ([]*DepositEvent, error) {
	results, err := ds.db.Query(ctx, sqlGetDepositEvents, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get deposit events: %w", err)
	}

	depositEvents := make([]*DepositEvent, len(results))
	for i, result := range results {
		eventId, ok := result["EventID"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert event id to string")
		}

		amount, ok := result["Amount"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert amount to string")
		}

		sender, ok := result["Sender"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert sender to string")
		}

		attesters, err := ds.getEventObservers(ctx, eventId)
		if err != nil {
			return nil, fmt.Errorf("failed to get event observers: %w", err)
		}

		depositEvents[i] = &DepositEvent{
			EventID:      eventId,
			Amount:       amount,
			Sender:       sender,
			Attestations: attesters,
		}
	}

	return depositEvents, nil
}
