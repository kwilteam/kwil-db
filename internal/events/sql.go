package events

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types/chain"
)

const (
	sqlInitTables = `
		-- events is used to keep track of all the events that have been heard by all the nodes
		-- Type is used to indicate the type of event (e.g. Deposit, Withdrawal, etc.)
		CREATE TABLE IF NOT EXISTS events (
			ID TEXT PRIMARY KEY,
			Type INTEGER NOT NULL, 
			Data BLOB NOT NULL
		) WITHOUT ROWID, STRICT;

		-- attesters is used to keep track of the attesters for each event
		CREATE TABLE IF NOT EXISTS attesters (
			ID TEXT REFERENCES events(ID) ON DELETE CASCADE,
			Validator BLOB NOT NULL,
			PRIMARY KEY (ID, Validator)
		) WITHOUT ROWID, STRICT;

		-- local_events is used to keep track of the events heard by the node
		-- IsBroadcasted is used to keep track of whether the event has been broadcasted to the other nodes
		CREATE TABLE IF NOT EXISTS local_events (
			ID TEXT REFERENCES events(ID) ON DELETE CASCADE,
			IsBroadcasted INTEGER NOT NULL,
			PRIMARY KEY (ID)
		) WITHOUT ROWID, STRICT;
		
		-- last_synced_height is used to keep track of the last height that was processed
		-- ID is always "height", so there is only one row in this table
		-- using ID just for simplicity of the code
		CREATE TABLE IF NOT EXISTS last_synced_height (
			ID TEXT PRIMARY KEY,
			HEIGHT INTEGER NOT NULL
		);

		-- schema_version is used to keep track of the version of the schema
		-- ID is always "version", so there is only one row in this table
		-- using ID just for simplicity of the code
		CREATE TABLE IF NOT EXISTS schema_version (
			ID TEXT PRIMARY KEY,
			VERSION INTEGER NOT NULL
		);`

	sqlAddEvent = `INSERT INTO events (ID, Type, Data) VALUES ($ID, $Type, $Data) ON CONFLICT DO NOTHING;`

	sqlAddAttester = `INSERT INTO attesters (ID, Validator) VALUES ($ID, $Validator) ON CONFLICT DO NOTHING;`

	sqlAddLocalEvent = `INSERT INTO local_events (ID, IsBroadcasted) VALUES ($ID, $IsBroadcasted) ON CONFLICT DO NOTHING;`

	sqlEventsToBroadcast = `SELECT ID FROM local_events WHERE IsBroadcasted = 0;`

	sqlGetEvent = `SELECT Type, Data FROM events WHERE ID = $ID;`

	//sqlDeleteLocalEvent = `DELETE FROM local_events WHERE ID = $ID`

	sqlSetVersion = `INSERT INTO schema_version (ID, VERSION) VALUES ($ID, $VERSION) ON CONFLICT DO UPDATE SET VERSION = $VERSION;`

	sqlGetVersion = `SELECT VERSION FROM schema_version WHERE ID = $ID;`

	sqlGetLastHeight = `SELECT HEIGHT FROM last_synced_height WHERE ID = $ID;`

	sqlSetLastHeight = `INSERT INTO last_synced_height (ID, HEIGHT) VALUES ($ID, $HEIGHT) ON CONFLICT DO UPDATE SET HEIGHT = $HEIGHT;`
)

var (
	eventStoreVersion = 1
)

func getTableInits() []string {
	inits := strings.Split(sqlInitTables, ";")
	return inits[:len(inits)-1]
}

func (ev *EventStore) initTables(ctx context.Context) error {
	inits := getTableInits()
	for _, init := range inits {
		_, err := ev.db.Execute(ctx, init, nil)
		if err != nil {
			return fmt.Errorf("failed to initialize tables: %w", err)
		}
	}

	if _, err := ev.db.Execute(ctx, sqlSetVersion, map[string]interface{}{
		"$ID":      "version",
		"$VERSION": eventStoreVersion,
	}); err != nil {
		return fmt.Errorf("failed to initialize schema version: %w", err)
	}

	// Get the last height
	_, err := ev.getLastHeight(ctx)
	if err != nil {
		// Height doesn't exist
		if err := ev.setLastHeight(ctx, 0); err != nil {
			return fmt.Errorf("failed to set last height: %w", err)
		}
	}

	_, err = ev.getVersion(ctx)
	if err != nil {
		// Version doesn't exist
		if err := ev.SetVersion(ctx, eventStoreVersion); err != nil {
			return fmt.Errorf("failed to set schema version: %w", err)
		}
	}

	return nil
}

// Operations:
// Add locally received event to the DB
// Add external event to the DB (vote extensions)
// Broadcast locally received event to other nodes

func (ev *EventStore) addLocalEvent(ctx context.Context, event *chain.Event) error {
	// Add the event to the events table, if it doesn't already exist
	if _, err := ev.db.Execute(ctx, sqlAddEvent, map[string]interface{}{
		"$ID":   event.ID,
		"$Type": event.Type,
		"$Data": event.Data,
	}); err != nil {
		return fmt.Errorf("failed to add event to events table: %w", err)
	}

	// Add the event to the local_events table, if it doesn't already exist
	if _, err := ev.db.Execute(ctx, sqlAddLocalEvent, map[string]interface{}{
		"$ID":            event.ID,
		"$IsBroadcasted": 0,
	}); err != nil {
		return fmt.Errorf("failed to add event (%s, %s) to local_events table: %w", event.ID, event.Type.String(), err)
	}

	// Add itself as an attester to the event
	if _, err := ev.db.Execute(ctx, sqlAddAttester, map[string]interface{}{
		"$ID":        event.ID,
		"$Validator": event.Observer,
	}); err != nil {
		return fmt.Errorf("failed to add attester (%s) to event (%s, %s): %w", event.Observer, event.ID, event.Type.String(), err)
	}
	return nil
}

func (ev *EventStore) addExternalEvent(ctx context.Context, event *chain.Event) error {
	// Add the event to the events table, if it doesn't already exist
	if _, err := ev.db.Execute(ctx, sqlAddEvent, map[string]interface{}{
		"$ID":   event.ID,
		"$Type": event.Type,
		"$Data": event.Data,
	}); err != nil {
		return fmt.Errorf("failed to add event to events table: %w", err)
	}

	// Add itself as an attester to the event
	if _, err := ev.db.Execute(ctx, sqlAddAttester, map[string]interface{}{
		"$ID":        event.ID,
		"$Validator": event.Observer,
	}); err != nil {
		return fmt.Errorf("failed to add attester (%s) to event (%s, %s): %w", event.Observer, event.ID, event.Type.String(), err)
	}

	return nil
}

func (ev *EventStore) eventsToBroadcast(ctx context.Context) ([]*chain.Event, error) {
	// Get all the eventID's that haven't been broadcasted yet
	results, err := ev.db.Query(ctx, sqlEventsToBroadcast, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get events to broadcast: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	events := make([]*chain.Event, len(results))
	for i, result := range results {
		resId, ok := result["$ID"]
		if !ok {
			return nil, fmt.Errorf("failed to get event ID from result")
		}
		id, ok := resId.(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert event ID to []byte")
		}

		event, err := ev.getEvent(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get event from record: %w", err)
		}
		events[i] = event
	}
	return events, nil
}

func (ev *EventStore) getEvent(ctx context.Context, eventID string) (*chain.Event, error) {
	result, err := ev.db.Query(ctx, sqlGetEvent, map[string]interface{}{
		"$ID": eventID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get event (%s) from events table: %w", eventID, err)
	}

	// Type
	resType, ok := result[0]["Type"]
	if !ok {
		return nil, fmt.Errorf("failed to get event type from result")
	}
	eventType, ok := resType.(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert event type to int64")
	}

	// Data
	resData, ok := result[0]["Data"]
	if !ok {
		return nil, fmt.Errorf("failed to get event data from result")
	}
	eventData, ok := resData.([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to convert event data to []byte")
	}

	return &chain.Event{
		ID:       eventID,
		Type:     chain.EventType(eventType),
		Data:     eventData,
		Observer: ev.address,
	}, nil
}

// Lol comeup with a new name
func (ev *EventStore) markEventAsBroadcasted(ctx context.Context, eventID string) error {
	if _, err := ev.db.Execute(ctx, sqlAddLocalEvent, map[string]interface{}{
		"$ID":            eventID,
		"$IsBroadcasted": 1,
	}); err != nil {
		return fmt.Errorf("failed to update event (%s) broadcasted status: %w", eventID, err)
	}
	return nil
}

func (ev *EventStore) markEventsAsBroadcasted(ctx context.Context, eventIDs []string) error {
	for _, eventID := range eventIDs {
		if err := ev.markEventAsBroadcasted(ctx, eventID); err != nil {
			return fmt.Errorf("failed to mark event (%s) as broadcasted: %w", eventID, err)
		}
	}
	return nil
}

func (ev *EventStore) getVersion(ctx context.Context) (int, error) {
	results, err := ev.db.Query(ctx, sqlGetVersion, map[string]interface{}{
		"$ID": "version",
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get schema version: %w", err)
	}

	if len(results) == 0 {
		return 0, fmt.Errorf("no schema version found")
	}

	resVersion, ok := results[0]["VERSION"]
	if !ok {
		return 0, fmt.Errorf("failed to get schema version from result")
	}
	version, ok := resVersion.(int64)
	if !ok {
		return 0, fmt.Errorf("failed to convert schema version to int64")
	}

	return int(version), nil
}

func (ev *EventStore) SetVersion(ctx context.Context, version int) error {
	if _, err := ev.db.Execute(ctx, sqlSetVersion, map[string]interface{}{
		"$ID":      "version",
		"$VERSION": version,
	}); err != nil {
		return fmt.Errorf("failed to set schema version: %w", err)
	}
	return nil
}

func (ev *EventStore) setLastHeight(ctx context.Context, height int64) error {
	if _, err := ev.db.Execute(ctx, sqlSetLastHeight, map[string]interface{}{
		"$ID":     "height",
		"$HEIGHT": height,
	}); err != nil {
		return fmt.Errorf("failed to set last height: %w", err)
	}
	return nil
}

// TODO: is int64 the right type for height?
func (ev *EventStore) getLastHeight(ctx context.Context) (int64, error) {
	results, err := ev.db.Query(ctx, sqlGetLastHeight, map[string]any{
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
