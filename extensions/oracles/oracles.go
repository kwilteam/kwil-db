package oracles

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/events"
)

var registeredOracles = make(map[string]Oracle)

type Oracle interface {
	Start(ctx context.Context, eventstore EventStore, config map[string]string, logger log.Logger) error
	Stop() error
}

func RegisterOracle(name string, oracle Oracle) error {
	if _, ok := registeredOracles[name]; ok {
		return fmt.Errorf("oracle with name %s already registered: ", name)
	}

	registeredOracles[name] = oracle
	return nil
}

func RegisteredOracles() map[string]Oracle {
	return registeredOracles
}

func GetOracle(name string) (Oracle, bool) {
	oracle, ok := registeredOracles[name]
	return oracle, ok
}

type EventStore interface {
	// KV returns a KV store to store metadata locally
	KV(prefix []byte) *events.KV

	// Store stores an event in the event store
	Store(ctx context.Context, data []byte, eventType string) error
}
