package oracles

import (
	"context"
	"time"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/extensions/oracles"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
	"go.uber.org/zap"
)

// OracleMgr listens to the Validator state changes and node catch up status
// and starts or stops the oracles accordingly
// It starts the oracles only when the node is a validator and is caught up
// It stops the running oracles when the node loses its validator status
type OracleMgr struct {
	ctx        context.Context
	config     map[string]map[string]string
	eventStore oracles.EventStore
	cometNode  *cometbft.CometBftNode

	// oraclesUp specifies whether the oracles are running currently or not
	oraclesUp bool
	// Status is a channel that will specify whether to run oracles or not
	// It depends on Validator Status and Node's Catch Up Status
	// Oracles can only be run when the node is a validator and is already caught up
	valStatus <-chan bool

	logger log.Logger
}

func NewOracleMgr(ctx context.Context, config map[string]map[string]string, eventStore oracles.EventStore, node *cometbft.CometBftNode, valStatusChan chan bool, logger log.Logger) *OracleMgr {
	return &OracleMgr{
		ctx:        ctx,
		config:     config,
		eventStore: eventStore,
		cometNode:  node,
		valStatus:  valStatusChan,
		oraclesUp:  false,
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
			case status := <-omgr.valStatus:
				if omgr.cometNode.IsCatchup() {
					continue
				}

				if !omgr.oraclesUp && status {
					// Start the oracles if they are not running
					omgr.oraclesUp = true
					omgr.logger.Info("Node's a validator and caught up with the network, starting oracles")
					omgr.startOracles()
				} else if omgr.oraclesUp && !status {
					// Stop the oracles if they are running
					omgr.logger.Info("Node's no longer the validator, stopping oracles")
					omgr.oraclesUp = false
					omgr.stopOracles()
				}
			case <-omgr.ctx.Done():
				return
			default:
				time.Sleep(100 * time.Millisecond)
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
