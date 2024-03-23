package listeners

import (
	"bytes"
	"context"
	"time"

	"go.uber.org/zap"

	common "github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/listeners"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
	"github.com/kwilteam/kwil-db/internal/voting"
)

// ListenerManager listens for any Validator state changes and node catch up status
// and starts or stops the listeners accordingly
// It starts the listeners only when the node is a validator and is caught up
// It stops the running listeners when the node loses its validator status
type ListenerManager struct {
	config     map[string]map[string]string
	eventStore *voting.EventStore
	vstore     ValidatorGetter
	cometNode  *cometbft.CometBftNode
	// pubKey is the public key of the node
	pubKey []byte

	logger log.Logger

	cancel context.CancelFunc // cancels the context for the listener manager
}

// ValidatorGetter is able to read the current validator set.
type ValidatorGetter interface {
	GetValidators(ctx context.Context) ([]*types.Validator, error)
}

func NewListenerManager(config map[string]map[string]string, eventStore *voting.EventStore, node *cometbft.CometBftNode, nodePubKey []byte, vstore ValidatorGetter, logger log.Logger) *ListenerManager {
	return &ListenerManager{
		config:     config,
		eventStore: eventStore,
		vstore:     vstore,
		cometNode:  node,
		pubKey:     nodePubKey,
		logger:     logger,
	}
}

// Start starts the listener manager.
// It will block until Stop is called.
func (omgr *ListenerManager) Start() error {
	ctx, cancel := context.WithCancel(context.Background()) // context that will be canceled when the manager shuts down
	omgr.cancel = cancel
	// Listen for status changes
	// Start oracles if the node is a validator and is caught up
	// Stop the oracles, if the node is not a validator
	// cancel function for the oracle instance
	// if it is nil, then the oracle is not running
	var listenerInstanceCancel context.CancelFunc

	var errChan = make(chan error, 1)

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errChan:
			return err
		case <-time.After(1 * time.Second):
			// still in the catch up mode, do nothing
			if omgr.cometNode.IsCatchup() {
				continue
			}

			validators, err := omgr.vstore.GetValidators(ctx)
			if err != nil {
				omgr.logger.Warn("failed to get validators", zap.Error(err))
				return err
			}

			isValidator := false
			for _, val := range validators {
				if bytes.Equal(val.PubKey, omgr.pubKey) {
					isValidator = true
					break
				}
			}

			if listenerInstanceCancel == nil && isValidator {
				// inner context to manage the listeners
				// this context will be cancelled when the node loses its validator status
				// it will also be cancelled when the listener manager is stopped
				ctx2, cancel2 := context.WithCancel(ctx)
				listenerInstanceCancel = cancel2

				omgr.logger.Info("Node is a validator and caught up with the network, starting listeners")

				for name, start := range listeners.RegisteredListeners() {
					go func(start listeners.ListenFunc, name string) {
						err := start(ctx2, &common.Service{
							Logger:           omgr.logger.Named(name).Sugar(),
							ExtensionConfigs: omgr.config,
						}, &scopedKVEventStore{
							ev: omgr.eventStore,
							// we add a space to prevent collisions in the KV
							// oracle names cannot have spaces
							KV: omgr.eventStore.KV([]byte(name + " ")),
						})
						if err != nil {
							// if error is returned, shutdown manager
							omgr.logger.Error("Oracle failed", zap.String("oracle", name), zap.Error(err))
							errChan <- err
						}
					}(start, name)

				}
			} else if listenerInstanceCancel != nil && !isValidator {
				// Stop the listeners if they are running
				omgr.logger.Info("Node is no longer a validator, stopping listeners")
				listenerInstanceCancel()
				listenerInstanceCancel = nil
			}
		}
	}
}

func (omgr *ListenerManager) Stop() {
	omgr.cancel()
}

// scopedKVEventStore scopes the event store's kv store to the listener's name
type scopedKVEventStore struct {
	ev *voting.EventStore
	*voting.KV
}

// Broadcast broadcasts an event to the event store.
func (e *scopedKVEventStore) Broadcast(ctx context.Context, eventType string, data []byte) error {
	return e.ev.Store(ctx, data, eventType)
}
