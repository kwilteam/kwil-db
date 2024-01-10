package oracles

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/log"
)

var registeredOracles = make(map[string]Oracle)

type Oracle interface {
	Initialize(ctx context.Context, eventstore EventStore, config map[string]string, logger log.Logger) error
	Start(ctx context.Context) error
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
