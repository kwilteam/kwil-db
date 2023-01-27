package ethereum

import (
	"fmt"
	chainClient "kwil/x/chain/client"
	chainClientDto "kwil/x/chain/client/dto"
	chainClientService "kwil/x/chain/client/service"
	"kwil/x/contracts/escrow"
	"kwil/x/contracts/token"
	"kwil/x/fund"
)

type Client struct {
	Escrow escrow.EscrowContract
	Token  token.TokenContract

	// TODO: rename this
	ChainClient chainClient.ChainClient

	Config *fund.Config
}

func NewClient(cfg *fund.Config) (*Client, error) {
	chnClient, err := chainClientService.NewChainClientExplicit(&chainClientDto.Config{
		ChainCode:             int64(cfg.ChainCode),
		Endpoint:              cfg.Provider,
		ReconnectionInterval:  cfg.ReConnectionInterval,
		RequiredConfirmations: cfg.RequiredConfirmations,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chain client: %v", err)
	}

	// escrow
	escrowCtr, err := escrow.New(chnClient, cfg.PrivateKey, cfg.PoolAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create escrow contract: %v", err)
	}

	// erc20 address
	tokenAddress := escrowCtr.TokenAddress()

	// erc20
	erc20Ctr, err := token.New(chnClient, cfg.PrivateKey, tokenAddress)
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

func (c *Client) GetConfig() *fund.Config {
	return c.Config
}
