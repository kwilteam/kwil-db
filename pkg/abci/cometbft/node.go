package cometbft

import (
	"errors"
	"fmt"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	cometConfig "github.com/cometbft/cometbft/config"
	cometEd25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cometLog "github.com/cometbft/cometbft/libs/log"
	cometNodes "github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/proxy"
	"github.com/kwilteam/kwil-db/pkg/abci/cometbft/privval"
	"github.com/kwilteam/kwil-db/pkg/log"
	"go.uber.org/zap"
)

type LogWrapper struct {
	log *log.Logger
}

func NewLogWrapper(log *log.Logger) *LogWrapper {
	log = log.WithOptions(zap.AddCallerSkip(1))
	return &LogWrapper{log}
}

func (lw *LogWrapper) Debug(msg string, kvs ...any) {
	lw.log.Debug(msg, keyValsToFields(kvs)...)
}

func (lw *LogWrapper) Info(msg string, kvs ...any) {
	lw.log.Info(msg, keyValsToFields(kvs)...)
}

func (lw *LogWrapper) Error(msg string, kvs ...any) {
	lw.log.Error(msg, keyValsToFields(kvs)...)
}

func (lw *LogWrapper) With(kvs ...any) cometLog.Logger {
	fields := keyValsToFields(kvs)
	return NewLogWrapper(lw.log.With(fields...))
}

func keyValsToFields(kvs []any) []zap.Field {
	if len(kvs)%2 != 0 {
		kvs = append(kvs, errors.New("missing value"))
	}
	n := len(kvs) / 2
	fields := make([]zap.Field, 0, n)
	for i := 0; i < n; i += 2 {
		fields = append(fields, zap.Any(toString(kvs[i]), kvs[i+1]))
	}
	return fields
}

func toString(x any) string {
	switch xt := x.(type) {
	case string:
		return xt
	case fmt.Stringer:
		return xt.String()
	}
	return fmt.Sprintf("%v", x)
}

type CometBftNode struct {
	Node *cometNodes.Node
}

// NewCometBftNode creates a new CometBFT node.
func NewCometBftNode(app abciTypes.Application, conf *cometConfig.Config, privateKey cometEd25519.PrivKey,
	atomicStore privval.AtomicReadWriter, log *log.Logger) (*CometBftNode, error) {

	logger := NewLogWrapper(log)

	privateValidator, err := privval.NewValidatorSigner(privateKey, atomicStore)
	if err != nil {
		return nil, fmt.Errorf("failed to create private validator: %v", err)
	}

	node, err := cometNodes.NewNode(
		conf,
		privateValidator,
		&p2p.NodeKey{
			PrivKey: privateKey,
		},
		proxy.NewLocalClientCreator(app),
		cometNodes.DefaultGenesisDocProviderFunc(conf),
		cometNodes.DefaultDBProvider,
		cometNodes.DefaultMetricsProvider(conf.Instrumentation),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CometBFT node: %v", err)
	}

	return &CometBftNode{
		Node: node,
	}, nil
}

// Start starts the CometBFT node.
func (n *CometBftNode) Start() error {
	return n.Node.Start()
}

// Stop stops the CometBFT node.
func (n *CometBftNode) Stop() error {
	return n.Node.Stop()
}
