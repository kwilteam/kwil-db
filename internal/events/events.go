package events

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types/chain"
)

type Datastore interface {
	Execute(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error)
	Query(ctx context.Context, query string, args map[string]any) ([]map[string]any, error)
}

type EventStore struct {
	address []byte // Local nodes address.
	db      Datastore
	log     log.Logger
}

func NewEventStore(ctx context.Context, datastore Datastore, address []byte, log log.Logger) (*EventStore, error) {
	es := &EventStore{
		address: address,
		db:      datastore,
		log:     log,
	}

	err := es.initTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database at version %d due to error: %w", eventStoreVersion, err)
	}

	return es, nil
}

func (ev *EventStore) AddLocalEvent(ctx context.Context, event *chain.Event) error {
	event.Observer = ev.address
	return ev.addLocalEvent(ctx, event)
}

func (ev *EventStore) AddExternalEvent(ctx context.Context, event *chain.Event) error {
	event.Observer = []byte("external")
	return ev.addExternalEvent(ctx, event)
}

func (ev *EventStore) LastProcessedBlock(ctx context.Context) (int64, error) {
	return ev.getLastHeight(ctx)
}

func (ev *EventStore) SetLastProcessedBlock(ctx context.Context, height int64) error {
	return ev.setLastHeight(ctx, height)
}
