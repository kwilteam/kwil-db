package events

import (
	"context"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/sql"
)

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
	if processed {
		return nil
	}

	// store the event only if not already processed
	return e.eventstore.Store(ctx, body, eventType)
}

func (e *EventMgr) KV(prefix []byte) sql.KVStore {
	return e.eventstore.KV(prefix)
}
