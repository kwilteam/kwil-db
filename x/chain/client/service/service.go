package service

import (
	"kwil/x/chain/client"
	"kwil/x/chain/client/dto"
	"kwil/x/chain/provider"
	"kwil/x/chain/types"
	"kwil/x/logx"
	"time"
)

// ChainClient implements the ChainClient interface
type chainClient struct {
	provider              provider.ChainProvider
	log                   logx.SugaredLogger
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
		log:                   logx.New().Named("chain-client").Sugar(),
		reconnectInterval:     time.Duration(conf.ReconnectionInterval) * time.Second,
		requiredConfirmations: conf.RequiredConfirmations,
		chainCode:             types.ChainCode(conf.ChainCode),
		lastBlock:             0,
	}, nil
}
