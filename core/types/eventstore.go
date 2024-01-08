package types

import (
	"context"
)

type Event struct {
	// Data is the data of the event.
	Data []byte
	// EventType is the type of the event.
	EventType string
}

func (e *Event) ID() UUID {
	return NewUUIDV5(e.Data)
}

type EventStore interface {
	// Store stores an event in the event store.
	// It is idempotent.
	Store(ctx context.Context, data []byte, eventType string) error

	// GetEvents gets all events in the event store.
	GetEvents(ctx context.Context) ([]*VotableEvent, error)

	// DeleteEvent deletes an event from the event store.
	// It is idempotent. If the event does not exist, it will not return an error.
	DeleteEvent(ctx context.Context, id UUID) error

	// KV returns a KVStore that is scoped to the given prefix.
	// It allows the user to define their own semantics
	// for tracking committed data. For example, it can be used to
	// track the latest block number of some other chain.
	// This allows users to implement complex logic for efficient
	// restart, to avoid re-processing events. Key uniqueness is
	// scoped to the event type.
	// It is up to each oracle to define their own sufficiently unique prefix(es).
	KV(prefix []byte) KVStore
}
