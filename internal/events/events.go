// package events is used to track events that need to be included in a Kwil block.
// It contains an event store that is outside of consensus.  The event store's primary
// purpose is to store events that are heard from other chains, and delete them once the
// node can verify that their event vote has been included in a block.
package events

import (
	"context"
	"fmt"
	"sync"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
)

// Datastore is a dependency required by the event store to store data.
type Datastore interface {
	Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error)
	Query(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error)

	// Set sets a key to a value.
	Set(ctx context.Context, kvTableName string, key []byte, value []byte) error
	// Get gets a value for a key.
	Get(ctx context.Context, kvTableName string, key []byte, sync bool) ([]byte, error)
	// Delete deletes a key. (Add when we need it)
	// Delete(ctx context.Context, key []byte) error
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
	mu sync.Mutex
}

func NewEventStore(ctx context.Context, db Datastore) (*EventStore, error) {
	if _, err := db.Execute(ctx, sqlCreateSchema); err != nil {
		return nil, err
	}
	if _, err := db.Execute(ctx, eventsTable); err != nil {
		return nil, err
	}
	err := pg.CreateKVTable(ctx, kvTableName, pg.WrapQueryFun(db.Execute))
	if err != nil {
		return nil, err
	}

	return &EventStore{
		db: db,
	}, nil
}

// kvDB emulates a sql.KV for consumers of the KV method, such as the oracle.
type kvDB struct{ *EventStore }

func (d *kvDB) Get(ctx context.Context, key []byte, sync bool) ([]byte, error) {
	d.EventStore.mu.Lock()
	defer d.EventStore.mu.Unlock()
	return d.EventStore.db.Get(ctx, kvTableName, key, sync)
}

func (d *kvDB) Set(ctx context.Context, key, value []byte) error {
	d.EventStore.mu.Lock()
	defer d.EventStore.mu.Unlock()
	return d.EventStore.db.Set(ctx, kvTableName, key, value)
}

// KV returns a kv store that is scoped to the given prefix.
// It allows the user to define their own semantics
// for tracking committed data. For example, it can be used to
// track the latest block number of some other chain.
// This allows users to implement complex logic for efficient
// restart, to avoid re-processing events. Key uniqueness is
// scoped to the event type.
// It is up to each oracle to define their own sufficiently unique prefix(es).
func (e *EventStore) KV(prefix []byte) *scopedKVStore {
	return &scopedKVStore{
		prefix: prefix,
		kv:     &kvDB{e},
	}
}

// Store stores an event in the event store.
// It is idempotent.
func (e *EventStore) Store(ctx context.Context, data []byte, eventType string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	event := &types.VotableEvent{
		Body: data,
		Type: eventType,
	}

	id := event.ID()

	_, err := e.db.Execute(ctx, insertEventIdempotent, id[:], data, eventType)
	return err
}

// GetEvents gets all events in the event store.
func (e *EventStore) GetEvents(ctx context.Context) ([]*types.VotableEvent, error) {
	res, err := e.db.Query(ctx, getEvents)
	if err != nil {
		return nil, err
	}

	var events []*types.VotableEvent
	if len(res.Columns) != 2 {
		return nil, fmt.Errorf("expected 2 columns getting events. this is an internal bug")
	}
	for _, row := range res.Rows {
		// rows[0] is the raw data of the event
		// rows[1] is the event type
		data, ok := row[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("expected data to be []byte, got %T", row[0])
		}
		eventType, ok := row[1].(string)
		if !ok {
			return nil, fmt.Errorf("expected event type to be string, got %T", row[1])
		}

		events = append(events, &types.VotableEvent{
			Body: data,
			Type: eventType,
		})
	}

	return events, nil
}

// DeleteEvent deletes an event from the event store.
// It is idempotent. If the event does not exist, it will not return an error.
func (e *EventStore) DeleteEvent(ctx context.Context, id types.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.db.Execute(ctx, deleteEvent, id[:])
	return err
}

// GetUnreceivedEvents retrieves events that are neither received by the network nor previously broadcasted.
// An event is considered "received" only after its inclusion in a block.
// The function excludes events that have been broadcasted but are still pending in the mempool, awaiting block inclusion.
func (e *EventStore) GetUnreceivedEvents(ctx context.Context) ([]*types.VotableEvent, error) {
	res, err := e.db.Query(ctx, getUnbroadcastedEvents)
	if err != nil {
		return nil, err
	}

	var events []*types.VotableEvent
	if len(res.Columns) != 2 {
		return nil, fmt.Errorf("expected 2 columns getting events. this is an internal bug")
	}
	for _, row := range res.Rows {
		// row[0] is the raw data of the event
		// row[1] is the event type

		data, ok := row[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("expected data to be []byte, got %T", row[0])
		}
		eventType, ok := row[1].(string)
		if !ok {
			return nil, fmt.Errorf("expected event type to be string, got %T", row[1])
		}

		events = append(events, &types.VotableEvent{
			Body: data,
			Type: eventType,
		})
	}

	return events, nil
}

// MarkReceived marks that an event has been received by the network, and should not be re-broadcasted.
func (e *EventStore) MarkReceived(ctx context.Context, id types.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.db.Execute(ctx, markReceived, id[:])
	return err
}

// MarkBroadcasted marks the event as broadcasted.
func (e *EventStore) MarkBroadcasted(ctx context.Context, ids []types.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.db.Execute(ctx, markBroadcasted, types.UUIDArray(ids))
	return err
}

// MarkRebroadcast marks the event to be rebroadcasted. Usually in scenarios where
// the transaction was rejected by mempool due to invalid nonces.
func (e *EventStore) MarkRebroadcast(ctx context.Context, ids []types.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.db.Execute(ctx, markRebroadcast, types.UUIDArray(ids))
	return err
}

// scopedKVStore is a KVStore that is scoped to an event type.
type scopedKVStore struct {
	prefix []byte
	kv     *kvDB
}

func (s *scopedKVStore) Get(ctx context.Context, key []byte, sync bool) ([]byte, error) {
	return s.kv.Get(ctx, append(s.prefix, key...), sync)
}

func (s *scopedKVStore) Set(ctx context.Context, key []byte, value []byte) error {
	return s.kv.Set(ctx, append(s.prefix, key...), value)
}

// func (s *scopedKVStore) Delete(ctx context.Context, key []byte) error {
// 	return s.store.Delete(ctx, append(s.prefix, key...))
// }
