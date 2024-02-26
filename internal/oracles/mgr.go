package oracles

import (
	"bytes"
	"context"
	"time"

	"go.uber.org/zap"

	common "github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/oracles"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
	"github.com/kwilteam/kwil-db/internal/events"
)

// OracleMgr listens for any Validator state changes and node catch up status
// and starts or stops the oracles accordingly
// It starts the oracles only when the node is a validator and is caught up
// It stops the running oracles when the node loses its validator status
type OracleMgr struct {
	config     map[string]map[string]string
	eventStore *events.EventStore
	vstore     ValidatorGetter
	cometNode  *cometbft.CometBftNode
	// pubKey is the public key of the node
	pubKey []byte

	logger log.Logger

	cancel context.CancelFunc // cancels the context for the oracle manager
}

// ValidatorGetter is able to read the current validator set.
type ValidatorGetter interface {
	GetValidators(ctx context.Context) ([]*types.Validator, error)
}

func NewOracleMgr(config map[string]map[string]string, eventStore *events.EventStore, node *cometbft.CometBftNode, nodePubKey []byte, vstore ValidatorGetter, logger log.Logger) *OracleMgr {
	return &OracleMgr{
		config:     config,
		eventStore: eventStore,
		vstore:     vstore,
		cometNode:  node,
		pubKey:     nodePubKey,
		logger:     logger,
	}
}

// Start starts the oracle manager.
// It will block until Stop is called.
func (omgr *OracleMgr) Start() {
	ctx, cancel := context.WithCancel(context.Background()) // context that will be canceled when the manager shuts down
	omgr.cancel = cancel
	// Listen for status changes
	// Start oracles if the node is a validator and is caught up
	// Stop the oracles, if the node is not a validator
	go func() {
		// cancel function for the oracle instance
		// if it is nil, then the oracle is not running
		var oracleInstanceCancel context.CancelFunc

		for {
			// still in the catch up mode, do nothing
			if omgr.cometNode.IsCatchup() {
				continue
			}

			validators, err := omgr.vstore.GetValidators(ctx)
			if err != nil {
				omgr.logger.Warn("failed to get validators", zap.Error(err))
				break
			}

			isValidator := false
			for _, val := range validators {
				if bytes.Equal(val.PubKey, omgr.pubKey) {
					isValidator = true
					break
				}
			}

			if oracleInstanceCancel == nil && isValidator {
				// inner context to manage the oracles
				// this context will be cancelled when the node loses its validator status
				// it will also be cancelled when the oracle manager is stopped
				ctx2, cancel2 := context.WithCancel(ctx)
				oracleInstanceCancel = cancel2

				omgr.logger.Info("Node is a validator and caught up with the network, starting oracles")

				for name, start := range oracles.RegisteredOracles() {
					go start(ctx2, &common.Service{
						Logger:           omgr.logger.Named(name).Sugar(),
						ExtensionConfigs: omgr.config,
					}, &scopedKVEventStore{
						ev: omgr.eventStore,
						// we add a space to prevent collisions in the KV
						// oracle names cannot have spaces
						KV: omgr.eventStore.KV([]byte(name + " ")),
					})
					if err != nil {
						omgr.logger.Error("failed to start oracle", zap.String("name", name), zap.Error(err))
					}
				}
			} else if oracleInstanceCancel != nil && !isValidator {
				// Stop the oracles if they are running
				omgr.logger.Info("Node is no longer a validator, stopping oracles")
				oracleInstanceCancel()
				oracleInstanceCancel = nil
			}
			time.Sleep(1 * time.Second)
		}
	}()

	<-ctx.Done()
}

func (omgr *OracleMgr) Stop() {
	omgr.cancel()
}

// scopedKVEventStore scopes the event store's kv store to the oracle's name
type scopedKVEventStore struct {
	ev *events.EventStore
	*events.KV
}

// Broadcast broadcasts an event to the event store.
func (e *scopedKVEventStore) Broadcast(ctx context.Context, eventType string, data []byte) error {
	return e.ev.Store(ctx, data, eventType)
}
