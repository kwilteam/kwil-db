package cometbft

import (
	"fmt"
	"os"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	cometConfig "github.com/cometbft/cometbft/config"
	cometFlags "github.com/cometbft/cometbft/libs/cli/flags"
	cometLog "github.com/cometbft/cometbft/libs/log"
	cometNodes "github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/proxy"
)

type CometBftNode struct {
	node *cometNodes.Node
}

type Config struct {
	// directory is the path to where files should be read and written for cometbft
	Directory string

	// LogLevel is the log level for cometbft
	LogLevel string
}

func NewCometBftNode(app abciTypes.Application, config *Config) (*CometBftNode, error) {
	conf := cometConfig.DefaultConfig().SetRoot(config.Directory)
	logger := cometLog.NewTMLogger(cometLog.NewSyncWriter(os.Stdout))
	logger, err := cometFlags.ParseLogLevel(conf.LogLevel, logger, config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %v", err)
	}

	privKey := &CometBftPrivateKey{}

	node, err := cometNodes.NewNode(
		conf,

		// ideally this takes our own custom implementation, since CometBFT
		// only supports signers created from files, or remote signers that need a connection
		newPrivateValidator(privKey),

		// ideally we can use our own custom implementation.
		// CometBFT supports both ED25519 and SECP256K1, but the default seems to be ED25519.
		// either translating our ED25519 and SECP256K1 keys to CometBFT's format, or
		// creating our own CometBet PrivKey implementation is ideal.
		// It does seems that internally, CometBFT is tied to their own implementations of PrivKey,
		// but I am not certain
		&p2p.NodeKey{
			PrivKey: &CometBftPrivateKey{},
		},
		proxy.NewLocalClientCreator(app),
		cometNodes.DefaultGenesisDocProviderFunc(conf),

		// There coukd be a good reason to switch this with our own implementation,
		// seems lower priority than others though
		cometNodes.DefaultDBProvider,
		cometNodes.DefaultMetricsProvider(conf.Instrumentation),
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CometBFT node: %v", err)
	}

	return &CometBftNode{
		node: node,
	}, nil
}
