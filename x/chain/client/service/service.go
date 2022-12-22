package service

import (
	"fmt"
	"kwil/x/cfgx"
	"kwil/x/chain"
	"kwil/x/chain/client/dto"
	provider "kwil/x/chain/provider/dto"
	"kwil/x/logx"
	"time"
)

// chainClient implements the ChainClient interface
type chainClient struct {
	provider              provider.ChainProvider
	log                   logx.SugaredLogger
	maxBlockInterval      time.Duration
	requiredConfirmations int64
	chainCode             chain.ChainCode
}

func NewChainClient(cfg cfgx.Config, prov provider.ChainProvider) (dto.ChainClient, error) {

	chainCode := chain.ChainCode(cfg.Int64("chain-code", 0))

	providerEndpoint := cfg.String("provider-endpoint")
	if providerEndpoint == "" {
		return nil, fmt.Errorf("provider endpoint is required")
	}

	// TODO: @Randal- I am dialing the ETH provider here... I also do this in the contract.  Is it ok to dial twice, or should we focus on using a shared connection?

	return &chainClient{
		provider:              prov,
		log:                   logx.New().Named("chain-client").Sugar(),
		maxBlockInterval:      time.Duration(cfg.Int64("reconnection-interval", 30)) * time.Second,
		requiredConfirmations: cfg.Int64("required-confirmations", 12),
		chainCode:             chainCode,
	}, nil
}

/*
func newChainSpecificClient(endpoint string, chainCode dto.ChainCode) (blockClient, error) {
	switch chainCode {
	case dto.ETHEREUM:
		return evmclient.New(endpoint, chainCode.ToChainId())
	case dto.GOERLI:
		return evmclient.New(endpoint, chainCode.ToChainId())
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainCode))
	}
}
*/
