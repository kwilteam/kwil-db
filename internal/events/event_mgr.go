package events

import (
	"context"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/sql"
)

// EventMgr is a wrapper around the EventStore and VoteStore
// It provides a single interface to store events and
// check if an event is already processed.
// This is to ensure that a node that's catching up doesn't
// broadcast events that are already processed and overwhelm
// the mempool.
type EventMgr struct {
	eventstore *EventStore
	votestore  VoteStore
}

type VoteStore interface {
	IsProcessed(ctx context.Context, resolutionID types.UUID) (bool, error)
}

func NewEventMgr(eventstore *EventStore, votestore VoteStore) *EventMgr {
	return &EventMgr{
		eventstore: eventstore,
		votestore:  votestore,
	}
}

// Store stores an event if it is not already processed.
func (e *EventMgr) Store(ctx context.Context, body []byte, eventType string) error {
	event := &types.VotableEvent{
		Body: body,
		Type: eventType,
	}
	id := event.ID()
	// is this event already processed?
	processed, err := e.votestore.IsProcessed(ctx, id)
	if err != nil {
		return err
	}

	// store the event only if not already processed
	if processed {
		return nil
	}
	return e.eventstore.Store(ctx, body, eventType)
}

func (e *EventMgr) GetEvents(ctx context.Context) ([]*types.VotableEvent, error) {
	return e.eventstore.GetEvents(ctx)
}

// KV returns a KV store
func (e *EventMgr) KV(prefix []byte) sql.KV {
	return e.eventstore.KV(prefix)
}
