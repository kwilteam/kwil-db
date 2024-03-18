// package events is used to track events that need to be included in a Kwil block.
// It contains an event store that is outside of consensus.  The event store's primary
// purpose is to store events that are heard from other chains, and delete them once the
// node can verify that their event vote has been included in a block.
package events

import (
	"context"
	"fmt"
	"sync"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/internal/sql/versioning"
	"github.com/kwilteam/kwil-db/internal/voting"
)

// DB is a database connection.
type DB interface {
	sql.ReadTxMaker
	sql.DB
}

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
	// eventWriter is a database used for writing events.
	// It holds a separate connection to the database, since
	// events are written outside of the consensus process.
	// Events are deleted during consensus and need to be atomic
	// with consensus, so these two cannot be managed with the same
	// connection.
	eventWriter DB

	writerMtx sync.Mutex // protects eventWriter, not applicable to read-only operations
}

// NewEventStore creates a new event store.
// It takes a database connection to write events to.
// WARNING: This connection cannot be the same connection
// used during consensus / in txapp.
func NewEventStore(ctx context.Context, writerDB DB) (*EventStore, error) {
	tx, err := writerDB.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	upgradeFns := map[int64]versioning.UpgradeFunc{
		0: initTables,
		1: upgradeV0ToV1,
	}

	err = versioning.Upgrade(ctx, tx, schemaName, upgradeFns, eventStoreVersion)
	if err != nil {
		return nil, err
	}

	return &EventStore{
		eventWriter: writerDB,
	}, tx.Commit(ctx)
}

// Store stores an event in the event store.
// It uses the local connection to the event store,
// instead of the consensus connection.
// It only stores unprocessed events. If an event is already processed, it's ignored.
func (e *EventStore) Store(ctx context.Context, data []byte, eventType string) error {
	e.writerMtx.Lock()
	defer e.writerMtx.Unlock()

	_, err := resolutions.GetResolution(eventType) // check if the event type is valid
	if err != nil {
		return err
	}

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
	processed, err := voting.IsProcessed(ctx, tx, id)
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

	return tx.Commit(ctx)
}

// GetEvents gets all events in the event store to which resolutions have not yet been created.
func GetEvents(ctx context.Context, db sql.Executor) ([]*types.VotableEvent, error) {
	res, err := db.Execute(ctx, getNewEvents)
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
func DeleteEvent(ctx context.Context, db sql.Executor, id types.UUID) error {
	_, err := db.Execute(ctx, deleteEvent, id[:])
	return err
}

// DeleteEvents deletes a list of events from the event store.
// It is idempotent. If the event does not exist, it will not return an error.
func DeleteEvents(ctx context.Context, db sql.DB, ids ...types.UUID) error {
	_, err := db.Execute(ctx, deleteEvents, types.UUIDArray(ids))
	return err
}

// GetObservedEvents filters out the events observed by the validator that are not previously broadcasted.
func (e *EventStore) GetObservedEvents(ctx context.Context, observedIDs []types.UUID) ([]types.UUID, error) {
	readTx, err := e.eventWriter.BeginReadTx(ctx)
	if err != nil {
		return nil, err
	}
	defer readTx.Rollback(ctx) // only reading, so we can always rollback

	res, err := readTx.Execute(ctx, filterObservedEvents, types.UUIDArray(observedIDs))
	if err != nil {
		return nil, err
	}

	var ids []types.UUID
	for _, row := range res.Rows {
		if len(row) != 1 {
			return nil, fmt.Errorf("expected 1 column, got %d", len(row))
		}

		id, ok := row[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("expected id to be types.UUID, got %T", row[0])
		}
		ids = append(ids, types.UUID(id))
	}

	return ids, nil
}

// MarkBroadcasted marks the event as broadcasted.
func (e *EventStore) MarkBroadcasted(ctx context.Context, ids []types.UUID) error {
	e.writerMtx.Lock()
	defer e.writerMtx.Unlock()

	_, err := e.eventWriter.Execute(ctx, markBroadcasted, types.UUIDArray(ids))
	return err
}

// MarkRebroadcast marks the event to be rebroadcasted. Usually in scenarios where
// the transaction was rejected by mempool due to invalid nonces.
func (e *EventStore) MarkRebroadcast(ctx context.Context, ids []types.UUID) error {
	e.writerMtx.Lock()
	defer e.writerMtx.Unlock()

	_, err := e.eventWriter.Execute(ctx, markRebroadcast, types.UUIDArray(ids))
	return err
}

// KV returns a kv store that is scoped to the given prefix.
// It allows the user to define their own semantics
// for tracking committed data. For example, it can be used to
// track the latest block number of some other chain.
// This allows users to implement complex logic for efficient
// restart, to avoid re-processing events. Key uniqueness is
// scoped to the event type.
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
	tx, err := s.es.eventWriter.BeginReadTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) // only reading, so we can always rollback

	res, err := tx.Execute(ctx, selectKvStmt, append(s.prefix, key...))
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
	s.es.writerMtx.Lock()
	defer s.es.writerMtx.Unlock()

	tx, err := s.es.eventWriter.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Execute(ctx, upsertKvStmt, append(s.prefix, key...), value)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *KV) Delete(ctx context.Context, key []byte) error {
	s.es.writerMtx.Lock()
	defer s.es.writerMtx.Unlock()

	tx, err := s.es.eventWriter.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Execute(ctx, deleteKvStmt, append(s.prefix, key...))
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
