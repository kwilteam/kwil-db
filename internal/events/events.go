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

// VoteStore is a store that tracks votes.
type VoteStore interface {
	// IsProcessed checks if a resolution has been processed.
	IsProcessed(ctx context.Context, db sql.DB, resolutionID types.UUID) (bool, error)
}

// EventStore stores events from external sources.
// Kwil uses the event store to track received events, and ensure that they are
// broadcasted to the network.
// It follows an at-least-once semantic, and so each event body should be unique.
// Events can be added idempotently; calling StoreEvent for an event that has already
// been stored or processed will incur some computational overhead, but will not
// cause any issues.
type EventStore struct {
	//eventWriterMu is a mutex to ensure that the eventWriter is not used concurrently.
	// Since the kv uses the autocommit mode, we need to ensure that this does not conflict
	// with transactions, since these cannot be used concurrently.
	eventWriterMu sync.Mutex

	// eventWriter is a database used for writing events.
	// It holds a separate connection to the database, since
	// events are written outside of the consensus process.
	// Events are deleted during consensus and need to be atomic
	// with consensus, so these two cannot be managed with the same
	// connection.
	eventWriter *pg.DB

	// voteStore is a store that tracks votes.
	votestore VoteStore
}

// NewEventStore creates a new event store.
// It takes a database connection to write events to.
// WARNING: This connection cannot be the same connection
// used during consensus / in txapp.
func NewEventStore(ctx context.Context, writerDB *pg.DB, voteStore VoteStore) (*EventStore, error) {
	tx, err := writerDB.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Execute(ctx, sqlCreateSchema); err != nil {
		return nil, err
	}
	if _, err := tx.Execute(ctx, eventsTable); err != nil {
		return nil, err
	}
	if _, err := tx.Execute(ctx, createKvTblStmt); err != nil {
		return nil, err
	}

	_, err = tx.Precommit(ctx) // outermost txs need to be precommitted
	if err != nil {
		return nil, err
	}

	return &EventStore{
		eventWriter: writerDB,
		votestore:   voteStore,
	}, tx.Commit(ctx)
}

// Store stores an event in the event store.
// It uses the local connection to the event store,
// instead of the consensus connection.
func (e *EventStore) Store(ctx context.Context, data []byte, eventType string) error {
	// tx guarantees that we are reading the same state of the database for both votestore and eventstore
	e.eventWriterMu.Lock()
	defer e.eventWriterMu.Unlock()

	tx, err := e.eventWriter.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	event := &types.VotableEvent{
		Body: data,
		Type: eventType,
	}
	id := event.ID()

	// is this event already processed?
	processed, err := e.votestore.IsProcessed(ctx, tx, id)
	if err != nil {
		return err
	}
	if processed {
		return nil
	}

	_, err = tx.Execute(ctx, insertEventIdempotent, id[:], data, eventType)
	if err != nil {
		return err
	}

	// we need to precommit then commit, since this is an outermost tx
	_, err = tx.Precommit(ctx)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetEvents gets all events in the event store.
func (e *EventStore) GetEvents(ctx context.Context, db sql.DB) ([]*types.VotableEvent, error) {
	res, err := db.Execute(ctx, getEvents)
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
func (e *EventStore) DeleteEvent(ctx context.Context, db sql.DB, id types.UUID) error {
	_, err := db.Execute(ctx, deleteEvent, id[:])
	return err
}

// GetUnreceivedEvents gets all events that have not been received by the network in a block.
// Events are only marked as "received" when they have been included in a block, by this validator.
// It uses the local connection to the event store, instead of the consensus connection.
func (e *EventStore) GetUnreceivedEvents(ctx context.Context) ([]*types.VotableEvent, error) {
	readTx, err := e.eventWriter.BeginReadTx(ctx)
	if err != nil {
		return nil, err
	}
	defer readTx.Rollback(ctx) // only reading, so we can always rollback

	res, err := readTx.Execute(ctx, getUnbroadcastedEvents)
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
func (e *EventStore) MarkReceived(ctx context.Context, db sql.DB, id types.UUID) error {
	_, err := db.Execute(ctx, markReceived, id[:])
	return err
}

// KV returns a kv store that is scoped to the given prefix.
// It allows the user to define their own semantics
// for tracking committed data. For example, it can be used to
// track the latest block number of some other chain.
// This allows users to implement complex logic for efficient
// restart, to avoid re-processing events. Key uniqueness is
// scoped to the event type.
// It is up to each oracle to define their own sufficiently unique prefix(es).
func (e *EventStore) KV(prefix []byte) *KV {
	return &KV{
		prefix: prefix,
		es:     e,
	}
}

// KV is a KVStore that is scoped to an event type.
type KV struct {
	prefix []byte
	es     *EventStore
}

func (s *KV) Get(ctx context.Context, key []byte) ([]byte, error) {
	s.es.eventWriterMu.Lock()
	defer s.es.eventWriterMu.Unlock()

	s.es.eventWriter.AutoCommit(true)
	defer s.es.eventWriter.AutoCommit(false)

	res, err := s.es.eventWriter.Execute(ctx, selectKvStmt, append(s.prefix, key...))
	if err != nil {
		return nil, err
	}

	if len(res.Rows) == 0 {
		return nil, nil
	}

	if len(res.Rows) > 1 {
		return nil, fmt.Errorf("expected 1 row, got %d", len(res.Rows))
	}

	data, ok := res.Rows[0][0].([]byte)
	if !ok {
		return nil, fmt.Errorf("expected data to be []byte, got %T", res.Rows[0][0])
	}

	return data, nil
}

func (s *KV) Set(ctx context.Context, key []byte, value []byte) error {
	s.es.eventWriterMu.Lock()
	defer s.es.eventWriterMu.Unlock()

	s.es.eventWriter.AutoCommit(true)
	defer s.es.eventWriter.AutoCommit(false)

	_, err := s.es.eventWriter.Execute(ctx, upsertKvStmt, append(s.prefix, key...), value)
	return err
}

func (s *KV) Delete(ctx context.Context, key []byte) error {
	s.es.eventWriterMu.Lock()
	defer s.es.eventWriterMu.Unlock()

	s.es.eventWriter.AutoCommit(true)
	defer s.es.eventWriter.AutoCommit(false)

	_, err := s.es.eventWriter.Execute(ctx, deleteKvStmt, append(s.prefix, key...))
	return err
}
