package service

import (
	"fmt"
	"kwil/pkg/chain/client"
	"kwil/pkg/chain/client/dto"
	"kwil/pkg/chain/contracts"
	"kwil/pkg/chain/provider"
	"kwil/pkg/chain/types"
	"kwil/pkg/log"
	"time"
)

const (
	// DefaultReconnectInterval is the default interval between reconnect attempts
	DefaultReconnectInterval = 30 * time.Second

	// DefaultRequiredConfirmations is the default number of confirmations required for a transaction to be considered final
	DefaultRequiredConfirmations = 12

	// DefaultChainCode is the default chain code.
	// We use Goerli by default for now.
	DefaultChainCode = types.ChainCode(2)

	// DefaultLastBlock is the default last block.
	DefaultLastBlock = int64(0)
)

// ChainClient implements the ChainClient interface
type chainClient struct {
	provider              provider.ChainProvider
	log                   log.Logger
	reconnectInterval     time.Duration
	requiredConfirmations int64
	chainCode             types.ChainCode
	lastBlock             int64
	contracts             contracts.Contracter
}

func NewChainClient(chainRpcUrl string, opts ...ChainClientOpts) (client.ChainClient, error) {
	cc := &chainClient{
		log:                   log.NewNoOp(),
		reconnectInterval:     DefaultReconnectInterval,
		requiredConfirmations: DefaultRequiredConfirmations,
		chainCode:             DefaultChainCode,
		lastBlock:             DefaultLastBlock,
	}

	for _, opt := range opts {
		opt(cc)
	}

	var err error
	cc.provider, err = provider.New(chainRpcUrl, types.ChainCode(cc.chainCode))
	if err != nil {
		return nil, fmt.Errorf("failed to create chain provider: %w", err)
	}

	cc.contracts = contracts.New(cc.provider)

	return cc, nil
}

// NewChainClientWithProvider creates a new ChainClient with a given provider.
// This is useful for testing and mocking.
func NewChainClientWithProvider(prov provider.ChainProvider, conf *dto.Config, logger log.Logger) (client.ChainClient, error) {
	return &chainClient{
		provider:              prov,
		log:                   logger.Named("chain_client"),
		reconnectInterval:     time.Duration(conf.ReconnectInterval) * time.Second,
		requiredConfirmations: conf.BlockConfirmation,
		chainCode:             types.ChainCode(conf.ChainCode),
		lastBlock:             0,
	}, nil
}

func (c *chainClient) Close() error {
	return c.provider.Close()
}

func (c *chainClient) Contracts() contracts.Contracter {
	return c.contracts
}
