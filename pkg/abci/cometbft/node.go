package cometbft

import (
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/abci/cometbft/privval"
	"github.com/kwilteam/kwil-db/pkg/log"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	cometConfig "github.com/cometbft/cometbft/config"
	cometEd25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cometLog "github.com/cometbft/cometbft/libs/log"
	cometNodes "github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/proxy"

	"go.uber.org/zap"
)

// demotedInfoMsgs contains extremely spammy and generally uninformative Info
// messages from cometBFT that we will log at Debug level. These are exact
// messages, no grep patterns.
var demotedInfoMsgs = map[string]string{
	"indexed block exents":             "",
	"indexed block events":             "",
	"committed state":                  "",
	"received proposal":                "",
	"received complete proposal block": "",
	"executed block":                   "",
	"Timed out":                        "consensus", // only the one from consensus.(*timeoutTicker).timeoutRoutine, which seems to be normal
}

// LogWrapper that implements cometbft's Logger interface.
type LogWrapper struct {
	log *log.Logger

	// comet bft typically sets a logger with the "module" key. We record this
	// whenever it is set so that we can gain context about the log call.
	module string
}

// NewLogWrapper creates a new LogWrapper using the provided Kwil Logger.
func NewLogWrapper(log *log.Logger) *LogWrapper {
	log = log.WithOptions(zap.AddCallerSkip(1))
	return &LogWrapper{
		log: log,
	}
}

func (lw *LogWrapper) Debug(msg string, kvs ...any) {
	lw.log.Debug(msg, keyValsToFields(kvs)...)
}

func (lw *LogWrapper) Info(msg string, kvs ...any) {
	logFun := lw.log.Info
	if module, quiet := demotedInfoMsgs[msg]; quiet &&
		(module == "" || lw.module == module) {
		logFun = lw.log.Debug
	}
	logFun(msg, keyValsToFields(kvs)...)
}

func (lw *LogWrapper) Error(msg string, kvs ...any) {
	// fields := append(keyValsToFields(kvs), zap.Stack("stacktrace"))
	// lw.log.Error(msg, fields...)
	lw.log.Error(msg, keyValsToFields(kvs)...)
}

func (lw *LogWrapper) With(kvs ...any) cometLog.Logger {
	fields := keyValsToFields(kvs)
	module := lw.module
	for _, f := range fields {
		if f.Key == "module" {
			module = f.String
		}
	}
	return &LogWrapper{
		log:    lw.log.With(fields...),
		module: module,
	}
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

	err := conf.ValidateBasic()
	if err != nil {
		return nil, fmt.Errorf("invalid node config: %w", err)
	}

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
		proxy.NewConnSyncLocalClientCreator(app), // "connection-synchronized" local client
		cometNodes.DefaultGenesisDocProviderFunc(conf),
		cometConfig.DefaultDBProvider,
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
