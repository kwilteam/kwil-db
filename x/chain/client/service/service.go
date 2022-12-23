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
	maxBlockInterval      time.Duration
	requiredConfirmations int64
	chainCode             chain.ChainCode
	lastBlock             int64
}

func NewChainClient(cfg cfgx.Config, prov provider.ChainProvider) dto.ChainClient {

	chainCode := cfg.Int64("chain-code", 0)
	blockInterval := cfg.Int64("reconnection-interval", 30)
	reqConfs := cfg.Int64("required-confirmations", 12)

	return NewChainClientNoConfig(prov, chainCode, blockInterval, reqConfs)
}

func NewChainClientNoConfig(prov provider.ChainProvider, chainCode int64, recInterval int64, reqConfs int64) dto.ChainClient {
	return &chainClient{
		provider:              prov,
		log:                   logx.New().Named("chain-client").Sugar(),
		maxBlockInterval:      time.Duration(recInterval) * time.Second,
		requiredConfirmations: reqConfs,
		chainCode:             chain.ChainCode(chainCode),
		lastBlock:             0,
	}
}
