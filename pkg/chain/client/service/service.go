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
	log                   logger.SugaredLogger
	reconnectInterval     time.Duration
	requiredConfirmations int64
	chainCode             types.ChainCode
	lastBlock             int64
}

func NewChainClientExplicit(conf *dto.Config) (client.ChainClient, error) {
	prov, err := provider.New(conf.Endpoint, types.ChainCode(conf.ChainCode))
	if err != nil {
		return nil, err
	}

	return &chainClient{
		provider:              prov,
		log:                   logger.New().Named("chain-client").Sugar(),
		reconnectInterval:     time.Duration(conf.ReconnectionInterval) * time.Second,
		requiredConfirmations: conf.RequiredConfirmations,
		chainCode:             types.ChainCode(conf.ChainCode),
		lastBlock:             0,
	}, nil
}
