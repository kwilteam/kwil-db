package oracles

import (
	"context"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
)

var registeredOracles = make(map[string]Oracle)

// Uses datastores, EventStore, and Logger from core/types

// Method1: using map[string]string
type Oracle interface {
	Name() string
	Start(ctx context.Context, datastores types.Datastores, eventstore types.EventStore, logger log.Logger, metadata map[string]string) error
	Stop() error
}

func RegisterOracle(name string, oracle Oracle) error {
	if _, ok := registeredOracles[name]; ok {
		panic("oracle of same name already registered: " + name)
	}

	registeredOracles[name] = oracle
	return nil
}

func RegisteredOracles() map[string]Oracle {
	return registeredOracles
}

func OracleByName(name string) (Oracle, bool) {
	oracle, ok := registeredOracles[name]
	return oracle, ok
}
