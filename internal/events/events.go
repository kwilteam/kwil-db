// package events is used to track events that need to be included in a Kwil block.
// It contains an event store that is outside of consensus.  The event store's primary
// purpose is to store events that are heard from other chains, and delete them once the
// node can verify that their event vote has been included in a block.
package events

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/sql"
)

// Datastore is a dependency required by the event store to store data.
type Datastore interface {
	Execute(ctx context.Context, stmt string, args map[string]any) (*sql.ResultSet, error)
	Query(ctx context.Context, stmt string, args map[string]any) (*sql.ResultSet, error)

	// Set sets a key to a value.
	Set(ctx context.Context, key []byte, value []byte) error
	// Get gets a value for a key.
	Get(ctx context.Context, key []byte, sync bool) ([]byte, error)
	// Delete deletes a key.
	Delete(ctx context.Context, key []byte) error
}

// EventStore stores events from external sources.
// Kwil uses the event store to track received events, and ensure that they are
// broadcasted to the network.
// It follows an at-least-once semantic, and so each event body should be unique.
// Events can be added idempotently; calling StoreEvent for an event that has already
// been stored or processed will incur some computational overhead, but will not
// cause any issues.
type EventStore struct {
	db Datastore
}

func NewEventStore(ctx context.Context, db Datastore) (*EventStore, error) {
	_, err := db.Execute(ctx, eventsTable, nil)
	if err != nil {
		return nil, err
	}

	return &EventStore{
		db: db,
	}, nil
}

// KV returns a KVStore that is scoped to the given prefix.
// It allows the user to define their own semantics
// for tracking committed data. For example, it can be used to
// track the latest block number of some other chain.
// This allows users to implement complex logic for efficient
// restart, to avoid re-processing events. Key uniqueness is
// scoped to the event type.
// It is up to each oracle to define their own sufficiently unique prefix(es).
func (e *EventStore) KV(prefix []byte) sql.KVStore {
	return &scopedKVStore{
		prefix: prefix,
		store:  e.db,
	}
}

// Store stores an event in the event store.
// It is idempotent.
func (e *EventStore) Store(ctx context.Context, data []byte, eventType string) error {
	event := &types.VotableEvent{
		Body: data,
		Type: eventType,
	}

	id := event.ID()

	_, err := e.db.Execute(ctx, insertEventIdempotent, map[string]any{
		"$id":         id[:],
		"$data":       data,
		"$event_type": eventType,
	})
	return err
}

// GetEvents gets all events in the event store.
func (e *EventStore) GetEvents(ctx context.Context) ([]*types.VotableEvent, error) {
	res, err := e.db.Query(ctx, getEvents, nil)
	if err != nil {
		return nil, err
	}

	var events []*types.VotableEvent
	if len(res.Columns()) != 2 {
		return nil, fmt.Errorf("expected 2 columns getting events. this is an internal bug")
	}
	for res.Next() {
		// res.Rows[0] is the raw data of the event
		// res.Rows[1] is the event type
		values, err := res.Values()
		if err != nil {
			return nil, err
		}

		data, ok := values[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("expected data to be []byte, got %T", values[0])
		}
		eventType, ok := values[1].(string)
		if !ok {
			return nil, fmt.Errorf("expected event type to be string, got %T", values[1])
		}

		events = append(events, &types.VotableEvent{
			Body: data,
			Type: eventType,
		})
	}
	if err := res.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// DeleteEvent deletes an event from the event store.
// It is idempotent. If the event does not exist, it will not return an error.
func (e *EventStore) DeleteEvent(ctx context.Context, id types.UUID) error {
	_, err := e.db.Execute(ctx, deleteEvent, map[string]any{
		"$id": id[:],
	})
	return err
}

// scopedKVStore is a KVStore that is scoped to an event type.
type scopedKVStore struct {
	prefix []byte
	store  Datastore
}

func (s *scopedKVStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	return s.store.Get(ctx, append(s.prefix, key...), true)
}

func (s *scopedKVStore) Set(ctx context.Context, key []byte, value []byte) error {
	return s.store.Set(ctx, append(s.prefix, key...), value)
}

func (s *scopedKVStore) Delete(ctx context.Context, key []byte) error {
	return s.store.Delete(ctx, append(s.prefix, key...))
}
