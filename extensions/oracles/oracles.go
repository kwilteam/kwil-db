// package oracles provides the interface and registration for custom
// event-driven oracles. Oracles are designed to be non-deterministic, and
// can be used to trigger "Resolutions" on the local network. See package
// "extensions/resolutions" to define resolutions that can be voted on by the
// network.
package oracles

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/common"
)

// registeredOracles is a map of all registered oracles.
var registeredOracles = make(map[string]OracleFunc)

// OracleFunc is a function that allows custom oracles to be built with Kwil.
// The function is called when a node has come online, synced with the network, and is currently a validator.
// The function is expected to run for as long as the implementer wants the oracle to continue running.
// The passed context will be canceled when the node is shutting down, or when it is no longer a validator.
// It is expected that any resources associated with the oracle are cleaned up when the context is canceled.
// The function can be called many times if a node's validator status changes (e.g. becomes a validator, is
// removed as a validator, then becomes a validator again).
// The function can block indefinitely, but all resources should be cleaned up when the context is canceled
// to prevent memory leaks.
type OracleFunc func(ctx context.Context, service *common.Service, eventstore EventStore)

// RegisterOracle registers an oracle with the Kwil network.
// It should be called in the init function of the oracle's package.
// The name cannot have spaces in it.
func RegisterOracle(name string, oracle OracleFunc) error {
	name = strings.ToLower(name)

	// we protect against spaces in the name, because the KV
	// gives each oracle its own namespace. Spaces are used
	// to prevent collisions in the KV.
	for _, r := range name {
		if r == ' ' {
			return fmt.Errorf("oracle name cannot have spaces")
		}
	}

	if _, ok := registeredOracles[name]; ok {
		return fmt.Errorf("oracle with name %s already registered: ", name)
	}

	registeredOracles[name] = oracle
	return nil
}

// RegisteredOracles returns a map of all registered oracles.
func RegisteredOracles() map[string]OracleFunc {
	return registeredOracles
}

// GetOracle returns an oracle by its name*.
func GetOracle(name string) (OracleFunc, bool) {
	oracle, ok := registeredOracles[name]
	return oracle, ok
}

// EventStore is an interface that allows oracles to persist events, and track
// arbitrary metadata about its external source. It should be used to signal
// to the local Kwil node that the validator has seen an event, and that the
// event should be broadcast to the network.
// Events should be broadcast to the network using the Broadcast method.
// The KV store data is never forwarded to the network, and is simply
// for the convenience of the oracle implementer to persist metadata about
// the data source.
type EventStore interface {
	// Broadcast broadcasts an event seen by the local node to the network.
	// The eventType is a string that identifies the network should interpret the data.
	// The eventType should directly correspond to a "resolution" type, implemented in
	// the resolutions package. The data argument will be passed to the resolution's
	// ResolveFunc if enough nodes vote on the resolution.
	Broadcast(ctx context.Context, eventType string, data []byte) error

	// Set sets a key-value pair in the KV store.
	// The data is scoped to the local node, and is not broadcast to the network.
	Set(ctx context.Context, key []byte, value []byte) error
	// Get gets a value from the local node's KV store.
	Get(ctx context.Context, key []byte) ([]byte, error)
	// Delete deletes a value from the local node's KV store.
	// The data is scoped to the local node, and is not broadcast to the network.
	Delete(ctx context.Context, key []byte) error
}
