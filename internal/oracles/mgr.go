package oracles

import (
	"context"
	"time"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/extensions/oracles"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
	"go.uber.org/zap"
)

// OracleMgr listens for any Validator state changes and node catch up status
// and starts or stops the oracles accordingly
// It starts the oracles only when the node is a validator and is caught up
// It stops the running oracles when the node loses its validator status
type OracleMgr struct {
	ctx        context.Context
	config     map[string]map[string]string
	eventStore oracles.EventStore
	vstore     ValidatorStore
	cometNode  *cometbft.CometBftNode
	// pubKey is the public key of the node
	pubKey []byte

	// oraclesUp specifies whether the oracles are running currently or not
	oraclesUp bool
	done      chan bool

	logger log.Logger
}

type ValidatorStore interface {
	// IsCurrent returns true if the validator is currently a validator.
	// It does not take into account uncommitted changes, but is thread-safe.
	IsCurrent(ctx context.Context, validator []byte) (bool, error)
}

func NewOracleMgr(ctx context.Context, config map[string]map[string]string, eventStore oracles.EventStore, node *cometbft.CometBftNode, nodePubKey []byte, vstore ValidatorStore, logger log.Logger) *OracleMgr {
	return &OracleMgr{
		ctx:        ctx,
		config:     config,
		eventStore: eventStore,
		vstore:     vstore,
		cometNode:  node,
		pubKey:     nodePubKey,
		oraclesUp:  false,
		done:       make(chan bool),
		logger:     logger,
	}
}

func (omgr *OracleMgr) Start() {
	omgr.listenForValidatorStatusChanges()
}

func (omgr *OracleMgr) listenForValidatorStatusChanges() {
	// Listen for status changes
	// Start oracles if the node is a validator and is caught up
	// Stop the oracles, if the node is not a validator
	go func() {
		omgr.logger.Info("starting oracle manager")
		for {
			select {
			// case status := <-omgr.valStatus:
			case <-omgr.ctx.Done():
				return
			case <-omgr.done:
				omgr.logger.Info("stopping oracle manager")
				return
			default:
				// still in the catch up mode, do nothing
				if omgr.cometNode.IsCatchup() {
					continue
				}

				// check if the node is a validator
				isVal, err := omgr.vstore.IsCurrent(omgr.ctx, omgr.pubKey)
				if err != nil {
					omgr.logger.Warn("failed to get validator status", zap.Error(err))
					continue
				}

				if !omgr.oraclesUp && isVal {
					// Start the oracles if they are not running
					omgr.oraclesUp = true
					omgr.logger.Info("Node's a validator and caught up with the network, starting oracles")
					omgr.startOracles()
				} else if omgr.oraclesUp && !isVal {
					// Stop the oracles if they are running
					omgr.logger.Info("Node's no longer the validator, stopping oracles")
					omgr.oraclesUp = false
					omgr.stopOracles()
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func (omgr *OracleMgr) Stop() {
	// Stop the oracles
	omgr.stopOracles()
}

func (omgr *OracleMgr) stopOracles() {
	oracles := oracles.RegisteredOracles()
	for name, oracle := range oracles {
		omgr.logger.Info("stopping oracle", zap.String("name", name))
		oracle.Stop()
	}
}

func (omgr *OracleMgr) startOracles() {
	oracles := oracles.RegisteredOracles()
	for name, oracle := range oracles {
		oracleName := name
		oracleInst := oracle

		omgr.logger.Info("starting oracle", zap.String("name", oracleName))
		go func() {
			if err := oracleInst.Start(omgr.ctx, omgr.eventStore, omgr.config[oracleName], *omgr.logger.Named(oracleName)); err != nil {
				omgr.logger.Warn("failed to start oracle", zap.String("name", oracleName), zap.Error(err))
			}
		}()
	}
}
