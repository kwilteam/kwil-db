package oracles

import (
	"bytes"
	"context"
	"time"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
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
	vstore     ValidatorGetter
	cometNode  *cometbft.CometBftNode
	// pubKey is the public key of the node
	pubKey []byte

	// oraclesUp specifies whether the oracles are running currently or not
	oraclesUp bool
	done      chan bool

	logger log.Logger
}

// ValidatorGetter is able to read the current validator set.
type ValidatorGetter interface {
	GetValidators(ctx context.Context) ([]*types.Validator, error)
}

func NewOracleMgr(ctx context.Context, config map[string]map[string]string, eventStore oracles.EventStore, node *cometbft.CometBftNode, nodePubKey []byte, vstore ValidatorGetter, logger log.Logger) *OracleMgr {
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

				validators, err := omgr.vstore.GetValidators(omgr.ctx)
				if err != nil {
					omgr.logger.Warn("failed to get validators", zap.Error(err))
					// panic?
					continue
				}

				isValidator := false
				for _, val := range validators {
					if bytes.Equal(val.PubKey, omgr.pubKey) {
						isValidator = true
						break
					}
				}

				if !omgr.oraclesUp && isValidator {
					// Start the oracles if they are not running
					omgr.oraclesUp = true
					omgr.logger.Info("Node's a validator and caught up with the network, starting oracles")
					omgr.startOracles()
				} else if omgr.oraclesUp && !isValidator {
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
	// stop configured oracles
	for oracleName := range omgr.config {
		omgr.logger.Info("stopping oracle", zap.String("name", oracleName))
		//check if oracle is registered
		oracleInst, ok := oracles.GetOracle(oracleName)
		if !ok {
			omgr.logger.Warn("Trying to stop unregistered oracle", zap.String("Oracle Name", oracleName), zap.Any("Config", omgr.config[oracleName]))
			return
		}

		// stop the oracle
		oracleInst.Stop()
	}
}

func (omgr *OracleMgr) startOracles() {
	// start configured oracles
	for oracleName := range omgr.config {
		//check if oracle is registered
		oracleInst, ok := oracles.GetOracle(oracleName)
		if !ok {
			omgr.logger.Warn("Trying to start unregistered oracle", zap.String("Oracle Name", oracleName), zap.Any("Config", omgr.config[oracleName]))
			return
		}

		omgr.logger.Info("starting oracle", zap.String("name", oracleName))
		go func(name string, inst oracles.Oracle) {
			if err := inst.Start(omgr.ctx, omgr.eventStore, omgr.config[name], *omgr.logger.Named(name)); err != nil {
				omgr.logger.Warn("failed to start oracle", zap.String("name", name), zap.Error(err))
			}
		}(oracleName, oracleInst)

	}
}
