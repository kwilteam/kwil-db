package ethereum

import (
	"fmt"
	chainClient "kwil/pkg/chain/client"
	chainClientService "kwil/pkg/chain/client/service"
	"kwil/pkg/fund"
	"kwil/pkg/log"
	"kwil/x/contracts/escrow"
	"kwil/x/contracts/token"
)

type Client struct {
	Escrow escrow.EscrowContract
	Token  token.TokenContract

	// TODO: rename this
	ChainClient chainClient.ChainClient

	Config *fund.Config
}

func NewClient(cfg *fund.Config, logger log.Logger) (*Client, error) {
	chnClient, err := chainClientService.NewChainClientExplicit(&cfg.Chain, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create chain client: %v", err)
	}

	// escrow
	escrowCtr, err := escrow.New(chnClient, cfg.Wallet, cfg.PoolAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create escrow contract: %v", err)
	}

	// erc20 address
	tokenAddress := escrowCtr.TokenAddress()

	// erc20
	erc20Ctr, err := token.New(chnClient, cfg.Wallet, tokenAddress)
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
