package ethereum

import (
	"fmt"
	chainClient "kwil/pkg/chain/client"
	ccs "kwil/pkg/chain/client/service"
	"kwil/pkg/chain/contracts/escrow"
	"kwil/pkg/chain/contracts/token"
	chainTypes "kwil/pkg/chain/types"
	"kwil/pkg/fund"
	"kwil/pkg/log"
)

type Client struct {
	Escrow escrow.EscrowContract
	Token  token.TokenContract

	// TODO: rename this
	ChainClient chainClient.ChainClient

	Config *fund.Config
}

func NewClient(cfg *fund.Config, logger log.Logger) (*Client, error) {
	chnClient, err := ccs.NewChainClient(cfg.Chain.RpcUrl,
		ccs.WithLogger(logger),
		ccs.WithChainCode(chainTypes.ChainCode(cfg.Chain.ChainCode)),
		ccs.WithRequiredConfirmations(cfg.Chain.BlockConfirmation),
		ccs.WithReconnectInterval(cfg.Chain.ReconnectInterval),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chain client: %v", err)
	}

	escrowCtr, err := chnClient.Contracts().Escrow(cfg.PoolAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create escrow contract: %v", err)
	}

	// erc20 address
	tokenAddress := escrowCtr.TokenAddress()

	// erc20
	erc20Ctr, err := chnClient.Contracts().Token(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create erc20 contract: %v", err)
	}

	return &Client{
		Escrow:      escrowCtr,
		Token:       erc20Ctr,
		ChainClient: chnClient,
		Config:      cfg,
	}, nil
}

func (c *Client) Close() error {
	return c.ChainClient.Close()
}
