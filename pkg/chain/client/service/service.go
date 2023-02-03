package service

import (
	"kwil/pkg/chain/client"
	"kwil/pkg/chain/client/dto"
	"kwil/pkg/chain/provider"
	"kwil/pkg/chain/types"
	"kwil/pkg/logger"
	"time"
)

// ChainClient implements the ChainClient interface
type chainClient struct {
	provider              provider.ChainProvider
	log                   logger.Logger
	reconnectInterval     time.Duration
	requiredConfirmations int64
	chainCode             types.ChainCode
	lastBlock             int64
}

func NewChainClientExplicit(conf *dto.Config, logger logger.Logger) (client.ChainClient, error) {
	prov, err := provider.New(conf.RpcUrl, types.ChainCode(conf.ChainCode))
	if err != nil {
		return nil, err
	}

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
