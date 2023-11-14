package events

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types/chain"
	"github.com/kwilteam/kwil-db/internal/sql"
)

type Datastore interface {
	Execute(ctx context.Context, stmt string, args map[string]any) error
	Query(ctx context.Context, query string, args map[string]any) ([]map[string]any, error)
	Prepare(stmt string) (sql.Statement, error)
}

type EventStore struct {
	address []byte
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
	event.Receiver = ev.address
	return ev.addLocalEvent(ctx, event)
}
