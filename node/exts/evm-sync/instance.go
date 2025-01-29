package evmsync

import (
	"context"
	"fmt"
	"sync"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/extensions/listeners"
	"github.com/kwilteam/kwil-db/node/exts/evm-sync/chains"
	orderedsync "github.com/kwilteam/kwil-db/node/exts/ordered-sync"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

// EventSyncer is the singleton that is responsible for syncing events from Ethereum.
var EventSyncer = &globalListenerManager{
	listeners: make(map[string]*listenerInfo),
}

// this file contains a thread-safe in-memory cache for the chains that the network cares about.

type globalListenerManager struct {
	// mu protects all fields in this struct
	mu sync.Mutex
	// listeners is a set of listeners
	listeners map[string]*listenerInfo
	// shouldListen is a flag that is set to true when the node should have
	// its listeners running.
	// We can probably get rid of this field, its mostly just to protect against unexpected
	// behavior by the rest of the code outside of this package.
	shouldListen bool
}

// EVMEventListenerConfig is the configuration for an EVM event listener.
type EVMEventListenerConfig struct {
	// UniqueName is a unique name for the listener.
	// It MUST be unique from all other listeners.
	UniqueName string
	// ContractAddresses is a list of contract addresses to listen to events from.
	ContractAddresses []string
	// EventSignatures is a list of event signatures to listen to.
	// All events from any contract configured matching any of these signatures will be emitted.
	// It is optional and defaults to all events.
	EventSignatures []string
	// Chain is the chain that the listener is listening to.
	Chain chains.Chain
	// Resolve is the function that will be called the Kwil network
	// has confirmed events from Ethereum.
	Resolve ResolveFunc
}

// RegisterListener registers a new listener.
// It should be called when a node starts up (e.g. on a precompile's
// OnStart method).
func (l *globalListenerManager) RegisterNewListener(ctx context.Context, db sql.DB, eng common.Engine, conf EVMEventListenerConfig) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, ok := l.listeners[conf.UniqueName]
	if ok {
		return fmt.Errorf("listener with name %s already registered", conf.UniqueName)
	}

	chainInfo, ok := chains.GetChainInfo(conf.Chain)
	if !ok {
		return fmt.Errorf("chain %s not found", conf.Chain)
	}

	err := orderedsync.Synchronizer.RegisterTopic(ctx, db, eng, conf.UniqueName,
		func(ctx context.Context, app *common.App, block *common.BlockContext, res *orderedsync.ResolutionMessage) error {
			// in our callback, we will deserialize the finalized message into ethereum logs and pass that to our own resolve function
			logs, err := deserializeLogs(res.Data)
			if err != nil {
				return err
			}

			return conf.Resolve(ctx, app, block, logs)
		})
	if err != nil {
		return err
	}

	doneCh := make(chan struct{})

	l.listeners[conf.UniqueName] = &listenerInfo{
		contractAddresses: conf.ContractAddresses,
		eventSignatures:   conf.EventSignatures,
		done:              doneCh,
		chain:             chainInfo,
		uniqueName:        conf.UniqueName,
	}

	return nil
}

// UnregisterListener unregisters a listener.
// It should be called when an extension gets unused
func (l *globalListenerManager) UnregisterListener(ctx context.Context, db sql.DB, eng common.Engine, uniqueName string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	info, ok := l.listeners[uniqueName]
	if !ok {
		return fmt.Errorf("listener with name %s not registered", uniqueName)
	}

	err := orderedsync.Synchronizer.UnregisterTopic(ctx, db, eng, uniqueName)
	if err != nil {
		return err
	}

	close(info.done)
	delete(l.listeners, uniqueName)

	return nil
}

// listen starts all listeners.
// If it returns an error, the node will stop
func (l *globalListenerManager) listen(ctx context.Context, service *common.Service, eventstore listeners.EventStore, syncConf *syncConfig) error {
	l.mu.Lock()

	if l.shouldListen {
		l.mu.Unlock()
		// protect against making duplicate connections
		return fmt.Errorf("already listening")
	}

	l.shouldListen = true
	defer func() {
		l.mu.Lock()
		l.shouldListen = false
		l.mu.Unlock()
	}()

	for _, info := range l.listeners {
		go info.listen(ctx, service, eventstore, syncConf)
	}

	l.mu.Unlock()
	<-ctx.Done()

	return nil
}

type ResolveFunc func(ctx context.Context, app *common.App, block *common.BlockContext, logs []ethtypes.Log) error

type listenerInfo struct {
	// done is a channel that is closed when the listener is done
	done chan struct{}
	// chain is the chain that the listener is listening to
	chain chains.ChainInfo
	// contractAddresses is a list of contract addresses to listen to events from
	contractAddresses []string
	// eventSignatures is a list of event signatures to listen to
	eventSignatures []string
	// uniqueName is the unique name of the listener
	uniqueName string
}

// listen makes a new client and starts listening for events.
// It does not return any errors because errors returned from listeners
// are fatal, and errors returned from this function are _very_ likely
// due to network errors (e.g. with the target RPC).
func (l *listenerInfo) listen(ctx context.Context, service *common.Service, eventstore listeners.EventStore, syncConf *syncConfig) {
	logger := service.Logger.New(l.uniqueName + "." + string(l.chain.Name))

	chainConf, err := getChainConf(service.LocalConfig.Extensions, l.chain.Name)
	if err != nil {
		logger.Error("failed to get chain config", "err", err)
		return
	}

	ethClient, err := newEthClient(ctx, chainConf.Provider, syncConf.MaxRetries, l.contractAddresses, l.eventSignatures, l.done, logger)
	if err != nil {
		logger.Error("failed to create evm client", "err", err)
		return
	}

	indiv := &individualListener{
		chain:            l.chain,
		syncConf:         syncConf,
		chainConf:        chainConf,
		client:           ethClient,
		orderedSyncTopic: l.uniqueName,
	}

	err = indiv.listen(ctx, eventstore, logger)
	if err != nil {
		logger.Error("error listening", "err", err)
	}
}
