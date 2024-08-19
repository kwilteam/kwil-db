package cometbft

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft/privval"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	cometConfig "github.com/cometbft/cometbft/config"
	cometEd25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cometLog "github.com/cometbft/cometbft/libs/log"
	cometNodes "github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/types"

	"go.uber.org/zap"
)

// demotedInfoMsgs contains extremely spammy and generally uninformative Info
// messages from cometBFT that we will log at Debug level. These are exact
// messages, no grep patterns.
var demotedInfoMsgs = map[string]string{
	"indexed block events":             "",
	"committed state":                  "",
	"received proposal":                "",
	"received complete proposal block": "",
	"executed block":                   "",
	"Starting localClient service":     "proxy",
	"Timed out":                        "consensus", // only the one from consensus.(*timeoutTicker).timeoutRoutine, which seems to be normal
	"Could not check tx":               "mempool",
}

// demotedErrMsgs contain error-level messages that should not be logged as
// errors, and are instead logged at Warn level. For example, a peer hanging up
// on us is barely log-worthy and certainly not an error that indicates a
// serious problem, which is logged with a large call stack dump. The logs in
// cometbft/p2p.(*Switch) are the most egregious, but most of cometbft uses
// error logging quite liberally, probably because they have no Warn.
var demotedErrMsgs = map[string]string{
	"Stopping peer for error":                        "p2p",
	"error while stopping peer":                      "p2p",
	"Error stopping pool":                            "blocksync", // this almost always happens on shutdown
	"Stopped accept routine, as transport is closed": "p2p",
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
	logFun := lw.log.Error
	if module, quiet := demotedErrMsgs[msg]; quiet &&
		(module == "" || lw.module == module) {
		logFun = lw.log.Warn
	}
	logFun(msg, keyValsToFields(kvs)...)
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
	n := len(kvs)
	if n%2 != 0 {
		kvs = append(kvs, errors.New("missing value"))
	}
	fields := make([]zap.Field, 0, n/2)
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

// Parses genesis file to extract cometbft specific config
func genesisDocProvider(genDoc *types.GenesisDoc) cometNodes.GenesisDocProvider {
	return func() (*types.GenesisDoc, error) {
		return genDoc, nil
	}
}

func writeCometBFTConfigs(conf *cometConfig.Config, genDoc *types.GenesisDoc) error {
	// Save a copy the cometbft genesis and config files in the "abci/config"
	// folder expected by the `cometbft` cli app, used for inspect, etc.
	cometConfigPath := filepath.Join(conf.RootDir, "config")
	os.MkdirAll(cometConfigPath, 0755)
	cmtGenesisFile := filepath.Join(cometConfigPath, GenesisJSONName)
	if err := genDoc.SaveAs(cmtGenesisFile); err != nil {
		return fmt.Errorf("failed to write cometbft genesis.json formatted file to %v: %w", cmtGenesisFile, err)
	}

	// Now "abci/config/config.toml"
	conf.RPC.TLSCertFile, conf.RPC.TLSKeyFile = "", "" // not needed for debugging and recovery
	cmtConfigFile := filepath.Join(cometConfigPath, "config.toml")
	cometConfig.WriteConfigFile(cmtConfigFile, conf)

	cfgREADME := `This config folder is used to echo the in-memory CometBFT configuration
that is generated from kwild's config and genesis files. This folder and the files
within are useful for running certain debugging commands with the official cometbft
command line application (github.com/cometbft/cometbft/cmd/cometbft).

WARNING: These files are overwritten on kwild startup.`
	cmtREADMEFile := filepath.Join(cometConfigPath, "README")
	if _, err := os.Stat(cmtREADMEFile); errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile(cmtREADMEFile, []byte(cfgREADME), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to write cometbft config folder README: %v", err)
		}
	}

	return nil
}

// NewCometBftNode creates a new CometBFT node.
func NewCometBftNode(ctx context.Context, app abciTypes.Application, conf *cometConfig.Config,
	genDoc *types.GenesisDoc, privateKey cometEd25519.PrivKey, atomicStore privval.AtomicReadWriter,
	logger *log.Logger) (*CometBftNode, error) {
	if err := writeCometBFTConfigs(conf, genDoc); err != nil {
		return nil, fmt.Errorf("failed to write the effective cometbft config files: %w", err)
	}

	logger.Debugf("%#v", *conf.StateSync)

	err := conf.ValidateBasic()
	if err != nil {
		return nil, fmt.Errorf("invalid node config: %w", err)
	}

	privateValidator, err := privval.NewValidatorSigner(privateKey, atomicStore)
	if err != nil {
		return nil, fmt.Errorf("failed to create private validator: %v", err)
	}

	node, err := cometNodes.NewNodeWithContext(
		ctx,
		conf,
		privateValidator,
		&p2p.NodeKey{
			PrivKey: privateKey,
		},
		proxy.NewConnSyncLocalClientCreator(app), // "connection-synchronized" local client
		genesisDocProvider(genDoc),
		cometConfig.DefaultDBProvider,
		cometNodes.DefaultMetricsProvider(conf.Instrumentation),
		NewLogWrapper(logger),
	)
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			err = context.Canceled // canceled and comet forgot to use %w in doHandshake and elsewhere
		}
		return nil, fmt.Errorf("failed to create CometBFT node: %w", err)
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

// IsCatchup returns true if the node is operating in catchup / blocksync
// mode.  If the node is caught up with the network, it returns false.
func (n *CometBftNode) IsCatchup() bool {
	return n.Node.ConsensusReactor().WaitSync()
}

func (n *CometBftNode) RemovePeer(nodeID string) error {
	peerInfo := n.Node.Switch().Peers()
	id := p2p.ID(nodeID)
	peer := peerInfo.Get(id)
	if peer == nil {
		return fmt.Errorf("peer %s not found", nodeID)
	}

	n.Node.Switch().StopPeerGracefully(peer)
	return nil
}
