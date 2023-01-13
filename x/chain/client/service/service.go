package service

import (
	client2 "kwil/kwil/client"
	"kwil/x/cfgx"
	"kwil/x/chain"
	"kwil/x/chain/client"
	"kwil/x/chain/client/dto"
	"kwil/x/chain/provider"
	"kwil/x/logx"
	"time"
)

// ChainClient implements the ChainClient interface
type chainClient struct {
	provider              provider.ChainProvider
	log                   logx.SugaredLogger
	reconnectInterval     time.Duration
	requiredConfirmations int64
	chainCode             chain.ChainCode
	lastBlock             int64
}

func NewChainClient(cfg cfgx.Config, prov provider.ChainProvider) (client.ChainClient, error) {

	chainCode := cfg.Int64(client2.ChainCodeFlag, 0)
	recInterval := cfg.Int64("reconnection-interval", 30)
	reqConfs := cfg.Int64("required-confirmations", 12)

	return NewChainClientExplicit(&dto.Config{
		Endpoint:              prov.Endpoint(),
		ChainCode:             chainCode,
		ReconnectionInterval:  recInterval,
		RequiredConfirmations: reqConfs,
	})
}

func NewChainClientExplicit(conf *dto.Config) (client.ChainClient, error) {
	prov, err := provider.New(conf.Endpoint, chain.ChainCode(conf.ChainCode))
	if err != nil {
		return nil, err
	}

	return &chainClient{
		provider:              prov,
		log:                   logx.New().Named("chain-client").Sugar(),
		reconnectInterval:     time.Duration(conf.ReconnectionInterval) * time.Second,
		requiredConfirmations: conf.RequiredConfirmations,
		chainCode:             chain.ChainCode(conf.ChainCode),
		lastBlock:             0,
	}, nil
}
