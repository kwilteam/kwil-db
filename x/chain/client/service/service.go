package service

import (
	"kwil/x/cfgx"
	"kwil/x/chain"
	"kwil/x/chain/client/dto"
	provider "kwil/x/chain/provider/dto"
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

func NewChainClient(cfg cfgx.Config, prov provider.ChainProvider) dto.ChainClient {

	chainCode := cfg.Int64("chain-code", 0)
	recInterval := cfg.Int64("reconnection-interval", 30)
	reqConfs := cfg.Int64("required-confirmations", 12)

	return NewChainClientExplicit(prov, &dto.Config{
		ChainCode:             chainCode,
		ReconnectionInterval:  recInterval,
		RequiredConfirmations: reqConfs,
	})
}

func NewChainClientExplicit(prov provider.ChainProvider, conf *dto.Config) dto.ChainClient {

	return &chainClient{
		provider:              prov,
		log:                   logx.New().Named("chain-client").Sugar(),
		reconnectInterval:     time.Duration(conf.ReconnectionInterval) * time.Second,
		requiredConfirmations: conf.RequiredConfirmations,
		chainCode:             chain.ChainCode(conf.ChainCode),
		lastBlock:             0,
	}
}
