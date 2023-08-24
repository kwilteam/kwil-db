package cometbft

import (
	"fmt"
	"os"
	"time"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	cometConfig "github.com/cometbft/cometbft/config"
	cometEd25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cometFlags "github.com/cometbft/cometbft/libs/cli/flags"
	cometLog "github.com/cometbft/cometbft/libs/log"
	cometNodes "github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/proxy"
	"github.com/kwilteam/kwil-db/pkg/abci/cometbft/privval"
)

type CometBftNode struct {
	Node *cometNodes.Node
}

// NewCometBftNode creates a new CometBFT node.
func NewCometBftNode(app abciTypes.Application, privateKey []byte, atomicStore privval.AtomicReadWriter, directory string, logLevel string) (*CometBftNode, error) {
	conf := cometConfig.DefaultConfig().SetRoot(directory)

	// TODO: this is temporary hack, we need to use KWILD config
	//conf.LogLevel = "debug"
	// create blocks every 5 seconds
	//conf.Consensus.CreateEmptyBlocks = true
	//conf.Consensus.TimeoutCommit = 5 * time.Second
	// create blocks every 5 seconds or when txs are received
	conf.Consensus.CreateEmptyBlocks = false
	conf.Consensus.CreateEmptyBlocksInterval = 5 * time.Second
	conf.RPC.ListenAddress = "tcp://0.0.0.0:26657"
	fmt.Printf("conf: %+v\n", conf)
	fmt.Printf("conf.concensus: %+v\n", conf.Consensus)
	fmt.Printf("conf.rpc: %+v\n", conf.RPC)
	fmt.Printf("conf.mempool %+v\n", conf.Mempool)
	fmt.Printf("conf.Blocksync %+v\n", conf.BlockSync)
	fmt.Printf("conf.txIndex %+v\n", conf.TxIndex)
	///////////////////////////////////////////

	logger := cometLog.NewTMLogger(cometLog.NewSyncWriter(os.Stdout))
	logger, err := cometFlags.ParseLogLevel(conf.LogLevel, logger, logLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %v", err)
	}

	privateValidator, err := privval.NewValidatorSigner(privateKey, atomicStore)
	if err != nil {
		return nil, fmt.Errorf("failed to create private validator: %v", err)
	}

	node, err := cometNodes.NewNode(
		conf,
		privateValidator,
		&p2p.NodeKey{
			PrivKey: cometEd25519.PrivKey(privateKey),
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
